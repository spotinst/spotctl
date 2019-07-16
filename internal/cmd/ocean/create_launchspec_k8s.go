package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdCreateLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdCreateLaunchSpecKubernetesOptions
	}

	CmdCreateLaunchSpecKubernetesOptions struct {
		*CmdCreateLaunchSpecOptions
	}
)

func NewCmdCreateLaunchSpecKubernetes(opts *CmdCreateLaunchSpecOptions) *cobra.Command {
	return newCmdCreateLaunchSpecKubernetes(opts).cmd
}

func newCmdCreateLaunchSpecKubernetes(opts *CmdCreateLaunchSpecOptions) *CmdCreateLaunchSpecKubernetes {
	var cmd CmdCreateLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Create a new Kubernetes launchspec",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdCreateLaunchSpecKubernetes) initSubCommands() {
	commands := []func(*CmdCreateLaunchSpecKubernetesOptions) *cobra.Command{
		NewCmdCreateLaunchSpecKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdCreateLaunchSpecKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdCreateLaunchSpecOptions) {
	x.CmdCreateLaunchSpecOptions = opts
}

func (x *CmdCreateLaunchSpecKubernetesOptions) Validate() error {
	return x.CmdCreateLaunchSpecOptions.Validate()
}
