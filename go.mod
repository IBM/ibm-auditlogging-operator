module github.com/IBM/ibm-auditlogging-operator

go 1.13

require (
	github.com/IBM/controller-filtered-cache v0.2.1
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/gobuffalo/flect v0.2.0 // indirect
	github.com/jetstack/cert-manager v0.10.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.20.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.25.4
	k8s.io/apimachinery v0.25.4
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.11.1
	golang.org/x/crypto => golang.org/x/crypto v0.7.0
	golang.org/x/net => golang.org/x/net v0.8.0
	golang.org/x/sys => golang.org/x/sys v0.6.0
	golang.org/x/text => golang.org/x/text v0.8.0
	k8s.io/api => k8s.io/api v0.25.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.25.4
	k8s.io/client-go => k8s.io/client-go v0.25.4
)
