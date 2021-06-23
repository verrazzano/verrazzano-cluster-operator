module github.com/verrazzano/verrazzano-cluster-operator

go 1.13

require (
	github.com/Jeffail/gabs/v2 v2.2.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/stretchr/testify v1.6.1
	github.com/verrazzano/pkg v0.0.2
	github.com/verrazzano/verrazzano-crd-generator v0.0.0-20201214161122-0330d094db41
	go.uber.org/zap v1.16.0
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	k8s.io/api => k8s.io/api v0.18.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.2
	k8s.io/client-go => k8s.io/client-go v0.18.2
)
