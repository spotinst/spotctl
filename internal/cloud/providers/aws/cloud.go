package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spotinst/spotinst-cli/internal/cloud"
)

// CloudProviderName is the name of this cloud provider.
const CloudProviderName cloud.ProviderName = "aws"

func init() {
	cloud.Register(CloudProviderName, factory)
}

func factory() (cloud.Interface, error) {
	return &Cloud{}, nil
}

type Cloud struct{}

func (c *Cloud) Name() cloud.ProviderName {
	return CloudProviderName
}

func (c *Cloud) DescribeRegions() ([]*cloud.Region, error) {
	sess, err := newSession("", "")
	if err != nil {
		return nil, err
	}

	svc := ec2.New(sess)
	input := new(ec2.DescribeRegionsInput)
	output, err := svc.DescribeRegions(input)
	if err != nil {
		return nil, err
	}

	var regions []*cloud.Region
	for _, region := range output.Regions {
		if region != nil && region.RegionName != nil {
			regions = append(regions, &cloud.Region{
				Name: aws.StringValue(region.RegionName),
			})
		}
	}

	return regions, nil
}

func (c *Cloud) DescribeZones(region string) ([]*cloud.Zone, error) {
	sess, err := newSession("", region)
	if err != nil {
		return nil, err
	}

	svc := ec2.New(sess)
	input := new(ec2.DescribeAvailabilityZonesInput)
	output, err := svc.DescribeAvailabilityZones(input)
	if err != nil {
		return nil, err
	}

	var zones []*cloud.Zone
	for _, zone := range output.AvailabilityZones {
		if zone != nil && zone.ZoneName != nil {
			zones = append(zones, &cloud.Zone{
				Id:   aws.StringValue(zone.ZoneId),
				Name: aws.StringValue(zone.ZoneName),
			})
		}
	}

	return zones, nil
}

func (c *Cloud) DescribeVPCs(region string) ([]*cloud.VPC, error) {
	sess, err := newSession("", region)
	if err != nil {
		return nil, err
	}

	svc := ec2.New(sess)
	input := new(ec2.DescribeVpcsInput)
	output, err := svc.DescribeVpcs(input)
	if err != nil {
		return nil, err
	}

	var vpcs []*cloud.VPC
	for _, vpc := range output.Vpcs {
		if vpc != nil && vpc.VpcId != nil {
			vpcs = append(vpcs, &cloud.VPC{
				Id:   aws.StringValue(vpc.VpcId),
				Name: c.findResourceNameFromTags(vpc.Tags),
			})
		}
	}

	return vpcs, nil
}

func (c *Cloud) DescribeSubnets(region, vpcID string) ([]*cloud.Subnet, error) {
	sess, err := newSession("", region)
	if err != nil {
		return nil, err
	}

	svc := ec2.New(sess)
	input := new(ec2.DescribeSubnetsInput)

	if vpcID != "" {
		input.Filters = []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(vpcID),
				},
			},
		}
	}

	output, err := svc.DescribeSubnets(input)
	if err != nil {
		return nil, err
	}

	var subnets []*cloud.Subnet
	for _, subnet := range output.Subnets {
		if subnet != nil && subnet.SubnetId != nil {
			subnets = append(subnets, &cloud.Subnet{
				Id:   aws.StringValue(subnet.SubnetId),
				Name: c.findResourceNameFromTags(subnet.Tags),
			})
		}
	}

	return subnets, nil
}

func (c *Cloud) findResourceNameFromTags(tags []*ec2.Tag) string {
	name := "unknown"

	for _, tag := range tags {
		if tag != nil && tag.Key != nil && aws.StringValue(tag.Key) == "Name" && tag.Value != nil {
			name = aws.StringValue(tag.Value)
		}
	}

	return name
}
