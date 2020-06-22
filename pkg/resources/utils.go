//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package resources

import (
	"os"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const AuditLoggingComponentName = "common-audit-logging"
const auditLoggingReleaseName = "common-audit-logging"
const auditLoggingCrType = "auditlogging_cr"
const productName = "IBM Cloud Platform Common Services"
const productID = "068a62892a1e4db39641342e592daa25"
const productVersion = "3.4.0"
const productMetric = "FREE"

const InstanceNamespace = "ibm-common-services"

var DefaultStatusForCR = []string{"none"}
var log = logf.Log.WithName("controller_auditlogging")

func EqualLabels(found map[string]string, expected map[string]string) bool {
	logger := log.WithValues("func", "EqualLabels")
	if !reflect.DeepEqual(found, expected) {
		logger.Info("Labels not equal", "Found", found, "Expected", expected)
		return false
	}
	return true
}

func EqualAnnotations(found map[string]string, expected map[string]string) bool {
	logger := log.WithValues("func", "EqualAnnotations")
	if !reflect.DeepEqual(found, expected) {
		logger.Info("Annotations not equal", "Found", found, "Expected", expected)
		return false
	}
	return true
}

func getImageID(imageRegistry, imageName, envVarName string) string {
	// determine if the image registry has been overridden by the CR
	var imageReg = DefaultImageRegistry
	var imageID string
	if imageRegistry != "" {
		if string(imageRegistry[len(imageRegistry)-1]) != "/" {
			imageRegistry += "/"
		}
		imageReg = imageRegistry
	}
	// determine if an image SHA or tag has been set in an env var.
	// if not, use the default tag (mainly used during development).
	imageTagOrSHA := os.Getenv(envVarName)
	if len(imageTagOrSHA) > 0 {
		// use the value from the env var to build the image ID.
		// a SHA value looks like "sha256:nnnn".
		// a tag value looks like "3.5.0".
		if strings.HasPrefix(imageTagOrSHA, "sha256:") {
			// use the SHA value
			imageID = imageReg + imageName + "@" + imageTagOrSHA
		} else {
			// use the tag value
			imageID = imageReg + imageName + ":" + imageTagOrSHA
		}
	} else {
		var tag string
		if imageName == DefaultPCImageName {
			tag = defaultPCImageTag
		} else {
			tag = defaultFluentdImageTag
		}
		// use the default tag to build the image ID
		imageID = imageReg + imageName + ":" + tag
	}
	return imageID
}

// GetPodNames returns the pod names of the array of pods passed in
func GetPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

//IBMDEV
func LabelsForMetadata(name string) map[string]string {
	return map[string]string{"app": name, "app.kubernetes.io/name": name, "app.kubernetes.io/component": AuditLoggingComponentName,
		"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": auditLoggingReleaseName, "release": auditLoggingReleaseName}
}

//IBMDEV
func LabelsForSelector(name string, crName string) map[string]string {
	return map[string]string{"app": name, "component": AuditLoggingComponentName, auditLoggingCrType: crName}
}

//IBMDEV
func LabelsForPodMetadata(deploymentName string, crName string) map[string]string {
	podLabels := LabelsForMetadata(deploymentName)
	selectorLabels := LabelsForSelector(deploymentName, crName)
	for key, value := range selectorLabels {
		podLabels[key] = value
	}
	return podLabels
}

//IBMDEV
func annotationsForMetering(deploymentName string) map[string]string {
	annotations := map[string]string{
		"productName":    productName,
		"productID":      productID,
		"productVersion": productVersion,
		"productMetric":  productMetric,
	}
	if deploymentName == FluentdName {
		annotations["seccomp.security.alpha.kubernetes.io/pod"] = "docker/default"
		annotations["clusterhealth.ibm.com/dependencies"] = "cert-manager"
		annotations["openshift.io/scc"] = "privileged"
	} else {
		annotations["openshift.io/scc"] = "restricted"
	}
	return annotations
}
