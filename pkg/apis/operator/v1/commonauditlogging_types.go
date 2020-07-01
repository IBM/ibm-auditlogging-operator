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

// CommonAuditLoggingSpec defines the desired state of CommonAuditLogging
type CommonAuditLoggingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	EnableAuditLoggingForwarding bool                         `json:"enabled,omitempty"`
	ImageRegistry                string                       `json:"imageRegistry,omitempty"`
	PullPolicy                   string                       `json:"pullPolicy,omitempty"`
	ClusterIssuer                string                       `json:"clusterIssuer,omitempty"`
	Output                       CommonAuditLoggingSpecOutput `json:"output,omitempty"`
}

// CommonAuditLoggingSpecOutput defines the configurations for forwarding audit logs to Splunk or QRadar
type CommonAuditLoggingSpecOutput struct {
	Splunk      CommonAuditLoggingSpecSplunk      `json:"splunk,omitempty"`
	QRadar      CommonAuditLoggingSpecQRadar      `json:"qradar,omitempty"`
	HostAliases []CommonAuditLoggingSpecHostAlias `json:"hostAlias,omitempty"`
}

// CommonAuditLoggingSpecSplunk defines the configurations for forwarding audit logs to Splunk
type CommonAuditLoggingSpecSplunk struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Token string `json:"token"`
}

// CommonAuditLoggingSpecQRadar defines the configurations for forwarding audit logs to QRadar
type CommonAuditLoggingSpecQRadar struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Hostname string `json:"hostname"`
}

// CommonAuditLoggingSpecHostAlias defines the host alias for an SIEM
type CommonAuditLoggingSpecHostAlias struct {
	HostIP   string `json:"hostIP"`
	Hostname string `json:"hostname"`
}

// CommonAuditLoggingStatus defines the observed state of CommonAuditLogging
type CommonAuditLoggingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CommonAuditLogging is the Schema for the commonauditloggings API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=commonauditloggings,scope=Namespaced
type CommonAuditLogging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CommonAuditLoggingSpec   `json:"spec,omitempty"`
	Status CommonAuditLoggingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CommonAuditLoggingList contains a list of CommonAuditLogging
type CommonAuditLoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CommonAuditLogging `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CommonAuditLogging{}, &CommonAuditLoggingList{})
}
