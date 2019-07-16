package ocean

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spotinst/spotinst-cli/internal/errors"
)

type (
	CmdDescribeClusterKubernetesAWS struct {
		cmd  *cobra.Command
		opts CmdDescribeClusterKubernetesAWSOptions
	}

	CmdDescribeClusterKubernetesAWSOptions struct {
		*CmdDescribeClusterKubernetesOptions
	}
)

func NewCmdDescribeClusterKubernetesAWS(opts *CmdDescribeClusterKubernetesOptions) *cobra.Command {
	return newCmdDescribeClusterKubernetesAWS(opts).cmd
}

func newCmdDescribeClusterKubernetesAWS(opts *CmdDescribeClusterKubernetesOptions) *CmdDescribeClusterKubernetesAWS {
	var cmd CmdDescribeClusterKubernetesAWS

	cmd.cmd = &cobra.Command{
		Use:           "aws",
		Short:         "Describe a Kubernetes cluster on AWS",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	return &cmd
}

func (x *CmdDescribeClusterKubernetesAWS) Run(ctx context.Context) error {
	return errors.NotImplemented()
}
