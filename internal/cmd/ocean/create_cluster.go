package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdCreateCluster struct {
		cmd  *cobra.Command
		opts CmdCreateClusterOptions
	}

	CmdCreateClusterOptions struct {
		*CmdCreateOptions
	}
)

func NewCmdCreateCluster(opts *CmdCreateOptions) *cobra.Command {
	return newCmdCreateCluster(opts).cmd
}

func newCmdCreateCluster(opts *CmdCreateOptions) *CmdCreateCluster {
	var cmd CmdCreateCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Create a new cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdCreateCluster) initSubCommands() {
	commands := []func(*CmdCreateClusterOptions) *cobra.Command{
		NewCmdCreateClusterKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdCreateClusterOptions) Init(fs *pflag.FlagSet, opts *CmdCreateOptions) {
	x.CmdCreateOptions = opts
}

func (x *CmdCreateClusterOptions) Validate() error {
	return x.CmdCreateOptions.Validate()
}
