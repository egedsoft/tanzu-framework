// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cmd ImgPkgClient defines functions to pull/push/List images
package cmd

type ImgpkgClient interface {
	CopyImageFromTar(sourceImageName string, destImageRepo string, customImageRepoCertificate string) error
	CopyImageToTar(sourceImageName string, destImageRepo string) error
	PullImage(sourceImageName string, destDir string) error
	GetImageTagList(sourceImageName string) []string
}
