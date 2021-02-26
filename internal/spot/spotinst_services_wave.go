package spot

import (
	"context"
	"fmt"
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

	if output.Cluster == nil {
		return nil, fmt.Errorf("cluster is nil")
	}

	var components []WaveComponent
	if output.Cluster.Config != nil && len(output.Cluster.Config.Components) > 0 {
		components = make([]WaveComponent, len(output.Cluster.Config.Components))
		for i, outputComponent := range output.Cluster.Config.Components {
			if outputComponent != nil {
				component := WaveComponent{
					Uid:             spotinst.StringValue(outputComponent.Uid),
					Name:            spotinst.StringValue(outputComponent.Name),
					OperatorVersion: spotinst.StringValue(outputComponent.OperatorVersion),
					Version:         spotinst.StringValue(outputComponent.Version),
					Properties:      outputComponent.Properties,
					State:           spotinst.StringValue(outputComponent.State),
				}
				components[i] = component
			}
		}
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
		Components: components,
		Obj:        output.Cluster,
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
