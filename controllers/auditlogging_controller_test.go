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

package controllers

import (
	"context"
	"os"

	batchv1 "k8s.io/api/batch/v1"

	operatorv1alpha1 "github.com/IBM/ibm-auditlogging-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"

	rbacv1 "k8s.io/api/rbac/v1"

	certmgr "github.com/ibm/ibm-cert-manager-operator/apis/certmanager/v1alpha1"

	corev1 "k8s.io/api/core/v1"

	res "github.com/IBM/ibm-auditlogging-operator/controllers/resources"

	opversion "github.com/IBM/ibm-auditlogging-operator/version"

	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testdata "github.com/IBM/ibm-auditlogging-operator/controllers/testutil"
)

var _ = Describe("AuditLogging controller", func() {
	const requestName = "example-auditlogging"
	var (
		ctx              context.Context
		requestNamespace string
		auditLogging     *operatorv1alpha1.AuditLogging
		namespacedName   types.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.Background()
		requestNamespace = createNSName(namespace)
		By("Creating the Namespace")
		Expect(k8sClient.Create(ctx, testdata.NamespaceObj(requestNamespace))).Should(Succeed())
		Expect(os.Setenv(constant.OperatorNamespaceKey, requestNamespace)).Should(Succeed())

		auditLogging = testdata.AuditLoggingObj(requestName)
		// AuditLogging is cluster scoped and does not have a namespace
		namespacedName = types.NamespacedName{Name: requestName, Namespace: ""}
		By("Creating a new AuditLogging")
		Expect(k8sClient.Create(ctx, auditLogging)).Should(Succeed())
	})

	AfterEach(func() {
		By("Deleting the AuditLogging")
		Expect(k8sClient.Delete(ctx, auditLogging)).Should(Succeed())
		By("Deleting the Namespace")
		Expect(k8sClient.Delete(ctx, testdata.NamespaceObj(requestNamespace))).Should(Succeed())
	})

	Context("When creating an AuditLogging instance", func() {
		It("Should create all secondary resources", func() {
			createdAuditLogging := &operatorv1alpha1.AuditLogging{}
			Eventually(func() error {
				return k8sClient.Get(ctx, namespacedName, createdAuditLogging)
			}, timeout, interval).Should(BeNil())

			By("Check status of AuditLogging")
			audit := &operatorv1alpha1.AuditLogging{}
			Eventually(func() string {
				Expect(k8sClient.Get(ctx, namespacedName, audit)).Should(Succeed())
				return audit.Status.Versions.Reconciled
			}, timeout, interval).Should(Equal(opversion.Version))

			By("Check if Job was created")
			foundJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.JobName, Namespace: requestNamespace}, foundJob)
			}, timeout, interval).Should(Succeed())

			By("Check if Policy Controller deployment was created")
			foundDeploy := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.AuditPolicyControllerDeploy, Namespace: requestNamespace}, foundDeploy)
			}, timeout, interval).Should(Succeed())

			By("Check if ConfigMaps were created")
			foundCM := &corev1.ConfigMap{}
			for _, cm := range res.FluentdConfigMaps {
				Eventually(func() error {
					return k8sClient.Get(ctx, types.NamespacedName{Name: cm, Namespace: requestNamespace}, foundCM)
				}, timeout, interval).Should(Succeed())
			}

			By("Check if Certificates were created")
			foundHTTPSCert := &certmgr.Certificate{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.AuditLoggingHTTPSCertName, Namespace: requestNamespace}, foundHTTPSCert)
			}, timeout, interval).Should(Succeed())

			foundCert := &certmgr.Certificate{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.AuditLoggingCertName, Namespace: requestNamespace}, foundCert)
			}, timeout, interval).Should(Succeed())

			By("Check if SA was created")
			foundSA := &corev1.ServiceAccount{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.OperandServiceAccount, Namespace: requestNamespace}, foundSA)
			}, timeout, interval).Should(Succeed())

			By("Check if Role was created")
			foundRole := &rbacv1.Role{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.FluentdDaemonSetName + "-role", Namespace: requestNamespace}, foundRole)
			}, timeout, interval).Should(Succeed())

			By("Check if RoleBinding was created")
			foundRB := &rbacv1.RoleBinding{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.FluentdDaemonSetName + "-rolebinding", Namespace: requestNamespace}, foundRB)
			}, timeout, interval).Should(Succeed())

			By("Check if Service was created")
			foundSvc := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: constant.AuditLoggingComponentName, Namespace: requestNamespace}, foundSvc)
			}, timeout, interval).Should(Succeed())

			By("Check if DaemonSet was created")
			foundDaemonset := &appsv1.DaemonSet{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.FluentdDaemonSetName, Namespace: requestNamespace}, foundDaemonset)
			}, timeout, interval).Should(Succeed())
		})
	})
})
