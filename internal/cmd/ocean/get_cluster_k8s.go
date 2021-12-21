package ocean

import (
	"context"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/writer"
)

type (
	CmdGetClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdGetClusterKubernetesOptions
	}

	CmdGetClusterKubernetesOptions struct {
		*CmdGetClusterOptions
	}
)

func NewCmdGetClusterKubernetes(opts *CmdGetClusterOptions) *cobra.Command {
	return newCmdGetClusterKubernetes(opts).cmd
}

func newCmdGetClusterKubernetes(opts *CmdGetClusterOptions) *CmdGetClusterKubernetes {
	var cmd CmdGetClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Display one or many Kubernetes clusters",
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

func (x *CmdGetClusterKubernetes) Run(ctx context.Context) error {
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

func (x *CmdGetClusterKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdGetClusterKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGetClusterKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdGetClusterKubernetes) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
		spot.WithDryRun(x.opts.DryRun),
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	clusters, err := oceanClient.ListClusters(ctx)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(writer.Format(x.opts.Output))
	if err != nil {
		return err
	}

	sort.Sort(&spot.OceanClustersSorter{Clusters: clusters})

	return w.Write(clusters)
}

func (x *CmdGetClusterKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdGetClusterOptions) {
	x.CmdGetClusterOptions = opts
}

func (x *CmdGetClusterKubernetesOptions) Validate() error {
	return x.CmdGetClusterOptions.Validate()
}
