package dep

import "path/filepath"

var (
	// See: https://kubernetes.io/docs/reference/kubectl.
	DependencyKubectl Dependency = &dependency{
		name:    "kubectl",
		version: "1.19.0",
		url: "https://storage.googleapis.com/kubernetes-release/" +
			"release/v{{.version}}/bin/{{.os}}/{{.arch}}/kubectl{{.extension}}",
	}

	// See: https://github.com/kubernetes/kops.
	DependencyKops Dependency = &dependency{
		name:    "kops",
		version: "1.18.2",
		url: "https://github.com/kubernetes/kops/releases/download/" +
			"v{{.version}}/kops-{{.os}}-{{.arch}}",
	}

	// See: https://github.com/weaveworks/eksctl.
	DependencyEksctl Dependency = &dependency{
		name:    "eksctl",
		version: "0.36.2-f5c273f8",
		url: "https://github.com/spotinst/weaveworks-eksctl/releases/download" +
			"/v{{.version}}/eksctl_{{.os}}_{{.arch}}.tar.gz",
	}
)

// DefaultDependencyListKubernetes returns the default list of packages needed
// to work with Kubernetes-based products, such as Ocean and Wave.
func DefaultDependencyListKubernetes() []Dependency {
	return []Dependency{
		DependencyKubectl,
		DependencyKops,
		DependencyEksctl,
	}
}

// DefaultBinaryDir returns default binary directory path.
//
// Builds the binary directory path based on the OS's platform.
//   - Linux/Unix: $HOME/.spotinst/bin
//   - Windows: %USERPROFILE%\.spotinst\bin
func DefaultBinaryDir() string {
	return filepath.Join(userHomeDir(), ".spotinst", "bin")
}
