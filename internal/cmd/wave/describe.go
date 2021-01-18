package wave

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/wave"
)

type CmdDescribe struct {
	cmd  *cobra.Command
	opts CmdDescribeOptions
}

type CmdDescribeOptions struct {
	*CmdOptions
	ClusterID   string
	ClusterName string
}

func (x *CmdDescribeOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "cluster id")
	fs.StringVar(&x.ClusterName, flags.FlagWaveClusterName, x.ClusterName, "cluster name")
}

func NewCmdDescribe(opts *CmdOptions) *cobra.Command {
	return newCmdDescribe(opts).cmd
}

func newCmdDescribe(opts *CmdOptions) *CmdDescribe {
	var cmd CmdDescribe

	cmd.cmd = &cobra.Command{
		Use:           "describe",
		Short:         "Describe a Wave installation",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdDescribeOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
	x.initFlags(fs)
}

func (x *CmdDescribe) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdDescribeOptions) Validate() error {
	if x.ClusterID == "" && x.ClusterName == "" {
		return errors.RequiredOr(flags.FlagWaveClusterID, flags.FlagWaveClusterName)
	}
	return x.CmdOptions.Validate()
}

func (x *CmdDescribe) Run(ctx context.Context) error {
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

func (x *CmdDescribe) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDescribe) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDescribe) run(ctx context.Context) error {
	if x.opts.ClusterID != "" {
		spotClientOpts := []spot.ClientOption{
			spot.WithCredentialsProfile(x.opts.Profile),
		}

		spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
		if err != nil {
			return err
		}

		oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
		if err != nil {
			return err
		}

		c, err := oceanClient.GetCluster(ctx, x.opts.ClusterID)
		if err != nil {
			return err
		}

		x.opts.ClusterName = c.Name
	}

	if err := validateClusterContext(x.opts.ClusterName); err != nil {
		return fmt.Errorf("cluster context validation failure, %w", err)
	}

	// TODO Move to tide
	manager, err := wave.NewManager(x.opts.ClusterName, getWaveLogger())
	if err != nil {
		return err
	}

	return manager.Describe()
}
