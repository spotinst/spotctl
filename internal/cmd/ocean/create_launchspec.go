package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdCreateLaunchSpec struct {
		cmd  *cobra.Command
		opts CmdCreateLaunchSpecOptions
	}

	CmdCreateLaunchSpecOptions struct {
		*CmdCreateOptions
	}
)

func NewCmdCreateLaunchSpec(opts *CmdCreateOptions) *cobra.Command {
	return newCmdCreateLaunchSpec(opts).cmd
}

func newCmdCreateLaunchSpec(opts *CmdCreateOptions) *CmdCreateLaunchSpec {
	var cmd CmdCreateLaunchSpec

	cmd.cmd = &cobra.Command{
		Use:           "launchspec",
		Short:         "Create a new launchspec",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdCreateLaunchSpec) initSubCommands() {
	commands := []func(*CmdCreateLaunchSpecOptions) *cobra.Command{
		NewCmdCreateLaunchSpecKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdCreateLaunchSpecOptions) Init(flags *pflag.FlagSet, opts *CmdCreateOptions) {
	x.CmdCreateOptions = opts
}

func (x *CmdCreateLaunchSpecOptions) Validate() error {
	return x.CmdCreateOptions.Validate()
}
