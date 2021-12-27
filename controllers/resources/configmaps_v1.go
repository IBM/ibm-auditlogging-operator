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
	"regexp"
	"strconv"
	"strings"

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"
	"github.com/IBM/ibm-auditlogging-operator/controllers/util"

	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"
)

// operatorv1 output config constants
var fluentdMainConfigV1Data = `
fluent.conf: |-
    # Input plugins (Supports Systemd and HTTP)
    @include /fluentd/etc/source.conf
    # Output plugins (Supports Splunk and Syslog)`
var fluentdOutputConfigV1Data = `
    <match icp-audit icp-audit.** records records.** syslog syslog.**>
        @type copy
`
var splunkConfigV1Data1 = `
splunkHEC.conf: |-
    <store>
        @type splunk_hec
`
var splunkConfigV1Data2 = `
        ca_file /fluentd/etc/tls/splunkCA.pem
        source ${tag}
    </store>`
var qRadarConfigV1Data1 = `
remoteSyslog.conf: |-
    <store>
        @type remote_syslog
`
var qRadarConfigV1Data2 = `
        protocol tcp
        ca_file /fluentd/etc/tls/qradar.crt
        packet_size 4096
        program fluentd
        <format>
            @type single_value
            message_key message
        </format>
    </store>`

var logDNAConfigV1Data1 = `
logDNA.conf: |-
    <store>
        @type logdna2
`
var logDNAConfigV1Data2 = `
        buffer_chunk_limit 1m
        flush_at_shutdown true
        <buffer>
            flush_mode interval
            retry_type exponential_backoff
            chunk_limit_size 1m
            flush_at_shutdown true
        </buffer>
    </store>`
var sourceConfigSyslog = `
    <source>
        @type syslog
        port 5140
        bind 0.0.0.0
        tag syslog
        <transport tls>
            ca_path /fluentd/etc/https/ca.crt
            cert_path /fluentd/etc/https/tls.crt
            private_key_path /fluentd/etc/https/tls.key
        </transport>
        <parse>
            @type regexp
            expression /^[^{]*(?<message>{.*})\s*$/
            types message:string
        </parse>
    </source>
`

var filterSyslog = `
    <filter syslog syslog.**>
        @type parser
        format json
        key_name message
        reserve_data true
    </filter>
`

var sourceConfigDataZenHTTP1 = `
    <source>
        @type http2
        # Tag is not supported in yaml, must be set by request path (/icp-audit.http is required for validation and export)
        port `
var sourceConfigDataZenHTTP2 = `
        bind 0.0.0.0
        body_size_limit 32m
        keepalive_timeout 10s
        <transport tls>
          ca_path /etc/internal-tls/certificate.pem
          cert_path /etc/internal-tls/tls.crt
          private_key_path /etc/internal-tls/tls.key
          client_cert_auth true
        </transport>
        <parse>
          @type json
        </parse>
    </source>
`
var sourceConfigDataZenHTTP21 = `
    <source>
        @type http2
        # Tag is not supported in yaml, must be set by request path (/icp-audit.http is required for validation and export)
        port `
var sourceConfigDataZenHTTP22 = `
        bind 0.0.0.0
        body_size_limit 32m
        keepalive_timeout 10s
        basic_auth true
        basic_auth_type file
        basic_auth_file /etc/zen-service-broker-secret/token
        basic_auth_file_type full
        <transport tls>
          ca_path /etc/internal-tls/certificate.pem
          cert_path /etc/internal-tls/tls.crt
          private_key_path /etc/internal-tls/tls.key
        </transport>
        <parse>
          @type json
        </parse>
    </source>
`
var filterCADFields = `
    <filter records records.**>
        @type cadfFields
    </filter>
`

const hecHost = `hec_host `
const hecPort = `hec_port `
const hecToken = `hec_token `
const host = `host `
const port = `port `
const hostname = `hostname `
const protocol = `protocol `
const tls = `tls `
const apiKey = `api_key `
const ingesterDomain = `ingester_domain `
const app = `app `

