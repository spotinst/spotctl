package ocean

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spotinst"
)

type (
	CmdRolloutStartKubernetes struct {
		cmd  *cobra.Command
		opts CmdRolloutStartKubernetesOptions
	}

	CmdRolloutStartKubernetesOptions struct {
		*CmdRolloutStartOptions
		spotinst.OceanRolloutOptions
	}
)

func NewCmdRolloutStartKubernetes(opts *CmdRolloutStartOptions) *cobra.Command {
	return newCmdRolloutStartKubernetes(opts).cmd
}

func newCmdRolloutStartKubernetes(opts *CmdRolloutStartOptions) *CmdRolloutStartKubernetes {
	var cmd CmdRolloutStartKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Start a new rollout of a Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdRolloutStartKubernetes) Run(ctx context.Context) error {
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

func (x *CmdRolloutStartKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdRolloutStartKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdRolloutStartKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdRolloutStartKubernetes) run(ctx context.Context) error {
	spotinstClientOpts := []spotinst.ClientOption{
		spotinst.WithCredentialsProfile(x.opts.Profile),
	}

	spotinstClient, err := x.opts.Clientset.NewSpotinst(spotinstClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotinstClient.Services().Ocean(x.opts.CloudProvider, spotinst.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	oceanRollout, err := oceanClient.NewRolloutBuilder(x.cmd.Flags(), &x.opts.OceanRolloutOptions).Build()
	if err != nil {
		return err
	}

	rollout, err := oceanClient.CreateRollout(ctx, oceanRollout)
	if err != nil {
		return err
	}

	fmt.Fprintln(x.opts.Out, fmt.Sprintf("Started (%q).", rollout.ID))
	return nil
}

func (x *CmdRolloutStartKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdRolloutStartOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdRolloutStartKubernetesOptions) initDefaults(opts *CmdRolloutStartOptions) {
	x.CmdRolloutStartOptions = opts
	x.Comment = "created by @spotinst/spotctl"
}

func (x *CmdRolloutStartKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	// Base.
	{
		fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "")
		fs.StringVar(&x.Comment, flags.FlagOceanRolloutComment, x.Comment, "")
	}

	// Parameters.
	{
		fs.IntVar(&x.BatchSizePercentage, flags.FlagOceanRolloutBatchSizePercentage, x.BatchSizePercentage, "")
		fs.BoolVar(&x.DisableAutoScaling, flags.FlagOceanRolloutDisableAutoScaling, x.DisableAutoScaling, "")
		fs.StringSliceVar(&x.SpecIDs, flags.FlagOceanRolloutSpecIDs, x.SpecIDs, "")
		fs.StringSliceVar(&x.InstanceIDs, flags.FlagOceanRolloutInstanceIDs, x.InstanceIDs, "")
	}
}

func (x *CmdRolloutStartKubernetesOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdRolloutStartOptions.Validate(); err != nil {
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
