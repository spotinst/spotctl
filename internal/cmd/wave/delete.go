package wave

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
)

type CmdDelete struct {
	cmd  *cobra.Command
	opts CmdDeleteOptions
}

type CmdDeleteOptions struct {
	*CmdOptions
	ClusterID   string
	DeleteOcean bool
}

func (x *CmdDeleteOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "cluster id")
	fs.BoolVar(&x.DeleteOcean, flags.FlagWaveDeleteOceanCluster, x.DeleteOcean, "delete ocean cluster")
}

func NewCmdDelete(opts *CmdOptions) *cobra.Command {
	return newCmdDelete(opts).cmd
}

func newCmdDelete(opts *CmdOptions) *CmdDelete {
	var cmd CmdDelete

	cmd.cmd = &cobra.Command{
		Use:           "delete",
		Short:         "Delete a Wave cluster",
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
		return errors.Required(flags.FlagWaveClusterID)
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

	waveClient, err := spotClient.Services().Wave()
	if err != nil {
		return err
	}

	return waveClient.DeleteCluster(ctx, x.opts.ClusterID, x.opts.DeleteOcean)

	/*oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
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

	// TODO Delete ocean cluster if it was provisioned
	// TODO Delete kubernetes cluster if it was provisioned

	if err := wave.ValidateClusterContext(c.Name); err != nil {
		return fmt.Errorf("cluster context validation failure, %w", err)
	}

	logger := getWaveLogger()

	manager, err := tide.NewManager(logger)
	if err != nil {
		return err
	}

	logger.Info("uninstalling wave")

	err = manager.Delete()
	if err != nil {
		return fmt.Errorf("could not delete wave, %w", err)
	}

	// Since we are running from CLI, we can do a full uninstall and remove the CRD too
	err = manager.DeleteConfiguration(true)
	if err != nil {
		return fmt.Errorf("could not delete wave configuration, %w", err)
	}

	err = manager.DeleteTideRBAC()
	if err != nil {
		return fmt.Errorf("could not delete tide rbac objects, %w", err)
	}

	logger.Info("wave has been uninstalled")

	return nil*/
}
