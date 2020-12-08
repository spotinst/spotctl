package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/log"
)

func (c *Cloud) GetInstanceProfile(ctx context.Context, profileName string) (*cloud.InstanceProfile, error) {
	log.Debugf("Getting IAM instance profile %q", profileName)

	svc := iam.New(c.session)
	input := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
	}
	output, err := svc.GetInstanceProfileWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	prof := &cloud.InstanceProfile{
		ID:   aws.StringValue(output.InstanceProfile.InstanceProfileId),
		Name: aws.StringValue(output.InstanceProfile.InstanceProfileName),
		ARN:  aws.StringValue(output.InstanceProfile.Arn),
	}

	if roles := output.InstanceProfile.Roles; len(roles) > 0 {
		prof.Roles = make([]*cloud.Role, len(roles))
		for i, role := range roles {
			prof.Roles[i] = &cloud.Role{
				ID:   aws.StringValue(role.RoleId),
				Name: aws.StringValue(role.RoleName),
				ARN:  aws.StringValue(role.Arn),
			}
		}
	}

	return prof, nil
}

func (c *Cloud) AttachRolePolicy(ctx context.Context, roleName, policyARN string) error {
	log.Debugf("Attaching IAM policy %q to role %q", policyARN, roleName)

	svc := iam.New(c.session)
	input := &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyARN),
	}
	_, err := svc.AttachRolePolicyWithContext(ctx, input)

	return err
}
