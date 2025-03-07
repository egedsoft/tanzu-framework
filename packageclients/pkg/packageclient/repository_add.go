// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package packageclient

import (
	"fmt"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

const (
	msgRunPackageRepositoryUpdate = "\n\nPlease consider using 'tanzu package repository update' to update the package repository with correct settings\n"
)

// AddRepository validates the provided input and adds the package repository CR to the cluster
func (p *pkgClient) AddRepository(o *packagedatamodel.RepositoryOptions, progress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType) {
	p.addRepository(o, progress, operationType)
}

// AddRepositorySync validates the provided input and adds the package repository CR to the cluster and returns error if any
func (p *pkgClient) AddRepositorySync(o *packagedatamodel.RepositoryOptions, operationType packagedatamodel.OperationType) error {
	pp := newPackageProgress()

	go p.addRepository(o, pp, operationType)

	initialMsg := fmt.Sprintf("Adding package repository '%s'", o.RepositoryName)
	if err := DisplayProgress(initialMsg, pp); err != nil {
		if err.Error() == packagedatamodel.ErrRepoAlreadyExists {
			log.Infof("Updated package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace)
			return nil
		}
		return err
	}
	log.Infof("Added package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace)
	return nil
}

func (p *pkgClient) addRepository(o *packagedatamodel.RepositoryOptions, progress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType) {
	var (
		pkgRepository *kappipkg.PackageRepository
		err           error
	)

	defer func() {
		if err != nil {
			progress.Err <- err
		}
		if operationType == packagedatamodel.OperationTypeInstall {
			close(progress.ProgressMsg)
			close(progress.Done)
		}
	}()

	if pkgRepository, err = p.kappClient.GetPackageRepository(o.RepositoryName, o.Namespace); err != nil {
		if !k8serror.IsNotFound(err) {
			return
		}
		err = nil
	}

	if pkgRepository != nil {
		progress.ProgressMsg <- fmt.Sprintf("Updating package repository '%s'", o.RepositoryName)
		p.UpdateRepository(o, progress, packagedatamodel.OperationTypeInstall)
		err = &packagedatamodel.PackagePluginNonCriticalError{Reason: packagedatamodel.ErrRepoAlreadyExists}
		return
	}

	progress.ProgressMsg <- "Validating provided settings for the package repository"
	if err = p.validateRepositoryAdd(o.RepositoryName, o.RepositoryURL, o.Namespace); err != nil {
		return
	}

	if o.CreateNamespace {
		progress.ProgressMsg <- fmt.Sprintf("Creating namespace '%s'", o.Namespace)
		if err = p.createNamespace(o.Namespace); err != nil {
			return
		}
	}

	newPackageRepo, err := p.newPackageRepository(o.RepositoryName, o.RepositoryURL, o.Namespace)
	if err != nil {
		return
	}

	progress.ProgressMsg <- "Creating package repository resource"

	if err = p.kappClient.CreatePackageRepository(newPackageRepo); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to create package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace))
		return
	}

	if o.Wait {
		if err = p.waitForResourceInstallation(o.RepositoryName, o.Namespace, o.PollInterval, o.PollTimeout, progress.ProgressMsg, packagedatamodel.ResourceTypePackageRepository); err != nil {
			log.Warning(msgRunPackageRepositoryUpdate)
			return
		}
	}
}

// newPackageRepository creates a new instance of the PackageRepository object
// If tag is empty, use tagSelection field to select the latest release tag
func (p *pkgClient) newPackageRepository(repositoryName, repositoryImg, namespace string) (*kappipkg.PackageRepository, error) {
	pkgr := &kappipkg.PackageRepository{
		TypeMeta:   metav1.TypeMeta{APIVersion: packagedatamodel.DefaultAPIVersion, Kind: packagedatamodel.KindPackageRepository},
		ObjectMeta: metav1.ObjectMeta{Name: repositoryName, Namespace: namespace},
		Spec: kappipkg.PackageRepositorySpec{Fetch: &kappipkg.PackageRepositoryFetch{
			ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{Image: repositoryImg},
		}},
	}

	_, tag, err := ParseRegistryImageURL(repositoryImg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse OCI registry URL")
	}

	if tag == "" {
		pkgr.Spec.Fetch.ImgpkgBundle.TagSelection = &versions.VersionSelection{
			Semver: &versions.VersionSelectionSemver{
				Constraints: packagedatamodel.DefaultRepositoryImageTagConstraint,
			},
		}
	}
	return pkgr, nil
}

// validateRepositoryAdd ensures that another repository (with the same name or same OCI registry URL) does not already exist in the cluster
func (p *pkgClient) validateRepositoryAdd(repositoryName, repositoryImg, namespace string) error {
	repositoryList, err := p.kappClient.ListPackageRepositories(namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list package repositories")
	}

	for _, repository := range repositoryList.Items { //nolint:gocritic
		if repository.Name == repositoryName {
			return fmt.Errorf("package repository name '%s' already exists in namespace '%s'", repositoryName, namespace)
		}

		if repository.Spec.Fetch != nil && repository.Spec.Fetch.ImgpkgBundle != nil &&
			repository.Spec.Fetch.ImgpkgBundle.Image == repositoryImg {
			return fmt.Errorf("package repository URL '%s' already exists in namespace '%s'", repositoryImg, namespace)
		}
	}

	return nil
}
