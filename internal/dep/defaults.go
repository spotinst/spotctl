package dep

import "path/filepath"

var (
	// See: https://kubernetes.io/docs/reference/kubectl.
	DependencyKubectl = Dependency{
		Name:    "kubectl",
		Version: "1.15.3",
		URL: "https://storage.googleapis.com/kubernetes-release/" +
			"release/v{{.version}}/bin/{{.os}}/{{.arch}}/kubectl{{.extension}}",
	}

	// See: https://github.com/kubernetes/kops.
	DependencyKops = Dependency{
		Name:    "kops",
		Version: "1.13.0",
		URL: "https://github.com/kubernetes/kops/" +
			"releases/download/{{.version}}/kops-{{.os}}-{{.arch}}",
	}
)

// DefaultDependencyListKubernetes returns default list of required dependencies
// to work with Kubernetes.
func DefaultDependencyListKubernetes() []Dependency {
	return []Dependency{
		DependencyKubectl,
		DependencyKops,
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
