package ocean

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spotinst/spotinst-cli/internal/errors"
	"github.com/spotinst/spotinst-cli/internal/spotinst"
	"github.com/spotinst/spotinst-cli/internal/utils/flags"
)

type (
	CmdEditClusterKubernetes struct {
		cmd  *cobra.Command
		opts CmdEditClusterKubernetesOptions
	}

	CmdEditClusterKubernetesOptions struct {
		*CmdEditClusterOptions

		ClusterID string
	}
)

func NewCmdEditClusterKubernetes(opts *CmdEditClusterOptions) *cobra.Command {
	return newCmdEditClusterKubernetes(opts).cmd
}

func newCmdEditClusterKubernetes(opts *CmdEditClusterOptions) *CmdEditClusterKubernetes {
	var cmd CmdEditClusterKubernetes

	cmd.cmd = &cobra.Command{
		Use:           "kubernetes",
		Short:         "Edit a Kubernetes cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

	return &cmd
}

func (x *CmdEditClusterKubernetes) Run(ctx context.Context) error {
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

func (x *CmdEditClusterKubernetes) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdEditClusterKubernetes) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdEditClusterKubernetes) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdEditClusterKubernetes) run(ctx context.Context) error {
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

	cluster, err := oceanClient.GetCluster(ctx, x.opts.ClusterID)
	if err != nil {
		return err
	}

	rawJSON, err := json.MarshalIndent(cluster.Obj, "", "  ")
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

	if err := json.Unmarshal(editedJSON, cluster.Obj); err != nil {
		return err
	}

	_, err = oceanClient.UpdateCluster(ctx, cluster)
	return err
}

func (x *CmdEditClusterKubernetesOptions) Init(fs *pflag.FlagSet, opts *CmdEditClusterOptions) {
	x.initFlags(fs)
	x.initDefaults(opts)
}

func (x *CmdEditClusterKubernetesOptions) initDefaults(opts *CmdEditClusterOptions) {
	x.CmdEditClusterOptions = opts
}

func (x *CmdEditClusterKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdEditClusterKubernetesOptions) Validate() error {
	if err := x.CmdEditClusterOptions.Validate(); err != nil {
		return err
	}

	if x.ClusterID == "" {
		return errors.Required("ClusterID")
	}

	return nil
}
