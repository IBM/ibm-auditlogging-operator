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

package controllers

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"

	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	operatorv1alpha1 "github.com/IBM/ibm-auditlogging-operator/api/v1alpha1"
	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	res "github.com/IBM/ibm-auditlogging-operator/controllers/resources"
	"github.com/IBM/ibm-auditlogging-operator/controllers/util"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *AuditLoggingReconciler) reconcileService(instance *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	expected := res.BuildAuditService(instance.Name, namespace)
	found := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Service", "Service.Namespace", expected.Namespace, "Service.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Service", "Service.Namespace", expected.Namespace,
				"Service.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	} else if result := res.EqualServices(expected, found); result {
		// If ports are incorrect, delete it and requeue
		r.Log.Info("Found ports are incorrect", "Found", found.Spec.Ports, "Expected", expected.Spec.Ports)
		err = r.Client.Delete(context.TODO(), found)
		if err != nil {
			r.Log.Error(err, "Failed to delete Service", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) reconcileJob(instance *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	found := &batchv1.Job{}
	expected := res.BuildJobForAuditLogging(instance.Name, namespace)
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: expected.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Job", "Job.Namespace", expected.Namespace, "Job.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Service", "Service.Namespace", expected.Namespace,
				"Service.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Job")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) removeOldPolicyControllerDeploy(instance *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	// Policy controller container has been moved to operator pod in 3.7, remove redundant deployment
	policyDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.AuditPolicyControllerDeploy,
			Namespace: namespace,
		},
	}
	// check if the deployment exists
	err := r.Client.Get(context.TODO(),
		types.NamespacedName{Name: res.AuditPolicyControllerDeploy, Namespace: namespace}, policyDeploy)
	if err == nil {
		// found deployment so delete it
		err := r.Client.Delete(context.TODO(), policyDeploy)
		if err != nil {
			r.Log.Error(err, "Failed to delete old policy controller deployment")
			return reconcile.Result{}, err
		}
		r.Log.Info("Deleted old policy controller deployment")
		return reconcile.Result{Requeue: true}, nil
	} else if !errors.IsNotFound(err) {
		// if err is NotFound do nothing, else print an error msg
		r.Log.Error(err, "Failed to get old policy controller deployment")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) reconcileAuditConfigMaps(instance *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error

	for _, cm := range res.FluentdConfigMaps {
		recResult, recErr = r.reconcileConfig(instance, cm, namespace)
		if recErr != nil || recResult.Requeue {
			return recResult, recErr
		}
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) reconcileConfig(instance *operatorv1alpha1.AuditLogging, configName string, namespace string) (reconcile.Result, error) {
	expected, err := res.BuildConfigMap(instance, configName, namespace)
	if err != nil {
		r.Log.Error(err, "Failed to create ConfigMap")
		return reconcile.Result{}, err
	}
	found := &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: configName, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ConfigMap
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", expected.Namespace, "ConfigMap.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", expected.Namespace,
				"ConfigMap.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// ConfigMap created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get ConfigMap")
		return reconcile.Result{}, err
	}
	// ConfigMap was found, check for expected values
	var update = false
	if !util.EqualLabels(found.ObjectMeta.Labels, expected.ObjectMeta.Labels) {
		found.ObjectMeta.Labels = expected.ObjectMeta.Labels
		update = true
	}
	switch configName {
	case res.FluentdDaemonSetName + "-" + res.ConfigName:
		if !res.EqualConfig(found, expected, res.EnableAuditLogForwardKey) {
			found.Data[res.EnableAuditLogForwardKey] = expected.Data[res.EnableAuditLogForwardKey]
			update = true
		}
	case res.FluentdDaemonSetName + "-" + res.SourceConfigName:
		if !res.EqualConfig(found, expected, res.SourceConfigKey) {
			found.Data[res.SourceConfigKey] = expected.Data[res.SourceConfigKey]
			update = true
		}
	case res.FluentdDaemonSetName + "-" + res.HTTPIngestName:
		if !res.EqualConfig(found, expected, res.HTTPIngestURLKey) {
			found.Data[res.HTTPIngestURLKey] = expected.Data[res.HTTPIngestURLKey]
			update = true
		}
	case res.FluentdDaemonSetName + "-" + res.SplunkConfigName:
		r.Log.Info("Checking output configs")
		fallthrough
	case res.FluentdDaemonSetName + "-" + res.QRadarConfigName:
		// Ensure match tags are correct
		if !res.EqualMatchTags(found) {
			// Keep customer SIEM configs
			data, err := res.BuildWithSIEMConfigs(found)
			if err != nil {
				r.Log.Error(err, "Failed to get SIEM configs", "Name", found.Name, "Found output config", data)
				return reconcile.Result{}, err
			}
			if configName == res.FluentdDaemonSetName+"-"+res.SplunkConfigName {
				found.Data[res.SplunkConfigKey] = data
			} else {
				found.Data[res.QRadarConfigKey] = data
			}
			update = true
		}
	default:
		r.Log.Info("Unknown ConfigMap name", "Name", configName)
	}
	if update {
		err = r.Client.Update(context.TODO(), found)
		if err != nil {
			r.Log.Error(err, "Failed to update ConfigMap", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Updated - return and requeue
		r.Log.Info("Updating ConfigMap", "ConfigMap.Name", found.Name)
		err = restartFluentdPods(r, instance, namespace)
		if err != nil {
			r.Log.Error(err, "Failed to restart fluentd pods")
		}
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func restartFluentdPods(r *AuditLoggingReconciler, instance *operatorv1alpha1.AuditLogging, namespace string) error {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(util.LabelsForSelector(constant.FluentdName, instance.Name)),
	}
	if err := r.Client.List(context.TODO(), podList, listOpts...); err != nil {
		r.Log.Error(err, "Failed to list pods")
		return err
	}
	for _, pod := range podList.Items {
		p := pod
		err := r.Client.Delete(context.TODO(), &p)
		if err != nil {
			r.Log.Error(err, "Failed to delete pod", "Pod.Name", pod.Name)
			return err
		}
	}
	r.Log.Info("Restarted pods", "Pods", util.GetPodNames(podList.Items))
	return nil
}

func (r *AuditLoggingReconciler) reconcileFluentdDaemonSet(instance *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	expected := res.BuildDaemonForFluentd(instance, namespace)
	found := &appsv1.DaemonSet{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new DaemonSet
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Fluentd DaemonSet", "Daemonset.Namespace", expected.Namespace, "Daemonset.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Fluentd DaemonSet", "Daemonset.Namespace", expected.Namespace,
				"Daemonset.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// DaemonSet created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get DaemonSet")
		return reconcile.Result{}, err
	} else if !res.EqualDaemonSets(expected, found) {
		// If spec is incorrect, update it and requeue
		found.ObjectMeta.Labels = expected.ObjectMeta.Labels
		// Keep hostAliases
		temp := found.Spec.Template.Spec.HostAliases
		found.Spec = expected.Spec
		found.Spec.Template.Spec.HostAliases = temp
		err = r.Client.Update(context.TODO(), found)
		if err != nil {
			r.Log.Error(err, "Failed to update Daemonset", "Namespace", namespace, "Name", found.Name)
			return reconcile.Result{}, err
		}
		r.Log.Info("Updating Fluentd DaemonSet", "Daemonset.Name", found.Name)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) reconcileAuditCerts(instance *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error
	recResult, recErr = r.reconcileAuditCertificate(instance, res.AuditLoggingHTTPSCertName, namespace)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.reconcileAuditCertificate(instance, res.AuditLoggingCertName, namespace)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) reconcileAuditCertificate(instance *operatorv1alpha1.AuditLogging, name string, namespace string) (reconcile.Result, error) {
	expectedCert := res.BuildCertsForAuditLogging(namespace, instance.Spec.Fluentd.ClusterIssuer, name)
	foundCert := &certmgr.Certificate{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expectedCert.Name, Namespace: expectedCert.ObjectMeta.Namespace}, foundCert)
	if err != nil && errors.IsNotFound(err) {
		// Set Audit Logging instance as the owner and controller of the Certificate
		if err := controllerutil.SetControllerReference(instance, expectedCert, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Fluentd Certificate", "Certificate.Namespace", expectedCert.Namespace, "Certificate.Name", expectedCert.Name)
		err = r.Client.Create(context.TODO(), expectedCert)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Fluentd Certificate", "Certificate.Namespace", expectedCert.Namespace,
				"Certificate.Name", expectedCert.Name)
			return reconcile.Result{}, err
		}
		// Certificate created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Certificate")
		return reconcile.Result{}, err
	} else if result := res.EqualCerts(expectedCert, foundCert); result {
		// If spec is incorrect, update it and requeue
		r.Log.Info("Found Certificate spec is incorrect", "Found", foundCert.Spec, "Expected", expectedCert.Spec)
		foundCert.Spec = expectedCert.Spec
		err = r.Client.Update(context.TODO(), foundCert)
		if err != nil {
			r.Log.Error(err, "Failed to update Certificate", "Namespace", foundCert.ObjectMeta.Namespace, "Name", foundCert.Name)
			return reconcile.Result{}, err
		}
		r.Log.Info("Updating Fluentd Certificate", "Certificate.Name", foundCert.Name)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) reconcileServiceAccount(cr *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	expectedRes := res.BuildServiceAccount(namespace)
	// Set CR instance as the owner and controller
	err := controllerutil.SetControllerReference(cr, expectedRes, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Failed to define expected resource")
		return reconcile.Result{}, err
	}

	// If ServiceAccount does not exist, create it and requeue
	foundSvcAcct := &corev1.ServiceAccount{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: expectedRes.Name, Namespace: namespace}, foundSvcAcct)
	if err != nil && errors.IsNotFound(err) {
		r.Log.Info("Creating a new ServiceAccount", "Namespace", namespace, "Name", expectedRes.Name)
		err = r.Client.Create(context.TODO(), expectedRes)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new ServiceAccount", "Namespace", namespace, "Name", expectedRes.Name)
			return reconcile.Result{}, err
		}
		// Created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get ServiceAccount")
		return reconcile.Result{}, err
	}
	// No extra validation of the service account required

	// No reconcile was necessary
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) removeOldRBAC(namespace string) {
	fluentdSA := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.FluentdDaemonSetName + "-svcacct",
			Namespace: namespace,
		},
	}
	// check if the service account exists
	err := r.Client.Get(context.TODO(),
		types.NamespacedName{Name: res.FluentdDaemonSetName + "-svcacct", Namespace: namespace}, fluentdSA)
	if err == nil {
		// found service account so delete it
		err := r.Client.Delete(context.TODO(), fluentdSA)
		if err != nil {
			r.Log.Error(err, "Failed to delete old fluentd service account")
		} else {
			r.Log.Info("Deleted old fluentd service account")
		}
	} else if !errors.IsNotFound(err) {
		// if err is NotFound do nothing, else print an error msg
		r.Log.Error(err, "Failed to get old fluentd service account")
	}

	policySA := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.AuditPolicyControllerDeploy + "-svcacct",
			Namespace: namespace,
		},
	}
	// check if the service account exists
	err = r.Client.Get(context.TODO(),
		types.NamespacedName{Name: res.AuditPolicyControllerDeploy + "-svcacct", Namespace: namespace}, policySA)
	if err == nil {
		// found service account so delete it
		err := r.Client.Delete(context.TODO(), policySA)
		if err != nil {
			r.Log.Error(err, "Failed to delete old policy controller service account")
		} else {
			r.Log.Info("Deleted old policy controller service account")
		}
	} else if !errors.IsNotFound(err) {
		// if err is NotFound do nothing, else print an error msg
		r.Log.Error(err, "Failed to get old policy controller service account")
	}

	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.AuditPolicyControllerDeploy + res.RolePostfix,
			Namespace: namespace,
		},
	}
	// check if the clusterrole exists
	err = r.Client.Get(context.TODO(),
		types.NamespacedName{Name: res.AuditPolicyControllerDeploy + res.RolePostfix, Namespace: namespace}, cr)
	if err == nil {
		// found clusterrole so delete it
		err := r.Client.Delete(context.TODO(), cr)
		if err != nil {
			r.Log.Error(err, "Failed to delete old policy controller clusterrole")
		} else {
			r.Log.Info("Deleted old policy controller clusterrole")
		}
	} else if !errors.IsNotFound(err) {
		// if err is NotFound do nothing, else print an error msg
		r.Log.Error(err, "Failed to get old policy controller clusterrole")
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.AuditPolicyControllerDeploy + res.RoleBindingPostfix,
			Namespace: namespace,
		},
	}
	// check if the clusterrolebinding exists
	err = r.Client.Get(context.TODO(),
		types.NamespacedName{Name: res.AuditPolicyControllerDeploy + res.RoleBindingPostfix, Namespace: namespace}, crb)
	if err == nil {
		// found clusterrolebinding so delete it
		err := r.Client.Delete(context.TODO(), crb)
		if err != nil {
			r.Log.Error(err, "Failed to delete old policy controller clusterrolebinding")
		} else {
			r.Log.Info("Deleted old policy controller clusterrolebinding")
		}
	} else if !errors.IsNotFound(err) {
		// if err is NotFound do nothing, else print an error msg
		r.Log.Error(err, "Failed to get old policy controller clusterrolebinding")
	}
}

