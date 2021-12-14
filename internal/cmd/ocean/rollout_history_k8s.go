package ocean

import (
	"context"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/writer"
)

type (
	CmdRolloutHistoryKubernetes struct {
		cmd  *cobra.Command
		opts CmdRolloutHistoryKubernetesOptions
	}

	CmdRolloutHistoryKubernetesOptions struct {
		*CmdRolloutHistoryOptions

		NoHeaders bool
		Output    string
		ClusterID string
	}
)

func NewCmdRolloutHistoryKubernetes(opts *CmdRolloutHistoryOptions) *cobra.Command {
	return newCmdRolloutHistoryKubernetes(opts).cmd
}

func newCmdRolloutHistoryKubernetes(opts *CmdRolloutHistoryOptions) *CmdRolloutHistoryKubernetes {
	var cmd CmdRolloutHistoryKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "View rollout history of a Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"k8s", "kube", "k"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdRolloutHistoryKubernetes) Run(ctx context.Context) error {
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

func (x *CmdRolloutHistoryKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdRolloutHistoryKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdRolloutHistoryKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdRolloutHistoryKubernetes) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
		spot.WithDryRun(x.opts.DryRun),
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	rollouts, err := oceanClient.ListRollouts(ctx, x.opts.ClusterID)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(writer.Format(x.opts.Output))
	if err != nil {
		return err
	}

	sort.Sort(&spot.OceanRolloutsSorter{Rollouts: rollouts})

	return w.Write(rollouts)
}

func (x *CmdRolloutHistoryKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdRolloutHistoryOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdRolloutHistoryKubernetesOptions) initDefaults(opts *CmdRolloutHistoryOptions) {
	x.CmdRolloutHistoryOptions = opts
	x.NoHeaders = false
	x.Output = "table"
}

func (x *CmdRolloutHistoryKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "")
}

func (x *CmdRolloutHistoryKubernetesOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdRolloutHistoryOptions.Validate(); err != nil {
		errg.Add(err)
	}

	if x.ClusterID == "" {
		errg.Add(errors.Required("ClusterID"))
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}
