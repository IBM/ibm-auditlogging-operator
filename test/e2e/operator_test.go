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
package e2e

import (
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	apis "github.com/ibm/ibm-auditlogging-operator/pkg/apis"
	operatorv1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1"
	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	"github.com/ibm/ibm-auditlogging-operator/test/config"
	"github.com/ibm/ibm-auditlogging-operator/test/helpers"
	testgroups "github.com/ibm/ibm-auditlogging-operator/test/testgroups"
)

func TestAuditLoggingOperator(t *testing.T) {
	auditLoggingList := &operatorv1alpha1.AuditLoggingList{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, auditLoggingList); err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	commonAuditList := &operatorv1.CommonAuditList{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, commonAuditList); err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	if err := deployOperator(t, ctx); err != nil {
		helpers.AssertNoError(t, err)
	}

	// Run group test
	t.Run("TestAuditLogging", testgroups.TestAuditLogging)
	t.Run("TestCommonAudit", testgroups.TestCommonAudit)
}

func deployOperator(t *testing.T, ctx *framework.TestCtx) error {
	err := ctx.InitializeClusterResources(
		&framework.CleanupOptions{
			TestContext:   ctx,
			Timeout:       config.CleanupTimeout,
			RetryInterval: config.CleanupRetry,
		},
	)
	if err != nil {
		t.Fatalf("failed to initialize cluster resources")
		return err
	}

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("failed to get the namespace")
		return err
	}

	return e2eutil.WaitForOperatorDeployment(
		t,
		framework.Global.KubeClient,
		namespace,
		config.TestOperatorName,
		1,
		config.APIRetry,
		config.APITimeout,
	)
}
