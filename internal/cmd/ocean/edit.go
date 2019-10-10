package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdEdit struct {
		cmd  *cobra.Command
		opts CmdEditOptions
	}

	CmdEditOptions struct {
		*CmdOptions
	}
)

func NewCmdEdit(opts *CmdOptions) *cobra.Command {
	return newCmdEdit(opts).cmd
}

func newCmdEdit(opts *CmdOptions) *CmdEdit {
	var cmd CmdEdit

	cmd.cmd = &cobra.Command{
		Use:           "edit",
		Short:         "Edit an existing resource",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdEdit) initSubCommands() {
	commands := []func(*CmdEditOptions) *cobra.Command{
		NewCmdEditCluster,
		NewCmdEditLaunchSpec,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdEditOptions) Init(flags *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdEditOptions) Validate() error {
	return x.CmdOptions.Validate()
}
