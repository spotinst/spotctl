package wave

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/writer"
	"github.com/spotinst/spotinst-sdk-go/service/wave"
	"sort"
	"strings"
)

type (
	CmdGetCluster struct {
		cmd  *cobra.Command
		opts CmdGetClusterOptions
	}

	CmdGetClusterOptions struct {
		*CmdGetOptions
		ClusterID    string
		ClusterName  string
		ClusterState string
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

	var clusters []*spot.WaveCluster
	if x.opts.ClusterID != "" {
		cluster, err := waveClient.GetCluster(ctx, x.opts.ClusterID)
		if err != nil {
			return err
		}
		clusters = make([]*spot.WaveCluster, 1)
		clusters[0] = cluster
	} else {
		clusterState := strings.ToUpper(x.opts.ClusterState)
		clusters, err = waveClient.ListClusters(ctx, x.opts.ClusterName, clusterState)
		if err != nil {
			return err
		}
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
	fs.StringVar(&x.ClusterState, flags.FlagWaveClusterState, x.ClusterState, "cluster state")
}

func (x *CmdGetClusterOptions) Validate() error {
	if x.ClusterID != "" && x.ClusterName != "" {
		return errors.RequiredXor(flags.FlagWaveClusterID, flags.FlagWaveClusterName)
	}
	if x.ClusterState != "" {
		if !validateClusterState(x.ClusterState) {
			return errors.Invalid(flags.FlagWaveClusterState, x.ClusterState)
		}
	}
	return x.CmdGetOptions.Validate()
}

func validateClusterState(state string) bool {
	clusterState := wave.ClusterState(strings.ToUpper(state))
	switch clusterState {
	case wave.ClusterDegraded, wave.ClusterAvailable, wave.ClusterFailing, wave.ClusterProgressing, wave.ClusterUnknown:
		return true
	default:
		return false
	}
}
