package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDescribeClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdDescribeClusterKubernetesOptions
	}

	CmdDescribeClusterKubernetesOptions struct {
		*CmdDescribeClusterOptions
	}
)

func NewCmdDescribeClusterKubernetes(opts *CmdDescribeClusterOptions) *cobra.Command {
	return newCmdDescribeClusterKubernetes(opts).cmd
}

func newCmdDescribeClusterKubernetes(opts *CmdDescribeClusterOptions) *CmdDescribeClusterKubernetes {
	var cmd CmdDescribeClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Describe a Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDescribeClusterKubernetes) initSubCommands() {
	commands := []func(*CmdDescribeClusterKubernetesOptions) *cobra.Command{
		NewCmdDescribeClusterKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDescribeClusterKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdDescribeClusterOptions) {
	x.CmdDescribeClusterOptions = opts
}

func (x *CmdDescribeClusterKubernetesOptions) Validate() error {
	return x.CmdDescribeClusterOptions.Validate()
}
