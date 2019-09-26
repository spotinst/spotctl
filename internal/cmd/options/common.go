package options

import (
	"io"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/cmd/clients"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
)

// CommonOptions contains common options and helper methods.
type CommonOptions struct {
	// In, Out, and Err represent the respective data streams that the command
	// may act upon. They are attached directly to any sub-process of the executed
	// command.
	In       io.Reader
	Out, Err io.Writer

	// Clients represents a factory interface that creates instances of each
	// client type. For example, to create an instance of the cloud provider
	// client interface, call the following method Clients.CloudProvider().
	Clients clients.Factory

	// Noninteractive disables the interactive mode user interface by quieting the
	// configuration prompts.
	Noninteractive bool

	// DryRun configures the command to print the actions that would be executed,
	// without executing them.
	DryRun bool

	// Timeout configures the maximum duration before timing out the execution of
	// the command.
	Timeout time.Duration

	// Verbose enables verbose logging.
	Verbose bool

	// Profile configures the name of the credentials profile to use.
	// Defaults to `default`.
	Profile string

	// PprofProfile and PprofOutput enables collecting of runtime profiling data
	// for the command's process in the format expected by the pprof visualization
	// tool.
	//
	// PprofProfile configures the type of profile to capture:
	// 	- cpu
	// 	- heap
	// 	- goroutine
	// 	- threadcreate
	// 	- block
	// 	- mutex
	//
	// PprofOutput configures the path of the file to write the profile to.
	PprofProfile, PprofOutput string
}

func NewCommonOptions(in io.Reader, out, err io.Writer) *CommonOptions {
	return &CommonOptions{
		In:  in,
		Out: out,
		Err: err,
	}
}

func (x *CommonOptions) Init(flags *pflag.FlagSet) {
	x.initIOStreams()
	x.initClientsFactory()
	x.initDefaults()
	x.initFlags(flags)
}

func (x *CommonOptions) initDefaults() {
	x.PprofProfile = "none"
	x.PprofOutput = "profile.pprof"
	x.Profile = credentials.DefaultProfile()
	x.Timeout = time.Minute
}

func (x *CommonOptions) initFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&x.Profile,
		"profile",
		"p",
		x.Profile,
		"name of credentials profile to use")

	flags.StringVar(
		&x.PprofProfile,
		"pprof-profile",
		x.PprofProfile,
		"name of profile to capture (none|cpu|heap|goroutine|threadcreate|block|mutex)")

	flags.StringVar(
		&x.PprofOutput,
		"pprof-output",
		x.PprofOutput,
		"name of the file to write the profile to")

	flags.BoolVarP(
		&x.Verbose,
		"verbose",
		"v",
		x.Verbose,
		"enable verbose logging")

	flags.BoolVarP(
		&x.Noninteractive,
		"noninteractive",
		"n",
		x.Noninteractive,
		"disable interactive mode user interface")

	flags.BoolVarP(
		&x.DryRun,
		"dry-run",
		"d",
		x.DryRun,
		"only print the actions that would be executed, without executing them")

	flags.DurationVarP(
		&x.Timeout,
		"timeout",
		"t",
		x.Timeout,
		"maximum duration before timing out the execution")
}

func (x *CommonOptions) initIOStreams() {
	// Standard input.
	if x.In == nil {
		x.In = os.Stdin
	}

	// Standard output.
	if x.Out == nil {
		x.Out = os.Stdout
	}

	// Standard error.
	if x.Err == nil {
		x.Err = os.Stderr
	}
}

func (x *CommonOptions) initClientsFactory() {
	x.Clients = clients.NewFactory(x.In, x.Out, x.Err)
}

func (x *CommonOptions) Validate() error {
	// TODO(liran): Validate all options.
	return nil
}