var RegexHecHost = regexp.MustCompile(hecHost + `.*`)
var RegexHecPort = regexp.MustCompile(hecPort + `.*`)
var RegexHecToken = regexp.MustCompile(hecToken + `.*`)
var RegexHost = regexp.MustCompile(host + `.*`)
var RegexPort = regexp.MustCompile(port + `.*`)
var RegexHostname = regexp.MustCompile(hostname + `.*`)
var RegexProtocol = regexp.MustCompile(protocol + `.*`)
var RegexTLS = regexp.MustCompile(tls + `.*`)
var RegexApiKey = regexp.MustCompile(apiKey + `.*`)
var RegexIngesterDomain = regexp.MustCompile(ingesterDomain + `.*`)
var RegexApp = regexp.MustCompile(app + `.*`)

var QradarPlugin = `@include /fluentd/etc/remoteSyslog.conf`
var SplunkPlugin = `@include /fluentd/etc/splunkHEC.conf`
var LogDNAPlugin = `@include /fluentd/etc/logDNA.conf`

const matchTags = `<match icp-audit icp-audit.**>`

var Protocols = map[bool]string{
	true:  "https",
	false: "http",
}

// SyslogIngestURLKey defines the Http endpoint
const SyslogIngestURLKey = "AuditLoggingSyslogIngestURL"

// New Endpoints configured with certificates issued by Issuer:  zen-tls-issuer
const ZenSyslogIngestURLKey = "AuditZenSyslogIngestURL"
const ZenHTTPBasicAuthIngestURLKey = "AuditZenHttpBasicAuthIngestURL"
const ZenHTTPMutualAuthIngestURLKey = "AuditZenHttpMutualAuthIngestURL"

// Zen Endpoints configured with certificates issued by Issuer:  zen-tls-issuer
const ZenSvcSyslogIngestURLKey = "AuditZenSvcSyslogIngestURL"
const ZenSvcHTTPBasicAuthIngestURLKey = "AuditZenSvcHttpBasicAuthIngestURL"
const ZenSvcHTTPMutualAuthIngestURLKey = "AuditZenSvcHttpMutualAuthIngestURL"

const zenHttpPath = "/records"

