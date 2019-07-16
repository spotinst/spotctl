package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/cmd/options"
	"github.com/spotinst/spotinst-cli/internal/dep"
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
		Use:           "ocean",
		Short:         "Manage Cmd resources",
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

func (x *Cmd) initSubCommands() {
	commands := []func(*CmdOptions) *cobra.Command{
		NewCmdCreate,
		NewCmdDelete,
		NewCmdGet,
		NewCmdDescribe,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *Cmd) preRun(ctx context.Context) error {
	// Call to the the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}

	return x.installDeps(ctx)
}

func (x *Cmd) installDeps(ctx context.Context) error {
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

func (x *CmdOptions) Init(flags *pflag.FlagSet, opts *options.CommonOptions) {
	x.CommonOptions = opts
}

func (x *CmdOptions) Validate() error {
	return x.CommonOptions.Validate()
}
