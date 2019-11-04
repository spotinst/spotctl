package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/errors"
	"github.com/spotinst/spotinst-cli/internal/spotinst"
	"github.com/spotinst/spotinst-cli/internal/utils/flags"
	"github.com/spotinst/spotinst-cli/internal/writer/writers/json"
)

type (
	CmdDescribeClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdDescribeClusterKubernetesOptions
	}

	CmdDescribeClusterKubernetesOptions struct {
		*CmdDescribeClusterOptions

		ClusterID string
	}
)

func NewCmdDescribeClusterKubernetes(opts *CmdDescribeClusterOptions) *cobra.Command {
	return newCmdDescribeClusterKubernetes(opts).cmd
}

func newCmdDescribeClusterKubernetes(opts *CmdDescribeClusterOptions) *CmdDescribeClusterKubernetes {
	var cmd CmdDescribeClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Describe a Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdDescribeClusterKubernetes) Run(ctx context.Context) error {
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

func (x *CmdDescribeClusterKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdDescribeClusterKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDescribeClusterKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDescribeClusterKubernetes) run(ctx context.Context) error {
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

	cluster, err := oceanClient.GetCluster(ctx, x.opts.ClusterID)
	if err != nil {
		return err
	}

	w, err := x.opts.Clients.NewWriter(json.WriterFormat)
	if err != nil {
		return err
	}

	return w.Write(cluster.Obj)
}

func (x *CmdDescribeClusterKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdDescribeClusterOptions) {
	x.initFlags(fs)
	x.initDefaults(opts)
}

func (x *CmdDescribeClusterKubernetesOptions) initDefaults(opts *CmdDescribeClusterOptions) {
	x.CmdDescribeClusterOptions = opts
}

func (x *CmdDescribeClusterKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdDescribeClusterKubernetesOptions) Validate() error {
	if err := x.CmdDescribeClusterOptions.Validate(); err != nil {
		return err
	}

	if x.ClusterID == "" {
		return errors.Required("ClusterID")
	}

	return nil
}
