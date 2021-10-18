//
// Copyright 2021 IBM Corporation
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
	"net"
	"os"
	"reflect"

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	"github.com/IBM/ibm-auditlogging-operator/controllers/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"
	operatorv1alpha1 "github.com/IBM/ibm-auditlogging-operator/api/v1alpha1"
)

var commonVolumes = []corev1.Volume{}
var architectureList = []string{"amd64", "ppc64le", "s390x"}
var seconds30 int64 = 30
var defaultReplicas = int32(1)

// FluentdDaemonSetName is the name of the fluentd daemonset name
const FluentdDaemonSetName = "audit-logging-fluentd-ds"

// FluentdDeploymentName is the name of the fluentd deployment
const FluentdDeploymentName = "audit-logging-fluentd"

// AuditPolicyControllerDeploy is the name of the audit-policy-controller deployment
const AuditPolicyControllerDeploy = "audit-policy-controller"

const fluentdInput = "/fluentd/etc/source.conf"
const qRadarOutput = "/fluentd/etc/remoteSyslog.conf"
const splunkOutput = "/fluentd/etc/splunkHEC.conf"

const defaultJournalPath = "/run/log/journal"

var commonNodeAffinity = &corev1.NodeAffinity{
	RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{
			{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "kubernetes.io/arch",
						Operator: corev1.NodeSelectorOpIn,
						Values:   architectureList,
					},
				},
			},
		},
	},
}

// BuildDeploymentForPolicyController returns a Deployment object
func BuildDeploymentForPolicyController(instance *operatorv1alpha1.AuditLogging, namespace string) *appsv1.Deployment {
	log.WithValues("func", "BuildDeploymentForPolicyController")
	metaLabels := util.LabelsForMetadata(AuditPolicyControllerDeploy)
	selectorLabels := util.LabelsForSelector(AuditPolicyControllerDeploy, instance.Name)
	podLabels := util.LabelsForPodMetadata(AuditPolicyControllerDeploy, instance.Name)
	annotations := util.AnnotationsForMetering(false)
	policyControllerMainContainer.Image = os.Getenv("AUDIT_POLICY_CONTROLLER_IMAGE")
	policyControllerMainContainer.ImagePullPolicy = getPullPolicy(instance.Spec.PolicyController.PullPolicy)

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
			Namespace: namespace,
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
					ServiceAccountName:            AuditPolicyServiceAccount,
					TerminationGracePeriodSeconds: &seconds30,
					Affinity: &corev1.Affinity{
						NodeAffinity: commonNodeAffinity,
					},
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

// BuildDeploymentForFluentd returns a Deployment object
func BuildDeploymentForFluentd(instance *operatorv1.CommonAudit) *appsv1.Deployment {
	log.WithValues("func", "BuildDeploymentForFluentd")
	metaLabels := util.LabelsForMetadata(constant.FluentdName)
	selectorLabels := util.LabelsForSelector(constant.FluentdName, instance.Name)
	podLabels := util.LabelsForPodMetadata(constant.FluentdName, instance.Name)
	annotations := util.AnnotationsForMetering(false)

	volumes := buildFluentdDeploymentVolumes()
	fluentdMainContainer.VolumeMounts = buildFluentdDeploymentVolumeMounts()
	fluentdMainContainer.Image = os.Getenv("FLUENTD_IMAGE")
	fluentdMainContainer.ImagePullPolicy = getPullPolicy(instance.Spec.Fluentd.PullPolicy)
	// Run fluentd as restricted
	fluentdMainContainer.SecurityContext = &restrictedSecurityContext
	var replicas = defaultReplicas
	if instance.Spec.Replicas > 0 {
		replicas = instance.Spec.Replicas
	}
	fluentdMainContainer.Resources = buildResources(instance.Spec.Fluentd.Resources, defaultFluentdResources)

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDeploymentName,
			Namespace: instance.Namespace,
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
						NodeAffinity: commonNodeAffinity,
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app.kubernetes.io/name",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													constant.FluentdName,
												},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					Tolerations: commonTolerations,
					Volumes:     volumes,
					Containers: []corev1.Container{
						fluentdMainContainer,
					},
				},
			},
		},
	}

	if len(instance.Spec.Outputs.HostAliases) > 0 {
		var hostAliases = []corev1.HostAlias{}
		for _, hostAlias := range instance.Spec.Outputs.HostAliases {
			if ip := net.ParseIP(hostAlias.HostIP); ip != nil {
				hostAliases = append(hostAliases, corev1.HostAlias{IP: hostAlias.HostIP, Hostnames: hostAlias.Hostnames})
			} else {
				log.Info("[WARNING] Invalid HostAliases IP. Update CommonAudit CR with a valid IP.", "Found", hostAlias.HostIP, "Instance", instance.Name)
			}
		}
		deploy.Spec.Template.Spec.HostAliases = hostAliases
	}

	return deploy
}

