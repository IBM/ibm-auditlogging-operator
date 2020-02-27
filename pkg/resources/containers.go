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

const journalPath = "/run/log/journal"
const FluentdDaemonSetName = "audit-logging-fluentd-ds"
const auditLoggingCertSecretName = "audit-certs"
const ConfigName = "config"
const FluentdConfigName = "main-config"
const SourceConfigName = "source-config"
const QRadarConfigName = "remote-syslog-config"
const SplunkConfigName = "splunk-hec-config"
const fluentdInput = "/fluentd/etc/source.conf"
const qRadarOutput = "/fluentd/etc/remoteSyslog.conf"
const splunkOutput = "/fluentd/etc/splunkHEC.conf"
const enableAuditLogForwardKey = "ENABLE_AUDIT_LOGGING_FORWARDING"
const fluentdConfigKey = "fluent.conf"
const sourceConfigKey = "source.conf"
const splunkConfigKey = "splunkHEC.conf"
const qRadarConfigKey = "remoteSyslog.conf"
const AuditLoggingCertName = "fluentd"
const AuditPolicyControllerDeploy = "audit-policy-controller"

var trueVar = true
var falseVar = false
var rootUser = int64(0)
var user100 = int64(100)
var replicas = int32(1)
var cpu25 = resource.NewMilliQuantity(25, resource.DecimalSI)          // 25m
var cpu300 = resource.NewMilliQuantity(300, resource.DecimalSI)        // 300m
var memory100 = resource.NewQuantity(100*1024*1024, resource.BinarySI) // 100Mi
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
  # Output plugins
  # Only use one output plugin conf file at a time. Comment or remove other files 
  # To forward audit logs to QRadar, uncommnet following line, add QRadar server information in the 'audit-logging-fluentd-ds-remote-syslog-config' ConfigMap and restart the 'audit-logging-fluentd-ds-*' pods
  #@include /fluentd/etc/remoteSyslog.conf
  #To forward audit logs to Splunk over HTTPS, uncomment following line, add Splunk server information in the 'audit-logging-fluentd-ds-splunk-hec-config' ConfigMap and restart the 'audit-logging-fluentd-ds-*' pods
  #@include /fluentd/etc/splunkHEC.conf
`

var sourceConfigData1 = `
source.conf: |-
    <source>
        @type systemd
        @id input_systemd_icp
        @log_level info
        tag icp-audit
        path `
var sourceConfigData2 = `
        matches '[{ "SYSLOG_IDENTIFIER": "icp-audit" }]'
        read_from_head true
        <storage>
          @type local
          persistent true
          path /icp-audit
        </storage>
        <entry>
          fields_strip_underscores true
          fields_lowercase true
        </entry>
    </source>
    <filter icp-audit>
        @type parser
        format json
        key_name message
        reserve_data true
    </filter>`

var splunkConfigData = `
splunkHEC.conf: |-
     <match icp-audit kube-audit>
        @type splunk_hec
        hec_host SPLUNK_SERVER_HOSTNAME
        hec_port SPLUNK_PORT
        hec_token SPLUNK_HEC_TOKEN
        ca_file /fluentd/etc/tls/splunkCA.pem
        source ${tag}
     </match>`

var qRadarConfigData = `
remoteSyslog.conf: |-
    <match icp-audit>
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

var policyControllerMainContainer = corev1.Container{
	Image:           "hyc-cloud-private-edge-docker-local.artifactory.swg-devops.com/ibmcom-amd64/audit-policy-controller:3.3.1",
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
	SecurityContext: &policyControllerSecurityContext,
}

var fluentdMainContainer = corev1.Container{
	Image:           "hyc-cloud-private-edge-docker-local.artifactory.swg-devops.com/ibmcom-amd64/fluentd:v1.6.2-rhc",
	Name:            "fluentd",
	ImagePullPolicy: corev1.PullIfNotPresent,
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      FluentdConfigName,
			MountPath: "/fluentd/etc/" + fluentdConfigKey,
			SubPath:   fluentdConfigKey,
		},
		{
			Name:      SourceConfigName,
			MountPath: fluentdInput,
			SubPath:   sourceConfigKey,
		},
		{
			Name:      QRadarConfigName,
			MountPath: qRadarOutput,
			SubPath:   qRadarConfigKey,
		},
		{
			Name:      SplunkConfigName,
			MountPath: splunkOutput,
			SubPath:   splunkConfigKey,
		},
		{
			Name:      "journal",
			MountPath: journalPath,
			ReadOnly:  true,
		},
		{
			Name:      "shared",
			MountPath: "/icp-audit",
		},
		{
			Name:      "shared",
			MountPath: "/tmp",
		},
		{
			Name:      "certs",
			MountPath: "/fluentd/etc/tls",
			ReadOnly:  true,
		},
	},
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
