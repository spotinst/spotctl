package wave

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/spotinst"
	"github.com/spotinst/spotctl/internal/wave"
)

type CmdCreate struct {
	cmd  *cobra.Command
	opts CmdCreateOptions
}

type CmdCreateOptions struct {
	*CmdOptions
	ClusterID string
}

func (x *CmdCreateOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
}

func NewCmdCreate(opts *CmdOptions) *cobra.Command {
	return newCmdCreate(opts).cmd
}

func newCmdCreate(opts *CmdOptions) *CmdCreate {
	var cmd CmdCreate

	cmd.cmd = &cobra.Command{
		Use:           "create",
		Short:         "Create a new wave installation",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdCreateOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
	x.initFlags(fs)
}

func (x *CmdCreate) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdCreateOptions) Validate() error {
	if x.ClusterID == "" {
		return fmt.Errorf("--cluster-id must be specified")
	}
	return x.CmdOptions.Validate()
}

func (x *CmdCreate) Run(ctx context.Context) error {
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

func (x *CmdCreate) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdCreate) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdCreate) run(ctx context.Context) error {
	spotinstClientOpts := []spotinst.ClientOption{
		spotinst.WithCredentialsProfile(x.opts.Profile),
	}

	spotClient, err := x.opts.Clientset.NewSpotinst(spotinstClientOpts...)
	if err != nil {
		return err
	}
	oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spotinst.OrchestratorKubernetes)
	if err != nil {
		return err
	}
	c, err := oceanClient.GetCluster(ctx, x.opts.ClusterID)
	if err != nil {
		return err
	}
	log.Infof("Verified cluster %s", c.Name)
	manager, err := wave.NewManager(c.Name, getWaveLogger())
	if err != nil {
		return err
	}

	return manager.Create()
}
