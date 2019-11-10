package cloud

import (
	"context"
	"errors"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("cloud: not implemented")

type (
	// ProviderName represents the name of a cloud provider.
	ProviderName string

	// Interface defines the interface that should be implemented by a cloud provider.
	Interface interface {
		// Name returns the cloud provider name.
		Name() ProviderName

		// Compute returns an instance of Compute interface.
		Compute() ComputeInterface

		// Storage returns an instance of Storage interface.
		Storage() StorageInterface
	}

	// ComputeInterface defines the interface of the Compute Services API.
	ComputeInterface interface {
		// DescribeRegions returns a list of regions.
		DescribeRegions(ctx context.Context) ([]*Region, error)

		// DescribeZones returns a list of availability zones within a region.
		DescribeZones(ctx context.Context) ([]*Zone, error)

		// DescribeVPCs returns a list of VPCs within a region.
		DescribeVPCs(ctx context.Context) ([]*VPC, error)

		// DescribeSubnets returns a list of subnets within a VPC.
		DescribeSubnets(ctx context.Context, vpcID string) ([]*Subnet, error)
	}

	// StorageInterface defines the interface of the Storage Services API.
	StorageInterface interface {
		// GetBucket returns a bucket by name.
		GetBucket(ctx context.Context, bucket string) (*Bucket, error)

		// CreateBucket creates a new bucket.
		CreateBucket(ctx context.Context, bucket string) (*Bucket, error)
	}

	// Describes a region.
	Region struct {
		Name string
	}

	// Describes a zone.
	Zone struct {
		ID   string
		Name string
	}

	// Describes a VPC.
	VPC struct {
		ID   string
		Name string
	}

	// Describes a subnet.
	Subnet struct {
		ID   string
		Name string
	}

	// Describes a bucket.
	Bucket struct {
		Name   string
		Region string
	}

	// Factory is a function that returns a Provider interface. An error is
	// returned if the cloud provider fails to initialize, nil otherwise.
	Factory func(options ...ProviderOption) (Interface, error)
)
