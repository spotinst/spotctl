package child

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// Command represents an external command being prepared or run.
type Command struct {
	// Command is the name of the command to execute. Args holds the command
	// arguments, if any.
	Name string
	Args []string

	// Stdin, Stdout, and Stderr represent the respective data streams that the
	// command may act upon. They are attached directly to the child process.
	Stdin          io.Reader
	Stdout, Stderr io.Writer

	// Env specifies the environment of the process. Only these environment
	// variables will be given to the command, so it is the responsibility of
	// the caller to include the parent processes environment, if required.
	// Each entry should be of the form "key=value".
	Env []string

	// Path is the path of the command to run. If Path is relative, it is
	// evaluated relative to Dir.
	Path string

	// Dir specifies the working directory of the command. If Dir is the
	// empty string, Run runs the command in the calling process's current
	// directory.
	Dir string

	// for internal use only
	options    []CommandOption
	cmd        *exec.Cmd
	started    time.Time
	ended      time.Time
	exitStatus int
}

// NewCommand creates a new Command.
func NewCommand(cmd string, options ...CommandOption) *Command {
	return &Command{
		Name:    cmd,
		options: options,
	}
}

// Start starts the specified command but does not wait for it to complete.
func (x *Command) Start(ctx context.Context) error {
	x.started = time.Now()
	var err error

	// Build up the command options.
	for _, opt := range x.options {
		if err = opt(x); err != nil {
			return err
		}
	}

	// Path of the command to run.
	name := x.Name
	if x.Path != "" {
		name = filepath.Join(x.Path, name)
	}

	// Initialize the internal command object.
	x.cmd = exec.CommandContext(ctx, name, x.Args...)

	// Set up the command options.
	if x.Stdin != nil {
		x.cmd.Stdin = x.Stdin
	}
	if x.Stdout != nil {
		x.cmd.Stdout = x.Stdout
	}
	if x.Stderr != nil {
		x.cmd.Stderr = x.Stderr
	}
	if x.Env != nil {
		x.cmd.Env = x.Env
	}
	if x.Dir != "" {
		x.cmd.Dir = x.Dir
	}

	// Finally, start the command.
	return x.cmd.Start()
}

// Wait waits for the command to exit and waits for any copying to
// stdin or copying from stdout or stderr to complete.
func (x *Command) Wait(ctx context.Context) error {
	errorsCh := make(chan error, 1)
	var err error

	go func() {
		err := x.cmd.Wait()
		x.ended = time.Now()
		errorsCh <- err
	}()

	select {
	case <-ctx.Done():
		if x.cmd.Process != nil {
			x.cmd.Process.Kill()
		}
		err = ctx.Err()
	case err = <-errorsCh:
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok { // exit status != 0
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				x.exitStatus = status.ExitStatus()
			}
		}
	}

	return err
}

// CombinedOutput runs the command and returns its combined standard output and
//
//	standard error.
func (x *Command) CombinedOutput(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer

	if x.Stdout == nil {
		x.cmd.Stdout = &buf
	}

	if x.Stderr == nil {
		x.cmd.Stderr = &buf
	}

	return buf.Bytes(), x.Run(ctx)
}

// Run calls Start(), then Wait(), and returns an error (if any). The error may
// be of many types including *exec.ExitError and context.Canceled,
// context.DeadlineExceeded.
func (x *Command) Run(ctx context.Context) error {
	if err := x.Start(ctx); err != nil {
		return err
	}

	return x.Wait(ctx)
}

// Runtime returns the amount of time the process is or was running.
func (x *Command) Runtime() time.Duration {
	return x.ended.Sub(x.started)
}

// Pid yields the pid of the process (dead or alive), or 0 if the process has
// not been run yet.
func (x *Command) Pid() uint32 {
	var pid uint32

	if x.cmd.Process != nil {
		pid = uint32(x.cmd.Process.Pid)
	}

	return pid
}

// ExitStatus returns the exit status of the process.
func (x *Command) ExitStatus() int {
	return x.exitStatus
}
