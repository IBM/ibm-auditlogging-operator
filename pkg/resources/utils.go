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
	"strconv"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	yaml "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	ver "github.com/ibm/ibm-auditlogging-operator/version"
)

const auditLoggingComponentName = "auditlogging_svc"
const auditLoggingReleaseName = "auditlogging"
const auditLoggingCrType = "auditlogging_cr"
const clusterRoleSuffix = "-role"
const clusterRoleBindingSuffix = "-rolebinding"
const productName = "IBM Cloud Platform Common Services"
const productID = "AuditLogging_3.5.0.0_Apache_00000"
const ServiceAcct = "-svcacct"
const defaultClusterIssuer = "cs-ca-clusterissuer"

var log = logf.Log.WithName("controller_auditlogging")
var seconds30 int64 = 30
var commonVolumes = []corev1.Volume{}

// BuildClusterRoleBinding returns a ClusterRoleBinding object
func BuildClusterRoleBindingForPolicyController(instance *operatorv1alpha1.AuditLogging) *rbacv1.ClusterRoleBinding {
	metaLabels := LabelsForMetadata(AuditPolicyControllerDeploy)
	rb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   AuditPolicyControllerDeploy + clusterRoleBindingSuffix,
			Labels: metaLabels,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      AuditPolicyControllerDeploy + ServiceAcct,
			Namespace: instance.Spec.InstanceNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     AuditPolicyControllerDeploy + clusterRoleSuffix,
		},
	}
	return rb
}

// BuildClusterRole returns a ClusterRole object
func BuildClusterRoleForPolicyController(instance *operatorv1alpha1.AuditLogging) *rbacv1.ClusterRole {
	metaLabels := LabelsForMetadata(AuditPolicyControllerDeploy)
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   AuditPolicyControllerDeploy + clusterRoleSuffix,
			Labels: metaLabels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{""},
				Resources: []string{"services"},
			},
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{""},
				Resources: []string{"secrets"},
			},
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"mutatingwebhookconfigurations", "validatingwebhookconfigurations"},
			},
			{
				Verbs:     []string{"get", "update", "patch"},
				APIGroups: []string{"audit.policies.ibm.com"},
				Resources: []string{"auditpolicies/status"},
			},
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{"audit.policies.ibm.com"},
				Resources: []string{"auditpolicies"},
			},
			{
				Verbs:     []string{"get", "update", "patch"},
				APIGroups: []string{"apps"},
				Resources: []string{"deployments/status"},
			},
			{
				Verbs:     []string{"get", "watch", "list", "create", "update", "patch", "delete"},
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
			},
			{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{""},
				Resources: []string{"pods", "namespaces"},
			},
			{
				Verbs:     []string{"get", "list", "watch", "update"},
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
			},
			{
				Verbs:     []string{"create", "get", "update", "patch"},
				APIGroups: []string{""},
				Resources: []string{"events"},
			},
			{
				Verbs:         []string{"use"},
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"anyuid"},
			},
		},
	}
	return cr
}

// BuildRoleBindingForFluentd returns a RoleBinding object for fluentd
func BuildRoleBindingForFluentd(instance *operatorv1alpha1.AuditLogging) *rbacv1.RoleBinding {
	metaLabels := LabelsForMetadata(FluentdName)
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDaemonSetName + clusterRoleBindingSuffix,
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    metaLabels,
		},
		Subjects: []rbacv1.Subject{{
			APIGroup:  "",
			Kind:      "ServiceAccount",
			Name:      FluentdDaemonSetName + ServiceAcct,
			Namespace: instance.Spec.InstanceNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     FluentdDaemonSetName + clusterRoleSuffix,
		},
	}
	return rb
}

// BuildRoleForFluentd returns a Role object for fluentd
func BuildRoleForFluentd(instance *operatorv1alpha1.AuditLogging) *rbacv1.Role {
	metaLabels := LabelsForMetadata(FluentdName)
	cr := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentdDaemonSetName + clusterRoleSuffix,
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    metaLabels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"use"},
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"privileged"},
			},
		},
	}
	return cr
}

