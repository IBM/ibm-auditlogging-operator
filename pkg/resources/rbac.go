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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const OperandRBAC = "ibm-auditlogging-operand"
const rolePostfix = "-role"
const roleBindingPostfix = "-rolebinding"

// BuildServiceAccount returns a ServiceAccoutn object
func BuildServiceAccount(instance *operatorv1alpha1.AuditLogging) *corev1.ServiceAccount {
	metaLabels := LabelsForMetadata(OperandRBAC)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      OperandRBAC,
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
		},
	}
	return sa
}

// BuildClusterRoleBinding returns a ClusterRoleBinding object
func BuildClusterRoleBinding(instance *operatorv1alpha1.AuditLogging) *rbacv1.ClusterRoleBinding {
	metaLabels := LabelsForMetadata(AuditPolicyControllerDeploy)
	rb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   AuditPolicyControllerDeploy + roleBindingPostfix,
			Labels: metaLabels,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      OperandRBAC,
			Namespace: InstanceNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     AuditPolicyControllerDeploy,
		},
	}
	return rb
}

// BuildClusterRole returns a ClusterRole object
func BuildClusterRole(instance *operatorv1alpha1.AuditLogging) *rbacv1.ClusterRole {
	metaLabels := LabelsForMetadata(AuditPolicyControllerDeploy)
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   AuditPolicyControllerDeploy + rolePostfix,
			Labels: metaLabels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{""},
				Resources: []string{"services"},
			},
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{""},
				Resources: []string{"secrets"},
			},
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"mutatingwebhookconfigurations", "validatingwebhookconfigurations"},
			},
			{
				Verbs:     []string{"get", "update", "patch"},
				APIGroups: []string{"audit.policies.ibm.com"},
				Resources: []string{"auditpolicies/status"},
			},
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{"audit.policies.ibm.com"},
				Resources: []string{"auditpolicies"},
			},
			{
				Verbs:     []string{"get", "update", "patch"},
				APIGroups: []string{"apps"},
				Resources: []string{"deployments/status"},
			},
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
			},
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{""},
				Resources: []string{"pods", "namespaces"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "update"},
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
			},
			{
				Verbs:     []string{"create", "get", "update", "patch"},
				APIGroups: []string{""},
				Resources: []string{"events"},
			},
			{
				Verbs:         []string{"use"},
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"restricted"},
			},
		},
	}
	return cr
}

// BuildRoleBindingForFluentd returns a RoleBinding object for fluentd
func BuildRoleBinding(instance *operatorv1alpha1.AuditLogging) *rbacv1.RoleBinding {
	metaLabels := LabelsForMetadata(FluentdName)
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDaemonSetName + roleBindingPostfix,
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
		},
		Subjects: []rbacv1.Subject{{
			APIGroup:  "",
			Kind:      "ServiceAccount",
			Name:      OperandRBAC,
			Namespace: InstanceNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     FluentdDaemonSetName,
		},
	}
	return rb
}

// BuildRoleForFluentd returns a Role object for fluentd
func BuildRole(instance *operatorv1alpha1.AuditLogging) *rbacv1.Role {
	metaLabels := LabelsForMetadata(FluentdName)
	cr := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDaemonSetName + rolePostfix,
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"use"},
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"privileged"},
			},
		},
	}
	return cr
}

func EqualRoles(expected *rbacv1.Role, found *rbacv1.Role) bool {
	return !reflect.DeepEqual(found.Rules, expected.Rules)
}

func EqualClusterRoles(expected *rbacv1.ClusterRole, found *rbacv1.ClusterRole) bool {
	return !reflect.DeepEqual(found.Rules, expected.Rules)
}

func EqualRoleBindings(expected *rbacv1.RoleBinding, found *rbacv1.RoleBinding) bool {
	return !reflect.DeepEqual(found.Subjects, expected.Subjects) ||
		!reflect.DeepEqual(found.RoleRef, expected.RoleRef)
}

func EqualClusterRoleBindings(expected *rbacv1.ClusterRoleBinding, found *rbacv1.ClusterRoleBinding) bool {
	return !reflect.DeepEqual(found.Subjects, expected.Subjects) ||
		!reflect.DeepEqual(found.RoleRef, expected.RoleRef)
}
