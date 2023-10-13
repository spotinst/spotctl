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
	CmdGetClusterEcs struct {
		cmd  *cobra.Command
		opts CmdGetClusterEcsOptions
	}

	CmdGetClusterEcsOptions struct {
		*CmdGetClusterOptions
	}
)

func NewCmdGetClusterEcs(opts *CmdGetClusterOptions) *cobra.Command {
	return newCmdGetClusterEcs(opts).cmd
}

func newCmdGetClusterEcs(opts *CmdGetClusterOptions) *CmdGetClusterEcs {
	var cmd CmdGetClusterEcs

	cmd.cmd = &cobra.Command{
		Use:           "ecs",
		Short:         "Display one or many ecs clusters",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"ecs"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdGetClusterEcs) Run(ctx context.Context) error {
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

func (x *CmdGetClusterEcs) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdGetClusterEcs) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGetClusterEcs) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdGetClusterEcs) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
		spot.WithDryRun(x.opts.DryRun),
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorECS)
	if err != nil {
		return err
	}

	clusters, err := oceanClient.ListClusters(ctx)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(writer.Format(x.opts.Output))
	if err != nil {
		return err
	}

	sort.Sort(&spot.OceanClustersSorter{Clusters: clusters})

	return w.Write(clusters)
}

func (x *CmdGetClusterEcsOptions) Init(fs *pflag.FlagSet, opts *CmdGetClusterOptions) {
	x.CmdGetClusterOptions = opts
}

func (x *CmdGetClusterEcsOptions) Validate() error {
	return x.CmdGetClusterOptions.Validate()
}
