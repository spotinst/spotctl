package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdRolloutHistory struct {
		cmd  *cobra.Command
		opts CmdRolloutHistoryOptions
	}

	CmdRolloutHistoryOptions struct {
		*CmdRolloutOptions
	}
)

func NewCmdRolloutHistory(opts *CmdRolloutOptions) *cobra.Command {
	return newCmdRolloutHistory(opts).cmd
}

func newCmdRolloutHistory(opts *CmdRolloutOptions) *CmdRolloutHistory {
	var cmd CmdRolloutHistory

	cmd.cmd = &cobra.Command{
		Use:           "history",
		Short:         "View rollout history",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdRolloutHistory) initSubCommands() {
	commands := []func(*CmdRolloutHistoryOptions) *cobra.Command{
		NewCmdRolloutHistoryKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdRolloutHistoryOptions) Init(fs *pflag.FlagSet, opts *CmdRolloutOptions) {
	x.CmdRolloutOptions = opts
}

func (x *CmdRolloutHistoryOptions) Validate() error {
	return x.CmdOptions.Validate()
}
