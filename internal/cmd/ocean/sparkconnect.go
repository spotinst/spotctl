package ocean

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/spot"
)

type (
	CmdSparkConnect struct {
		cmd  *cobra.Command
		opts CmdSparkConnectOptions
	}

	CmdSparkConnectOptions struct {
		*CmdOptions

		// options
		NoHeaders bool
		Output    string
	}
)

func NewCmdSparkConnect(opts *CmdOptions) *cobra.Command {
	return newCmdSparkConnect(opts).cmd
}

func newCmdSparkConnect(opts *CmdOptions) *CmdSparkConnect {
	var cmd CmdSparkConnect

	cmd.cmd = &cobra.Command{
		Use:           "sparkconnect",
		Short:         "Spark Connect to Ocean Spark",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	//cmd.initSubCommands()

	return &cmd
}

func (x *CmdSparkConnect) preRun(ctx context.Context) error {
	// Call to the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}
	return nil
}

/*func (x *CmdSparkConnect) initSubCommands() {
	commands := []func(*CmdSparkConnectOptions) *cobra.Command{
		NewCmdSparkCreate,
		NewCmdSparkGet,
		NewCmdSparkDescribe,
		NewCmdSparkDelete,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}*/

func (x *CmdSparkConnectOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdSparkConnectOptions) Validate() error {
	if x.CmdOptions.CloudProvider != spot.CloudProviderAWS {
		// Ocean for Apache Spark currently only supports AWS
		return fmt.Errorf("unsupported cloud provider: %q", x.CmdOptions.CloudProvider)
	}
	return x.CmdOptions.Validate()
}