// BuildConfigMap returns a ConfigMap object
func BuildConfigMap(instance *operatorv1alpha1.AuditLogging, name string) (*corev1.ConfigMap, error) {
	reqLogger := log.WithValues("ConfigMap.Namespace", instance.Spec.InstanceNamespace, "ConfigMap.Name", name)
	metaLabels := LabelsForMetadata(FluentdName)
	dataMap := make(map[string]string)
	var err error
	switch name {
	case FluentdDaemonSetName + "-" + ConfigName:
		dataMap[enableAuditLogForwardKey] = strconv.FormatBool(instance.Spec.Fluentd.EnableAuditLoggingForwarding)
		type Data struct {
			Value string `yaml:"fluent.conf"`
		}
		d := Data{}
		err = yaml.Unmarshal([]byte(fluentdMainConfigData), &d)
		dataMap[fluentdConfigKey] = d.Value
	case FluentdDaemonSetName + "-" + SourceConfigName:
		type DataS struct {
			Value string `yaml:"source.conf"`
		}
		ds := DataS{}
		var result string
		if instance.Spec.Fluentd.JournalPath != "" {
			result = sourceConfigData1 + instance.Spec.Fluentd.JournalPath + sourceConfigData2
		} else {
			result = sourceConfigData1 + journalPath + sourceConfigData2
		}
		err = yaml.Unmarshal([]byte(result), &ds)
		dataMap[sourceConfigKey] = ds.Value
	case FluentdDaemonSetName + "-" + SplunkConfigName:
		type DataSplunk struct {
			Value string `yaml:"splunkHEC.conf"`
		}
		dsplunk := DataSplunk{}
		err = yaml.Unmarshal([]byte(splunkConfigData), &dsplunk)
		if err != nil {
			reqLogger.Error(err, "Failed to unmarshall data for "+name)
		}
		dataMap[splunkConfigKey] = dsplunk.Value
	case FluentdDaemonSetName + "-" + QRadarConfigName:
		type DataQRadar struct {
			Value string `yaml:"remoteSyslog.conf"`
		}
		dq := DataQRadar{}
		err = yaml.Unmarshal([]byte(qRadarConfigData), &dq)
		dataMap[qRadarConfigKey] = dq.Value
	default:
		reqLogger.Info("Unknown ConfigMap name")
	}
	if err != nil {
		reqLogger.Error(err, "Failed to unmarshall data for "+name)
		return nil, err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Labels:    metaLabels,
			Namespace: instance.Spec.InstanceNamespace,
		},
		Data: dataMap,
	}
	return cm, nil
}

// BuildDeploymentForPolicyController returns a Deployment object
func BuildDeploymentForPolicyController(instance *operatorv1alpha1.AuditLogging) *appsv1.Deployment {
	reqLogger := log.WithValues("deploymentForPolicyController", "Entry", "instance.Name", instance.Name)
	metaLabels := LabelsForMetadata(AuditPolicyControllerDeploy)
	selectorLabels := LabelsForSelector(AuditPolicyControllerDeploy, instance.Name)
	podLabels := LabelsForPodMetadata(AuditPolicyControllerDeploy, instance.Name)
	annotations := annotationsForMetering(AuditPolicyControllerDeploy)

	var tag, imageRegistry string
	if instance.Spec.PolicyController.ImageRegistry != "" || instance.Spec.PolicyController.ImageTag != "" {
		if instance.Spec.PolicyController.ImageRegistry != "" {
			imageRegistry = instance.Spec.PolicyController.ImageRegistry
		} else {
			imageRegistry = defaultImageRegistry
		}
		if instance.Spec.PolicyController.ImageTag != "" {
			tag = instance.Spec.PolicyController.ImageTag
		} else {
			tag = defaultPCImageTag
		}
		policyControllerMainContainer.Image = imageRegistry + defaultPCImageName + ":" + tag
	}

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
	}
	if instance.Spec.PolicyController.Duration != "" {
		args = append(args, "--default-duration="+instance.Spec.PolicyController.Duration)
		reqLogger.Info("Test", "Duration", instance.Spec.PolicyController.Duration)
	}
	if instance.Spec.PolicyController.Frequency != "" {
		args = append(args, "--update-frequency="+instance.Spec.PolicyController.Frequency)
	}
	policyControllerMainContainer.Args = args

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuditPolicyControllerDeploy,
			Namespace: instance.Spec.InstanceNamespace,
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
					ServiceAccountName:            AuditPolicyControllerDeploy + ServiceAcct,
					TerminationGracePeriodSeconds: &seconds30,
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

