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
	"reflect"
	"sort"
	"time"

	operatorv1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1"
	res "github.com/ibm/ibm-auditlogging-operator/pkg/resources"

	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_commonaudit")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new CommonAudit Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCommonAudit{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("commonaudit-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CommonAudit
	err = c.Watch(&source.Kind{Type: &operatorv1.CommonAudit{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner CommonAudit
	secondaryResourceTypes := []runtime.Object{
		&appsv1.Deployment{},
		&corev1.ConfigMap{},
		&certmgr.Certificate{},
		&corev1.ServiceAccount{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
		&corev1.Service{},
	}
	for _, restype := range secondaryResourceTypes {
		log.Info("Watching", "restype", reflect.TypeOf(restype))
		//err = c.Watch(&kind, &handler.EnqueueRequestForOwner{
		err = c.Watch(&source.Kind{Type: restype}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &operatorv1.CommonAudit{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileCommonAudit implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCommonAudit{}

// ReconcileCommonAudit reconciles a CommonAudit object
type ReconcileCommonAudit struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a CommonAudit object and makes changes based on the state read
// and what is in the CommonAudit.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCommonAudit) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CommonAudit")

	// Fetch the CommonAudit instance
	instance := &operatorv1.CommonAudit{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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
		instance.Status.Nodes = res.DefaultStatusForCR
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to set AuditLogging default status")
			return reconcile.Result{}, err
		}
	}

	var recResult reconcile.Result
	var recErr error

	reconcilers := []func(*operatorv1.CommonAudit) (reconcile.Result, error){
		r.reconcileAuditConfigMaps,
		r.reconcileAuditCerts,
		r.reconcileServiceAccount,
		r.reconcileRole,
		r.reconcileRoleBinding,
		r.reconcileService,
		r.reconcileFluentdDeployment,
		r.updateStatus,
		r.checkIfBeingDeleted,
	}
	for _, rec := range reconcilers {
		recResult, recErr = rec(instance)
		if recErr != nil || recResult.Requeue {
			return recResult, recErr
		}
	}

	reqLogger.Info("Reconciliation successful!", "Name", instance.Name)
	// since we updated the status in the Audit Logging CR, sleep 5 seconds to allow the CR to be refreshed.
	time.Sleep(5 * time.Second)

	return reconcile.Result{}, nil
}

func (r *ReconcileCommonAudit) updateStatus(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	reqLogger := log.WithValues("Namespace", instance.Namespace, "Name", instance.Name)

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.FluentdName, instance.Name)),
	}
	if err := r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "AuditLogging.Namespace", instance.Namespace, "AuditLogging.Name", instance.Name)
		return reconcile.Result{}, err
	}
	podNames := []string{}
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	sort.Strings(podNames)

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

func (r *ReconcileCommonAudit) checkIfBeingDeleted(instance *operatorv1.CommonAudit) (reconcile.Result, error) {
	reqLogger := log.WithValues("func", "checkIfBeingDeleted")
	// Credit: kubebuilder book
	finalizerName := "commonaudit.operator.ibm.com"
	// Determine if the AuditLogging CR is going to be deleted
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object not being deleted, but add our finalizer so we know to remove this object later when it is going to be deleted
		if !res.ContainsString(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				reqLogger.Error(err, "Error adding the finalizer to the CR")
				return reconcile.Result{}, err
			}
		}
	} else {
		// Object scheduled to be deleted
		if res.ContainsString(instance.ObjectMeta.Finalizers, finalizerName) {
			if err := r.deleteExternalResources(); err != nil {
				reqLogger.Error(err, "Error deleting resources created by this operator")
				return reconcile.Result{}, err
			}
			instance.ObjectMeta.Finalizers = res.RemoveString(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				reqLogger.Error(err, "Error updating the CR to remove the finalizer")
				return reconcile.Result{}, err
			}
			reqLogger.Info("Successfully deleted external resources")
		}
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileCommonAudit) deleteExternalResources() error {
	// Remove CustomResourceDefinition
	if err := removeCRD(r.client, res.AuditPolicyCRDName); err != nil {
		return err
	}
	return nil
}
