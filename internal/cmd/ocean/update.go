package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdUpdate struct {
		cmd  *cobra.Command
		opts CmdUpdateOptions
	}

	CmdUpdateOptions struct {
		*CmdOptions
	}
)

func NewCmdUpdate(opts *CmdOptions) *cobra.Command {
	return newCmdUpdate(opts).cmd
}

func newCmdUpdate(opts *CmdOptions) *CmdUpdate {
	var cmd CmdUpdate

	cmd.cmd = &cobra.Command{
		Use:           "update",
		Short:         "Update an existing resource",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdUpdate) initSubCommands() {
	commands := []func(*CmdUpdateOptions) *cobra.Command{
		NewCmdUpdateCluster,
		NewCmdUpdateLaunchSpec,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdUpdateOptions) Init(flags *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdUpdateOptions) Validate() error {
	return x.CmdOptions.Validate()
}
