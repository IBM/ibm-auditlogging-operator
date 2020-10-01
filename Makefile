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

.DEFAULT_GOAL:=help

# Specify whether this repo is build locally or not, default values is '1';
# If set to 1, then you need to also set 'DOCKER_USERNAME' and 'DOCKER_PASSWORD'
# environment variables before build the repo.
BUILD_LOCALLY ?= 1

VCS_URL ?= https://github.com/IBM/ibm-auditlogging-operator
VCS_REF ?= $(shell git rev-parse HEAD)
VERSION ?= $(shell git describe --exact-match 2> /dev/null || \
                git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)
RELEASE_VERSION ?= $(shell cat ./version/version.go | grep "Version =" | awk '{ print $$3}' | tr -d '"')
LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
    TARGET_OS ?= linux
    XARGS_FLAGS="-r"
	STRIP_FLAGS=
else ifeq ($(LOCAL_OS),Darwin)
    TARGET_OS ?= darwin
    XARGS_FLAGS=
	STRIP_FLAGS="-x"
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

# Default image repo
IMAGE_REPO ?= quay.io/opencloudio

ifeq ($(BUILD_LOCALLY),0)
REGISTRY ?= "hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom"
else
REGISTRY ?= "hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com/ibmcom"
endif

# Current Operator image name
OPERATOR_IMAGE_NAME ?= ibm-auditlogging-operator
# Current Operator bundle image name
BUNDLE_IMAGE_NAME ?= ibm-auditlogging-operator-bundle
# Current Operator version
OPERATOR_VERSION ?= 3.7.2

# Image URL to use all building/pushing image targets
IMG ?= $(OPERATOR_IMAGE_NAME):latest

# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
    export CONFIG_DOCKER_TARGET_QUAY = config-docker-quay
endif

include common/Makefile.common.mk

##@ Development

install-controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
endif

install-kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
endif

# find or download kubebuilder
# download kubebuilder if necessary
kube-builder:
ifeq (, $(wildcard /usr/local/kubebuilder))
	@./common/scripts/install-kubebuilder.sh
endif


install-operator-sdk:
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then ./common/scripts/install-operator-sdk.sh; fi

check: lint-all ## Check all files lint error

code-dev: ## Run the default dev commands which are the go tidy, fmt, vet then execute the $ make code-gen
	@echo Running the common required commands for developments purposes
	- make bundle
	- make code-tidy
	- make code-fmt
	@echo Running the common required commands for code delivery
	- make check
	- make test

manager: generate code-fmt code-vet ## Build manager binary
	go build -o bin/manager main.go

run: generate code-fmt code-vet manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	WATCH_NAMESPACE="" go run ./main.go

install: manifests  ## Install CRDs into a cluster
	kustomize build config/crd | kubectl apply -f -

uninstall: manifests ## Uninstall CRDs from a cluster
	kustomize build config/crd | kubectl delete -f -

deploy: manifests ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && kustomize edit set image controller=$(IMAGE_REPO)/$(OPERATOR_IMAGE_NAME):$(OPERATOR_VERSION)
	kustomize build config/default | kubectl apply -f -

##@ Generate code and manifests

manifests: ## Generate manifests e.g. CRD, RBAC etc.
	controller-gen $(CRD_OPTIONS) rbac:roleName=ibm-auditlogging-operator webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: ## Generate code e.g. API etc.
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

package-manifests:
	@{ \
	PKG_MANIFESTS_DIR=deploy/olm-catalog/ibm-auditlogging-operator/$(OPERATOR_VERSION) ;\
	BUNDLE_MANIFESTS_DIR=bundle/manifests ;\
	[[ ! -d $$PKG_MANIFESTS_DIR ]] && mkdir $$PKG_MANIFESTS_DIR ;\
	cp $$BUNDLE_MANIFESTS_DIR/ibm-auditlogging-operator.clusterserviceversion.yaml $$PKG_MANIFESTS_DIR/ibm-auditlogging-operator.v$(OPERATOR_VERSION).clusterserviceversion.yaml ;\
	cp $$BUNDLE_MANIFESTS_DIR/operator.ibm.com_auditloggings.yaml $$PKG_MANIFESTS_DIR/operator.ibm.com_auditloggins_crd.yaml ;\
	cp $$BUNDLE_MANIFESTS_DIR/operator.ibm.com_commonaudits.yaml $$PKG_MANIFESTS_DIR/operator.ibm.com_commonaudits_crd.yaml ;\
	}

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests
	operator-sdk generate kustomize manifests -q
	cd config/manager && kustomize edit set image controller=$(IMG)
	kustomize build config/manifests | operator-sdk generate bundle -q --overwrite --version $(OPERATOR_VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

##@ Test
test: ## Run unit test on prow
	@rm -rf crds
	- make find-certmgr-crds
	@echo "Running unit tests for the controllers."
	@go test ./controllers/... -coverprofile cover.out
	@rm -rf crds

scorecard: operator-sdk ## Run scorecard test
	@echo ... Running the scorecard test
	- operator-sdk scorecard --verbose

##@ Coverage
coverage: ## Run code coverage test
	@common/scripts/codecov.sh ${BUILD_LOCALLY} "pkg/controller"

unit-test: generate code-fmt code-vet manifests ## Run unit test
ifeq (, $(USE_EXISTING_CLUSTER))
	@rm -rf crds
	- make kube-builder
	- make find-certmgr-crds
endif
	@echo "Running unit tests for the controllers."
	@go test ./controllers/... -coverprofile cover.out
	@rm -rf crds


##@ Build

build-operator-image: ## Build the operator image.
	@echo "Building the $(OPERATOR_IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker build -t $(REGISTRY)/$(OPERATOR_IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) \
	--build-arg VCS_REF=$(VCS_REF) --build-arg VCS_URL=$(VCS_URL) \
	--build-arg GOARCH=$(LOCAL_ARCH) -f Dockerfile .

build-bundle-image: ## Build the operator bundle image.
	docker build -f bundle.Dockerfile -t $(IMAGE_REPO)/$(BUNDLE_IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) .

##@ Release

build-push-image: $(CONFIG_DOCKER_TARGET) build-operator-image  ## Build and push the operator images.
	@echo "Pushing the $(OPERATOR_IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker push $(REGISTRY)/$(OPERATOR_IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION)

build-push-bundle-image: $(CONFIG_DOCKER_TARGET_QUAY) build-bundle-image ## Build and push the bundle images.
	@echo "Pushing the $(BUNDLE_IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker push $(IMAGE_REPO)/$(BUNDLE_IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION)

multiarch-image: $(CONFIG_DOCKER_TARGET) ## Generate multiarch images for operator image.
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(REGISTRY) $(OPERATOR_IMAGE_NAME) $(VERSION) $(RELEASE_VERSION)

multiarch-bundle-image: $(CONFIG_DOCKER_TARGET_QUAY) ## Generate multiarch images for bundle image.
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(IMAGE_REPO) $(BUNDLE_IMAGE_NAME) $(VERSION) $(RELEASE_VERSION)

##@ Help
help: ## Display this help
	@echo "Usage:\n  make \033[36m<target>\033[0m"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)