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

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	utils "github.com/IBM/ibm-auditlogging-operator/controllers/util"

	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const AuditLoggingClientCertSecName = "audit-certs"
const AuditLoggingHTTPSCertName = "fluentd-https"
const AuditLoggingServerCertSecName = "audit-server-certs"
const AuditLoggingCertName = "fluentd"
const DefaultIssuer = "cs-ca-issuer"
const RootCert = "audit-root-ca-cert"

// BuildCertsForAuditLogging returns a Certificate object
func BuildCertsForAuditLogging(namespace string, issuer string, name string) *certmgr.Certificate {
	metaLabels := utils.LabelsForMetadata(constant.FluentdName)
	var certIssuer string
	if issuer != "" {
		certIssuer = issuer
	} else {
		certIssuer = DefaultIssuer
	}

	certificate := &certmgr.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Spec: certmgr.CertificateSpec{
			CommonName: name,
			IssuerRef: certmgr.ObjectReference{
				Name: certIssuer,
				Kind: certmgr.IssuerKind,
			},
		},
	}

	if name == AuditLoggingHTTPSCertName {
		certificate.Spec.SecretName = AuditLoggingServerCertSecName
		certificate.Spec.DNSNames = []string{
			constant.AuditLoggingComponentName,
			constant.AuditLoggingComponentName + "." + namespace,
			constant.AuditLoggingComponentName + "." + namespace + ".svc.cluster.local",
		}
	} else {
		certificate.Spec.SecretName = AuditLoggingClientCertSecName
	}
	return certificate
}

// BuildRootCACert returns a Certificate object
func BuildRootCACert(namespace string) *certmgr.Certificate {
	metaLabels := utils.LabelsForMetadata(constant.FluentdName)
	return &certmgr.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RootCert,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Spec: certmgr.CertificateSpec{
			SecretName: RootCert,
			IsCA:       true,
			CommonName: RootCert,
			IssuerRef: certmgr.ObjectReference{
				Name: GodIssuer,
				Kind: certmgr.IssuerKind,
			},
		},
	}
}

// EqualCerts returns a Boolean
func EqualCerts(expected *certmgr.Certificate, found *certmgr.Certificate) bool {
	return !reflect.DeepEqual(found.Spec, expected.Spec)
}
