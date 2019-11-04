package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdUpdateLaunchSpec struct {
		cmd  *cobra.Command
		opts CmdUpdateLaunchSpecOptions
	}

	CmdUpdateLaunchSpecOptions struct {
		*CmdUpdateOptions
	}
)

func NewCmdUpdateLaunchSpec(opts *CmdUpdateOptions) *cobra.Command {
	return newCmdUpdateLaunchSpec(opts).cmd
}

func newCmdUpdateLaunchSpec(opts *CmdUpdateOptions) *CmdUpdateLaunchSpec {
	var cmd CmdUpdateLaunchSpec

	cmd.cmd = &cobra.Command{
		Use:           "launchspec",
		Short:         "Update an existing launch spec",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdUpdateLaunchSpec) initSubCommands() {
	commands := []func(*CmdUpdateLaunchSpecOptions) *cobra.Command{
		NewCmdUpdateLaunchSpecKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdUpdateLaunchSpecOptions) Init(fs *pflag.FlagSet, opts *CmdUpdateOptions) {
	x.CmdUpdateOptions = opts
}

func (x *CmdUpdateLaunchSpecOptions) Validate() error {
	return x.CmdUpdateOptions.Validate()
}
