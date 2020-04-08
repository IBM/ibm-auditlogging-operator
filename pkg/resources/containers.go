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

package resources

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const ConfigName = "config"
const FluentdConfigName = "main-config"
const SourceConfigName = "source-config"
const QRadarConfigName = "remote-syslog-config"
const SplunkConfigName = "splunk-hec-config"
const ELKConfigName = "elk-config"

const fluentdInput = "/fluentd/etc/source.conf"
const qRadarOutput = "/fluentd/etc/remoteSyslog.conf"
const splunkOutput = "/fluentd/etc/splunkHEC.conf"
const elkOutput = "/fluentd/etc/elk.conf"

const enableAuditLogForwardKey = "ENABLE_AUDIT_LOGGING_FORWARDING"

const fluentdConfigKey = "fluent.conf"
const sourceConfigKey = "source.conf"
const splunkConfigKey = "splunkHEC.conf"
const qRadarConfigKey = "remoteSyslog.conf"
const elkConfigKey = "elk.conf"

const FluentdDaemonSetName = "audit-logging-fluentd-ds"
const auditLoggingCertSecretName = "audit-certs"
const AuditLoggingCertName = "fluentd"
const AuditPolicyControllerDeploy = "audit-policy-controller"
const FluentdName = "fluentd"
const AuditPolicyCRDName = "auditpolicies.audit.policies.ibm.com"

const defaultImageRegistry = "quay.io/opencloudio/"
const defaultFluentdImageName = "fluentd"
const defaultFluentdImageTag = "v1.6.2-ubi7"
const defaultPCImageName = "audit-policy-controller"
const defaultPCImageTag = "3.4.0"
const defaultJournalPath = "/run/log/journal"
const defaultSourceTag = "icp-audit"
const defaultSourcePath = "/icp-audit"
const defaultSourceID = "input_systemd_icp"

var trueVar = true
var falseVar = false
var rootUser = int64(0)
var user100 = int64(100)
var replicas = int32(1)
var cpu25 = resource.NewMilliQuantity(25, resource.DecimalSI)          // 25m
var cpu100 = resource.NewMilliQuantity(100, resource.DecimalSI)        // 100m
var cpu200 = resource.NewMilliQuantity(200, resource.DecimalSI)        // 200m
var cpu300 = resource.NewMilliQuantity(300, resource.DecimalSI)        // 300m
var memory100 = resource.NewQuantity(100*1024*1024, resource.BinarySI) // 100Mi
var memory150 = resource.NewQuantity(150*1024*1024, resource.BinarySI) // 150Mi
var memory300 = resource.NewQuantity(300*1024*1024, resource.BinarySI) // 300Mi
var memory400 = resource.NewQuantity(400*1024*1024, resource.BinarySI) // 400Mi

var commonCapabilities = corev1.Capabilities{
	Drop: []corev1.Capability{
		"ALL",
	},
}
var fluentdSecurityContext = corev1.SecurityContext{
	AllowPrivilegeEscalation: &trueVar,
	Privileged:               &trueVar,
	ReadOnlyRootFilesystem:   &trueVar,
	RunAsNonRoot:             &falseVar,
	RunAsUser:                &rootUser,
	Capabilities:             &commonCapabilities,
}

var policyControllerSecurityContext = corev1.SecurityContext{
	AllowPrivilegeEscalation: &falseVar,
	Privileged:               &falseVar,
	ReadOnlyRootFilesystem:   &trueVar,
	RunAsNonRoot:             &trueVar,
	RunAsUser:                &user100,
	Capabilities:             &commonCapabilities,
}

var commonTolerations = []corev1.Toleration{
	{
		Key:      "dedicated",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	},
	{
		Key:      "CriticalAddonsOnly",
		Operator: corev1.TolerationOpExists,
	},
}

