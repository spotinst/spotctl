package dep

import (
	"context"
	"errors"
	"net/url"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("dep: not implemented")

type (
	// Manager defines the interface of a Dependency Manager.
	Manager interface {
		// Install installs a new dependency.
		Install(ctx context.Context, dep Dependency, options ...InstallOption) error

		// InstallBulk installs a bulk of new dependencies.
		InstallBulk(ctx context.Context, deps []Dependency, options ...InstallOption) error
	}

	// Dependency represents an executable package.
	Dependency interface {
		// Name returns the name of the dependency.
		Name() string

		// Version returns the version of the dependency.
		Version() string

		// URL returns the download link of the dependency.
		URL() (*url.URL, error)

		// Extension returns the extension type of the dependency.
		Extension() string

		// Executable returns the name of the dependency executable.
		Executable() string
	}
)

// InstallPolicy describes a policy for if/when to install a dependency.
type InstallPolicy string

const (
	// InstallAlways means that the Manager always attempts to install the Dependency.
	InstallAlways InstallPolicy = "Always"
	// InstallNever means that the Manager never installs a Dependency, but only uses a local one.
	InstallNever InstallPolicy = "Never"
	// InstallIfNotPresent means that the Manager installs if the Dependency isn't present.
	InstallIfNotPresent InstallPolicy = "IfNotPresent"
)
