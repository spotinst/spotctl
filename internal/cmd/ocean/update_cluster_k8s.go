package ocean

import (
	"context"
	"fmt"
	"github.com/spotinst/spotinst-cli/internal/errors"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/spotinst"
	"github.com/spotinst/spotinst-cli/internal/utils"
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
		ClusterID           string
		Name                string
		ControllerClusterId string

		// Strategy
		SpotPercentage           float64
		UtilizeReservedInstances bool
		FallbackToOnDemand       bool
		DrainingTimeout          int

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
		IAMInstanceProfileName   string
		IAMInstanceProfileArn    string
		LoadBalancerName         string
		LoadBalancerArn          string
		LoadBalancerType         string
		//TODO add slice of tags in spotinst sdk

		// Auto Scaling
		EnableAutoScaler       bool
		EnableAutoConfig       bool
		Cooldown               int
		HeadroomCpuPerUnit     int
		HeadroomMemoryPerUnit  int
		HeadroomGpuPerUnit     int
		HeadroomNumPerUnit     int
		ResourceLimitMaxVCpu   int
		ResourceLimitMaxMemory int
		EvaluationPeriods      int
		MaxScaleDownPercentage int
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

	if x.cmd.Flags().Changed(utils.Name) {
		cluster.SetName(spotinstsdk.String(x.opts.Name))
	}
	if x.cmd.Flags().Changed(utils.ControllerClusterId) {
		cluster.SetControllerClusterId(spotinstsdk.String(x.opts.ControllerClusterId))
	}

	cluster.SetStrategy(updateStrategy(x))
	cluster.SetCapacity(updateCapacity(x))
	cluster.SetAutoScaler(updateAutoScaler(x))
	cluster.SetCompute(updateCompute(x))

	if changed = !reflect.DeepEqual(cluster, createEmptyCluster()); changed {
		cluster.SetId(spotinstsdk.String(x.opts.ClusterID))
	}

	return cluster, changed
}

func updateStrategy(x *CmdUpdateClusterKubernetes) *aws.Strategy {
	var strategy *aws.Strategy

	if x.cmd.Flags().Changed(utils.SpotPercentage) || x.cmd.Flags().Changed(utils.UtilizeReserveInstances) ||
		x.cmd.Flags().Changed(utils.FallbackOnDemand) || x.cmd.Flags().Changed(utils.DrainingTimeout) {
		strategy = new(aws.Strategy)

		if x.cmd.Flags().Changed(utils.SpotPercentage) {
			strategy.SetSpotPercentage(spotinstsdk.Float64(x.opts.SpotPercentage))
		}
		if x.cmd.Flags().Changed(utils.UtilizeReserveInstances) {
			strategy.SetUtilizeReservedInstances(spotinstsdk.Bool(x.opts.UtilizeReservedInstances))
		}
		if x.cmd.Flags().Changed(utils.FallbackOnDemand) {
			strategy.SetFallbackToOnDemand(spotinstsdk.Bool(x.opts.FallbackToOnDemand))
		}
		if x.cmd.Flags().Changed(utils.DrainingTimeout) {
			strategy.SetDrainingTimeout(spotinstsdk.Int(x.opts.DrainingTimeout))
		}
	}

	return strategy
}

func updateCapacity(x *CmdUpdateClusterKubernetes) *aws.Capacity {
	var capacity *aws.Capacity

	if x.cmd.Flags().Changed(utils.MaxSize) || x.cmd.Flags().Changed(utils.MinSize) ||
		x.cmd.Flags().Changed(utils.TargetSize) {
		capacity = new(aws.Capacity)

		if x.cmd.Flags().Changed(utils.MinSize) {
			capacity.SetMinimum(spotinstsdk.Int(x.opts.MinSize))
		}
		if x.cmd.Flags().Changed(utils.MaxSize) {
			capacity.SetMaximum(spotinstsdk.Int(x.opts.MaxSize))
		}
		if x.cmd.Flags().Changed(utils.TargetSize) {
			capacity.SetTarget(spotinstsdk.Int(x.opts.TargetSize))
		}
	}

	return capacity
}

