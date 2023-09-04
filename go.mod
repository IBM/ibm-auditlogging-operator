module github.com/IBM/ibm-auditlogging-operator

go 1.13

require (
	github.com/IBM/controller-filtered-cache v0.3.6
	github.com/go-logr/logr v1.2.4
	//github.com/jetstack/cert-manager v0.10.1
	github.com/ibm/ibm-cert-manager-operator v0.0.0-20220602233809-3a62073266c7
	// github.com/onsi/ginkgo v1.12.1
	// github.com/onsi/gomega v1.10.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.27.10
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.28.1
	k8s.io/apimachinery v0.28.1
	k8s.io/client-go v0.28.1
	//k8s.io/apimachinery v0.18.6
	//k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0
	//sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-runtime v0.16.1
)

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.11.1
	golang.org/x/crypto => golang.org/x/crypto v0.7.0
	golang.org/x/net => golang.org/x/net v0.8.0
	golang.org/x/sys => golang.org/x/sys v0.6.0
	golang.org/x/text => golang.org/x/text v0.8.0
//k8s.io/client-go => k8s.io/client-go v0.18.6
)
