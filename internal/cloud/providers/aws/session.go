package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spotinst/spotctl/internal/log"
)

func newSession(region, profile string) (*session.Session, error) {
	conf := &aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{Profile: profile},
			}),
		Logger: aws.LoggerFunc(func(args ...interface{}) {
			log.Debugf("%s", args...)
		}),
		LogLevel: aws.LogLevel(aws.LogDebug),
	}

	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *conf,
		Profile:           profile,
	}

	return session.NewSessionWithOptions(opts)
}
