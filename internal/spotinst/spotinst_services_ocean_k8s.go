package spotinst

import (
	"context"

	"github.com/spotinst/spotinst-cli/internal/log"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

type oceanKubernetesAWS struct {
	svc aws.Service
}

func (x *oceanKubernetesAWS) ListClusters(ctx context.Context) ([]*OceanCluster, error) {
	log.Debugf("Listing all Kubernetes clusters")

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
				Obj:       cluster,
			},
		}
	}

	return clusters, nil
}

func (x *oceanKubernetesAWS) ListLaunchSpecs(ctx context.Context) ([]*OceanLaunchSpec, error) {
	log.Debugf("Listing all Kubernetes launch specs")

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
				Obj:       spec,
			},
		}
	}

	return specs, nil
}
