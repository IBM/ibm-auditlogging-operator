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
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const auditLoggingComponentName = "common-audit-logging"
const auditLoggingReleaseName = "common-audit-logging"
const auditLoggingCrType = "auditlogging_cr"
const productName = "IBM Cloud Platform Common Services"
const productID = "068a62892a1e4db39641342e592daa25"
const productVersion = "3.3.0"
const productMetric = "FREE"

const InstanceNamespace = "ibm-common-services"

var DefaultStatusForCR = []string{"none"}
var log = logf.Log.WithName("controller_auditlogging")
var seconds30 int64 = 30
var commonVolumes = []corev1.Volume{}

// BuildAuditService returns a Service object
func BuildAuditService(instance *operatorv1alpha1.AuditLogging) *corev1.Service {
	metaLabels := LabelsForMetadata(FluentdName)
	selectorLabels := LabelsForSelector(FluentdName, instance.Name)

	var httpPort int32
	if res, port := getHTTPPort(instance.Spec.Fluentd.HTTPPort); res {
		httpPort = port
	} else {
		httpPort = defaultHTTPPort
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      auditLoggingComponentName,
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{
					Name:     auditLoggingComponentName,
					Protocol: "TCP",
					Port:     httpPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: httpPort,
					},
				},
			},
			Selector: selectorLabels,
		},
	}
	return service
}

// BuildAuditPolicyCRD returns a CRD object
func BuildAuditPolicyCRD(instance *operatorv1alpha1.AuditLogging) *extv1beta1.CustomResourceDefinition {
	metaLabels := LabelsForMetadata(AuditPolicyControllerDeploy)
	metaLabels["controller-tools.k8s.io"] = "1.0"
	crd := &extv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   AuditPolicyCRDName,
			Labels: metaLabels,
		},
		Spec: extv1beta1.CustomResourceDefinitionSpec{
			Group: "audit.policies.ibm.com",
			Names: extv1beta1.CustomResourceDefinitionNames{
				Kind:       "AuditPolicy",
				Plural:     "auditpolicies",
				ShortNames: []string{"ap"},
			},
			Scope: "Namespaced",
			Validation: &extv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &extv1beta1.JSONSchemaProps{
					Properties: map[string]extv1beta1.JSONSchemaProps{
						"apiVersion": {
							Description: "APIVersion defines the versioned schema of this representation of an object. " +
								"Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. " +
								"More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type: "string",
						},
						"kind": {
							Description: "'Kind is a string value representing the REST resource this object represents. " +
								"Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. " +
								"More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type: "string",
						},
						"metadata": {
							Type: "object",
						},
						"spec": {
							Properties: map[string]extv1beta1.JSONSchemaProps{
								"labelSelector": {
									Description: "selecting a list of namespaces where the policy applies",
									Type:        "object",
								},
								"namespaceSelector": {
									Description: "namespaces on which to run the policy",
									Properties: map[string]extv1beta1.JSONSchemaProps{
										"exclude": {
											Items: &extv1beta1.JSONSchemaPropsOrArray{
												Schema: &extv1beta1.JSONSchemaProps{
													Type: "string",
												},
											},
											Type: "array",
										},
										"include": {
											Items: &extv1beta1.JSONSchemaPropsOrArray{
												Schema: &extv1beta1.JSONSchemaProps{
													Type: "string",
												},
											},
											Type: "array",
										},
									},
									Type: "object",
								},
								"remediationAction": {
									Description: "remediate or enforce",
									Type:        "string",
								},
								"clusterAuditPolicy": {
									Description: "enforce, inform",
									Type:        "object",
								},
							},
							Type: "object",
						},
						"status": {
							Properties: map[string]extv1beta1.JSONSchemaProps{
								"auditDetails": {
									Description: "selecting a list of services to validate",
									Type:        "object",
								},
								"compliant": {
									Type: "string",
								},
							},
							Type: "object",
						},
					},
				},
			},
			Version: "v1alpha1",
		},
		Status: extv1beta1.CustomResourceDefinitionStatus{
			AcceptedNames: extv1beta1.CustomResourceDefinitionNames{
				Kind:   "",
				Plural: "",
			},
			Conditions:     []extv1beta1.CustomResourceDefinitionCondition{},
			StoredVersions: []string{},
		},
	}

	return crd
}

