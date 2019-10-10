package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdEditLaunchSpec struct {
		cmd  *cobra.Command
		opts CmdEditLaunchSpecOptions
	}

	CmdEditLaunchSpecOptions struct {
		*CmdEditOptions
	}
)

func NewCmdEditLaunchSpec(opts *CmdEditOptions) *cobra.Command {
	return newCmdEditLaunchSpec(opts).cmd
}

func newCmdEditLaunchSpec(opts *CmdEditOptions) *CmdEditLaunchSpec {
	var cmd CmdEditLaunchSpec

	cmd.cmd = &cobra.Command{
		Use:           "launchspec",
		Short:         "Edit an existing launch spec",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdEditLaunchSpec) initSubCommands() {
	commands := []func(*CmdEditLaunchSpecOptions) *cobra.Command{
		NewCmdEditLaunchSpecKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdEditLaunchSpecOptions) Init(flags *pflag.FlagSet, opts *CmdEditOptions) {
	x.CmdEditOptions = opts
}

func (x *CmdEditLaunchSpecOptions) Validate() error {
	return x.CmdEditOptions.Validate()
}
