package wave

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

type Cluster struct {
	ID                *string      `json:"id,omitempty"`
	ClusterIdentifier *string      `json:"clusterIdentifier,omitempty"`
	Environment       *Environment `json:"environment,omitempty"`
	Config            *Config      `json:"config,omitempty"`
	State             *string      `json:"state,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type Environment struct {
	OperatorVersion         *string `json:"operatorVersion,omitempty"`
	CertManagerDeployed     *bool   `json:"certManagerDeployed,omitempty"`
	K8sClusterProvisioned   *bool   `json:"k8sClusterProvisioned,omitempty"`
	OceanClusterProvisioned *bool   `json:"oceanClusterProvisioned,omitempty"`
	EnvironmentNamespace    *string `json:"environmentNamespace,omitempty"`
	OceanClusterId          *string `json:"oceanClusterId,omitempty"`
}

type Config struct {
	Components []*Component `json:"components,omitempty"`
}

type Component struct {
	Uid             *string           `json:"uid,omitempty"`
	Name            *string           `json:"name,omitempty"`
	OperatorVersion *string           `json:"operatorVersion,omitempty"`
	Version         *string           `json:"version,omitempty"`
	Properties      map[string]string `json:"properties,omitempty"`
	State           *string           `json:"state,omitempty"`
}

type SparkApplication struct {
	ID                *string `json:"id,omitempty"`
	ApplicationID     *string `json:"applicationId,omitempty"`
	ClusterIdentifier *string `json:"clusterIdentifier,omitempty"`
	Name              *string `json:"name,omitempty"`
	Namespace         *string `json:"namespace,omitempty"`
	Heritage          *string `json:"heritage,omitempty"`
	ApplicationState  *string `json:"applicationState,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type ListClustersInput struct {
	ClusterIdentifier *string `json:"clusterIdentifier,omitempty"`
	ClusterState      *string `json:"clusterState,omitempty"`
}

type ListClustersOutput struct {
	Clusters []*Cluster `json:"clusters,omitempty"`
}

type ReadClusterInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type ReadClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type DeleteClusterInput struct {
	ClusterID         *string `json:"clusterId,omitempty"`
	ShouldDeleteOcean *bool   `json:"shouldDeleteOcean,omitempty"`
}

type DeleteClusterOutput struct{}

type ListSparkApplicationsInput struct {
	ClusterIdentifier *string `json:"clusterIdentifier,omitempty"`
	Name              *string `json:"name,omitempty"`
	Namespace         *string `json:"namespace,omitempty"`
	ApplicationId     *string `json:"applicationId,omitempty"`
	ApplicationState  *string `json:"applicationState,omitempty"`
	Heritage          *string `json:"heritage,omitempty"`
}

type ListSparkApplicationsOutput struct {
	SparkApplications []*SparkApplication `json:"sparkApplications,omitempty"`
}

type ReadSparkApplicationInput struct {
	ID *string `json:"id,omitempty"`
}

type ReadSparkApplicationOutput struct {
	SparkApplication *SparkApplication `json:"sparkApplication,omitempty"`
}

func (s *ServiceOp) ListClusters(ctx context.Context, input *ListClustersInput) (*ListClustersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/wave/cluster")

	if input != nil {
		if input.ClusterIdentifier != nil {
			r.Params.Set("clusterIdentifier", spotinst.StringValue(input.ClusterIdentifier))
		}

		if input.ClusterState != nil {
			r.Params.Set("state", spotinst.StringValue(input.ClusterState))
		}
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	clusters, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListClustersOutput{Clusters: clusters}, nil
}

func (s *ServiceOp) ReadCluster(ctx context.Context, input *ReadClusterInput) (*ReadClusterOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	path, err := uritemplates.Expand("/wave/cluster/{clusterId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	clusters, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadClusterOutput)
	if len(clusters) > 0 {
		output.Cluster = clusters[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteCluster(ctx context.Context, input *DeleteClusterInput) (*DeleteClusterOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	path, err := uritemplates.Expand("/wave/cluster/{clusterId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)

	if input.ShouldDeleteOcean != nil {
		r.Params.Set("shouldDeleteOcean",
			strconv.FormatBool(spotinst.BoolValue(input.ShouldDeleteOcean)))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteClusterOutput{}, nil
}

func (s *ServiceOp) ListSparkApplications(ctx context.Context, input *ListSparkApplicationsInput) (*ListSparkApplicationsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/wave/spark/application")

	if input != nil {
		if input.ClusterIdentifier != nil {
			r.Params.Set("clusterIdentifier", spotinst.StringValue(input.ClusterIdentifier))
		}

		if input.Name != nil {
			r.Params.Set("name", spotinst.StringValue(input.Name))
		}

		if input.ApplicationId != nil {
			r.Params.Set("applicationId", spotinst.StringValue(input.ApplicationId))
		}

		if input.ApplicationState != nil {
			r.Params.Set("applicationState", spotinst.StringValue(input.ApplicationState))
		}

		if input.Heritage != nil {
			r.Params.Set("heritage", spotinst.StringValue(input.Heritage))
		}

		if input.Namespace != nil {
			r.Params.Set("namespace", spotinst.StringValue(input.Namespace))
		}
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sparkApplications, err := sparkApplicationsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListSparkApplicationsOutput{SparkApplications: sparkApplications}, nil
}

func (s *ServiceOp) ReadSparkApplication(ctx context.Context, input *ReadSparkApplicationInput) (*ReadSparkApplicationOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	path, err := uritemplates.Expand("/wave/spark/application/{id}", uritemplates.Values{
		"id": spotinst.StringValue(input.ID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sparkApplications, err := sparkApplicationsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadSparkApplicationOutput)
	if len(sparkApplications) > 0 {
		output.SparkApplication = sparkApplications[0]
	}

	return output, nil
}

func clustersFromHttpResponse(resp *http.Response) ([]*Cluster, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return clustersFromJSON(body)
}

func clustersFromJSON(in []byte) ([]*Cluster, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Cluster, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := clusterFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func clusterFromJSON(in []byte) (*Cluster, error) {
	b := new(Cluster)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func sparkApplicationsFromHttpResponse(resp *http.Response) ([]*SparkApplication, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return sparkApplicationsFromJSON(body)
}

func sparkApplicationsFromJSON(in []byte) ([]*SparkApplication, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*SparkApplication, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := sparkApplicationFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func sparkApplicationFromJSON(in []byte) (*SparkApplication, error) {
	b := new(SparkApplication)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}
