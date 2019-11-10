package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spotinst/spotctl/internal/cloud"
)

func (c *Cloud) DescribeRegions(ctx context.Context) ([]*cloud.Region, error) {
	svc := ec2.New(c.session)
	input := new(ec2.DescribeRegionsInput)
	output, err := svc.DescribeRegionsWithContext(ctx, input)
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

func (c *Cloud) DescribeZones(ctx context.Context) ([]*cloud.Zone, error) {
	svc := ec2.New(c.session)
	input := new(ec2.DescribeAvailabilityZonesInput)
	output, err := svc.DescribeAvailabilityZonesWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	var zones []*cloud.Zone
	for _, zone := range output.AvailabilityZones {
		if zone != nil && zone.ZoneName != nil {
			zones = append(zones, &cloud.Zone{
				ID:   aws.StringValue(zone.ZoneId),
				Name: aws.StringValue(zone.ZoneName),
			})
		}
	}

	return zones, nil
}

func (c *Cloud) DescribeVPCs(ctx context.Context) ([]*cloud.VPC, error) {
	svc := ec2.New(c.session)
	input := new(ec2.DescribeVpcsInput)
	output, err := svc.DescribeVpcsWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	var vpcs []*cloud.VPC
	for _, vpc := range output.Vpcs {
		if vpc != nil && vpc.VpcId != nil {
			vpcs = append(vpcs, &cloud.VPC{
				ID:   aws.StringValue(vpc.VpcId),
				Name: c.findResourceNameFromTags(vpc.Tags),
			})
		}
	}

	return vpcs, nil
}

func (c *Cloud) DescribeSubnets(ctx context.Context, vpcID string) ([]*cloud.Subnet, error) {
	svc := ec2.New(c.session)
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

	output, err := svc.DescribeSubnetsWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	var subnets []*cloud.Subnet
	for _, subnet := range output.Subnets {
		if subnet != nil && subnet.SubnetId != nil {
			subnets = append(subnets, &cloud.Subnet{
				ID:   aws.StringValue(subnet.SubnetId),
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
