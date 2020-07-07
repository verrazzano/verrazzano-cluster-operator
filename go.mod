module github.com/verrazzano/verrazzano-cluster-operator

go 1.13

require (
	github.com/Jeffail/gabs/v2 v2.2.0
	github.com/kylelemons/godebug v1.1.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/rs/zerolog v1.19.0
	github.com/spf13/pflag v1.0.5
	github.com/verrazzano/verrazzano-crd-generator v0.3.32
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.2
