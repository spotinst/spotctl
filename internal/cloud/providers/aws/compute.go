package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/log"
)

func (c *Cloud) DescribeRegions(ctx context.Context) ([]*cloud.Region, error) {
	log.Debugf("Describing regions")

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
	log.Debugf("Describing availability zones")

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
	log.Debugf("Describing VPCs")

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
	log.Debugf("Describing subnets of VPC %q", vpcID)

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

func (c *Cloud) DescribeInstances(ctx context.Context, filters ...*cloud.Filter) ([]*cloud.Instance, error) {
	log.Debugf("Describing instances")

	svc := ec2.New(c.session)
	input := new(ec2.DescribeInstancesInput)

	for _, filter := range filters {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String(filter.Name),
			Values: aws.StringSlice(filter.Values),
		})
	}

	output, err := svc.DescribeInstancesWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	var instances []*cloud.Instance
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			i := &cloud.Instance{
				ID:   aws.StringValue(instance.InstanceId),
				Name: c.findResourceNameFromTags(instance.Tags),
			}
			if ip := instance.IamInstanceProfile; ip != nil {
				a, _ := arn.Parse(aws.StringValue(ip.Arn))
				i.InstanceProfile = &cloud.InstanceProfile{
					ID:  aws.StringValue(ip.Id),
					ARN: aws.StringValue(ip.Arn),

					// arn:partition:service:region:account-id:resource-type/resource-id
					// Ref: https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
					Name: strings.TrimPrefix(a.Resource, "instance-profile/"),
				}
			}
			instances = append(instances, i)
		}
	}

	return instances, nil
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
