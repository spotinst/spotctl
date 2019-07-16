package spotinst

import (
	"os"
)

type ClientOptions struct {
	// User credentials.
	Token, Account string

	// BaseURL configures the default base URL of the Spotinst API.
	// Defaults to https://api.spotinst.io.
	BaseURL string

	// DryRun configures the client to print the actions that would be executed,
	// without executing them.
	DryRun bool
}

// ClientOption allows specifying various settings configurable by the client.
type ClientOption func(*ClientOptions)

// WithCredentials specifies static credentials.
func WithCredentials(token, account string) ClientOption {
	return func(opts *ClientOptions) {
		opts.Token = token
		opts.Account = account
	}
}

// WithBaseURL defines the base URL of the Spotinst API.
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
