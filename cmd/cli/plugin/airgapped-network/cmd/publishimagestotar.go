// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	tkrv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/pkg/tkr/v1"
)

var totalImgCopiedCounter int

const outputDir = "tmp"

type PublishImagesToTarOptions struct {
	TkgImageRepo    string
	TkgVersion      string
	CustomImageRepo string
	PkgClient       ImgpkgClient
	ImageDetails    map[string]string
}

var pullImage = &PublishImagesToTarOptions{}

var PublishImagestotarCmd = &cobra.Command{
	Use:          "publish-image-to-tar",
	Short:        "Copy images from public repo to tar files",
	RunE:         publishImagesToTar,
	SilenceUsage: true,
}

func init() {
	PublishImagestotarCmd.Flags().StringVarP(&pullImage.TkgImageRepo, "tkgImageRepository", "", "projects.registry.vmware.com/tkg", "TKG public repository")
	PublishImagestotarCmd.Flags().StringVarP(&pullImage.TkgVersion, "tkgVersion", "", "", "TKG version")
	PublishImagestotarCmd.Flags().StringVarP(&pullImage.CustomImageRepo, "customImageRepo", "", "", "custom images repository for airgapped network")
	pullImage.ImageDetails = map[string]string{}
}

func (p *PublishImagesToTarOptions) DownloadTkgCompatibilityImage() error {
	if p.CustomImageRepo == "" || p.TkgVersion == "" {
		return errors.New("input parameter CustomImageRepo or TkgVersion is empty")
	}

	tkgCompatibilityRelativeImagePath := "tkg-compatibility"

	if !isTKGRTMVersion(p.TkgVersion) {
		tkgCompatibilityRelativeImagePath = filepath.Join(p.TkgVersion, tkgCompatibilityRelativeImagePath)
	}
	tkgCompatibilityImagePath := path.Join(p.TkgImageRepo, tkgCompatibilityRelativeImagePath)
	imageTags := p.PkgClient.GetImageTagList(tkgCompatibilityImagePath)
	if len(imageTags) == 0 {
		return errors.New("image doesn't have any tags")
	}
	sourceImageName := tkgCompatibilityImagePath + ":" + imageTags[len(imageTags)-1]
	tarFilename := "tkg-compatibility" + "-" + imageTags[len(imageTags)-1] + ".tar"
	err := p.PkgClient.CopyImageToTar(sourceImageName, tarFilename)
	if err != nil {
		return err
	}
	destRepo := path.Join(p.CustomImageRepo, tkgCompatibilityRelativeImagePath)
	p.ImageDetails[tarFilename] = destRepo
	return nil
}

func (p *PublishImagesToTarOptions) DownloadTkgBomAndComponentImages() (string, error) {
	if p.TkgImageRepo == "" || p.TkgVersion == "" {
		return "", errors.New("input parameter TkgImageRepo or TkgVersion is empty")
	}
	tkgBomImagePath := path.Join(p.TkgImageRepo, "tkg-bom")

	sourceImageName := tkgBomImagePath + ":" + p.TkgVersion
	tarnames := "tkg-bom" + "-" + p.TkgVersion + ".tar"
	destRepo := path.Join(p.CustomImageRepo, tkgBomImagePath)
	p.ImageDetails[tarnames] = destRepo
	err := p.PkgClient.CopyImageToTar(sourceImageName, tarnames)
	if err != nil {
		return "", errors.New("error while downloading tkg-bom")
	}
	err = p.PkgClient.PullImage(sourceImageName, outputDir)
	if err != nil {
		return "", err
	}
	// read the tkg-bom file
	tkgBomFilePath := filepath.Join(outputDir, fmt.Sprintf("tkg-bom-%s.yaml", p.TkgVersion))
	b, err := os.ReadFile(tkgBomFilePath)

	// read the tkg-bom file
	if err != nil {
		return "", errors.Wrapf(err, "read tkg-bom file from %s faild", tkgBomFilePath)
	}
	tkgBom, _ := tkrv1.NewBom(b)
	// imgpkg copy each component's artifacts
	components, _ := tkgBom.Components()
	for _, compInfos := range components {
		for _, compInfo := range compInfos {
			for _, imageInfo := range compInfo.Images {
				sourceImageName = filepath.Join(p.TkgImageRepo, imageInfo.ImagePath) + ":" + imageInfo.Tag
				destImageRepo := path.Join(p.CustomImageRepo, imageInfo.ImagePath)
				imageInfo.ImagePath = replaceSlash(imageInfo.ImagePath)
				tarname := imageInfo.ImagePath + "-" + imageInfo.Tag + ".tar"
				p.ImageDetails[tarname] = destImageRepo
				err := p.PkgClient.CopyImageToTar(sourceImageName, tarname)
				if err != nil {
					return "", err
				}
			}
		}
	}
	return tkgBom.GetCompatibility(), nil
}

