package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDeleteLaunchSpec struct {
		cmd  *cobra.Command
		opts CmdDeleteLaunchSpecOptions
	}

	CmdDeleteLaunchSpecOptions struct {
		*CmdDeleteOptions
	}
)

func NewCmdDeleteLaunchSpec(opts *CmdDeleteOptions) *cobra.Command {
	return newCmdDeleteLaunchSpec(opts).cmd
}

func newCmdDeleteLaunchSpec(opts *CmdDeleteOptions) *CmdDeleteLaunchSpec {
	var cmd CmdDeleteLaunchSpec

	cmd.cmd = &cobra.Command{
		Use:           "launchspec",
		Short:         "Delete an existing launch spec",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDeleteLaunchSpec) initSubCommands() {
	commands := []func(*CmdDeleteLaunchSpecOptions) *cobra.Command{
		NewCmdDeleteLaunchSpecKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDeleteLaunchSpecOptions) Init(flags *pflag.FlagSet, opts *CmdDeleteOptions) {
	x.CmdDeleteOptions = opts
}

func (x *CmdDeleteLaunchSpecOptions) Validate() error {
	return x.CmdDeleteOptions.Validate()
}
