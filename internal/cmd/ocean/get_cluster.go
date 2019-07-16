package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdGetCluster struct {
		cmd  *cobra.Command
		opts CmdGetClusterOptions
	}

	CmdGetClusterOptions struct {
		*CmdGetOptions
	}
)

func NewCmdGetCluster(opts *CmdGetOptions) *cobra.Command {
	return newCmdGetCluster(opts).cmd
}

func newCmdGetCluster(opts *CmdGetOptions) *CmdGetCluster {
	var cmd CmdGetCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Display one or many clusters",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdGetCluster) initSubCommands() {
	commands := []func(*CmdGetClusterOptions) *cobra.Command{
		NewCmdGetClusterKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdGetClusterOptions) Init(flags *pflag.FlagSet, opts *CmdGetOptions) {
	x.CmdGetOptions = opts
}

func (x *CmdGetClusterOptions) Validate() error {
	return x.CmdGetOptions.Validate()
}
