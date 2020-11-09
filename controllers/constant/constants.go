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

package constant

const (
	AuditLoggingComponentName     = "common-audit-logging"
	AuditLoggingReleaseName       = "common-audit-logging"
	AuditLoggingCrType            = "auditlogging_cr"
	DefaultPCImageName            = "audit-policy-controller"
	ProductName                   = "IBM Cloud Platform Common Services"
	ProductID                     = "068a62892a1e4db39641342e592daa25"
	ProductMetric                 = "FREE"
	InstanceNamespace             = "ibm-common-services"
	FluentdName                   = "fluentd"
	DefaultImageRegistry          = "quay.io/opencloudio/"
	TestImageRegistry             = "quay.io/hbradfield/"
	DefaultFluentdImageName       = "fluentd"
	DefaultFluentdImageTag        = "v1.6.2-bedrock-1"
	FluentdEnvVar                 = "FLUENTD_TAG_OR_SHA"
	OperatorNamespaceKey          = "POD_NAMESPACE"
	DefaultJobImageName           = "audit-garbage-collector"
	DefaultJobImageTag            = "1.0.0"
	JobEnvVar                     = "JOB_TAG_OR_SHA"
	PolicyControllerEnvVar        = "POLICY_CTRL_TAG_OR_SHA"
	DefaultPCImageTag             = "3.5.3"
	AuditTypeLabel                = "operator.ibm.com/managedBy-audit"
	DefaultEnablePolicyController = "true"
)
