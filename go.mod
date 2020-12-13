module github.com/verrazzano/verrazzano-cluster-operator

go 1.13

require (
	github.com/Jeffail/gabs/v2 v2.2.0
	github.com/kylelemons/godebug v1.1.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/stretchr/testify v1.5.1
	github.com/verrazzano/verrazzano-crd-generator v0.3.35-0.20201211145006-a6882074bd8d
	go.uber.org/zap v1.16.0
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace k8s.io/client-go => k8s.io/client-go v0.18.2
