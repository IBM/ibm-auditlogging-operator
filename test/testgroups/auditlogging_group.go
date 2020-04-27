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

package testgroups

import (
	"testing"

	"github.com/operator-framework/operator-sdk/pkg/test"

	"github.com/ibm/ibm-auditlogging-operator/test/helpers"
)

// TestAuditLogging is the test group for testing Audit Logging
func TestAuditLogging(t *testing.T) {
	t.Run("TestAuditLoggingCRUD", TestAuditLoggingCRUD)
}

// TestAuditLoggingCRUD is for testing AuditLogging
func TestAuditLoggingCRUD(t *testing.T) {

	ctx := test.NewTestCtx(t)
	defer ctx.Cleanup()

	// get global framework variables
	f := test.Global

	if err := helpers.CreateAuditLogging(f, ctx); err != nil {
		t.Fatal(err)
	}

	reqCr, err := helpers.RetrieveAuditLogging(f, ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := helpers.UpdateAuditLogging(f, ctx); err != nil {
		t.Fatal(err)
	}

	if err = helpers.DeleteAuditLogging(reqCr, f); err != nil {
		t.Fatal(err)
	}
}
