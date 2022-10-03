package spot

import (
	"context"
	"strings"

	"github.com/spotinst/spotinst-sdk-go/service/ocean/spark"
	"github.com/spotinst/spotinst-sdk-go/spotinst"

	"github.com/spotinst/spotctl/internal/log"
)

type oceanSpark struct {
	svc spark.Service
}

func (x *oceanSpark) ListClusters(ctx context.Context, controllerClusterID string, state string) ([]*OceanSparkCluster, error) {
	log.Debugf("Listing Ocean Spark clusters")

	filter := &spark.ListClustersInput{}

	if controllerClusterID != "" {
		filter.ControllerClusterID = spotinst.String(controllerClusterID)
	}

	if state != "" {
		filter.ClusterState = spotinst.String(strings.ToUpper(state))
	}

	output, err := x.svc.ListClusters(ctx, filter)
	if err != nil {
		return nil, err
	}

	clusters := make([]*OceanSparkCluster, len(output.Clusters))
	for i, cluster := range output.Clusters {
		clusters[i] = &OceanSparkCluster{
			TypeMeta: TypeMeta{
				Kind: typeOf(OceanSparkCluster{}),
			},
			ObjectMeta: ObjectMeta{
				ID:        spotinst.StringValue(cluster.ID),
				Name:      spotinst.StringValue(cluster.ControllerClusterID),
				CreatedAt: spotinst.TimeValue(cluster.CreatedAt),
				UpdatedAt: spotinst.TimeValue(cluster.UpdatedAt),
			},
			OceanClusterID: spotinst.StringValue(cluster.OceanClusterID),
			State:          spotinst.StringValue(cluster.State),
			Obj:            cluster,
		}
	}

	return clusters, nil
}

func (x *oceanSpark) GetCluster(ctx context.Context, clusterID string) (*OceanSparkCluster, error) {
	log.Debugf("Getting Ocean Spark cluster by ID: %s", clusterID)

	input := &spark.ReadClusterInput{
		ClusterID: spotinst.String(clusterID),
	}

	output, err := x.svc.ReadCluster(ctx, input)
	if err != nil {
		return nil, err
	}

	cluster := &OceanSparkCluster{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanSparkCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Cluster.ID),
			Name:      spotinst.StringValue(output.Cluster.ControllerClusterID),
			CreatedAt: spotinst.TimeValue(output.Cluster.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Cluster.UpdatedAt),
		},
		Obj: output.Cluster,
	}

	return cluster, nil
}

func (x *oceanSpark) CreateCluster(ctx context.Context, oceanClusterID string) (*OceanSparkCluster, error) {
	log.Debugf("Creating Ocean Spark cluster on Ocean cluster: %s", oceanClusterID)

	input := &spark.CreateClusterInput{
		Cluster: &spark.CreateClusterRequest{
			OceanClusterID: spotinst.String(oceanClusterID),
		},
	}

	output, err := x.svc.CreateCluster(ctx, input)
	if err != nil {
		return nil, err
	}

	cluster := &OceanSparkCluster{
		TypeMeta: TypeMeta{
			Kind: typeOf(OceanSparkCluster{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(output.Cluster.ID),
			Name:      spotinst.StringValue(output.Cluster.ControllerClusterID),
			CreatedAt: spotinst.TimeValue(output.Cluster.CreatedAt),
			UpdatedAt: spotinst.TimeValue(output.Cluster.UpdatedAt),
		},
		Obj: output.Cluster,
	}

	return cluster, nil
}

func (x *oceanSpark) DeleteCluster(ctx context.Context, clusterID string) error {
	log.Debugf("Deleting Ocean Spark cluster: %s", clusterID)

	input := &spark.DeleteClusterInput{
		ClusterID: spotinst.String(clusterID),
	}

	_, err := x.svc.DeleteCluster(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
