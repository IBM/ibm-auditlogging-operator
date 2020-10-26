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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"
	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	"github.com/IBM/ibm-auditlogging-operator/controllers/util"
)

const splunkKey = "splunkCA.pem"
const qRadarKey = "qradar.crt"

// BuildSecret returns a Secret object
func BuildSecret(instance *operatorv1.CommonAudit) *corev1.Secret {
	metaLabels := util.LabelsForMetadata(constant.FluentdName)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuditLoggingClientCertSecName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Data: map[string][]byte{
			splunkKey: {},
			qRadarKey: {},
		},
	}
	return secret
}
