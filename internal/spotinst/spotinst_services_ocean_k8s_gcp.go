package spotinst

import (
	"context"

	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/gcp"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

type oceanKubernetesGCP struct {
	svc gcp.Service
}

func (x *oceanKubernetesGCP) NewClusterBuilder(fs *pflag.FlagSet, opts *OceanClusterOptions) OceanClusterBuilder {
	return &oceanKubernetesGCPClusterBuilder{fs, opts}
}

func (x *oceanKubernetesGCP) NewLaunchSpecBuilder(fs *pflag.FlagSet, opts *OceanLaunchSpecOptions) OceanLaunchSpecBuilder {
	return &oceanKubernetesGCPLaunchSpecBuilder{fs, opts}
}

func (x *oceanKubernetesGCP) NewRolloutBuilder(fs *pflag.FlagSet, opts *OceanRolloutOptions) OceanRolloutBuilder {
	return &oceanKubernetesGCPRolloutBuilder{fs, opts}
}

func (x *oceanKubernetesGCP) ListClusters(ctx context.Context) ([]*OceanCluster, error) {
	log.Debugf("Listing all Kubernetes clusters")

	output, err := x.svc.ListClusters(ctx, &gcp.ListClustersInput{})
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

func (x *oceanKubernetesGCP) ListLaunchSpecs(ctx context.Context) ([]*OceanLaunchSpec, error) {
	log.Debugf("Listing all Kubernetes launch specs")

	output, err := x.svc.ListLaunchSpecs(ctx, &gcp.ListLaunchSpecsInput{})
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
				ID: spotinst.StringValue(spec.ID),
				//Name:      spotinst.StringValue(spec.Name),
				//CreatedAt: spotinst.TimeValue(spec.CreatedAt),
				//UpdatedAt: spotinst.TimeValue(spec.UpdatedAt),
			},
			Obj: spec,
		}
	}

	return specs, nil
}

func (x *oceanKubernetesGCP) ListRollouts(ctx context.Context, clusterID string) ([]*OceanRollout, error) {
	return nil, ErrNotImplemented
}

func (x *oceanKubernetesGCP) GetCluster(ctx context.Context, clusterID string) (*OceanCluster, error) {
	log.Debugf("Getting a Kubernetes cluster by ID: %s", clusterID)

	input := &gcp.ReadClusterInput{
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

func (x *oceanKubernetesGCP) GetLaunchSpec(ctx context.Context, specID string) (*OceanLaunchSpec, error) {
	log.Debugf("Getting a Kubernetes launch spec by ID: %s", specID)

	input := &gcp.ReadLaunchSpecInput{
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
			ID: spotinst.StringValue(output.LaunchSpec.ID),
			//Name:      spotinst.StringValue(output.LaunchSpec.Name),
			//CreatedAt: spotinst.TimeValue(output.LaunchSpec.CreatedAt),
			//UpdatedAt: spotinst.TimeValue(output.LaunchSpec.UpdatedAt),
		},
		Obj: output.LaunchSpec,
	}

	return spec, nil
}

func (x *oceanKubernetesGCP) GetRollout(ctx context.Context, clusterID, rolloutID string) (*OceanRollout, error) {
	return nil, ErrNotImplemented
}

func (x *oceanKubernetesGCP) CreateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error) {
	log.Debugf("Creating a new Kubernetes cluster")

	input := &gcp.CreateClusterInput{
		Cluster: cluster.Obj.(*gcp.Cluster),
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

func (x *oceanKubernetesGCP) CreateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error) {
	log.Debugf("Creating a new Kubernetes launch spec")

	input := &gcp.CreateLaunchSpecInput{
		LaunchSpec: spec.Obj.(*gcp.LaunchSpec),
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
			ID: spotinst.StringValue(output.LaunchSpec.ID),
			//Name:      spotinst.StringValue(output.LaunchSpec.Name),
			//CreatedAt: spotinst.TimeValue(output.LaunchSpec.CreatedAt),
			//UpdatedAt: spotinst.TimeValue(output.LaunchSpec.UpdatedAt),
		},
		Obj: output.LaunchSpec,
	}

	return created, nil
}

func (x *oceanKubernetesGCP) CreateRollout(ctx context.Context, rollout *OceanRollout) (*OceanRollout, error) {
	return nil, ErrNotImplemented
}

func (x *oceanKubernetesGCP) UpdateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error) {
	log.Debugf("Updating a Kubernetes cluster by ID: %s", cluster.ID)

	input := &gcp.UpdateClusterInput{
		Cluster: cluster.Obj.(*gcp.Cluster),
	}

	// Remove read-only fields.
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

func (x *oceanKubernetesGCP) UpdateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error) {
	log.Debugf("Updating a Kubernetes launch spec by ID: %s", spec.ID)

	input := &gcp.UpdateLaunchSpecInput{
		LaunchSpec: spec.Obj.(*gcp.LaunchSpec),
	}

	// Remove read-only fields.
	//input.LaunchSpec.UpdatedAt = nil
	//input.LaunchSpec.CreatedAt = nil

	output, err := x.svc.UpdateLaunchSpec(ctx, input)
	if err != nil {
		return nil, err
	}

	updated := &OceanLaunchSpec{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID: spotinst.StringValue(output.LaunchSpec.ID),
			//Name:      spotinst.StringValue(output.LaunchSpec.Name),
			//CreatedAt: spotinst.TimeValue(output.LaunchSpec.CreatedAt),
			//UpdatedAt: spotinst.TimeValue(output.LaunchSpec.UpdatedAt),
		},
		Obj: output.LaunchSpec,
	}

	return updated, nil
}

func (x *oceanKubernetesGCP) UpdateRollout(ctx context.Context, rollout *OceanRollout) (*OceanRollout, error) {
	return nil, ErrNotImplemented
}

func (x *oceanKubernetesGCP) DeleteCluster(ctx context.Context, clusterID string) error {
	log.Debugf("Deleting a Kubernetes cluster by ID: %s", clusterID)

	input := &gcp.DeleteClusterInput{
		ClusterID: spotinst.String(clusterID),
	}

	_, err := x.svc.DeleteCluster(ctx, input)
	return err
}

func (x *oceanKubernetesGCP) DeleteLaunchSpec(ctx context.Context, specID string) error {
	log.Debugf("Deleting a Kubernetes launch spec by ID: %s", specID)

	input := &gcp.DeleteLaunchSpecInput{
		LaunchSpecID: spotinst.String(specID),
	}

	_, err := x.svc.DeleteLaunchSpec(ctx, input)
	return err
}

type oceanKubernetesGCPClusterBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanClusterOptions
}

func (x *oceanKubernetesGCPClusterBuilder) Build() (*OceanCluster, error) {
	return nil, ErrNotImplemented
}

type oceanKubernetesGCPLaunchSpecBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanLaunchSpecOptions
}

func (x *oceanKubernetesGCPLaunchSpecBuilder) Build() (*OceanLaunchSpec, error) {
	return nil, ErrNotImplemented
}

type oceanKubernetesGCPRolloutBuilder struct {
	fs   *pflag.FlagSet
	opts *OceanRolloutOptions
}

func (x *oceanKubernetesGCPRolloutBuilder) Build() (*OceanRollout, error) {
	return nil, ErrNotImplemented
}
