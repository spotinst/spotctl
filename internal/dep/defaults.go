package dep

import "path/filepath"

var (
	// See: https://kubernetes.io/docs/reference/kubectl.
	DependencyKubectl Dependency = &dependency{
		name:    "kubectl",
		version: "1.19.6",
		url: "https://storage.googleapis.com/kubernetes-release/" +
			"release/v{{.version}}/bin/{{.os}}/{{.arch}}/kubectl{{.extension}}",
	}

	// See: https://github.com/kubernetes/kops.
	DependencyKops Dependency = &dependency{
		name:    "kops",
		version: "1.19.1",
		url: "https://github.com/kubernetes/kops/releases/download/" +
			"v{{.version}}/kops-{{.os}}-{{.arch}}",
	}

	// Spot fork of eksctl
	// See: https://github.com/spotinst/weaveworks-eksctl.
	DependencyEksctlSpot Dependency = &dependency{
		name:               "eksctl-spot",
		upstreamBinaryName: "eksctl",
		version:            "0.108.0-cfe0db21",
		url: "https://github.com/spotinst/weaveworks-eksctl/releases/download" +
			"/v{{.version}}/eksctl_{{.os}}_{{.arch}}.tar.gz",
		// TODO Remove this override when the binary has been released for ARM (M1 macs)
		rosettaArchOverride: true,
	}
)

// DefaultDependencyListKubernetes returns the default list of packages needed
// to work with Kubernetes-based products, such as Ocean.
func DefaultDependencyListKubernetes() []Dependency {
	return []Dependency{
		DependencyKubectl,
		DependencyKops,
		DependencyEksctlSpot,
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
