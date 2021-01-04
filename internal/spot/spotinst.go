package spot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/version"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	sdklog "github.com/spotinst/spotinst-sdk-go/spotinst/log"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

type api struct {
	client  *client.Client
	session *session.Session
	config  *spotinst.Config
}

func New(options ...ClientOption) Client {
	cfg := spotinst.DefaultConfig()

	// Initialize options.
	opts := initDefaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	// Configure the base URL.
	{
		if opts.BaseURL != "" {
			cfg.WithBaseURL(opts.BaseURL)
			log.Debugf("Configured base URL: %q", opts.BaseURL)
		}
	}

	// Configure credentials.
	{
		if opts.Profile != "" && opts.Profile != credentials.DefaultProfile() {
			cfg.WithCredentials(credentials.NewFileCredentials(opts.Profile, credentials.DefaultFilename()))
			log.Debugf("Configured file credentials")
		}
		if opts.Token != "" || opts.Account != "" {
			cfg.WithCredentials(credentials.NewStaticCredentials(opts.Token, opts.Account))
			log.Debugf("Configured static credentials")
		}
	}

	// Configure the SDK to use a dry-run mode.
	{
		if opts.DryRun {
			cfg.HTTPClient.Transport = new(roundTripperMock)
			cfg.WithCredentials(credentials.NewStaticCredentials("dry-run", "dry-run"))
			log.Debugf("Configured dry-run mode")
		}
	}

	// Configure the user agent.
	{
		userAgent := fmt.Sprintf("spotctl/%s", version.String())
		cfg.WithUserAgent(userAgent)
		log.Debugf("Configured user agent: %q", userAgent)
	}

	// Configure the logger.
	{
		cfg.WithLogger(sdklog.LoggerFunc(func(format string, args ...interface{}) {
			log.Debugf(format, args...)
		}))
	}

	return &api{
		client:  client.New(cfg),
		session: session.New(cfg),
		config:  cfg,
	}
}

func (x *api) Accounts() AccountsInterface {
	return &apiAccounts{client: x.client}
}

func (x *api) Services() ServicesInterface {
	return &apiServices{session: x.session}
}

type roundTripperMock struct{}

func (x *roundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		Request:    req,
	}, nil
}
