package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdSparkDescribe struct {
		cmd  *cobra.Command
		opts CmdSparkDescribeOptions
	}

	CmdSparkDescribeOptions struct {
		*CmdSparkOptions
	}
)

func NewCmdSparkDescribe(opts *CmdSparkOptions) *cobra.Command {
	return newCmdSparkDescribe(opts).cmd
}

func newCmdSparkDescribe(opts *CmdSparkOptions) *CmdSparkDescribe {
	var cmd CmdSparkDescribe

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

func (x *CmdSparkDescribe) initSubCommands() {
	commands := []func(*CmdSparkDescribeOptions) *cobra.Command{
		NewCmdSparkDescribeCluster,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdSparkDescribeOptions) Init(fs *pflag.FlagSet, opts *CmdSparkOptions) {
	x.CmdSparkOptions = opts
}

func (x *CmdSparkDescribeOptions) Validate() error {
	return x.CmdOptions.Validate()
}
