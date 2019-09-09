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
	CmdGetClusterKubernetesAWS struct {
		cmd  *cobra.Command
		opts CmdGetClusterKubernetesAWSOptions
	}

	CmdGetClusterKubernetesAWSOptions struct {
		*CmdGetClusterKubernetesOptions
	}
)

func NewCmdGetClusterKubernetesAWS(opts *CmdGetClusterKubernetesOptions) *cobra.Command {
	return newCmdGetClusterKubernetesAWS(opts).cmd
}

func newCmdGetClusterKubernetesAWS(opts *CmdGetClusterKubernetesOptions) *CmdGetClusterKubernetesAWS {
	var cmd CmdGetClusterKubernetesAWS

	cmd.cmd = &cobra.Command{
		Use:           "aws",
		Short:         "Display one or many Kubernetes clusters on AWS",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdGetClusterKubernetesAWS) Run(ctx context.Context) error {
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

func (x *CmdGetClusterKubernetesAWS) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdGetClusterKubernetesAWS) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGetClusterKubernetesAWS) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdGetClusterKubernetesAWS) run(ctx context.Context) error {
	spotinstClient, err := x.opts.Clients.NewSpotinst()
	if err != nil {
		return err
	}

	oceanClient, err := spotinstClient.Services().Ocean(spotinst.CloudProviderAWS, spotinst.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	clusters, err := oceanClient.ListClusters(ctx)
	if err != nil {
		return err
	}

	w, err := x.opts.Clients.NewWriter(writer.Format(x.opts.Output))
	if err != nil {
		return err
	}

	return w.Write(clusters)
}

func (x *CmdGetClusterKubernetesAWSOptions) Init(flags *pflag.FlagSet, opts *CmdGetClusterKubernetesOptions) {
	x.CmdGetClusterKubernetesOptions = opts
}

func (x *CmdGetClusterKubernetesAWSOptions) Validate() error {
	if err := x.CmdGetClusterKubernetesOptions.Validate(); err != nil {
		return err
	}

	return nil
}
