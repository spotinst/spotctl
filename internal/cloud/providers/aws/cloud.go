package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spotinst/spotctl/internal/cloud"
)

// CloudProviderName is the name of this cloud provider.
const CloudProviderName cloud.ProviderName = "aws"

func init() {
	cloud.Register(CloudProviderName, factory)
}

func factory(options ...cloud.ProviderOption) (cloud.Provider, error) {
	opts := cloud.DefaultProviderOptions()

	for _, opt := range options {
		opt(opts)
	}

	sess, err := newSession(opts.Region, opts.Profile)
	if err != nil {
		return nil, err
	}

	return &Cloud{opts, sess}, nil
}

type Cloud struct {
	options *cloud.ProviderOptions
	session *session.Session
}

func (c *Cloud) Name() cloud.ProviderName { return CloudProviderName }
func (c *Cloud) Compute() cloud.Compute   { return c }
func (c *Cloud) Storage() cloud.Storage   { return c }
func (c *Cloud) IAM() cloud.IAM           { return c }
func (c *Cloud) Session(region, profile string) (interface{}, error) {
	return newSession(region, profile)
}
