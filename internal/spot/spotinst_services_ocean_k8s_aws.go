package spot

import (
	"context"
	"encoding/base64"

	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

type oceanKubernetesAWS struct {
	svc aws.Service
}

func (x *oceanKubernetesAWS) NewClusterBuilder(fs *pflag.FlagSet, opts *OceanClusterOptions) OceanClusterBuilder {
	return &oceanKubernetesAWSClusterBuilder{fs, opts}
}

func (x *oceanKubernetesAWS) NewLaunchSpecBuilder(fs *pflag.FlagSet, opts *OceanLaunchSpecOptions) OceanLaunchSpecBuilder {
	return &oceanKubernetesAWSLaunchSpecBuilder{fs, opts}
}

func (x *oceanKubernetesAWS) NewRolloutBuilder(fs *pflag.FlagSet, opts *OceanRolloutOptions) OceanRolloutBuilder {
	return &oceanKubernetesAWSRolloutBuilder{fs, opts}
}

func (x *oceanKubernetesAWS) ListClusters(ctx context.Context) ([]*OceanCluster, error) {
	log.Debugf("Listing Ocean (Kubernetes) clusters")

	output, err := x.svc.ListClusters(ctx, &aws.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	clusters := make([]*OceanCluster, len(output.Clusters))
	for i, cluster := range output.Clusters {
		clusters[i] = &OceanCluster{
			TypeMeta: TypeMeta{
				Kind: typeOf(OceanCluster{}),
			},
			ObjectMeta: ObjectMeta{
				ID:        spotinst.StringValue(cluster.ID),
				Name:      spotinst.StringValue(cluster.Name),
				CreatedAt: spotinst.TimeValue(cluster.CreatedAt),
				UpdatedAt: spotinst.TimeValue(cluster.UpdatedAt),
			},
			Obj: cluster,
		}
	}

	return clusters, nil
}

func (x *oceanKubernetesAWS) ListLaunchSpecs(ctx context.Context) ([]*OceanLaunchSpec, error) {
	log.Debugf("Listing Ocean (Kubernetes) launch specs")

	output, err := x.svc.ListLaunchSpecs(ctx, &aws.ListLaunchSpecsInput{})
	if err != nil {
		return nil, err
	}

	specs := make([]*OceanLaunchSpec, len(output.LaunchSpecs))
	for i, spec := range output.LaunchSpecs {
		specs[i] = &OceanLaunchSpec{
			TypeMeta: TypeMeta{
				Kind: typeOf(OceanLaunchSpec{}),
			},
			ObjectMeta: ObjectMeta{
				ID:        spotinst.StringValue(spec.ID),
				Name:      spotinst.StringValue(spec.Name),
				CreatedAt: spotinst.TimeValue(spec.CreatedAt),
				UpdatedAt: spotinst.TimeValue(spec.UpdatedAt),
			},
			Obj: spec,
		}
	}

	return specs, nil
}

func (x *oceanKubernetesAWS) ListRollouts(ctx context.Context, clusterID string) ([]*OceanRollout, error) {
	log.Debugf("Listing Kubernetes rollouts")

	input := &aws.ListRollsInput{
		ClusterID: spotinst.String(clusterID),
	}

	output, err := x.svc.ListRolls(ctx, input)
	if err != nil {
		return nil, err
	}

	rollouts := make([]*OceanRollout, len(output.Rolls))
	for i, rollout := range output.Rolls {
		rollouts[i] = &OceanRollout{
			TypeMeta: TypeMeta{
				Kind: typeOf(OceanRollout{}),
			},
			ObjectMeta: ObjectMeta{
				ID:        spotinst.StringValue(rollout.ID),
				Name:      spotinst.StringValue(rollout.Comment),
				CreatedAt: spotinst.TimeValue(rollout.CreatedAt),
				UpdatedAt: spotinst.TimeValue(rollout.UpdatedAt),
			},
			Obj: rollout,
		}
	}

	return rollouts, nil
}

func (x *oceanKubernetesAWS) GetCluster(ctx context.Context, clusterID string) (*OceanCluster, error) {
	log.Debugf("Getting Ocean (Kubernetes) cluster by ID: %s", clusterID)

	input := &aws.ReadClusterInput{
		ClusterID: spotinst.String(clusterID),
	}

	output, err := x.svc.ReadCluster(ctx, input)
	if err != nil {
		return nil, err
	}

	cluster := &OceanCluster{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Cluster.ID),
			Name:      spotinst.StringValue(output.Cluster.Name),
			CreatedAt: spotinst.TimeValue(output.Cluster.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Cluster.UpdatedAt),
		},
		Obj: output.Cluster,
	}

	return cluster, nil
}

