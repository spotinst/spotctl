package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdGetClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdGetClusterKubernetesOptions
	}

	CmdGetClusterKubernetesOptions struct {
		*CmdGetClusterOptions
	}
)

func NewCmdGetClusterKubernetes(opts *CmdGetClusterOptions) *cobra.Command {
	return newCmdGetClusterKubernetes(opts).cmd
}

func newCmdGetClusterKubernetes(opts *CmdGetClusterOptions) *CmdGetClusterKubernetes {
	var cmd CmdGetClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Display one or many Kubernetes clusters",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdGetClusterKubernetes) initSubCommands() {
	commands := []func(*CmdGetClusterKubernetesOptions) *cobra.Command{
		NewCmdGetClusterKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdGetClusterKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdGetClusterOptions) {
	x.CmdGetClusterOptions = opts
}

func (x *CmdGetClusterKubernetesOptions) Validate() error {
	return x.CmdGetClusterOptions.Validate()
}
