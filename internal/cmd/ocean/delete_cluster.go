package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDeleteCluster struct {
		cmd  *cobra.Command
		opts CmdDeleteClusterOptions
	}

	CmdDeleteClusterOptions struct {
		*CmdDeleteOptions
	}
)

func NewCmdDeleteCluster(opts *CmdDeleteOptions) *cobra.Command {
	return newCmdDeleteCluster(opts).cmd
}

func newCmdDeleteCluster(opts *CmdDeleteOptions) *CmdDeleteCluster {
	var cmd CmdDeleteCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Delete an existing cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDeleteCluster) initSubCommands() {
	commands := []func(*CmdDeleteClusterOptions) *cobra.Command{
		NewCmdDeleteClusterKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDeleteClusterOptions) Init(flags *pflag.FlagSet, opts *CmdDeleteOptions) {
	x.CmdDeleteOptions = opts
}

func (x *CmdDeleteClusterOptions) Validate() error {
	return x.CmdDeleteOptions.Validate()
}
