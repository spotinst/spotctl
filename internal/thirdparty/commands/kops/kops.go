package kops

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

	cmdOptions := []child.CommandOption{
		child.WithArgs("version"),
		child.WithStdio(nil, &buf, nil),
		child.WithPath(x.opts.Path),
	}

	if err := child.NewCommand(string(CommandName), cmdOptions...).Run(ctx); err != nil {
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
		env = append(env, fmt.Sprintf(`%s=+%s`,
			featureFlagKey,
			featureFlagVal))
	}

	for _, kv := range env {
		if strings.HasPrefix(kv, "KOPS") {
			log.Debugf("ENV: %s", kv)
		}
	}

	cmdOptions := []child.CommandOption{
		child.WithArgs(args...),
		child.WithStdio(x.opts.In, x.opts.Out, x.opts.Err),
		child.WithEnv(env),
		child.WithPath(x.opts.Path),
	}

	return child.NewCommand(string(CommandName), cmdOptions...).Run(ctx)
}
