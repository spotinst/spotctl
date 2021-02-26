package wave

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
	CmdDescribeCluster struct {
		cmd  *cobra.Command
		opts CmdDescribeClusterOptions
	}

	CmdDescribeClusterOptions struct {
		*CmdDescribeOptions
		ClusterID string
	}
)

func NewCmdDescribeCluster(opts *CmdDescribeOptions) *cobra.Command {
	return newCmdDescribeCluster(opts).cmd
}

func newCmdDescribeCluster(opts *CmdDescribeOptions) *CmdDescribeCluster {
	var cmd CmdDescribeCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Describe a Wave cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdDescribeClusterOptions) Init(fs *pflag.FlagSet, opts *CmdDescribeOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdDescribeClusterOptions) initDefaults(opts *CmdDescribeOptions) {
	x.CmdDescribeOptions = opts
}

func (x *CmdDescribeClusterOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdDescribeCluster) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdDescribeClusterOptions) Validate() error {
	if x.ClusterID == "" {
		return errors.Required(flags.FlagWaveClusterID)
	}
	return x.CmdDescribeOptions.Validate()
}

func (x *CmdDescribeCluster) Run(ctx context.Context) error {
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

func (x *CmdDescribeCluster) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDescribeCluster) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDescribeCluster) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	waveClient, err := spotClient.Services().Wave()
	if err != nil {
		return err
	}

	cluster, err := waveClient.GetCluster(ctx, x.opts.ClusterID)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(json.WriterFormat)
	if err != nil {
		return err
	}

	return w.Write(cluster.Obj)
}
