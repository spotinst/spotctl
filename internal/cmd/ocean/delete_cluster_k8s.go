package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDeleteClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdDeleteClusterKubernetesOptions
	}

	CmdDeleteClusterKubernetesOptions struct {
		*CmdDeleteClusterOptions
	}
)

func NewCmdDeleteClusterKubernetes(opts *CmdDeleteClusterOptions) *cobra.Command {
	return newCmdDeleteClusterKubernetes(opts).cmd
}

func newCmdDeleteClusterKubernetes(opts *CmdDeleteClusterOptions) *CmdDeleteClusterKubernetes {
	var cmd CmdDeleteClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Delete an existing Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDeleteClusterKubernetes) initSubCommands() {
	commands := []func(*CmdDeleteClusterKubernetesOptions) *cobra.Command{
		NewCmdDeleteClusterKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDeleteClusterKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdDeleteClusterOptions) {
	x.CmdDeleteClusterOptions = opts
}

func (x *CmdDeleteClusterKubernetesOptions) Validate() error {
	return x.CmdDeleteClusterOptions.Validate()
}
