package ocean

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
)

type (
	CmdCreateClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdCreateClusterKubernetesOptions
	}

	CmdCreateClusterKubernetesOptions struct {
		*CmdCreateClusterOptions
		spot.OceanClusterOptions
	}
)

func NewCmdCreateClusterKubernetes(opts *CmdCreateClusterOptions) *cobra.Command {
	return newCmdCreateClusterKubernetes(opts).cmd
}

func newCmdCreateClusterKubernetes(opts *CmdCreateClusterOptions) *CmdCreateClusterKubernetes {
	var cmd CmdCreateClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Create a new Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"k8s", "kube", "k"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdCreateClusterKubernetes) Run(ctx context.Context) error {
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

func (x *CmdCreateClusterKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdCreateClusterKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdCreateClusterKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdCreateClusterKubernetes) run(ctx context.Context) error {
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
		spot.WithDryRun(x.opts.DryRun),
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	oceanCluster, err := oceanClient.NewClusterBuilder(x.cmd.Flags(), &x.opts.OceanClusterOptions).Build()
	if err != nil {
		return err
	}

	cluster, err := oceanClient.CreateCluster(ctx, oceanCluster)
	if err != nil {
		return err
	}

	fmt.Fprintln(x.opts.Out, fmt.Sprintf("Created (%q).", cluster.ID))
	return nil
}

func (x *CmdCreateClusterKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdCreateClusterOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdCreateClusterKubernetesOptions) initDefaults(opts *CmdCreateClusterOptions) {
	x.CmdCreateClusterOptions = opts
}

func (x *CmdCreateClusterKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	// Base.
	{
		fs.StringVar(&x.Name, flags.FlagOceanName, x.Name, "")
		fs.StringVar(&x.ControllerID, flags.FlagOceanControllerID, x.ControllerID, "")
		fs.StringVar(&x.Region, flags.FlagOceanRegion, x.Region, "")
	}

	// Strategy.
	{
		fs.Float64Var(&x.SpotPercentage, flags.FlagOceanSpotPercentage, x.SpotPercentage, "")
		fs.IntVar(&x.DrainingTimeout, flags.FlagOceanDrainingTimeout, x.DrainingTimeout, "")
		fs.BoolVar(&x.UtilizeReservedInstances, flags.FlagOceanUtilizeReserveInstances, x.UtilizeReservedInstances, "")
		fs.BoolVar(&x.FallbackToOnDemand, flags.FlagOceanFallbackOnDemand, x.FallbackToOnDemand, "")
		fs.StringVar(&x.ClusterOrientation, flags.FlagOceanClusterOrientation, x.ClusterOrientation, "")
	}

	// Capacity.
	{
		fs.IntVar(&x.MinSize, flags.FlagOceanMinSize, x.MinSize, "")
		fs.IntVar(&x.MaxSize, flags.FlagOceanMaxSize, x.MaxSize, "")
		fs.IntVar(&x.TargetSize, flags.FlagOceanTargetSize, x.TargetSize, "")
	}

	// Compute.
	{
		fs.StringSliceVar(&x.SubnetIDs, flags.FlagOceanSubnetIDs, x.SubnetIDs, "")
		fs.StringSliceVar(&x.InstanceTypesWhitelist, flags.FlagOceanInstancesTypesWhitelist, x.InstanceTypesWhitelist, "")
		fs.StringSliceVar(&x.InstanceTypesBlacklist, flags.FlagOceanInstancesTypesBlacklist, x.InstanceTypesBlacklist, "")
		fs.StringSliceVar(&x.SecurityGroupIDs, flags.FlagOceanSecurityGroupIDs, x.SecurityGroupIDs, "")
		fs.StringVar(&x.ImageID, flags.FlagOceanImageID, x.ImageID, "")
		fs.StringVar(&x.KeyPair, flags.FlagOceanKeyPair, x.KeyPair, "")
		fs.StringVar(&x.UserData, flags.FlagOceanUserData, x.UserData, "")
		fs.IntVar(&x.RootVolumeSize, flags.FlagOceanRootVolumeSize, x.RootVolumeSize, "")
		fs.BoolVar(&x.AssociatePublicIPAddress, flags.FlagOceanAssociatePublicIPAddress, x.AssociatePublicIPAddress, "")
		fs.BoolVar(&x.EnableMonitoring, flags.FlagOceanEnableMonitoring, x.EnableMonitoring, "")
		fs.BoolVar(&x.EnableEBSOptimization, flags.FlagOceanEnableEBSOptimization, x.EnableEBSOptimization, "")
		fs.StringVar(&x.IAMInstanceProfileName, flags.FlagOceanIamInstanceProfileName, x.IAMInstanceProfileName, "")
		fs.StringVar(&x.IAMInstanceProfileARN, flags.FlagOceanIamInstanceProfileARN, x.IAMInstanceProfileARN, "")
		fs.StringSliceVar(&x.LoadBalancerNames, flags.FlagOceanLoadBalancerName, x.LoadBalancerNames, "")
		fs.StringSliceVar(&x.LoadBalancerARNs, flags.FlagOceanLoadBalancerARN, x.LoadBalancerARNs, "")
		fs.StringVar(&x.LoadBalancerType, flags.FlagOceanLoadBalancerType, x.LoadBalancerType, "")
	}

	// Auto Scaling.
	{
		fs.BoolVar(&x.EnableAutoScaler, flags.FlagOceanEnableAutoScaler, x.EnableAutoScaler, "")
		fs.BoolVar(&x.EnableAutoConfig, flags.FlagOceanEnableAutoScalerAutoConfig, x.EnableAutoConfig, "")
		fs.IntVar(&x.Cooldown, flags.FlagOceanCooldown, x.Cooldown, "")
		fs.IntVar(&x.HeadroomCPUPerUnit, flags.FlagOceanHeadroomCPUPerUnit, x.HeadroomCPUPerUnit, "")
		fs.IntVar(&x.HeadroomMemoryPerUnit, flags.FlagOceanHeadroomMemoryPerUnit, x.HeadroomMemoryPerUnit, "")
		fs.IntVar(&x.HeadroomGPUPerUnit, flags.FlagOceanHeadroomGPUPerUnit, x.HeadroomGPUPerUnit, "")
		fs.IntVar(&x.HeadroomNumPerUnit, flags.FlagOceanHeadroomNumPerUnit, x.HeadroomNumPerUnit, "")
		fs.IntVar(&x.ResourceLimitMaxVCPU, flags.FlagOceanResourceLimitMaxVCPU, x.ResourceLimitMaxVCPU, "")
		fs.IntVar(&x.ResourceLimitMaxMemory, flags.FlagOceanResourceLimitMaxMemory, x.ResourceLimitMaxMemory, "")
		fs.IntVar(&x.EvaluationPeriods, flags.FlagOceanEvaluationPeriods, x.EvaluationPeriods, "")
		fs.IntVar(&x.MaxScaleDownPercentage, flags.FlagOceanMaxScaleDownPercentage, x.MaxScaleDownPercentage, "")
	}
}

func (x *CmdCreateClusterKubernetesOptions) Validate() error {
	return x.CmdCreateClusterOptions.Validate()
}
