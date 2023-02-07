package ocean

import (
	"context"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/writer"
)

type (
	CmdSparkGetCluster struct {
		cmd  *cobra.Command
		opts CmdSparkGetClusterOptions
	}

	CmdSparkGetClusterOptions struct {
		*CmdSparkGetOptions

		ControllerClusterID string
		State               string
	}
)

func NewCmdSparkGetCluster(opts *CmdSparkGetOptions) *cobra.Command {
	return newCmdSparkGetCluster(opts).cmd
}

func newCmdSparkGetCluster(opts *CmdSparkGetOptions) *CmdSparkGetCluster {
	var cmd CmdSparkGetCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Display one or many Ocean for Apache Spark clusters",
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

func (x *CmdSparkGetCluster) Run(ctx context.Context) error {
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

func (x *CmdSparkGetCluster) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdSparkGetCluster) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdSparkGetCluster) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdSparkGetCluster) run(ctx context.Context) error {
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

	clusters, err := oceanSparkClient.ListClusters(ctx, x.opts.ControllerClusterID, x.opts.State)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(writer.Format(x.opts.Output))
	if err != nil {
		return err
	}

	sort.Sort(&spot.OceanSparkClustersSorter{Clusters: clusters})

	return w.Write(clusters)
}

func (x *CmdSparkGetClusterOptions) Init(fs *pflag.FlagSet, opts *CmdSparkGetOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdSparkGetClusterOptions) initDefaults(opts *CmdSparkGetOptions) {
	x.CmdSparkGetOptions = opts
	x.NoHeaders = false
	x.Output = "table"
}

func (x *CmdSparkGetClusterOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ControllerClusterID, flags.FlagOFASName, x.ControllerClusterID, "")
	fs.StringVar(&x.State, flags.FlagOFASState, x.State, "")
}

func (x *CmdSparkGetClusterOptions) Validate() error {
	return x.CmdSparkGetOptions.Validate()
}
