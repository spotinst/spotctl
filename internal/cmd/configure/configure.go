package configure

import (
	"context"
	"fmt"

	"github.com/go-ini/ini"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/cmd/options"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/spotinst"
	"github.com/spotinst/spotctl/internal/survey"
	"github.com/spotinst/spotctl/internal/thirdparty/commands/aws"

	credsaws "github.com/aws/aws-sdk-go/aws/credentials"
	credsspot "github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
)

type (
	Cmd struct {
		cmd  *cobra.Command
		opts CmdOptions
	}

	CmdOptions struct {
		*options.CommonOptions

		CredentialsSpotinst
	}

	CredentialsSpotinst struct {
		credsspot.Value
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
	steps := []func(context.Context) error{
		x.configureCredentialsSpotinst,
		x.configureCredentialsAWS,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (x *Cmd) configureCredentialsSpotinst(ctx context.Context) error {
	log.Debugf("Configuring Spotinst credentials")

	// Survey.
	{
		if x.opts.Noninteractive {
			return errors.Noninteractive(x.cmd.Name())
		}

		log.Debugf("Starting survey...")
		surv, err := x.opts.Clients.NewSurvey()
		if err != nil {
			return err
		}

		// Token.
		{
			if x.opts.CredentialsSpotinst.Token == "" {
				input := &survey.Input{
					Message:  "Enter your access token",
					Help:     "Your access token acts on your behalf when interacting with the Spotinst API",
					Required: true,
				}

				if x.opts.CredentialsSpotinst.Token, err = surv.Password(input); err != nil {
					return err
				}
			}
		}

		// Account.
		{
			if x.opts.CredentialsSpotinst.Account == "" {
				// Instantiate a Spotinst client instance.
				spotinstClientOpts := []spotinst.ClientOption{
					spotinst.WithCredentialsStatic(x.opts.CredentialsSpotinst.Token, ""),
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

				if x.opts.CredentialsSpotinst.Account, err = surv.Select(input); err != nil {
					return err
				}
			}
		}
	}

	// Configure.
	{
		// Configuration filename.
		filename := credsspot.DefaultFilename()

		// Create or update configuration.
		cfg, err := ini.LooseLoad(filename)
		if err != nil {
			return err
		}

		// Create a new `default` section.
		sec, err := cfg.NewSection(x.opts.Profile)
		if err != nil {
			return err
		}

		// Create a new `token` key.
		if _, err := sec.NewKey("token", x.opts.Token); err != nil {
			return err
		}

		// Create a new `account` key.
		if x.opts.Account != "" {
			if _, err := sec.NewKey("account", x.opts.Account); err != nil {
				return err
			}
		}

		// Write out configuration to stdout.
		if x.opts.DryRun {
			_, err := cfg.WriteTo(x.opts.Out)
			return err
		}

		// Write out configuration to a file.
		if err := cfg.SaveTo(filename); err != nil {
			return err
		}
	}

	log.Debugf("Configured Spotinst credentials")
	return nil
}

func (x *Cmd) configureCredentialsAWS(ctx context.Context) error {
	log.Debugf("Configuring AWS credentials")

	// Survey.
	{
		if x.opts.Noninteractive {
			return errors.Noninteractive(x.cmd.Name())
		}

		provider := credsaws.NewChainCredentials(
			[]credsaws.Provider{
				&credsaws.EnvProvider{},
				&credsaws.SharedCredentialsProvider{Profile: x.opts.Profile},
			})

		if _, err := provider.Get(); err == nil {
			log.Debugf("Skipping AWS credential configuration because credentials are already configured")
			return nil
		}
	}

	// Configure.
	{
		log.Debugf("Starting survey...")
		surv, err := x.opts.Clients.NewSurvey()
		if err != nil {
			return err
		}

		// Confirm.
		{
			input := &survey.Input{
				Message: "Configure AWS credentials",
			}

			configure, err := surv.Confirm(input)
			if err != nil {
				return err
			}

			if !configure {
				log.Debugf("Skipping AWS credential configuration because user selection")
				return nil
			}
		}

		// TODO(liran): Use the dependency manager to install the AWS CLI, if needed.
		cmd, err := x.opts.Clients.NewCommand(aws.CommandName)
		if err != nil {
			return err
		}

		if err := cmd.Run(ctx, "configure", "--profile", x.opts.Profile); err != nil {
			return err
		}
	}

	log.Debugf("Configured AWS credentials")
	return nil
}

func (x *CmdOptions) Init(fs *pflag.FlagSet, opts *options.CommonOptions) {
	x.initDefaults(opts)
}

func (x *CmdOptions) initDefaults(opts *options.CommonOptions) {
	x.CommonOptions = opts
}
