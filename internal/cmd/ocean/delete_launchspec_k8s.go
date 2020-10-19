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
	CmdDeleteLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdDeleteLaunchSpecKubernetesOptions
	}

	CmdDeleteLaunchSpecKubernetesOptions struct {
		*CmdDeleteLaunchSpecOptions

		LaunchSpecID string
	}
)

func NewCmdDeleteLaunchSpecKubernetes(opts *CmdDeleteLaunchSpecOptions) *cobra.Command {
	return newCmdDeleteLaunchSpecKubernetes(opts).cmd
}

func newCmdDeleteLaunchSpecKubernetes(opts *CmdDeleteLaunchSpecOptions) *CmdDeleteLaunchSpecKubernetes {
	var cmd CmdDeleteLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Delete a Kubernetes launch spec",
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

func (x *CmdDeleteLaunchSpecKubernetes) Run(ctx context.Context) error {
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

func (x *CmdDeleteLaunchSpecKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdDeleteLaunchSpecKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDeleteLaunchSpecKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDeleteLaunchSpecKubernetes) run(ctx context.Context) error {
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

	return oceanClient.DeleteLaunchSpec(ctx, x.opts.LaunchSpecID)
}

func (x *CmdDeleteLaunchSpecKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdDeleteLaunchSpecOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdDeleteLaunchSpecKubernetesOptions) initDefaults(opts *CmdDeleteLaunchSpecOptions) {
	x.CmdDeleteLaunchSpecOptions = opts
}

func (x *CmdDeleteLaunchSpecKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.LaunchSpecID, flags.FlagOceanSpecID, x.LaunchSpecID, "id of the launch spec")
}

func (x *CmdDeleteLaunchSpecKubernetesOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdDeleteLaunchSpecOptions.Validate(); err != nil {
		errg.Add(err)
	}

	if x.LaunchSpecID == "" {
		errg.Add(errors.Required("LaunchSpecID"))
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}
