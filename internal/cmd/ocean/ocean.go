package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/cmd/options"
	"github.com/spotinst/spotctl/internal/spot"
)

type (
	Cmd struct {
		cmd  *cobra.Command
		opts CmdOptions
	}

	CmdOptions struct {
		*options.CommonOptions

		// CloudProvider configures the name of the cloud provider associated with
		// the account.
		//
		// Populated by a pre-run function.
		CloudProvider spot.CloudProviderName
	}
)

func NewCmd(opts *options.CommonOptions) *cobra.Command {
	return newCmd(opts).cmd
}

func newCmd(opts *options.CommonOptions) *Cmd {
	var cmd Cmd

	cmd.cmd = &cobra.Command{
		Use:           "ocean",
		Short:         "Manage Ocean resources",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *Cmd) preRun(ctx context.Context) error {
	// Call to the the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}

	// TODO(liran): Use the Spotinst API to figure out the cloud provider
	//  associated with the configured account. We support a single cloud
	//  provider (AWS) it's okay to to hard coded it here for now.
	x.opts.CloudProvider = spot.CloudProviderAWS

	return nil
}

func (x *Cmd) initSubCommands() {
	commands := []func(*CmdOptions) *cobra.Command{
		NewCmdQuickstart,
		NewCmdCreate,
		NewCmdGet,
		NewCmdDescribe,
		NewCmdUpdate,
		NewCmdEdit,
		NewCmdDelete,
		NewCmdRollout,
		NewCmdOperator,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdOptions) Init(fs *pflag.FlagSet, opts *options.CommonOptions) {
	x.CommonOptions = opts
}

func (x *CmdOptions) Validate() error {
	return x.CommonOptions.Validate()
}