// BuildCertsForAuditLogging returns a Certificate object
func BuildCertsForAuditLogging(instance *operatorv1alpha1.AuditLogging, issuer string) *certmgr.Certificate {
	metaLabels := LabelsForMetadata(FluentdName)
	var clusterIssuer string
	if issuer != "" {
		clusterIssuer = issuer
	} else {
		clusterIssuer = defaultClusterIssuer
	}

	certificate := &certmgr.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuditLoggingCertName,
			Labels:    metaLabels,
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: certmgr.CertificateSpec{
			CommonName: AuditLoggingCertName,
			SecretName: auditLoggingCertSecretName,
			IssuerRef: certmgr.ObjectReference{
				Name: clusterIssuer,
				Kind: certmgr.ClusterIssuerKind,
			},
		},
	}
	return certificate
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

	var tag, imageRegistry string
	if instance.Spec.Fluentd.ImageRegistry != "" || instance.Spec.Fluentd.ImageTag != "" {
		if instance.Spec.Fluentd.ImageRegistry != "" {
			imageRegistry = instance.Spec.Fluentd.ImageRegistry
		} else {
			imageRegistry = defaultImageRegistry
		}
		if instance.Spec.Fluentd.ImageTag != "" {
			tag = instance.Spec.Fluentd.ImageTag
		} else {
			tag = defaultFluentdImageTag
		}
		fluentdMainContainer.Image = imageRegistry + defaultFluentdImageName + ":" + tag
	}

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
			Namespace: instance.Spec.InstanceNamespace,
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
					ServiceAccountName:            FluentdDaemonSetName + ServiceAcct,
					TerminationGracePeriodSeconds: &seconds30,
					// NodeSelector:                  {},
					Tolerations: []corev1.Toleration{
						{
							Key:      "dedicated",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
					},
					Volumes: commonVolumes,
					Containers: []corev1.Container{
						fluentdMainContainer,
					},
				},
			},
		},
	}
	return daemon
}

// BuildCommonVolumeMounts returns an array of VolumeMount objects
func BuildCommonVolumeMounts(instance *operatorv1alpha1.AuditLogging) []corev1.VolumeMount {
	var journal = journalPath
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
			Name:      "certs",
			MountPath: "/fluentd/etc/tls",
			ReadOnly:  true,
		},
	}
	return commonVolumeMounts
}

// BuildCommonVolumes returns an array of Volume objects
func BuildCommonVolumes(instance *operatorv1alpha1.AuditLogging) []corev1.Volume {
	var journal = journalPath
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
							Key:  sourceConfigKey,
							Path: sourceConfigKey,
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
							Key:  qRadarConfigKey,
							Path: qRadarConfigKey,
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
							Key:  splunkConfigKey,
							Path: splunkConfigKey,
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
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: auditLoggingCertSecretName,
				},
			},
		},
	}
	return commonVolumes
}

func EqualCerts(expected *certmgr.Certificate, found *certmgr.Certificate) bool {
	return !reflect.DeepEqual(found.Spec, expected.Spec)
}

