package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdCreate struct {
		cmd  *cobra.Command
		opts CmdCreateOptions
	}

	CmdCreateOptions struct {
		*CmdOptions
	}
)

func NewCmdCreate(opts *CmdOptions) *cobra.Command {
	return newCmdCreate(opts).cmd
}

func newCmdCreate(opts *CmdOptions) *CmdCreate {
	var cmd CmdCreate

	cmd.cmd = &cobra.Command{
		Use:           "create",
		Short:         "Create a new resource",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdCreate) initSubCommands() {
	commands := []func(*CmdCreateOptions) *cobra.Command{
		NewCmdCreateCluster,
		NewCmdCreateLaunchSpec,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdCreateOptions) Init(flags *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdCreateOptions) Validate() error {
	return x.CmdOptions.Validate()
}
