package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdSparkDelete struct {
		cmd  *cobra.Command
		opts CmdSparkDeleteOptions
	}

	CmdSparkDeleteOptions struct {
		*CmdSparkOptions
	}
)

func NewCmdSparkDelete(opts *CmdSparkOptions) *cobra.Command {
	return newCmdSparkDelete(opts).cmd
}

func newCmdSparkDelete(opts *CmdSparkOptions) *CmdSparkDelete {
	var cmd CmdSparkDelete

	cmd.cmd = &cobra.Command{
		Use:           "delete",
		Short:         "Delete a resource",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdSparkDelete) preRun(ctx context.Context) error {
	// Call to the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}
	return nil
}

func (x *CmdSparkDelete) initSubCommands() {
	commands := []func(*CmdSparkDeleteOptions) *cobra.Command{
		NewCmdSparkDeleteCluster,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdSparkDeleteOptions) Init(fs *pflag.FlagSet, opts *CmdSparkOptions) {
	x.CmdSparkOptions = opts
}

func (x *CmdSparkDeleteOptions) Validate() error {
	return x.CmdSparkOptions.Validate()
}
