package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdSparkCreate struct {
		cmd  *cobra.Command
		opts CmdSparkCreateOptions
	}

	CmdSparkCreateOptions struct {
		*CmdSparkOptions
	}
)

func NewCmdSparkCreate(opts *CmdSparkOptions) *cobra.Command {
	return newCmdSparkCreate(opts).cmd
}

func newCmdSparkCreate(opts *CmdSparkOptions) *CmdSparkCreate {
	var cmd CmdSparkCreate

	cmd.cmd = &cobra.Command{
		Use:           "create",
		Short:         "Create a new resource",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdSparkCreate) initSubCommands() {
	commands := []func(*CmdSparkCreateOptions) *cobra.Command{
		NewCmdSparkCreateCluster,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdSparkCreateOptions) Init(fs *pflag.FlagSet, opts *CmdSparkOptions) {
	x.CmdSparkOptions = opts
}

func (x *CmdSparkCreateOptions) Validate() error {
	return x.CmdSparkOptions.Validate()
}
