package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdRolloutStop struct {
		cmd  *cobra.Command
		opts CmdRolloutStopOptions
	}

	CmdRolloutStopOptions struct {
		*CmdRolloutOptions
	}
)

func NewCmdRolloutStop(opts *CmdRolloutOptions) *cobra.Command {
	return newCmdRolloutStop(opts).cmd
}

func newCmdRolloutStop(opts *CmdRolloutOptions) *CmdRolloutStop {
	var cmd CmdRolloutStop

	cmd.cmd = &cobra.Command{
		Use:           "stop",
		Short:         "Stop an in-progress rollout",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdRolloutStop) initSubCommands() {
	commands := []func(*CmdRolloutStopOptions) *cobra.Command{
		NewCmdRolloutStopKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdRolloutStopOptions) Init(fs *pflag.FlagSet, opts *CmdRolloutOptions) {
	x.CmdRolloutOptions = opts
}

func (x *CmdRolloutStopOptions) Validate() error {
	return x.CmdOptions.Validate()
}
