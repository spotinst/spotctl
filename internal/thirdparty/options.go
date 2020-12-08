package thirdparty

import "io"

type CommandOptions struct {
	// Stdin, Stdout, and Stderr represent the respective data streams that the
	// command may act upon.
	In       io.Reader
	Out, Err io.Writer

	// Path is the path of the command to run. If Path is relative, it is
	// evaluated relative to Dir.
	Path string
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

// WithPath specifies the path of the command to run from.
func WithPath(path string) CommandOption {
	return func(opts *CommandOptions) {
		opts.Path = path
	}
}

func initDefaultOptions() *CommandOptions {
	return &CommandOptions{}
}
