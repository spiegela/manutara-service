module github.com/manutara/service

go 1.12

require (
	github.com/fvbock/endless v0.0.0-20170109170031-447134032cb6
	github.com/graphql-go/graphql v0.7.8
	github.com/graphql-go/handler v0.2.3
	github.com/manutara/manutara v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	sigs.k8s.io/controller-runtime v0.2.2
)

replace github.com/manutara/manutara => /Users/spiegela/src/proj/manutara/manutara
