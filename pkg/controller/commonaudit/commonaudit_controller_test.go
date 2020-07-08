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
	"crypto/rand"
	"math/big"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	operatorv1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1"
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

const dummyHost = "master"
const dummyPort = "8088"
const dummyToken = "abc-123"
const dummySplunkConfig = `
splunkHEC.conf: |-
     <match icp-audit k8s-audit>
        @type splunk_hec
        hec_host master
        hec_port 8089
        hec_token abc-123
        ca_file /fluentd/etc/tls/splunkCA.pem
        source ${tag}
        <buffer>
          # ...
        </buffer>
     </match>`
const dummyFluentdSHA = "sha256:abc"
const dummyHostAliasIP = "9.12.34.56"
const dummyHostAliasName = "test.fyre.ibm.com"
const namespace = "test"

var dummyHostAliases = []corev1.HostAlias{
	{
		IP:        dummyHostAliasIP,
		Hostnames: []string{dummyHostAliasName},
	},
}
var replicas = 3

// TestConfigConfig runs ReconcileOperandConfig.Reconcile() against a
// fake client that tracks a OperandConfig object.
func TestCommonAuditController(t *testing.T) {
	// USE THIS
	// logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	// logf.SetLogger(logf.ZapLogger(true))
	var (
		name = "example-commonaudit"
	)

	os.Setenv(res.FluentdEnvVar, dummyFluentdSHA)
	os.Setenv("WATCH_NAMESPACE", "")

	req := getReconcileRequest(name)
	cr := buildCommonAudit(name)
	r := getReconciler(cr)

	reconcileResources(t, r, req, true)
	checkMountAndRBACPreReqs(t, r, req, cr)
	checkFluentdConfig(t, r, req, cr)
	al := getAuditLogging(t, r, req)
	updateAuditLoggingCR(al, t, r, req)
	checkInPlaceUpdate(t, r, req, cr)
}

func checkMountAndRBACPreReqs(t *testing.T, r ReconcileCommonAudit, req reconcile.Request, cr *operatorv1.CommonAudit) {
	var err error
	// Check if ConfigMaps are created and have data
	foundCM := &corev1.ConfigMap{}
	for _, cm := range res.FluentdConfigMaps {
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: cm, Namespace: cr.Namespace}, foundCM)
		if err != nil {
			t.Fatalf("get configmap: (%v), namespace: (%s)", err, cr.Namespace)
		}
		reconcileResources(t, r, req, true)
	}

	// Check if Certs are created
	foundCert := &certmgr.Certificate{}
	certs := []string{res.AuditLoggingHTTPSCertName, res.AuditLoggingCertName}
	for _, c := range certs {
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: c, Namespace: cr.Namespace}, foundCert)
		if err != nil {
			t.Fatalf("get cert: (%v)", err)
		}
		if foundCert.Spec.IssuerRef.Name != res.DefaultClusterIssuer {
			t.Fatalf("incorrect clusterissuer. Found: (%s), Expected: (%s).", foundCert.Spec.IssuerRef.Name, res.DefaultClusterIssuer)
		}
		reconcileResources(t, r, req, true)
	}

	// Check if ServiceAccount is created
	foundSA := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.OperandServiceAccount, Namespace: cr.Namespace}, foundSA)
	if err != nil {
		t.Fatalf("get service account: (%v)", err)
	}
	reconcileResources(t, r, req, true)
}

func checkFluentdConfig(t *testing.T, r ReconcileCommonAudit, req reconcile.Request, cr *operatorv1.CommonAudit) {
	var err error

	// Check if Role and Role Binding are created
	foundRole := &rbacv1.Role{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName + "-role", Namespace: cr.Namespace}, foundRole)
	if err != nil {
		t.Fatalf("get role: (%v)", err)
	}
	reconcileResources(t, r, req, true)

	foundRB := &rbacv1.RoleBinding{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName + "-rolebinding", Namespace: cr.Namespace}, foundRB)
	if err != nil {
		t.Fatalf("get rolebinding: (%v)", err)
	}
	reconcileResources(t, r, req, true)

	// Check if Service is created
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.AuditLoggingComponentName, Namespace: cr.Namespace}, foundSvc)
	if err != nil {
		t.Fatalf("get service: (%v)", err)
	}
	reconcileResources(t, r, req, true)

	// Check if fluentd deployment is created
	dep := getFluentd(t, r, cr)
	image := res.DefaultImageRegistry + res.DefaultFluentdImageName + "@" + dummyFluentdSHA
	if dep.Spec.Template.Spec.Containers[0].Image != image {
		t.Fatalf("Incorrect fluentd image. Found: (%s), Expected: (%s)", dep.Spec.Template.Spec.Containers[0].Image, image)
	}
	reconcileResources(t, r, req, false)

	// Create fake pods in namespace and collect their names to check against Status
	var podLabels = res.LabelsForPodMetadata(res.FluentdName, cr.Name)
	var pod = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Labels:    podLabels,
		},
	}
	podNames := make([]string, replicas)
	var randInt *big.Int
	var i int
	for i = 0; i < replicas; i++ {
		randInt, _ = rand.Int(rand.Reader, big.NewInt(99999))
		pod.ObjectMeta.Name = res.FluentdDeploymentName + "-" + randInt.String()
		podNames[i] = pod.ObjectMeta.Name
		if err = r.client.Create(context.TODO(), pod.DeepCopy()); err != nil {
			t.Fatalf("create pod %d: (%v)", i, err)
		}
	}
	// Reconcile again so Reconcile() checks pods and updates the AuditLogging
	// resources' Status.
	reconcileResources(t, r, req, false)

	// Check status

	// Get the updated AuditLogging object.
	al := getAuditLogging(t, r, req)
	nodes := al.Status.Nodes
	sort.Strings(podNames)
	if !reflect.DeepEqual(podNames, nodes) {
		t.Errorf("pod names %v did not match expected %v", nodes, podNames)
	}
}