func (p *PublishImagesToTarOptions) DownloadTkrCompatibilityImage(tkrCompatibilityRelativeImagePath string) (tkgVersion []string, err error) {
	if p.TkgImageRepo == "" || p.TkgVersion == "" {
		return nil, errors.New("input parameter TkgImageRepo or TkgVersion is empty")
	}

	// get the latest tag of tkr-compatibility image
	tkrCompatibilityImagePath := path.Join(p.TkgImageRepo, tkrCompatibilityRelativeImagePath)
	imageTags := p.PkgClient.GetImageTagList(tkrCompatibilityImagePath)
	if len(imageTags) == 0 {
		return nil, errors.New("image doesn't have any tags")
	}
	// inspect the tkr-compatibility image to get the list of compatible tkrs
	tkrCompatibilityImageURL := tkrCompatibilityImagePath + ":" + imageTags[len(imageTags)-1]

	sourceImageName := tkrCompatibilityImageURL
	err1 := p.PkgClient.PullImage(sourceImageName, outputDir)
	if err1 != nil {
		return nil, err1
	}
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, errors.Wrapf(err, "read directory tmp failed")
	}
	if len(files) != 1 || files[0].IsDir() {
		return nil, fmt.Errorf("tkr-compatibility image should only has exact one file inside")
	}
	tkrCompatibilityFilePath := filepath.Join(outputDir, files[0].Name())
	b, err := os.ReadFile(tkrCompatibilityFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read tkr-compatibility file from %s faild", tkrCompatibilityFilePath)
	}
	tkrCompatibility := &tkrv1.CompatibilityMetadata{}
	if err := yaml.Unmarshal(b, tkrCompatibility); err != nil {
		return nil, errors.Wrapf(err, "Unmarshal tkr-compatibility file %s failed", tkrCompatibilityFilePath)
	}
	// find the corresponding tkg-bom entry
	var tkrVersions []string
	var found = false
	for _, compatibilityInfo := range tkrCompatibility.ManagementClusterVersions {
		if compatibilityInfo.TKGVersion == p.TkgVersion {
			found = true
			tkrVersions = compatibilityInfo.SupportedKubernetesVersions
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("couldn't find the corresponding tkg-bom version in the tkr-compatibility file")
	}
	// imgpkg copy the tkr-compatibility image
	sourceImageName = tkrCompatibilityImageURL
	tarFilename := "tkr-compatibility" + "-" + imageTags[len(imageTags)-1] + ".tar"
	destImageRepo := path.Join(p.CustomImageRepo, tkrCompatibilityRelativeImagePath)
	p.ImageDetails[tarFilename] = destImageRepo
	err = p.PkgClient.CopyImageToTar(sourceImageName, tarFilename)
	if err != nil {
		return nil, err
	}
	return tkrVersions, nil
}

func (p *PublishImagesToTarOptions) DownloadTkrBomAndComponentImages(tkrVersion string) error {
	if p.TkgImageRepo == "" || p.CustomImageRepo == "" {
		return errors.New("input parameter TkgImageRepo or CustomImageRepo is empty")
	}
	tkrTag := underscoredPlus(tkrVersion)
	tkrBomImagePath := path.Join(p.TkgImageRepo, "tkr-bom")
	sourceImageName := tkrBomImagePath + ":" + tkrTag
	tarFilename := "tkr-bom" + "-" + tkrTag + ".tar"
	destImageRepo := path.Join(p.CustomImageRepo, "tkr-bom")
	p.ImageDetails[tarFilename] = destImageRepo
	err := p.PkgClient.CopyImageToTar(sourceImageName, tarFilename)
	if err != nil {
		return err
	}
	sourceImageName = tkrBomImagePath + ":" + tkrTag
	err = p.PkgClient.PullImage(sourceImageName, outputDir)
	if err != nil {
		return err
	}
	// read the tkr-bom file
	tkrBomFilePath := filepath.Join(outputDir, fmt.Sprintf("tkr-bom-%s.yaml", tkrVersion))
	b, err := os.ReadFile(tkrBomFilePath)
	if err != nil {
		return errors.Wrapf(err, "read tkr-bom file from %s faild", tkrBomFilePath)
	}
	tkgBom, _ := tkrv1.NewBom(b)
	// imgpkg copy each component's artifacts
	components, _ := tkgBom.Components()
	for _, compInfos := range components {
		for _, compInfo := range compInfos {
			for _, imageInfo := range compInfo.Images {
				sourceImageName = filepath.Join(p.TkgImageRepo, imageInfo.ImagePath) + ":" + imageInfo.Tag
				destImageRepo := path.Join(p.CustomImageRepo, imageInfo.ImagePath)
				imageInfo.ImagePath = replaceSlash(imageInfo.ImagePath)
				tarname := imageInfo.ImagePath + "-" + imageInfo.Tag + ".tar"
				p.ImageDetails[tarname] = destImageRepo
				err = p.PkgClient.CopyImageToTar(sourceImageName, tarname)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func publishImagesToTar(cmd *cobra.Command, args []string) error {
	pullImage.PkgClient = &imgpkgClient{}
	if pullImage.TkgVersion == "" || !strings.HasPrefix(pullImage.TkgVersion, "v") {
		return fmt.Errorf("invalid TKG Tag %s", pullImage.TkgVersion)
	}
	if pullImage.TkgImageRepo == "" { // TODO : Put more validation
		return fmt.Errorf("invalid TkgImageRepository %s", pullImage.TkgImageRepo)
	}
	if pullImage.CustomImageRepo == "" {
		return fmt.Errorf("invalid customImageRepo %s", pullImage.CustomImageRepo)
	}
	err := pullImage.DownloadTkgCompatibilityImage()
	if err != nil {
		return err
	}
	tkrCompatibilityRelativeImagePath, err := pullImage.DownloadTkgBomAndComponentImages()

	if err != nil {
		return err
	}
	tkrVersions, err := pullImage.DownloadTkrCompatibilityImage(tkrCompatibilityRelativeImagePath)
	if err != nil {
		return errors.Wrapf(err, "Error while retrieving tkrVersions")
	}

	for _, tkrVersion := range tkrVersions {
		err = pullImage.DownloadTkrBomAndComponentImages(tkrVersion)
		if err != nil {
			return err
		}
	}

	data, _ := yaml.Marshal(&pullImage.ImageDetails)
	err2 := os.WriteFile("publish-images-fromtar.yaml", data, 0666)
	if err2 != nil {
		return errors.Wrapf(err2, "Error while writing publish-images-fromtar.yaml file")
	}

	return nil
}
