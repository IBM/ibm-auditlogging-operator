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

package testutil

import (
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"
	operatorv1alpha1 "github.com/IBM/ibm-auditlogging-operator/api/v1alpha1"
	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
)

const (
	Forwarding     = true
	ClusterIssuer  = "test-ca-issuer"
	ImageRegistry  = "test-registry.com/test-repo"
	PullPolicy     = "Always"
	SplunkHost     = "test-splunk.fyre.ibm.com"
	SplunkIP       = "7.7.7.7"
	SplunkPort     = 8088
	SplunkToken    = "aaaa"
	SplunkTLS      = false
	SplunkEnable   = true
	QRadarHost     = "test-qradar.fyre.ibm.com"
	QRadarIP       = "6.6.6.6"
	QRadarPort     = 514
	QRadarHostname = "test-syslog"
	QRadarTLS      = false
	QRadarEnable   = true

	BadPort = 1111
)

const DummySplunkConfig = `
splunkHEC.conf: |-
     <store>
        @type splunk_hec
        hec_host master
        hec_port 8089
        hec_token abc-123
        protocol https
        ca_file /fluentd/etc/tls/splunkCA.pem
        source ${tag}
        <buffer>
          # ...
        </buffer>
     </store>`

var trueVar = true
var falseVar = false
var rootUser = int64(0)
var cpu25 = resource.NewMilliQuantity(300, resource.DecimalSI)         // 300m
var memory100 = resource.NewQuantity(100*1024*1024, resource.BinarySI) // 100Mi
var Resources = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu25,
		corev1.ResourceMemory: *memory100},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu25,
		corev1.ResourceMemory: *memory100},
}
var HostAliases = []corev1.HostAlias{
	{
		IP:        SplunkIP,
		Hostnames: []string{SplunkHost},
	},
	{
		IP:        QRadarIP,
		Hostnames: []string{QRadarHost},
	},
}
var Replicas = int32(3)
var BadPorts = []corev1.ServicePort{
	{
		Name:     constant.AuditLoggingComponentName,
		Protocol: "TCP",
		Port:     BadPort,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: BadPort,
		},
	},
}
var BadCommonAuditSecurityCtx = corev1.SecurityContext{
	AllowPrivilegeEscalation: &trueVar,
	Privileged:               &trueVar,
	ReadOnlyRootFilesystem:   &trueVar,
	RunAsNonRoot:             &falseVar,
	RunAsUser:                &rootUser,
}

func CommonAuditObj(name, namespace string) *operatorv1.CommonAudit {
	return &operatorv1.CommonAudit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: operatorv1.CommonAuditSpec{},
	}
}

func AuditLoggingObj(name string) *operatorv1alpha1.AuditLogging {
	return &operatorv1alpha1.AuditLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constant.InstanceNamespace,
		},
		Spec: operatorv1alpha1.AuditLoggingSpec{},
	}
}

func NamespaceObj(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func GetFluentdConfig(rg *regexp.Regexp, cmData string) string {
	return strings.Split(rg.FindStringSubmatch(cmData)[0], " ")[1]
}
