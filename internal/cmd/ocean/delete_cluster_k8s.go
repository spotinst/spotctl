package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spotinst"
)

type (
	CmdDeleteClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdDeleteClusterKubernetesOptions
	}

	CmdDeleteClusterKubernetesOptions struct {
		*CmdDeleteClusterOptions

		ClusterID string
	}
)

func NewCmdDeleteClusterKubernetes(opts *CmdDeleteClusterOptions) *cobra.Command {
	return newCmdDeleteClusterKubernetes(opts).cmd
}

func newCmdDeleteClusterKubernetes(opts *CmdDeleteClusterOptions) *CmdDeleteClusterKubernetes {
	var cmd CmdDeleteClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Delete a Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"k8s", "kube", "k"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdDeleteClusterKubernetes) Run(ctx context.Context) error {
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

func (x *CmdDeleteClusterKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdDeleteClusterKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDeleteClusterKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDeleteClusterKubernetes) run(ctx context.Context) error {
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

	return oceanClient.DeleteCluster(ctx, x.opts.ClusterID)
}

func (x *CmdDeleteClusterKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdDeleteClusterOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdDeleteClusterKubernetesOptions) initDefaults(opts *CmdDeleteClusterOptions) {
	x.CmdDeleteClusterOptions = opts
}

func (x *CmdDeleteClusterKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdDeleteClusterKubernetesOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdDeleteClusterOptions.Validate(); err != nil {
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
