package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/spot"
)

type (
	CmdSparkDeleteCluster struct {
		cmd  *cobra.Command
		opts CmdSparkDeleteClusterOptions
	}

	CmdSparkDeleteClusterOptions struct {
		*CmdSparkDeleteOptions

		ClusterID string
	}
)

func NewCmdSparkDeleteCluster(opts *CmdSparkDeleteOptions) *cobra.Command {
	return newCmdSparkDeleteCluster(opts).cmd
}

func newCmdSparkDeleteCluster(opts *CmdSparkDeleteOptions) *CmdSparkDeleteCluster {
	var cmd CmdSparkDeleteCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Delete an Ocean Spark cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"cl"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdSparkDeleteCluster) Run(ctx context.Context) error {
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

func (x *CmdSparkDeleteCluster) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdSparkDeleteCluster) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdSparkDeleteCluster) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdSparkDeleteCluster) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
		spot.WithDryRun(x.opts.DryRun),
	}

	// TODO Confirmation prompt

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	oceanSparkClient, err := spotClient.Services().OceanSpark()
	if err != nil {
		return err
	}

	log.Infof("Deleting Ocean Spark cluster %s", x.opts.ClusterID)
	if err := oceanSparkClient.DeleteCluster(ctx, x.opts.ClusterID); err != nil {
		return err
	}

	log.Infof("Ocean Spark cluster %s successfully deleted", x.opts.ClusterID)

	return nil
}

func (x *CmdSparkDeleteClusterOptions) Init(fs *pflag.FlagSet, opts *CmdSparkDeleteOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdSparkDeleteClusterOptions) initDefaults(opts *CmdSparkDeleteOptions) {
	x.CmdSparkDeleteOptions = opts
}

func (x *CmdSparkDeleteClusterOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOFASClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdSparkDeleteClusterOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdSparkDeleteOptions.Validate(); err != nil {
		errg.Add(err)
	}

	if x.ClusterID == "" {
		errg.Add(errors.Required("--cluster-id"))
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}
