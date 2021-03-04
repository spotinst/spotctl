package wave

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
)

type (
	CmdDescribeComponents struct {
		cmd  *cobra.Command
		opts CmdDescribeComponentsOptions
	}

	CmdDescribeComponentsOptions struct {
		*CmdDescribeOptions
		ClusterID string
	}
)

func NewCmdDescribeComponents(opts *CmdDescribeOptions) *cobra.Command {
	return newCmdDescribeComponents(opts).cmd
}

func newCmdDescribeComponents(opts *CmdDescribeOptions) *CmdDescribeComponents {
	var cmd CmdDescribeComponents

	cmd.cmd = &cobra.Command{
		Use:           "components",
		Short:         "Describe Wave components",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdDescribeComponentsOptions) Init(fs *pflag.FlagSet, opts *CmdDescribeOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdDescribeComponentsOptions) initDefaults(opts *CmdDescribeOptions) {
	x.CmdDescribeOptions = opts
}

func (x *CmdDescribeComponentsOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdDescribeComponents) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdDescribeComponentsOptions) Validate() error {
	if x.ClusterID == "" {
		return errors.Required(flags.FlagWaveClusterID)
	}
	return x.CmdDescribeOptions.Validate()
}

func (x *CmdDescribeComponents) Run(ctx context.Context) error {
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

func (x *CmdDescribeComponents) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDescribeComponents) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDescribeComponents) run(ctx context.Context) error {
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

	cluster, err := waveClient.GetCluster(ctx, x.opts.ClusterID)
	if err != nil {
		return err
	}

	return printWaveComponentDescriptions(cluster.Components)
}

func printWaveComponentDescriptions(components []spot.WaveComponent) error {

	width := 20
	writer := tabwriter.NewWriter(os.Stdout, width, 8, 1, '\t', tabwriter.AlignRight)
	bar := strings.Repeat("-", width)
	boundary := bar + "\t" + bar + "\t" + bar + "\t" + bar

	_, err := fmt.Fprintln(writer, "component\tstate\tproperty\tvalue")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, boundary)
	if err != nil {
		return err
	}

	for _, component := range components {
		if len(component.Properties) == 0 {
			_, err = fmt.Fprintln(writer, component.Name+"\t"+component.State+"\t\t")
			if err != nil {
				return err
			}
		} else {
			h := component.Name + "\t" + component.State
			for k, v := range component.Properties {
				_, err = fmt.Fprintln(writer, h+"\t"+k+"\t"+v)
				if err != nil {
					return err
				}
				h = "\t"
			}
		}

		_, err = fmt.Fprintln(writer, boundary)
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}
