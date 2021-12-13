package ocean

import (
	"context"

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
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdSparkCreate) preRun(ctx context.Context) error {
	// Call to the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}
	return nil
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
