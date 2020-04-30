#
# Copyright 2020 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Specify whether this repo is build locally or not, default values is '1';
# If set to 1, then you need to also set 'DOCKER_USERNAME' and 'DOCKER_PASSWORD'
# environment variables before build the repo.
BUILD_LOCALLY ?= 1

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overridding the IMG and REGISTRY environment variable.
# IBMDEV Set image and repo
IMAGE_NAME ?= ibm-auditlogging-operator
IMAGE_REPO ?= quay.io/opencloudio
CSV_VERSION ?= 3.6.0

# The namespcethat operator will be deployed in
NAMESPACE=ibm-common-services

# Github host to use for checking the source tree;
# Override this variable ue with your own value if you're working on forked repo.
GIT_HOST ?= github.com/IBM

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))

# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT := $(GOPATH)/bin
export GOBIN ?= $(GOBIN_DEFAULT)
TESTARGS_DEFAULT := "-v"
export TESTARGS ?= $(TESTARGS_DEFAULT)
DEST := $(GOPATH)/src/$(GIT_HOST)/$(BASE_DIR)
VERSION ?= $(shell git describe --exact-match 2> /dev/null || \
                 git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
    TARGET_OS ?= linux
    XARGS_FLAGS="-r"
else ifeq ($(LOCAL_OS),Darwin)
    TARGET_OS ?= darwin
    XARGS_FLAGS=
else
    $(error "This system's OS $(LOCAL_OS) isn't recognized/supported")
endif

ARCH := $(shell uname -m)
LOCAL_ARCH := "amd64"
ifeq ($(ARCH),x86_64)
    LOCAL_ARCH="amd64"
else ifeq ($(ARCH),ppc64le)
    LOCAL_ARCH="ppc64le"
else ifeq ($(ARCH),s390x)
    LOCAL_ARCH="s390x"
else
    $(error "This system's ARCH $(ARCH) isn't recognized/supported")
endif


all: fmt check test coverage build images

#IBMDEV If not using a set GOPATH env var, this check fails
ifeq (,$(wildcard go.mod))
	ifneq ("$(realpath $(DEST))", "$(realpath $(PWD))")
			$(error Please run 'make' from $(DEST). Current directory is $(PWD))
	endif
endif

include common/Makefile.common.mk

############################################################
# work section
############################################################
$(GOBIN):
	@echo "create gobin"
	@mkdir -p $(GOBIN)

work: $(GOBIN)

############################################################
# format section
############################################################

# All available format: format-go format-python
# Default value will run all formats, override these make target with your requirements:
#    eg: fmt: format-go format-protos
fmt: format-go format-python

############################################################
# check section
############################################################

check: lint

# All available linters: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-python lint-helm lint-markdown
# Default value will run all linters, override these make target with your requirements:
#    eg: lint: lint-go lint-yaml
# The MARKDOWN_LINT_WHITELIST variable can be set with comma separated urls you want to whitelist
lint: lint-all

############################################################
# test section
############################################################

test: ## Run unit test
	@echo "Running the tests for $(IMAGE_NAME) on $(LOCAL_ARCH)..."
	@go test $(TESTARGS) ./pkg/controller/...

test-e2e: ## Run integration e2e tests
	@echo ... Running e2e tests ...
	@echo ... Running locally ...
	- operator-sdk test local ./test/e2e --verbose --up-local --namespace=${NAMESPACE}

############################################################
# coverage section
############################################################

coverage: ## Run code coverage test
	@common/scripts/codecov.sh ${BUILD_LOCALLY} "pkg/controller"

############################################################
# install operator sdk section
############################################################

install-operator-sdk:
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then ./common/scripts/install-operator-sdk.sh; fi

############################################################
# build section
############################################################

build: build-amd64 build-ppc64le build-s390x

build-amd64:
	@echo "Building the ${IMAGE_NAME} amd64 binary..."
	@GOARCH=amd64 common/scripts/gobuild.sh build/_output/bin/$(IMAGE_NAME) ./cmd/manager

build-ppc64le:
	@echo "Building the ${IMAGE_NAME} ppc64le binary..."
	@GOARCH=ppc64le common/scripts/gobuild.sh build/_output/bin/$(IMAGE_NAME)-ppc64le ./cmd/manager

build-s390x:
	@echo "Building the ${IMAGE_NAME} s390x binary..."
	@GOARCH=s390x common/scripts/gobuild.sh build/_output/bin/$(IMAGE_NAME)-s390x ./cmd/manager

############################################################
# images section
############################################################

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
endif

build-image-amd64: build-amd64
	@docker build -t $(IMAGE_REPO)/$(IMAGE_NAME)-amd64:$(VERSION) -f build/Dockerfile .

build-image-ppc64le: build-ppc64le
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(IMAGE_REPO)/$(IMAGE_NAME)-ppc64le:$(VERSION) -f build/Dockerfile.ppc64le .

build-image-s390x: build-s390x
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(IMAGE_REPO)/$(IMAGE_NAME)-s390x:$(VERSION) -f build/Dockerfile.s390x .

push-image-amd64: $(CONFIG_DOCKER_TARGET) build-image-amd64
	@docker push $(IMAGE_REPO)/$(IMAGE_NAME)-amd64:$(VERSION)

push-image-ppc64le: $(CONFIG_DOCKER_TARGET) build-image-ppc64le
	@docker push $(IMAGE_REPO)/$(IMAGE_NAME)-ppc64le:$(VERSION)

push-image-s390x: $(CONFIG_DOCKER_TARGET) build-image-s390x
	@docker push $(IMAGE_REPO)/$(IMAGE_NAME)-s390x:$(VERSION)

############################################################
# multiarch-image section
############################################################

images: push-image-amd64 push-image-ppc64le push-image-s390x multiarch-image

multiarch-image:
	@curl -L -o /tmp/manifest-tool https://github.com/estesp/manifest-tool/releases/download/v1.0.0/manifest-tool-linux-amd64
	@chmod +x /tmp/manifest-tool
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(IMAGE_REPO)/$(IMAGE_NAME)-ARCH:$(VERSION) --target $(IMAGE_REPO)/$(IMAGE_NAME) --ignore-missing
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(IMAGE_REPO)/$(IMAGE_NAME)-ARCH:$(VERSION) --target $(IMAGE_REPO)/$(IMAGE_NAME):$(VERSION) --ignore-missing


############################################################
# local testing section
############################################################

install: ## Install all resources (CR/CRD's, RBCA and Operator)
	@echo ....... Installing .......
	@echo ....... Applying CRD and Operator .......
	- kubectl apply -f deploy/crds/operator.ibm.com_auditloggings_crd.yaml
	@echo ....... Applying RBAC .......
	- kubectl apply -f deploy/service_account.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role_binding.yaml -n ${NAMESPACE}
	@echo ....... Applying Operator .......
	- kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}
	@echo ....... Creating AuditLogging CR .......
	- kubectl apply -f deploy/crds/operator.ibm.com_v1alpha1_auditlogging_cr.yaml -n ${NAMESPACE}

uninstall: ## Uninstall all that all performed in the $ make install
	@echo ....... Uninstalling .......
	@echo ....... Deleting CRs .......
	- kubectl get auditpolicy -n ${NAMESPACE} | grep audit | awk '{print $$1}' | xargs kubectl delete auditpolicy -n ${NAMESPACE}
	- kubectl delete -f deploy/crds/operator.ibm.com_v1alpha1_auditlogging_cr.yaml -n ${NAMESPACE}
	@echo ....... Deleting Operator .......
	- kubectl delete -f deploy/operator.yaml -n ${NAMESPACE}
	@echo ....... Deleting CRD.......
	- kubectl delete -f deploy/crds/operator.ibm.com_auditloggings_crd.yaml
	@echo ....... Deleting Rules and Service Account .......
	- kubectl delete -f deploy/role_binding.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/role.yaml -n ${NAMESPACE}

############################################################
# CSV section
############################################################
csv: ## Push CSV package to the catalog
	@RELEASE=${CSV_VERSION} common/scripts/push-csv.sh

############################################################
# clean section
############################################################
clean:
	@rm -rf build/_output

.PHONY: all work fmt check coverage lint test build image images multiarch-image clean
