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
	CmdGetLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdGetLaunchSpecKubernetesOptions
	}

	CmdGetLaunchSpecKubernetesOptions struct {
		*CmdGetLaunchSpecOptions
	}
)

func NewCmdGetLaunchSpecKubernetes(opts *CmdGetLaunchSpecOptions) *cobra.Command {
	return newCmdGetLaunchSpecKubernetes(opts).cmd
}

func newCmdGetLaunchSpecKubernetes(opts *CmdGetLaunchSpecOptions) *CmdGetLaunchSpecKubernetes {
	var cmd CmdGetLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Display one or many Kubernetes launch specs",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdGetLaunchSpecKubernetes) Run(ctx context.Context) error {
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

func (x *CmdGetLaunchSpecKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdGetLaunchSpecKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGetLaunchSpecKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdGetLaunchSpecKubernetes) run(ctx context.Context) error {
	spotinstClientOpts := []spotinst.ClientOption{
		spotinst.WithCredentialsProfile(x.opts.Profile),
	}

	spotinstClient, err := x.opts.Clients.NewSpotinst(spotinstClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotinstClient.Services().Ocean(x.opts.CloudProvider, spotinst.OrchestratorKubernetes)
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

func (x *CmdGetLaunchSpecKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdGetLaunchSpecOptions) {
	x.CmdGetLaunchSpecOptions = opts
}

func (x *CmdGetLaunchSpecKubernetesOptions) Validate() error {
	return x.CmdGetLaunchSpecOptions.Validate()
}