func updateAutoScaler(x *CmdUpdateClusterKubernetes) *aws.AutoScaler {
	autoScaler := new(aws.AutoScaler)

	if x.cmd.Flags().Changed(utils.EnableAutoScaler) {
		autoScaler.SetIsEnabled(spotinstsdk.Bool(x.opts.EnableAutoScaler))
	}
	if x.cmd.Flags().Changed(utils.EnableAutoScalerAutoconfig) {
		autoScaler.SetIsAutoConfig(spotinstsdk.Bool(x.opts.EnableAutoConfig))
	}
	if x.cmd.Flags().Changed(utils.Cooldown) {
		autoScaler.SetCooldown(spotinstsdk.Int(x.opts.Cooldown))
	}

	if x.cmd.Flags().Changed(utils.HeadroomCpuPerUnit) || x.cmd.Flags().Changed(utils.HeadroomMemoryPerUnit) ||
		x.cmd.Flags().Changed(utils.HeadroomGpuPerUnit) || x.cmd.Flags().Changed(utils.HeadroomNumPerUnit) {
		headroom := new(aws.AutoScalerHeadroom)

		if x.cmd.Flags().Changed(utils.HeadroomCpuPerUnit) {
			headroom.SetCPUPerUnit(spotinstsdk.Int(x.opts.HeadroomCpuPerUnit))
		}
		if x.cmd.Flags().Changed(utils.HeadroomMemoryPerUnit) {
			headroom.SetMemoryPerUnit(spotinstsdk.Int(x.opts.HeadroomMemoryPerUnit))
		}
		if x.cmd.Flags().Changed(utils.HeadroomGpuPerUnit) {
			headroom.SetGPUPerUnit(spotinstsdk.Int(x.opts.HeadroomGpuPerUnit))
		}
		if x.cmd.Flags().Changed(utils.HeadroomNumPerUnit) {
			headroom.SetNumOfUnits(spotinstsdk.Int(x.opts.HeadroomNumPerUnit))
		}

		autoScaler.SetHeadroom(headroom)
	}

	if x.cmd.Flags().Changed(utils.ResourceLimitMaxVcpu) || x.cmd.Flags().Changed(utils.ResourceLimitMaxMemory) {
		resourceLimit := new(aws.AutoScalerResourceLimits)

		if x.cmd.Flags().Changed(utils.ResourceLimitMaxMemory) {
			resourceLimit.SetMaxMemoryGiB(spotinstsdk.Int(x.opts.ResourceLimitMaxMemory))
		}
		if x.cmd.Flags().Changed(utils.ResourceLimitMaxVcpu) {
			resourceLimit.SetMaxVCPU(spotinstsdk.Int(x.opts.ResourceLimitMaxVCpu))
		}

		autoScaler.SetResourceLimits(resourceLimit)
	}

	//TODO add support to MaxScaleDownPercentage in spotinst sdk
	if x.cmd.Flags().Changed(utils.EvaluationPeriods) {
		down := new(aws.AutoScalerDown)
		down.SetEvaluationPeriods(spotinstsdk.Int(x.opts.EvaluationPeriods))

		autoScaler.SetDown(down)
	}

	return autoScaler
}

func updateCompute(x *CmdUpdateClusterKubernetes) *aws.Compute {
	compute := new(aws.Compute)

	//Subnet Ids
	if x.cmd.Flags().Changed(utils.SubnetIds) {
		compute.SetSubnetIDs(x.opts.SubnetIDs)
	}

	//Instances types
	if x.cmd.Flags().Changed(utils.InstancesTypesBlacklist) || x.cmd.Flags().Changed(utils.InstancesTypesWhitelist) {
		instanceTypes := new(aws.InstanceTypes)
		if x.cmd.Flags().Changed(utils.InstancesTypesBlacklist) {
			instanceTypes.SetBlacklist(x.opts.InstanceTypesBlacklist)
		}
		if x.cmd.Flags().Changed(utils.InstancesTypesWhitelist) {
			instanceTypes.SetWhitelist(x.opts.InstanceTypesWhitelist)
		}

		compute.SetInstanceTypes(instanceTypes)
	}

	//Launch specification
	shouldUpdateLaunchSpec := shouldUpdateLaunchSpec(x)
	if shouldUpdateLaunchSpec {
		launchSpec := new(aws.LaunchSpecification)

		if x.cmd.Flags().Changed(utils.AssociatePublicIpAddress) {
			launchSpec.SetAssociatePublicIPAddress(spotinstsdk.Bool(x.opts.AssociatePublicIPAddress))
		}
		if x.cmd.Flags().Changed(utils.SecurityGroupIds) {
			launchSpec.SetSecurityGroupIDs(x.opts.SecurityGroupIDs)
		}
		if x.cmd.Flags().Changed(utils.ImageIds) {
			launchSpec.SetImageId(spotinstsdk.String(x.opts.ImageID))
		}
		if x.cmd.Flags().Changed(utils.KeyPair) {
			launchSpec.SetKeyPair(spotinstsdk.String(x.opts.KeyPair))
		}
		if x.cmd.Flags().Changed(utils.UserData) {
			launchSpec.SetUserData(spotinstsdk.String(x.opts.UserData))
		}
		if x.cmd.Flags().Changed(utils.RootVolumeSize) {
			launchSpec.SetRootVolumeSize(spotinstsdk.Int(x.opts.RootVolumeSize))
		}
		if x.cmd.Flags().Changed(utils.EnableMonitoring) {
			launchSpec.SetMonitoring(spotinstsdk.Bool(x.opts.EnableMonitoring))
		}
		if x.cmd.Flags().Changed(utils.EnableEbsOptimization) {
			launchSpec.SetEBSOptimized(spotinstsdk.Bool(x.opts.EnableEBSOptimization))
		}

		if x.cmd.Flags().Changed(utils.IamInstanceProfileName) || x.cmd.Flags().Changed(utils.IamInstanceProfileArn) {
			iam := new(aws.IAMInstanceProfile)

			if x.cmd.Flags().Changed(utils.IamInstanceProfileName) {
				iam.SetName(spotinstsdk.String(x.opts.IAMInstanceProfileName))
			}
			if x.cmd.Flags().Changed(utils.IamInstanceProfileArn) {
				iam.SetArn(spotinstsdk.String(x.opts.IAMInstanceProfileArn))
			}
			launchSpec.SetIAMInstanceProfile(iam)
		}

		if x.cmd.Flags().Changed(utils.LoadBalancerName) || x.cmd.Flags().Changed(utils.LoadBalancerArn) || x.cmd.Flags().Changed(utils.LoadBalancerType) {
			loadBalancer := new(aws.LoadBalancer)

			if x.cmd.Flags().Changed(utils.LoadBalancerName) {
				loadBalancer.SetName(spotinstsdk.String(x.opts.LoadBalancerName))
			}
			if x.cmd.Flags().Changed(utils.LoadBalancerArn) {
				loadBalancer.SetArn(spotinstsdk.String(x.opts.LoadBalancerArn))
			}
			if x.cmd.Flags().Changed(utils.LoadBalancerType) {
				loadBalancer.SetType(spotinstsdk.String(x.opts.LoadBalancerType))
			}
			loadBalancers := []*aws.LoadBalancer{loadBalancer}

			launchSpec.SetLoadBalancers(loadBalancers)
		} //TODO add tags

		compute.SetLaunchSpecification(launchSpec)
	}

	return compute
}