func (x *oceanKubernetesAWS) GetLaunchSpec(ctx context.Context, specID string) (*OceanLaunchSpec, error) {
	log.Debugf("Getting Ocean (Kubernetes) launch spec by ID: %s", specID)

	input := &aws.ReadLaunchSpecInput{
		LaunchSpecID: spotinst.String(specID),
	}

	output, err := x.svc.ReadLaunchSpec(ctx, input)
	if err != nil {
		return nil, err
	}

	spec := &OceanLaunchSpec{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanLaunchSpec{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.LaunchSpec.ID),
			Name:      spotinst.StringValue(output.LaunchSpec.Name),
			CreatedAt: spotinst.TimeValue(output.LaunchSpec.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.LaunchSpec.UpdatedAt),
		},
		Obj: output.LaunchSpec,
	}

	return spec, nil
}

func (x *oceanKubernetesAWS) GetRollout(ctx context.Context, clusterID, rolloutID string) (*OceanRollout, error) {
	log.Debugf("GettingKubernetes rollout by ID: %s", rolloutID)

	input := &aws.ReadRollInput{
		ClusterID: spotinst.String(clusterID),
		RollID:    spotinst.String(rolloutID),
	}

	output, err := x.svc.ReadRoll(ctx, input)
	if err != nil {
		return nil, err
	}

	rollout := &OceanRollout{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanRollout{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Roll.ID),
			Name:      spotinst.StringValue(output.Roll.Comment),
			CreatedAt: spotinst.TimeValue(output.Roll.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Roll.UpdatedAt),
		},
		Obj: output.Roll,
	}

	return rollout, nil
}

func (x *oceanKubernetesAWS) CreateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error) {
	log.Debugf("Creating Ocean (Kubernetes) cluster")

	input := &aws.CreateClusterInput{
		Cluster: cluster.Obj.(*aws.Cluster),
	}

	output, err := x.svc.CreateCluster(ctx, input)
	if err != nil {
		return nil, err
	}

	created := &OceanCluster{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Cluster.ID),
			Name:      spotinst.StringValue(output.Cluster.Name),
			CreatedAt: spotinst.TimeValue(output.Cluster.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Cluster.UpdatedAt),
		},
		Obj: output.Cluster,
	}

	return created, nil
}

func (x *oceanKubernetesAWS) CreateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error) {
	log.Debugf("Creating Ocean (Kubernetes) launch spec")

	input := &aws.CreateLaunchSpecInput{
		LaunchSpec: spec.Obj.(*aws.LaunchSpec),
	}

	output, err := x.svc.CreateLaunchSpec(ctx, input)
	if err != nil {
		return nil, err
	}

	created := &OceanLaunchSpec{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.LaunchSpec.ID),
			Name:      spotinst.StringValue(output.LaunchSpec.Name),
			CreatedAt: spotinst.TimeValue(output.LaunchSpec.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.LaunchSpec.UpdatedAt),
		},
		Obj: output.LaunchSpec,
	}

	return created, nil
}

func (x *oceanKubernetesAWS) CreateRollout(ctx context.Context, rollout *OceanRollout) (*OceanRollout, error) {
	log.Debugf("Creating Kubernetes rollout")

	input := &aws.CreateRollInput{
		Roll: rollout.Obj.(*aws.RollSpec),
	}

	output, err := x.svc.CreateRoll(ctx, input)
	if err != nil {
		return nil, err
	}

	created := &OceanRollout{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanRollout{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Roll.ID),
			Name:      spotinst.StringValue(output.Roll.Comment),
			CreatedAt: spotinst.TimeValue(output.Roll.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Roll.UpdatedAt),
		},
		Obj: output.Roll,
	}

	return created, nil
}

