module sigs.k8s.io/kubebuilder-release-tools/verify

go 1.15

replace sigs.k8s.io/kubebuilder-release-tools/notes => ../notes

require (
	github.com/google/go-github/v32 v32.1.0
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	sigs.k8s.io/kubebuilder-release-tools/notes v0.0.0-00010101000000-000000000000
)
