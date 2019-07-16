package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDeleteLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdDeleteLaunchSpecKubernetesOptions
	}

	CmdDeleteLaunchSpecKubernetesOptions struct {
		*CmdDeleteLaunchSpecOptions
	}
)

func NewCmdDeleteLaunchSpecKubernetes(opts *CmdDeleteLaunchSpecOptions) *cobra.Command {
	return newCmdDeleteLaunchSpecKubernetes(opts).cmd
}

func newCmdDeleteLaunchSpecKubernetes(opts *CmdDeleteLaunchSpecOptions) *CmdDeleteLaunchSpecKubernetes {
	var cmd CmdDeleteLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Delete an existing Kubernetes launchspec",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDeleteLaunchSpecKubernetes) initSubCommands() {
	commands := []func(*CmdDeleteLaunchSpecKubernetesOptions) *cobra.Command{
		NewCmdDeleteLaunchSpecKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDeleteLaunchSpecKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdDeleteLaunchSpecOptions) {
	x.CmdDeleteLaunchSpecOptions = opts
}

func (x *CmdDeleteLaunchSpecKubernetesOptions) Validate() error {
	return x.CmdDeleteLaunchSpecOptions.Validate()
}
