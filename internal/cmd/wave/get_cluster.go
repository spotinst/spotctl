package wave

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/writer"
	"sort"

	"github.com/spotinst/spotctl/internal/flags"
)

type (
	CmdGetCluster struct {
		cmd  *cobra.Command
		opts CmdGetClusterOptions
	}

	CmdGetClusterOptions struct {
		*CmdGetOptions
		ClusterID   string
		ClusterName string
	}
)

func NewCmdGetCluster(opts *CmdGetOptions) *cobra.Command {
	return newCmdGetCluster(opts).cmd
}

func newCmdGetCluster(opts *CmdGetOptions) *CmdGetCluster {
	var cmd CmdGetCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Display one or many Wave clusters",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdGetCluster) Run(ctx context.Context) error {
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

func (x *CmdGetCluster) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdGetCluster) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGetCluster) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdGetCluster) run(ctx context.Context) error {

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

	clusters, err := waveClient.ListClusters(ctx)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(writer.Format(x.opts.Output))
	if err != nil {
		return err
	}

	sort.Sort(&spot.WaveClustersSorter{Clusters: clusters})

	return w.Write(clusters)
}

func (x *CmdGetClusterOptions) Init(fs *pflag.FlagSet, opts *CmdGetOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdGetClusterOptions) initDefaults(opts *CmdGetOptions) {
	x.CmdGetOptions = opts
}

func (x *CmdGetClusterOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "cluster id")
	fs.StringVar(&x.ClusterName, flags.FlagWaveClusterName, x.ClusterName, "cluster name")
}

func (x *CmdGetClusterOptions) Validate() error {
	/*if x.ClusterID == "" && x.ClusterName == "" {
		return errors.RequiredOr(flags.FlagWaveClusterID, flags.FlagWaveClusterName)
	}*/

	return x.CmdGetOptions.Validate()
}
