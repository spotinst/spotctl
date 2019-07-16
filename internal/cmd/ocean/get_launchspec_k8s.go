package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdGetLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdGetLaunchSpecKubernetesOptions
	}

	CmdGetLaunchSpecKubernetesOptions struct {
		*CmdGetLaunchSpecOptions
	}
)

func NewCmdGetLaunchSpecKubernetes(opts *CmdGetLaunchSpecOptions) *cobra.Command {
	return newCmdGetLaunchSpecKubernetes(opts).cmd
}

func newCmdGetLaunchSpecKubernetes(opts *CmdGetLaunchSpecOptions) *CmdGetLaunchSpecKubernetes {
	var cmd CmdGetLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Display one or many Kubernetes launch specs",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdGetLaunchSpecKubernetes) initSubCommands() {
	commands := []func(*CmdGetLaunchSpecKubernetesOptions) *cobra.Command{
		NewCmdGetLaunchSpecKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdGetLaunchSpecKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdGetLaunchSpecOptions) {
	x.CmdGetLaunchSpecOptions = opts
}

func (x *CmdGetLaunchSpecKubernetesOptions) Validate() error {
	return x.CmdGetLaunchSpecOptions.Validate()
}
