package wave

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
)

type CmdGet struct {
	cmd  *cobra.Command
	opts CmdGetOptions
}

type CmdGetOptions struct {
	*CmdOptions
	ClusterID string
}

func (x *CmdGetOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
}

func NewCmdGet(opts *CmdOptions) *cobra.Command {
	return newCmdGet(opts).cmd
}

func newCmdGet(opts *CmdOptions) *CmdGet {
	var cmd CmdGet

	cmd.cmd = &cobra.Command{
		Use:           "get",
		Short:         "Get a Wave installation",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdGetOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
	x.initFlags(fs)
}

func (x *CmdGet) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdGetOptions) Validate() error {
	if x.ClusterID == "" {
		return fmt.Errorf("--cluster-id must be specified")
	}
	return x.CmdOptions.Validate()
}

func (x *CmdGet) Run(ctx context.Context) error {
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

func (x *CmdGet) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGet) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdGet) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
	}

	_, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	fmt.Fprintln(x.opts.Out, fmt.Sprintf("blah blah blah"))
	return nil
}
