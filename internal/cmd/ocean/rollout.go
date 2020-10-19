package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdRollout struct {
		cmd  *cobra.Command
		opts CmdRolloutOptions
	}

	CmdRolloutOptions struct {
		*CmdOptions
	}
)

func NewCmdRollout(opts *CmdOptions) *cobra.Command {
	return newCmdRollout(opts).cmd
}

func newCmdRollout(opts *CmdOptions) *CmdRollout {
	var cmd CmdRollout

	cmd.cmd = &cobra.Command{
		Use:           "rollout",
		Short:         "Manage the rollout of a resource",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdRollout) initSubCommands() {
	commands := []func(*CmdRolloutOptions) *cobra.Command{
		NewCmdRolloutStart,
		NewCmdRolloutStop,
		NewCmdRolloutStatus,
		NewCmdRolloutHistory,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdRolloutOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdRolloutOptions) Validate() error {
	return x.CmdOptions.Validate()
}
