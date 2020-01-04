module github.com/bythepowerof/kmake-controller

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/microsoft/azure-databricks-operator v0.0.0-20191219224929-593794c9ecf5
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.0
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
