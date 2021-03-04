package spot

import (
	"context"
	"fmt"

	wavesdk "github.com/spotinst/spotinst-sdk-go/service/wave"
	"github.com/spotinst/spotinst-sdk-go/spotinst"

	"github.com/spotinst/spotctl/internal/log"
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
		return nil, parseError(err)
	}

	return buildCluster(output.Cluster)
}

func (x *wave) DeleteCluster(ctx context.Context, clusterID string, shouldDeleteOcean bool) error {
	log.Debugf("Deleting Wave cluster by ID: %s (shouldDeleteOcean: %t)", clusterID, shouldDeleteOcean)

	input := &wavesdk.DeleteClusterInput{
		ClusterID:         spotinst.String(clusterID),
		ShouldDeleteOcean: spotinst.Bool(shouldDeleteOcean),
	}

	_, err := x.svc.DeleteCluster(ctx, input)
	return err
}

func (x *wave) ListClusters(ctx context.Context, filter *WaveClustersFilter) ([]*WaveCluster, error) {
	log.Debugf("Listing Wave clusters, filter: %+v", filter)

	input := &wavesdk.ListClustersInput{}
	if filter != nil {
		if filter.ClusterIdentifier != "" {
			input.ClusterIdentifier = spotinst.String(filter.ClusterIdentifier)
		}
		if filter.ClusterState != "" {
			input.ClusterState = spotinst.String(filter.ClusterState)
		}
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

func (x *wave) ListSparkApplications(ctx context.Context, filter *SparkApplicationsFilter) ([]*SparkApplication, error) {
	log.Debugf("Listing Spark applications, filter: %+v", filter)

	input := &wavesdk.ListSparkApplicationsInput{}
	if filter != nil {
		if filter.ClusterIdentifier != "" {
			input.ClusterIdentifier = spotinst.String(filter.ClusterIdentifier)
		}
		if filter.ApplicationState != "" {
			input.ApplicationState = spotinst.String(filter.ApplicationState)
		}
		if filter.Name != "" {
			input.Name = spotinst.String(filter.Name)
		}
		if filter.ApplicationID != "" {
			input.ApplicationID = spotinst.String(filter.ApplicationID)
		}
	}

	output, err := x.svc.ListSparkApplications(ctx, input)
	if err != nil {
		return nil, err
	}

	sparkApplications := make([]*SparkApplication, len(output.SparkApplications))
	for i, outputSparkApplication := range output.SparkApplications {
		sparkApplication, err := buildSparkApplication(outputSparkApplication)
		if err != nil {
			return nil, err
		}
		sparkApplications[i] = sparkApplication
	}

	return sparkApplications, nil
}

func (x *wave) GetSparkApplication(ctx context.Context, id string) (*SparkApplication, error) {
	log.Debugf("Getting Wave Spark application by ID: %s", id)

	input := &wavesdk.ReadSparkApplicationInput{
		ID: spotinst.String(id),
	}

	output, err := x.svc.ReadSparkApplication(ctx, input)
	if err != nil {
		return nil, err
	}

	return buildSparkApplication(output.SparkApplication)
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
					UID:             spotinst.StringValue(comp.UID),
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

func buildSparkApplication(sparkApplication *wavesdk.SparkApplication) (*SparkApplication, error) {
	if sparkApplication == nil {
		return nil, fmt.Errorf("spark application is nil")
	}

	return &SparkApplication{
		TypeMeta: TypeMeta{
			Kind: typeOf(SparkApplication{}),
		},
		ObjectMeta: ObjectMeta{
			ID:        spotinst.StringValue(sparkApplication.ID),
			Name:      spotinst.StringValue(sparkApplication.Name),
			CreatedAt: spotinst.TimeValue(sparkApplication.CreatedAt),
			UpdatedAt: spotinst.TimeValue(sparkApplication.UpdatedAt),
		},
		State:             spotinst.StringValue(sparkApplication.ApplicationState),
		ClusterIdentifier: spotinst.StringValue(sparkApplication.ClusterIdentifier),
		ApplicationID:     spotinst.StringValue(sparkApplication.ApplicationID),
		Obj:               sparkApplication,
	}, nil
}
