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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AuditLoggingSpec defines the desired state of AuditLogging
type AuditLoggingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	OperatorVersion   string                           `json:"operatorVersion,omitempty"`
	Fluentd           AuditLoggingSpecFluentd          `json:"fluentd,omitempty"`
	PolicyController  AuditLoggingSpecPolicyController `json:"policyController,omitempty"`
	InstanceNamespace string                           `json:"instanceNamespace"`
}

// AuditLoggingSpecFluentd defines the desired state of Fluentd
type AuditLoggingSpecFluentd struct {
	EnableAuditLoggingForwarding bool   `json:"enabled,omitempty"`
	ImageRegistry                string `json:"imageRegistry,omitempty"`
	ImageTagPostfix              string `json:"imageTagPostfix,omitempty"`
	PullPolicy                   string `json:"pullPolicy,omitempty"`
	JournalPath                  string `json:"journalPath,omitempty"`
	ClusterIssuer                string `json:"clusterIssuer,omitempty"`
}

// AuditLoggingSpecPolicyController defines the policy controller configuration in the the audit logging spec
type AuditLoggingSpecPolicyController struct {
	ImageRegistry   string `json:"imageRegistry,omitempty"`
	ImageTagPostfix string `json:"imageTagPostfix,omitempty"`
	PullPolicy      string `json:"pullPolicy,omitempty"`
	Verbosity       string `json:"verbosity,omitempty"`
	Frequency       string `json:"frequency,omitempty"`
	Duration        string `json:"duration,omitempty"`
}

// AuditLoggingStatus defines the observed state of AuditLogging
type AuditLoggingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AuditLogging is the Schema for the auditloggings API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=auditloggings,scope=Cluster
type AuditLogging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuditLoggingSpec   `json:"spec,omitempty"`
	Status AuditLoggingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AuditLoggingList contains a list of AuditLogging
type AuditLoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuditLogging `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuditLogging{}, &AuditLoggingList{})
}
