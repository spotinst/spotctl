package ocean

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spotinst/spotctl/internal/thirdparty/commands/kops"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/cloud/providers/aws"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/survey"
	"github.com/spotinst/spotctl/internal/uuid"
)

type (
	CmdQuickstartClusterKubernetesAWS struct {
		cmd  *cobra.Command
		opts CmdQuickstartClusterKubernetesAWSOptions
	}

	CmdQuickstartClusterKubernetesAWSOptions struct {
		*CmdQuickstartClusterKubernetesOptions

		// Basic
		ClusterName string
		StateStore  string

		// Networking
		Region  string
		Zones   []string
		VPC     string
		Subnets []string

		// Infrastructure
		MasterCount        int64
		NodeCount          int64
		MasterMachineTypes []string
		NodeMachineTypes   []string
		Image              string
		KubernetesVersion  string
		SSHPublicKey       string

		// Security
		Authorization string

		// Metadata
		Tags []string
	}
)

func NewCmdQuickstartClusterKubernetesAWS(opts *CmdQuickstartClusterKubernetesOptions) *cobra.Command {
	return newCmdQuickstartClusterKubernetesAWS(opts).cmd
}

func newCmdQuickstartClusterKubernetesAWS(opts *CmdQuickstartClusterKubernetesOptions) *CmdQuickstartClusterKubernetesAWS {
	var cmd CmdQuickstartClusterKubernetesAWS

	cmd.cmd = &cobra.Command{
		Use:           "aws",
		Short:         "Quickstart a new Ocean cluster on AWS (using kops)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdQuickstartClusterKubernetesAWS) Run(ctx context.Context) error {
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

func (x *CmdQuickstartClusterKubernetesAWS) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	log.Debugf("Starting survey...")
	surv, err := x.opts.Clientset.NewSurvey()
	if err != nil {
		return err
	}

	// Instantiate a cloud provider instance.
	cloudProviderOpts := []cloud.ProviderOption{
		cloud.WithProfile(x.opts.Profile),
		cloud.WithRegion(x.opts.Region),
	}
	cloudProvider, err := x.opts.Clientset.NewCloud(aws.CloudProviderName, cloudProviderOpts...)
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
				Default:  x.opts.ClusterName,
				Required: true,
			}

			if x.opts.ClusterName, err = surv.InputString(input); err != nil {
				return err
			}
		}
	}

	// Networking.
	{
		// Region.
		{
			if x.opts.Region == "" {
				regions, err := cloudProvider.Compute().DescribeRegions(ctx)
				if err != nil {
					return err
				}

				regionOpts := make([]interface{}, len(regions))
				for i, region := range regions {
					regionOpts[i] = region.Name
				}

				input := &survey.Select{
					Message: "Region",
					Help: "The region in which your cluster nodes (control plane" +
						" and nodes) will be created",
					Options: regionOpts,
				}

				if x.opts.Region, err = surv.Select(input); err != nil {
					return err
				}

				// Instantiate a cloud provider instance.
				cloudProviderOpts := []cloud.ProviderOption{
					cloud.WithProfile(x.opts.Profile),
					cloud.WithRegion(x.opts.Region),
				}
				cloudProvider, err = x.opts.Clientset.NewCloud(aws.CloudProviderName, cloudProviderOpts...)
				if err != nil {
					return err
				}
			}
		}

		// Availability zones.
		{
			if len(x.opts.Zones) == 0 {
				zones, err := cloudProvider.Compute().DescribeZones(ctx)
				if err != nil {
					return err
				}

				zoneOpts := make([]interface{}, len(zones))
				for i, zone := range zones {
					zoneOpts[i] = zone.Name
				}

				input := &survey.Select{
					Message: "Availability zones",
					Help: "The availability zones in which your cluster nodes (control plane" +
						" and nodes) will be created",
					Options: zoneOpts,
				}

				if x.opts.Zones, err = surv.SelectMulti(input); err != nil {
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
			log.Debugf("Skipping advanced configuration because user selection")
			return nil
		}

		// KOPS specific.
		{
			// State.
			{
				input := &survey.Input{
					Message:  "State store",
					Help:     "See: https://git.io/fjH5V",
					Default:  x.opts.StateStore,
					Required: true,
				}

				if x.opts.StateStore, err = surv.InputString(input); err != nil {
					return err
				}
			}
		}

		// Node count.
		{
			// Masters.
			{
				input := &survey.Input{
					Message:  "Number of master nodes",
					Default:  x.opts.MasterCount,
					Required: true,
					Validate: survey.ValidateInt64,
				}

				if x.opts.MasterCount, err = surv.InputInt64(input); err != nil {
					return err
				}
			}

			// Nodes.
			{
				input := &survey.Input{
					Message:  "Number of worker nodes",
					Default:  x.opts.NodeCount,
					Required: true,
					Validate: survey.ValidateInt64,
				}

				if x.opts.NodeCount, err = surv.InputInt64(input); err != nil {
					return err
				}
			}
		}

		// Networking.
		{
			// VPC.
			{
				if x.opts.VPC == "" {
					var selectVPC bool

					input := &survey.Input{
						Message: "Launch into shared VPC",
					}

					if selectVPC, err = surv.Confirm(input); err != nil {
						return err
					}

					if selectVPC {
						vpcs, err := cloudProvider.Compute().DescribeVPCs(ctx)
						if err != nil {
							return err
						}

						vpcOpts := make([]interface{}, len(vpcs))
						for i, vpc := range vpcs {
							vpcOpts[i] = fmt.Sprintf("%s (%s)", vpc.ID, vpc.Name)
						}

						input := &survey.Select{
							Message: "VPC",
							Help: "The VPC in which your cluster nodes (control plane" +
								" and nodes) will be created",
							Options:   vpcOpts,
							Transform: survey.TransformOnlyId,
						}

						if x.opts.VPC, err = surv.Select(input); err != nil {
							return err
						}
					}
				}
			}

			// Subnets.
			{
				if len(x.opts.Subnets) == 0 {
					var selectSubnets bool

					input := &survey.Input{
						Message: "Launch into shared subnets",
					}

					if selectSubnets, err = surv.Confirm(input); err != nil {
						return err
					}

					if selectSubnets {
						subnets, err := cloudProvider.Compute().DescribeSubnets(ctx, x.opts.VPC)
						if err != nil {
							return err
						}

						subnetOpts := make([]interface{}, len(subnets))
						for i, subnet := range subnets {
							subnetOpts[i] = fmt.Sprintf("%s (%s)", subnet.ID, subnet.Name)
						}

						input := &survey.Select{
							Message: "Subnets",
							Help: "The subnets in which your cluster nodes (control plane" +
								" and nodes) will be created",
							Options:   subnetOpts,
							Transform: survey.TransformOnlyId,
						}

						if x.opts.Subnets, err = surv.SelectMulti(input); err != nil {
							return err
						}
					}
				}
			}
		}

		// Kubernetes version.
		{
			input := &survey.Input{
				Message: "Kubernetes version",
				Help:    "Version of Kubernetes to run (defaults to the version in the stable channel)",
			}

			if x.opts.KubernetesVersion, err = surv.InputString(input); err != nil {
				return err
			}
		}

		// Security.
		{
			// Authorization.
			{
				input := &survey.Select{
					Message: "Authorization mode",
					Help:    "See: https://bit.ly/31xUE3h",
					Options: []interface{}{
						"RBAC",
						"AlwaysAllow",
					},
					Defaults: []interface{}{
						"RBAC",
					},
				}

				if x.opts.Authorization, err = surv.Select(input); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (x *CmdQuickstartClusterKubernetesAWS) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdQuickstartClusterKubernetesAWS) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdQuickstartClusterKubernetesAWS) run(ctx context.Context) error {
	// Instantiate a cloud provider instance.
	cloudProviderOpts := []cloud.ProviderOption{
		cloud.WithProfile(x.opts.Profile),
		cloud.WithRegion(x.opts.Region),
	}
	cloudProvider, err := x.opts.Clientset.NewCloud(aws.CloudProviderName, cloudProviderOpts...)
	if err != nil {
		return err
	}

	// Create an S3 bucket to store the cluster state.
	if _, err = cloudProvider.Storage().CreateBucket(ctx, x.opts.StateStore); err != nil {
		return err
	}

	// Instantiate new command.
	cmd, err := x.opts.Clientset.NewCommand(kops.CommandName)
	if err != nil {
		return err
	}

	// Build cluster configuration.
	cmdArgs := x.buildKopsArgs()

	// Finally, create the cluster.
	return cmd.Run(ctx, cmdArgs...)
}

func (x *CmdQuickstartClusterKubernetesAWS) buildKopsArgs() []string {
	log.Debugf("Building up command arguments")

	args := []string{
		"create", "cluster",
		"--state", x.opts.StateStore,
		"--name", x.opts.ClusterName,
		"--cloud", string(aws.CloudProviderName),
		"--master-count", fmt.Sprintf("%d", x.opts.MasterCount),
		"--node-count", fmt.Sprintf("%d", x.opts.NodeCount),
	}

	if len(x.opts.MasterMachineTypes) > 0 {
		args = append(args, "--master-size", strings.Join(x.opts.MasterMachineTypes, ","))
	}

	if len(x.opts.NodeMachineTypes) > 0 {
		args = append(args, "--node-size", strings.Join(x.opts.NodeMachineTypes, ","))
	}

	if len(x.opts.VPC) > 0 {
		args = append(args, "--vpc", x.opts.VPC)
	}

	if len(x.opts.Subnets) > 0 {
		args = append(args, "--subnets", strings.Join(x.opts.Subnets, ","))
	}

	if len(x.opts.Zones) > 0 {
		args = append(args, "--zones", strings.Join(x.opts.Zones, ","))
	}

	if len(x.opts.KubernetesVersion) > 0 {
		args = append(args, "--kubernetes-version", x.opts.KubernetesVersion)
	}

	if len(x.opts.Image) > 0 {
		args = append(args, "--image", x.opts.Image)
	}

	if len(x.opts.Tags) > 0 {
		args = append(args, "--cloud-labels", strings.Join(x.opts.Tags, ","))
	}

	if len(x.opts.SSHPublicKey) > 0 {
		args = append(args, "--ssh-public-key", x.opts.SSHPublicKey)
	}

	if len(x.opts.Authorization) > 0 {
		args = append(args, "--authorization", x.opts.Authorization)
	}

	if x.opts.DryRun {
		args = append(args, "--dry-run", "--output", "yaml")
	} else {
		args = append(args, "--yes")
	}

	if x.opts.Verbose {
		args = append(args, "--logtostderr", "--v", "10")
	} else {
		args = append(args, "--logtostderr", "--v", "0")
	}

	return args
}

func (x *CmdQuickstartClusterKubernetesAWSOptions) Init(fs *pflag.FlagSet, opts *CmdQuickstartClusterKubernetesOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdQuickstartClusterKubernetesAWSOptions) initDefaults(opts *CmdQuickstartClusterKubernetesOptions) {
	x.CmdQuickstartClusterKubernetesOptions = opts
	x.MasterCount = 3
	x.NodeCount = 3
	x.Authorization = "RBAC"
	x.SSHPublicKey = "~/.ssh/id_rsa.pub"
	x.Region = os.Getenv("AWS_DEFAULT_REGION")
	x.ClusterName = os.Getenv("KOPS_CLUSTER_NAME")

	x.StateStore = os.Getenv("KOPS_STATE_STORE")
	if x.StateStore == "" {
		// TODO(liran): Clean up after cluster deletion.
		x.StateStore = fmt.Sprintf("s3://spot-ocean-quickstart-%s", uuid.NewV4().Short())
	}
}

func (x *CmdQuickstartClusterKubernetesAWSOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterName, "cluster-name", x.ClusterName, "name of the cluster")
	fs.Int64Var(&x.MasterCount, "master-count", x.MasterCount, "master count")
	fs.Int64Var(&x.NodeCount, "node-count", x.NodeCount, "node count")
	fs.StringVar(&x.Region, "region", x.Region, "region in which your cluster (control plane and nodes) will be created")
	fs.StringVar(&x.VPC, "vpc", x.VPC, "region in which your cluster (control plane and nodes) will be created")
	fs.StringSliceVar(&x.Zones, "zones", x.Zones, "availability zones in which your cluster (control plane and nodes) will be created")
	fs.StringSliceVar(&x.MasterMachineTypes, "master-machine-types", x.MasterMachineTypes, "list of machine types for masters")
	fs.StringSliceVar(&x.NodeMachineTypes, "node-machine-types", x.NodeMachineTypes, "list of machine types for nodes")
	fs.StringVar(&x.StateStore, "state", x.StateStore, "s3 bucket used to store the state of the cluster")
	fs.StringVar(&x.SSHPublicKey, "ssh-public-key", x.SSHPublicKey, "ssh public key to use for nodes")
	fs.StringSliceVar(&x.Tags, "tags", x.Tags, "list of K/V pairs used to tag all cloud resources (eg: \"Owner=john@example.com,Team=DevOps\")")
	fs.StringVar(&x.Authorization, "authorization", x.Authorization, "authorization mode to use")
	fs.StringVar(&x.Image, "image", x.Image, "image to use in your cluster (control plane and nodes)")
	fs.StringVar(&x.KubernetesVersion, "kubernetes-version", x.KubernetesVersion, "kubernetes version")
}

func (x *CmdQuickstartClusterKubernetesAWSOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdQuickstartClusterKubernetesOptions.Validate(); err != nil {
		errg.Add(err)
	}

	if x.StateStore == "" {
		errg.Add(errors.Required("StateStore"))
	}

	if x.ClusterName == "" {
		errg.Add(errors.Required("ClusterName"))
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}
