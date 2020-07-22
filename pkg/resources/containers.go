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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const DefaultImageRegistry = "quay.io/opencloudio/"
const DefaultFluentdImageName = "fluentd"
const defaultFluentdImageTag = "v1.6.2-bedrock-0"
const DefaultPCImageName = "audit-policy-controller"
const defaultPCImageTag = "3.5.0"

const FluentdEnvVar = "FLUENTD_TAG_OR_SHA"
const PolicyConrtollerEnvVar = "POLICY_CTRL_TAG_OR_SHA"

var trueVar = true
var falseVar = false
var rootUser = int64(0)
var cpu25 = resource.NewMilliQuantity(25, resource.DecimalSI)          // 25m
var cpu300 = resource.NewMilliQuantity(300, resource.DecimalSI)        // 300m
var memory100 = resource.NewQuantity(100*1024*1024, resource.BinarySI) // 100Mi
var memory400 = resource.NewQuantity(400*1024*1024, resource.BinarySI) // 400Mi

var commonCapabilities = corev1.Capabilities{
	Drop: []corev1.Capability{
		"ALL",
	},
}
var fluentdPrivilegedSecurityContext = corev1.SecurityContext{
	AllowPrivilegeEscalation: &trueVar,
	Privileged:               &trueVar,
	ReadOnlyRootFilesystem:   &trueVar,
	RunAsNonRoot:             &falseVar,
	RunAsUser:                &rootUser,
	Capabilities:             &commonCapabilities,
}

var fluentdRestrictedSecurityContext = corev1.SecurityContext{
	AllowPrivilegeEscalation: &falseVar,
	Privileged:               &falseVar,
	ReadOnlyRootFilesystem:   &trueVar,
	RunAsNonRoot:             &trueVar,
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

var fluentdMainContainer = corev1.Container{
	Image:           DefaultImageRegistry + DefaultFluentdImageName + ":" + defaultFluentdImageTag,
	Name:            FluentdName,
	ImagePullPolicy: corev1.PullIfNotPresent,
	// CommonEnvVars
	Env: []corev1.EnvVar{
		{
			Name: EnableAuditLogForwardKey,
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: FluentdDaemonSetName + "-" + ConfigName,
					},
					Key: EnableAuditLogForwardKey,
				},
			},
		},
	},
	Ports: []corev1.ContainerPort{
		{
			ContainerPort: defaultHTTPPort,
			Protocol:      "TCP",
		},
	},
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
}

// EqualContainers returns a Boolean
func EqualContainers(expected corev1.Container, found corev1.Container, allowModify bool) bool {
	logger := log.WithValues("func", "EqualContainers")
	if !reflect.DeepEqual(found.Name, expected.Name) {
		logger.Info("Container name not equal", "Found", found.Name, "Expected", expected.Name)
		return false
	}
	if !reflect.DeepEqual(found.Image, expected.Image) {
		logger.Info("Image not equal", "Found", found.Image, "Expected", expected.Image)
		return false
	}
	if !reflect.DeepEqual(found.ImagePullPolicy, expected.ImagePullPolicy) {
		logger.Info("ImagePullPolicy not equal", "Found", found.ImagePullPolicy, "Expected", expected.ImagePullPolicy)
		return false
	}
	if !reflect.DeepEqual(found.VolumeMounts, expected.VolumeMounts) {
		logger.Info("VolumeMounts not equal", "Found", found.VolumeMounts, "Expected", expected.VolumeMounts)
		return false
	}
	if !reflect.DeepEqual(found.SecurityContext, expected.SecurityContext) {
		logger.Info("SecurityContext not equal", "Found", found.SecurityContext, "Expected", expected.SecurityContext)
		return false
	}
	if !reflect.DeepEqual(found.Ports, expected.Ports) {
		logger.Info("Ports not equal", "Found", found.Ports, "Expected", expected.Ports)
		return false
	}
	if !reflect.DeepEqual(found.Args, expected.Args) {
		logger.Info("Args not equal", "Found", found.Args, "Expected", expected.Args)
		return false
	}
	if !reflect.DeepEqual(found.Env, expected.Env) {
		logger.Info("Env not equal", "Found", found.Env, "Expected", expected.Env)
		return false
	}
	if !allowModify {
		if !reflect.DeepEqual(found.Resources, expected.Resources) {
			logger.Info("Resources not equal", "Found", found.Resources, "Expected", expected.Resources)
			return false
		}
	}
	return true
}
