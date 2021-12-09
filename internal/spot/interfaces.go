package spot

import (
	"context"
	"errors"

	"github.com/spf13/pflag"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("spot: not implemented")

type (
	// Client the interface of the Spot API.
	Client interface {
		// Accounts returns an instance of Accounts interface.
		Accounts() AccountsInterface

		// Services returns an instance of Services interface.
		Services() ServicesInterface
	}

	// AccountsInterface defines the interface of the Spot Accounts API.
	AccountsInterface interface {
		// ListAccounts returns a list of Spot accounts.
		ListAccounts(ctx context.Context) ([]*Account, error)
	}

	// ServicesInterface defines the interface of the Spot Services API.
	ServicesInterface interface {
		// Ocean returns an instance of Ocean interface by cloud provider and
		// orchestrator names.
		Ocean(provider CloudProviderName, orchestrator OrchestratorName) (OceanInterface, error)
	}

	// OceanInterface defines the interface of the Spot Ocean API.
	OceanInterface interface {
		// NewClusterBuilder returns new instance of OceanClusterBuilder
		// interface for defining fresh Ocean cluster.
		NewClusterBuilder(fs *pflag.FlagSet, opts *OceanClusterOptions) OceanClusterBuilder

		// NewLaunchSpecBuilder returns new instance of OceanLaunchSpecBuilder
		// interface for defining fresh Ocean launch spec.
		NewLaunchSpecBuilder(fs *pflag.FlagSet, opts *OceanLaunchSpecOptions) OceanLaunchSpecBuilder

		// NewRolloutBuilder returns new instance of OceanRolloutBuilder
		// interface for defining fresh Ocean rollout.
		NewRolloutBuilder(fs *pflag.FlagSet, opts *OceanRolloutOptions) OceanRolloutBuilder

		// ListClusters returns a list of Ocean clusters.
		ListClusters(ctx context.Context) ([]*OceanCluster, error)

		// ListLaunchSpecs returns a list of Ocean launch specs.
		ListLaunchSpecs(ctx context.Context) ([]*OceanLaunchSpec, error)

		// ListRollouts returns a list of Ocean rollouts.
		ListRollouts(ctx context.Context, clusterID string) ([]*OceanRollout, error)

		// GetCluster returns an Ocean cluster spec by ID.
		GetCluster(ctx context.Context, clusterID string) (*OceanCluster, error)

		// GetLaunchSpec returns an Ocean launch spec by ID.
		GetLaunchSpec(ctx context.Context, specID string) (*OceanLaunchSpec, error)

		// GetRollout returns an Ocean rollout by ID.
		GetRollout(ctx context.Context, clusterID, rolloutID string) (*OceanRollout, error)

		// CreateCluster creates a new Ocean cluster.
		CreateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error)

		// CreateLaunchSpec creates a new Ocean launch spec.
		CreateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error)

		// CreateRollout creates a new Ocean rollout.
		CreateRollout(ctx context.Context, rollout *OceanRollout) (*OceanRollout, error)

		// UpdateCluster updates an existing Ocean cluster by ID.
		UpdateCluster(ctx context.Context, cluster *OceanCluster) (*OceanCluster, error)

		// UpdateLaunchSpec updates an existing Ocean launch spec by ID.
		UpdateLaunchSpec(ctx context.Context, spec *OceanLaunchSpec) (*OceanLaunchSpec, error)

		// UpdateRollout updates an existing Ocean rollout by ID.
		UpdateRollout(ctx context.Context, rollout *OceanRollout) (*OceanRollout, error)

		// DeleteCluster deletes an Ocean cluster spec by ID.
		DeleteCluster(ctx context.Context, clusterID string) error

		// DeleteLaunchSpec deletes an Ocean launch spec by ID.
		DeleteLaunchSpec(ctx context.Context, specID string) error
	}

	// OceanClusterBuilder is the interface that every Ocean cluster
	// concrete implementation should obey.
	OceanClusterBuilder interface {
		Build() (*OceanCluster, error)
	}

	// OceanLaunchSpecBuilder is the interface that every Ocean launch spec
	// concrete implementation should obey.
	OceanLaunchSpecBuilder interface {
		Build() (*OceanLaunchSpec, error)
	}

	// OceanRolloutBuilder is the interface that every Ocean rollout
	// concrete implementation should obey.
	OceanRolloutBuilder interface {
		Build() (*OceanRollout, error)
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
