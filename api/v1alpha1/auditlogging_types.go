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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AuditLoggingSpec defines the desired state of AuditLogging
type AuditLoggingSpec struct {
	// Fluentd defines the desired state of Fluentd
	Fluentd AuditLoggingSpecFluentd `json:"fluentd,omitempty"`
	// PolicyController has been deprecated.
	PolicyController AuditLoggingSpecPolicyController `json:"policyController,omitempty"`
}

// AuditLoggingSpecFluentd defines the desired state of Fluentd
type AuditLoggingSpecFluentd struct {
	EnableAuditLoggingForwarding bool   `json:"enabled,omitempty"`
	ImageRegistry                string `json:"imageRegistry,omitempty"`
	// ImageTag no longer supported. Define image sha or tag in operator.yaml
	ImageTag    string `json:"imageTag,omitempty"`
	PullPolicy  string `json:"pullPolicy,omitempty"`
	JournalPath string `json:"journalPath,omitempty"`
	// ClusterIssuer deprecated, use Issuer
	ClusterIssuer string                      `json:"clusterIssuer,omitempty"`
	Issuer        string                      `json:"issuer,omitempty"`
	Resources     corev1.ResourceRequirements `json:"resources,omitempty"`
}

// AuditLoggingSpecPolicyController defines the policy controller configuration in the the audit logging spec.
type AuditLoggingSpecPolicyController struct {
	// +kubebuilder:validation:Pattern=true|false
	EnableAuditPolicy string `json:"enabled,omitempty"`
	ImageRegistry     string `json:"imageRegistry,omitempty"`
	ImageTag          string `json:"imageTag,omitempty"`
	PullPolicy        string `json:"pullPolicy,omitempty"`
	Verbosity         string `json:"verbosity,omitempty"`
	Frequency         string `json:"frequency,omitempty"`
}

// StatusVersion defines the Operator versions
type StatusVersion struct {
	Reconciled string `json:"reconciled"`
}

// AuditLoggingStatus defines the observed state of AuditLogging
type AuditLoggingStatus struct {
	// Nodes defines the names of the audit pods
	Nodes    []string      `json:"nodes"`
	Versions StatusVersion `json:"versions,omitempty"`
}

// +kubebuilder:object:root=true

// AuditLogging is the Schema for the auditloggings API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=auditloggings,scope=Cluster
type AuditLogging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuditLoggingSpec   `json:"spec,omitempty"`
	Status AuditLoggingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AuditLoggingList contains a list of AuditLogging
type AuditLoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuditLogging `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuditLogging{}, &AuditLoggingList{})
}
