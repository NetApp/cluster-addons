module sigs.k8s.io/cluster-addons/cert-manager

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jetstack/cert-manager v0.14.2
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	golang.org/x/sys v0.0.0-20200113162924-86b910548bc1 // indirect
	k8s.io/apimachinery v0.18.1
	k8s.io/client-go v0.18.1
	sigs.k8s.io/cluster-addons/util v0.0.0-8711d1fd448f10dc2adce23ce02e702d2e1efaa2
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/kubebuilder-declarative-pattern v0.0.0-20200415210853-85eb326a6add
)

replace (
	sigs.k8s.io/cluster-addons/util => ../util
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.1-0.20200414221803-bac7e8aaf90a
	sigs.k8s.io/kubebuilder-declarative-pattern => ../../kubebuilder-declarative-pattern
)
