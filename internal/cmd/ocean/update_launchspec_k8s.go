package ocean

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/spotinst"
	"github.com/spotinst/spotinst-cli/internal/utils/flags"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	spotinstsdk "github.com/spotinst/spotinst-sdk-go/spotinst"
)

type (
	CmdUpdateLaunchSpecKubernetes struct {
		cmd  *cobra.Command
		opts CmdUpdateLaunchSpecKubernetesOptions
	}

	CmdUpdateLaunchSpecKubernetesOptions struct {
		*CmdUpdateLaunchSpecOptions

		Name             string
		SpecID           string
		OceanID          string
		ImageID          string
		UserData         string
		SecurityGroupIDs []string
	}
)

func NewCmdUpdateLaunchSpecKubernetes(opts *CmdUpdateLaunchSpecOptions) *cobra.Command {
	return newCmdUpdateLaunchSpecKubernetes(opts).cmd
}

func newCmdUpdateLaunchSpecKubernetes(opts *CmdUpdateLaunchSpecOptions) *CmdUpdateLaunchSpecKubernetes {
	var cmd CmdUpdateLaunchSpecKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Update an existing Kubernetes launchspec",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdUpdateLaunchSpecKubernetes) Run(ctx context.Context) error {
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

func (x *CmdUpdateLaunchSpecKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdUpdateLaunchSpecKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdUpdateLaunchSpecKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdUpdateLaunchSpecKubernetes) run(ctx context.Context) error {
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

	spec, ok := x.buildLaunchSpecFromOpts()
	if !ok {
		fmt.Fprintln(x.opts.Out, "Update cancelled, no changes made.")
		return nil
	}

	_, err = oceanClient.UpdateLaunchSpec(ctx, spec)
	return err
}

func (x *CmdUpdateLaunchSpecKubernetes) buildLaunchSpecFromOpts() (*spotinst.OceanLaunchSpec, bool) {
	var spec interface{}
	var changed bool

	switch x.opts.CloudProvider {
	case spotinst.CloudProviderAWS:
		spec, changed = x.buildLaunchSpecFromOptsAWS()
	}

	return &spotinst.OceanLaunchSpec{Obj: spec}, changed
}

func (x *CmdUpdateLaunchSpecKubernetes) buildLaunchSpecFromOptsAWS() (*aws.LaunchSpec, bool) {
	spec := new(aws.LaunchSpec)
	changed := false

	if x.opts.Name != "" {
		spec.SetName(spotinstsdk.String(x.opts.Name))
	}

	if x.opts.OceanID != "" {
		spec.SetOceanId(spotinstsdk.String(x.opts.OceanID))
	}

	if x.opts.ImageID != "" {
		spec.SetImageId(spotinstsdk.String(x.opts.ImageID))
	}

	if x.opts.UserData != "" {
		if _, err := base64.StdEncoding.DecodeString(x.opts.UserData); err != nil {
			x.opts.UserData = base64.StdEncoding.EncodeToString([]byte(x.opts.UserData))
		}

		spec.SetUserData(spotinstsdk.String(x.opts.UserData))
	}

	if len(x.opts.SecurityGroupIDs) > 0 {
		spec.SetSecurityGroupIDs(x.opts.SecurityGroupIDs)
	}

	if changed = !reflect.DeepEqual(spec, new(aws.LaunchSpec)); changed {
		spec.SetId(spotinstsdk.String(x.opts.SpecID))
	}

	return spec, changed
}

func (x *CmdUpdateLaunchSpecKubernetesOptions) Init(flags *pflag.FlagSet, opts *CmdUpdateLaunchSpecOptions) {
	x.initDefaults(opts)
	x.initFlags(flags)
}

func (x *CmdUpdateLaunchSpecKubernetesOptions) initDefaults(opts *CmdUpdateLaunchSpecOptions) {
	x.CmdUpdateLaunchSpecOptions = opts
}

func (x *CmdUpdateLaunchSpecKubernetesOptions) initFlags(flags *pflag.FlagSet) {
	flags.StringVar(
		&x.Name,
		"name",
		x.Name,
		"name of the launch spec")

	flags.StringVar(
		&x.SpecID,
		"spec-id",
		x.SpecID,
		"id of the launch spec")

	flags.StringVar(
		&x.OceanID,
		"ocean-id",
		x.OceanID,
		"id of the cluster")

	flags.StringVar(
		&x.ImageID,
		"image-id",
		x.ImageID,
		"id of the image")

	flags.StringVar(
		&x.UserData,
		"user-data",
		x.UserData,
		"user data to provide when launching a node (plain-text or base64-encoded)")
}

func (x *CmdUpdateLaunchSpecKubernetesOptions) Validate() error {
	return x.CmdUpdateLaunchSpecOptions.Validate()
}
