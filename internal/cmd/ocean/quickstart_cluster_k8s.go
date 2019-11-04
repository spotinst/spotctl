package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdQuickstartClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdQuickstartClusterKubernetesOptions
	}

	CmdQuickstartClusterKubernetesOptions struct {
		*CmdQuickstartClusterOptions
	}
)

func NewCmdQuickstartClusterKubernetes(opts *CmdQuickstartClusterOptions) *cobra.Command {
	return newCmdQuickstartClusterKubernetes(opts).cmd
}

func newCmdQuickstartClusterKubernetes(opts *CmdQuickstartClusterOptions) *CmdQuickstartClusterKubernetes {
	var cmd CmdQuickstartClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Create a quickstart Ocean cluster (Kubernetes)",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdQuickstartClusterKubernetes) initSubCommands() {
	commands := []func(*CmdQuickstartClusterKubernetesOptions) *cobra.Command{
		NewCmdQuickstartClusterKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdQuickstartClusterKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdQuickstartClusterOptions) {
	x.CmdQuickstartClusterOptions = opts
}

func (x *CmdQuickstartClusterKubernetesOptions) Validate() error {
	return x.CmdQuickstartClusterOptions.Validate()
}
