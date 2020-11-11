package wave

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spotinst"
)

type CmdDelete struct {
	cmd  *cobra.Command
	opts CmdDeleteOptions
}

type CmdDeleteOptions struct {
	*CmdOptions
	ClusterID string
}

func (x *CmdDeleteOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
}

func NewCmdDelete(opts *CmdOptions) *cobra.Command {
	return newCmdDelete(opts).cmd
}

func newCmdDelete(opts *CmdOptions) *CmdDelete {
	var cmd CmdDelete

	cmd.cmd = &cobra.Command{
		Use:           "delete",
		Short:         "Delete a wave installation",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdDeleteOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
	x.initFlags(fs)
}

func (x *CmdDelete) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdDeleteOptions) Validate() error {
	if x.ClusterID == "" {
		return fmt.Errorf("--cluster-id must be specified")
	}
	return x.CmdOptions.Validate()
}

func (x *CmdDelete) Run(ctx context.Context) error {
	steps := []func(context.Context) error{
		x.survey,
		x.log,
		x.validate,
		x.run,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (x *CmdDelete) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDelete) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDelete) run(ctx context.Context) error {
	spotinstClientOpts := []spotinst.ClientOption{
		spotinst.WithCredentialsProfile(x.opts.Profile),
	}

	_, err := x.opts.Clientset.NewSpotinst(spotinstClientOpts...)
	if err != nil {
		return err
	}

	fmt.Fprintln(x.opts.Out, fmt.Sprintf("blah blah blah"))
	return nil
}
