package dep

type InstallOptions struct {
	// BinaryDir specifies the binary directory that the manager should download
	// and install the binary to.
	BinaryDir string

	// Noninteractive disables the interactive mode user interface by quieting the
	// configuration prompts.
	Noninteractive bool

	// DryRun configures the dependency manager to print the actions that would
	// be executed, without executing them.
	DryRun bool
}

// InstallOption allows specifying various settings configurable by the dependency
// manager for overriding the defaults used when calling the `Install` method.
type InstallOption func(*InstallOptions)

// WithBinaryDir sets the binary directory.
func WithBinaryDir(path string) InstallOption {
	return func(opts *InstallOptions) {
		opts.BinaryDir = path
	}
}

// WithNoninteractive toggles the noninteractive mode on/off.
func WithNoninteractive(value bool) InstallOption {
	return func(opts *InstallOptions) {
		opts.Noninteractive = value
	}
}

// WithDryRun toggles the dry-run mode on/off.
func WithDryRun(value bool) InstallOption {
	return func(opts *InstallOptions) {
		opts.DryRun = value
	}
}

func initDefaultOptions() *InstallOptions {
	return &InstallOptions{
		BinaryDir:      DefaultBinaryDir(),
		Noninteractive: false,
		DryRun:         false,
	}
}
