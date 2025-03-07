# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build tooling that is used for the main project

# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL := /usr/bin/env bash

.DEFAULT_GOAL := help

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq (,$(strip $(GOPROXY)))
GOPROXY := https://proxy.golang.org
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE := on

# Directories.
ROOT_DIR:=$(shell git rev-parse --show-toplevel)
BIN_DIR := bin
TOOLS_SCRIPTS_DIR := $(ROOT_DIR)/pinniped-components/hack/tools
SRCS := go.mod go.sum

GOLANGCI_LINT_VERSION=1.46.0
IMGPKG_VERSION=v0.31.0
KAPP_VERSION=v0.52.0
KBLD_VERSION=v0.31.0
# darwin arm64 available in K8S_VERSION=1.24.1
K8S_VERSION=1.23.5
VALE_VERSION=2.17.0
VENDIR_VERSION=v0.30.0
YQ_VERSION=v4.5.0
YTT_VERSION=v0.42.0

# Host information.
HOST_OS=$(shell go env GOOS)
HOST_ARCH=$(shell go env GOARCH)

# Binaries.
CONTROLLER_GEN := $(BIN_DIR)/controller-gen
GINKGO             := $(BIN_DIR)/ginkgo
GOLANGCI_LINT      := $(BIN_DIR)/golangci-lint
GOIMPORTS          := $(BIN_DIR)/goimports
KUBEBUILDER        := $(BIN_DIR)/kubebuilder
KUSTOMIZE := $(BIN_DIR)/kustomize
GOBINDATA          := $(BIN_DIR)/gobindata
VALE               := $(BIN_DIR)/vale
YTT                := $(BIN_DIR)/ytt
KBLD               := $(BIN_DIR)/kbld
VENDIR             := $(BIN_DIR)/vendir
IMGPKG             := $(BIN_DIR)/imgpkg
YQ                 := $(BIN_DIR)/yq
TRIVY              := $(BIN_DIR)/trivy
KAPP               := $(BIN_DIR)/kapp
CONVERSION_GEN     := $(BIN_DIR)/conversion-gen

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Binaries
## --------------------------------------

controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): go.mod go.sum
	go build -tags=tools -o $@ sigs.k8s.io/controller-tools/cmd/controller-gen

ginkgo: $(GINKGO)
$(GINKGO): go.mod
	mkdir -p $(BIN_DIR)
	go build -tags=tools -o $(BIN_DIR)/ginkgo github.com/onsi/ginkgo/ginkgo

COUNTERFEITER := $(BIN_DIR)/counterfeiter
$(COUNTERFEITER): $(BIN_DIR) go.mod go.sum
	go build -tags=tools -o $@ github.com/maxbrunsfeld/counterfeiter/v6

golangci-lint: $(GOLANGCI_LINT) ## Install golangci-lint
$(GOLANGCI_LINT):
	mkdir -p $(BIN_DIR)
	curl -L https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(HOST_OS)-$(HOST_ARCH).tar.gz | tar -xz -C /tmp/
	mv /tmp/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(HOST_OS)-$(HOST_ARCH)/golangci-lint $(@)

goimports: $(GOIMPORTS) $(SRCS) ## Build goimports
$(GOIMPORTS): go.mod go.sum
	go build -tags=tools -o $@ golang.org/x/tools/cmd/goimports

kubebuilder: $(KUBEBUILDER) ## Install kubebuilder
$(KUBEBUILDER):
	mkdir -p $(BIN_DIR)
	curl -L https://go.kubebuilder.io/test-tools/${K8S_VERSION}/${HOST_OS}/${HOST_ARCH} | tar -xz -C /tmp/
	mv /tmp/kubebuilder $(@)

kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(BIN_DIR) go.mod go.sum # Build kustomize from tools folder.
	CGO_ENABLED=0 go build -tags=tools -o $@ sigs.k8s.io/kustomize/kustomize/v4

gobindata: $(GOBINDATA)
$(GOBINDATA): go.mod go.sum # Build go-bindata
	mkdir -p $(BIN_DIR)
	go build -tags=tools -o $(BIN_DIR) github.com/shuLhan/go-bindata/...; mv $(BIN_DIR)/go-bindata $(GOBINDATA)

vale: $(VALE)
$(VALE):
	mkdir -p $(BIN_DIR)
	# vale uses macOS for darwin, Linux for linux and 64 for amd64, so manually map them
	curl -sfL https://github.com/errata-ai/vale/releases/download/v$(VALE_VERSION)/vale_$(VALE_VERSION)_$(shell echo $(HOST_OS) | sed 's,darwin,macOS,g'| sed 's,linux,Linux,g')_$(shell echo $(HOST_ARCH) | sed 's,amd,,g')-bit.tar.gz | tar -xz -C /tmp/
	mv /tmp/vale $(@)
	chmod a+x $(@)

yq: $(YQ) ## Download yq
$(YQ):
	mkdir -p $(BIN_DIR)
	echo $(HOST_OS) $(HOST_ARCH)
	curl -sL https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$(HOST_OS)_$(HOST_ARCH) -o $(@) && \
	chmod a+x $(@)

conversion-gen: $(CONVERSION_GEN) $(SRCS) ## Build conversion-gen
$(CONVERSION_GEN): go.mod
	mkdir -p $(BIN_DIR)
	go build -tags=tools -o $(BIN_DIR) k8s.io/code-generator/cmd/conversion-gen

## --------------------------------------------------------------------------------------------------------
## Carvel tools
## --------------------------------------------------------------------------------------------------------

ytt: $(YTT) ## Install ytt
$(YTT):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-ytt/releases/download/$(YTT_VERSION)/ytt-$(HOST_OS)-$(HOST_ARCH)
	mv ytt-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

kbld: $(KBLD) ## Download kbld
$(KBLD):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-kbld/releases/download/$(KBLD_VERSION)/kbld-$(HOST_OS)-$(HOST_ARCH)
	mv kbld-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

vendir: $(VENDIR) ## Download vendir
$(VENDIR):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-vendir/releases/download/$(VENDIR_VERSION)/vendir-$(HOST_OS)-$(HOST_ARCH)
	mv vendir-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

imgpkg: $(IMGPKG) ## Download imgpkg
$(IMGPKG):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-imgpkg/releases/download/$(IMGPKG_VERSION)/imgpkg-$(HOST_OS)-$(HOST_ARCH)
	mv imgpkg-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

kapp: $(KAPP) ## Download kapp
$(KAPP):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-kapp/releases/download/$(KAPP_VERSION)/kapp-$(HOST_OS)-$(HOST_ARCH)
	mv kapp-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

## --------------------------------------
## Trivy
## --------------------------------------

trivy: $(TRIVY) ## Download trivy
$(TRIVY):
	curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b $(BIN_DIR)

## --------------------------------------
## Cleanup
## --------------------------------------

.PHONY: clean
clean: ## Remove all generated binaries
	rm -rf bin
