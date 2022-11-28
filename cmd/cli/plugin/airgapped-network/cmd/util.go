// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
type ImagePathInfo struct {
	ImagePath string `yaml:"imagePath"`
}

type BomCompatibilitySection struct {
	Compatibility ImagePathInfo `yaml:"tkr-compatibility"`
}

*/
func printErrorAndExit(err error) {
	fmt.Printf("failed with error %s\n", err.Error())
	os.Exit(1)
}

func isTKGRTMVersion(tag string) bool {
	tag = strings.TrimPrefix(tag, "v")
	versions := strings.Split(tag, ".")
	if len(versions) != 3 {
		return false
	}
	for _, version := range versions {
		if _, err := strconv.Atoi(version); err != nil {
			return false
		}
	}
	return true
}

func underscoredPlus(s string) string {
	return strings.Replace(s, "+", "_", -1)
}

func replaceSlash(s string) string {
	return strings.Replace(s, "/", "-", -1)
}
