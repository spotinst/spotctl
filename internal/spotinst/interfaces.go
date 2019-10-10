package spotinst

import (
	"context"
	"errors"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("spotinst: not implemented")

type (
	// Interface the interface of the Spotinst API.
	Interface interface {
		// Accounts returns an instance of Accounts interface.
		Accounts() AccountsInterface

		// Services returns an instance of Services interface.
		Services() ServicesInterface
	}

	// AccountsInterface defines the interface of the Spotinst Accounts API.
	AccountsInterface interface {
		// ListAccounts returns a list of Spotinst accounts.
		ListAccounts(ctx context.Context) ([]*Account, error)
	}

	// ServicesInterface defines the interface of the Spotinst Services API.
	ServicesInterface interface {
		// Ocean returns an instance of Ocean interface by cloud provider and
		// orchestrator names.
		Ocean(provider CloudProviderName, orchestrator OrchestratorName) (OceanInterface, error)
	}

	// OceanInterface defines the interface of the Spotinst Ocean API.
	OceanInterface interface {
		// ListClusters returns a list of Ocean clusters.
		ListClusters(ctx context.Context) ([]*OceanCluster, error)

		// ListLaunchSpecs returns a list of Ocean launch specs.
		ListLaunchSpecs(ctx context.Context) ([]*OceanLaunchSpec, error)

		// GetCluster returns an Ocean cluster spec by ID.
		GetCluster(ctx context.Context, clusterID string) (*OceanCluster, error)

		// GetLaunchSpec returns an Ocean launch spec by ID.
		GetLaunchSpec(ctx context.Context, specID string) (*OceanLaunchSpec, error)

		// CreateCluster creates a new Ocean cluster.
		CreateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error)

		// CreateLaunchSpec creates a new Ocean launch spec.
		CreateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error)

		// UpdateCluster updates an existing Ocean cluster by ID.
		UpdateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error)

		// UpdateLaunchSpec updates an existing Ocean launch spec by ID.
		UpdateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error)

		// DeleteCluster deletes an Ocean cluster spec by ID.
		DeleteCluster(ctx context.Context, clusterID string) error

		// DeleteLaunchSpec deletes an Ocean launch spec by ID.
		DeleteLaunchSpec(ctx context.Context, specID string) error
	}

	// CloudProviderName represents the name of a cloud provider.
	CloudProviderName string

	// OrchestratorName represents the name of a container orchestrator.
	OrchestratorName string
)

// Cloud Providers.
const (
	CloudProviderAWS   CloudProviderName = "aws"
	CloudProviderGCP   CloudProviderName = "gcp"
	CloudProviderAzure CloudProviderName = "azure"
)

// Orchestrators.
const (
	OrchestratorKubernetes OrchestratorName = "kubernetes"
	OrchestratorECS        OrchestratorName = "ecs"
)