func (x *oceanKubernetesAWS) UpdateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error) {
	log.Debugf("Updating Ocean (Kubernetes) cluster by ID: %s", cluster.ID)

	input := &aws.UpdateClusterInput{
		Cluster: cluster.Obj.(*aws.Cluster),
	}

	// Remove read-only fields.
	input.Cluster.Region = nil
	input.Cluster.UpdatedAt = nil
	input.Cluster.CreatedAt = nil

	output, err := x.svc.UpdateCluster(ctx, input)
	if err != nil {
		return nil, err
	}

	updated := &OceanCluster{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Cluster.ID),
			Name:      spotinst.StringValue(output.Cluster.Name),
			CreatedAt: spotinst.TimeValue(output.Cluster.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Cluster.UpdatedAt),
		},
		Obj: output.Cluster,
	}

	return updated, nil
}

func (x *oceanKubernetesAWS) UpdateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error) {
	log.Debugf("Updating Ocean (Kubernetes) launch spec by ID: %s", spec.ID)

	input := &aws.UpdateLaunchSpecInput{
		LaunchSpec: spec.Obj.(*aws.LaunchSpec),
	}

	// Remove read-only fields.
	input.LaunchSpec.UpdatedAt = nil
	input.LaunchSpec.CreatedAt = nil

	output, err := x.svc.UpdateLaunchSpec(ctx, input)
	if err != nil {
		return nil, err
	}

	updated := &OceanLaunchSpec{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.LaunchSpec.ID),
			Name:      spotinst.StringValue(output.LaunchSpec.Name),
			CreatedAt: spotinst.TimeValue(output.LaunchSpec.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.LaunchSpec.UpdatedAt),
		},
		Obj: output.LaunchSpec,
	}

	return updated, nil
}

func (x *oceanKubernetesAWS) UpdateRollout(ctx context.Context, rollout *OceanRollout) (*OceanRollout, error) {
	log.Debugf("Updating Kubernetes rollout by ID: %s", rollout.ID)

	input := &aws.UpdateRollInput{
		Roll: rollout.Obj.(*aws.RollSpec),
	}

	output, err := x.svc.UpdateRoll(ctx, input)
	if err != nil {
		return nil, err
	}

	updated := &OceanRollout{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanRollout{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Roll.ID),
			Name:      spotinst.StringValue(output.Roll.Comment),
			CreatedAt: spotinst.TimeValue(output.Roll.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Roll.UpdatedAt),
		},
		Obj: output.Roll,
	}

	return updated, nil
}

func (x *oceanKubernetesAWS) DeleteCluster(ctx context.Context, clusterID string) error {
	log.Debugf("Deleting Ocean (Kubernetes) cluster by ID: %s", clusterID)

	input := &aws.DeleteClusterInput{
		ClusterID: spotinst.String(clusterID),
	}

	_, err := x.svc.DeleteCluster(ctx, input)
	return err
}

func (x *oceanKubernetesAWS) DeleteLaunchSpec(ctx context.Context, specID string) error {
	log.Debugf("Deleting Ocean (Kubernetes) launch spec by ID: %s", specID)

	input := &aws.DeleteLaunchSpecInput{
		LaunchSpecID: spotinst.String(specID),
	}

	_, err := x.svc.DeleteLaunchSpec(ctx, input)
	return err
}

type oceanKubernetesAWSClusterBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanClusterOptions
}

func (x *oceanKubernetesAWSClusterBuilder) Build() (*OceanCluster, error) {
	return &OceanCluster{Obj: x.buildCluster()}, nil
}

func (x *oceanKubernetesAWSClusterBuilder) buildCluster() *aws.Cluster {
	cluster := new(aws.Cluster)

	if x.fs.Changed(flags.FlagOceanClusterID) {
		cluster.SetId(spotinst.String(x.opts.ClusterID))
	}

	if x.fs.Changed(flags.FlagOceanName) {
		cluster.SetName(spotinst.String(x.opts.Name))
	}

	if x.fs.Changed(flags.FlagOceanControllerID) {
		cluster.SetControllerClusterId(spotinst.String(x.opts.ControllerID))
	} else if x.fs.Changed(flags.FlagOceanName) {
		cluster.SetControllerClusterId(spotinst.String(x.opts.Name))
	}

	if x.fs.Changed(flags.FlagOceanRegion) {
		cluster.SetRegion(spotinst.String(x.opts.Region))
	}

	cluster.SetStrategy(x.buildStrategy())
	cluster.SetCapacity(x.buildCapacity())
	cluster.SetAutoScaler(x.buildAutoScaler())
	cluster.SetCompute(x.buildCompute())

	return cluster
}

