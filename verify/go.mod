module sigs.k8s.io/kubebuilder-release-tools/verify

go 1.20

replace sigs.k8s.io/kubebuilder-release-tools/notes => ../notes

require (
	github.com/google/go-github/v32 v32.1.0
	golang.org/x/oauth2 v0.8.0
	sigs.k8s.io/kubebuilder-release-tools/notes v0.3.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
