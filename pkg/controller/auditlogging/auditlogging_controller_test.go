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
	"testing"
	"time"

	"github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-auditlogging-operator/pkg/resources"
	olmv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/apis/scorecard/v1alpha2"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TestConfigConfig runs ReconcileOperandConfig.Reconcile() against a
// fake client that tracks a OperandConfig object.
func TestAuditLoggingController(t *testing.T) {
	var (
		name      = "auditlogging"
		namespace = res.InstanceNamespace
	)

	req := getReconcileRequest(name, namespace)
	r := getReconciler(name, namespace)
	cr := &operatorv1alpha1.AuditLogging{}

	initReconcile(t, r, req, cr)

	// check policy controller deployment, fluentd ds, and service

	updateAuditLoggingCR(t, r, req)

	deleteAuditLoggingCR(t, r, req, cr)
}

// Init reconcile the OperandRequest
func initReconcile(t *testing.T, r ReconcileAuditLogging, req reconcile.Request, cr *operatorv1alpha1.AuditLogging) {
	assert := assert.New(t)
	res, err := r.Reconcile(req)
	if res.Requeue {
		t.Error("Reconcile requeued request as not expected")
	}
	assert.NoError(err)
	// cfg, err := config.GetConfig()
	// assert.NoError(err)
	// workers := len(cfg.)

	// Retrieve AuditLogging CR
	retrieveAuditLogging(t, r, req, cr, 3)

	// TODO
	err = checkAuditLoggingCR(t, r, cr, true, "/var/log/audit", "10")
	assert.NoError(err)
}

// Retrieve AuditLogging instance
func retrieveAuditLogging(t *testing.T, r ReconcileAuditLogging, req reconcile.Request, cr *operatorv1alpha1.AuditLogging, expectedPodNum int) {
	assert := assert.New(t)
	err := r.client.Get(context.TODO(), req.NamespacedName, cr)
	assert.NoError(err)
	assert.Equal(expectedPodNum, len(cr.Status.Nodes), "audit logging status node list should have three elements")
}

func updateAuditLoggingCR(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) {
}

func checkAuditLoggingCR(t *testing.T, r ReconcileAuditLogging, cr *operatorv1alpha1.AuditLogging, forward bool, journal string, verbosity string) error {
	return nil
}

// Mock delete AuditLogging instance, mark AuditLogging instance as delete state
func deleteAuditLoggingCR(t *testing.T, r ReconcileAuditLogging, req reconcile.Request, cr *operatorv1alpha1.AuditLogging) {
	assert := assert.New(t)
	deleteTime := metav1.NewTime(time.Now())
	cr.SetDeletionTimestamp(&deleteTime)
	err := r.client.Update(context.TODO(), cr)
	assert.NoError(err)
	res, _ := r.Reconcile(req)
	if res.Requeue {
		t.Error("Reconcile requeued request as not expected")
	}

	// Check if AuditLogging instance deleted
	err = r.client.Delete(context.TODO(), cr)
	assert.NoError(err)
	err = r.client.Get(context.TODO(), req.NamespacedName, cr)
	assert.True(errors.IsNotFound(err), "retrieve audit logging should be return an error of type is 'NotFound'")
}

type DataObj struct {
	objs []runtime.Object
}

func initClientData(name, namespace string) *DataObj {
	return &DataObj{
		objs: []runtime.Object{
			auditLogging(name, namespace),
		},
	}
}

func auditLogging(name, namespace string) *operatorv1alpha1.AuditLogging {
	return &operatorv1alpha1.AuditLogging{}
}

func getReconciler(name, namespace string) ReconcileAuditLogging {
	s := scheme.Scheme
	v1alpha1.SchemeBuilder.AddToScheme(s)
	olmv1.SchemeBuilder.AddToScheme(s)
	olmv1alpha1.SchemeBuilder.AddToScheme(s)
	v1beta2.SchemeBuilder.AddToScheme(s)
	v1alpha2.SchemeBuilder.AddToScheme(s)

	initData := initClientData(name, namespace)

	// Create a fake client to mock API calls.
	client := fake.NewFakeClient(initData.objs...)

	// Return a ReconcileOperandRequest object with the scheme and fake client.
	return ReconcileAuditLogging{
		scheme: s,
		client: client,
	}
}

// Mock request to simulate Reconcile() being called on an event for a watched resource
func getReconcileRequest(name, namespace string) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
}
