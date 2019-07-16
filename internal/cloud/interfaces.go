package cloud

import "errors"

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("cloud: not implemented")

type (
	// ProviderName represents the name of a cloud provider.
	ProviderName string

	// Interface defines the interface that should be implemented by a cloud provider.
	Interface interface {
		// Name returns the cloud provider name.
		Name() ProviderName

		// DescribeRegions returns a list of regions.
		DescribeRegions() ([]*Region, error)

		// DescribeZones returns a list of availability zones within a region.
		DescribeZones(region string) ([]*Zone, error)

		// DescribeVPCs returns a list of VPCs within a region.
		DescribeVPCs(region string) ([]*VPC, error)

		// DescribeSubnets returns a list of subnets within a VPC.
		DescribeSubnets(region, vpcID string) ([]*Subnet, error)
	}

	// Describes a region.
	Region struct {
		Name string
	}

	// Describes a zone.
	Zone struct {
		Id   string
		Name string
	}

	// Describes a VPC.
	VPC struct {
		Id   string
		Name string
	}

	// Describes a subnet.
	Subnet struct {
		Id   string
		Name string
	}

	// Factory is a function that returns a Provider interface. An error is
	// returned if the cloud provider fails to initialize, nil otherwise.
	Factory func() (Interface, error)
)
