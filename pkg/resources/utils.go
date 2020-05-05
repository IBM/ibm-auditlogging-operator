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
	"regexp"
	"strconv"
	"strings"

	operatorv1alpha1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1alpha1"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	yaml "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const AuditLoggingComponentName = "common-audit-logging"
const auditLoggingReleaseName = "common-audit-logging"
const auditLoggingCrType = "auditlogging_cr"
const productName = "IBM Cloud Platform Common Services"
const productID = "068a62892a1e4db39641342e592daa25"
const productVersion = "3.3.0"
const productMetric = "FREE"

const InstanceNamespace = "ibm-common-services"

var architectureList = []string{"amd64", "ppc64le", "s390x"}

var DefaultStatusForCR = []string{"none"}
var log = logf.Log.WithName("controller_auditlogging")
var seconds30 int64 = 30
var commonVolumes = []corev1.Volume{}

type DataSplunk struct {
	Value string `yaml:"splunkHEC.conf"`
}

type DataQRadar struct {
	Value string `yaml:"remoteSyslog.conf"`
}

// BuildAuditService returns a Service object
func BuildAuditService(instance *operatorv1alpha1.AuditLogging) *corev1.Service {
	metaLabels := LabelsForMetadata(FluentdName)
	selectorLabels := LabelsForSelector(FluentdName, instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuditLoggingComponentName,
			Namespace: InstanceNamespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{
					Name:     AuditLoggingComponentName,
					Protocol: "TCP",
					Port:     defaultHTTPPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: defaultHTTPPort,
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
		p := strconv.Itoa(defaultHTTPPort)
		result += sourceConfigData3 + p + sourceConfigData4
		err = yaml.Unmarshal([]byte(result), &ds)
		dataMap[SourceConfigKey] = ds.Value
	case FluentdDaemonSetName + "-" + SplunkConfigName:
		dsplunk := DataSplunk{}
		err = yaml.Unmarshal([]byte(splunkConfigData1+splunkDefaults+splunkConfigData2), &dsplunk)
		if err != nil {
			reqLogger.Error(err, "Failed to unmarshall data for "+name)
		}
		dataMap[SplunkConfigKey] = dsplunk.Value
	case FluentdDaemonSetName + "-" + QRadarConfigName:
		dq := DataQRadar{}
		err = yaml.Unmarshal([]byte(qRadarConfigData1+qRadarDefaults+qRadarConfigData2), &dq)
		dataMap[QRadarConfigKey] = dq.Value
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

	if instance.Spec.PolicyController.ImageRegistry != "" {
		imageRegistry := instance.Spec.PolicyController.ImageRegistry
		if string(imageRegistry[len(imageRegistry)-1]) != "/" {
			imageRegistry += "/"
		}
		policyControllerMainContainer.Image = imageRegistry + defaultPCImageName + ":" + defaultPCImageTag
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
		certificate.Spec.DNSNames = []string{AuditLoggingComponentName}
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

	if instance.Spec.Fluentd.ImageRegistry != "" {
		imageRegistry := instance.Spec.Fluentd.ImageRegistry
		if string(imageRegistry[len(imageRegistry)-1]) != "/" {
			imageRegistry += "/"
		}
		fluentdMainContainer.Image = imageRegistry + defaultFluentdImageName + ":" + defaultFluentdImageTag
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

func EqualServices(expected *corev1.Service, found *corev1.Service) bool {
	return !reflect.DeepEqual(found.Spec.Ports, expected.Spec.Ports)
}

func EqualCerts(expected *certmgr.Certificate, found *certmgr.Certificate) bool {
	return !reflect.DeepEqual(found.Spec, expected.Spec)
}

func EqualDeployments(expected *appsv1.Deployment, found *appsv1.Deployment) bool {
	return EqualPods(expected.Spec.Template, found.Spec.Template)
}

func EqualDaemonSets(expected *appsv1.DaemonSet, found *appsv1.DaemonSet) bool {
	return EqualPods(expected.Spec.Template, found.Spec.Template)
}

func EqualPods(expected corev1.PodTemplateSpec, found corev1.PodTemplateSpec) bool {
	logger := log.WithValues("func", "EqualPods")
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

func EqualContainers(expected corev1.Container, found corev1.Container) bool {
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
	return true
}

func EqualMatchTags(found *corev1.ConfigMap) bool {
	logger := log.WithValues("func", "EqualMatchTags")
	var key string
	if found.Name == FluentdDaemonSetName+"-"+SplunkConfigName {
		key = SplunkConfigKey
	} else {
		key = QRadarConfigKey
	}
	re := regexp.MustCompile(`match icp-audit icp-audit\.\*\*`)
	var match = re.FindStringSubmatch(found.Data[key])
	if len(match) < 1 {
		logger.Info("Match tags not equal", "Expected", OutputPluginMatches)
		return false
	}
	return len(match) >= 1
}

func EqualSourceConfig(expected *corev1.ConfigMap, found *corev1.ConfigMap) (bool, []string) {
	var ports []string
	var foundPort string
	re := regexp.MustCompile("port [0-9]+")
	var match = re.FindStringSubmatch(found.Data[SourceConfigKey])
	if len(match) < 1 {
		foundPort = ""
	} else {
		foundPort = strings.Split(match[0], " ")[1]
	}
	ports = append(ports, foundPort)
	match = re.FindStringSubmatch(expected.Data[SourceConfigKey])
	expectedPort := strings.Split(match[0], " ")[1]
	return (foundPort == expectedPort), append(ports, expectedPort)
}

func EqualLabels(found map[string]string, expected map[string]string) bool {
	logger := log.WithValues("func", "EqualLabels")
	if !reflect.DeepEqual(found, expected) {
		logger.Info("Labels not equal", "Found", found, "Expected", expected)
		return false
	}
	return true
}

func BuildWithSIEMCreds(found *corev1.ConfigMap) (string, error) {
	var key string
	var creds string
	var result string
	var err error
	if found.Name == FluentdDaemonSetName+"-"+SplunkConfigName {
		key = SplunkConfigKey
		reHost := regexp.MustCompile(`hec_host .*`)
		host := reHost.FindStringSubmatch(found.Data[key])[0]
		rePort := regexp.MustCompile(`hec_port .*`)
		port := rePort.FindStringSubmatch(found.Data[key])[0]
		reToken := regexp.MustCompile(`hec_token .*`)
		token := reToken.FindStringSubmatch(found.Data[key])[0]
		creds = yamlLine(2, host, true) + yamlLine(2, port, true) + yamlLine(2, token, false)
		ds := DataSplunk{}
		err = yaml.Unmarshal([]byte(splunkConfigData1+"\n"+creds+splunkConfigData2), &ds)
		result = ds.Value
	} else {
		key = QRadarConfigKey
		reHost := regexp.MustCompile(`host .*`)
		host := reHost.FindStringSubmatch(found.Data[key])[0]
		rePort := regexp.MustCompile(`port .*`)
		port := rePort.FindStringSubmatch(found.Data[key])[0]
		reHostname := regexp.MustCompile(`hostname .*`)
		hostname := reHostname.FindStringSubmatch(found.Data[key])[0]
		creds = yamlLine(3, host, true) + yamlLine(3, port, true) + yamlLine(3, hostname, false)
		dq := DataQRadar{}
		err = yaml.Unmarshal([]byte(qRadarConfigData1+"\n"+creds+qRadarConfigData2), &dq)
		result = dq.Value
	}
	return result, err
}

func yamlLine(tabs int, line string, newline bool) string {
	spaces := strings.Repeat(`    `, tabs)
	if !newline {
		return spaces + line
	}
	return spaces + line + "\n"
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
	return map[string]string{"app": name, "app.kubernetes.io/name": name, "app.kubernetes.io/component": AuditLoggingComponentName,
		"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": auditLoggingReleaseName, "release": auditLoggingReleaseName}
}

//IBMDEV
func LabelsForSelector(name string, crName string) map[string]string {
	return map[string]string{"app": name, "component": AuditLoggingComponentName, auditLoggingCrType: crName}
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
