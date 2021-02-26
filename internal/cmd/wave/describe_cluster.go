package wave

import (
	"context"
	"fmt"
	"github.com/spotinst/spotctl/internal/writer/writers/json"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/wave-operator/api/v1alpha1"
	"text/tabwriter"

	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
)

type (
	CmdDescribeCluster struct {
		cmd  *cobra.Command
		opts CmdDescribeClusterOptions
	}

	CmdDescribeClusterOptions struct {
		*CmdDescribeOptions
		ClusterID string
	}
)

func NewCmdDescribeCluster(opts *CmdDescribeOptions) *cobra.Command {
	return newCmdDescribeCluster(opts).cmd
}

func newCmdDescribeCluster(opts *CmdDescribeOptions) *CmdDescribeCluster {
	var cmd CmdDescribeCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Describe a Wave cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)

	return &cmd
}

func (x *CmdDescribeClusterOptions) Init(fs *pflag.FlagSet, opts *CmdDescribeOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdDescribeClusterOptions) initDefaults(opts *CmdDescribeOptions) {
	x.CmdDescribeOptions = opts
}

func (x *CmdDescribeClusterOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdDescribeCluster) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}
	return nil
}

func (x *CmdDescribeClusterOptions) Validate() error {
	if x.ClusterID == "" {
		return errors.Required(flags.FlagWaveClusterID)
	}
	return x.CmdDescribeOptions.Validate()
}

func (x *CmdDescribeCluster) Run(ctx context.Context) error {
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

func (x *CmdDescribeCluster) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDescribeCluster) validate(ctx context.Context) error {
	return x.opts.Validate()
}

// TODO describe_cluster describe_components
func (x *CmdDescribeCluster) run(ctx context.Context) error {
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

	w, err := x.opts.Clientset.NewWriter(json.WriterFormat)
	if err != nil {
		return err
	}

	return w.Write(cluster.Obj)

	/*if x.opts.ClusterID != "" {
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

	if err := wave.ValidateClusterContext(x.opts.ClusterName); err != nil {
		return fmt.Errorf("cluster context validation failure, %w", err)
	}

	waveComponents, err := wave.ListComponents()
	if err != nil {
		return err
	}

	return printWaveComponentDescriptions(waveComponents)*/
}

func printWaveComponentDescriptions(components *v1alpha1.WaveComponentList) error {

	width := 20
	writer := tabwriter.NewWriter(os.Stdout, width, 8, 1, '\t', tabwriter.AlignRight)
	bar := strings.Repeat("-", width)
	boundary := bar + "\t" + bar + "\t" + bar + "\t" + bar

	_, err := fmt.Fprintln(writer, "component\tcondition\tproperty\tvalue")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, boundary)
	if err != nil {
		return err
	}

	for _, wc := range components.Items {
		sort.Slice(wc.Status.Conditions, func(i, j int) bool {
			return wc.Status.Conditions[i].LastUpdateTime.Time.After(wc.Status.Conditions[j].LastUpdateTime.Time)
		})
		condition := "Unknown"

		if len(wc.Status.Conditions) > 0 {
			condition = fmt.Sprintf("%s=%s", wc.Status.Conditions[0].Type, wc.Status.Conditions[0].Status)
		}

		if len(wc.Status.Properties) == 0 {
			_, err = fmt.Fprintln(writer, wc.Name+"\t"+condition+"\t\t")
			if err != nil {
				return err
			}
		} else {
			h := wc.Name + "\t" + condition
			for k, v := range wc.Status.Properties {
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
