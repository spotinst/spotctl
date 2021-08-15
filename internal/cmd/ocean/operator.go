package ocean

import (
	"github.com/spf13/cobra"
)

type (
	CmdOperator struct {
		cmd  *cobra.Command
		opts CmdOperatorOptions
	}

	CmdOperatorOptions struct {
		*CmdOptions
	}
)

func NewCmdOperator(opts *CmdOptions) *cobra.Command {
	return newCmdOperator(opts).cmd
}

func newCmdOperator(opts *CmdOptions) *CmdOperator {
	var cmd CmdOperator

	cmd.cmd = &cobra.Command{
		Use:           "operator",
		Short:         "Manage the Ocean Operator",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.CmdOptions = opts
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdOperator) initSubCommands() {
	commands := []func(*CmdOperatorOptions) *cobra.Command{
		NewCmdOperatorInstall,
		NewCmdOperatorUninstall,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}
