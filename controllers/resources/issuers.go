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

	certmgr "github.com/ibm/ibm-cert-manager-operator/apis/certmanager/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GodIssuer = "audit-god-issuer"
const RootIssuer = "audit-root-ca-issuer"

// BuildGodIssuer returns an Issuer object
func BuildGodIssuer(namespace string) *certmgr.Issuer {
	metaLabels := utils.LabelsForMetadata(constant.FluentdName)
	return &certmgr.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GodIssuer,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Spec: certmgr.IssuerSpec{
			IssuerConfig: certmgr.IssuerConfig{
				SelfSigned: &certmgr.SelfSignedIssuer{},
			},
		},
	}
}

// BuildRootCAIssuer returns an Issuer object
func BuildRootCAIssuer(namespace string) *certmgr.Issuer {
	metaLabels := utils.LabelsForMetadata(constant.FluentdName)
	return &certmgr.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RootIssuer,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Spec: certmgr.IssuerSpec{
			IssuerConfig: certmgr.IssuerConfig{
				CA: &certmgr.CAIssuer{
					SecretName: RootCert,
				},
			},
		},
	}
}

// EqualIssuers returns a boolean
func EqualIssuers(expected *certmgr.Issuer, found *certmgr.Issuer) bool {
	return !reflect.DeepEqual(expected.Spec, found.Spec)
}
