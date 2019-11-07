package ocean

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/spotinst"
	"github.com/spotinst/spotctl/internal/utils/flags"
)

type (
	CmdEditLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdEditLaunchSpecKubernetesOptions
	}

	CmdEditLaunchSpecKubernetesOptions struct {
		*CmdEditLaunchSpecOptions

		LaunchSpecID string
	}
)

func NewCmdEditLaunchSpecKubernetes(opts *CmdEditLaunchSpecOptions) *cobra.Command {
	return newCmdEditLaunchSpecKubernetes(opts).cmd
}

func newCmdEditLaunchSpecKubernetes(opts *CmdEditLaunchSpecOptions) *CmdEditLaunchSpecKubernetes {
	var cmd CmdEditLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Edit a Kubernetes launch spec",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdEditLaunchSpecKubernetes) Run(ctx context.Context) error {
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

func (x *CmdEditLaunchSpecKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdEditLaunchSpecKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdEditLaunchSpecKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdEditLaunchSpecKubernetes) run(ctx context.Context) error {
	spotinstClientOpts := []spotinst.ClientOption{
		spotinst.WithCredentialsProfile(x.opts.Profile),
	}

	spotinstClient, err := x.opts.Clients.NewSpotinst(spotinstClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotinstClient.Services().Ocean(x.opts.CloudProvider, spotinst.OrchestratorKubernetes)
	if err != nil {
		return err
	}

	spec, err := oceanClient.GetLaunchSpec(ctx, x.opts.LaunchSpecID)
	if err != nil {
		return err
	}

	rawJSON, err := json.MarshalIndent(spec.Obj, "", "  ")
	if err != nil {
		return err
	}

	editor, err := x.opts.Clients.NewEditor()
	if err != nil {
		return err
	}

	editedJSON, path, err := editor.OpenTempFile(ctx, "spotinst", ".json", bytes.NewBuffer(rawJSON))
	if err != nil {
		return err
	}

	if bytes.Equal(rawJSON, editedJSON) {
		os.Remove(path)
		fmt.Fprintln(x.opts.Out, "Edit cancelled, no changes made.")
		return nil
	}

	if err := json.Unmarshal(editedJSON, spec.Obj); err != nil {
		return err
	}

	_, err = oceanClient.UpdateLaunchSpec(ctx, spec)
	return nil
}

func (x *CmdEditLaunchSpecKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdEditLaunchSpecOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdEditLaunchSpecKubernetesOptions) initDefaults(opts *CmdEditLaunchSpecOptions) {
	x.CmdEditLaunchSpecOptions = opts
}

func (x *CmdEditLaunchSpecKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.LaunchSpecID, flags.FlagOceanSpecID, x.LaunchSpecID, "id of the launch spec")
}

func (x *CmdEditLaunchSpecKubernetesOptions) Validate() error {
	if err := x.CmdEditLaunchSpecOptions.Validate(); err != nil {
		return err
	}

	if x.LaunchSpecID == "" {
		return errors.Required("LaunchSpecID")
	}

	return nil
}
