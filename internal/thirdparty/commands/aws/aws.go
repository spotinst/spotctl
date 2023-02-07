package aws

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
const CommandName thirdparty.CommandName = "aws"

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
	log.Debugf("Executing command: %s %s", CommandName, strings.Join(args, " "))

	steps := []func(ctx context.Context, args ...string) error{
		x.runVersion,
		x.run,
	}

	for _, step := range steps {
		if err := step(ctx, args...); err != nil {
			return err
		}
	}

	return nil
}

func (x *Command) RunWithStdin(_ context.Context, _ io.Reader, _ ...string) error {
	return thirdparty.ErrNotImplemented
}

func (x *Command) runVersion(ctx context.Context, args ...string) error {
	var buf bytes.Buffer

	// TODO(liran/dep-awscli): Configure the local execution path.
	cmdOptions := []child.CommandOption{
		child.WithArgs("--version"),
		child.WithStdio(nil, &buf, nil),
	}

	if err := child.NewCommand(string(CommandName), cmdOptions...).Run(ctx); err != nil {
		return err
	}

	log.Debugf("%s", buf.String())
	return nil
}

func (x *Command) run(ctx context.Context, args ...string) error {
	// TODO(liran/dep-awscli): Configure the local execution path.
	cmdOptions := []child.CommandOption{
		child.WithArgs(args...),
		child.WithStdio(x.opts.In, x.opts.Out, x.opts.Err),
	}

	return child.NewCommand(string(CommandName), cmdOptions...).Run(ctx)
}
