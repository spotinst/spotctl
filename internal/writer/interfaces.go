package writer

import (
	"errors"
	"io"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("writer: not implemented")

type (
	// Format represents the format of a writer.
	Format string

	// Writer defines the interface that should be implemented by a writer.
	Writer interface {
		Write(obj interface{}) error
	}

	// Factory is a function that returns a writer interface. An error is
	// returned if the writer fails to initialize, nil otherwise.
	Factory func(w io.Writer) (Writer, error)
)