// BuildFluentdConfigMap returns a ConfigMap object
func BuildFluentdConfigMap(instance *operatorv1.CommonAudit, name string) (*corev1.ConfigMap, error) {
	log.WithValues("ConfigMap.Namespace", instance.Namespace, "ConfigMap.Name", name)
	metaLabels := util.LabelsForMetadata(constant.FluentdName)
	dataMap := make(map[string]string)
	var err error
	var data string
	switch name {
	case FluentdDaemonSetName + "-" + ConfigName:
		dataMap[EnableAuditLogForwardKey] = strconv.FormatBool(instance.Spec.EnableAuditLoggingForwarding)
		d := Data{}
		data = buildFluentdConfig(instance)
		err = yaml.Unmarshal([]byte(data), &d)
		if err != nil {
			break
		}
		dataMap[FluentdConfigKey] = d.Value
	case FluentdDaemonSetName + "-" + SourceConfigName:
		ds := DataS{}
		var result string
		p := strconv.Itoa(defaultHTTPPort)
		mutualAuthPort := strconv.Itoa(mutualCertAuthHTTP2Port)
		basicAuthPort := strconv.Itoa(basicAuthHTTP2Port)
		if instance.Spec.Fluentd.ZenEnabled {
			result += sourceConfigDataKey + sourceConfigDataHTTP1 + p + sourceConfigDataHTTP2 + sourceConfigDataZenHTTP1 + mutualAuthPort + sourceConfigDataZenHTTP2 + sourceConfigDataZenHTTP21 + basicAuthPort + sourceConfigDataZenHTTP22 + sourceConfigSyslog + filterHTTP + filterSyslog + filterCADFields
		} else {
			result += sourceConfigDataKey + sourceConfigDataHTTP1 + p + sourceConfigDataHTTP2 + sourceConfigSyslog + filterHTTP + filterSyslog
		}
		err = yaml.Unmarshal([]byte(result), &ds)
		if err != nil {
			break
		}
		dataMap[SourceConfigKey] = ds.Value
	case FluentdDaemonSetName + "-" + SplunkConfigName:
		dsplunk := DataSplunk{}
		data = buildFluentdSplunkConfig(instance)
		err = yaml.Unmarshal([]byte(data), &dsplunk)
		if err != nil {
			break
		}
		dataMap[SplunkConfigKey] = dsplunk.Value
	case FluentdDaemonSetName + "-" + QRadarConfigName:
		dq := DataQRadar{}
		data = buildFluentdQRadarConfig(instance)
		err = yaml.Unmarshal([]byte(data), &dq)
		if err != nil {
			break
		}
		dataMap[QRadarConfigKey] = dq.Value
	case FluentdDaemonSetName + "-" + LogDNAConfigName:
		dq := DataLogDNA{}
		data = buildFluentdLogDNAConfig(instance)
		err = yaml.Unmarshal([]byte(data), &dq)
		if err != nil {
			break
		}
		dataMap[LogDNAConfigKey] = dq.Value
	case FluentdDaemonSetName + "-" + HTTPIngestName:
		var p = strconv.Itoa(defaultHTTPPort)
		dataMap[HTTPIngestURLKey] = "https://" + constant.AuditLoggingComponentName + "." + instance.Namespace + ":" + p + httpPath
		p = strconv.Itoa(defaultSyslogPort)
		dataMap[SyslogIngestURLKey] = "https://" + constant.AuditLoggingComponentName + "." + instance.Namespace + ":" + p

		// New Endpoints configured with certificates issued by Issuer:  zen-tls-issuer
		p = strconv.Itoa(mutualCertAuthHTTP2Port)
		dataMap[ZenHTTPMutualAuthIngestURLKey] = "https://" + constant.AuditLoggingComponentName + "." + instance.Namespace + ":" + p + httpPath
		p = strconv.Itoa(basicAuthHTTP2Port)
		dataMap[ZenHTTPBasicAuthIngestURLKey] = "https://" + constant.AuditLoggingComponentName + "." + instance.Namespace + ":" + p + httpPath
		p = strconv.Itoa(zenSyslogPort)
		dataMap[ZenSyslogIngestURLKey] = "https://" + constant.AuditLoggingComponentName + "." + instance.Namespace + ":" + p

		// Zen Endpoints configured with certificates issued by Issuer:  zen-tls-issuer
		p = strconv.Itoa(defaultHTTPPort)
		dataMap[ZenSvcHTTPMutualAuthIngestURLKey] = "https://" + constant.ZenAuditService + "." + instance.Namespace + ":" + p + zenHttpPath
		p = strconv.Itoa(basicAuthHTTP2Port)
		dataMap[ZenSvcHTTPBasicAuthIngestURLKey] = "https://" + constant.ZenAuditService + "." + instance.Namespace + ":" + p + zenHttpPath
		p = strconv.Itoa(defaultSyslogPort)
		dataMap[ZenSvcSyslogIngestURLKey] = "https://" + constant.ZenAuditService + "." + instance.Namespace + ":" + p
	default:
		log.Info("Unknown ConfigMap name")
	}
	if err != nil {
		log.Error(err, "Failed to unmarshall data for "+name)
		return nil, err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Data: dataMap,
	}
	return cm, nil
}

func buildFluentdConfig(instance *operatorv1.CommonAudit) string {
	var result = fluentdMainConfigV1Data
	if instance.Spec.Outputs.Splunk.EnableSIEM || instance.Spec.Outputs.Syslog.EnableSIEM || instance.Spec.Outputs.LogDNA.EnableSIEM {
		result += fluentdOutputConfigV1Data
		if instance.Spec.Outputs.Splunk.EnableSIEM {
			result += yamlLine(2, SplunkPlugin, true)
		}
		if instance.Spec.Outputs.Syslog.EnableSIEM {
			result += yamlLine(2, QradarPlugin, true)
		}
		if instance.Spec.Outputs.LogDNA.EnableSIEM {
			result += yamlLine(2, LogDNAPlugin, true)
		}
		result += yamlLine(1, `</match>`, false)
	}
	return result
}

func buildFluentdSplunkConfig(instance *operatorv1.CommonAudit) string {
	var result = splunkConfigV1Data1
	if instance.Spec.Outputs.Splunk.Host != "" && instance.Spec.Outputs.Splunk.Port != 0 &&
		instance.Spec.Outputs.Splunk.Token != "" {
		result += yamlLine(2, hecHost+instance.Spec.Outputs.Splunk.Host, true)
		result += yamlLine(2, hecPort+strconv.Itoa(int(instance.Spec.Outputs.Splunk.Port)), true)
		result += yamlLine(2, hecToken+instance.Spec.Outputs.Splunk.Token, true)
	} else {
		result += yamlLine(2, hecHost+`SPLUNK_SERVER_HOSTNAME`, true)
		result += yamlLine(2, hecPort+`SPLUNK_PORT`, true)
		result += yamlLine(2, hecToken+`SPLUNK_HEC_TOKEN`, true)
	}
	result += yamlLine(2, protocol+Protocols[instance.Spec.Outputs.Splunk.TLS], false)
	return result + splunkConfigV1Data2
}

