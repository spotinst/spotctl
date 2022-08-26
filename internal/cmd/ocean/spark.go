package ocean

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/spot"
)

type (
	CmdSpark struct {
		cmd  *cobra.Command
		opts CmdSparkOptions
	}

	CmdSparkOptions struct {
		*CmdOptions
	}
)

func NewCmdSpark(opts *CmdOptions) *cobra.Command {
	return newCmdSpark(opts).cmd
}

func newCmdSpark(opts *CmdOptions) *CmdSpark {
	var cmd CmdSpark

	cmd.cmd = &cobra.Command{
		Use:           "spark",
		Short:         "Manage Ocean for Apache Spark resources",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.PersistentFlags(), opts)
	cmd.initSubCommands()

	return &cmd
}

func (x *CmdSpark) preRun(ctx context.Context) error {
	// Call to the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}
	return nil
}

func (x *CmdSpark) initSubCommands() {
	commands := []func(*CmdSparkOptions) *cobra.Command{
		NewCmdSparkCreate,
		NewCmdSparkGet,
		NewCmdSparkDescribe,
	}

	for _, cmd := range commands {
		x.cmd.AddCommand(cmd(&x.opts))
	}
}

func (x *CmdSparkOptions) Init(fs *pflag.FlagSet, opts *CmdOptions) {
	x.CmdOptions = opts
}

func (x *CmdSparkOptions) Validate() error {
	if x.CmdOptions.CloudProvider != spot.CloudProviderAWS {
		// Ocean for Apache Spark currently only supports AWS
		return fmt.Errorf("unsupported cloud provider: %q", x.CmdOptions.CloudProvider)
	}
	return x.CmdOptions.Validate()
}
