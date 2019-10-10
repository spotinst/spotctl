package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdQuickstartCluster struct {
		cmd  *cobra.Command
		opts CmdQuickstartClusterOptions
	}

	CmdQuickstartClusterOptions struct {
		*CmdQuickstartOptions
	}
)

func NewCmdQuickstartCluster(opts *CmdQuickstartOptions) *cobra.Command {
	return newCmdQuickstartCluster(opts).cmd
}

func newCmdQuickstartCluster(opts *CmdQuickstartOptions) *CmdQuickstartCluster {
	var cmd CmdQuickstartCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Create a quickstart Ocean cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdQuickstartCluster) initSubCommands() {
	commands := []func(*CmdQuickstartClusterOptions) *cobra.Command{
		NewCmdQuickstartClusterKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdQuickstartClusterOptions) Init(flags *pflag.FlagSet, opts *CmdQuickstartOptions) {
	x.CmdQuickstartOptions = opts
}

func (x *CmdQuickstartClusterOptions) Validate() error {
	return x.CmdQuickstartOptions.Validate()
}
