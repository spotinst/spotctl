package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdCreateClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdCreateClusterKubernetesOptions
	}

	CmdCreateClusterKubernetesOptions struct {
		*CmdCreateClusterOptions
	}
)

func NewCmdCreateClusterKubernetes(opts *CmdCreateClusterOptions) *cobra.Command {
	return newCmdCreateClusterKubernetes(opts).cmd
}

func newCmdCreateClusterKubernetes(opts *CmdCreateClusterOptions) *CmdCreateClusterKubernetes {
	var cmd CmdCreateClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Create a new Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdCreateClusterKubernetes) initSubCommands() {
	commands := []func(*CmdCreateClusterKubernetesOptions) *cobra.Command{
		NewCmdCreateClusterKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdCreateClusterKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdCreateClusterOptions) {
	x.CmdCreateClusterOptions = opts
}

func (x *CmdCreateClusterKubernetesOptions) Validate() error {
	return x.CmdCreateClusterOptions.Validate()
}
