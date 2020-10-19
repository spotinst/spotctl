package ocean

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spotinst"
)

type (
	CmdUpdateLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdUpdateLaunchSpecKubernetesOptions
	}

	CmdUpdateLaunchSpecKubernetesOptions struct {
		*CmdUpdateLaunchSpecOptions
		spotinst.OceanLaunchSpecOptions
	}
)

func NewCmdUpdateLaunchSpecKubernetes(opts *CmdUpdateLaunchSpecOptions) *cobra.Command {
	return newCmdUpdateLaunchSpecKubernetes(opts).cmd
}

func newCmdUpdateLaunchSpecKubernetes(opts *CmdUpdateLaunchSpecOptions) *CmdUpdateLaunchSpecKubernetes {
	var cmd CmdUpdateLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Update an existing Kubernetes launch spec",
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

func (x *CmdUpdateLaunchSpecKubernetes) Run(ctx context.Context) error {
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

func (x *CmdUpdateLaunchSpecKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdUpdateLaunchSpecKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdUpdateLaunchSpecKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdUpdateLaunchSpecKubernetes) run(ctx context.Context) error {
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

	oceanLaunchSpec, err := oceanClient.NewLaunchSpecBuilder(x.cmd.Flags(), &x.opts.OceanLaunchSpecOptions).Build()
	if err != nil {
		return err
	}

	spec, err := oceanClient.UpdateLaunchSpec(ctx, oceanLaunchSpec)
	if err != nil {
		return err
	}

	fmt.Fprintln(x.opts.Out, fmt.Sprintf("Updated (%q).", spec.ID))
	return err
}

func (x *CmdUpdateLaunchSpecKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdUpdateLaunchSpecOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdUpdateLaunchSpecKubernetesOptions) initDefaults(opts *CmdUpdateLaunchSpecOptions) {
	x.CmdUpdateLaunchSpecOptions = opts
}

func (x *CmdUpdateLaunchSpecKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	// Base.
	{
		fs.StringVar(&x.SpecID, flags.FlagOceanSpecID, x.SpecID, "name of the launch spec")
		fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
		fs.StringVar(&x.Name, flags.FlagOceanName, x.Name, "name of the launch spec")
	}

	// Compute.
	{
		fs.StringVar(&x.ImageID, flags.FlagOceanImageID, x.ImageID, "id of the image")
		fs.StringVar(&x.UserData, flags.FlagOceanUserData, x.UserData, "user data to provide when launching a node (plain-text or base64-encoded)")
	}
}

func (x *CmdUpdateLaunchSpecKubernetesOptions) Validate() error {
	return x.CmdUpdateLaunchSpecOptions.Validate()
}
