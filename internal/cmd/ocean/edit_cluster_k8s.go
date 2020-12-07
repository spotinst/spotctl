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
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/spot"
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
		Aliases:       []string{"k8s", "kube", "k"},
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
	spotClientOpts := []spot.ClientOption{
		spot.WithCredentialsProfile(x.opts.Profile),
	}

	spotClient, err := x.opts.Clientset.NewSpotClient(spotClientOpts...)
	if err != nil {
		return err
	}

	oceanClient, err := spotClient.Services().Ocean(x.opts.CloudProvider, spot.OrchestratorKubernetes)
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

	editor, err := x.opts.Clientset.NewEditor()
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
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdEditClusterKubernetesOptions) initDefaults(opts *CmdEditClusterOptions) {
	x.CmdEditClusterOptions = opts
}

func (x *CmdEditClusterKubernetesOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&x.ClusterID, flags.FlagOceanClusterID, x.ClusterID, "id of the cluster")
}

func (x *CmdEditClusterKubernetesOptions) Validate() error {
	errg := errors.NewErrorGroup()

	if err := x.CmdEditClusterOptions.Validate(); err != nil {
		errg.Add(err)
	}

	if x.ClusterID == "" {
		errg.Add(errors.Required("ClusterID"))
	}

	if errg.Len() > 0 {
		return errg
	}

	return nil
}
