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

	return buildCluster(output.Cluster)
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

func (x *wave) ListClusters(ctx context.Context, clusterIdentifier string, state string) ([]*WaveCluster, error) {
	log.Debugf("Listing Wave clusters (clusterIdentifier: %q, state: %q)", clusterIdentifier, state)

	input := &wavesdk.ListClustersInput{}
	if clusterIdentifier != "" {
		input.ClusterIdentifier = spotinst.String(clusterIdentifier)
	}
	if state != "" {
		clusterState := wavesdk.ClusterState(state)
		input.ClusterState = &clusterState
	}

	output, err := x.svc.ListClusters(ctx, input)
	if err != nil {
		return nil, err
	}

	clusters := make([]*WaveCluster, len(output.Clusters))
	for i, outputCluster := range output.Clusters {
		cluster, err := buildCluster(outputCluster)
		if err != nil {
			return nil, err
		}
		clusters[i] = cluster
	}

	return clusters, nil
}

func buildCluster(cluster *wavesdk.Cluster) (*WaveCluster, error) {
	if cluster == nil {
		return nil, fmt.Errorf("cluster is nil")
	}

	var components []WaveComponent
	if cluster.Config != nil && len(cluster.Config.Components) > 0 {
		components = make([]WaveComponent, len(cluster.Config.Components))
		for i, comp := range cluster.Config.Components {
			if comp != nil {
				component := WaveComponent{
					Uid:             spotinst.StringValue(comp.Uid),
					Name:            spotinst.StringValue(comp.Name),
					OperatorVersion: spotinst.StringValue(comp.OperatorVersion),
					Version:         spotinst.StringValue(comp.Version),
					Properties:      comp.Properties,
					State:           spotinst.StringValue(comp.State),
				}
				components[i] = component
			}
		}
	}

	return &WaveCluster{
		TypeMeta: TypeMeta{
			Kind: typeOf(WaveCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(cluster.ID),
			Name:      spotinst.StringValue(cluster.ClusterIdentifier),
			CreatedAt: spotinst.TimeValue(cluster.CreatedAt),
			UpdatedAt: spotinst.TimeValue(cluster.UpdatedAt),
		},
		State:      spotinst.StringValue(cluster.State),
		Components: components,
		Obj:        cluster,
	}, nil
}
