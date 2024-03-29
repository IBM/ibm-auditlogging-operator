//
// Copyright 2021 IBM Corporation
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

package util

import (
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"

	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var DefaultStatusForCR = []string{"none"}
var log = logf.Log.WithName("controller_utils")

func EqualAnnotations(found map[string]string, expected map[string]string) bool {
	logger := log.WithValues("func", "EqualAnnotations")
	if !reflect.DeepEqual(found, expected) {
		logger.Info("Annotations not equal", "Found", found, "Expected", expected)
		return false
	}
	return true
}

func GetImageID(imageRegistry, imageName, envVarName string) string {
	// determine if the image registry has been overridden by the CR
	var imageReg = constant.DefaultImageRegistry
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
		if imageName == constant.DefaultFluentdImageName {
			tag = constant.DefaultFluentdImageTag
		} else if imageName == constant.DefaultJobImageName {
			tag = constant.DefaultJobImageTag
		} else if imageName == constant.DefaultPCImageName {
			tag = constant.DefaultPCImageTag
		} else {
			tag = "latest"
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

// ContainsString returns a Boolean
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

//RemoveString returns a Boolean
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

//IBMDEV
func LabelsForMetadata(name string) map[string]string {
	return map[string]string{"app": name, "app.kubernetes.io/name": name, "app.kubernetes.io/component": constant.AuditLoggingComponentName,
		"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": constant.AuditLoggingReleaseName, "release": constant.AuditLoggingReleaseName,
		constant.AuditTypeLabel: name,
	}
}

//IBMDEV
func LabelsForSelector(name string, crName string) map[string]string {
	return map[string]string{"app": name, "component": constant.AuditLoggingComponentName, constant.AuditLoggingCrType: crName}
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
func AnnotationsForMetering() map[string]string {
	var scc = "restricted"
	annotations := map[string]string{
		"productName":                        constant.ProductName,
		"productID":                          constant.ProductID,
		"productMetric":                      constant.ProductMetric,
		"clusterhealth.ibm.com/dependencies": "cert-manager",
	}
	annotations["openshift.io/scc"] = scc
	return annotations
}

// GetCSNamespace returns the Namespace the operator is running in
func GetCSNamespace() (string, error) {
	var err error
	csNamespace, set := os.LookupEnv(constant.OperatorNamespaceKey)
	if !set {
		err = errors.New("missing ENV variable: " + constant.OperatorNamespaceKey)
	}
	return csNamespace, err
}

func CopyLables(origMap map[string]string) map[string]string {
	targetMap := make(map[string]string)
	certManagerLabel := "certmanager.k8s.io/time-restarted"
	for index, element := range origMap {
		targetMap[index] = element
	}
	delete(targetMap, certManagerLabel)
	return targetMap
}
