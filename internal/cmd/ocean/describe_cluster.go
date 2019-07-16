package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDescribeCluster struct {
		cmd  *cobra.Command
		opts CmdDescribeClusterOptions
	}

	CmdDescribeClusterOptions struct {
		*CmdDescribeOptions
	}
)

func NewCmdDescribeCluster(opts *CmdDescribeOptions) *cobra.Command {
	return newCmdDescribeCluster(opts).cmd
}

func newCmdDescribeCluster(opts *CmdDescribeOptions) *CmdDescribeCluster {
	var cmd CmdDescribeCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Describe a cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDescribeCluster) initSubCommands() {
	commands := []func(*CmdDescribeClusterOptions) *cobra.Command{
		NewCmdDescribeClusterKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDescribeClusterOptions) Init(flags *pflag.FlagSet, opts *CmdDescribeOptions) {
	x.CmdDescribeOptions = opts
}

func (x *CmdDescribeClusterOptions) Validate() error {
	return x.CmdDescribeOptions.Validate()
}
