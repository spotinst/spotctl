package spot

import (
	"os"
)

type ClientOptions struct {
	// Credentials profile. Defaults to `default`.
	Profile string

	// Static user credentials.
	Token, Account string

	// BaseURL configures the default base URL of the Spot API.
	// Defaults to `https://api.spotinst.io`.
	BaseURL string

	// DryRun configures the client to print the actions that would be executed,
	// without executing them.
	DryRun bool
}

// ClientOption allows specifying various settings configurable by the client.
type ClientOption func(*ClientOptions)

// WithCredentialsProfile specifies credentials profile to use.
func WithCredentialsProfile(profile string) ClientOption {
	return func(opts *ClientOptions) {
		opts.Profile = profile
	}
}

// WithCredentialsStatic specifies static credentials.
func WithCredentialsStatic(token, account string) ClientOption {
	return func(opts *ClientOptions) {
		opts.Token = token
		opts.Account = account
	}
}

// WithBaseURL defines the base URL of the Spot API.
func WithBaseURL(url string) ClientOption {
	return func(opts *ClientOptions) {
		opts.BaseURL = url
	}
}

// WithDryRun toggles the dry-run mode on/off.
func WithDryRun(value bool) ClientOption {
	return func(opts *ClientOptions) {
		opts.DryRun = value
	}
}

func initDefaultOptions() *ClientOptions {
	return &ClientOptions{
		BaseURL: os.Getenv("SPOTINST_BASE_URL"),
	}
}
