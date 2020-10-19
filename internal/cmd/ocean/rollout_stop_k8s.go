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
	CmdRolloutStopKubernetes struct {
		cmd  *cobra.Command
		opts CmdRolloutStopKubernetesOptions
	}

	CmdRolloutStopKubernetesOptions struct {
		*CmdRolloutStopOptions
		spotinst.OceanRolloutOptions
	}
)

func NewCmdRolloutStopKubernetes(opts *CmdRolloutStopOptions) *cobra.Command {
	return newCmdRolloutStopKubernetes(opts).cmd
}

func newCmdRolloutStopKubernetes(opts *CmdRolloutStopOptions) *CmdRolloutStopKubernetes {
	var cmd CmdRolloutStopKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Stop an in-progress rollout of a Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdRolloutStopKubernetes) Run(ctx context.Context) error {
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

func (x *CmdRolloutStopKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdRolloutStopKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdRolloutStopKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdRolloutStopKubernetes) run(ctx context.Context) error {
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

	rollout, err := oceanClient.UpdateRollout(ctx, oceanRollout)
	if err != nil {
		return err
	}

	fmt.Fprintln(x.opts.Out, fmt.Sprintf("Stopped (%q).", rollout.ID))
	return nil
}

func (x *CmdRolloutStopKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdRolloutStopOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdRolloutStopKubernetesOptions) initDefaults(opts *CmdRolloutStopOptions) {
	x.CmdRolloutStopOptions = opts
	x.OceanRolloutOptions.Status = "STOPPED"
}

func (x *CmdRolloutStopKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "")
	fs.StringVar(&x.RolloutID, flags.FlagOceanRolloutID, x.RolloutID, "")
}

func (x *CmdRolloutStopKubernetesOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdRolloutStopOptions.Validate(); err != nil {
		errg.Add(err)
	}

	if x.ClusterID == "" {
		errg.Add(errors.Required("ClusterID"))
	}

	if x.RolloutID == "" {
		errg.Add(errors.Required("RolloutID"))
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}
