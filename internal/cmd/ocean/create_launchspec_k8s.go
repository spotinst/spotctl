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
	CmdCreateLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdCreateLaunchSpecKubernetesOptions
	}

	CmdCreateLaunchSpecKubernetesOptions struct {
		*CmdCreateLaunchSpecOptions
		spotinst.OceanLaunchSpecOptions
	}
)

func NewCmdCreateLaunchSpecKubernetes(opts *CmdCreateLaunchSpecOptions) *cobra.Command {
	return newCmdCreateLaunchSpecKubernetes(opts).cmd
}

func newCmdCreateLaunchSpecKubernetes(opts *CmdCreateLaunchSpecOptions) *CmdCreateLaunchSpecKubernetes {
	var cmd CmdCreateLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Create a new Kubernetes launch spec",
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

func (x *CmdCreateLaunchSpecKubernetes) Run(ctx context.Context) error {
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

func (x *CmdCreateLaunchSpecKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdCreateLaunchSpecKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdCreateLaunchSpecKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdCreateLaunchSpecKubernetes) run(ctx context.Context) error {
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

	spec, err := oceanClient.CreateLaunchSpec(ctx, oceanLaunchSpec)
	if err != nil {
		return err
	}

	fmt.Fprintln(x.opts.Out, fmt.Sprintf("Created (%q).", spec.ID))
	return nil
}

func (x *CmdCreateLaunchSpecKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdCreateLaunchSpecOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdCreateLaunchSpecKubernetesOptions) initDefaults(opts *CmdCreateLaunchSpecOptions) {
	x.CmdCreateLaunchSpecOptions = opts
}

func (x *CmdCreateLaunchSpecKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	// Base.
	{
		fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
		fs.StringVar(&x.Name, flags.FlagOceanName, x.Name, "name of the launch spec")
	}

	// Compute.
	{
		fs.StringVar(&x.ImageID, flags.FlagOceanImageID, x.ImageID, "id of the image")
		fs.StringVar(&x.UserData, flags.FlagOceanUserData, x.UserData, "user data to provide when launching a node (plain-text or base64-encoded)")
	}
}

func (x *CmdCreateLaunchSpecKubernetesOptions) Validate() error {
	return x.CmdCreateLaunchSpecOptions.Validate()
}