func buildFluentdQRadarConfig(instance *operatorv1.CommonAudit) string {
	var result = qRadarConfigV1Data1
	if instance.Spec.Outputs.Syslog.Host != "" && instance.Spec.Outputs.Syslog.Port != 0 &&
		instance.Spec.Outputs.Syslog.Hostname != "" {
		result += yamlLine(2, host+instance.Spec.Outputs.Syslog.Host, true)
		result += yamlLine(2, port+strconv.Itoa(int(instance.Spec.Outputs.Syslog.Port)), true)
		result += yamlLine(2, hostname+instance.Spec.Outputs.Syslog.Hostname, true)
	} else {
		result += yamlLine(2, host+`QRADAR_SERVER_HOSTNAME`, true)
		result += yamlLine(2, port+`QRADAR_PORT_FOR_icp-audit`, true)
		result += yamlLine(2, hostname+`QRADAR_LOG_SOURCE_IDENTIFIER_FOR_icp-audit`, true)
	}
	result += yamlLine(2, tls+strconv.FormatBool(instance.Spec.Outputs.Syslog.TLS), false)
	return result + qRadarConfigV1Data2
}

func buildFluentdLogDNAConfig(instance *operatorv1.CommonAudit) string {
	var result = logDNAConfigV1Data1
	if instance.Spec.Outputs.LogDNA.HostName != "" && instance.Spec.Outputs.LogDNA.ApiKey != "" &&
		instance.Spec.Outputs.LogDNA.IngesterDomain != "" {
		result += yamlLine(2, hostname+instance.Spec.Outputs.LogDNA.HostName, true)
		result += yamlLine(2, ingesterDomain+instance.Spec.Outputs.LogDNA.IngesterDomain, true)
		result += yamlLine(2, apiKey+instance.Spec.Outputs.LogDNA.ApiKey, true)
	} else {
		result += yamlLine(2, hostname+`LOGDNA_SERVER_HOSTNAME`, true)
		result += yamlLine(2, ingesterDomain+`LOGDNA_INGESTER_DOMAIN`, true)
		result += yamlLine(2, apiKey+`LOGDNA_APIKEY`, true)
	}
	result += yamlLine(2, app+instance.Spec.Outputs.LogDNA.App, false)
	return result + logDNAConfigV1Data2
}

// UpdateSIEMConfig returns a String
func UpdateSIEMConfig(instance *operatorv1.CommonAudit, found *corev1.ConfigMap) string {
	var newData, d1, d2, d3, d4 string
	if found.Name == FluentdDaemonSetName+"-"+SplunkConfigName {
		if instance.Spec.Outputs.Splunk != (operatorv1.CommonAuditSpecSplunk{}) {
			newData = found.Data[SplunkConfigKey]
			d1 = RegexHecHost.ReplaceAllString(newData, hecHost+instance.Spec.Outputs.Splunk.Host)
			d2 = RegexHecPort.ReplaceAllString(d1, hecPort+strconv.Itoa(int(instance.Spec.Outputs.Splunk.Port)))
			d3 = RegexHecToken.ReplaceAllString(d2, hecToken+instance.Spec.Outputs.Splunk.Token)
			d4 = RegexProtocol.ReplaceAllString(d3, protocol+Protocols[instance.Spec.Outputs.Splunk.TLS])
		}
	} else if found.Name == FluentdDaemonSetName+"-"+QRadarConfigName {
		if instance.Spec.Outputs.Syslog != (operatorv1.CommonAuditSpecSyslog{}) {
			newData = found.Data[QRadarConfigKey]
			d1 = RegexHost.ReplaceAllString(newData, host+instance.Spec.Outputs.Syslog.Host)
			d2 = RegexPort.ReplaceAllString(d1, port+strconv.Itoa(int(instance.Spec.Outputs.Syslog.Port)))
			d3 = RegexHostname.ReplaceAllString(d2, hostname+instance.Spec.Outputs.Syslog.Hostname)
			d4 = RegexTLS.ReplaceAllString(d3, tls+strconv.FormatBool(instance.Spec.Outputs.Syslog.TLS))
		}
	} else {
		if instance.Spec.Outputs.LogDNA != (operatorv1.CommonAuditSpecLogDNA{}) {
			newData = found.Data[LogDNAConfigKey]
			d1 = RegexHostname.ReplaceAllString(newData, hostname+instance.Spec.Outputs.LogDNA.HostName)
			d2 = RegexApiKey.ReplaceAllString(d1, apiKey+instance.Spec.Outputs.LogDNA.ApiKey)
			d3 = RegexIngesterDomain.ReplaceAllString(d2, ingesterDomain+instance.Spec.Outputs.LogDNA.IngesterDomain)
			d4 = RegexApp.ReplaceAllString(d3, app+instance.Spec.Outputs.LogDNA.App)
		}

	}
	return d4
}

