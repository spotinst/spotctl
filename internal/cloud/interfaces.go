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
	Provider interface {
		// Name returns the cloud provider name.
		Name() ProviderName

		// Compute returns an instance of Compute interface.
		Compute() Compute

		// Storage returns an instance of Storage interface.
		Storage() Storage

		// IAM returns an instance of IAM interface.
		IAM() IAM
	}

	// Compute defines the interface of the Compute Services API.
	Compute interface {
		// DescribeRegions returns a list of regions.
		DescribeRegions(ctx context.Context) ([]*Region, error)

		// DescribeZones returns a list of availability zones within a region.
		DescribeZones(ctx context.Context) ([]*Zone, error)

		// DescribeVPCs returns a list of VPCs within a region.
		DescribeVPCs(ctx context.Context) ([]*VPC, error)

		// DescribeSubnets returns a list of subnets within a VPC.
		DescribeSubnets(ctx context.Context, vpcID string) ([]*Subnet, error)

		// DescribeInstances returns a list of instances.
		DescribeInstances(ctx context.Context, filters ...*Filter) ([]*Instance, error)
	}

	// Storage defines the interface of the Storage Services API.
	Storage interface {
		// GetBucket returns a bucket by name.
		GetBucket(ctx context.Context, bucket string) (*Bucket, error)

		// CreateBucket creates a new bucket.
		CreateBucket(ctx context.Context, bucket string) (*Bucket, error)
	}

	IAM interface {
		// GetInstanceProfile returns an instance profile by name.
		GetInstanceProfile(ctx context.Context, profileName string) (*InstanceProfile, error)

		// AttachRolePolicy attaches the specified policy to the specified IAM role.
		AttachRolePolicy(ctx context.Context, roleName, policyARN string) error
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

	// Describes an instance.
	Instance struct {
		ID              string
		Name            string
		InstanceProfile *InstanceProfile
	}

	// Describes an IAM instance profile.
	InstanceProfile struct {
		ID    string
		Name  string
		ARN   string
		Roles []*Role
	}

	// Describes an IAM role.
	Role struct {
		ID   string
		Name string
		ARN  string
	}

	// Filter represents a name and value pair that is used to return a more
	// specific list of results from a describe operation.
	Filter struct {
		Name   string
		Values []string
	}

	// Factory is a function that returns a Provider interface. An error is
	// returned if the cloud provider fails to initialize, nil otherwise.
	Factory func(options ...ProviderOption) (Provider, error)
)
