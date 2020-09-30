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

package controllers

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"

	rbacv1 "k8s.io/api/rbac/v1"

	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

	corev1 "k8s.io/api/core/v1"

	res "github.com/IBM/ibm-auditlogging-operator/controllers/resources"

	opversion "github.com/IBM/ibm-auditlogging-operator/version"

	"k8s.io/apimachinery/pkg/types"

	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testdata "github.com/IBM/ibm-auditlogging-operator/controllers/testutil"
)

var _ = Describe("CommonAudit controller", func() {
	const requestName = "example-commonaudit"
	const namespace = "test"
	var (
		ctx              context.Context
		requestNamespace string
		commonAudit      *operatorv1.CommonAudit
		namespacedName   types.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.Background()
		requestNamespace = createNSName(namespace)
		By("Creating the Namespace")
		Expect(k8sClient.Create(ctx, testdata.NamespaceObj(requestNamespace))).Should(Succeed())

		commonAudit = testdata.CommonAuditObj(requestName, requestNamespace)
		namespacedName = types.NamespacedName{Name: requestName, Namespace: requestNamespace}
		By("Creating a new CommonAudit")
		Expect(k8sClient.Create(ctx, commonAudit)).Should(Succeed())
	})

	AfterEach(func() {
		By("Deleting the CommonAudit")
		Expect(k8sClient.Delete(ctx, commonAudit)).Should(Succeed())
		By("Deleting the Namespace")
		Expect(k8sClient.Delete(ctx, testdata.NamespaceObj(requestNamespace))).Should(Succeed())
	})

	Context("When creating a CommonAudit instance", func() {
		It("Should create all secondary resources", func() {
			createdCommonAudit := &operatorv1.CommonAudit{}
			Eventually(func() error {
				return k8sClient.Get(ctx, namespacedName, createdCommonAudit)
			}, timeout, interval).Should(BeNil())

			By("Check status of CommonAudit")
			caInstance := &operatorv1.CommonAudit{}
			Eventually(func() string {
				Expect(k8sClient.Get(ctx, namespacedName, caInstance)).Should(Succeed())
				return caInstance.Status.Versions.Reconciled
			}, timeout, interval).Should(Equal(opversion.Version))

			By("Check if ConfigMaps were created")
			foundCM := &corev1.ConfigMap{}
			for _, cm := range res.FluentdConfigMaps {
				Eventually(func() error {
					return k8sClient.Get(ctx, types.NamespacedName{Name: cm, Namespace: requestNamespace}, foundCM)
				}, timeout, interval).Should(Succeed())
			}

			By("Check if Certificate was created")
			foundCert := &certmgr.Certificate{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.AuditLoggingHTTPSCertName, Namespace: requestNamespace}, foundCert)
			}, timeout, interval).Should(Succeed())

			By("Check if Secret was created")
			foundSecret := &corev1.Secret{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.AuditLoggingClientCertSecName, Namespace: requestNamespace}, foundSecret)
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

			By("Check if Deployment was created")
			foundDeploy := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.FluentdDeploymentName, Namespace: requestNamespace}, foundDeploy)
			}, timeout, interval).Should(Succeed())
		})
	})
	Context("When updating the CommonAudit CR", func() {
		It("Should update secondary resources", func() {
			ca := &operatorv1.CommonAudit{}
			Eventually(func() error {
				return k8sClient.Get(ctx, namespacedName, ca)
			}, timeout, interval).Should(Succeed())

			By("Updating fields in CommonAudit CR")
			// Spec
			ca.Spec.EnableAuditLoggingForwarding = testdata.Forwarding
			ca.Spec.ClusterIssuer = testdata.ClusterIssuer
			ca.Spec.Replicas = testdata.Replicas

			// Fluentd
			ca.Spec.Fluentd.ImageRegistry = testdata.ImageRegistry
			ca.Spec.Fluentd.PullPolicy = testdata.PullPolicy
			ca.Spec.Fluentd.Resources = testdata.Resources

			// Outputs
			ca.Spec.Outputs.Splunk.Host = testdata.SplunkHost
			ca.Spec.Outputs.Splunk.Token = testdata.SplunkToken
			ca.Spec.Outputs.Splunk.Port = testdata.SplunkPort
			ca.Spec.Outputs.Splunk.TLS = testdata.SplunkTLS
			ca.Spec.Outputs.Splunk.EnableSIEM = testdata.SplunkEnable
			ca.Spec.Outputs.Syslog.Host = testdata.QRadarHost
			ca.Spec.Outputs.Syslog.Hostname = testdata.QRadarHostname
			ca.Spec.Outputs.Syslog.Port = testdata.QRadarPort
			ca.Spec.Outputs.Syslog.TLS = testdata.QRadarTLS
			ca.Spec.Outputs.Syslog.EnableSIEM = testdata.QRadarEnable
			ca.Spec.Outputs.HostAliases = append(ca.Spec.Outputs.HostAliases, operatorv1.CommonAuditSpecHostAliases{
				HostIP: testdata.SplunkIP, Hostnames: []string{testdata.SplunkHost},
			})
			ca.Spec.Outputs.HostAliases = append(ca.Spec.Outputs.HostAliases, operatorv1.CommonAuditSpecHostAliases{
				HostIP: testdata.QRadarIP, Hostnames: []string{testdata.QRadarHost},
			})
			Eventually(func() error {
				return k8sClient.Update(ctx, ca)
			}, timeout, interval).Should(Succeed())

			By("Checking ConfigMaps are updated")
			splunkCM := &corev1.ConfigMap{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.FluentdDaemonSetName + "-" +
					res.SplunkConfigName, Namespace: requestNamespace}, splunkCM)
			}, timeout, interval).Should(Succeed())
			splunkData := splunkCM.Data[res.SplunkConfigKey]

			host := testdata.GetFluentdConfig(res.RegexHecHost, splunkData)
			Expect(host).Should(Equal(testdata.SplunkHost))

			port := testdata.GetFluentdConfig(res.RegexHecPort, splunkData)
			Expect(port).Should(Equal(strconv.Itoa(testdata.SplunkPort)))

			token := testdata.GetFluentdConfig(res.RegexHecToken, splunkData)
			Expect(token).Should(Equal(testdata.SplunkToken))

			tls := testdata.GetFluentdConfig(res.RegexProtocol, splunkData)
			Expect(tls).Should(Equal(res.Protocols[testdata.SplunkTLS]))

			qRadarCM := &corev1.ConfigMap{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.FluentdDaemonSetName + "-" +
					res.QRadarConfigName, Namespace: requestNamespace}, qRadarCM)
			}, timeout, interval).Should(Succeed())
			qRadarData := qRadarCM.Data[res.QRadarConfigKey]

			host = testdata.GetFluentdConfig(res.RegexHost, qRadarData)
			Expect(host).Should(Equal(testdata.QRadarHost))

			port = testdata.GetFluentdConfig(res.RegexPort, qRadarData)
			Expect(port).Should(Equal(strconv.Itoa(testdata.QRadarPort)))

			hostname := testdata.GetFluentdConfig(res.RegexHostname, qRadarData)
			Expect(hostname).Should(Equal(testdata.QRadarHostname))

			tls = testdata.GetFluentdConfig(res.RegexTLS, qRadarData)
			Expect(tls).Should(Equal(strconv.FormatBool(testdata.QRadarTLS)))

			By("Checking Deployment is updated")
			deploy := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: res.FluentdDeploymentName, Namespace: requestNamespace}, deploy)
			}, timeout, interval).Should(Succeed())
			Expect(deploy.Spec.Template.Spec.HostAliases).Should(Equal(testdata.HostAliases))
		})
	})

	Context("When deleting secondary resources", func() {
		It("Should recreate them", func() {
			By("Deleting ConfigMaps")
			foundCM := &corev1.ConfigMap{}
			for _, cm := range res.FluentdConfigMaps {
				Eventually(func() error {
					return k8sClient.Get(ctx, types.NamespacedName{Name: cm, Namespace: requestNamespace}, foundCM)
				}, timeout, interval).Should(Succeed())
				Expect(k8sClient.Delete(ctx, foundCM)).Should(Succeed())
			}
			By("Getting ConfigMaps")
			for _, cm := range res.FluentdConfigMaps {
				Eventually(func() error {
					return k8sClient.Get(ctx, types.NamespacedName{Name: cm, Namespace: requestNamespace}, foundCM)
				}, timeout, interval).Should(Succeed())
			}
		})
	})
})
