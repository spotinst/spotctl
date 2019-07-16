package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spotinst/spotinst-cli/internal/errors"
)

type (
	CmdDescribeLaunchSpecKubernetesAWS struct {
		cmd  *cobra.Command
		opts CmdDescribeLaunchSpecKubernetesAWSOptions
	}

	CmdDescribeLaunchSpecKubernetesAWSOptions struct {
		*CmdDescribeLaunchSpecKubernetesOptions
	}
)

func NewCmdDescribeLaunchSpecKubernetesAWS(opts *CmdDescribeLaunchSpecKubernetesOptions) *cobra.Command {
	return newCmdDescribeLaunchSpecKubernetesAWS(opts).cmd
}

func newCmdDescribeLaunchSpecKubernetesAWS(opts *CmdDescribeLaunchSpecKubernetesOptions) *CmdDescribeLaunchSpecKubernetesAWS {
	var cmd CmdDescribeLaunchSpecKubernetesAWS

	cmd.cmd = &cobra.Command{
		Use:           "aws",
		Short:         "Describe a Kubernetes launch spec on AWS",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	return &cmd
}

func (x *CmdDescribeLaunchSpecKubernetesAWS) Run(ctx context.Context) error {
	return errors.NotImplemented()
}
