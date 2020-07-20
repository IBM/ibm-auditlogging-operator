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

package commonaudit

import (
	"context"

	operatorv1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1"
	res "github.com/ibm/ibm-auditlogging-operator/pkg/resources"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileCommonAudit) reconcileService(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	reqLogger := log.WithValues("Service.Namespace", instance.Namespace, "instance.Name", instance.Name)
	expected := res.BuildAuditService(instance.Name, instance.Namespace)
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expected.Name, Namespace: instance.Namespace}, found)
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
		// If ports are incorrect, delete it and requeue
		reqLogger.Info("Found ports are incorrect", "Found", found.Spec.Ports, "Expected", expected.Spec.Ports)
		err = r.client.Delete(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to delete Service", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileCommonAudit) reconcileAuditConfigMaps(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
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

func (r *ReconcileCommonAudit) reconcileConfig(instance *operatorv1.CommonAudit, configName string) (reconcile.Result, error) {
	reqLogger := log.WithValues("ConfigMap.Namespace", instance.Namespace, "instance.Name", instance.Name)
	expected, err := res.BuildFluentdConfigMap(instance, configName)
	if err != nil {
		reqLogger.Error(err, "Failed to create ConfigMap")
		return reconcile.Result{}, err
	}
	found := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: configName, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new ConfigMap
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", expected.Namespace, "ConfigMap.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", expected.Namespace,
				"ConfigMap.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// ConfigMap created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get ConfigMap")
		return reconcile.Result{}, err
	}
	// ConfigMap was found, check for expected values
	var update = false
	if !res.EqualLabels(found.ObjectMeta.Labels, expected.ObjectMeta.Labels) {
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
	case res.FluentdDaemonSetName + "-" + res.SplunkConfigName:
		reqLogger.Info("Checking output configs")
		fallthrough
	case res.FluentdDaemonSetName + "-" + res.QRadarConfigName:
		// Ensure match tags are correct
		if !res.EqualMatchTags(found) {
			data := res.UpdateMatchTags(found)
			if configName == res.FluentdDaemonSetName+"-"+res.SplunkConfigName {
				found.Data[res.SplunkConfigKey] = data
			} else {
				found.Data[res.QRadarConfigKey] = data
			}
			update = true
		}
		if equal, missing := res.EqualSIEMConfig(instance, found); !equal {
			if missing {
				// Missing required config fields in cm
				err = r.client.Delete(context.TODO(), found)
				if err != nil {
					reqLogger.Error(err, "Failed to delete ConfigMap", "Name", found.Name)
					return reconcile.Result{}, err
				}
				reqLogger.Info("[WARNING] Missing required fields. Recreating ConfigMap", "Name", configName)
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
		reqLogger.Info("Unknown ConfigMap name", "Name", configName)
	}
	if update {
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update ConfigMap", "Name", found.Name)
			return reconcile.Result{}, err
		}
		// Updated - return and requeue
		reqLogger.Info("Updating ConfigMap", "ConfigMap.Name", found.Name)
		err = restartFluentdPods(r, instance)
		if err != nil {
			reqLogger.Error(err, "Failed to restart fluentd pods")
		}
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func restartFluentdPods(r *ReconcileCommonAudit, instance *operatorv1.CommonAudit) error {
	reqLogger := log.WithValues("func", "restartFluentdPods")
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.FluentdName, instance.Name)),
	}
	if err := r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods")
		return err
	}
	for _, pod := range podList.Items {
		p := pod
		err := r.client.Delete(context.TODO(), &p)
		if err != nil {
			reqLogger.Error(err, "Failed to delete pod", "Pod.Name", pod.Name)
			return err
		}
	}
	reqLogger.Info("Restarted pods", "Pods", res.GetPodNames(podList.Items))
	return nil
}

func (r *ReconcileCommonAudit) reconcileFluentdDeployment(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	reqLogger := log.WithValues("Deployment.Namespace", instance.Namespace, "instance.Name", instance.Name)
	expected := res.BuildDeploymentForFluentd(instance)
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDeploymentName, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Deployment
		if err := controllerutil.SetControllerReference(instance, expected, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Fluentd Deployment", "Deployment.Namespace", expected.Namespace, "Deployment.Name", expected.Name)
		err = r.client.Create(context.TODO(), expected)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new Fluentd Deployment", "Deployment.Namespace", expected.Namespace,
				"Deployment.Name", expected.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	} else if !res.EqualDeployments(expected, found) {
		// If spec is incorrect, update it and requeue
		found.ObjectMeta.Labels = expected.ObjectMeta.Labels
		found.Spec = expected.Spec
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Namespace", instance.Namespace, "Name", found.Name)
			return reconcile.Result{}, err
		}
		reqLogger.Info("Updating Fluentd Deployment", "Deployment.Name", found.Name)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileCommonAudit) reconcileAuditCerts(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	var recResult reconcile.Result
	var recErr error
	recResult, recErr = r.reconcileAuditCertificate(instance, res.AuditLoggingHTTPSCertName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	recResult, recErr = r.reconcileAuditCertificate(instance, res.AuditLoggingCertName)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileCommonAudit) reconcileAuditCertificate(instance *operatorv1.CommonAudit, name string) (reconcile.Result, error) {
	reqLogger := log.WithValues("Certificate.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	expectedCert := res.BuildCertsForAuditLogging(instance.Namespace, instance.Spec.ClusterIssuer, name)
	foundCert := &certmgr.Certificate{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: expectedCert.Name, Namespace: expectedCert.ObjectMeta.Namespace}, foundCert)
	if err != nil && errors.IsNotFound(err) {
		// Set Audit Logging instance as the owner and controller of the Certificate
		if err := controllerutil.SetControllerReference(instance, expectedCert, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Fluentd Certificate", "Certificate.Namespace", expectedCert.Namespace, "Certificate.Name", expectedCert.Name)
		err = r.client.Create(context.TODO(), expectedCert)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue.
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
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
		reqLogger.Info("Updating Fluentd Certificate", "Certificate.Name", foundCert.Name)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}