func updateAuditLoggingCR(al *operatorv1.CommonAudit, t *testing.T, r ReconcileCommonAudit, req reconcile.Request) {
	al.Spec.Fluentd.Output.Splunk.Host = dummyHost
	al.Spec.Fluentd.Output.Splunk.Token = dummyToken
	al.Spec.Fluentd.Output.Splunk.Port, _ = strconv.Atoi(dummyPort)
	al.Spec.Fluentd.Output.HostAliases = append(al.Spec.Fluentd.Output.HostAliases, operatorv1.CommonAuditSpecHostAliases{
		HostIP: dummyHostAliasIP, Hostnames: []string{dummyHostAliasName},
	})
	err := r.client.Update(context.TODO(), al)
	if err != nil {
		t.Fatalf("Failed to update CR: (%v)", err)
	}
	// update resources
	reconcileResources(t, r, req, true)
}

func checkInPlaceUpdate(t *testing.T, r ReconcileCommonAudit, req reconcile.Request, cr *operatorv1.CommonAudit) {
	var err error
	var emptyLabels map[string]string
	assert := assert.New(t)

	foundCM := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName + "-" +
		res.SplunkConfigName, Namespace: cr.Namespace}, foundCM)
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
	reconcileResources(t, r, req, true)

	updatedCM := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDaemonSetName + "-" +
		res.SplunkConfigName, Namespace: cr.Namespace}, updatedCM)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	reHost := regexp.MustCompile(`hec_host .*`)
	host := strings.Split(reHost.FindStringSubmatch(updatedCM.Data[res.SplunkConfigKey])[0], " ")[1]
	rePort := regexp.MustCompile(`hec_port .*`)
	port := strings.Split(rePort.FindStringSubmatch(updatedCM.Data[res.SplunkConfigKey])[0], " ")[1]
	reToken := regexp.MustCompile(`hec_token .*`)
	token := strings.Split(reToken.FindStringSubmatch(updatedCM.Data[res.SplunkConfigKey])[0], " ")[1]
	reBuffer := regexp.MustCompile(`buffer`)
	buffer := reBuffer.FindAllString(updatedCM.Data[res.SplunkConfigKey], -1)
	if host != dummyHost || port != dummyPort || token != dummyToken {
		t.Fatalf("SIEM creds not preserved: Found: (%s), (%s), (%s). Expected: (%s), (%s), (%s). ",
			host, port, token, dummyHost, dummyPort, dummyToken)
	}
	if !reflect.DeepEqual(updatedCM.ObjectMeta.Labels, res.LabelsForMetadata(res.FluentdName)) {
		t.Fatalf("Labels not correct")
	}
	if len(buffer) < 2 {
		t.Fatalf("Buffer config not preserved. Found: (%s)", buffer)
	}

	// add hostaliases
	fluentd := getFluentd(t, r, cr)
	// fluentd.Spec.Template.Spec.HostAliases = dummyHostAliases
	// trigger found != expected
	fluentd.ObjectMeta.Labels = map[string]string{}
	err = r.client.Update(context.TODO(), fluentd)
	if err != nil {
		t.Fatalf("Failed to update fluentd daemonset: (%v)", err)
	}
	reconcileResources(t, r, req, true)

	fluentd = getFluentd(t, r, cr)
	if !reflect.DeepEqual(fluentd.Spec.Template.Spec.HostAliases, dummyHostAliases) {
		t.Fatalf("HostAliases not saved. Found: (%v). Expected: (%v)", fluentd.Spec.Template.Spec.HostAliases, dummyHostAliases)
	}
}

func getAuditLogging(t *testing.T, r ReconcileCommonAudit, req reconcile.Request) *operatorv1.CommonAudit {
	al := &operatorv1.CommonAudit{}
	err := r.client.Get(context.TODO(), req.NamespacedName, al)
	if err != nil {
		t.Fatalf("get commonaudit: (%v)", err)
	}
	return al
}

func getFluentd(t *testing.T, r ReconcileCommonAudit, cr *operatorv1.CommonAudit) *appsv1.Deployment {
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.FluentdDeploymentName, Namespace: cr.Namespace}, found)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
	return found
}

func buildCommonAudit(name string) *operatorv1.CommonAudit {
	return &operatorv1.CommonAudit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: operatorv1.CommonAuditSpec{},
	}
}

func getReconciler(cr *operatorv1.CommonAudit) ReconcileCommonAudit {
	s := scheme.Scheme
	operatorv1.SchemeBuilder.AddToScheme(s)
	certmgr.SchemeBuilder.AddToScheme(s)
	extv1beta1.AddToScheme(s)

	// Objects to track in the fake client.
	objs := []runtime.Object{
		cr,
	}

	// Create a fake client to mock API calls.
	client := fake.NewFakeClient(objs...)

	// Return a ReconcileOperandRequest object with the scheme and fake client.
	return ReconcileCommonAudit{
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

func reconcileResources(t *testing.T, r ReconcileCommonAudit, req reconcile.Request, requeue bool) {
	assert := assert.New(t)
	result, err := r.Reconcile(req)
	if requeue {
		// Check the result of reconciliation to make sure it has the desired state.
		if !result.Requeue {
			t.Error("reconcile did not requeue request as expected")
		}
	}
	assert.NoError(err)
}
