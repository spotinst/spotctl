package version

import (
	_ "embed"
	"strings"

	"github.com/hashicorp/go-version"
)

var (
	// Version is an instance of version.Version used to verify that the full
	// version complies with the Semantic Versioning specification (https://semver.org).
	//
	// Populated at runtime.
	// Read-only.
	Version *version.Version

	// _version represents the full version that must comply with the Semantic
	// Versioning specification (https://semver.org).
	//
	// Populated at build-time.
	// Read-only.
	//go:embed VERSION
	_version string
)

func init() {
	// Parse and verify the given version.
	Version = version.Must(version.NewSemver(strings.TrimSpace(_version)))
}

// Prerelease is an alias of version.Prerelease.
func Prerelease() string {
	return Version.Prerelease()
}

// Metadata is an alias of version.Metadata.
func Metadata() string {
	return Version.Metadata()
}

// String is an alias of version.String.
func String() string {
	return Version.String()
}
