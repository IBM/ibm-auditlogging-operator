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

	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var architectureList = []string{"amd64", "ppc64le", "s390x"}
var seconds30 int64 = 30

// BuildDeploymentForPolicyController returns a Deployment object
func BuildDeploymentForPolicyController(instance *operatorv1alpha1.AuditLogging) *appsv1.Deployment {
	reqLogger := log.WithValues("deploymentForPolicyController", "Entry", "instance.Name", instance.Name)
	metaLabels := LabelsForMetadata(AuditPolicyControllerDeploy)
	selectorLabels := LabelsForSelector(AuditPolicyControllerDeploy, instance.Name)
	podLabels := LabelsForPodMetadata(AuditPolicyControllerDeploy, instance.Name)
	annotations := annotationsForMetering(AuditPolicyControllerDeploy)
	policyControllerMainContainer.Image = getImageID(instance.Spec.PolicyController.ImageRegistry, DefaultPCImageName, PolicyConrtollerEnvVar)

	if instance.Spec.PolicyController.PullPolicy != "" {
		switch instance.Spec.PolicyController.PullPolicy {
		case "Always":
			policyControllerMainContainer.ImagePullPolicy = corev1.PullAlways
		case "PullNever":
			policyControllerMainContainer.ImagePullPolicy = corev1.PullNever
		case "IfNotPresent":
			policyControllerMainContainer.ImagePullPolicy = corev1.PullIfNotPresent
		default:
			reqLogger.Info("Trying to update PullPolicy", "NOT SUPPORTED", instance.Spec.PolicyController.PullPolicy)
		}
	}
	var args = make([]string, 0)
	if instance.Spec.PolicyController.Verbosity != "" {
		args = append(args, "--v="+instance.Spec.PolicyController.Verbosity)
	} else {
		args = append(args, "--v=0")
	}
	if instance.Spec.PolicyController.Frequency != "" {
		args = append(args, "--update-frequency="+instance.Spec.PolicyController.Frequency)
	}
	policyControllerMainContainer.Args = args

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuditPolicyControllerDeploy,
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            OperandServiceAccount,
					TerminationGracePeriodSeconds: &seconds30,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "beta.kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values:   architectureList,
											},
										},
									},
								},
							},
						},
					},

					// NodeSelector:                  {},
					Tolerations: commonTolerations,
					Volumes: []corev1.Volume{
						{
							Name: "tmp",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []corev1.Container{
						policyControllerMainContainer,
					},
				},
			},
		},
	}
	return deploy
}

// BuildDaemonForFluentd returns a Daemonset object
func BuildDaemonForFluentd(instance *operatorv1alpha1.AuditLogging) *appsv1.DaemonSet {
	reqLogger := log.WithValues("dameonForFluentd", "Entry", "instance.Name", instance.Name)
	metaLabels := LabelsForMetadata(FluentdName)
	selectorLabels := LabelsForSelector(FluentdName, instance.Name)
	podLabels := LabelsForPodMetadata(FluentdName, instance.Name)
	annotations := annotationsForMetering(FluentdName)
	commonVolumes = BuildCommonVolumes(instance)
	fluentdMainContainer.VolumeMounts = BuildCommonVolumeMounts(instance)
	fluentdMainContainer.Image = getImageID(instance.Spec.Fluentd.ImageRegistry, DefaultFluentdImageName, FluentdEnvVar)

	if instance.Spec.Fluentd.PullPolicy != "" {
		switch instance.Spec.Fluentd.PullPolicy {
		case "Always":
			fluentdMainContainer.ImagePullPolicy = corev1.PullAlways
		case "PullNever":
			fluentdMainContainer.ImagePullPolicy = corev1.PullNever
		case "IfNotPresent":
			fluentdMainContainer.ImagePullPolicy = corev1.PullIfNotPresent
		default:
			reqLogger.Info("Trying to update PullPolicy", "NOT SUPPORTED", instance.Spec.Fluentd.PullPolicy)
		}
	}

	daemon := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDaemonSetName,
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			MinReadySeconds: 5,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            OperandServiceAccount,
					TerminationGracePeriodSeconds: &seconds30,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "beta.kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values:   architectureList,
											},
										},
									},
								},
							},
						},
					},
					// NodeSelector:                  {},
					Tolerations: commonTolerations,
					Volumes:     commonVolumes,
					Containers: []corev1.Container{
						fluentdMainContainer,
					},
				},
			},
		},
	}
	return daemon
}

// EqualDeployments returns a Boolean
func EqualDeployments(expected *appsv1.Deployment, found *appsv1.Deployment) bool {
	if !EqualLabels(found.ObjectMeta.Labels, expected.ObjectMeta.Labels) {
		return false
	}
	if !EqualPods(expected.Spec.Template, found.Spec.Template) {
		return false
	}
	return true
}

// EqualDaemonSets returns a Boolean
func EqualDaemonSets(expected *appsv1.DaemonSet, found *appsv1.DaemonSet) bool {
	if !EqualLabels(found.ObjectMeta.Labels, expected.ObjectMeta.Labels) {
		return false
	}
	if !EqualPods(expected.Spec.Template, found.Spec.Template) {
		return false
	}
	return true
}

// EqualPods returns a Boolean
func EqualPods(expected corev1.PodTemplateSpec, found corev1.PodTemplateSpec) bool {
	logger := log.WithValues("func", "EqualPods")
	if !EqualLabels(found.ObjectMeta.Labels, expected.ObjectMeta.Labels) {
		return false
	}
	if !EqualAnnotations(found.ObjectMeta.Annotations, expected.ObjectMeta.Annotations) {
		return false
	}
	if !reflect.DeepEqual(found.Spec.ServiceAccountName, expected.Spec.ServiceAccountName) {
		logger.Info("ServiceAccount not equal", "Found", found.Spec.ServiceAccountName, "Expected", expected.Spec.ServiceAccountName)
		return false
	}
	if len(found.Spec.Containers) != len(expected.Spec.Containers) {
		logger.Info("Number of containers not equal", "Found", len(found.Spec.Containers), "Expected", len(expected.Spec.Containers))
		return false
	}
	if !EqualContainers(expected.Spec.Containers[0], found.Spec.Containers[0]) {
		return false
	}
	return true
}
