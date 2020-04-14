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
	"strconv"
	"testing"
	"time"

	"github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-auditlogging-operator/pkg/resources"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const journalPath = "/var/log/audit"
const forwardingBool = true
const verbosity = "10"

// TestConfigConfig runs ReconcileOperandConfig.Reconcile() against a
// fake client that tracks a OperandConfig object.
func TestAuditLoggingController(t *testing.T) {
	var (
		name      = "example-auditlogging"
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

	nodeList := &corev1.NodeList{}
	t.Log(nodeList)
	opts := []client.ListOption{
		client.MatchingLabels{"kubernetes.io/hostname": "worker1.audit-op.os.fyre.ibm.com"},
	}
	ctx := context.TODO()
	err = r.client.List(ctx, nodeList, opts...)
	assert.NoError(err)

	// Retrieve AuditLogging CR
	retrieveAuditLogging(t, r, req, cr, len(nodeList.Items)+1)

	// TODO
	err = checkAuditLoggingCR(t, r, cr, forwardingBool, journalPath, verbosity)
	assert.NoError(err)
}

// Retrieve AuditLogging instance
func retrieveAuditLogging(t *testing.T, r ReconcileAuditLogging, req reconcile.Request, cr *operatorv1alpha1.AuditLogging, expectedPodNum int) {
	assert := assert.New(t)
	err := r.client.Get(context.TODO(), req.NamespacedName, cr)
	assert.NoError(err)
	assert.Equal(expectedPodNum, len(cr.Status.Nodes), "Audit logging status node list should have "+strconv.Itoa(expectedPodNum)+" elements")
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
	corev1.SchemeBuilder.AddToScheme(s)

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
