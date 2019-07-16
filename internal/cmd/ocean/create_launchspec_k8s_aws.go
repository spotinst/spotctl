package ocean

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/cloud/providers/aws"
	"github.com/spotinst/spotinst-cli/internal/errors"
	"github.com/spotinst/spotinst-cli/internal/log"
	"github.com/spotinst/spotinst-cli/internal/survey"
	"github.com/spotinst/spotinst-cli/internal/thirdparty/commands/kops"
)

type (
	CmdCreateLaunchSpecKubernetesAWS struct {
		cmd  *cobra.Command
		opts CmdCreateLaunchSpecKubernetesAWSOptions
	}

	CmdCreateLaunchSpecKubernetesAWSOptions struct {
		*CmdCreateLaunchSpecKubernetesOptions

		// Basic
		ClusterName string
		SpecName    string
		State       string

		// Networking
		Region string
		Zones  []string

		// Infrastructure
		MinNodeCount int64
		MaxNodeCount int64
		MachineTypes []string

		// Metadata
		Labels []string
		Tags   []string
	}
)

func NewCmdCreateLaunchSpecKubernetesAWS(opts *CmdCreateLaunchSpecKubernetesOptions) *cobra.Command {
	return newCmdCreateLaunchSpecKubernetesAWS(opts).cmd
}

func newCmdCreateLaunchSpecKubernetesAWS(opts *CmdCreateLaunchSpecKubernetesOptions) *CmdCreateLaunchSpecKubernetesAWS {
	var cmd CmdCreateLaunchSpecKubernetesAWS

	cmd.cmd = &cobra.Command{
		Use:           "aws",
		Short:         "Create a new Ocean launch spec on AWS (using kops)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdCreateLaunchSpecKubernetesAWS) Run(ctx context.Context) error {
	steps := []func(context.Context) error{x.survey, x.validate, x.run}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (x *CmdCreateLaunchSpecKubernetesAWS) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	log.Debugf("Starting survey...")
	surv, err := x.opts.Clients.NewSurvey()
	if err != nil {
		return err
	}

	// Instantiate a cloud provider instance.
	cloudProvider, err := x.opts.Clients.NewCloud(aws.CloudProviderName)
	if err != nil {
		return err
	}

	// Cluster name.
	{
		if x.opts.ClusterName == "" {
			input := &survey.Input{
				Message: "Cluster name",
				Help: "Name must start with a lowercase letter followed by up to " +
					"39 lowercase letters, numbers, or hyphens, and cannot end with a hyphen",
				Required: true,
			}

			if x.opts.ClusterName, err = surv.InputString(input); err != nil {
				return err
			}
		}
	}

	// Spec name.
	{
		if x.opts.SpecName == "" {
			input := &survey.Input{
				Message: "Spec name",
				Help: "Name must start with a lowercase letter followed by up to " +
					"39 lowercase letters, numbers, or hyphens, and cannot end with a hyphen",
				Required: true,
			}

			if x.opts.SpecName, err = surv.InputString(input); err != nil {
				return err
			}
		}
	}

	// Networking.
	{
		// Region.
		{
			if x.opts.Region == "" {
				regions, err := cloudProvider.DescribeRegions()
				if err != nil {
					return err
				}

				regionOpts := make([]interface{}, len(regions))
				for i, region := range regions {
					regionOpts[i] = region.Name
				}

				input := &survey.Select{
					Message: "Region",
					Help:    "The region in which your cluster is located",
					Options: regionOpts,
				}

				if x.opts.Region, err = surv.Select(input); err != nil {
					return err
				}
			}
		}

		// Availability zones.
		{
			if len(x.opts.Zones) == 0 {
				zones, err := cloudProvider.DescribeZones(x.opts.Region)
				if err != nil {
					return err
				}

				zoneOpts := make([]interface{}, len(zones))
				for i, zone := range zones {
					zoneOpts[i] = zone.Name
				}

				input := &survey.Select{
					Message: "Availability zones",
					Help:    "The availability zones in which your nodes will be created",
					Options: zoneOpts,
				}

				if x.opts.Zones, err = surv.SelectMulti(input); err != nil {
					return err
				}
			}
		}
	}

	// KOPS specific.
	{
		// State.
		{
			if x.opts.State == "" {
				input := &survey.Input{
					Message:  "Location of state store",
					Help:     "See: https://git.io/fjH5V",
					Required: true,
				}

				if x.opts.State, err = surv.InputString(input); err != nil {
					return err
				}
			}
		}
	}

	// Advanced.
	{
		input := &survey.Input{
			Message: "Edit advanced configuration",
		}

		if x.opts.Advanced, err = surv.Confirm(input); err != nil {
			return err
		}

		if !x.opts.Advanced {
			return nil
		}

		// Node count.
		{
			// Minimum.
			{
				if x.opts.MinNodeCount == 0 {
					input := &survey.Input{
						Message:  "Minimum node count",
						Default:  "0",
						Required: true,
						Validate: survey.ValidateInt64,
					}

					if x.opts.MinNodeCount, err = surv.InputInt64(input); err != nil {
						return err
					}
				}
			}

			// Maximum.
			{
				if x.opts.MaxNodeCount == 0 {
					input := &survey.Input{
						Message:  "Maximum node count",
						Default:  "3",
						Required: true,
						Validate: survey.ValidateInt64,
					}

					if x.opts.MaxNodeCount, err = surv.InputInt64(input); err != nil {
						return err
					}
				}
			}
		}

		// Labels.
		{
			input := &survey.Input{
				Message: "Labels",
				Help:    "List of K/V pairs used to label all nodes (eg: \"spotinst.io/foo=bar,spotinst.io/alice=bob\")",
			}

			labels, err := surv.InputString(input)
			if err != nil {
				return err
			}

			x.opts.Labels = strings.Split(labels, " ")
		}

		// Tags.
		{
			input := &survey.Input{
				Message: "Tags",
				Help:    "List of K/V pairs used to tag all cloud resources (eg: \"Owner=john@example.com,Team=DevOps\")",
			}

			tags, err := surv.InputString(input)
			if err != nil {
				return err
			}

			x.opts.Tags = strings.Split(tags, " ")
		}
	}

	return nil
}

func (x *CmdCreateLaunchSpecKubernetesAWS) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdCreateLaunchSpecKubernetesAWS) run(ctx context.Context) error {
	cmd, err := x.opts.Clients.NewCommand(kops.CommandName)
	if err != nil {
		return err
	}

	return cmd.Run(ctx, x.buildKopsArgs()...)
}

