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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CommonAuditSpec defines the desired state of CommonAudit
type CommonAuditSpec struct {
	// EnableAuditLoggingForwarding defines if audit logs should be forwarded to an SIEM or not
	EnableAuditLoggingForwarding bool `json:"enabled,omitempty"`
	// ClusterIssuer deprecated, use Issuer
	ClusterIssuer string                 `json:"clusterIssuer,omitempty"`
	Issuer        string                 `json:"issuer,omitempty"`
	Replicas      int32                  `json:"replicas,omitempty"`
	Fluentd       CommonAuditSpecFluentd `json:"fluentd,omitempty"`
	Outputs       CommonAuditSpecOutputs `json:"outputs,omitempty"`
}

// CommonAuditSpecFluentd defines the desired state of Fluentd
type CommonAuditSpecFluentd struct {
	// ImageRegistry deprecated, define image in operator.yaml
	ImageRegistry string                      `json:"imageRegistry,omitempty"`
	PullPolicy    string                      `json:"pullPolicy,omitempty"`
	Resources     corev1.ResourceRequirements `json:"resources,omitempty"`
}

// CommonAuditSpecOutputs defines the configurations for forwarding audit logs to Splunk or Syslog
type CommonAuditSpecOutputs struct {
	Splunk      CommonAuditSpecSplunk        `json:"splunk,omitempty"`
	Syslog      CommonAuditSpecSyslog        `json:"syslog,omitempty"`
	HostAliases []CommonAuditSpecHostAliases `json:"hostAliases,omitempty"`
}

// CommonAuditSpecSplunk defines the configurations for forwarding audit logs to Splunk
type CommonAuditSpecSplunk struct {
	EnableSIEM bool   `json:"enableSIEM"`
	Host       string `json:"host"`
	Port       int32  `json:"port"`
	Token      string `json:"token"`
	TLS        bool   `json:"enableTLS"`
}

// CommonAuditSpecSyslog defines the configurations for forwarding audit logs to a syslog SIEM
type CommonAuditSpecSyslog struct {
	EnableSIEM bool   `json:"enableSIEM"`
	Host       string `json:"host"`
	Port       int32  `json:"port"`
	Hostname   string `json:"hostname"`
	TLS        bool   `json:"enableTLS"`
}

// CommonAuditSpecHostAliases defines the host alias for an SIEM
type CommonAuditSpecHostAliases struct {
	HostIP    string   `json:"ip"`
	Hostnames []string `json:"hostnames"`
}

// StatusVersion defines the Operator versions
type StatusVersion struct {
	Reconciled string `json:"reconciled"`
}

// CommonAuditStatus defines the observed state of CommonAudit
type CommonAuditStatus struct {
	// The list of pod names for fluentd
	Nodes    []string      `json:"nodes"`
	Versions StatusVersion `json:"versions,omitempty"`
}

// +kubebuilder:object:root=true

// CommonAudit is the Schema for the commonaudits API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=commonaudits,scope=Namespaced
type CommonAudit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CommonAuditSpec   `json:"spec,omitempty"`
	Status CommonAuditStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CommonAuditList contains a list of CommonAudit
type CommonAuditList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CommonAudit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CommonAudit{}, &CommonAuditList{})
}
