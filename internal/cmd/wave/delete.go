package wave

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/wave"
)

type CmdDelete struct {
	cmd  *cobra.Command
	opts CmdDeleteOptions
}

type CmdDeleteOptions struct {
	*CmdOptions
	ClusterID   string
	ClusterName string
}

func (x *CmdDeleteOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "cluster id")
	fs.StringVar(&x.ClusterName, flags.FlagWaveClusterName, x.ClusterName, "cluster name")
}

func NewCmdDelete(opts *CmdOptions) *cobra.Command {
	return newCmdDelete(opts).cmd
}

func newCmdDelete(opts *CmdOptions) *CmdDelete {
	var cmd CmdDelete

	cmd.cmd = &cobra.Command{
		Use:           "delete",
		Short:         "Delete a Wave installation",
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
	if x.ClusterID == "" && x.ClusterName == "" {
		return errors.RequiredOr(flags.FlagWaveClusterID, flags.FlagWaveClusterName)
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
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
	}

	_, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	c, err := oceanClient.GetCluster(ctx, x.opts.ClusterID)
	if err != nil {
		return err
	}

	// TODO Remove option to specify cluster-name on command line, or look up Ocean cluster by name,
	// This will override the user supplied command line flag
	x.opts.ClusterName = c.Name

	manager, err := wave.NewManager(x.opts.ClusterName, getWaveLogger()) // pass in name to validate ocean controller configuration
	if err != nil {
		return err
	}

	err = manager.Delete()
	if err != nil {
		return err
	}
	return nil
}
