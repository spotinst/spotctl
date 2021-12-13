package eks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	ekssdk "github.com/aws/aws-sdk-go/service/eks"

	"github.com/spotinst/spotctl/internal/cloud"
)

type ErrClusterNotFound struct {
	wrappedError error
}

func (e ErrClusterNotFound) Error() string {
	if e.wrappedError != nil {
		return fmt.Sprintf("cluster not found: %s", e.wrappedError.Error())
	}
	return "cluster not found"
}

func (e ErrClusterNotFound) Unwrap() error {
	return e.wrappedError
}

func errClusterNotFound(err error) ErrClusterNotFound {
	return ErrClusterNotFound{
		wrappedError: err,
	}
}

func GetEKSCluster(cloudProvider cloud.Provider, profile string, region string, clusterName string) (*ekssdk.Cluster, error) {
	svc, err := newEKSService(cloudProvider, profile, region)
	if err != nil {
		return nil, fmt.Errorf("could not get eks service, %w", err)
	}

	describeOutput, err := svc.DescribeCluster(&ekssdk.DescribeClusterInput{Name: &clusterName})
	if err != nil {
		if awsError, ok := err.(awserr.Error); ok {
			if awsError.Code() == ekssdk.ErrCodeResourceNotFoundException {
				return nil, errClusterNotFound(err)
			}
		}
		return nil, fmt.Errorf("could not describe cluster, %w", err)
	}

	if describeOutput == nil {
		return nil, fmt.Errorf("describe output is nil")
	}

	return describeOutput.Cluster, nil
}

func newEKSService(cloudProvider cloud.Provider, profile string, region string) (*ekssdk.EKS, error) {
	sess, err := cloudProvider.Session(region, profile)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud provider session, %w", err)
	}

	return ekssdk.New(sess.(*session.Session)), nil
}
