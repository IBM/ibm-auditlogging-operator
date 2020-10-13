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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"
	testdata "github.com/IBM/ibm-auditlogging-operator/controllers/testutil"
)

var _ = Describe("ConfigMaps", func() {
	const requestName = "example-commonaudit"
	const requestNamespace = "test"
	var (
		commonAudit *operatorv1.CommonAudit
		//namespacedName = types.NamespacedName{Name: requestName, Namespace: requestNamespace}
	)

	BeforeEach(func() {
		commonAudit = testdata.CommonAuditObj(requestName, requestNamespace)
		commonAudit.Spec.Outputs.Splunk.EnableSIEM = testdata.SplunkEnable
		commonAudit.Spec.Outputs.Syslog.EnableSIEM = testdata.QRadarEnable

		commonAudit.Spec.Outputs.Splunk.Host = testdata.SplunkHost
		commonAudit.Spec.Outputs.Splunk.Port = testdata.SplunkPort
		commonAudit.Spec.Outputs.Splunk.Token = testdata.SplunkToken
		commonAudit.Spec.Outputs.Splunk.TLS = testdata.SplunkTLS

		commonAudit.Spec.Outputs.Syslog.Host = testdata.QRadarHost
		commonAudit.Spec.Outputs.Syslog.Port = testdata.QRadarPort
		commonAudit.Spec.Outputs.Syslog.Hostname = testdata.QRadarHostname
		commonAudit.Spec.Outputs.Syslog.TLS = testdata.QRadarTLS
	})

	Context("Build Fluentd Config", func() {
		It("Should include enabled output plugins", func() {

			result := buildFluentdConfig(commonAudit)
			expectedResult := `
fluent.conf: |-
    # Input plugins (Supports Systemd and HTTP)
    @include /fluentd/etc/source.conf
    # Output plugins (Supports Splunk and Syslog)
    <match icp-audit icp-audit.** syslog syslog.**>
        @type copy
        @include /fluentd/etc/splunkHEC.conf
        @include /fluentd/etc/remoteSyslog.conf
    </match>`
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Build Fluentd Splunk Config", func() {
		It("Should build Splunk configmap with instance host, port, and token", func() {

			result := buildFluentdSplunkConfig(commonAudit)
			expectedResult := `
splunkHEC.conf: |-
    <store>
        @type splunk_hec
        hec_host test-splunk.fyre.ibm.com
        hec_port 8088
        hec_token aaaa
        protocol http
        ca_file /fluentd/etc/tls/splunkCA.pem
        source ${tag}
    </store>`
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Build Fluentd QRadar Config", func() {
		It("Should build Syslog configmap with instance host, port, and hostname", func() {
			result := buildFluentdQRadarConfig(commonAudit)
			expectedResult := `
remoteSyslog.conf: |-
    <store>
        @type remote_syslog
        host test-qradar.fyre.ibm.com
        port 514
        hostname test-syslog
        tls false
        protocol tcp
        ca_file /fluentd/etc/tls/qradar.crt
        packet_size 4096
        program fluentd
        <format>
            @type single_value
            message_key message
        </format>
    </store>`
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Update SIEM Config", func() {
		It("Should update found configmap with instance configs", func() {
			dq := DataQRadar{}
			dataMap := make(map[string]string)
			data := `
remoteSyslog.conf: |-
    <store>
        @type remote_syslog
        host qradar.fyre.ibm.com
        port 614
        hostname test-syslog
        tls true
        protocol tcp
        ca_file /fluentd/etc/tls/qradar.crt
        packet_size 4096
        program fluentd
        <buffer>
            @type file
        </buffer>
        <format>
            @type single_value
            message_key message
        </format>
    </store>`
			err := yaml.Unmarshal([]byte(data), &dq)
			Expect(err).NotTo(HaveOccurred())
			dataMap[QRadarConfigKey] = dq.Value
			foundCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FluentdDaemonSetName + "-" + QRadarConfigName,
					Namespace: requestNamespace,
				},
				Data: dataMap,
			}
			result := UpdateSIEMConfig(commonAudit, foundCM)
			expectedResult := `<store>
    @type remote_syslog
    host test-qradar.fyre.ibm.com
    port 514
    hostname test-syslog
    tls false
    protocol tcp
    ca_file /fluentd/etc/tls/qradar.crt
    packet_size 4096
    program fluentd
    <buffer>
        @type file
    </buffer>
    <format>
        @type single_value
        message_key message
    </format>
</store>`
			Expect(result).Should(Equal(expectedResult))
		})
	})
	Context("Equal SIEM Config", func() {
		It("Should return whether or not instance configs are equal to configmap data", func() {
			dq := DataQRadar{}
			dataMap := make(map[string]string)
			// tls is missing
			data := `
remoteSyslog.conf: |-
    <store>
        @type remote_syslog
        host test-qradar.fyre.ibm.com
        port 514
        hostname test-syslog
        protocol tcp
        ca_file /fluentd/etc/tls/qradar.crt
        packet_size 4096
        program fluentd
        <buffer>
            @type file
        </buffer>
        <format>
            @type single_value
            message_key message
        </format>
    </store>`
			err := yaml.Unmarshal([]byte(data), &dq)
			Expect(err).NotTo(HaveOccurred())
			dataMap[QRadarConfigKey] = dq.Value
			foundCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FluentdDaemonSetName + "-" + QRadarConfigName,
					Namespace: requestNamespace,
				},
				Data: dataMap,
			}
			result, missing := EqualSIEMConfig(commonAudit, foundCM)
			Expect(result).Should(BeFalse())
			Expect(missing).Should(BeTrue())
		})
	})
})
