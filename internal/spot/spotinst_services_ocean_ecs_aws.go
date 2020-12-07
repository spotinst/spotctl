package spot

import (
	"context"

	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

type oceanECS struct {
	svc aws.Service
}

func (x *oceanECS) NewClusterBuilder(fs *pflag.FlagSet, opts *OceanClusterOptions) OceanClusterBuilder {
	return &oceanECSClusterBuilder{fs, opts}
}

func (x *oceanECS) NewLaunchSpecBuilder(fs *pflag.FlagSet, opts *OceanLaunchSpecOptions) OceanLaunchSpecBuilder {
	return &oceanECSLaunchSpecBuilder{fs, opts}
}

func (x *oceanECS) NewRolloutBuilder(fs *pflag.FlagSet, opts *OceanRolloutOptions) OceanRolloutBuilder {
	return &oceanECSRolloutBuilder{fs, opts}
}

func (x *oceanECS) ListClusters(ctx context.Context) ([]*OceanCluster, error) {
	log.Debugf("Listing all ECS clusters")

	output, err := x.svc.ListECSClusters(ctx, &aws.ListECSClustersInput{})
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

func (x *oceanECS) ListLaunchSpecs(ctx context.Context) ([]*OceanLaunchSpec, error) {
	log.Debugf("Listing all ECS launch specs")

	output, err := x.svc.ListECSLaunchSpecs(ctx, &aws.ListECSLaunchSpecsInput{})
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

func (x *oceanECS) ListRollouts(ctx context.Context, clusterID string) ([]*OceanRollout, error) {
	return nil, ErrNotImplemented
}

func (x *oceanECS) GetCluster(ctx context.Context, clusterID string) (*OceanCluster, error) {
	log.Debugf("Getting a ECS cluster by ID: %s", clusterID)

	input := &aws.ReadECSClusterInput{
		ClusterID: spotinst.String(clusterID),
	}

	output, err := x.svc.ReadECSCluster(ctx, input)
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

func (x *oceanECS) GetLaunchSpec(ctx context.Context, specID string) (*OceanLaunchSpec, error) {
	log.Debugf("Getting a ECS launch spec by ID: %s", specID)

	input := &aws.ReadECSLaunchSpecInput{
		LaunchSpecID: spotinst.String(specID),
	}

	output, err := x.svc.ReadECSLaunchSpec(ctx, input)
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

func (x *oceanECS) GetRollout(ctx context.Context, clusterID, rolloutID string) (*OceanRollout, error) {
	return nil, ErrNotImplemented
}

func (x *oceanECS) CreateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error) {
	log.Debugf("Creating a new ECS cluster")

	input := &aws.CreateECSClusterInput{
		Cluster: cluster.Obj.(*aws.ECSCluster),
	}

	output, err := x.svc.CreateECSCluster(ctx, input)
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

func (x *oceanECS) CreateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error) {
	log.Debugf("Creating a new ECS launch spec")

	input := &aws.CreateECSLaunchSpecInput{
		LaunchSpec: spec.Obj.(*aws.ECSLaunchSpec),
	}

	output, err := x.svc.CreateECSLaunchSpec(ctx, input)
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

func (x *oceanECS) CreateRollout(ctx context.Context, rollout *OceanRollout) (*OceanRollout, error) {
	return nil, ErrNotImplemented
}

func (x *oceanECS) UpdateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error) {
	log.Debugf("Updating a ECS cluster by ID: %s", cluster.ID)

	input := &aws.UpdateECSClusterInput{
		Cluster: cluster.Obj.(*aws.ECSCluster),
	}

	// Remove read-only fields.
	input.Cluster.Region = nil
	input.Cluster.UpdatedAt = nil
	input.Cluster.CreatedAt = nil

	output, err := x.svc.UpdateECSCluster(ctx, input)
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

func (x *oceanECS) UpdateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error) {
	log.Debugf("Updating a ECS launch spec by ID: %s", spec.ID)

	input := &aws.UpdateECSLaunchSpecInput{
		LaunchSpec: spec.Obj.(*aws.ECSLaunchSpec),
	}

	// Remove read-only fields.
	input.LaunchSpec.UpdatedAt = nil
	input.LaunchSpec.CreatedAt = nil

	output, err := x.svc.UpdateECSLaunchSpec(ctx, input)
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

func (x *oceanECS) UpdateRollout(ctx context.Context, rollout *OceanRollout) (*OceanRollout, error) {
	return nil, ErrNotImplemented
}

func (x *oceanECS) DeleteCluster(ctx context.Context, clusterID string) error {
	log.Debugf("Deleting a ECS cluster by ID: %s", clusterID)

	input := &aws.DeleteECSClusterInput{
		ClusterID: spotinst.String(clusterID),
	}

	_, err := x.svc.DeleteECSCluster(ctx, input)
	return err
}

func (x *oceanECS) DeleteLaunchSpec(ctx context.Context, specID string) error {
	log.Debugf("Deleting a ECS launch spec by ID: %s", specID)

	input := &aws.DeleteECSLaunchSpecInput{
		LaunchSpecID: spotinst.String(specID),
	}

	_, err := x.svc.DeleteECSLaunchSpec(ctx, input)
	return err
}

type oceanECSClusterBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanClusterOptions
}

func (x *oceanECSClusterBuilder) Build() (*OceanCluster, error) {
	return nil, ErrNotImplemented
}

type oceanECSLaunchSpecBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanLaunchSpecOptions
}

func (x *oceanECSLaunchSpecBuilder) Build() (*OceanLaunchSpec, error) {
	return nil, ErrNotImplemented
}

type oceanECSRolloutBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanRolloutOptions
}

func (x *oceanECSRolloutBuilder) Build() (*OceanRollout, error) {
	return nil, ErrNotImplemented
}
