package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDescribeLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdDescribeLaunchSpecKubernetesOptions
	}

	CmdDescribeLaunchSpecKubernetesOptions struct {
		*CmdDescribeLaunchSpecOptions
	}
)

func NewCmdDescribeLaunchSpecKubernetes(opts *CmdDescribeLaunchSpecOptions) *cobra.Command {
	return newCmdDescribeLaunchSpecKubernetes(opts).cmd
}

func newCmdDescribeLaunchSpecKubernetes(opts *CmdDescribeLaunchSpecOptions) *CmdDescribeLaunchSpecKubernetes {
	var cmd CmdDescribeLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Describe a Kubernetes launch spec",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDescribeLaunchSpecKubernetes) initSubCommands() {
	commands := []func(*CmdDescribeLaunchSpecKubernetesOptions) *cobra.Command{
		NewCmdDescribeLaunchSpecKubernetesAWS,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDescribeLaunchSpecKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdDescribeLaunchSpecOptions) {
	x.CmdDescribeLaunchSpecOptions = opts
}

func (x *CmdDescribeLaunchSpecKubernetesOptions) Validate() error {
	return x.CmdDescribeLaunchSpecOptions.Validate()
}
