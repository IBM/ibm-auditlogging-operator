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

	"github.com/operator-framework/operator-sdk/pkg/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	apis "github.com/ibm/ibm-auditlogging-operator/pkg/apis"
	operator "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1"
	"github.com/ibm/ibm-auditlogging-operator/test/config"
	"github.com/ibm/ibm-auditlogging-operator/test/helpers"
	testgroups "github.com/ibm/ibm-auditlogging-operator/test/testgroups"
)

func TestAuditLoggingOperator(t *testing.T) {
	auditLoggingList := &operator.AuditLoggingList{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, auditLoggingList); err != nil {
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
}

func deployOperator(t *testing.T, ctx *test.TestCtx) error {
	err := ctx.InitializeClusterResources(
		&test.CleanupOptions{
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
		test.Global.KubeClient,
		namespace,
		config.TestOperatorName,
		1,
		config.APIRetry,
		config.APITimeout,
	)
}
