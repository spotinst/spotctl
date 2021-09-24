package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdSpark struct {
		cmd  *cobra.Command
		opts CmdSparkOptions
	}

	CmdSparkOptions struct {
		*CmdOptions
	}
)

func NewCmdSpark(opts *CmdOptions) *cobra.Command {
	return newCmdSpark(opts).cmd
}

func newCmdSpark(opts *CmdOptions) *CmdSpark {
	var cmd CmdSpark

	cmd.cmd = &cobra.Command{
		Use:           "spark",
		Short:         "Manage Ocean for Apache Spark resources",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdSpark) initSubCommands() {
	commands := []func(*CmdSparkOptions) *cobra.Command{
		NewCmdSparkCreate,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdSparkOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdSparkOptions) Validate() error {
	return x.CmdOptions.Validate()
}
