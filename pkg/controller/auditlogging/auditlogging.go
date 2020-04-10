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

package auditlogging

import (
	"context"
	"reflect"

	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-auditlogging-operator/pkg/resources"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileAuditLogging) updateStatus(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("Namespace", res.InstanceNamespace, "Name", instance.Name)

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(res.InstanceNamespace),
		client.MatchingLabels(res.LabelsForSelector(res.FluentdName, instance.Name)),
	}
	if err := r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "AuditLogging.Namespace", res.InstanceNamespace, "AuditLogging.Name", instance.Name)
		return reconcile.Result{}, err
	}
	podNames := []string{}
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	// Get audit-policy-controller pod too
	listOpts = []client.ListOption{
		client.InNamespace(res.InstanceNamespace),
		client.MatchingLabels(res.LabelsForSelector(res.AuditPolicyControllerDeploy, instance.Name)),
	}
	if err := r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "AuditLogging.Namespace", res.InstanceNamespace, "AuditLogging.Name", instance.Name)
		return reconcile.Result{}, err
	}
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		reqLogger.Info("Updating Audit Logging status", "Name", instance.Name)
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateService(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("Service.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)
	expected := res.BuildAuditService(instance)
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: res.InstanceNamespace}, found)
	if err != nil && errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Service", "Service.Namespace", expected.Namespace, "Service.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", expected.Namespace,
				"Service.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	} else if result := res.EqualServices(expected, found); result {
		// If role permissions are incorrect, update it and requeue
		reqLogger.Info("Found ports is incorrect", "Found", found.Spec.Ports, "Expected", expected.Spec.Ports)
		found.Spec.Ports = expected.Spec.Ports
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Service", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createAuditPolicyCRD(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("CRD.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)
	expected := res.BuildAuditPolicyCRD(instance)
	found := &extv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new CRD
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Audit Policy CRD", "CRD.Namespace", expected.Namespace, "CRD.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new CRD", "CRD.Namespace", expected.Namespace,
				"CRD.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// CRD created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get CRD")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateServiceAccounts(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error
	recResult, recErr = r.serviceAccount(instance, res.FluentdDaemonSetName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.serviceAccount(instance, res.AuditPolicyControllerDeploy)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	return reconcile.Result{}, nil
}

// IBMDEV serviceAccountForCR returns (reconcile.Result, error)
func (r *ReconcileAuditLogging) serviceAccount(cr *operatorv1alpha1.AuditLogging, name string) (reconcile.Result, error) {
	reqLogger := log.WithValues("cr.Name", cr.Name)

	expectedRes := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + res.ServiceAcct,
			Namespace: res.InstanceNamespace,
		},
	}
	// Set CR instance as the owner and controller
	err := controllerutil.SetControllerReference(cr, expectedRes, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to define expected resource")
		return reconcile.Result{}, err
	}

	// If ServiceAccount does not exist, create it and requeue
	foundSvcAcct := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: expectedRes.Name, Namespace: res.InstanceNamespace}, foundSvcAcct)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "Namespace", res.InstanceNamespace, "Name", expectedRes.Name)
		err = r.client.Create(context.TODO(), expectedRes)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new ServiceAccount", "Namespace", res.InstanceNamespace, "Name", expectedRes.Name)
			return reconcile.Result{}, err
		}
		// Created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get ServiceAccount")
		return reconcile.Result{}, err
	}
	// No extra validation of the service account required

	// No reconcile was necessary
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateClusterRole(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("ClusterRole.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)
	expected := res.BuildClusterRoleForPolicyController(instance)
	found := &rbacv1.ClusterRole{}
	// Note: clusterroles are cluster-scoped, so this does not search using namespace (unlike other resources above)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ClusterRole
		// newClusterRole := res.BuildClusterRole(instance)
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new ClusterRole", "ClusterRole.Namespace", expected.Namespace, "ClusterRole.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new ClusterRole", "ClusterRole.Namespace", expected.Namespace,
				"ClusterRole.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// ClusterRole created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get ClusterRole")
		return reconcile.Result{}, err
	} else if result := res.EqualClusterRoles(expected, found); result {
		// If role permissions are incorrect, update it and requeue
		reqLogger.Info("Found role is incorrect", "Found", found.Rules, "Expected", expected.Rules)
		found.Rules = expected.Rules
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update role", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateClusterRoleBinding(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("ClusterRoleBinding.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)
	expected := res.BuildClusterRoleBindingForPolicyController(instance)
	found := &rbacv1.ClusterRoleBinding{}
	// Note: clusterroles are cluster-scoped, so this does not search using namespace (unlike other resources above)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Role
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new ClusterRoleBinding", "ClusterRole.Namespace", expected.Namespace, "ClusterRoleBinding.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new ClusterRoleBinding", "ClusterRoleBinding.Namespace", expected.Namespace,
				"RoleBinding.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// ClusterRoleBinding created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get ClusterRoleBinding")
		return reconcile.Result{}, err
	} else if result := res.EqualClusterRoleBindings(expected, found); result {
		// If rolebinding is incorrect, delete it and requeue
		reqLogger.Info("Found rolebinding is incorrect", "Found", found.Subjects, "Expected", expected.Subjects)
		err = r.client.Delete(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to delete rolebinding", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Deleted - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateRole(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("Role.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)
	expected := res.BuildRoleForFluentd(instance)
	found := &rbacv1.Role{}
	// Note: clusterroles are cluster-scoped, so this does not search using namespace (unlike other resources above)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: res.InstanceNamespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Role
		// newClusterRole := res.BuildRole(instance)
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Role", "Role.Namespace", expected.Namespace, "Role.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			reqLogger.Info("Already exists", "Role.Namespace", expected.Namespace, "Role.Name", expected.Name)
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new Role", "Role.Namespace", expected.Namespace,
				"Role.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Role created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Role")
		return reconcile.Result{}, err
	} else if result := res.EqualRoles(expected, found); result {
		// If role permissions are incorrect, update it and requeue
		reqLogger.Info("Found role is incorrect", "Found", found.Rules, "Expected", expected.Rules)
		found.Rules = expected.Rules
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update role", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateRoleBinding(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("RoleBinding.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)
	expected := res.BuildRoleBindingForFluentd(instance)
	found := &rbacv1.RoleBinding{}
	// Note: clusterroles are cluster-scoped, so this does not search using namespace (unlike other resources above)
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: res.InstanceNamespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Role
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new RoleBinding", "Role.Namespace", expected.Namespace, "RoleBinding.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new RoleBinding", "RoleBinding.Namespace", expected.Namespace,
				"RoleBinding.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// RoleBinding created successfully - return and requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get RoleBinding")
		return reconcile.Result{}, err
	} else if result := res.EqualRoleBindings(expected, found); result {
		// If rolebinding is incorrect, delete it and requeue
		reqLogger.Info("Found rolebinding is incorrect", "Found", found.Subjects, "Expected", expected.Subjects)
		err = r.client.Delete(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to delete rolebinding", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Deleted - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdatePolicyControllerDeployment(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("Deployment.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)

	expected := res.BuildDeploymentForPolicyController(instance)
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.AuditPolicyControllerDeploy, Namespace: res.InstanceNamespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Deployment
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Audit Policy Controller Deployment", "Deployment.Namespace", expected.Namespace, "Deployment.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new Audit Policy Controller Deployment", "Deployment.Namespace", expected.Namespace,
				"Deployment.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	} else if result := res.EqualDeployments(expected, found); result {
		// If spec is incorrect, update it and requeue
		reqLogger.Info("Found deployment spec is incorrect", "Found", found.Spec.Template.Spec, "Expected", expected.Spec.Template.Spec)
		found.Spec.Template.Spec.Volumes = expected.Spec.Template.Spec.Volumes
		found.Spec.Template.Spec.Containers = expected.Spec.Template.Spec.Containers
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Namespace", res.InstanceNamespace, "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateAuditConfigMaps(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error
	recResult, recErr = r.createOrUpdateConfig(instance, res.FluentdDaemonSetName+"-"+res.ConfigName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	// FIX
	recResult, recErr = r.createOrUpdateConfig(instance, res.FluentdDaemonSetName+"-"+res.SourceConfigName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.createOrUpdateConfig(instance, res.FluentdDaemonSetName+"-"+res.SplunkConfigName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.createOrUpdateConfig(instance, res.FluentdDaemonSetName+"-"+res.QRadarConfigName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.createOrUpdateConfig(instance, res.FluentdDaemonSetName+"-"+res.ELKConfigName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateConfig(instance *operatorv1alpha1.AuditLogging, configName string) (reconcile.Result, error) {
	reqLogger := log.WithValues("ConfigMap.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)
	configMapFound := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: configName, Namespace: res.InstanceNamespace}, configMapFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ConfigMap
		newConfigMap, err := res.BuildConfigMap(instance, configName)
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap")
			return reconcile.Result{}, err
		}
		if err := controllerutil.SetControllerReference(instance, newConfigMap, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", newConfigMap.Namespace, "ConfigMap.Name", newConfigMap.Name)
		err = r.client.Create(context.TODO(), newConfigMap)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", newConfigMap.Namespace,
				"ConfigMap.Name", newConfigMap.Name)
			return reconcile.Result{}, err
		}
		// ConfigMap created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get ConfigMap")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateFluentdDaemonSet(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	reqLogger := log.WithValues("Daemonset.Namespace", res.InstanceNamespace, "instance.Name", instance.Name)
	expected := res.BuildDaemonForFluentd(instance)
	found := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName, Namespace: res.InstanceNamespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new DaemonSet
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Fluentd DaemonSet", "Daemonset.Namespace", expected.Namespace, "Daemonset.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new Fluentd DaemonSet", "Daemonset.Namespace", expected.Namespace,
				"Daemonset.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// DaemonSet created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get DaemonSet")
		return reconcile.Result{}, err
	} else if result := res.EqualDaemonSets(expected, found); result {
		// If spec is incorrect, update it and requeue
		reqLogger.Info("Found daemonset spec is incorrect", "Found", found.Spec.Template.Spec, "Expected", expected.Spec.Template.Spec)
		found.Spec.Template.Spec.Volumes = expected.Spec.Template.Spec.Volumes
		found.Spec.Template.Spec.Containers = expected.Spec.Template.Spec.Containers
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Daemonset", "Namespace", res.InstanceNamespace, "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateAuditCerts(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error
	recResult, recErr = r.createOrUpdateAuditCertificate(instance, res.AuditLoggingHTTPSCertName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.createOrUpdateAuditCertificate(instance, res.AuditLoggingCertName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAuditLogging) createOrUpdateAuditCertificate(instance *operatorv1alpha1.AuditLogging, name string) (reconcile.Result, error) {
	reqLogger := log.WithValues("Certificate.Namespace", res.InstanceNamespace, "Instance.Name", instance.Name)
	expectedCert := res.BuildCertsForAuditLogging(instance, instance.Spec.Fluentd.ClusterIssuer, name)
	foundCert := &certmgr.Certificate{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expectedCert.Name, Namespace: expectedCert.ObjectMeta.Namespace}, foundCert)
	if err != nil && errors.IsNotFound(err) {
		// Set Audit Logging instance as the owner and controller of the Certificate
		err := controllerutil.SetControllerReference(instance, expectedCert, r.scheme)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to set owner for Certificate")
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Fluentd Certificate", "Certificate.Namespace", expectedCert.Namespace, "Certificate.Name", expectedCert.Name)
		err = r.client.Create(context.TODO(), expectedCert)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Fluentd Certificate", "Certificate.Namespace", expectedCert.Namespace,
				"Certificate.Name", expectedCert.Name)
			return reconcile.Result{}, err
		}
		// Certificate created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Certificate")
		return reconcile.Result{}, err
	} else if result := res.EqualCerts(expectedCert, foundCert); result {
		// If spec is incorrect, update it and requeue
		reqLogger.Info("Found Certificate spec is incorrect", "Found", foundCert.Spec, "Expected", expectedCert.Spec)
		foundCert.Spec = expectedCert.Spec
		err = r.client.Update(context.TODO(), foundCert)
		if err != nil {
			reqLogger.Error(err, "Failed to update Certificate", "Namespace", foundCert.ObjectMeta.Namespace, "Name", foundCert.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}
