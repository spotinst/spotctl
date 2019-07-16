package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func newSession(profile, region string) (*session.Session, error) {
	conf := aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{Profile: profile},
			}),
	}

	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            conf,
		Profile:           profile,
	}

	return session.NewSessionWithOptions(opts)
}
