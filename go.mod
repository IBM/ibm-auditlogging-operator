module github.com/IBM/ibm-auditlogging-operator

go 1.13

require (
	github.com/IBM/controller-filtered-cache v0.2.1
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/jetstack/cert-manager v0.10.1
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.2
)

replace k8s.io/client-go => k8s.io/client-go v0.18.6

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