// BuildConfigMap returns a ConfigMap object
func BuildConfigMap(instance *operatorv1alpha1.AuditLogging, name string) (*corev1.ConfigMap, error) {
	reqLogger := log.WithValues("ConfigMap.Namespace", InstanceNamespace, "ConfigMap.Name", name)
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
			result = sourceConfigData1 + defaultJournalPath + sourceConfigData2
		}
		var p string
		if res, port := getHTTPPort(instance.Spec.Fluentd.HTTPPort); res {
			p = strconv.Itoa(int(port))
		} else {
			p = strconv.Itoa(defaultHTTPPort)
		}
		result += sourceConfigData3 + p + sourceConfigData4
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
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
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
					ServiceAccountName:            OperandRBAC,
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
func BuildCertsForAuditLogging(instance *operatorv1alpha1.AuditLogging, issuer string, name string) *certmgr.Certificate {
	metaLabels := LabelsForMetadata(FluentdName)
	var clusterIssuer string
	if issuer != "" {
		clusterIssuer = issuer
	} else {
		clusterIssuer = defaultClusterIssuer
	}

	certificate := &certmgr.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
		},
		Spec: certmgr.CertificateSpec{
			CommonName: name,
			IssuerRef: certmgr.ObjectReference{
				Name: clusterIssuer,
				Kind: certmgr.ClusterIssuerKind,
			},
		},
	}

	if name == AuditLoggingHTTPSCertName {
		certificate.Spec.SecretName = AuditLoggingServerCertSecName
		certificate.Spec.DNSNames = []string{auditLoggingComponentName}
	} else {
		certificate.Spec.SecretName = AuditLoggingClientCertSecName
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

	if result, port := getHTTPPort(instance.Spec.Fluentd.HTTPPort); result {
		fluentdMainContainer.Ports[0].ContainerPort = port
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
					ServiceAccountName:            OperandRBAC,
					TerminationGracePeriodSeconds: &seconds30,
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

func getHTTPPort(port string) (bool, int32) {
	if port == "" {
		return false, 0
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return false, 0
	}
	if p > 0 && p <= 65535 {
		return true, int32(p)
	}
	return false, 0
}

func EqualServices(expected *corev1.Service, found *corev1.Service) bool {
	return !reflect.DeepEqual(found.Spec.Ports, expected.Spec.Ports)
}

func EqualCerts(expected *certmgr.Certificate, found *certmgr.Certificate) bool {
	return !reflect.DeepEqual(found.Spec, expected.Spec)
}

func EqualDeployments(expectedDeployment *appsv1.Deployment, foundDeployment *appsv1.Deployment) bool {
	return !reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Volumes, expectedDeployment.Spec.Template.Spec.Volumes) ||
		len(foundDeployment.Spec.Template.Spec.Containers) != len(expectedDeployment.Spec.Template.Spec.Containers) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Name, expectedDeployment.Spec.Template.Spec.Containers[0].Name) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Image, expectedDeployment.Spec.Template.Spec.Containers[0].Image) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy, expectedDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Args, expectedDeployment.Spec.Template.Spec.Containers[0].Args) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].VolumeMounts, expectedDeployment.Spec.Template.Spec.Containers[0].VolumeMounts) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].SecurityContext, expectedDeployment.Spec.Template.Spec.Containers[0].SecurityContext) ||
		!reflect.DeepEqual(foundDeployment.Spec.Template.Spec.ServiceAccountName, expectedDeployment.Spec.Template.Spec.ServiceAccountName)
}

func EqualDaemonSets(expected *appsv1.DaemonSet, found *appsv1.DaemonSet) bool {
	return len(found.Spec.Template.Spec.Containers) != len(expected.Spec.Template.Spec.Containers) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].Name, expected.Spec.Template.Spec.Containers[0].Name) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].Image, expected.Spec.Template.Spec.Containers[0].Image) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].ImagePullPolicy, expected.Spec.Template.Spec.Containers[0].ImagePullPolicy) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].VolumeMounts, expected.Spec.Template.Spec.Containers[0].VolumeMounts) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].SecurityContext, expected.Spec.Template.Spec.Containers[0].SecurityContext) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.ServiceAccountName, expected.Spec.Template.Spec.ServiceAccountName)
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
		"productID":      productID,
		"productVersion": productVersion,
		"productMetric":  productMetric,
	}
	if deploymentName == FluentdName {
		annotations["seccomp.security.alpha.kubernetes.io/pod"] = "docker/default"
		annotations["clusterhealth.ibm.com/dependencies"] = "cert-manager"
		annotations["openshift.io/scc"] = "privileged"
	} else {
		annotations["openshift.io/scc"] = "restricted"
	}
	return annotations
}
