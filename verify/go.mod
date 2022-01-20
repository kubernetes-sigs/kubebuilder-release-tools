module sigs.k8s.io/kubebuilder-release-tools/verify

go 1.17

replace sigs.k8s.io/kubebuilder-release-tools/notes => ../notes

require (
	github.com/google/go-github/v32 v32.1.0
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	sigs.k8s.io/kubebuilder-release-tools/notes v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073 // indirect
	golang.org/x/net v0.0.0-20200520004742-59133d7f0dd7 // indirect
	google.golang.org/appengine v1.1.0 // indirect
	google.golang.org/protobuf v1.23.0 // indirect
)
