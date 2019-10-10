package ocean

import (
	"context"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/spotinst"
	"github.com/spotinst/spotinst-cli/internal/utils/flags"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	spotinstsdk "github.com/spotinst/spotinst-sdk-go/spotinst"
)

type (
	CmdUpdateClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdUpdateClusterKubernetesOptions
	}

	CmdUpdateClusterKubernetesOptions struct {
		*CmdUpdateClusterOptions

		// Base
		ClusterID string
		Name      string

		// Strategy
		SpotPercentage           float64
		UtilizeReservedInstances bool
		FallbackToOnDemand       bool

		// Capacity
		MinSize    int
		MaxSize    int
		TargetSize int

		// Compute
		SubnetIDs                []string
		InstanceTypesWhitelist   []string
		InstanceTypesBlacklist   []string
		SecurityGroupIDs         []string
		ImageID                  string
		KeyPair                  string
		UserData                 string
		RootVolumeSize           int
		AssociatePublicIPAddress bool
		EnableMonitoring         bool
		EnableEBSOptimization    bool

		// Auto Scaling
		EnableAutoScaler bool
		EnableAutoConfig bool
		Cooldown         int
	}
)

func NewCmdUpdateClusterKubernetes(opts *CmdUpdateClusterOptions) *cobra.Command {
	return newCmdUpdateClusterKubernetes(opts).cmd
}

func newCmdUpdateClusterKubernetes(opts *CmdUpdateClusterOptions) *CmdUpdateClusterKubernetes {
	var cmd CmdUpdateClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Update an existing Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdUpdateClusterKubernetes) Run(ctx context.Context) error {
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

func (x *CmdUpdateClusterKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdUpdateClusterKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdUpdateClusterKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdUpdateClusterKubernetes) run(ctx context.Context) error {
	spotinstClientOpts := []spotinst.ClientOption{
		spotinst.WithCredentialsProfile(x.opts.Profile),
	}

	spotinstClient, err := x.opts.Clients.NewSpotinst(spotinstClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotinstClient.Services().Ocean(x.opts.CloudProvider, spotinst.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	cluster, ok := x.buildClusterFromOpts()
	if !ok {
		fmt.Fprintln(x.opts.Out, "Update cancelled, no changes made.")
		return nil
	}

	_, err = oceanClient.UpdateCluster(ctx, cluster)
	return err
}

func (x *CmdUpdateClusterKubernetes) buildClusterFromOpts() (*spotinst.OceanCluster, bool) {
	var cluster interface{}
	var changed bool

	switch x.opts.CloudProvider {
	case spotinst.CloudProviderAWS:
		cluster, changed = x.buildClusterFromOptsAWS()
	}

	return &spotinst.OceanCluster{Obj: cluster}, changed
}

func (x *CmdUpdateClusterKubernetes) buildClusterFromOptsAWS() (*aws.Cluster, bool) {
	cluster := new(aws.Cluster)
	changed := false

	if x.opts.Name != "" {
		cluster.SetName(spotinstsdk.String(x.opts.Name))
	}

	if changed = !reflect.DeepEqual(cluster, new(aws.Cluster)); changed {
		cluster.SetId(spotinstsdk.String(x.opts.ClusterID))
	}

	return cluster, changed
}

func (x *CmdUpdateClusterKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdUpdateClusterOptions) {
	x.initDefaults(opts)
	x.initFlags(flags)
}

func (x *CmdUpdateClusterKubernetesOptions) initDefaults(opts *CmdUpdateClusterOptions) {
	x.CmdUpdateClusterOptions = opts
}

func (x *CmdUpdateClusterKubernetesOptions) initFlags(flags *pflag.FlagSet) {
	// Base
	{
		flags.StringVar(
			&x.ClusterID,
			"cluster-id",
			x.ClusterID,
			"id of the cluster")

		flags.StringVar(
			&x.Name,
			"name",
			x.Name,
			"name of the cluster")
	}

	// Strategy
	{
		flags.Float64Var(
			&x.SpotPercentage,
			"spot-percentage",
			x.SpotPercentage,
			"")

		flags.BoolVar(
			&x.UtilizeReservedInstances,
			"utilize-reserved-instances",
			x.UtilizeReservedInstances,
			"")

		flags.BoolVar(
			&x.FallbackToOnDemand,
			"fallback-ondemand",
			x.FallbackToOnDemand,
			"")
	}

	// Capacity
	{
		flags.IntVar(
			&x.MinSize,
			"min-size",
			x.MinSize,
			"")

		flags.IntVar(
			&x.MaxSize,
			"max-size",
			x.MaxSize,
			"")

		flags.IntVar(
			&x.TargetSize,
			"target-size",
			x.TargetSize,
			"")
	}

	// Compute
	{
		flags.StringSliceVar(
			&x.SubnetIDs,
			"subnet-ids",
			x.SubnetIDs,
			"")

		flags.StringSliceVar(
			&x.InstanceTypesWhitelist,
			"instance-types-whitelist",
			x.InstanceTypesWhitelist,
			"")

		flags.StringSliceVar(
			&x.InstanceTypesBlacklist,
			"instance-types-blacklist",
			x.InstanceTypesBlacklist,
			"")

		flags.StringSliceVar(
			&x.SecurityGroupIDs,
			"security-group-ids",
			x.SecurityGroupIDs,
			"")

		flags.StringVar(
			&x.ImageID,
			"image-id",
			x.ImageID,
			"")

		flags.StringVar(
			&x.KeyPair,
			"key-pair",
			x.KeyPair,
			"")

		flags.StringVar(
			&x.UserData,
			"user-data",
			x.UserData,
			"")

		flags.IntVar(
			&x.RootVolumeSize,
			"root-volume-size",
			x.RootVolumeSize,
			"")

		flags.BoolVar(
			&x.AssociatePublicIPAddress,
			"associate-public-ip-address",
			x.AssociatePublicIPAddress,
			"")

		flags.BoolVar(
			&x.EnableMonitoring,
			"enable-monitoring",
			x.EnableMonitoring,
			"")

		flags.BoolVar(
			&x.EnableEBSOptimization,
			"enable-ebs-optimization",
			x.EnableEBSOptimization,
			"")

	}

	// Auto Scaling
	{
		flags.BoolVar(
			&x.EnableAutoScaler,
			"enable-auto-scaler",
			x.EnableAutoScaler,
			"")

		flags.BoolVar(
			&x.EnableAutoConfig,
			"enable-auto-scaler-autoconfig",
			x.EnableAutoConfig,
			"")

		flags.IntVar(
			&x.Cooldown,
			"cooldown",
			x.Cooldown,
			"")
	}
}

func (x *CmdUpdateClusterKubernetesOptions) Validate() error {
	return x.CmdUpdateClusterOptions.Validate()
}
