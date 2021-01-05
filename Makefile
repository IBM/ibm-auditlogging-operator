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
VERSION ?= $(shell cat ./version/version.go | grep "Version =" | awk '{ print $$3}' | tr -d '"')
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


ifeq ($(BUILD_LOCALLY),0)
IMAGE_REPO ?= "hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom"
else
IMAGE_REPO ?= "quay.io/hbradfield"
endif
OPERAND_REGISTRY ?= $(IMAGE_REPO)

# Current Operator image name
IMAGE_NAME ?= ibm-auditlogging-operator
# Current Operator bundle image name
BUNDLE_IMAGE_NAME ?= ibm-auditlogging-operator-bundle
# Current Operator version
OPERATOR_VERSION ?= $(VERSION)
CSV_VERSION ?= $(OPERATOR_VERSION)

# The namespce that operator and auditlogging will be deployed in
NAMESPACE=ibm-common-services
# The namespce that commonaudit will be deployed in
CA_NAMESPACE = test

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_NAME):latest

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

install-controller-gen: ## Install controller-gen
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

install-kustomize: ## Install kustomize
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


install-operator-sdk: ## Install operator-sdk
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then ./common/scripts/install-operator-sdk.sh; fi

check: lint-all ## Check all files for lint error

code-dev: ## Run the default dev commands (go tidy, fmt, vet) then execute $ make code-gen
	@echo Running the common required commands for developments purposes
	- make code-tidy
	- make code-fmt
	- make code-vet
	@echo Running the common required commands for code delivery
	- make check
	- make test

manager: generate code-fmt code-vet ## Generate code e.g. API etc and build manager binary
	go build -o bin/manager main.go

run: generate code-fmt code-vet manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	WATCH_NAMESPACE="ibm-common-services" go run ./main.go

install: manifests  ## Install CRDs into a cluster
	kustomize build config/crd | kubectl apply -f -

uninstall: manifests ## Uninstall CRDs from a cluster
	kustomize build config/crd | kubectl delete -f -

