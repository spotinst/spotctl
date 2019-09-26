package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/spotinst"
	"github.com/spotinst/spotinst-cli/internal/utils/flags"
	"github.com/spotinst/spotinst-cli/internal/writer"
)

type (
	CmdGetLaunchSpecKubernetesAWS struct {
		cmd  *cobra.Command
		opts CmdGetLaunchSpecKubernetesAWSOptions
	}

	CmdGetLaunchSpecKubernetesAWSOptions struct {
		*CmdGetLaunchSpecKubernetesOptions
	}
)

func NewCmdGetLaunchSpecKubernetesAWS(opts *CmdGetLaunchSpecKubernetesOptions) *cobra.Command {
	return newCmdGetLaunchSpecKubernetesAWS(opts).cmd
}

func newCmdGetLaunchSpecKubernetesAWS(opts *CmdGetLaunchSpecKubernetesOptions) *CmdGetLaunchSpecKubernetesAWS {
	var cmd CmdGetLaunchSpecKubernetesAWS

	cmd.cmd = &cobra.Command{
		Use:           "aws",
		Short:         "Display one or many Kubernetes launch specs on AWS",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdGetLaunchSpecKubernetesAWS) Run(ctx context.Context) error {
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

func (x *CmdGetLaunchSpecKubernetesAWS) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdGetLaunchSpecKubernetesAWS) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGetLaunchSpecKubernetesAWS) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdGetLaunchSpecKubernetesAWS) run(ctx context.Context) error {
	spotinstClientOpts := []spotinst.ClientOption{
		spotinst.WithCredentialsProfile(x.opts.Profile),
	}

	spotinstClient, err := x.opts.Clients.NewSpotinst(spotinstClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotinstClient.Services().Ocean(spotinst.CloudProviderAWS, spotinst.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	specs, err := oceanClient.ListLaunchSpecs(ctx)
	if err != nil {
		return err
	}

	w, err := x.opts.Clients.NewWriter(writer.Format(x.opts.Output))
	if err != nil {
		return err
	}

	return w.Write(specs)
}

func (x *CmdGetLaunchSpecKubernetesAWSOptions) Init(flags *pflag.FlagSet, opts *CmdGetLaunchSpecKubernetesOptions) {
	x.CmdGetLaunchSpecKubernetesOptions = opts
}

func (x *CmdGetLaunchSpecKubernetesAWSOptions) Validate() error {
	return x.CmdGetLaunchSpecKubernetesOptions.Validate()
}
