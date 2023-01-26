package eksctl

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/spotinst/spotctl/internal/child"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/thirdparty"
)

// CommandName is the name of this command.
const CommandName thirdparty.CommandName = "eksctl-spot"

func init() {
	thirdparty.Register(CommandName, factory)
}

func factory(options *thirdparty.CommandOptions) (thirdparty.Command, error) {
	return &Command{options}, nil
}

type Command struct {
	opts *thirdparty.CommandOptions
}

func (x *Command) Name() thirdparty.CommandName {
	return CommandName
}

func (x *Command) Run(ctx context.Context, args ...string) error {
	return x.RunWithStdin(ctx, nil, args...)
}

func (x *Command) RunWithStdin(ctx context.Context, stdin io.Reader, args ...string) error {
	log.Debugf("Executing command: %s %s", CommandName, strings.Join(args, " "))

	steps := []func(ctx context.Context, stdin io.Reader, args ...string) error{
		x.runVersion,
		x.run,
	}

	for _, step := range steps {
		if err := step(ctx, stdin, args...); err != nil {
			return err
		}
	}

	return nil
}

func (x *Command) runVersion(ctx context.Context, _ io.Reader, _ ...string) error {
	var buf bytes.Buffer

	cmdOptions := []child.CommandOption{
		child.WithArgs("version", "--output", "json"),
		child.WithStdio(nil, &buf, nil),
		child.WithPath(x.opts.Path),
	}

	if err := child.NewCommand(string(CommandName), cmdOptions...).Run(ctx); err != nil {
		return err
	}

	log.Debugf("%s", buf.String())
	return nil
}

func (x *Command) run(ctx context.Context, stdin io.Reader, args ...string) error {
	stdinReader := x.opts.In
	if stdin != nil {
		stdinReader = stdin
	}

	cmdOptions := []child.CommandOption{
		child.WithArgs(args...),
		child.WithStdio(stdinReader, x.opts.Out, x.opts.Err),
		child.WithPath(x.opts.Path),
	}

	return child.NewCommand(string(CommandName), cmdOptions...).Run(ctx)
}
