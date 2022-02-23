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
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	const (
		testNamespace = "test-ns"
		testString    = "test"
	)
	Context("Equal Annotations", func() {
		It("Should report whether found annotations are equal to expected annotations", func() {
			foundAnnotations := map[string]string{
				"app": "name",
			}
			expectedAnnotations := map[string]string{
				"app": "name",
			}
			result := EqualAnnotations(foundAnnotations, expectedAnnotations)
			Expect(result).Should(BeTrue())
		})
	})
	Context("Get Image ID", func() {
		It("Should return concatenated image with proper tag or SHA", func() {
			imageRegistry := "quay.io/test"
			imageName := "test-image"
			envVarName := "TEST_IMG_SHA"
			testSHA := "sha256:a1a1a1a15d44814ccabb21545673e9424b1c34449e8936182d8c1f416297b9a7"
			expectedResult := imageRegistry + "/" + imageName + "@" + testSHA
			err := os.Setenv(envVarName, testSHA)
			Expect(err).ToNot(HaveOccurred())

			result := GetImageID(imageRegistry, imageName, envVarName)
			Expect(result).Should(Equal(expectedResult))

			imageName = "fluentd"
			envVarName = "NONE"
			expectedResult = imageRegistry + "/" + imageName + ":" + constant.DefaultFluentdImageTag
			result = GetImageID(imageRegistry, imageName, envVarName)
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Get Pod Names", func() {
		It("Should return the pod names of the array of pods passed in", func() {
			pod1Name := "test1"
			pod2Name := "test2"
			pods := []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pod1Name,
						Namespace: testNamespace,
					},
					Spec: corev1.PodSpec{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pod2Name,
						Namespace: testNamespace,
					},
					Spec: corev1.PodSpec{},
				},
			}
			expectedResult := []string{pod1Name, pod2Name}
			result := GetPodNames(pods)
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Contains string", func() {
		It("Should report whether a slice contains a certain string", func() {
			slice := []string{"a", "b", testString, "c"}
			result := ContainsString(slice, testString)
			Expect(result).Should(BeTrue())
		})
	})
	Context("Remove string", func() {
		It("Should remove a string from a slice", func() {
			searchFor := testString
			slice := []string{"a", "b", testString, "c"}
			expectedResult := []string{"a", "b", "c"}
			result := RemoveString(slice, searchFor)
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Labels for Metadata", func() {
		It("Should return labels with the operand name", func() {
			expectedResult := map[string]string{"app": testString, "app.kubernetes.io/name": testString,
				"app.kubernetes.io/component": constant.AuditLoggingComponentName, "app.kubernetes.io/managed-by": "operator",
				"app.kubernetes.io/instance": constant.AuditLoggingReleaseName, "release": constant.AuditLoggingReleaseName,
				constant.AuditTypeLabel: testString}
			result := LabelsForMetadata(testString)
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Labels for Selector", func() {
		It("Should return labels with the operand name", func() {
			crName := "test-cr"
			expectedResult := map[string]string{"app": testString, "component": constant.AuditLoggingComponentName,
				constant.AuditLoggingCrType: crName}
			result := LabelsForSelector(testString, crName)
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Labels for Pod Metadata", func() {
		It("Should return Metadata and Selector labels with the operand name", func() {
			crName := "test-cr"
			expectedResult := map[string]string{"app": testString, "app.kubernetes.io/name": testString,
				"app.kubernetes.io/component": constant.AuditLoggingComponentName, "app.kubernetes.io/managed-by": "operator",
				"app.kubernetes.io/instance": constant.AuditLoggingReleaseName, "release": constant.AuditLoggingReleaseName,
				constant.AuditTypeLabel: testString, "component": constant.AuditLoggingComponentName, constant.AuditLoggingCrType: crName}
			result := LabelsForPodMetadata(testString, crName)
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Annotations For Metering", func() {
		It("Should return product annotations", func() {
			expectedResult := map[string]string{
				"productName":                        constant.ProductName,
				"productID":                          constant.ProductID,
				"productMetric":                      constant.ProductMetric,
				"clusterhealth.ibm.com/dependencies": "cert-manager",
				"openshift.io/scc":                   "restricted",
			}
			result := AnnotationsForMetering()
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Get CS Namespace", func() {
		It("Should return the namespace the operator is running in", func() {
			Expect(os.Setenv(constant.OperatorNamespaceKey, testNamespace)).Should(Succeed())
			result, err := GetCSNamespace()
			Expect(result).Should(Equal(testNamespace))
			Expect(err).Should(BeNil())
		})
	})
})