func (r *AuditLoggingReconciler) reconcileRole(instance *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	expected := res.BuildRole(namespace, true)
	found := &rbacv1.Role{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Role
		// newClusterRole := res.BuildRole(instance)
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Role", "Role.Namespace", expected.Namespace, "Role.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			r.Log.Info("Already exists", "Role.Namespace", expected.Namespace, "Role.Name", expected.Name)
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Role", "Role.Namespace", expected.Namespace,
				"Role.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Role created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Role")
		return reconcile.Result{}, err
	} else if result := res.EqualRoles(expected, found); result {
		// If role permissions are incorrect, update it and requeue
		r.Log.Info("Found role is incorrect", "Found", found.Rules, "Expected", expected.Rules)
		found.Rules = expected.Rules
		err = r.Client.Update(context.TODO(), found)
		if err != nil {
			r.Log.Error(err, "Failed to update role", "Name", found.Name)
			return reconcile.Result{}, err
		}
		r.Log.Info("Updating Role", "Role.Name", found.Name)
		// Updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) reconcileRoleBinding(instance *operatorv1alpha1.AuditLogging, namespace string) (reconcile.Result, error) {
	expected := res.BuildRoleBinding(namespace)
	found := &rbacv1.RoleBinding{}
	// Note: clusterroles are cluster-scoped, so this does not search using namespace (unlike other resources above)
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Role
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new RoleBinding", "Role.Namespace", expected.Namespace, "RoleBinding.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new RoleBinding", "RoleBinding.Namespace", expected.Namespace,
				"RoleBinding.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// RoleBinding created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get RoleBinding")
		return reconcile.Result{}, err
	} else if result := res.EqualRoleBindings(expected, found); result {
		// If rolebinding is incorrect, delete it and requeue
		r.Log.Info("Found rolebinding is incorrect", "Found", found.Subjects, "Expected", expected.Subjects)
		err = r.Client.Delete(context.TODO(), found)
		if err != nil {
			r.Log.Error(err, "Failed to delete rolebinding", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Deleted - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}
