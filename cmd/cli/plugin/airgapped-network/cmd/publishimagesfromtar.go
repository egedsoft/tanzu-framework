// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type PublishImagesFromTarOptions struct {
	TkgTarFilePath             string
	CustomImageRepoCertificate string
	PkgClient                  ImgpkgClient
}

var pushImage = &PublishImagesFromTarOptions{}

var PublishImagesfromtarCmd = &cobra.Command{
	Use:          "publish-image-from-tar",
	Short:        "Copy images from tar files to private repo",
	RunE:         publishImagesFromTar,
	SilenceUsage: true,
}

func init() {
	PublishImagesfromtarCmd.Flags().StringVarP(&pushImage.TkgTarFilePath, "tkgTarFilePath", "", "", "Path to the folder where the downloaded tar files are in the disk")
	PublishImagesfromtarCmd.Flags().StringVarP(&pushImage.CustomImageRepoCertificate, "customRepoCertificate", "", "", "custom repo certificate")
}

func (p *PublishImagesFromTarOptions) PushImageToRepo() error {
	yamlFile := filepath.Join(p.TkgTarFilePath, "publish-images-fromtar.yaml")
	yfile, err := os.ReadFile(yamlFile)
	if err != nil {
		return errors.Wrapf(err, "Error while reading %s file", yamlFile)
	}

	data := make(map[string]string)
	err = yaml.Unmarshal(yfile, &data)

	if err != nil {
		return errors.Wrapf(err, "Error while parsing publish-images-fromtar.yaml file")
	}

	for tarfile, path := range data {
		tarfile = filepath.Join(p.TkgTarFilePath, tarfile)
		err = p.PkgClient.CopyImageFromTar(tarfile, path, p.CustomImageRepoCertificate)
		if err != nil {
			return err
		}
	}
	return nil
}

func publishImagesFromTar(cmd *cobra.Command, args []string) error {
	pushImage.PkgClient = &imgpkgClient{}
	err := pushImage.PushImageToRepo()
	if err != nil {
		return err
	}
	return nil
}