func (x *oceanKubernetesAWSClusterBuilder) buildStrategy() *aws.Strategy {
	strategy := new(aws.Strategy)

	if x.fs.Changed(flags.FlagOceanSpotPercentage) {
		strategy.SetSpotPercentage(spotinst.Float64(x.opts.SpotPercentage))
	}

	if x.fs.Changed(flags.FlagOceanUtilizeReserveInstances) {
		strategy.SetUtilizeReservedInstances(spotinst.Bool(x.opts.UtilizeReservedInstances))
	}

	if x.fs.Changed(flags.FlagOceanFallbackOnDemand) {
		strategy.SetFallbackToOnDemand(spotinst.Bool(x.opts.FallbackToOnDemand))
	}

	if x.fs.Changed(flags.FlagOceanDrainingTimeout) {
		strategy.SetDrainingTimeout(spotinst.Int(x.opts.DrainingTimeout))
	}

	if x.fs.Changed(flags.FlagOceanClusterOrientation) {
		//		data := aws.Strategy{}spotinst.String(flags.FlagOceanClusterOrientation)
		test := aws.ClusterOrientation{}
		test.SetAvailabilityVsCost(spotinst.String(x.opts.ClusterOrientation))
		strategy.SetClusterOrientation(&test)
	}

	return strategy
}

func (x *oceanKubernetesAWSClusterBuilder) buildCapacity() *aws.Capacity {
	capacity := new(aws.Capacity)

	if x.fs.Changed(flags.FlagOceanMinSize) {
		capacity.SetMinimum(spotinst.Int(x.opts.MinSize))
	}

	if x.fs.Changed(flags.FlagOceanMaxSize) {
		capacity.SetMaximum(spotinst.Int(x.opts.MaxSize))
	}

	if x.fs.Changed(flags.FlagOceanTargetSize) {
		capacity.SetTarget(spotinst.Int(x.opts.TargetSize))
	}

	return capacity
}

func (x *oceanKubernetesAWSClusterBuilder) buildAutoScaler() *aws.AutoScaler {
	autoScaler := new(aws.AutoScaler)

	if x.fs.Changed(flags.FlagOceanEnableAutoScaler) {
		autoScaler.SetIsEnabled(spotinst.Bool(x.opts.EnableAutoScaler))
	}

	if x.fs.Changed(flags.FlagOceanEnableAutoScalerAutoConfig) {
		autoScaler.SetIsAutoConfig(spotinst.Bool(x.opts.EnableAutoConfig))
	}

	if x.fs.Changed(flags.FlagOceanCooldown) {
		autoScaler.SetCooldown(spotinst.Int(x.opts.Cooldown))
	}

	if x.fs.Changed(flags.FlagOceanHeadroomCPUPerUnit) ||
		x.fs.Changed(flags.FlagOceanHeadroomMemoryPerUnit) ||
		x.fs.Changed(flags.FlagOceanHeadroomGPUPerUnit) ||
		x.fs.Changed(flags.FlagOceanHeadroomNumPerUnit) {
		headroom := new(aws.AutoScalerHeadroom)

		if x.fs.Changed(flags.FlagOceanHeadroomCPUPerUnit) {
			headroom.SetCPUPerUnit(spotinst.Int(x.opts.HeadroomCPUPerUnit))
		}

		if x.fs.Changed(flags.FlagOceanHeadroomMemoryPerUnit) {
			headroom.SetMemoryPerUnit(spotinst.Int(x.opts.HeadroomMemoryPerUnit))
		}

		if x.fs.Changed(flags.FlagOceanHeadroomGPUPerUnit) {
			headroom.SetGPUPerUnit(spotinst.Int(x.opts.HeadroomGPUPerUnit))
		}

		if x.fs.Changed(flags.FlagOceanHeadroomNumPerUnit) {
			headroom.SetNumOfUnits(spotinst.Int(x.opts.HeadroomNumPerUnit))
		}

		autoScaler.SetHeadroom(headroom)
	}

	if x.fs.Changed(flags.FlagOceanResourceLimitMaxVCPU) ||
		x.fs.Changed(flags.FlagOceanResourceLimitMaxMemory) {
		resourceLimit := new(aws.AutoScalerResourceLimits)

		if x.fs.Changed(flags.FlagOceanResourceLimitMaxMemory) {
			resourceLimit.SetMaxMemoryGiB(spotinst.Int(x.opts.ResourceLimitMaxMemory))
		}

		if x.fs.Changed(flags.FlagOceanResourceLimitMaxVCPU) {
			resourceLimit.SetMaxVCPU(spotinst.Int(x.opts.ResourceLimitMaxVCPU))
		}

		autoScaler.SetResourceLimits(resourceLimit)
	}

	if x.fs.Changed(flags.FlagOceanEvaluationPeriods) {
		down := new(aws.AutoScalerDown)
		down.SetEvaluationPeriods(spotinst.Int(x.opts.EvaluationPeriods))

		autoScaler.SetDown(down)
	}

	return autoScaler
}

