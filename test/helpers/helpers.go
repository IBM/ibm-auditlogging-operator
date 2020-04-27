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
package helpers

import (
	goctx "context"
	"fmt"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operator "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	"github.com/ibm/ibm-auditlogging-operator/test/config"
)

// CreateAuditLogging creates a AuditLogging instance
func CreateAuditLogging(f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	// Create AuditLogging instance
	fmt.Println("--- CREATE: AuditLogging Instance")
	ci := newAuditLoggingCR(config.AuditLoggingCrName, namespace)
	err = f.Client.Create(goctx.TODO(), ci, &framework.CleanupOptions{TestContext: ctx, Timeout: config.CleanupTimeout, RetryInterval: config.CleanupRetry})
	if err != nil {
		return err
	}

	return nil
}

// RetrieveAuditLogging gets a AuditLogging instance
func RetrieveAuditLogging(f *framework.Framework, ctx *framework.TestCtx) (*operator.AuditLogging, error) {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return nil, err
	}
	ci := &operator.AuditLogging{}
	fmt.Println("--- GET: AuditLogging Instance")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: config.AuditLoggingCrName, Namespace: namespace}, ci)
	if err != nil {
		return nil, err
	}

	return ci, nil
}

// UpdateAuditLogging updates a AuditLogging instance
func UpdateAuditLogging(f *framework.Framework, ctx *framework.TestCtx) error {
	fmt.Println("--- UPDATE: AuditLogging Instance")
	if err := utilwait.PollImmediate(config.WaitForRetry, config.WaitForTimeout, func() (done bool, err error) {
		conCr, err := RetrieveAuditLogging(f, ctx)
		if err != nil {
			return false, err
		}
		conCr.Spec.Fluentd.EnableAuditLoggingForwarding = config.TestForwardingEnvVar
		conCr.Spec.Fluentd.JournalPath = config.TestJournalPath
		if err := f.Client.Update(goctx.TODO(), conCr); err != nil {
			fmt.Println("    --- Waiting for AuditLogging instance stable ...")
			return false, nil
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

// DeleteAuditLogging deletes a AuditLogging instance
func DeleteAuditLogging(reqCr *operator.AuditLogging, f *framework.Framework) error {
	fmt.Println("--- DELETE: AuditLogging Instance")
	if err := f.Client.Delete(goctx.TODO(), reqCr); err != nil {
		return err
	}
	return nil
}

// AssertNoError confirms the error returned is nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

//AuditLogging instance
func newAuditLoggingCR(name, namespace string) *operator.AuditLogging {
	return &operator.AuditLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: operator.AuditLoggingSpec{
			// TODO
		},
	}
}

func GetReconcileRequest(name, namespace string) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
}
