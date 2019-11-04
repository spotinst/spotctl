package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdEditCluster struct {
		cmd  *cobra.Command
		opts CmdEditClusterOptions
	}

	CmdEditClusterOptions struct {
		*CmdEditOptions
	}
)

func NewCmdEditCluster(opts *CmdEditOptions) *cobra.Command {
	return newCmdEditCluster(opts).cmd
}

func newCmdEditCluster(opts *CmdEditOptions) *CmdEditCluster {
	var cmd CmdEditCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Edit an existing cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdEditCluster) initSubCommands() {
	commands := []func(*CmdEditClusterOptions) *cobra.Command{
		NewCmdEditClusterKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdEditClusterOptions) Init(fs *pflag.FlagSet, opts *CmdEditOptions) {
	x.CmdEditOptions = opts
}

func (x *CmdEditClusterOptions) Validate() error {
	return x.CmdEditOptions.Validate()
}
