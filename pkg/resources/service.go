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

	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// BuildAuditService returns a Service object
func BuildAuditService(instance *operatorv1alpha1.AuditLogging) *corev1.Service {
	metaLabels := LabelsForMetadata(FluentdName)
	selectorLabels := LabelsForSelector(FluentdName, instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuditLoggingComponentName,
			Namespace: InstanceNamespace,
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
