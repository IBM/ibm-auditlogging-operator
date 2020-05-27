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
	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var commonVolumes = []corev1.Volume{}

// BuildCommonVolumeMounts returns an array of VolumeMount objects
func BuildCommonVolumeMounts(instance *operatorv1alpha1.AuditLogging) []corev1.VolumeMount {
	var journal = defaultJournalPath
	if instance.Spec.Fluentd.JournalPath != "" {
		journal = instance.Spec.Fluentd.JournalPath
	}
	commonVolumeMounts := []corev1.VolumeMount{
		{
			Name:      FluentdConfigName,
			MountPath: "/fluentd/etc/" + fluentdConfigKey,
			SubPath:   fluentdConfigKey,
		},
		{
			Name:      SourceConfigName,
			MountPath: fluentdInput,
			SubPath:   SourceConfigKey,
		},
		{
			Name:      QRadarConfigName,
			MountPath: qRadarOutput,
			SubPath:   QRadarConfigKey,
		},
		{
			Name:      SplunkConfigName,
			MountPath: splunkOutput,
			SubPath:   SplunkConfigKey,
		},
		{
			Name:      "journal",
			MountPath: journal,
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
			Name:      AuditLoggingClientCertSecName,
			MountPath: "/fluentd/etc/tls",
			ReadOnly:  true,
		},
		{
			Name:      AuditLoggingServerCertSecName,
			MountPath: "/fluentd/etc/https",
			ReadOnly:  true,
		},
	}
	return commonVolumeMounts
}

// BuildCommonVolumes returns an array of Volume objects
func BuildCommonVolumes(instance *operatorv1alpha1.AuditLogging) []corev1.Volume {
	var journal = defaultJournalPath
	if instance.Spec.Fluentd.JournalPath != "" {
		journal = instance.Spec.Fluentd.JournalPath
	}
	commonVolumes := []corev1.Volume{
		{
			Name: "journal",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: journal,
					Type: nil,
				},
			},
		},
		{
			Name: FluentdConfigName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: FluentdDaemonSetName + "-" + ConfigName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  fluentdConfigKey,
							Path: fluentdConfigKey,
						},
					},
				},
			},
		},
		{
			Name: SourceConfigName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: FluentdDaemonSetName + "-" + SourceConfigName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  SourceConfigKey,
							Path: SourceConfigKey,
						},
					},
				},
			},
		},
		{
			Name: QRadarConfigName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: FluentdDaemonSetName + "-" + QRadarConfigName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  QRadarConfigKey,
							Path: QRadarConfigKey,
						},
					},
				},
			},
		},
		{
			Name: SplunkConfigName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: FluentdDaemonSetName + "-" + SplunkConfigName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  SplunkConfigKey,
							Path: SplunkConfigKey,
						},
					},
				},
			},
		},
		{
			Name: "shared",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: AuditLoggingClientCertSecName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: AuditLoggingClientCertSecName,
				},
			},
		},
		{
			Name: AuditLoggingServerCertSecName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: AuditLoggingServerCertSecName,
				},
			},
		},
	}
	return commonVolumes
}
