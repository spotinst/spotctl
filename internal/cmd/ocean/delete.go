package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDelete struct {
		cmd  *cobra.Command
		opts CmdDeleteOptions
	}

	CmdDeleteOptions struct {
		*CmdOptions
	}
)

func NewCmdDelete(opts *CmdOptions) *cobra.Command {
	return newCmdDelete(opts).cmd
}

func newCmdDelete(opts *CmdOptions) *CmdDelete {
	var cmd CmdDelete

	cmd.cmd = &cobra.Command{
		Use:           "delete",
		Short:         "Delete an existing resource",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDelete) initSubCommands() {
	commands := []func(*CmdDeleteOptions) *cobra.Command{
		NewCmdDeleteCluster,
		NewCmdDeleteLaunchSpec,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDeleteOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdDeleteOptions) Validate() error {
	return x.CmdOptions.Validate()
}
