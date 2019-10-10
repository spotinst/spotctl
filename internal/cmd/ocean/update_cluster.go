package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdUpdateCluster struct {
		cmd  *cobra.Command
		opts CmdUpdateClusterOptions
	}

	CmdUpdateClusterOptions struct {
		*CmdUpdateOptions
	}
)

func NewCmdUpdateCluster(opts *CmdUpdateOptions) *cobra.Command {
	return newCmdUpdateCluster(opts).cmd
}

func newCmdUpdateCluster(opts *CmdUpdateOptions) *CmdUpdateCluster {
	var cmd CmdUpdateCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Update an existing cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdUpdateCluster) initSubCommands() {
	commands := []func(*CmdUpdateClusterOptions) *cobra.Command{
		NewCmdUpdateClusterKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdUpdateClusterOptions) Init(flags *pflag.FlagSet, opts *CmdUpdateOptions) {
	x.CmdUpdateOptions = opts
}

func (x *CmdUpdateClusterOptions) Validate() error {
	return x.CmdUpdateOptions.Validate()
}
