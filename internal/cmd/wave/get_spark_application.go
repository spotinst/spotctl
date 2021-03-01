package wave

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/writer"
	"sort"
)

type (
	CmdGetSparkApplication struct {
		cmd  *cobra.Command
		opts CmdGetSparkApplicationOptions
	}

	CmdGetSparkApplicationOptions struct {
		*CmdGetOptions
		ClusterName          string
		ApplicationName      string
		ApplicationNamespace string
		ApplicationId        string
		ApplicationSparkId   string
		ApplicationState     string
		ApplicationHeritage  string
	}
)

func NewCmdGetSparkApplication(opts *CmdGetOptions) *cobra.Command {
	return newCmdGetSparkApplication(opts).cmd
}

func newCmdGetSparkApplication(opts *CmdGetOptions) *CmdGetSparkApplication {
	var cmd CmdGetSparkApplication

	cmd.cmd = &cobra.Command{
		Use:           "sparkapplication",
		Short:         "Display one or many Wave Spark applications",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdGetSparkApplication) Run(ctx context.Context) error {
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

func (x *CmdGetSparkApplication) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdGetSparkApplication) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdGetSparkApplication) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdGetSparkApplication) run(ctx context.Context) error {

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

	var sparkApplications []*spot.SparkApplication
	if x.opts.ApplicationId != "" {
		sparkApplication, err := waveClient.GetSparkApplication(ctx, x.opts.ApplicationId)
		if err != nil {
			return err
		}
		sparkApplications = make([]*spot.SparkApplication, 1)
		sparkApplications[0] = sparkApplication
	} else {
		filter := &spot.SparkApplicationsFilter{
			ClusterIdentifier: x.opts.ClusterName,
			Name:              x.opts.ApplicationName,
			Namespace:         x.opts.ApplicationNamespace,
			ApplicationId:     x.opts.ApplicationSparkId,
			ApplicationState:  x.opts.ApplicationState,
			Heritage:          x.opts.ApplicationHeritage,
		}
		sparkApplications, err = waveClient.ListSparkApplications(ctx, filter)
		if err != nil {
			return err
		}
	}

	// Should the json writer just write out the json as is? like in describe (cluster.obj)
	w, err := x.opts.Clientset.NewWriter(writer.Format(x.opts.Output))
	if err != nil {
		return err
	}

	sort.Sort(&spot.SparkApplicationsSorter{SparkApplications: sparkApplications})

	return w.Write(sparkApplications)
}

func (x *CmdGetSparkApplicationOptions) Init(fs *pflag.FlagSet, opts *CmdGetOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdGetSparkApplicationOptions) initDefaults(opts *CmdGetOptions) {
	x.CmdGetOptions = opts
}

func (x *CmdGetSparkApplicationOptions) initFlags(fs *pflag.FlagSet) {
	// TODO Remove some of these flags ... only allow filtering by table columns?
	fs.StringVar(&x.ClusterName, flags.FlagWaveClusterName, x.ClusterName, "cluster name")
	fs.StringVar(&x.ApplicationName, flags.FlagWaveSparkApplicationName, x.ApplicationName, "application name")
	fs.StringVar(&x.ApplicationNamespace, flags.FlagWaveSparkApplicationNamespace, x.ApplicationNamespace, "application namespace")
	fs.StringVar(&x.ApplicationId, flags.FlagWaveSparkApplicationEntityId, x.ApplicationId, "application id")
	fs.StringVar(&x.ApplicationSparkId, flags.FlagWaveSparkApplicationSparkId, x.ApplicationSparkId, "the application's spark id (spark-xxx)")
	fs.StringVar(&x.ApplicationState, flags.FlagWaveSparkApplicationState, x.ApplicationState, "application state")
	fs.StringVar(&x.ApplicationHeritage, flags.FlagWaveSparkApplicationHeritage, x.ApplicationHeritage, "application heritage")
}

func (x *CmdGetSparkApplicationOptions) Validate() error {
	// TODO Do this for list clusters too
	if x.ApplicationId != "" {
		if x.ClusterName != "" {
			return errors.RequiredXor(flags.FlagWaveSparkApplicationEntityId, flags.FlagWaveClusterName)
		}
		if x.ApplicationName != "" {
			return errors.RequiredXor(flags.FlagWaveSparkApplicationEntityId, flags.FlagWaveSparkApplicationName)
		}
		if x.ApplicationNamespace != "" {
			return errors.RequiredXor(flags.FlagWaveSparkApplicationEntityId, flags.FlagWaveSparkApplicationNamespace)
		}
		if x.ApplicationSparkId != "" {
			return errors.RequiredXor(flags.FlagWaveSparkApplicationEntityId, flags.FlagWaveSparkApplicationSparkId)
		}
		if x.ApplicationState != "" {
			return errors.RequiredXor(flags.FlagWaveSparkApplicationEntityId, flags.FlagWaveSparkApplicationState)
		}
		if x.ApplicationHeritage != "" {
			return errors.RequiredXor(flags.FlagWaveSparkApplicationEntityId, flags.FlagWaveSparkApplicationHeritage)
		}
	}
	return x.CmdGetOptions.Validate()
}
