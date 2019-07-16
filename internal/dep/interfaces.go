package dep

import (
	"context"
	"errors"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("deps: not implemented")

type (
	// Interface defines the interface of a Dependency Manager.
	Interface interface {
		// Install installs a new dependency.
		Install(ctx context.Context, dep Dependency, options ...InstallOption) error

		// InstallBulk installs a bulk of new dependencies.
		InstallBulk(ctx context.Context, deps []Dependency, options ...InstallOption) error
	}

	// Dependency represents an executable package.
	Dependency struct {
		Name    string
		Version string
		URL     string
	}
)