var fluentdMainConfigData = `
fluent.conf: |-
  # Input plugins
  @include /fluentd/etc/source.conf

  # Output plugins (Only use one output plugin conf file at a time. Comment or remove other files)
  #@include /fluentd/etc/remoteSyslog.conf
  #@include /fluentd/etc/splunkHEC.conf
  #@include /fluentd/etc/elk.conf
`

var sourceConfigData1 = `
    <source>
        @type systemd
        @log_level info
`
var sourceConfigData2 = `
        read_from_head true
        <storage>
            @type local
            persistent true
`
var sourceConfigData3 = `
        </storage>
        <entry>
            fields_strip_underscores true
            fields_lowercase true
        </entry>
    </source>`

var icpAuditSourceFilter = `
        @type parser
        format json
        key_name message
        reserve_data true
    </filter>`

// IBMDEV systemd debug
/*
   <match mcm-audit>
       @type copy
       <store>
         @type "stdout"
       </store>
   </match>`
*/
var splunkConfigData = `
        @type splunk_hec
        hec_host SPLUNK_SERVER_HOSTNAME
        hec_port SPLUNK_PORT
        hec_token SPLUNK_HEC_TOKEN
        ca_file /fluentd/etc/tls/splunkCA.pem
        source ${tag}
    </match>`

var qRadarConfigData = `
        @type copy
        <store>
            @type remote_syslog
            host QRADAR_SERVER_HOSTNAME
            port QRADAR_PORT_FOR_icp-audit
            hostname QRADAR_LOG_SOURCE_IDENTIFIER_FOR_icp-audit
            protocol tcp
            tls true
            ca_file /fluentd/etc/tls/qradar.crt
            packet_size 4096
            program fluentd
            <format>
                @type single_value
                message_key message
            </format>
        </store>
    </match>`

var elkFilter = `
    <filter icp-audit>
        @type elasticsearch_genid
        hash_id_key _hash
    </filter>`

var elkConfigData = `
        @type elasticsearch
        @log_level info
        type_name fluentd
        id_key _hash
        remove_keys _hash
        logstash_prefix audit
        scheme http
        host ELASTICSEARCH_SERVER_HOSTNAME
        port ELASTICSEARCH_PORT
        user ELASTICSEARCH_USERNAME
        password ELASTICSEARCH_PASSWORD
        <buffer>
            flush_thread_count 2
            flush_interval 15s
            chunk_limit_size 2M
            queue_limit_length 32
            retry_max_interval 30
            retry_forever true
        </buffer>
    </match>`

var policyControllerMainContainer = corev1.Container{
	Image:           defaultImageRegistry + defaultPCImageName + ":" + defaultPCImageTag,
	Name:            AuditPolicyControllerDeploy,
	ImagePullPolicy: corev1.PullIfNotPresent,
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "tmp",
			MountPath: "/tmp",
		},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"sh",
					"-c",
					"pgrep audit-policy -l",
				},
			},
		},
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"sh",
					"-c",
					"exec echo start audit-policy-controller",
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
	},
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu200,
			corev1.ResourceMemory: *memory300},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory150},
	},
	SecurityContext: &policyControllerSecurityContext,
}

var fluentdMainContainer = corev1.Container{
	Image:           defaultImageRegistry + defaultFluentdImageName + ":" + defaultFluentdImageTag,
	Name:            FluentdName,
	ImagePullPolicy: corev1.PullIfNotPresent,
	// CommonEnvVars
	Env: []corev1.EnvVar{
		{
			Name: enableAuditLogForwardKey,
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: FluentdDaemonSetName + "-" + ConfigName,
					},
					Key: enableAuditLogForwardKey,
				},
			},
		},
	}, //CS??? TODO
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"ls",
					"/tmp",
				},
			},
		},
		InitialDelaySeconds: 2,
		TimeoutSeconds:      1,
		PeriodSeconds:       2,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"bash",
					"-c",
					"exec echo start",
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu300,
			corev1.ResourceMemory: *memory400},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu25,
			corev1.ResourceMemory: *memory100},
	},
	SecurityContext: &fluentdSecurityContext,
}
