package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdGet struct {
		cmd  *cobra.Command
		opts CmdGetOptions
	}

	CmdGetOptions struct {
		*CmdOptions

		// Get options
		NoHeaders bool
		Output    string
	}
)

func NewCmdGet(opts *CmdOptions) *cobra.Command {
	return newCmdGet(opts).cmd
}

func newCmdGet(opts *CmdOptions) *CmdGet {
	var cmd CmdGet

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

func (x *CmdGet) initSubCommands() {
	commands := []func(*CmdGetOptions) *cobra.Command{
		NewCmdGetCluster,
		NewCmdGetLaunchSpec,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdGetOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.initFlags(fs)
	x.initDefaults(opts)
}

func (x *CmdGetOptions) initDefaults(opts *CmdOptions) {
	x.CmdOptions = opts
	x.NoHeaders = false
	x.Output = "table"
}

func (x *CmdGetOptions) initFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&x.NoHeaders, "no-headers", x.NoHeaders, "when using the `table` output format, don't print headers")
	fs.StringVarP(&x.Output, "output", "o", x.Output, "output format (table|json|yaml)")
}

func (x *CmdGetOptions) Validate() error {
	// TODO(liran): Validate output format.

	return x.CmdOptions.Validate()
}
