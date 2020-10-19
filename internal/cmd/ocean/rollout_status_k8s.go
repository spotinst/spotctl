package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spotinst"
	"github.com/spotinst/spotctl/internal/writer/writers/json"
)

type (
	CmdRolloutStatusKubernetes struct {
		cmd  *cobra.Command
		opts CmdRolloutStatusKubernetesOptions
	}

	CmdRolloutStatusKubernetesOptions struct {
		*CmdRolloutStatusOptions

		ClusterID string
		RolloutID string
	}
)

func NewCmdRolloutStatusKubernetes(opts *CmdRolloutStatusOptions) *cobra.Command {
	return newCmdRolloutStatusKubernetes(opts).cmd
}

func newCmdRolloutStatusKubernetes(opts *CmdRolloutStatusOptions) *CmdRolloutStatusKubernetes {
	var cmd CmdRolloutStatusKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Show the status of a Kubernetes rollout",
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

func (x *CmdRolloutStatusKubernetes) Run(ctx context.Context) error {
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

func (x *CmdRolloutStatusKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdRolloutStatusKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdRolloutStatusKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdRolloutStatusKubernetes) run(ctx context.Context) error {
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

	rollout, err := oceanClient.GetRollout(ctx, x.opts.ClusterID, x.opts.RolloutID)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(json.WriterFormat)
	if err != nil {
		return err
	}

	return w.Write(rollout.Obj)
}

func (x *CmdRolloutStatusKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdRolloutStatusOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdRolloutStatusKubernetesOptions) initDefaults(opts *CmdRolloutStatusOptions) {
	x.CmdRolloutStatusOptions = opts
}

func (x *CmdRolloutStatusKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "")
	fs.StringVar(&x.RolloutID, flags.FlagOceanRolloutID, x.RolloutID, "")
}

func (x *CmdRolloutStatusKubernetesOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdRolloutStatusOptions.Validate(); err != nil {
		errg.Add(err)
	}

	if x.ClusterID == "" {
		errg.Add(errors.Required("ClusterID"))
	}

	if x.ClusterID == "" {
		errg.Add(errors.Required("RolloutID"))
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}
