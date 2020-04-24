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
	"crypto/rand"
	"math/big"
	"reflect"
	"testing"

	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-auditlogging-operator/pkg/resources"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const journalPath = "/var/log/audit"
const forwardingBool = true
const verbosity = "10"

// TestConfigConfig runs ReconcileOperandConfig.Reconcile() against a
// fake client that tracks a OperandConfig object.
func TestAuditLoggingController(t *testing.T) {
	// USE THIS
	logf.SetLogger(logf.ZapLogger(true))
	var (
		name = "example-auditlogging"
	)

	req := getReconcileRequest(name)
	cr := buildAuditLogging(name)
	r := getReconciler(cr)

	initReconcile(t, r, req)
	checkMountAndRBACPreReqs(t, r, req)
	checkPolicyControllerConfig(t, r, req)
	checkFluentdConfig(t, r, req, cr)
}

// Init reconcile the AuditLogging CR
func initReconcile(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) {
	assert := assert.New(t)
	result, err := r.Reconcile(req)
	assert.NoError(err)
	// Check the result of reconciliation to make sure it has the desired state.
	if !result.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}

func checkMountAndRBACPreReqs(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) {
	assert := assert.New(t)
	var err error
	// Check if ConfigMaps are created and have data
	foundCM := &corev1.ConfigMap{}
	configmaps := []string{
		res.FluentdDaemonSetName + "-" + res.ConfigName,
		res.FluentdDaemonSetName + "-" + res.SourceConfigName,
		res.FluentdDaemonSetName + "-" + res.SplunkConfigName,
		res.FluentdDaemonSetName + "-" + res.QRadarConfigName,
	}
	for _, cm := range configmaps {
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: cm, Namespace: res.InstanceNamespace}, foundCM)
		if err != nil {
			t.Fatalf("get configmap: (%v)", err)
		}
		_, err = r.Reconcile(req)
		assert.NoError(err)
	}

	// Check if Certs are created
	foundCert := &certmgr.Certificate{}
	certs := []string{res.AuditLoggingHTTPSCertName, res.AuditLoggingCertName}
	for _, c := range certs {
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: c, Namespace: res.InstanceNamespace}, foundCert)
		if err != nil {
			t.Fatalf("get cert: (%v)", err)
		}
		_, err = r.Reconcile(req)
		assert.NoError(err)
	}

	// Check if ServiceAccount is created
	foundSA := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.OperandRBAC, Namespace: res.InstanceNamespace}, foundSA)
	if err != nil {
		t.Fatalf("get service account: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)
}

func checkPolicyControllerConfig(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) {
	assert := assert.New(t)
	var err error

	// Check if ClusterRole and ClusterRoleBinding are created
	foundCR := &rbacv1.ClusterRole{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.AuditPolicyControllerDeploy + "-role"}, foundCR)
	if err != nil {
		t.Fatalf("get clusterrole: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)

	foundCRB := &rbacv1.ClusterRoleBinding{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.AuditPolicyControllerDeploy + "-rolebinding"}, foundCRB)
	if err != nil {
		t.Fatalf("get clusterrolebinding: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)

	// Check Audit Policy CRD is created
	foundCRD := &extv1beta1.CustomResourceDefinition{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.AuditPolicyCRDName}, foundCRD)
	if err != nil {
		t.Fatalf("get CRD: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)

	// Check if Policy Controller Deployment has been created and has the correct arguments
	foundDep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.AuditPolicyControllerDeploy, Namespace: res.InstanceNamespace}, foundDep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)
}

func checkFluentdConfig(t *testing.T, r ReconcileAuditLogging, req reconcile.Request, cr *operatorv1alpha1.AuditLogging) {
	assert := assert.New(t)
	var err error

	// Check if Role and Role Binding are created
	foundRole := &rbacv1.Role{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName + "-role", Namespace: res.InstanceNamespace}, foundRole)
	if err != nil {
		t.Fatalf("get role: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)

	foundRB := &rbacv1.RoleBinding{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName + "-rolebinding", Namespace: res.InstanceNamespace}, foundRB)
	if err != nil {
		t.Fatalf("get rolebinding: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)

	// Check if Service is created
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.AuditLoggingComponentName, Namespace: res.InstanceNamespace}, foundSvc)
	if err != nil {
		t.Fatalf("get service: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)

	// Check if fluentd DaemonSet is created
	foundDS := &appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName, Namespace: res.InstanceNamespace}, foundDS)
	if err != nil {
		t.Fatalf("get daemonset: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)

	// TODO
	updateAuditLoggingCR(t, r, req)

	// Create fake pods in namespace and collect their names to check against Status
	podLabels := res.LabelsForPodMetadata(res.FluentdName, cr.Name)
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: res.InstanceNamespace,
			Labels:    podLabels,
		},
	}
	podNames := make([]string, 3)
	for i := 0; i < 3; i++ {
		randInt, _ := rand.Int(rand.Reader, big.NewInt(99999))
		pod.ObjectMeta.Name = res.FluentdDaemonSetName + "-" + randInt.String()
		podNames[i] = pod.ObjectMeta.Name
		if err = r.client.Create(context.TODO(), pod.DeepCopy()); err != nil {
			t.Fatalf("create pod %d: (%v)", i, err)
		}
	}

	// Reconcile again so Reconcile() checks pods and updates the AuditLogging
	// resources' Status.
	result, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if result != (reconcile.Result{}) {
		t.Error("reconcile did not return an empty Result")
	}

	// Check status

	// Get the updated AuditLogging object.
	al := &operatorv1alpha1.AuditLogging{}
	err = r.client.Get(context.TODO(), req.NamespacedName, al)
	if err != nil {
		t.Errorf("get auditlogging: (%v)", err)
	}
	nodes := al.Status.Nodes
	if !reflect.DeepEqual(podNames, nodes) {
		t.Errorf("pod names %v did not match expected %v", nodes, podNames)
	}

	// TODO
	err = checkAuditLoggingCR(t, r, cr, forwardingBool, journalPath, verbosity)
	assert.NoError(err)
}

func updateAuditLoggingCR(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) {
}

func checkAuditLoggingCR(t *testing.T, r ReconcileAuditLogging, cr *operatorv1alpha1.AuditLogging, forward bool, journal string, verbosity string) error {
	return nil
}

func buildAuditLogging(name string) *operatorv1alpha1.AuditLogging {
	return &operatorv1alpha1.AuditLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: operatorv1alpha1.AuditLoggingSpec{
			Fluentd:          operatorv1alpha1.AuditLoggingSpecFluentd{},
			PolicyController: operatorv1alpha1.AuditLoggingSpecPolicyController{},
		},
	}
}

func getReconciler(cr *operatorv1alpha1.AuditLogging) ReconcileAuditLogging {
	s := scheme.Scheme
	operatorv1alpha1.SchemeBuilder.AddToScheme(s)
	certmgr.SchemeBuilder.AddToScheme(s)
	extv1beta1.AddToScheme(s)

	// Objects to track in the fake client.
	objs := []runtime.Object{
		cr,
	}

	// Create a fake client to mock API calls.
	client := fake.NewFakeClient(objs...)

	// Return a ReconcileOperandRequest object with the scheme and fake client.
	return ReconcileAuditLogging{
		scheme: s,
		client: client,
	}
}

// Mock request to simulate Reconcile() being called on an event for a watched resource
func getReconcileRequest(name string) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: name,
		},
	}
}
