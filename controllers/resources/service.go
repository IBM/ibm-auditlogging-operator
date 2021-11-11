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
	"reflect"

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	"github.com/IBM/ibm-auditlogging-operator/controllers/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const defaultHTTPPort = 9880
const defaultSyslogPort = 5140
const mutualCertAuthHTTP2Port = 9881
const basicAuthHTTP2Port = 9890

// BuildAuditService returns a Service object
func BuildAuditService(instanceName string, namespace string) *corev1.Service {
	metaLabels := util.LabelsForMetadata(constant.FluentdName)
	selectorLabels := util.LabelsForSelector(constant.FluentdName, instanceName)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constant.AuditLoggingComponentName,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{
					Name:     constant.AuditLoggingComponentName,
					Protocol: "TCP",
					Port:     defaultHTTPPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: defaultHTTPPort,
					},
				},
				{
					Name:     constant.AuditLoggingComponentName + "-syslog",
					Protocol: "TCP",
					Port:     defaultSyslogPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: defaultSyslogPort,
					},
				},
				{
					Name:     constant.AuditLoggingComponentName + "-muatual-auth-http2",
					Protocol: "TCP",
					Port:     mutualCertAuthHTTP2Port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: mutualCertAuthHTTP2Port,
					},
				},
				{
					Name:     constant.AuditLoggingComponentName + "-basic-auth-http2",
					Protocol: "TCP",
					Port:     basicAuthHTTP2Port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: basicAuthHTTP2Port,
					},
				},
			},
			Selector: selectorLabels,
		},
	}
	return service
}
func BuildZenAuditService(instanceName string, namespace string) *corev1.Service {
	metaLabels := util.LabelsForMetadata(constant.FluentdName)
	selectorLabels := util.LabelsForSelector(constant.FluentdName, instanceName)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constant.ZenAuditService,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{
					Name:     constant.ZenAuditService + "-muatual-auth-http2",
					Protocol: "TCP",
					Port:     mutualCertAuthHTTP2Port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: mutualCertAuthHTTP2Port,
					},
				},
				{
					Name:     constant.ZenAuditService + "-basic-auth-http2",
					Protocol: "TCP",
					Port:     basicAuthHTTP2Port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: basicAuthHTTP2Port,
					},
				},
			},
			Selector: selectorLabels,
		},
	}
	return service
}

// EqualServices returns a Boolean
func EqualServices(expected *corev1.Service, found *corev1.Service) bool {
	return !reflect.DeepEqual(found.Spec.Ports, expected.Spec.Ports)
}
