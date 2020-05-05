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
	"regexp"
	"testing"

	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-auditlogging-operator/pkg/resources"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
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
)

const journalPath = "/var/log/audit"
const verbosity = "10"
const numPods = 4
const dummyHost = "hec_host master"
const dummyPort = "hec_port 8088"
const dummyToken = "hec_token abc-123"
const dummySplunkConfig = `
splunkHEC.conf: |-
     <match icp-audit k8s-audit>
        @type splunk_hec
        hec_host master
        hec_port 8088
        hec_token abc-123
        ca_file /fluentd/etc/tls/splunkCA.pem
        source ${tag}
     </match>`

// TestConfigConfig runs ReconcileOperandConfig.Reconcile() against a
// fake client that tracks a OperandConfig object.
func TestAuditLoggingController(t *testing.T) {
	// USE THIS
	// logf.SetLogger(logf.ZapLogger(true))
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
	checkInPlaceUpdate(t, r, req)
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
	getAuditPolicyController(t, r, req)
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
	getFluentd(t, r, req)

	// Create fake pods in namespace and collect their names to check against Status
	var podLabels = res.LabelsForPodMetadata(res.FluentdName, cr.Name)
	var pod = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: res.InstanceNamespace,
			Labels:    podLabels,
		},
	}
	podNames := make([]string, numPods)
	var randInt *big.Int
	var i int
	for i = 0; i < numPods-1; i++ {
		randInt, _ = rand.Int(rand.Reader, big.NewInt(99999))
		pod.ObjectMeta.Name = res.FluentdDaemonSetName + "-" + randInt.String()
		podNames[i] = pod.ObjectMeta.Name
		if err = r.client.Create(context.TODO(), pod.DeepCopy()); err != nil {
			t.Fatalf("create pod %d: (%v)", i, err)
		}
	}
	podLabels = res.LabelsForPodMetadata(res.AuditPolicyControllerDeploy, cr.Name)
	pod = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: res.InstanceNamespace,
			Labels:    podLabels,
		},
	}
	randInt, _ = rand.Int(rand.Reader, big.NewInt(99999))
	pod.ObjectMeta.Name = res.AuditPolicyControllerDeploy + "-" + randInt.String()
	podNames[i] = pod.ObjectMeta.Name
	if err = r.client.Create(context.TODO(), pod.DeepCopy()); err != nil {
		t.Fatalf("create pod %d: (%v)", i, err)
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
	al := getAuditLogging(t, r, req)
	nodes := al.Status.Nodes
	if !reflect.DeepEqual(podNames, nodes) {
		t.Errorf("pod names %v did not match expected %v", nodes, podNames)
	}

	updateAuditLoggingCR(al, t, r, req)
	checkAuditLogging(t, r, req)
}

func updateAuditLoggingCR(al *operatorv1alpha1.AuditLogging, t *testing.T, r ReconcileAuditLogging, req reconcile.Request) {
	assert := assert.New(t)
	al.Spec.Fluentd.JournalPath = journalPath
	al.Spec.PolicyController.Verbosity = verbosity
	err := r.client.Update(context.TODO(), al)
	if err != nil {
		t.Fatalf("Failed to update CR: (%v)", err)
	}
	// update resources
	result, err := r.Reconcile(req)
	if !result.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	assert.NoError(err)
}

func checkAuditLogging(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) {
	policyController := getAuditPolicyController(t, r, req)
	fluentd := getFluentd(t, r, req)
	var found = false
	for _, arg := range policyController.Spec.Template.Spec.Containers[0].Args {
		if arg == "--v="+verbosity {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Policy controller not updated with verbosity: (%s)", verbosity)
	}
	found = false
	for _, v := range fluentd.Spec.Template.Spec.Containers[0].VolumeMounts {
		if v.MountPath == journalPath {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Fluentd ds not updated with journal path: (%s)", journalPath)
	}
}

func checkInPlaceUpdate(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) {
	var err error
	var emptyLabels map[string]string
	assert := assert.New(t)

	foundCM := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName + "-" +
		res.SplunkConfigName, Namespace: res.InstanceNamespace}, foundCM)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	var ds res.DataSplunk
	err = yaml.Unmarshal([]byte(dummySplunkConfig), &ds)
	assert.NoError(err)
	foundCM.Data[res.SplunkConfigKey] = ds.Value
	foundCM.ObjectMeta.Labels = emptyLabels
	err = r.client.Update(context.TODO(), foundCM)
	if err != nil {
		t.Fatalf("Failed to update CR: (%v)", err)
	}
	// update splunk configmap
	result, err := r.Reconcile(req)
	if !result.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	assert.NoError(err)

	updatedCM := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName + "-" +
		res.SplunkConfigName, Namespace: res.InstanceNamespace}, updatedCM)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	reHost := regexp.MustCompile(`hec_host .*`)
	host := reHost.FindStringSubmatch(updatedCM.Data[res.SplunkConfigKey])[0]
	rePort := regexp.MustCompile(`hec_port .*`)
	port := rePort.FindStringSubmatch(updatedCM.Data[res.SplunkConfigKey])[0]
	reToken := regexp.MustCompile(`hec_token .*`)
	token := reToken.FindStringSubmatch(updatedCM.Data[res.SplunkConfigKey])[0]
	if host != dummyHost || port != dummyPort || token != dummyToken {
		t.Fatalf("SIEM creds not retained: found host = (%s), found port = (%s), found token = (%s)", host, port, token)
	}
	if !reflect.DeepEqual(updatedCM.ObjectMeta.Labels, res.LabelsForMetadata(res.FluentdName)) {
		t.Fatalf("Labels not correct")
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)
}

func getAuditLogging(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) *operatorv1alpha1.AuditLogging {
	assert := assert.New(t)
	al := &operatorv1alpha1.AuditLogging{}
	err := r.client.Get(context.TODO(), req.NamespacedName, al)
	if err != nil {
		t.Fatalf("get auditlogging: (%v)", err)
	}
	assert.NoError(err)
	return al
}

func getAuditPolicyController(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) *appsv1.Deployment {
	assert := assert.New(t)
	foundDep := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.AuditPolicyControllerDeploy, Namespace: res.InstanceNamespace}, foundDep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)
	return foundDep
}

func getFluentd(t *testing.T, r ReconcileAuditLogging, req reconcile.Request) *appsv1.DaemonSet {
	assert := assert.New(t)
	foundDS := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName, Namespace: res.InstanceNamespace}, foundDS)
	if err != nil {
		t.Fatalf("get daemonset: (%v)", err)
	}
	_, err = r.Reconcile(req)
	assert.NoError(err)
	return foundDS
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
