package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/writer/writers/json"
)

type (
	CmdSparkDescribeCluster struct {
		cmd  *cobra.Command
		opts CmdSparkDescribeClusterOptions
	}

	CmdSparkDescribeClusterOptions struct {
		*CmdSparkDescribeOptions

		ClusterID string
	}
)

func NewCmdSparkDescribeCluster(opts *CmdSparkDescribeOptions) *cobra.Command {
	return newCmdSparkDescribeCluster(opts).cmd
}

func newCmdSparkDescribeCluster(opts *CmdSparkDescribeOptions) *CmdSparkDescribeCluster {
	var cmd CmdSparkDescribeCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Describe an Ocean Spark cluster",
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

func (x *CmdSparkDescribeCluster) Run(ctx context.Context) error {
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

func (x *CmdSparkDescribeCluster) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdSparkDescribeCluster) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdSparkDescribeCluster) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdSparkDescribeCluster) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
		spot.WithDryRun(x.opts.DryRun),
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	oceanSparkClient, err := spotClient.Services().OceanSpark()
	if err != nil {
		return err
	}

	cluster, err := oceanSparkClient.GetCluster(ctx, x.opts.ClusterID)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(json.WriterFormat)
	if err != nil {
		return err
	}

	return w.Write(cluster.Obj)
}

func (x *CmdSparkDescribeClusterOptions) Init(fs *pflag.FlagSet, opts *CmdSparkDescribeOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdSparkDescribeClusterOptions) initDefaults(opts *CmdSparkDescribeOptions) {
	x.CmdSparkDescribeOptions = opts
}

func (x *CmdSparkDescribeClusterOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOFASClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdSparkDescribeClusterOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdSparkDescribeOptions.Validate(); err != nil {
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
