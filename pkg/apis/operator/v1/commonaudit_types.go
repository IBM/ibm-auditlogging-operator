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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CommonAuditSpec defines the desired state of CommonAudit
type CommonAuditSpec struct {
	EnableAuditLoggingForwarding bool                   `json:"enabled,omitempty"`
	ClusterIssuer                string                 `json:"clusterIssuer,omitempty"`
	Size                         string                 `json:"size,omitempty"`
	Fluentd                      CommonAuditSpecFluentd `json:"fluentd,omitempty"`
	Outputs                      CommonAuditSpecOutputs `json:"outputs,omitempty"`
}

// CommonAuditSpecFluentd defines the desired state of Fluentd
type CommonAuditSpecFluentd struct {
	ImageRegistry string                   `json:"imageRegistry,omitempty"`
	PullPolicy    string                   `json:"pullPolicy,omitempty"`
	Replicas      int                      `json:"replicas,omitempty"`
	Resources     CommonAuditSpecResources `json:"resources,omitempty"`
}

// CommonAuditSpecResources defines the resources for the fluentd deployment
type CommonAuditSpecResources struct {
	Requests CommonAuditSpecRequirements `json:"requests"`
	Limits   CommonAuditSpecRequirements `json:"limits"`
}

// CommonAuditSpecRequirements defines cpu and memory
type CommonAuditSpecRequirements struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// CommonAuditSpecOutputs defines the configurations for forwarding audit logs to Splunk or Syslog
type CommonAuditSpecOutputs struct {
	Splunk      CommonAuditSpecSplunk        `json:"splunk,omitempty"`
	Syslog      CommonAuditSpecSyslog        `json:"syslog,omitempty"`
	HostAliases []CommonAuditSpecHostAliases `json:"hostAliases,omitempty"`
}

// CommonAuditSpecSplunk defines the configurations for forwarding audit logs to Splunk
type CommonAuditSpecSplunk struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Token string `json:"token"`
	// This is the protocol to use for calling the HEC API.
	// Valid values are: "https" or "http"
	Protocol protocol `json:"protocol"`
}

// +kubebuilder:validation:Enum=https;http
type protocol string

// CommonAuditSpecSyslog defines the configurations for forwarding audit logs to a syslog SIEM
type CommonAuditSpecSyslog struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Hostname string `json:"hostname"`
	TLS      bool   `json:"tls,omitempty"`
}

// CommonAuditSpecHostAliases defines the host alias for an SIEM
type CommonAuditSpecHostAliases struct {
	HostIP    string   `json:"hostIP"`
	Hostnames []string `json:"hostnames"`
}

// CommonAuditStatus defines the observed state of CommonAudit
type CommonAuditStatus struct {
	// The list of pod names for fluentd
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CommonAudit is the Schema for the commonaudits API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=commonaudits,scope=Namespaced
type CommonAudit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CommonAuditSpec   `json:"spec,omitempty"`
	Status CommonAuditStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CommonAuditList contains a list of CommonAudit
type CommonAuditList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CommonAudit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CommonAudit{}, &CommonAuditList{})
}
