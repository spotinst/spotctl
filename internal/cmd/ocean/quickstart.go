package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/dep"
)

type (
	CmdQuickstart struct {
		cmd  *cobra.Command
		opts CmdQuickstartOptions
	}

	CmdQuickstartOptions struct {
		*CmdOptions

		// Quickstart options
		Advanced bool
	}
)

func NewCmdQuickstart(opts *CmdOptions) *cobra.Command {
	return newCmdQuickstart(opts).cmd
}

func newCmdQuickstart(opts *CmdOptions) *CmdQuickstart {
	var cmd CmdQuickstart

	cmd.cmd = &cobra.Command{
		Use:           "quickstart",
		Short:         "Create a quickstart environment",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdQuickstart) initSubCommands() {
	commands := []func(*CmdQuickstartOptions) *cobra.Command{
		NewCmdQuickstartCluster,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdQuickstart) preRun(ctx context.Context) error {
	// Call to the the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}

	return x.installDeps(ctx)
}

func (x *CmdQuickstart) installDeps(ctx context.Context) error {
	// Initialize a new dependency manager.
	dm, err := x.opts.Clients.NewDep()
	if err != nil {
		return err
	}

	// Install options.
	installOpts := []dep.InstallOption{
		dep.WithNoninteractive(x.opts.Noninteractive),
		dep.WithDryRun(x.opts.DryRun),
	}

	// Install!
	return dm.InstallBulk(ctx, dep.DefaultDependencyListKubernetes(), installOpts...)
}

func (x *CmdQuickstartOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdQuickstartOptions) Validate() error {
	return x.CmdOptions.Validate()
}