// EqualSIEMConfig returns a Boolean
func EqualSIEMConfig(instance *operatorv1.CommonAudit, found *corev1.ConfigMap) (bool, bool) {
	var key string
	type config struct {
		name  string
		value string
		regex *regexp.Regexp
	}
	var configs = []config{}
	logger := log.WithValues("func", "EqualSIEMConfigs")
	if found.Name == FluentdDaemonSetName+"-"+SplunkConfigName {
		if instance.Spec.Outputs.Splunk != (operatorv1.CommonAuditSpecSplunk{}) {
			key = SplunkConfigKey
			configs = []config{
				{
					name:  "Host",
					value: instance.Spec.Outputs.Splunk.Host,
					regex: RegexHecHost,
				},
				{
					name:  "Port",
					value: strconv.Itoa(int(instance.Spec.Outputs.Splunk.Port)),
					regex: RegexHecPort,
				},
				{
					name:  "Token",
					value: instance.Spec.Outputs.Splunk.Token,
					regex: RegexHecToken,
				},
				{
					name:  "Protocol",
					value: Protocols[instance.Spec.Outputs.Splunk.TLS],
					regex: RegexProtocol,
				},
			}
		}
	} else if found.Name == FluentdDaemonSetName+"-"+QRadarConfigName {
		if instance.Spec.Outputs.Syslog != (operatorv1.CommonAuditSpecSyslog{}) {
			key = QRadarConfigKey
			configs = []config{
				{
					name:  "Host",
					value: instance.Spec.Outputs.Syslog.Host,
					regex: RegexHost,
				},
				{
					name:  "Port",
					value: strconv.Itoa(int(instance.Spec.Outputs.Syslog.Port)),
					regex: RegexPort,
				},
				{
					name:  "Hostname",
					value: instance.Spec.Outputs.Syslog.Hostname,
					regex: RegexHostname,
				},
				{
					name:  "TLS",
					value: strconv.FormatBool(instance.Spec.Outputs.Syslog.TLS),
					regex: RegexTLS,
				},
			}
		}
	} else {
		if instance.Spec.Outputs.LogDNA != (operatorv1.CommonAuditSpecLogDNA{}) {
			key = LogDNAConfigKey
			configs = []config{
				{
					name:  "Hostname",
					value: instance.Spec.Outputs.LogDNA.HostName,
					regex: RegexHostname,
				},
				{
					name:  "Apikey",
					value: instance.Spec.Outputs.LogDNA.ApiKey,
					regex: RegexApiKey,
				},
				{
					name:  "IngesterDomain",
					value: instance.Spec.Outputs.LogDNA.IngesterDomain,
					regex: RegexIngesterDomain,
				},
				{
					name:  "App",
					value: instance.Spec.Outputs.LogDNA.App,
					regex: RegexApp,
				},
			}
		}

	}
	var equal = true
	var missing = false
	for _, c := range configs {
		equal = true
		missing = false
		found := c.regex.FindStringSubmatch(found.Data[key])
		if len(found) > 0 {
			value := strings.Split(found[0], " ")
			if len(value) > 1 {
				if value[1] != c.value {
					equal = false
				}
			} else {
				equal = false
				missing = true
			}
		} else {
			equal = false
			missing = true
		}
		if !equal {
			logger.Info(c.name+" incorrect", "Expected", c.value)
			break
		}
	}
	return equal, missing
}
