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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const defaultHTTPPort = 9880
const defaultSyslogPort = 5140

// BuildAuditService returns a Service object
func BuildAuditService(instanceName string, namespace string) *corev1.Service {
	metaLabels := LabelsForMetadata(FluentdName)
	selectorLabels := LabelsForSelector(FluentdName, instanceName)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuditLoggingComponentName,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{
					Name:     AuditLoggingComponentName,
					Protocol: "TCP",
					Port:     defaultHTTPPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: defaultHTTPPort,
					},
				},
				{
					Name:     AuditLoggingComponentName + "-syslog",
					Protocol: "TCP",
					Port:     defaultSyslogPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: defaultSyslogPort,
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
