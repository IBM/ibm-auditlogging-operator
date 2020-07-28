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

package config

import (
	"time"
)

const (
	// APIRetry defines the frequency at which we check for updates against the
	// k8s api when waiting for a specific condition to be true.
	APIRetry = time.Second * 5

	// APITimeout defines the amount of time we should spend querying the k8s api
	// when waiting for a specific condition to be true.
	APITimeout = time.Minute * 60

	// CleanupRetry is the interval at which test framework attempts cleanup
	CleanupRetry = time.Second * 10

	// CleanupTimeout is the wait time for test framework cleanup
	CleanupTimeout = time.Second * 180

	// WaitForTimeout is the wait time for cluster resource
	WaitForTimeout = time.Second * 200

	// WaitForRetry is the the interval at checking cluster resource
	WaitForRetry = time.Second * 10

	// TestOperatorName specifies the name of the operator being tested
	TestOperatorName = "ibm-auditlogging-operator"

	// AuditLoggingCrName specifies the name of the custom resource of the AuditLogging
	AuditLoggingCrName = "test-auditlogging"

	// CommonAuditCrName specifies the name of the custom resource of the AuditLogging
	CommonAuditCrName = "test-commonaudit"

	// TestJournalPath specifies the journal path of fluentd being tested
	TestJournalPath = "/var/log/audit"

	// TestForwardingEnvVar specifies the value of fluentd env var ENABLE_AUDIT_LOGGING_FORWARDING being tested
	TestForwardingEnvVar = true

	// TestSplunkHost specifies the splunk server host name
	TestSplunkHost = "master"

	// TestSplunkPort specifies the splunk server port
	TestSplunkPort = 8088

	// TestSplunkToken specifies the splunk hec token
	TestSplunkToken = "123-abc"

	// TestSplunkTLS specifies if tls for splunk is enabled
	TestSplunkTLS = false
)
