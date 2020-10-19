package ocean

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	CmdRolloutStatus struct {
		cmd  *cobra.Command
		opts CmdRolloutStatusOptions
	}

	CmdRolloutStatusOptions struct {
		*CmdRolloutOptions
	}
)

func NewCmdRolloutStatus(opts *CmdRolloutOptions) *cobra.Command {
	return newCmdRolloutStatus(opts).cmd
}

func newCmdRolloutStatus(opts *CmdRolloutOptions) *CmdRolloutStatus {
	var cmd CmdRolloutStatus

	cmd.cmd = &cobra.Command{
		Use:           "status",
		Short:         "Show the status of a rollout",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdRolloutStatus) initSubCommands() {
	commands := []func(*CmdRolloutStatusOptions) *cobra.Command{
		NewCmdRolloutStatusKubernetes,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdRolloutStatusOptions) Init(fs *pflag.FlagSet, opts *CmdRolloutOptions) {
	x.CmdRolloutOptions = opts
}

func (x *CmdRolloutStatusOptions) Validate() error {
	return x.CmdOptions.Validate()
}
