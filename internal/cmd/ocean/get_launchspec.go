package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdGetLaunchSpec struct {
		cmd  *cobra.Command
		opts CmdGetLaunchSpecOptions
	}

	CmdGetLaunchSpecOptions struct {
		*CmdGetOptions
	}
)

func NewCmdGetLaunchSpec(opts *CmdGetOptions) *cobra.Command {
	return newCmdGetLaunchSpec(opts).cmd
}

func newCmdGetLaunchSpec(opts *CmdGetOptions) *CmdGetLaunchSpec {
	var cmd CmdGetLaunchSpec

	cmd.cmd = &cobra.Command{
		Use:           "launchspec",
		Short:         "Display one or many launchspecs",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdGetLaunchSpec) initSubCommands() {
	commands := []func(*CmdGetLaunchSpecOptions) *cobra.Command{
		NewCmdGetLaunchSpecKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdGetLaunchSpecOptions) Init(flags *pflag.FlagSet, opts *CmdGetOptions) {
	x.CmdGetOptions = opts
}

func (x *CmdGetLaunchSpecOptions) Validate() error {
	return x.CmdGetOptions.Validate()
}
