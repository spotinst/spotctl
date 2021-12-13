package eks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/spotinst/spotctl/internal/cloud"
	"github.com/spotinst/spotctl/internal/log"
)

type Stack = cloudformation.StackSummary

type stackCollection struct {
	clusterName string
	svc         *cloudformation.CloudFormation
}

func GetStacksForCluster(cloudProvider cloud.Provider, profile string, region string, clusterName string) ([]*Stack, error) {
	stackCollection, err := newStackCollection(cloudProvider, profile, region, clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not get stack collection, %w", err)
	}

	stacks, err := stackCollection.describeStacks()
	if err != nil {
		return nil, fmt.Errorf("could not describe stacks, %w", err)
	}

	log.Debugf("Stacks for cluster %q:\n%s", clusterName, strings.Join(StacksToStrings(stacks), "\n"))

	// Filter out deleted stacks
	stacks = FilterStacks(stacks, func(stack *Stack) bool {
		return !IsStackDeleted(stack)
	})

	return stacks, nil
}

func FilterStacks(stacks []*Stack, filter func(stack *Stack) bool) []*Stack {
	res := make([]*Stack, 0)
	for i := range stacks {
		stack := stacks[i]
		if filter(stack) {
			res = append(res, stack)
		}
	}
	return res
}

func IsStackCreated(stack *Stack) bool {
	return isStackOfStatus(stack, cloudformation.StackStatusCreateComplete)
}

func IsStackDeleted(stack *Stack) bool {
	return isStackOfStatus(stack, cloudformation.StackStatusDeleteComplete)
}

func isStackOfStatus(stack *Stack, status string) bool {
	if stack != nil && stack.StackStatus != nil && *stack.StackStatus == status {
		return true
	}
	return false
}

func IsClusterStack(stack *Stack) bool {
	if stack != nil && stack.StackName != nil && strings.HasSuffix(*stack.StackName, "-cluster") {
		return true
	}
	return false
}

func IsNodegroupStack(stack *Stack) bool {
	if stack != nil && stack.StackName != nil && strings.Contains(*stack.StackName, "nodegroup-ocean") {
		return true
	}
	return false
}

func newStackCollection(cloudProvider cloud.Provider, profile string, region string, clusterName string) (*stackCollection, error) {
	sess, err := cloudProvider.Session(region, profile)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud provider session, %w", err)
	}

	return &stackCollection{
		clusterName: clusterName,
		svc:         cloudformation.New(sess.(*session.Session)),
	}, nil
}

func (c *stackCollection) listStacks(statusFilters ...string) ([]*Stack, error) {
	return c.listStacksMatching(fmtStacksRegexForCluster(c.clusterName), statusFilters...)
}

// listStacksMatching gets all of CloudFormation stacks with names matching nameRegex.
func (c *stackCollection) listStacksMatching(nameRegex string, statusFilters ...string) ([]*Stack, error) {
	re, err := regexp.Compile(nameRegex)
	if err != nil {
		return nil, fmt.Errorf("cannot list stacks: %w", err)
	}

	input := &cloudformation.ListStacksInput{}
	if len(statusFilters) > 0 {
		input.StackStatusFilter = aws.StringSlice(statusFilters)
	}

	var stacks []*Stack
	pager := func(p *cloudformation.ListStacksOutput, _ bool) bool {
		for i := range p.StackSummaries {
			summary := p.StackSummaries[i]
			if summary != nil && summary.StackName != nil && re.MatchString(*summary.StackName) {
				stacks = append(stacks, summary)
			}
		}
		return true
	}

	if err = c.svc.ListStacksPages(input, pager); err != nil {
		return nil, err
	}

	return stacks, nil
}

// describeStacks describes cloudformation stacks.
func (c *stackCollection) describeStacks() ([]*Stack, error) {
	log.Debugf("Describing stacks")

	stacks, err := c.listStacks()
	if err != nil {
		return nil, fmt.Errorf("could not list CloudFormation stacks for %q: %w", c.clusterName, err)
	}

	if len(stacks) == 0 {
		log.Debugf("no eksctl-managed CloudFormation stacks found for %q", c.clusterName)
	}

	return stacks, nil
}

func fmtStacksRegexForCluster(name string) string {
	const ourStackRegexFmt = "^(eksctl|EKS)-%s-((cluster|nodegroup-.+|addon-.+)|(VPC|ServiceRole|ControlPlane|DefaultNodeGroup))$"
	return fmt.Sprintf(ourStackRegexFmt, name)
}

func StacksToStrings(stacks []*Stack) []string {
	out := make([]string, len(stacks))
	for i := range stacks {
		out[i] = stackToString(stacks[i])
	}
	return out
}

func stackToString(s *Stack) string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprintf("Stack - ID: %q, Status: %q", safeDerefString(s.StackId), safeDerefString(s.StackStatus))
}

func safeDerefString(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}
