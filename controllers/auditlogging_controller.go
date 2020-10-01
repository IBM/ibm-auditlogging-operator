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
	"reflect"

	"sort"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/IBM/ibm-auditlogging-operator/controllers/util"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"
	operatorv1alpha1 "github.com/IBM/ibm-auditlogging-operator/api/v1alpha1"
	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	opversion "github.com/IBM/ibm-auditlogging-operator/version"
)

// AuditLoggingReconciler reconciles a AuditLogging object
type AuditLoggingReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *AuditLoggingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithName("controller_auditlogging").WithValues("request", req.NamespacedName)

	instance := &operatorv1alpha1.AuditLogging{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Set a default Status value
	if len(instance.Status.Nodes) == 0 {
		instance.Status.Nodes = util.DefaultStatusForCR
		instance.Status.Versions.Reconciled = opversion.Version
		err = r.Client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.Log.Error(err, "Failed to set AuditLogging default status")
			return reconcile.Result{}, err
		}
	}

	auditLoggingList := &operatorv1alpha1.AuditLoggingList{}
	if err := r.Client.List(context.TODO(), auditLoggingList); err == nil && len(auditLoggingList.Items) > 1 {
		msg := "Only one instance of AuditLogging per cluster. Delete other instances to proceed."
		r.Log.Info(msg)
		r.updateEvent(instance, msg, corev1.EventTypeWarning, "Not Allowed")
		// Return and don't requeue
		return reconcile.Result{}, nil
	}

	commonAuditList := &operatorv1.CommonAuditList{}
	if err := r.Client.List(context.TODO(), commonAuditList, client.InNamespace(constant.InstanceNamespace)); err == nil &&
		len(commonAuditList.Items) > 0 {
		msg := "CommonAudit cannot run alongside AuditLogging in the same namespace. Delete one or the other to proceed."
		r.Log.Info(msg)
		r.updateEvent(instance, msg, corev1.EventTypeWarning, "Not Allowed")
		// Return and don't requeue
		return reconcile.Result{}, nil
	}

	r.updateEvent(instance, "Instance found", corev1.EventTypeNormal, "Initializing")

	var recResult reconcile.Result
	var recErr error
	reconcilers := []func(*operatorv1alpha1.AuditLogging) (reconcile.Result, error){
		r.removeOldPolicyControllerDeploy,
		r.reconcileAuditConfigMaps,
		r.reconcileAuditCerts,
		r.reconcileServiceAccount,
		r.reconcileRole,
		r.reconcileRoleBinding,
		r.reconcileService,
		r.reconcileFluentdDaemonSet,
		r.updateStatus,
	}
	for _, rec := range reconcilers {
		recResult, recErr = rec(instance)
		if recErr != nil || recResult.Requeue {
			return recResult, recErr
		}
	}

	// Prior to version 3.6, audit-logging used two separate service accounts.
	// Delete service accounts if they were leftover from a previous version.
	// Policy controller deployment has been moved to operator pod in 3.7, remove redundant rbac
	r.removeOldRBAC()
	r.updateEvent(instance, "Deployed "+constant.AuditLoggingComponentName+" successfully", corev1.EventTypeNormal, "Deployed")
	r.Log.Info("Reconciliation successful!", "Name", instance.Name)
	// since we updated the status in the Audit Logging CR, sleep 5 seconds to allow the CR to be refreshed.
	time.Sleep(5 * time.Second)

	return ctrl.Result{}, nil
}

func (r *AuditLoggingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AuditLogging{}).
		Owns(&appsv1.DaemonSet{}).Owns(&corev1.ConfigMap{}).Owns(&certmgr.Certificate{}).Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).Owns(&rbacv1.RoleBinding{}).Owns(&corev1.Service{}).
		Complete(r)
}

func (r *AuditLoggingReconciler) updateStatus(instance *operatorv1alpha1.AuditLogging) (reconcile.Result, error) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(constant.InstanceNamespace),
		client.MatchingLabels(util.LabelsForSelector(constant.FluentdName, instance.Name)),
	}
	if err := r.Client.List(context.TODO(), podList, listOpts...); err != nil {
		r.Log.Error(err, "Failed to list pods", "AuditLogging.Namespace", constant.InstanceNamespace, "AuditLogging.Name", instance.Name)
		return reconcile.Result{}, err
	}
	var podNames []string
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}
	// if no pods were found set the default status
	if len(podNames) == 0 {
		podNames = util.DefaultStatusForCR
	} else {
		sort.Strings(podNames)
	}

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) || opversion.Version != instance.Status.Versions.Reconciled {
		instance.Status.Nodes = podNames
		instance.Status.Versions.Reconciled = opversion.Version
		r.Log.Info("Updating Audit Logging status", "Name", instance.Name)
		err := r.Client.Status().Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *AuditLoggingReconciler) updateEvent(instance *operatorv1alpha1.AuditLogging, message, event, reason string) {
	r.Recorder.Event(instance, event, reason, message)
}