func (x *oceanKubernetesAWSClusterBuilder) buildCompute() *aws.Compute {
	compute := new(aws.Compute)

	if x.fs.Changed(flags.FlagOceanSubnetIDs) {
		compute.SetSubnetIDs(x.opts.SubnetIDs)
	}

	if x.fs.Changed(flags.FlagOceanInstancesTypesBlacklist) ||
		x.fs.Changed(flags.FlagOceanInstancesTypesWhitelist) {
		instanceTypes := new(aws.InstanceTypes)
		instanceTypes.SetBlacklist(x.opts.InstanceTypesBlacklist)
		instanceTypes.SetWhitelist(x.opts.InstanceTypesWhitelist)

		compute.SetInstanceTypes(instanceTypes)
	}

	launchSpec := new(aws.LaunchSpecification)

	if x.fs.Changed(flags.FlagOceanAssociatePublicIPAddress) {
		launchSpec.SetAssociatePublicIPAddress(spotinst.Bool(x.opts.AssociatePublicIPAddress))
	}

	if x.fs.Changed(flags.FlagOceanSecurityGroupIDs) {
		launchSpec.SetSecurityGroupIDs(x.opts.SecurityGroupIDs)
	}

	if x.fs.Changed(flags.FlagOceanImageID) {
		launchSpec.SetImageId(spotinst.String(x.opts.ImageID))
	}

	if x.fs.Changed(flags.FlagOceanKeyPair) {
		launchSpec.SetKeyPair(spotinst.String(x.opts.KeyPair))
	}

	if x.fs.Changed(flags.FlagOceanUserData) {
		launchSpec.SetUserData(spotinst.String(x.opts.UserData))
	}

	if x.fs.Changed(flags.FlagOceanRootVolumeSize) {
		launchSpec.SetRootVolumeSize(spotinst.Int(x.opts.RootVolumeSize))
	}

	if x.fs.Changed(flags.FlagOceanEnableMonitoring) {
		launchSpec.SetMonitoring(spotinst.Bool(x.opts.EnableMonitoring))
	}

	if x.fs.Changed(flags.FlagOceanEnableEBSOptimization) {
		launchSpec.SetEBSOptimized(spotinst.Bool(x.opts.EnableEBSOptimization))
	}

	if x.fs.Changed(flags.FlagOceanIamInstanceProfileName) ||
		x.fs.Changed(flags.FlagOceanIamInstanceProfileARN) {
		iam := new(aws.IAMInstanceProfile)

		if x.fs.Changed(flags.FlagOceanIamInstanceProfileName) {
			iam.SetName(spotinst.String(x.opts.IAMInstanceProfileName))
		}

		if x.fs.Changed(flags.FlagOceanIamInstanceProfileARN) {
			iam.SetArn(spotinst.String(x.opts.IAMInstanceProfileARN))
		}

		launchSpec.SetIAMInstanceProfile(iam)
	}

	if x.fs.Changed(flags.FlagOceanLoadBalancerName) ||
		x.fs.Changed(flags.FlagOceanLoadBalancerARN) {

		// Classic Load Balancers.
		if x.fs.Changed(flags.FlagOceanLoadBalancerName) && len(x.opts.LoadBalancerNames) > 0 {
			for _, name := range x.opts.LoadBalancerNames {
				launchSpec.LoadBalancers = append(launchSpec.LoadBalancers, &aws.LoadBalancer{
					Name: spotinst.String(name),
					Type: spotinst.String("CLASSIC"),
				})
			}
		}

		// Target Groups.
		if x.fs.Changed(flags.FlagOceanLoadBalancerARN) && len(x.opts.LoadBalancerARNs) > 0 {
			for _, arn := range x.opts.LoadBalancerARNs {
				launchSpec.LoadBalancers = append(launchSpec.LoadBalancers, &aws.LoadBalancer{
					Arn:  spotinst.String(arn),
					Type: spotinst.String("TARGET_GROUP"),
				})
			}
		}
	}

	compute.SetLaunchSpecification(launchSpec)

	return compute
}

type oceanKubernetesAWSLaunchSpecBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanLaunchSpecOptions
}

func (x *oceanKubernetesAWSLaunchSpecBuilder) Build() (*OceanLaunchSpec, error) {
	return &OceanLaunchSpec{Obj: x.buildSpec()}, nil
}

func (x *oceanKubernetesAWSLaunchSpecBuilder) buildSpec() *aws.LaunchSpec {
	spec := new(aws.LaunchSpec)

	if x.fs.Changed(flags.FlagOceanSpecID) {
		spec.SetId(spotinst.String(x.opts.SpecID))
	}

	if x.fs.Changed(flags.FlagOceanName) {
		spec.SetName(spotinst.String(x.opts.Name))
	}

	if x.fs.Changed(flags.FlagOceanClusterID) {
		spec.SetOceanId(spotinst.String(x.opts.ClusterID))
	}

	if x.fs.Changed(flags.FlagOceanImageID) {
		spec.SetImageId(spotinst.String(x.opts.ImageID))
	}

	if x.fs.Changed(flags.FlagOceanUserData) {
		if _, err := base64.StdEncoding.DecodeString(x.opts.UserData); err != nil {
			x.opts.UserData = base64.StdEncoding.EncodeToString([]byte(x.opts.UserData))
		}

		spec.SetUserData(spotinst.String(x.opts.UserData))
	}

	if x.fs.Changed(flags.FlagOceanSecurityGroupIDs) {
		spec.SetSecurityGroupIDs(x.opts.SecurityGroupIDs)
	}

	return spec
}

type oceanKubernetesAWSRolloutBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanRolloutOptions
}

func (x *oceanKubernetesAWSRolloutBuilder) Build() (*OceanRollout, error) {
	return &OceanRollout{Obj: x.buildRollout()}, nil
}

func (x *oceanKubernetesAWSRolloutBuilder) buildRollout() *aws.RollSpec {
	spec := new(aws.RollSpec)

	if x.fs.Changed(flags.FlagOceanClusterID) {
		spec.ClusterID = spotinst.String(x.opts.ClusterID)
	}

	if x.fs.Changed(flags.FlagOceanRolloutID) {
		spec.ID = spotinst.String(x.opts.RolloutID)
	}

	if x.fs.Changed(flags.FlagOceanRolloutBatchSizePercentage) {
		spec.SetBatchSizePercentage(spotinst.Int(x.opts.BatchSizePercentage))
	}

	if x.fs.Changed(flags.FlagOceanRolloutDisableAutoScaling) {
		spec.SetDisableLaunchSpecAutoScaling(spotinst.Bool(x.opts.DisableAutoScaling))
	}

	if x.fs.Changed(flags.FlagOceanRolloutSpecIDs) {
		spec.SetLaunchSpecIDs(x.opts.SpecIDs)
	}

	if x.fs.Changed(flags.FlagOceanRolloutInstanceIDs) {
		spec.SetInstanceIDs(x.opts.InstanceIDs)
	}

	if len(x.opts.Comment) > 0 {
		spec.SetComment(spotinst.String(x.opts.Comment))
	}

	if len(x.opts.Status) > 0 {
		spec.SetStatus(spotinst.String(x.opts.Status))
	}

	if x.fs.Changed(flags.FlagOceanRolloutRespectPDB) {
		spec.SetRespectPDB(spotinst.Bool(x.opts.RespectPDB))
	}

	return spec
}
