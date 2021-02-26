package spot

import (
	"context"
	"github.com/spotinst/spotinst-sdk-go/spotinst"

	"github.com/spotinst/spotctl/internal/log"
	wavesdk "github.com/spotinst/spotinst-sdk-go/service/wave"
)

type wave struct {
	svc wavesdk.Service
}

func (x *wave) GetCluster(ctx context.Context, clusterID string) (*WaveCluster, error) {
	log.Debugf("Getting Wave cluster by ID: %s", clusterID)

	input := &wavesdk.ReadClusterInput{
		ClusterID: spotinst.String(clusterID),
	}

	output, err := x.svc.ReadCluster(ctx, input)
	if err != nil {
		return nil, err
	}

	cluster := &WaveCluster{
		TypeMeta: TypeMeta{
			Kind: typeOf(WaveCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Cluster.ID),
			Name:      spotinst.StringValue(output.Cluster.ClusterIdentifier),
			CreatedAt: spotinst.TimeValue(output.Cluster.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Cluster.UpdatedAt),
		},
		Obj: output.Cluster,
	}

	return cluster, nil
}

func (x *wave) DeleteCluster(ctx context.Context, clusterID string, shouldDeleteOcean bool, forceDelete bool) error {
	log.Debugf("Deleting Wave cluster by ID: %s (shouldDeleteOcean: %t, forceDelete: %t)", clusterID, shouldDeleteOcean, forceDelete)

	input := &wavesdk.DeleteClusterInput{
		ClusterID:         spotinst.String(clusterID),
		ShouldDeleteOcean: spotinst.Bool(shouldDeleteOcean),
		ForceDelete:       spotinst.Bool(forceDelete),
	}

	_, err := x.svc.DeleteCluster(ctx, input)
	return err
}

func (x *wave) ListClusters(ctx context.Context) ([]*WaveCluster, error) {
	log.Debugf("Listing Wave clusters")

	output, err := x.svc.ListClusters(ctx, &wavesdk.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	clusters := make([]*WaveCluster, len(output.Clusters))
	for i, cluster := range output.Clusters {
		clusters[i] = &WaveCluster{
			TypeMeta: TypeMeta{
				Kind: typeOf(WaveCluster{}),
			},
			ObjectMeta: ObjectMeta{
				ID:        spotinst.StringValue(cluster.ID),
				Name:      spotinst.StringValue(cluster.ClusterIdentifier),
				CreatedAt: spotinst.TimeValue(cluster.CreatedAt),
				UpdatedAt: spotinst.TimeValue(cluster.UpdatedAt),
			},
			Obj: cluster,
		}
	}

	return clusters, nil
}
