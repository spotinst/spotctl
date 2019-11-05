package kops

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spotinst/spotctl/internal/child"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/thirdparty"
)

// CommandName is the name of this command.
const CommandName thirdparty.CommandName = "kops"

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
	log.Debugf("Executing command: kops %s", strings.Join(args, " "))

	fns := []func(ctx context.Context, args ...string) error{
		x.runVersion,
		x.run,
	}

	for _, fn := range fns {
		if err := fn(ctx, args...); err != nil {
			return err
		}
	}

	return nil
}

func (x *Command) runVersion(ctx context.Context, args ...string) error {
	var buf bytes.Buffer

	cmdOptions := []child.CommandOption{
		child.WithArgs("version"),
		child.WithStdio(nil, &buf, nil),
	}

	if err := child.NewCommand("kops", cmdOptions...).Run(ctx); err != nil {
		return err
	}

	log.Debugf("%s", buf.String())
	return nil
}

func (x *Command) run(ctx context.Context, args ...string) error {
	const (
		featureFlagKey = "KOPS_FEATURE_FLAGS"
		featureFlagVal = "Spotinst,SpotinstOcean"
	)

	env := os.Environ()
	envFeatured := false

	for _, kv := range env {
		if strings.Contains(kv, featureFlagKey) && strings.Contains(kv, featureFlagVal) {
			envFeatured = true
		}
	}
	if !envFeatured {
		env = append(env, fmt.Sprintf(`%s="+%s"`,
			featureFlagKey,
			featureFlagVal))
	}

	cmdOptions := []child.CommandOption{
		child.WithArgs(args...),
		child.WithStdio(x.opts.In, x.opts.Out, x.opts.Err),
		child.WithEnv(env),
	}

	return child.NewCommand("kops", cmdOptions...).Run(ctx)
}
