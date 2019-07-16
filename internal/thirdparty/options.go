package thirdparty

import "io"

type CommandOptions struct {
	// Stdin, Stdout, and Stderr represent the respective data streams that the
	// command may act upon.
	In       io.Reader
	Out, Err io.Writer
}

// CommandOption allows specifying various settings configurable by a command.
type CommandOption func(*CommandOptions)

// WithStdio specifies the standard input, output and error files data streams.
func WithStdio(in io.Reader, out, err io.Writer) CommandOption {
	return func(opts *CommandOptions) {
		opts.In = in
		opts.Out = out
		opts.Err = err
	}
}

func initDefaultOptions() *CommandOptions {
	return &CommandOptions{}
}
