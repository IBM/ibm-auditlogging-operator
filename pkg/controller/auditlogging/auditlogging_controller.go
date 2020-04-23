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
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-auditlogging-operator/pkg/resources"

	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

var log = logf.Log.WithName("controller_auditlogging")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AuditLogging Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAuditLogging{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("auditlogging-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AuditLogging
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.AuditLogging{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Deployments and requeue the owner AuditLogging
	secondaryResourceTypes := []runtime.Object{
		&appsv1.DaemonSet{},
		&appsv1.Deployment{},
		&corev1.ConfigMap{},
		&certmgr.Certificate{},
		&corev1.ServiceAccount{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&extv1beta1.CustomResourceDefinition{},
		&corev1.Service{},
	}
	for _, restype := range secondaryResourceTypes {
		log.Info("Watching", "restype", reflect.TypeOf(restype))
		//err = c.Watch(&kind, &handler.EnqueueRequestForOwner{
		err = c.Watch(&source.Kind{Type: restype}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &operatorv1alpha1.AuditLogging{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileAuditLogging implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAuditLogging{}

// ReconcileAuditLogging reconciles a AuditLogging object
type ReconcileAuditLogging struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AuditLogging object and makes changes based on the state read
// and what is in the AuditLogging.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAuditLogging) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AuditLogging")
	// if we need to create several resources, set a flag so we just requeue one time instead of after each create.
	// Fetch the AuditLogging instance
	instance := &operatorv1alpha1.AuditLogging{}
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

	// Reconcile the expected configmaps
	recResult, recErr = r.createOrUpdateAuditConfigMaps(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected cert
	recResult, recErr = r.createOrUpdateAuditCerts(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected ServiceAccount for operands
	recResult, recErr = r.createOrUpdateServiceAccount(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected Role
	recResult, recErr = r.createOrUpdateClusterRole(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected RoleBinding
	recResult, recErr = r.createOrUpdateClusterRoleBinding(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the AuditPolicy CRD
	recResult, recErr = r.createAuditPolicyCRD(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected bridge deployment
	recResult, recErr = r.createOrUpdatePolicyControllerDeployment(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected Role
	recResult, recErr = r.createOrUpdateRole(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected RoleBinding
	recResult, recErr = r.createOrUpdateRoleBinding(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected Service
	recResult, recErr = r.createOrUpdateService(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected daemonset
	recResult, recErr = r.createOrUpdateFluentdDaemonSet(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Prior to version 3.6, audit-logging used two seperate service accounts.
	// Delete service accounts if they were leftover from a previous version.
	r.checkOldServiceAccounts(instance)

	// Reconcile the expected Status
	recResult, recErr = r.updateStatus(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	reqLogger.Info("Reconciliation successful!", "Name", instance.Name)
	// since we updated the status in the Audit Logging CR, sleep 5 seconds to allow the CR to be refreshed.
	time.Sleep(5 * time.Second)

	return reconcile.Result{}, nil
}
