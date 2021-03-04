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

	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"
	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	res "github.com/IBM/ibm-auditlogging-operator/controllers/resources"
	"github.com/IBM/ibm-auditlogging-operator/controllers/util"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *CommonAuditReconciler) reconcileService(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	expected := res.BuildAuditService(instance.Name, instance.Namespace)
	found := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: instance.Namespace}, found)
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

func (r *CommonAuditReconciler) reconcileAuditConfigMaps(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error

	for _, cm := range res.FluentdConfigMaps {
		recResult, recErr = r.reconcileConfig(instance, cm)
		if recErr != nil || recResult.Requeue {
			return recResult, recErr
		}
	}
	return reconcile.Result{}, nil
}

func (r *CommonAuditReconciler) reconcileConfig(instance *operatorv1.CommonAudit, configName string) (reconcile.Result, error) {
	expected, err := res.BuildFluentdConfigMap(instance, configName)
	if err != nil {
		r.Log.Error(err, "Failed to create ConfigMap")
		return reconcile.Result{}, err
	}
	found := &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: configName, Namespace: instance.Namespace}, found)
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
		// Fix so that cp4d can add inputs
		if !res.EqualConfig(found, expected, res.FluentdConfigKey) {
			found.Data[res.FluentdConfigKey] = expected.Data[res.FluentdConfigKey]
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
		if equal, missing := res.EqualSIEMConfig(instance, found); !equal {
			if missing {
				// Missing required config fields in cm
				err = r.Client.Delete(context.TODO(), found)
				if err != nil {
					r.Log.Error(err, "Failed to delete ConfigMap", "Name", found.Name)
					return reconcile.Result{}, err
				}
				r.Log.Info("[WARNING] Missing required fields. Recreating ConfigMap", "Name", configName)
				return reconcile.Result{Requeue: true}, nil
			}
			data := res.UpdateSIEMConfig(instance, found)
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
		err = r.restartFluentdPods(instance)
		if err != nil {
			r.Log.Error(err, "Failed to restart fluentd pods")
		}
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *CommonAuditReconciler) restartFluentdPods(instance *operatorv1.CommonAudit) error {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
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

func (r *CommonAuditReconciler) reconcileSecret(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	expected := res.BuildSecret(instance)
	found := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Secret
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Secret", "Secret.Namespace", expected.Namespace, "Secret.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Secret", "Secret.Namespace", expected.Namespace,
				"Secret.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Secret created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Secret")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *CommonAuditReconciler) reconcileFluentdDeployment(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	expected := res.BuildDeploymentForFluentd(instance)
	found := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDeploymentName, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Deployment
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Fluentd Deployment", "Deployment.Namespace", expected.Namespace, "Deployment.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Fluentd Deployment", "Deployment.Namespace", expected.Namespace,
				"Deployment.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	} else if !res.EqualDeployments(expected, found, false) {
		// If spec is incorrect, update it and requeue
		// Keep label changes
		tempPodLabels := found.Spec.Template.ObjectMeta.Labels
		found.Spec = expected.Spec
		found.Spec.Template.ObjectMeta.Labels = tempPodLabels
		err = r.Client.Update(context.TODO(), found)
		if err != nil {
			r.Log.Error(err, "Failed to update Deployment", "Namespace", instance.Namespace, "Name", found.Name)
			return reconcile.Result{}, err
		}
		r.Log.Info("Updating Fluentd Deployment", "Deployment.Name", found.Name)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *CommonAuditReconciler) reconcileCertPreReqs(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error
	recResult, recErr = r.reconcileIssuer(res.BuildGodIssuer, instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.reconcileRootCACert(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.reconcileIssuer(res.BuildRootCAIssuer, instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	return reconcile.Result{}, nil
}

func (r *CommonAuditReconciler) reconcileIssuer(f func(namespace string) *certmgr.Issuer, instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	expected := f(instance.Namespace)
	found := &certmgr.Issuer{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: expected.ObjectMeta.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Set instance as the owner and controller of the Issuer
		if err := controllerutil.SetControllerReference(instance, expected, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Issuer", "Issuer.Namespace", expected.Namespace, "Issuer.Name", expected.Name)
		err = r.Client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Issuer", "Issuer.Namespace", expected.Namespace,
				"Certificate.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Issuer created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Issuer")
		return reconcile.Result{}, err
	} else if result := res.EqualIssuers(expected, found); result {
		// If spec is incorrect, update it and requeue
		r.Log.Info("Found Issuer spec is incorrect", "Found", found.Spec, "Expected", expected.Spec)
		found.Spec = expected.Spec
		err = r.Client.Update(context.TODO(), found)
		if err != nil {
			r.Log.Error(err, "Failed to update Issuer", "Namespace", found.ObjectMeta.Namespace, "Name", found.Name)
			return reconcile.Result{}, err
		}
		r.Log.Info("Updating Issuer", "Issuer.Name", found.Name)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *CommonAuditReconciler) reconcileRootCACert(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	expectedCert := res.BuildRootCACert(instance.Namespace)
	foundCert := &certmgr.Certificate{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expectedCert.Name, Namespace: expectedCert.ObjectMeta.Namespace}, foundCert)
	if err != nil && errors.IsNotFound(err) {
		// Set instance as the owner and controller of the Certificate
		if err := controllerutil.SetControllerReference(instance, expectedCert, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		r.Log.Info("Creating a new Certificate", "Certificate.Namespace", expectedCert.Namespace, "Certificate.Name", expectedCert.Name)
		err = r.Client.Create(context.TODO(), expectedCert)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new Certificate", "Certificate.Namespace", expectedCert.Namespace,
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
		r.Log.Info("Updating Certificate", "Certificate.Name", foundCert.Name)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *CommonAuditReconciler) reconcileAuditCerts(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error
	recResult, recErr = r.reconcileAuditCertificate(instance, res.AuditLoggingHTTPSCertName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	return reconcile.Result{}, nil
}

func (r *CommonAuditReconciler) reconcileAuditCertificate(instance *operatorv1.CommonAudit, name string) (reconcile.Result, error) {
	issuer := res.RootIssuer
	if instance.Spec.Issuer != "" {
		issuer = instance.Spec.Issuer
	}
	expectedCert := res.BuildCertsForAuditLogging(instance.Namespace, issuer, name)
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

func (r *CommonAuditReconciler) reconcileServiceAccount(cr *operatorv1.CommonAudit) (reconcile.Result, error) {
	expectedRes := res.BuildServiceAccount(cr.Namespace)
	// Set CR instance as the owner and controller
	err := controllerutil.SetControllerReference(cr, expectedRes, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Failed to define expected resource")
		return reconcile.Result{}, err
	}

	// If ServiceAccount does not exist, create it and requeue
	foundSvcAcct := &corev1.ServiceAccount{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: expectedRes.Name, Namespace: cr.Namespace}, foundSvcAcct)
	if err != nil && errors.IsNotFound(err) {
		r.Log.Info("Creating a new ServiceAccount", "Namespace", cr.Namespace, "Name", expectedRes.Name)
		err = r.Client.Create(context.TODO(), expectedRes)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			r.Log.Error(err, "Failed to create new ServiceAccount", "Namespace", cr.Namespace, "Name", expectedRes.Name)
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

func (r *CommonAuditReconciler) reconcileRole(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	expected := res.BuildRole(instance.Namespace, false)
	found := &rbacv1.Role{}
	// Note: clusterroles are cluster-scoped, so this does not search using namespace (unlike other resources above)
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: instance.Namespace}, found)
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

func (r *CommonAuditReconciler) reconcileRoleBinding(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	expected := res.BuildRoleBinding(instance.Namespace)
	found := &rbacv1.RoleBinding{}
	// Note: clusterroles are cluster-scoped, so this does not search using namespace (unlike other resources above)
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: instance.Namespace}, found)
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
