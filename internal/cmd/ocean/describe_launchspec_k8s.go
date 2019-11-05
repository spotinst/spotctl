package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/spotinst"
	"github.com/spotinst/spotctl/internal/utils/flags"
	"github.com/spotinst/spotctl/internal/writer/writers/json"
)

type (
	CmdDescribeLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdDescribeLaunchSpecKubernetesOptions
	}

	CmdDescribeLaunchSpecKubernetesOptions struct {
		*CmdDescribeLaunchSpecOptions

		LaunchSpecID string
	}
)

func NewCmdDescribeLaunchSpecKubernetes(opts *CmdDescribeLaunchSpecOptions) *cobra.Command {
	return newCmdDescribeLaunchSpecKubernetes(opts).cmd
}

func newCmdDescribeLaunchSpecKubernetes(opts *CmdDescribeLaunchSpecOptions) *CmdDescribeLaunchSpecKubernetes {
	var cmd CmdDescribeLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Describe a Kubernetes launch spec",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdDescribeLaunchSpecKubernetes) Run(ctx context.Context) error {
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

func (x *CmdDescribeLaunchSpecKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdDescribeLaunchSpecKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDescribeLaunchSpecKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDescribeLaunchSpecKubernetes) run(ctx context.Context) error {
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

	spec, err := oceanClient.GetLaunchSpec(ctx, x.opts.LaunchSpecID)
	if err != nil {
		return err
	}

	w, err := x.opts.Clients.NewWriter(json.WriterFormat)
	if err != nil {
		return err
	}

	return w.Write(spec.Obj)
}

func (x *CmdDescribeLaunchSpecKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdDescribeLaunchSpecOptions) {
	x.initFlags(fs)
	x.initDefaults(opts)
}

func (x *CmdDescribeLaunchSpecKubernetesOptions) initDefaults(opts *CmdDescribeLaunchSpecOptions) {
	x.CmdDescribeLaunchSpecOptions = opts
}

func (x *CmdDescribeLaunchSpecKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.LaunchSpecID, flags.FlagOceanSpecID, x.LaunchSpecID, "id of the launch spec")
}

func (x *CmdDescribeLaunchSpecKubernetesOptions) Validate() error {
	if err := x.CmdDescribeLaunchSpecOptions.Validate(); err != nil {
		return err
	}

	if x.LaunchSpecID == "" {
		return errors.Required("LaunchSpecID")
	}

	return nil
}
