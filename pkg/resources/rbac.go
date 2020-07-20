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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperandServiceAccount defines the name of the operands' ServiceAccount
const OperandServiceAccount = "ibm-auditlogging-operand"
const RolePostfix = "-role"
const RoleBindingPostfix = "-rolebinding"

// BuildServiceAccount returns a ServiceAccoutn object
func BuildServiceAccount(namespace string) *corev1.ServiceAccount {
	metaLabels := LabelsForMetadata(OperandServiceAccount)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      OperandServiceAccount,
			Namespace: namespace,
			Labels:    metaLabels,
		},
	}
	return sa
}

// BuildRoleBinding returns a RoleBinding object for fluentd
func BuildRoleBinding(namespace string) *rbacv1.RoleBinding {
	metaLabels := LabelsForMetadata(FluentdName)
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDaemonSetName + RoleBindingPostfix,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Subjects: []rbacv1.Subject{{
			APIGroup:  "",
			Kind:      "ServiceAccount",
			Name:      OperandServiceAccount,
			Namespace: namespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     FluentdDaemonSetName + RolePostfix,
		},
	}
	return rb
}

// BuildRole returns a Role object for fluentd
func BuildRole(namespace string, journalAccess bool) *rbacv1.Role {
	metaLabels := LabelsForMetadata(FluentdName)
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDaemonSetName + RolePostfix,
			Namespace: namespace,
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
	return role
}

// EqualRoles returns a Boolean
func EqualRoles(expected *rbacv1.Role, found *rbacv1.Role) bool {
	return !reflect.DeepEqual(found.Rules, expected.Rules)
}

// EqualClusterRoles returns a Boolean
func EqualClusterRoles(expected *rbacv1.ClusterRole, found *rbacv1.ClusterRole) bool {
	return !reflect.DeepEqual(found.Rules, expected.Rules)
}

// EqualRoleBindings returns a Boolean
func EqualRoleBindings(expected *rbacv1.RoleBinding, found *rbacv1.RoleBinding) bool {
	return !reflect.DeepEqual(found.Subjects, expected.Subjects) ||
		!reflect.DeepEqual(found.RoleRef, expected.RoleRef)
}

// EqualClusterRoleBindings returns a Boolean
func EqualClusterRoleBindings(expected *rbacv1.ClusterRoleBinding, found *rbacv1.ClusterRoleBinding) bool {
	return !reflect.DeepEqual(found.Subjects, expected.Subjects) ||
		!reflect.DeepEqual(found.RoleRef, expected.RoleRef)
}
