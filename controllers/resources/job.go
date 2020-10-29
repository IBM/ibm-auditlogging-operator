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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-auditlogging-operator/api/v1alpha1"
	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	utils "github.com/IBM/ibm-auditlogging-operator/controllers/util"
)

const jobServiceAccountName = "ibm-auditlogging-cleanup"
const JobName = "audit-logging-cleanup"
const jobNamespaceEnvVar = "CR_NAMESPACE"

func BuildJobForAuditLogging(instance *operatorv1alpha1.AuditLogging, namespace string) *batchv1.Job {
	metaLabels := utils.LabelsForMetadata(JobName)
	podLabels := utils.LabelsForPodMetadata(JobName, instance.Name)
	annotations := utils.AnnotationsForMetering(false)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      JobName,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: jobServiceAccountName,
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					Tolerations:        commonTolerations,
					Affinity: &corev1.Affinity{
						NodeAffinity: commonNodeAffinity,
					},
					Containers: buildJobContainer(namespace, instance.Spec.Fluentd.ImageRegistry),
				},
			},
		},
	}
	return job
}

func buildJobContainer(namespace string, imageRegistry string) []corev1.Container {
	return []corev1.Container{
		{
			Name:            JobName,
			Image:           utils.GetImageID(imageRegistry, constant.DefaultJobImageName, constant.JobEnvVar),
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &restrictedSecurityContext,
			Resources: corev1.ResourceRequirements{
				Limits: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceCPU:    *cpu25,
					corev1.ResourceMemory: *memory100},
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceCPU:    *cpu25,
					corev1.ResourceMemory: *memory100},
			},
			Command: []string{"/manager"},
			Env: []corev1.EnvVar{
				{
					Name:  jobNamespaceEnvVar,
					Value: namespace,
				},
			},
		},
	}
}
