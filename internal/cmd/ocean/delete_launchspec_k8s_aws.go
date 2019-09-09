package ocean

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/errors"
	"github.com/spotinst/spotinst-cli/internal/log"
	"github.com/spotinst/spotinst-cli/internal/survey"
	"github.com/spotinst/spotinst-cli/internal/thirdparty/commands/kops"
	"github.com/spotinst/spotinst-cli/internal/utils/flags"
)

type (
	CmdDeleteLaunchSpecKubernetesAWS struct {
		cmd  *cobra.Command
		opts CmdDeleteLaunchSpecKubernetesAWSOptions
	}

	CmdDeleteLaunchSpecKubernetesAWSOptions struct {
		*CmdDeleteLaunchSpecKubernetesOptions

		// Basic
		ClusterName string
		State       string

		// LaunchSpec
		SpecName string
	}
)

func NewCmdDeleteLaunchSpecKubernetesAWS(opts *CmdDeleteLaunchSpecKubernetesOptions) *cobra.Command {
	return newCmdDeleteLaunchSpecKubernetesAWS(opts).cmd
}

func newCmdDeleteLaunchSpecKubernetesAWS(opts *CmdDeleteLaunchSpecKubernetesOptions) *CmdDeleteLaunchSpecKubernetesAWS {
	var cmd CmdDeleteLaunchSpecKubernetesAWS

	cmd.cmd = &cobra.Command{
		Use:           "aws",
		Short:         "Delete an existing Ocean launch spec on AWS (using kops)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdDeleteLaunchSpecKubernetesAWS) Run(ctx context.Context) error {
	steps := []func(context.Context) error{
		x.survey,
		x.log,
		x.validate,
		x.run,
	}

	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (x *CmdDeleteLaunchSpecKubernetesAWS) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	log.Debugf("Starting survey...")
	surv, err := x.opts.Clients.NewSurvey()
	if err != nil {
		return err
	}

	// Cluster name.
	{
		if x.opts.ClusterName == "" {
			input := &survey.Input{
				Message: "Cluster name",
				Help: "Name must start with a lowercase letter followed by up to " +
					"39 lowercase letters, numbers, or hyphens, and cannot end with a hyphen",
				Required: true,
			}

			if x.opts.ClusterName, err = surv.InputString(input); err != nil {
				return err
			}
		}
	}

	// Spec name.
	{
		if x.opts.SpecName == "" {
			input := &survey.Input{
				Message: "LaunchSpec name",
				Help: "Name must start with a lowercase letter followed by up to " +
					"39 lowercase letters, numbers, or hyphens, and cannot end with a hyphen",
				Required: true,
			}

			if x.opts.SpecName, err = surv.InputString(input); err != nil {
				return err
			}
		}
	}

	// KOPS specific.
	{
		// State.
		{
			if x.opts.State == "" {
				input := &survey.Input{
					Message:  "Location of state store",
					Help:     "See: https://git.io/fjH5V",
					Required: true,
				}

				if x.opts.State, err = surv.InputString(input); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (x *CmdDeleteLaunchSpecKubernetesAWS) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdDeleteLaunchSpecKubernetesAWS) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdDeleteLaunchSpecKubernetesAWS) run(ctx context.Context) error {
	cmd, err := x.opts.Clients.NewCommand(kops.CommandName)
	if err != nil {
		return err
	}

	return cmd.Run(ctx, x.buildKopsArgs()...)
}

func (x *CmdDeleteLaunchSpecKubernetesAWS) buildKopsArgs() []string {
	log.Debugf("Building up command arguments")

	args := []string{
		"delete", "ig",
		"--state", x.opts.State,
		"--name", x.opts.ClusterName,
		x.opts.SpecName,
	}

	if !x.opts.DryRun {
		args = append(args, "--yes")
	}

	if x.opts.Verbose {
		args = append(args, "--logtostderr", "--v", "10")
	} else {
		args = append(args, "--logtostderr", "--v", "0")
	}

	return args
}

func (x *CmdDeleteLaunchSpecKubernetesAWSOptions) Init(flags *pflag.FlagSet, opts *CmdDeleteLaunchSpecKubernetesOptions) {
	x.initDefaults(opts)
	x.initFlags(flags)
}

func (x *CmdDeleteLaunchSpecKubernetesAWSOptions) initDefaults(opts *CmdDeleteLaunchSpecKubernetesOptions) {
	x.CmdDeleteLaunchSpecKubernetesOptions = opts
	x.ClusterName = os.Getenv("KOPS_CLUSTER_NAME")
	x.State = os.Getenv("KOPS_STATE_STORE")
}

func (x *CmdDeleteLaunchSpecKubernetesAWSOptions) initFlags(flags *pflag.FlagSet) {
	flags.StringVar(
		&x.ClusterName,
		"cluster-name",
		x.ClusterName,
		"name of the cluster")

	flags.StringVar(
		&x.SpecName,
		"spec-name",
		x.SpecName,
		"name of the launchspec")

	flags.StringVar(
		&x.State,
		"state",
		x.State,
		"s3 bucket used to store the state of the cluster")
}

func (x *CmdDeleteLaunchSpecKubernetesAWSOptions) Validate() error {
	if err := x.CmdDeleteLaunchSpecKubernetesOptions.Validate(); err != nil {
		return err
	}

	if x.State == "" {
		return errors.Required("state")
	}

	if x.ClusterName == "" {
		return errors.Required("cluster-name")
	}

	if x.SpecName == "" {
		return errors.Required("spec-name")
	}

	return nil
}