func shouldUpdateLaunchSpec(x *CmdUpdateClusterKubernetes) bool {
	shouldUpdate := false

	shouldUpdate = x.cmd.Flags().Changed(utils.AssociatePublicIpAddress) || x.cmd.Flags().Changed(utils.SecurityGroupIds) ||
		x.cmd.Flags().Changed(utils.ImageIds) || x.cmd.Flags().Changed(utils.KeyPair) || x.cmd.Flags().Changed(utils.UserData) ||
		x.cmd.Flags().Changed(utils.RootVolumeSize) || x.cmd.Flags().Changed(utils.EnableMonitoring) ||
		x.cmd.Flags().Changed(utils.EnableEbsOptimization) || x.cmd.Flags().Changed(utils.IamInstanceProfileName) ||
		x.cmd.Flags().Changed(utils.IamInstanceProfileArn) || x.cmd.Flags().Changed(utils.LoadBalancerName) ||
		x.cmd.Flags().Changed(utils.LoadBalancerArn) || x.cmd.Flags().Changed(utils.LoadBalancerType)

	return shouldUpdate
}

func createEmptyCluster() *aws.Cluster {
	cluster := new(aws.Cluster)
	cluster.SetAutoScaler(new(aws.AutoScaler))
	cluster.SetCompute(new(aws.Compute))
	cluster.SetStrategy(nil)
	cluster.SetCapacity(nil)

	return cluster
}

//TODO add support to scheduling object in sdk-go
//func buildScheduling(x *CmdCreateClusterKubernetes) *aws.Scheduling {

