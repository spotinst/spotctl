package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/spf13/cobra"
	"github.com/spotinst/spotctl/internal/cmd/completion"
	"github.com/spotinst/spotctl/internal/cmd/configure"
	"github.com/spotinst/spotctl/internal/cmd/ocean"
	"github.com/spotinst/spotctl/internal/cmd/options"
	"github.com/spotinst/spotctl/internal/cmd/version"
	"github.com/spotinst/spotctl/internal/cmd/wave"
	"github.com/spotinst/spotctl/internal/log"

	_ "github.com/spotinst/spotctl/internal/cloud/providers"
	_ "github.com/spotinst/spotctl/internal/thirdparty/commands"
	_ "github.com/spotinst/spotctl/internal/writer/writers"
)

type (
	CmdRoot struct {
		cmd  *cobra.Command
		opts CmdRootOptions
	}

	CmdRootOptions struct {
		*options.CommonOptions
	}
)

func New(in io.Reader, out, err io.Writer) *cobra.Command {
	return newCmd(in, out, err).cmd
}

func newCmd(in io.Reader, out, err io.Writer) *CmdRoot {
	var cmd CmdRoot

	cmd.cmd = &cobra.Command{
		Use:   "spotctl",
		Short: `A unified command-line interface to manage your Spot resources`,
		Long: `
A unified command-line interface to manage your Spot by NetApp resources. 
See the home page (https://github.com/spotinst/spotctl) for installation, 
usage, documentation, changelog and configuration walkthroughs.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return cmd.postRun(context.Background())
		},
	}

	cmd.initOptions(in, out, err)
	cmd.initUsage()
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdRoot) initOptions(in io.Reader, out, err io.Writer) {
	x.opts.CommonOptions = options.NewCommonOptions(in, out, err)
	x.opts.CommonOptions.Init(x.cmd.PersistentFlags())
}

func (x *CmdRoot) initUsage() {
	cobra.AddTemplateFunc("showCommands", func(cmd *cobra.Command) bool {
		return cmd.CalledAs() != "options"
	})

	cobra.AddTemplateFunc("showLocalFlags", func(cmd *cobra.Command) bool {
		// Don't show local flags (which are the global ones on the root) on
		// "spotinst" (=x.cmd.Use) and "spotinst help" (which shows the global help).
		return cmd.CalledAs() != x.cmd.Use && cmd.CalledAs() != ""
	})

	cobra.AddTemplateFunc("showGlobalFlags", func(cmd *cobra.Command) bool {
		return cmd.CalledAs() == "options"
	})

	// Set an alternative usage template.
	x.cmd.SetUsageTemplate(usageTemplate)
}

func (x *CmdRoot) initSubCommands() {
	commands := []func(*options.CommonOptions) *cobra.Command{
		// Resource management commands.
		ocean.NewCmd,
		wave.NewCmd,

		// Settings commands.
		completion.NewCmd,
		configure.NewCmd,

		// Other commands.
		options.NewCmd,
		version.NewCmd,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(x.opts.CommonOptions))
	}
}

func (x *CmdRoot) preRun(ctx context.Context) error {
	fns := []func() error{
		x.initLogger,
		x.initProfiling,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (x *CmdRoot) postRun(ctx context.Context) error {
	fns := []func() error{
		x.flushProfiling,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (x *CmdRoot) initLogger() error {
	// Logger options.
	logOpts := []log.LoggerOption{
		log.WithOutput(x.opts.Out),
		log.WithFormat(log.FormatText),
	}

	// Logger verbosity level.
	logLevel := log.LevelInfo
	if x.opts.Verbose {
		logLevel = log.LevelDebug
	}
	logOpts = append(logOpts, log.WithLevel(logLevel))

	// Initialize the default logger.
	log.InitDefaultLogger(logOpts...)

	return nil
}

func (x *CmdRoot) initProfiling() error {
	switch x.opts.PprofProfile {
	case "none":
		return nil
	case "cpu":
		f, err := os.Create(x.opts.PprofOutput)
		if err != nil {
			return err
		}
		return pprof.StartCPUProfile(f)
	// Block and mutex profiles need a call to Set{Block,Mutex}ProfileRate to
	// output anything. We choose to sample all events.
	case "block":
		runtime.SetBlockProfileRate(1)
		return nil
	case "mutex":
		runtime.SetMutexProfileFraction(1)
		return nil
	default:
		// Check the profile name is valid.
		if profile := pprof.Lookup(x.opts.PprofProfile); profile == nil {
			return fmt.Errorf("unknown profile %q", x.opts.PprofProfile)
		}
	}

	return nil
}

func (x *CmdRoot) flushProfiling() error {
	switch x.opts.PprofProfile {
	case "none":
		return nil
	case "cpu":
		pprof.StopCPUProfile()
	case "heap":
		runtime.GC()
		fallthrough
	default:
		profile := pprof.Lookup(x.opts.PprofProfile)
		if profile == nil {
			return nil
		}
		f, err := os.Create(x.opts.PprofOutput)
		if err != nil {
			return err
		}
		_ = profile.WriteTo(f, 0)
	}

	return nil
}
