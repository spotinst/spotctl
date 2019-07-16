package version

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/cmd/options"
	"github.com/spotinst/spotinst-cli/internal/version"
)

type (
	Cmd struct {
		cmd  *cobra.Command
		opts CmdOptions
	}

	CmdOptions struct {
		*options.CommonOptions
	}
)

func NewCmd(opts *options.CommonOptions) *cobra.Command {
	return newCmd(opts).cmd
}

func newCmd(opts *options.CommonOptions) *Cmd {
	var cmd Cmd

	cmd.cmd = &cobra.Command{
		Use:           "version",
		Short:         "Print version information",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *Cmd) Run(ctx context.Context) error {
	_, err := fmt.Fprintln(x.opts.Out, version.String())
	return err
}

func (x *CmdOptions) Init(flags *pflag.FlagSet, opts *options.CommonOptions) {
	x.CommonOptions = opts
}

func (x *CmdOptions) Validate() error {
	return x.CommonOptions.Validate()
}
