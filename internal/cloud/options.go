package cloud

type ProviderOptions struct {
	// Credentials profile.
	Profile string

	// Region name.
	Region string

	// DryRun configures the provider to print the actions that would be executed,
	// without executing them.
	DryRun bool
}

// ProviderOption allows specifying various settings configurable by the provider.
type ProviderOption func(*ProviderOptions)

// WithProfile specifies credentials profile to use.
func WithProfile(profile string) ProviderOption {
	return func(opts *ProviderOptions) {
		opts.Profile = profile
	}
}

// WithRegion specifies the region to use.
func WithRegion(region string) ProviderOption {
	return func(opts *ProviderOptions) {
		opts.Region = region
	}
}

// DefaultProviderOptions returns the default provider options.
func DefaultProviderOptions() *ProviderOptions {
	return &ProviderOptions{}
}