//TODO add support to security object in sdk-go
//func buildSecurity(x *CmdCreateClusterKubernetes) *aws.Security

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
			utils.ClusterId,
			x.ClusterID,
			"id of the cluster")

		flags.StringVar(
			&x.ControllerClusterId,
			utils.ControllerClusterId,
			x.ControllerClusterId,
			"id of the cluster controller")

		flags.StringVar(
			&x.Name,
			utils.Name,
			x.Name,
			"name of the cluster")
	}

	// Strategy
	{
		flags.Float64Var(
			&x.SpotPercentage,
			utils.SpotPercentage,
			x.SpotPercentage,
			"")

		flags.IntVar(
			&x.DrainingTimeout,
			utils.DrainingTimeout,
			x.DrainingTimeout,
			"")

		flags.BoolVar(
			&x.UtilizeReservedInstances,
			utils.UtilizeReserveInstances,
			x.UtilizeReservedInstances,
			"")

		flags.BoolVar(
			&x.FallbackToOnDemand,
			utils.FallbackOnDemand,
			x.FallbackToOnDemand,
			"")
	}

	// Capacity
	{
		flags.IntVar(
			&x.MinSize,
			utils.MinSize,
			x.MinSize,
			"")

		flags.IntVar(
			&x.MaxSize,
			utils.MaxSize,
			x.MaxSize,
			"")

		flags.IntVar(
			&x.TargetSize,
			utils.TargetSize,
			x.TargetSize,
			"")
	}

	// Compute
	{
		flags.StringSliceVar(
			&x.SubnetIDs,
			utils.SubnetIds,
			x.SubnetIDs,
			"")

		flags.StringSliceVar(
			&x.InstanceTypesWhitelist,
			utils.InstancesTypesWhitelist,
			x.InstanceTypesWhitelist,
			"")

		flags.StringSliceVar(
			&x.InstanceTypesBlacklist,
			utils.InstancesTypesBlacklist,
			x.InstanceTypesBlacklist,
			"")

		flags.StringSliceVar(
			&x.SecurityGroupIDs,
			utils.SecurityGroupIds,
			x.SecurityGroupIDs,
			"")

		flags.StringVar(
			&x.ImageID,
			utils.ImageIds,
			x.ImageID,
			"")

		flags.StringVar(
			&x.KeyPair,
			utils.KeyPair,
			x.KeyPair,
			"")

		flags.StringVar(
			&x.UserData,
			utils.UserData,
			x.UserData,
			"")

		flags.IntVar(
			&x.RootVolumeSize,
			utils.RootVolumeSize,
			x.RootVolumeSize,
			"")

		flags.BoolVar(
			&x.AssociatePublicIPAddress,
			utils.AssociatePublicIpAddress,
			x.AssociatePublicIPAddress,
			"")

		flags.BoolVar(
			&x.EnableMonitoring,
			utils.EnableMonitoring,
			x.EnableMonitoring,
			"")

		flags.BoolVar(
			&x.EnableEBSOptimization,
			utils.EnableEbsOptimization,
			x.EnableEBSOptimization,
			"")

		flags.StringVar(
			&x.IAMInstanceProfileName,
			utils.IamInstanceProfileName,
			x.IAMInstanceProfileName,
			"")

		flags.StringVar(
			&x.IAMInstanceProfileArn,
			utils.IamInstanceProfileArn,
			x.IAMInstanceProfileArn,
			"")

		flags.StringVar(
			&x.LoadBalancerName,
			utils.LoadBalancerName,
			x.LoadBalancerName,
			"")

		flags.StringVar(
			&x.LoadBalancerArn,
			utils.LoadBalancerArn,
			x.LoadBalancerArn,
			"")

		flags.StringVar(
			&x.LoadBalancerType,
			utils.LoadBalancerType,
			x.LoadBalancerType,
			"")

	}

	// Auto Scaling
	{
		flags.BoolVar(
			&x.EnableAutoScaler,
			utils.EnableAutoScaler,
			x.EnableAutoScaler,
			"")

		flags.BoolVar(
			&x.EnableAutoConfig,
			utils.EnableAutoScalerAutoconfig,
			x.EnableAutoConfig,
			"")

		flags.IntVar(
			&x.Cooldown,
			utils.Cooldown,
			x.Cooldown,
			"")

		flags.IntVar(
			&x.HeadroomCpuPerUnit,
			utils.HeadroomCpuPerUnit,
			x.HeadroomCpuPerUnit,
			"")

		flags.IntVar(
			&x.HeadroomMemoryPerUnit,
			utils.HeadroomMemoryPerUnit,
			x.HeadroomMemoryPerUnit,
			"")

		flags.IntVar(
			&x.HeadroomGpuPerUnit,
			utils.HeadroomGpuPerUnit,
			x.HeadroomGpuPerUnit,
			"")

		flags.IntVar(
			&x.HeadroomNumPerUnit,
			utils.HeadroomNumPerUnit,
			x.HeadroomNumPerUnit,
			"")

		flags.IntVar(
			&x.ResourceLimitMaxVCpu,
			utils.ResourceLimitMaxVcpu,
			x.ResourceLimitMaxVCpu,
			"")

		flags.IntVar(
			&x.ResourceLimitMaxMemory,
			utils.ResourceLimitMaxMemory,
			x.ResourceLimitMaxMemory,
			"")

		flags.IntVar(
			&x.EvaluationPeriods,
			utils.EvaluationPeriods,
			x.EvaluationPeriods,
			"")

		flags.IntVar(
			&x.MaxScaleDownPercentage,
			utils.MaxScaleDownPercentage,
			x.MaxScaleDownPercentage,
			"")
	}
}

func (x *CmdUpdateClusterKubernetesOptions) Validate() error {
	if err := x.CmdUpdateClusterOptions.Validate(); err != nil {
		return err
	}

	if x.ClusterID == "" {
		return errors.Required("ClusterID")
	}

	return nil
}
