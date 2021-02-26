package wave

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdDescribe struct {
		cmd  *cobra.Command
		opts CmdDescribeOptions
	}

	CmdDescribeOptions struct {
		*CmdOptions
	}
)

func NewCmdDescribe(opts *CmdOptions) *cobra.Command {
	return newCmdDescribe(opts).cmd
}

func newCmdDescribe(opts *CmdOptions) *CmdDescribe {
	var cmd CmdDescribe

	cmd.cmd = &cobra.Command{
		Use:           "describe",
		Short:         "Describe a resource",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdDescribe) initSubCommands() {
	commands := []func(*CmdDescribeOptions) *cobra.Command{
		NewCmdDescribeCluster,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdDescribeOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdDescribeOptions) Validate() error {
	return x.CmdOptions.Validate()
}
