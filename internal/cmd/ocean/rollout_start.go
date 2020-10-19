package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdRolloutStart struct {
		cmd  *cobra.Command
		opts CmdRolloutStartOptions
	}

	CmdRolloutStartOptions struct {
		*CmdRolloutOptions
	}
)

func NewCmdRolloutStart(opts *CmdRolloutOptions) *cobra.Command {
	return newCmdRolloutStart(opts).cmd
}

func newCmdRolloutStart(opts *CmdRolloutOptions) *CmdRolloutStart {
	var cmd CmdRolloutStart

	cmd.cmd = &cobra.Command{
		Use:           "start",
		Short:         "Start a new rollout",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdRolloutStart) initSubCommands() {
	commands := []func(*CmdRolloutStartOptions) *cobra.Command{
		NewCmdRolloutStartKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdRolloutStartOptions) Init(fs *pflag.FlagSet, opts *CmdRolloutOptions) {
	x.CmdRolloutOptions = opts
}

func (x *CmdRolloutStartOptions) Validate() error {
	return x.CmdOptions.Validate()
}
