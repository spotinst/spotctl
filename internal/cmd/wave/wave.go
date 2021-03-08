package wave

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/cmd/options"
	"github.com/spotinst/spotctl/internal/spot"
)

type Cmd struct {
	cmd  *cobra.Command
	opts CmdOptions
}

type CmdOptions struct {
	*options.CommonOptions
	CloudProvider spot.CloudProviderName
}

func NewCmd(opts *options.CommonOptions) *cobra.Command {
	return newCmd(opts).cmd
}

func newCmd(opts *options.CommonOptions) *Cmd {
	var cmd Cmd

	cmd.cmd = &cobra.Command{
		Use:           "wave",
		Short:         "Manage Wave resources",
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

	// ... yeah, no
	x.opts.CloudProvider = spot.CloudProviderAWS

	return nil
}

func (x *Cmd) initSubCommands() {
	commands := []func(*CmdOptions) *cobra.Command{
		NewCmdCreate,
		NewCmdGet,
		NewCmdDescribe,
		NewCmdDelete,
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
