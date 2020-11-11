package wave

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spotinst"
)

type CmdDescribe struct {
	cmd  *cobra.Command
	opts CmdDescribeOptions
}

type CmdDescribeOptions struct {
	*CmdOptions
	ClusterID string
}

func (x *CmdDescribeOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
}

func NewCmdDescribe(opts *CmdOptions) *cobra.Command {
	return newCmdDescribe(opts).cmd
}

func newCmdDescribe(opts *CmdOptions) *CmdDescribe {
	var cmd CmdDescribe

	cmd.cmd = &cobra.Command{
		Use:           "describe",
		Short:         "Describe a wave installation",
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
	if x.ClusterID == "" {
		return fmt.Errorf("--cluster-id must be specified")
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
	spotinstClientOpts := []spotinst.ClientOption{
		spotinst.WithCredentialsProfile(x.opts.Profile),
	}

	_, err := x.opts.Clientset.NewSpotinst(spotinstClientOpts...)
	if err != nil {
		return err
	}

	fmt.Fprintln(x.opts.Out, fmt.Sprintf("blah blah blah"))
	return nil
}
