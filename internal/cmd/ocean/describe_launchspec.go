package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDescribeLaunchSpec struct {
		cmd  *cobra.Command
		opts CmdDescribeLaunchSpecOptions
	}

	CmdDescribeLaunchSpecOptions struct {
		*CmdDescribeOptions
	}
)

func NewCmdDescribeLaunchSpec(opts *CmdDescribeOptions) *cobra.Command {
	return newCmdDescribeLaunchSpec(opts).cmd
}

func newCmdDescribeLaunchSpec(opts *CmdDescribeOptions) *CmdDescribeLaunchSpec {
	var cmd CmdDescribeLaunchSpec

	cmd.cmd = &cobra.Command{
		Use:           "launchspec",
		Short:         "Describe a launch spec",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDescribeLaunchSpec) initSubCommands() {
	commands := []func(*CmdDescribeLaunchSpecOptions) *cobra.Command{
		NewCmdDescribeLaunchSpecKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDescribeLaunchSpecOptions) Init(fs *pflag.FlagSet, opts *CmdDescribeOptions) {
	x.CmdDescribeOptions = opts
}

func (x *CmdDescribeLaunchSpecOptions) Validate() error {
	return x.CmdDescribeOptions.Validate()
}
