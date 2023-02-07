package thirdparty

import (
	"context"
	"errors"
	"io"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("thirdparty: not implemented")

type (
	// CommandName represents the name of a third-party command.
	CommandName string

	// Command defines the interface that should be implemented by a third-party command.
	Command interface {
		// Name returns the third-party command name.
		Name() CommandName

		// Run invokes the command with optional arguments. An error is
		// returned if the command fails, nil otherwise.
		Run(ctx context.Context, args ...string) error

		// RunWithStdin invokes the command with stdin override and optional arguments. An error is
		// returned if the command fails, nil otherwise.
		RunWithStdin(ctx context.Context, stdin io.Reader, args ...string) error
	}

	// Factory is a function that returns a command interface. An error is
	// returned if the command fails to initialize, nil otherwise.
	Factory func(options *CommandOptions) (Command, error)
)
