package wave

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/writer/writers/json"
)

type (
	CmdDescribeSparkApplication struct {
		cmd  *cobra.Command
		opts CmdDescribeSparkApplicationOptions
	}

	CmdDescribeSparkApplicationOptions struct {
		*CmdDescribeOptions
		ID string
	}
)

func NewCmdDescribeSparkApplication(opts *CmdDescribeOptions) *cobra.Command {
	return newCmdDescribeSparkApplication(opts).cmd
}

func newCmdDescribeSparkApplication(opts *CmdDescribeOptions) *CmdDescribeSparkApplication {
	var cmd CmdDescribeSparkApplication

	cmd.cmd = &cobra.Command{
		Use:           "sparkapplication",
		Short:         "Describe a Wave Spark application",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdDescribeSparkApplicationOptions) Init(fs *pflag.FlagSet, opts *CmdDescribeOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdDescribeSparkApplicationOptions) initDefaults(opts *CmdDescribeOptions) {
	x.CmdDescribeOptions = opts
}

func (x *CmdDescribeSparkApplicationOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ID, flags.FlagWaveSparkApplicationEntityID, x.ID, "id of the spark application")
}

func (x *CmdDescribeSparkApplication) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdDescribeSparkApplicationOptions) Validate() error {
	if x.ID == "" {
		return errors.Required(flags.FlagWaveSparkApplicationEntityID)
	}
	return x.CmdDescribeOptions.Validate()
}

func (x *CmdDescribeSparkApplication) Run(ctx context.Context) error {
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

func (x *CmdDescribeSparkApplication) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDescribeSparkApplication) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDescribeSparkApplication) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	waveClient, err := spotClient.Services().Wave()
	if err != nil {
		return err
	}

	sparkApplication, err := waveClient.GetSparkApplication(ctx, x.opts.ID)
	if err != nil {
		return err
	}

	w, err := x.opts.Clientset.NewWriter(json.WriterFormat)
	if err != nil {
		return err
	}

	return w.Write(sparkApplication.Obj)
}
