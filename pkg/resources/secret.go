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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	operatorv1 "github.com/ibm/ibm-auditlogging-operator/pkg/apis/operator/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const splunkKey = "splunkCA.pem"
const qRadarKey = "qradar.crt"

// BuildSecret returns a Secret object
func BuildSecret(instance *operatorv1.CommonAudit) *corev1.Secret {
	metaLabels := LabelsForMetadata(FluentdName)
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

func createKeyValuePairs(m map[string][]byte) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b := new(bytes.Buffer)
	for _, k := range keys {
		fmt.Fprintf(b, "%s=\"%s\"\n", k, m[k])
	}
	return b.String()
}

// GenerateHash returns the string value of the SHA256 checksum for sec.Data
func GenerateHash(sec *corev1.Secret) string {
	logger := log.WithValues("func", "GenerateHash")
	h := sha256.New()
	data := createKeyValuePairs(sec.Data)
	if _, err := h.Write([]byte(data)); err != nil {
		logger.Error(err, "Failed to generate hash for secret.")
	}
	return hex.EncodeToString(h.Sum(nil))
}
