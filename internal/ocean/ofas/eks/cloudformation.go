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

type Stack = cloudformation.Stack
type ResourceType string

type stackCollection struct {
	clusterName string
	svc         *cloudformation.CloudFormation
}

const (
	ResourceTypeCluster   ResourceType = "cluster"
	ResourceTypeNodegroup ResourceType = "nodegroup"
	ResourceTypeUnknown   ResourceType = "unknown"

	StackStatusUnknown = "unknown"
)

func GetStacksForCluster(cloudProvider cloud.Provider, profile string, region string, clusterName string) ([]*Stack, error) {
	stackCollection, err := newStackCollection(cloudProvider, profile, region, clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not get stack collection, %w", err)
	}

	stacks, err := stackCollection.describeStacks()
	if err != nil {
		return nil, fmt.Errorf("could not describe stacks, %w", err)
	}

	return stacks, nil
}

func GroupStacksByResourceType(stacks []*Stack) map[ResourceType][]*Stack {
	res := make(map[ResourceType][]*Stack)

	for i := range stacks {
		stack := stacks[i]
		isClusterStack := isClusterStack(stack)
		isNodegroupStack := isNodegroupStack(stack)
		if isClusterStack && isNodegroupStack {
			res[ResourceTypeUnknown] = append(res[ResourceTypeUnknown], stack)
		} else if isClusterStack {
			res[ResourceTypeCluster] = append(res[ResourceTypeCluster], stack)
		} else if isNodegroupStack {
			res[ResourceTypeNodegroup] = append(res[ResourceTypeNodegroup], stack)
		} else {
			res[ResourceTypeUnknown] = append(res[ResourceTypeUnknown], stack)
		}
	}

	return res
}

func GroupStacksByStatus(stacks []*Stack) map[string][]*Stack {
	res := make(map[string][]*Stack)

	for i := range stacks {
		stack := stacks[i]
		if stack == nil {
			res[StackStatusUnknown] = append(res[StackStatusUnknown], stack)
			continue
		}
		if stack.StackStatus == nil {
			res[StackStatusUnknown] = append(res[StackStatusUnknown], stack)
			continue
		}
		status := *stack.StackStatus
		res[status] = append(res[status], stack)
	}

	return res
}

func isClusterStack(stack *Stack) bool {
	if stack == nil {
		return false
	}
	name := stack.StackName
	if name != nil && strings.HasSuffix(*name, "-cluster") {
		return true
	}
	return false
}

func isNodegroupStack(stack *Stack) bool {
	if stack == nil {
		return false
	}
	name := stack.StackName
	if name != nil && strings.Contains(*name, "nodegroup-ocean") {
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

// describeStack describes a cloudformation stack.
func (c *stackCollection) describeStack(i *Stack) (*Stack, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: i.StackName,
	}
	resp, err := c.svc.DescribeStacks(input)
	if err != nil {
		return nil, fmt.Errorf("describing CloudFormation stack %q: %w", *i.StackName, err)
	}
	return resp.Stacks[0], nil
}

// listStacksMatching gets all of CloudFormation stacks with names matching nameRegex.
func (c *stackCollection) listStacksMatching(nameRegex string, statusFilters ...string) ([]*Stack, error) {
	var (
		subErr error
		stack  *Stack
	)

	re, err := regexp.Compile(nameRegex)
	if err != nil {
		return nil, fmt.Errorf("cannot list stacks: %w", err)
	}
	input := &cloudformation.ListStacksInput{
		StackStatusFilter: defaultStackStatusFilter(),
	}
	if len(statusFilters) > 0 {
		input.StackStatusFilter = aws.StringSlice(statusFilters)
	}
	var stacks []*Stack
	pager := func(p *cloudformation.ListStacksOutput, _ bool) bool {
		for _, s := range p.StackSummaries {
			if re.MatchString(*s.StackName) {
				stack, subErr = c.describeStack(&Stack{
					StackName: s.StackName,
					StackId:   s.StackId,
				})
				if subErr != nil {
					return false
				}
				stacks = append(stacks, stack)
			}
		}
		return true
	}

	if err = c.svc.ListStacksPages(input, pager); err != nil {
		return nil, err
	}
	if subErr != nil {
		return nil, subErr
	}

	return stacks, nil
}

// describeStacks describes cloudformation stacks.
func (c *stackCollection) describeStacks() ([]*Stack, error) {
	log.Debugf("Describing stacks")

	stacks, err := c.listStacks()
	if err != nil {
		return nil, fmt.Errorf("describing CloudFormation stacks for %q: %w", c.clusterName, err)
	}

	if len(stacks) == 0 {
		log.Debugf("no eksctl-managed CloudFormation stacks found for %q", c.clusterName)
	}

	out := make([]*Stack, 0)
	for _, s := range stacks {
		if *s.StackStatus == cloudformation.StackStatusDeleteComplete {
			// Ignore deleted stacks
			continue
		}
		out = append(out, s)
	}

	return out, nil
}

func fmtStacksRegexForCluster(name string) string {
	const ourStackRegexFmt = "^(eksctl|EKS)-%s-((cluster|nodegroup-.+|addon-.+)|(VPC|ServiceRole|ControlPlane|DefaultNodeGroup))$"
	return fmt.Sprintf(ourStackRegexFmt, name)
}

func defaultStackStatusFilter() []*string {
	return aws.StringSlice(allNonDeletedStackStatuses())
}

func finalizedStackStatuses() []string {
	return []string{
		cloudformation.StackStatusDeleteComplete,
		cloudformation.StackStatusCreateComplete,
	}
}

func allNonDeletedStackStatuses() []string {
	return []string{
		// X StackStatusCreateInProgress,
		cloudformation.StackStatusCreateInProgress,
		// X StackStatusCreateFailed,
		cloudformation.StackStatusCreateFailed,
		// X StackStatusCreateComplete,
		cloudformation.StackStatusCreateComplete,
		// X StackStatusRollbackInProgress,
		cloudformation.StackStatusRollbackInProgress,
		// X StackStatusRollbackFailed,
		cloudformation.StackStatusRollbackFailed,
		// X StackStatusRollbackComplete,
		cloudformation.StackStatusRollbackComplete,
		// X StackStatusDeleteInProgress,
		cloudformation.StackStatusDeleteInProgress,
		// X StackStatusDeleteFailed,
		cloudformation.StackStatusDeleteFailed,
		// X StackStatusUpdateInProgress,
		cloudformation.StackStatusUpdateInProgress,
		// X StackStatusUpdateCompleteCleanupInProgress,
		cloudformation.StackStatusUpdateCompleteCleanupInProgress,
		// X StackStatusUpdateComplete,
		cloudformation.StackStatusUpdateComplete,
		// StackStatusUpdateRollbackInProgress,
		cloudformation.StackStatusUpdateRollbackInProgress,
		// StackStatusUpdateRollbackFailed,
		cloudformation.StackStatusUpdateRollbackFailed,
		// StackStatusUpdateRollbackCompleteCleanupInProgress,
		cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress,
		// StackStatusUpdateRollbackComplete,
		cloudformation.StackStatusUpdateRollbackComplete,
		// StackStatusReviewInProgress,
		cloudformation.StackStatusReviewInProgress,
	}

	/*








		StackStatusDeleteComplete,



		StackStatusUpdateFailed,





		StackStatusImportInProgress,
		StackStatusImportComplete,
		StackStatusImportRollbackInProgress,
		StackStatusImportRollbackFailed,
		StackStatusImportRollbackComplete,
	*/
}

func StacksToString(stacks []*Stack) []string {
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