func EqualRoles(expected *rbacv1.Role, found *rbacv1.Role) bool {
	return !reflect.DeepEqual(found.Rules, expected.Rules)
}

func EqualClusterRoles(expected *rbacv1.ClusterRole, found *rbacv1.ClusterRole) bool {
	return !reflect.DeepEqual(found.Rules, expected.Rules)
}

func EqualRoleBindings(expected *rbacv1.RoleBinding, found *rbacv1.RoleBinding) bool {
	return !reflect.DeepEqual(found.Subjects, expected.Subjects) ||
		!reflect.DeepEqual(found.RoleRef, expected.RoleRef)
}

func EqualClusterRoleBindings(expected *rbacv1.ClusterRoleBinding, found *rbacv1.ClusterRoleBinding) bool {
	return !reflect.DeepEqual(found.Subjects, expected.Subjects) ||
		!reflect.DeepEqual(found.RoleRef, expected.RoleRef)
}

func EqualDeployments(expectedDeployment *appsv1.Deployment, foundDeployment *appsv1.Deployment) bool {
	return !reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Volumes, expectedDeployment.Spec.Template.Spec.Volumes) ||
		len(foundDeployment.Spec.Template.Spec.Containers) != len(expectedDeployment.Spec.Template.Spec.Containers) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Name, expectedDeployment.Spec.Template.Spec.Containers[0].Name) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Image, expectedDeployment.Spec.Template.Spec.Containers[0].Image) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy, expectedDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Args, expectedDeployment.Spec.Template.Spec.Containers[0].Args) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].VolumeMounts, expectedDeployment.Spec.Template.Spec.Containers[0].VolumeMounts) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].SecurityContext, expectedDeployment.Spec.Template.Spec.Containers[0].SecurityContext)
}

func EqualDaemonSets(expected *appsv1.DaemonSet, found *appsv1.DaemonSet) bool {
	return len(found.Spec.Template.Spec.Containers) != len(expected.Spec.Template.Spec.Containers) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].Name, expected.Spec.Template.Spec.Containers[0].Name) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].Image, expected.Spec.Template.Spec.Containers[0].Image) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].ImagePullPolicy, expected.Spec.Template.Spec.Containers[0].ImagePullPolicy) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].VolumeMounts, expected.Spec.Template.Spec.Containers[0].VolumeMounts) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].SecurityContext, expected.Spec.Template.Spec.Containers[0].SecurityContext)
}

// GetPodNames returns the pod names of the array of pods passed in
func GetPodNames(pods []corev1.Pod) []string {
	reqLogger := log.WithValues("func", "getPodNames")
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
		reqLogger.Info("CS??? pod name=" + pod.Name)
	}
	return podNames
}

//IBMDEV
func LabelsForMetadata(name string) map[string]string {
	return map[string]string{"app": name, "app.kubernetes.io/name": name, "app.kubernetes.io/component": auditLoggingComponentName,
		"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": auditLoggingReleaseName, "release": auditLoggingReleaseName}
}

//IBMDEV
func LabelsForSelector(name string, crName string) map[string]string {
	return map[string]string{"app": name, "component": auditLoggingComponentName, auditLoggingCrType: crName}
}

//IBMDEV
func LabelsForPodMetadata(deploymentName string, crName string) map[string]string {
	podLabels := LabelsForMetadata(deploymentName)
	selectorLabels := LabelsForSelector(deploymentName, crName)
	for key, value := range selectorLabels {
		podLabels[key] = value
	}
	return podLabels
}

//IBMDEV
func annotationsForMetering(deploymentName string) map[string]string {
	annotations := map[string]string{
		"productName":    productName,
		"productVersion": ver.Version,
		"productID":      productID,
	}
	if deploymentName == FluentdName {
		annotations["seccomp.security.alpha.kubernetes.io/pod"] = "docker/default"
	}
	return annotations
}
