package child

import (
	"io"
)

// CommandOption sets an optional parameter for commands.
type CommandOption func(*Command) error

// WithArgs specifies the command arguments.
func WithArgs(args ...string) CommandOption {
	return func(cmd *Command) error {
		cmd.Args = args
		return nil
	}
}

// WithStdio specifies the standard input, output and error files data streams.
func WithStdio(in io.Reader, out, err io.Writer) CommandOption {
	return func(cmd *Command) error {
		cmd.Stdin = in
		cmd.Stdout = out
		cmd.Stderr = err
		return nil
	}
}

// WithEnv specifies the environment of the process created by the command.
func WithEnv(env []string) CommandOption {
	return func(cmd *Command) error {
		cmd.Env = env
		return nil
	}
}

// WithPath specifies the path of the command to run from.
func WithPath(path string) CommandOption {
	return func(cmd *Command) error {
		cmd.Path = path
		return nil
	}
}

// WithDir specifies the working directory of the command.
func WithDir(dir string) CommandOption {
	return func(cmd *Command) error {
		cmd.Dir = dir
		return nil
	}
}
