package options

import (
	"github.com/spf13/cobra"
)

type Cmd struct {
	cmd  *cobra.Command
	opts *CommonOptions
}

func NewCmd(opts *CommonOptions) *cobra.Command {
	return newCmd(opts).cmd
}

func newCmd(opts *CommonOptions) *Cmd {
	var cmd Cmd

	cmd.cmd = &cobra.Command{
		Use:           "options",
		SilenceErrors: true,
		SilenceUsage:  true,
		Hidden:        true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	return &cmd
}
