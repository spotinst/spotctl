package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdSparkGet struct {
		cmd  *cobra.Command
		opts CmdSparkGetOptions
	}

	CmdSparkGetOptions struct {
		*CmdSparkOptions

		// Get options
		NoHeaders bool
		Output    string
	}
)

func NewCmdSparkGet(opts *CmdSparkOptions) *cobra.Command {
	return newCmdSparkGet(opts).cmd
}

func newCmdSparkGet(opts *CmdSparkOptions) *CmdSparkGet {
	var cmd CmdSparkGet

	cmd.cmd = &cobra.Command{
		Use:           "get",
		Short:         "Display one or many resources",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdSparkGet) initSubCommands() {
	commands := []func(*CmdSparkGetOptions) *cobra.Command{
		NewCmdSparkGetCluster,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdSparkGetOptions) Init(fs *pflag.FlagSet, opts *CmdSparkOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdSparkGetOptions) initDefaults(opts *CmdSparkOptions) {
	x.CmdSparkOptions = opts
	x.NoHeaders = false
	x.Output = "table"
}

func (x *CmdSparkGetOptions) initFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&x.NoHeaders, "no-headers", x.NoHeaders, "when using the `table` output format, don't print headers")
	fs.StringVarP(&x.Output, "output", "o", x.Output, "output format (table|json|yaml)")
}

func (x *CmdSparkGetOptions) Validate() error {
	return x.CmdOptions.Validate()
}
