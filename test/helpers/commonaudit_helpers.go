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

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilwait "k8s.io/apimachinery/pkg/util/wait"

	operator "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1"
	"github.com/ibm/ibm-auditlogging-operator/test/config"
)

// CreateCommonAudit creates a CommonAudit instance
func CreateCommonAudit(f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	// Create CommonAudit instance
	fmt.Println("--- CREATE: CommonAudit Instance")
	ci := newCommonAuditCR(config.CommonAuditCrName, namespace)
	err = f.Client.Create(goctx.TODO(), ci, &framework.CleanupOptions{TestContext: ctx, Timeout: config.CleanupTimeout, RetryInterval: config.CleanupRetry})
	if err != nil {
		return err
	}

	return nil
}

// RetrieveCommonAudit gets a CommonAudit instance
func RetrieveCommonAudit(f *framework.Framework, ctx *framework.TestCtx) (*operator.CommonAudit, error) {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return nil, err
	}
	ci := &operator.CommonAudit{}
	fmt.Println("--- GET: CommonAudit Instance")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: config.CommonAuditCrName, Namespace: namespace}, ci)
	if err != nil {
		return nil, err
	}

	return ci, nil
}

// UpdateCommonAudit updates a CommonAudit instance
func UpdateCommonAudit(f *framework.Framework, ctx *framework.TestCtx) error {
	fmt.Println("--- UPDATE: CommonAudit Instance")
	if err := utilwait.PollImmediate(config.WaitForRetry, config.WaitForTimeout, func() (done bool, err error) {
		conCr, err := RetrieveCommonAudit(f, ctx)
		if err != nil {
			return false, err
		}
		conCr.Spec.Outputs.Splunk.Host = config.TestSplunkHost
		conCr.Spec.Outputs.Splunk.Port = config.TestSplunkPort
		conCr.Spec.Outputs.Splunk.Token = config.TestSplunkToken
		conCr.Spec.Outputs.Splunk.Protocol = config.TestSplunkProtocol
		if err := f.Client.Update(goctx.TODO(), conCr); err != nil {
			fmt.Println("    --- Waiting for CommonAudit instance stable ...")
			return false, nil
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

// DeleteCommonAudit deletes a CommonAudit instance
func DeleteCommonAudit(reqCr *operator.CommonAudit, f *framework.Framework) error {
	fmt.Println("--- DELETE: CommonAudit Instance")
	if err := f.Client.Delete(goctx.TODO(), reqCr); err != nil {
		return err
	}
	return nil
}

//CommonAudit instance
func newCommonAuditCR(name, namespace string) *operator.CommonAudit {
	return &operator.CommonAudit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: operator.CommonAuditSpec{
			Outputs: operator.CommonAuditSpecOutputs{
				Splunk: operator.CommonAuditSpecSplunk{
					Host:     "worker",
					Port:     0,
					Token:    "abc-123",
					Protocol: "https",
				},
			},
		},
	}
}
