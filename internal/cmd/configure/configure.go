package configure

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/cmd/options"
	"github.com/spotinst/spotinst-cli/internal/errors"
	"github.com/spotinst/spotinst-cli/internal/log"
	"github.com/spotinst/spotinst-cli/internal/spotinst"
	"github.com/spotinst/spotinst-cli/internal/survey"
)

type (
	Cmd struct {
		cmd  *cobra.Command
		opts CmdOptions
	}

	CmdOptions struct {
		*options.CommonOptions

		Token   string
		Account string
	}
)

func NewCmd(opts *options.CommonOptions) *cobra.Command {
	return newCmd(opts).cmd
}

func newCmd(opts *options.CommonOptions) *Cmd {
	var cmd Cmd

	cmd.cmd = &cobra.Command{
		Use:           "configure",
		Short:         "Configure options",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *Cmd) Run(ctx context.Context) error {
	steps := []func(context.Context) error{x.survey, x.validate, x.run}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (x *Cmd) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	log.Debugf("Starting survey...")
	surv, err := x.opts.Clients.NewSurvey()
	if err != nil {
		return err
	}

	// Token.
	{
		if x.opts.Token == "" {
			input := &survey.Input{
				Message:  "Enter your access token",
				Help:     "Your access token acts on your behalf when interacting with the Spotinst API",
				Required: true,
			}

			if x.opts.Token, err = surv.Password(input); err != nil {
				return err
			}
		}
	}

	// Account.
	{
		if x.opts.Account == "" {
			// Instantiate a Spotinst client instance.
			spotinstClientOpts := []spotinst.ClientOption{
				spotinst.WithCredentials(x.opts.Token, ""),
				spotinst.WithDryRun(x.opts.DryRun),
			}

			spotinstClient, err := x.opts.Clients.NewSpotinst(spotinstClientOpts...)
			if err != nil {
				return err
			}

			accounts, err := spotinstClient.Accounts().ListAccounts(ctx)
			if err != nil {
				return err
			}

			accountOpts := make([]interface{}, 0, len(accounts))
			for _, account := range accounts {
				accountOpts = append(accountOpts,
					fmt.Sprintf("%s (%s)", account.ID, account.Name))
			}

			input := &survey.Select{
				Message:   "Select your default account",
				Help:      "The default account in which your resources will be created",
				Options:   accountOpts,
				Transform: survey.TransformOnlyId,
			}

			if x.opts.Account, err = surv.Select(input); err != nil {
				return err
			}
		}
	}

	return nil
}

func (x *Cmd) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *Cmd) run(ctx context.Context) error {
	return errors.NotImplemented()
}

func (x *CmdOptions) Init(flags *pflag.FlagSet, opts *options.CommonOptions) {
	x.CommonOptions = opts
}

func (x *CmdOptions) Validate() error {
	if err := x.CommonOptions.Validate(); err != nil {
		return err
	}

	if x.Token == "" {
		return errors.Required("token")
	}

	if x.Account != "" && !strings.HasPrefix(x.Account, "act") {
		return errors.Invalid("account", x.Account)
	}

	return nil
}