func (x *CmdCreateLaunchSpecKubernetesAWS) buildKopsArgs() []string {
	log.Debugf("Building up command arguments")

	args := []string{
		"create",
		"--state", x.opts.State,
		"--filename", "",
	}

	if x.opts.Verbose {
		args = append(args, "--logtostderr", "--v", "10")
	} else {
		args = append(args, "--logtostderr", "--v", "0")
	}

	return args
}

func (x *CmdCreateLaunchSpecKubernetesAWSOptions) Init(flags *pflag.FlagSet, opts *CmdCreateLaunchSpecKubernetesOptions) {
	x.initDefaults(opts)
	x.initFlags(flags)
}

func (x *CmdCreateLaunchSpecKubernetesAWSOptions) initDefaults(opts *CmdCreateLaunchSpecKubernetesOptions) {
	x.CmdCreateLaunchSpecKubernetesOptions = opts
	x.ClusterName = os.Getenv("KOPS_CLUSTER_NAME")
	x.State = os.Getenv("KOPS_STATE_STORE")
	x.MaxNodeCount = 3
}

func (x *CmdCreateLaunchSpecKubernetesAWSOptions) initFlags(flags *pflag.FlagSet) {
	flags.StringVar(
		&x.ClusterName,
		"cluster-name",
		x.ClusterName,
		"name of the cluster")

	flags.StringVar(
		&x.SpecName,
		"spec-name",
		x.SpecName,
		"name of the launchspec")

	flags.Int64Var(
		&x.MinNodeCount,
		"min-node-count",
		x.MinNodeCount,
		"minimum node count")

	flags.Int64Var(
		&x.MaxNodeCount,
		"max-node-count",
		x.MaxNodeCount,
		"maximum node count")

	flags.StringSliceVar(
		&x.Zones,
		"zones",
		x.Zones,
		"availability zones in which your nodes will be created")

	flags.StringSliceVar(
		&x.MachineTypes,
		"machine-types",
		x.MachineTypes,
		"list of machine types")

	flags.StringVar(
		&x.State,
		"state",
		x.State,
		"s3 bucket used to store the state of the cluster")

	flags.StringSliceVar(
		&x.Labels,
		"labels",
		x.Labels,
		"list of K/V pairs used to label all nodes (eg: \"spotinst.io/foo=bar,spotinst.io/alice=bob\")")

	flags.StringSliceVar(
		&x.Tags,
		"tags",
		x.Tags,
		"list of K/V pairs used to tag all cloud resources (eg: \"Owner=john@example.com,Team=DevOps\")")

}

func (x *CmdCreateLaunchSpecKubernetesAWSOptions) Validate() error {
	if err := x.CmdCreateLaunchSpecKubernetesOptions.Validate(); err != nil {
		return err
	}

	if x.State == "" {
		return errors.Required("state")
	}

	if x.ClusterName == "" {
		return errors.Required("cluster-name")
	}

	return nil
}