func buildFluentdDeploymentVolumes() []corev1.Volume {
	commonVolumes := []corev1.Volume{
		{
			Name: FluentdConfigName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: FluentdDaemonSetName + "-" + ConfigName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  FluentdConfigKey,
							Path: FluentdConfigKey,
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

func buildFluentdDeploymentVolumeMounts() []corev1.VolumeMount {
	commonVolumeMounts := []corev1.VolumeMount{
		{
			Name:      FluentdConfigName,
			MountPath: "/fluentd/etc/" + FluentdConfigKey,
			SubPath:   FluentdConfigKey,
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

// BuildDaemonForFluentd returns a Daemonset object
func BuildDaemonForFluentd(instance *operatorv1alpha1.AuditLogging, namespace string) *appsv1.DaemonSet {
	metaLabels := util.LabelsForMetadata(constant.FluentdName)
	selectorLabels := util.LabelsForSelector(constant.FluentdName, instance.Name)
	podLabels := util.LabelsForPodMetadata(constant.FluentdName, instance.Name)
	annotations := util.AnnotationsForMetering(true)
	commonVolumes = buildDaemonsetVolumes(instance)
	fluentdMainContainer.VolumeMounts = buildDaemonsetVolumeMounts(instance)
	fluentdMainContainer.Image = os.Getenv("FLUENTD_IMAGE")
	fluentdMainContainer.ImagePullPolicy = getPullPolicy(instance.Spec.Fluentd.PullPolicy)
	// Run fluentd as privileged
	fluentdMainContainer.SecurityContext = &fluentdPrivilegedSecurityContext
	// setup the resource requirements
	fluentdMainContainer.Resources = buildResources(instance.Spec.Fluentd.Resources, defaultFluentdResources)

	daemon := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDaemonSetName,
			Namespace: namespace,
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
						NodeAffinity: commonNodeAffinity,
					},
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

func buildDaemonsetVolumes(instance *operatorv1alpha1.AuditLogging) []corev1.Volume {
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
							Key:  FluentdConfigKey,
							Path: FluentdConfigKey,
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

func buildDaemonsetVolumeMounts(instance *operatorv1alpha1.AuditLogging) []corev1.VolumeMount {
	var journal = defaultJournalPath
	if instance.Spec.Fluentd.JournalPath != "" {
		journal = instance.Spec.Fluentd.JournalPath
	}
	commonVolumeMounts := []corev1.VolumeMount{
		{
			Name:      FluentdConfigName,
			MountPath: "/fluentd/etc/" + FluentdConfigKey,
			SubPath:   FluentdConfigKey,
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

func buildResources(requestedResources, defaultResources corev1.ResourceRequirements) corev1.ResourceRequirements {
	var resourceRequirements = corev1.ResourceRequirements{
		Limits:   defaultResources.Limits.DeepCopy(),
		Requests: defaultResources.Requests.DeepCopy(),
	}
	if requestedResources.Limits != nil {
		// check CPU limits
		cpuLimit := requestedResources.Limits.Cpu()
		if !cpuLimit.IsZero() {
			resourceRequirements.Limits[corev1.ResourceCPU] = *cpuLimit
		}
		// check Memory limits
		memoryLimit := requestedResources.Limits.Memory()
		if !memoryLimit.IsZero() {
			resourceRequirements.Limits[corev1.ResourceMemory] = *memoryLimit
		}
	}
	if requestedResources.Requests != nil {
		// check CPU requests
		cpuRequest := requestedResources.Requests.Cpu()
		if !cpuRequest.IsZero() {
			resourceRequirements.Requests[corev1.ResourceCPU] = *cpuRequest
		}
		// check Memory requests
		memoryRequest := requestedResources.Requests.Memory()
		if !memoryRequest.IsZero() {
			resourceRequirements.Requests[corev1.ResourceMemory] = *memoryRequest
		}
	}
	return resourceRequirements
}

// EqualDeployments returns a Boolean
func EqualDeployments(expected *appsv1.Deployment, found *appsv1.Deployment, allowModify bool) bool {
	logger := log.WithValues("func", "EqualDeployments")
	if !util.EqualLabels(found.ObjectMeta.Labels, expected.ObjectMeta.Labels) {
		return false
	}
	if !reflect.DeepEqual(expected.Spec.Replicas, found.Spec.Replicas) {
		logger.Info("Replicas not equal", "Found", found.Spec.Replicas, "Expected", expected.Spec.Replicas)
		return false
	}
	if !EqualPods(expected.Spec.Template, found.Spec.Template, allowModify) {
		return false
	}
	return true
}

// EqualDaemonSets returns a Boolean
func EqualDaemonSets(expected *appsv1.DaemonSet, found *appsv1.DaemonSet) bool {
	if !util.EqualLabels(found.ObjectMeta.Labels, expected.ObjectMeta.Labels) {
		return false
	}
	if !EqualPods(expected.Spec.Template, found.Spec.Template, true) {
		return false
	}
	return true
}

// EqualPods returns a Boolean
func EqualPods(expected corev1.PodTemplateSpec, found corev1.PodTemplateSpec, allowModify bool) bool {
	logger := log.WithValues("func", "EqualPods")
	if !util.EqualLabels(found.ObjectMeta.Labels, expected.ObjectMeta.Labels) {
		return false
	}
	if !util.EqualAnnotations(found.ObjectMeta.Annotations, expected.ObjectMeta.Annotations) {
		return false
	}
	if !reflect.DeepEqual(found.Spec.ServiceAccountName, expected.Spec.ServiceAccountName) {
		logger.Info("ServiceAccount not equal", "Found", found.Spec.ServiceAccountName, "Expected", expected.Spec.ServiceAccountName)
		return false
	}
	if !allowModify {
		if !equalHostAliases(found.Spec.HostAliases, expected.Spec.HostAliases) {
			logger.Info("HostAliases not equal", "Found", found.Spec.HostAliases, "Expected", expected.Spec.HostAliases)
			return false
		}
	}
	if len(found.Spec.Containers) != len(expected.Spec.Containers) {
		logger.Info("Number of containers not equal", "Found", len(found.Spec.Containers), "Expected", len(expected.Spec.Containers))
		return false
	}
	if !EqualContainers(expected.Spec.Containers[0], found.Spec.Containers[0], allowModify) {
		return false
	}
	return true
}

func equalHostAliases(found []corev1.HostAlias, expected []corev1.HostAlias) bool {
	if (found == nil) != (expected == nil) {
		return false
	}
	if len(found) != len(expected) {
		return false
	}
	for i := range found {
		if found[i].IP != expected[i].IP {
			return false
		}
		if !reflect.DeepEqual(found[i].Hostnames, expected[i].Hostnames) {
			return false
		}
	}
	return true
}

func getPullPolicy(pullPolicy string) corev1.PullPolicy {
	switch pullPolicy {
	case "Always":
		return corev1.PullAlways
	case "PullNever":
		return corev1.PullNever
	default:
		return corev1.PullIfNotPresent
	}
}
