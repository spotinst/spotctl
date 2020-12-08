package wave

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/spot"
	"github.com/spotinst/spotctl/internal/thirdparty/commands/eksctl"
	"github.com/spotinst/spotctl/internal/uuid"
	"github.com/spotinst/spotctl/internal/wave"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type CmdCreate struct {
	cmd  *cobra.Command
	opts CmdCreateOptions
}

type CmdCreateOptions struct {
	*CmdOptions
	ConfigFile        string
	ClusterID         string
	ClusterName       string
	Region            string
	Zones             []string
	Tags              []string
	KubernetesVersion string
}

func (x *CmdCreateOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&x.ConfigFile, flags.FlagWaveConfigFile, "f", x.ConfigFile, "load configuration from a file (or stdin if set to '-')")
	fs.StringVar(&x.ClusterID, flags.FlagWaveClusterID, x.ClusterID, "cluster id (will be created if empty)")
	fs.StringVar(&x.ClusterName, flags.FlagWaveClusterName, x.ClusterName, "cluster name (generated if unspecified, e.g. \"wave-9d4afe95\")")
	fs.StringVar(&x.Region, flags.FlagWaveRegion, os.Getenv("AWS_REGION"), "region in which your cluster (control plane and nodes) will be created")
	fs.StringSliceVar(&x.Zones, "zones", x.Zones, "availability zones in which your cluster (control plane and nodes) will be created")
	fs.StringSliceVar(&x.Tags, "tags", x.Tags, "list of K/V pairs used to tag all cloud resources (eg: \"Owner=john@example.com,Team=DevOps\")")
	fs.StringVar(&x.KubernetesVersion, "kubernetes-version", "1.18", "kubernetes version")
}

func NewCmdCreate(opts *CmdOptions) *cobra.Command {
	return newCmdCreate(opts).cmd
}

func newCmdCreate(opts *CmdOptions) *CmdCreate {
	var cmd CmdCreate

	cmd.cmd = &cobra.Command{
		Use:           "create",
		Short:         "Create a new Wave installation",
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
	if x.ClusterID == "" && x.ClusterName == "" && x.ConfigFile == "" {
		return errors.RequiredOr(flags.FlagWaveClusterID, flags.FlagWaveClusterName, flags.FlagWaveConfigFile)
	}
	if x.ClusterID != "" && x.ClusterName != "" {
		return errors.RequiredXor(flags.FlagWaveClusterID, flags.FlagWaveClusterName)
	}
	if x.ClusterID != "" && x.ConfigFile != "" {
		return errors.RequiredXor(flags.FlagWaveClusterID, flags.FlagWaveClusterName)
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
	if x.opts.ClusterID == "" { // create a new cluster
		if x.opts.ConfigFile != "" { // extract from config
			type clusterConfig struct {
				Metadata struct {
					Name   string `json:"name"`
					Region string `json:"region"`
				} `json:"metadata"`
			}

			var r io.Reader
			var err error

			if x.opts.ConfigFile == "-" { // read from standard input
				r = os.Stdin
			} else { // read from file
				f, err := os.Open(x.opts.ConfigFile)
				if err != nil {
					return err
				}
				defer f.Close()
				r = f
			}

			c := new(clusterConfig)
			if err = yaml.NewYAMLOrJSONDecoder(r, 4096).Decode(c); err != nil {
				return err
			}

			x.opts.ClusterName = c.Metadata.Name
			x.opts.Region = c.Metadata.Region
		} else { // generate unique name
			x.opts.ClusterName = fmt.Sprintf("wave-%s", uuid.NewV4().Short())
		}

		// TODO(liran): Validate it somewhere else.
		if x.opts.ClusterName == "" {
			return errors.Required("ClusterName")
		}

		cmdEksctl, err := x.opts.Clientset.NewCommand(eksctl.CommandName)
		if err != nil {
			return err
		}

		if err = cmdEksctl.Run(ctx, x.buildEksctlArgs()...); err != nil {
			return err
		}

	} else { // import an existing cluster
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

	log.Infof("Verified cluster %q", x.opts.ClusterName)

	// TODO(liran): Validate it somewhere else.
	if x.opts.Region == "" {
		return errors.Required("Region")
	}

	// Instantiate a cloud provider instance.
	cloudProviderOpts := []cloud.ProviderOption{
		cloud.WithProfile(x.opts.Profile),
		cloud.WithRegion(x.opts.Region),
	}
	cloudProvider, err := x.opts.Clientset.NewCloud(
		cloud.ProviderName(x.opts.CloudProvider), cloudProviderOpts...)
	if err != nil {
		return err
	}

	// Describe cluster instances.
	var instances []*cloud.Instance
	filters := []*cloud.Filter{
		{
			Name:   "tag:alpha.eksctl.io/cluster-name",
			Values: []string{x.opts.ClusterName},
		},
	}
	err = wait.Poll(5*time.Second, x.opts.Timeout, func() (bool, error) {
		instances, err = cloudProvider.Compute().DescribeInstances(ctx, filters...)
		if err != nil {
			return false, err
		}
		return len(instances) > 0, nil
	})
	for _, i := range instances {
		profile, err := cloudProvider.IAM().GetInstanceProfile(ctx, i.InstanceProfile.Name)
		if err != nil {
			return err
		}
		if len(profile.Roles) > 0 {
			const s3FullAccessProfileARN = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
			for _, role := range profile.Roles {
				if err = cloudProvider.IAM().AttachRolePolicy(ctx, role.Name,
					s3FullAccessProfileARN); err != nil {
					return err
				}
			}
		}
	}

	manager, err := wave.NewManager(x.opts.ClusterName, getWaveLogger()) // pass in name to validate ocean controller configuration
	if err != nil {
		return err
	}

	return manager.Create()
}

func (x *CmdCreate) buildEksctlArgs() []string {
	log.Debugf("Building up command arguments")

	args := []string{
		"create", "cluster",
		"--timeout", "60m",
		"--color", "false",
	}

	if len(x.opts.ConfigFile) > 0 {
		args = append(args, "--config-file", x.opts.ConfigFile)

	} else {
		if len(x.opts.ClusterName) > 0 {
			args = append(args, "--name", x.opts.ClusterName, "--nodegroup-name", x.opts.ClusterName)
		}

		if len(x.opts.Region) > 0 {
			args = append(args, "--region", x.opts.Region)
		}

		if len(x.opts.Zones) > 0 {
			args = append(args, "--zones", strings.Join(x.opts.Zones, ","))
		}

		if len(x.opts.Tags) > 0 {
			args = append(args, "--tags", strings.Join(x.opts.Tags, ","))
		}

		if len(x.opts.KubernetesVersion) > 0 {
			args = append(args, "--version", x.opts.KubernetesVersion)
		}

		args = append(args, "--spot-ocean", "--spot-profile", x.opts.Profile)
	}

	if x.opts.Verbose {
		args = append(args, "--verbose", "4")
	} else {
		args = append(args, "--verbose", "0")
	}

	return args
}
