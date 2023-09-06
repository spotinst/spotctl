package ocean

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	spotctlerrors "github.com/spotinst/spotctl/internal/errors"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/kubernetes"
	"github.com/spotinst/spotctl/internal/log"
)

type (
	CmdSparkConnect struct {
		cmd  *cobra.Command
		opts CmdSparkConnectOptions
	}

	CmdSparkConnectOptions struct {
		*CmdSparkOptions
		OceanClusterID    string
		ClusterName       string
		Region            string
		Tags              []string
		KubernetesVersion string
		KubeConfigPath    string
	}
)

func NewCmdSparkConnect(opts *CmdSparkOptions) *cobra.Command {
	return newCmdSparkConnect(opts).cmd
}

func newCmdSparkConnect(opts *CmdSparkOptions) *CmdSparkConnect {
	var cmd CmdSparkConnect

	cmd.cmd = &cobra.Command{
		Use:           "connect",
		Short:         "connect to Ocean Spark",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"t"},
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return cmd.preRun(context.Background())
		},
	}

	cmd.opts.Init(cmd.cmd.Flags(), opts)

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

func (x *CmdSparkConnect) Run(ctx context.Context) error {
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

func (x *CmdSparkConnect) survey(_ context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdSparkConnect) log(_ context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdSparkConnect) validate(_ context.Context) error {
	return x.opts.Validate()
}

func (x *CmdSparkConnect) run(ctx context.Context) error {
	log.Infof("Spark connect will now run")

	return nil
}

func (x *CmdSparkConnectOptions) Init(fs *pflag.FlagSet, opts *CmdSparkOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdSparkConnectOptions) initDefaults(opts *CmdSparkOptions) {
	x.CmdSparkOptions = opts
}

func (x *CmdSparkConnectOptions) initFlags(fs *pflag.FlagSet) {
	// todo check what flags we need
	fs.StringVar(&x.OceanClusterID, flags.FlagOFASOceanClusterID, x.OceanClusterID, "ID of Ocean cluster that should be imported into Ocean for Apache Spark. Note that your machine must be configured to access the cluster.")
	fs.StringVar(&x.KubeConfigPath, flags.FlagOFASKubeConfigPath, kubernetes.GetDefaultKubeConfigPath(), "path to local kubeconfig")
}

func (x *CmdSparkConnectOptions) Validate() error {
	if x.OceanClusterID != "" && x.ClusterName != "" {
		return spotctlerrors.RequiredXor(flags.FlagOFASOceanClusterID, flags.FlagOFASClusterName)
	}
	if x.KubeConfigPath == "" {
		return spotctlerrors.Required(flags.FlagOFASKubeConfigPath)
	}
	return x.CmdSparkOptions.Validate()
}