install-all: ## Install all resources (CR/CRD's, RBCA and Operator)
	@echo ....... Creating namespace .......
	- kubectl create namespace ${CA_NAMESPACE}
	@echo ....... Applying manifests .......
	- kubectl create sa ibm-auditlogging-operator -n ${NAMESPACE}
	- kubectl create sa ibm-audit-policy-controller -n ${NAMESPACE}
	- kubectl create sa ibm-auditlogging-cleanup -n ${NAMESPACE}
	- kubectl create -f config/rbac/role.yaml
	- kubectl create -f config/rbac/role_binding.yaml
	- kubectl create -f config/rbac/leader_election_role.yaml
	- kubectl create -f config/rbac/leader_election_role_binding.yaml
	- for manifest in $(shell ls bundle/manifests/*.yaml); do kubectl apply -f $${manifest} -n ${NAMESPACE}; done
	@echo ....... Creating the Instances .......
# 	- kubectl apply -f config/samples/operator_v1_commonaudit.yaml -n ${CA_NAMESPACE}
	- kubectl apply -f config/samples/operator_v1alpha1_auditlogging.yaml -n ${NAMESPACE}

uninstall-all: ## Uninstall all resources (CR/CRD's, RBCA and Operator)
	@echo ....... Deleting namespace .......
	- kubectl delete namespace ${CA_NAMESPACE}
	@echo ....... Deleting the Instances .......
	- kubectl delete --all commonaudit --all-namespaces
	- kubectl delete --all auditpolicy --all-namespaces
	@echo ....... Deleting manifests .......
	- kubectl delete sa ibm-auditlogging-operator -n ${NAMESPACE}
	- kubectl delete sa ibm-audit-policy-controller -n ${NAMESPACE}
	- kubectl delete sa ibm-auditlogging-cleanup -n ${NAMESPACE}
	- kubectl delete -f config/rbac/role.yaml
	- kubectl delete -f config/rbac/role_binding.yaml
	- kubectl delete -f config/rbac/leader_election_role.yaml
	- kubectl delete -f config/rbac/leader_election_role_binding.yaml
	- for manifest in $(shell ls bundle/manifests/*.yaml); do kubectl delete -f $${manifest} -n ${NAMESPACE}; done

deploy: manifests ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && kustomize edit set image controller=$(IMAGE_REPO)/$(IMAGE_NAME):$(VERSION)
	kustomize build config/default | kubectl apply -f -

##@ Generate code and manifests

manifests: ## Generate manifests e.g. CRD, RBAC etc.
	controller-gen $(CRD_OPTIONS) rbac:roleName=ibm-auditlogging-operator webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: ## Generate code e.g. API etc.
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

bundle-manifests: ## Generate bundle manifests
	kustomize build config/manifests | operator-sdk generate bundle \
	-q --overwrite --version $(CSV_VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

generate-all: manifests ## Generate bundle manifests, metadata and package manifests
	operator-sdk generate kustomize manifests -q
	- make bundle-manifests CHANNELS=beta,stable-v1 DEFAULT_CHANNEL=stable-v1

##@ Test
test: ## Run unit test on prow
	@rm -rf crds
	- make find-certmgr-crds
	@echo "Running unit tests for the controllers."
	@go test -v ./controllers/...
	@rm -rf crds

unit-test: generate code-fmt code-vet manifests ## Run unit test
ifeq (, $(USE_EXISTING_CLUSTER))
	@rm -rf crds
	- make kube-builder
	- make find-certmgr-crds
endif
	@echo "Running unit tests for the controllers."
	@go test -v ./controllers/... -coverprofile cover.out
	@rm -rf crds

scorecard: operator-sdk ## Run scorecard test
	@echo ... Running the scorecard test
	- operator-sdk scorecard --verbose

##@ Coverage
coverage: ## Run code coverage test
	@rm -rf crds
	- make find-certmgr-crds
	@common/scripts/codecov.sh ${BUILD_LOCALLY} "controllers/"
	@rm -rf crds

##@ Build

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
endif

build: ## Build operator binary
	@echo "Building the ibm-auditlogging-operator binary"
	@CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o bin/manager main.go

build-bundle-image: ## Build operator bundle image
	$(eval ARCH := $(shell uname -m|sed 's/x86_64/amd64/'))
	docker build -f bundle.Dockerfile -t $(IMAGE_REPO)/$(BUNDLE_IMAGE_NAME)-$(ARCH):$(VERSION) .

build-image-amd64: ## Build amd64 operator image
	@docker build -t $(IMAGE_REPO)/$(IMAGE_NAME)-amd64:$(VERSION) \
	--build-arg VCS_REF=$(VCS_REF) --build-arg VCS_URL=$(VCS_URL) -f Dockerfile .

build-image-ppc64le: ## Build ppcle64 operator image
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(IMAGE_REPO)/$(IMAGE_NAME)-ppc64le:$(VERSION) \
	--build-arg VCS_REF=$(VCS_REF) --build-arg VCS_URL=$(VCS_URL) -f Dockerfile.ppc64le .

build-image-s390x: ## Build s390x operator image
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(IMAGE_REPO)/$(IMAGE_NAME)-s390x:$(VERSION) \
	--build-arg VCS_REF=$(VCS_REF) --build-arg VCS_URL=$(VCS_URL) -f Dockerfile.s390x .

push-image-amd64: $(CONFIG_DOCKER_TARGET) build-image-amd64 ## Push amd64 operator image
	@docker push $(IMAGE_REPO)/$(IMAGE_NAME)-amd64:$(VERSION)

push-image-ppc64le: $(CONFIG_DOCKER_TARGET) build-image-ppc64le ## Push ppc64le operator image
	@docker push $(IMAGE_REPO)/$(IMAGE_NAME)-ppc64le:$(VERSION)

push-image-s390x: $(CONFIG_DOCKER_TARGET) build-image-s390x ## Push s390x operator image
	@docker push $(IMAGE_REPO)/$(IMAGE_NAME)-s390x:$(VERSION)

push-bundle-image: $(CONFIG_DOCKER_TARGET) build-bundle-image ## Push operator bundle image
	@docker push $(IMAGE_REPO)/$(BUNDLE_IMAGE_NAME)-$(ARCH):$(VERSION)

##@ Release

images: push-image-amd64 push-image-ppc64le push-image-s390x multiarch-image ## Generate all images

multiarch-image: ## Generate multiarch images for operator image
	@curl -L -o /tmp/manifest-tool https://github.com/estesp/manifest-tool/releases/download/v1.0.0/manifest-tool-linux-amd64
	@chmod +x /tmp/manifest-tool
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(IMAGE_REPO)/$(IMAGE_NAME)-ARCH:$(VERSION) --target $(IMAGE_REPO)/$(IMAGE_NAME) --ignore-missing
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(IMAGE_REPO)/$(IMAGE_NAME)-ARCH:$(VERSION) --target $(IMAGE_REPO)/$(IMAGE_NAME):$(VERSION) --ignore-missing

multiarch-bundle-image: ## Generate multiarch images for operator bundle image
	@curl -L -o /tmp/manifest-tool https://github.com/estesp/manifest-tool/releases/download/v1.0.0/manifest-tool-linux-amd64
	@chmod +x /tmp/manifest-tool
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(IMAGE_REPO)/$(BUNDLE_IMAGE_NAME):$(VERSION) --target $(IMAGE_REPO)/$(BUNDLE_IMAGE_NAME) --ignore-missing
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(IMAGE_REPO)/$(IMAGE_NAME)-ARCH:$(VERSION) --target $(IMAGE_REPO)/$(IMAGE_NAME):$(VERSION) --ignore-missing

##@ Help
help: ## Display this help
	@echo "Usage:\n  make \033[36m<target>\033[0m"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
