package ocean

import (
	"context"
	"fmt"
	"github.com/spotinst/spotctl/internal/errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/spotinst/spotctl/internal/dep"
	"github.com/spotinst/spotctl/internal/flags"
)

type (
	CmdSparkCreateCluster struct {
		cmd  *cobra.Command
		opts CmdSparkCreateClusterOptions
	}

	CmdSparkCreateClusterOptions struct {
		*CmdSparkCreateOptions
		ConfigFile        string
		ClusterID         string
		ClusterName       string
		Region            string
		Tags              []string
		KubernetesVersion string
	}
)

func NewCmdSparkCreateCluster(opts *CmdSparkCreateOptions) *cobra.Command {
	return newCmdSparkCreateCluster(opts).cmd
}

func newCmdSparkCreateCluster(opts *CmdSparkCreateOptions) *CmdSparkCreateCluster {
	var cmd CmdSparkCreateCluster

	cmd.cmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Create a new Ocean for Apache Spark cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		Aliases:       []string{"c"},
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

func (x *CmdSparkCreateCluster) preRun(ctx context.Context) error {
	// Call to the parent command's PersistentPreRunE.
	// See: https://github.com/spf13/cobra/issues/216.
	if parent := x.cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
		if err := parent.PersistentPreRunE(parent, nil); err != nil {
			return err
		}
	}
	return x.installDeps(ctx)
}

func (x *CmdSparkCreateCluster) Run(ctx context.Context) error {
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

func (x *CmdSparkCreateCluster) survey(ctx context.Context) error {
	if x.opts.Noninteractive {
		return nil
	}

	return nil
}

func (x *CmdSparkCreateCluster) log(ctx context.Context) error {
	flags.Log(x.cmd)
	return nil
}

func (x *CmdSparkCreateCluster) validate(ctx context.Context) error {
	return x.opts.Validate()
}

func (x *CmdSparkCreateCluster) run(ctx context.Context) error {
	fmt.Println("SHOULD CREATE OFAS CLUSTER")

	return nil
}

func (x *CmdSparkCreateClusterOptions) Init(fs *pflag.FlagSet, opts *CmdSparkCreateOptions) {
	x.initDefaults(opts)
	x.initFlags(fs)
}

func (x *CmdSparkCreateClusterOptions) initDefaults(opts *CmdSparkCreateOptions) {
	x.CmdSparkCreateOptions = opts
}

func (x *CmdSparkCreateClusterOptions) initFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&x.ConfigFile, flags.FlagOFASConfigFile, "f", x.ConfigFile, "load configuration from a file (or stdin if set to '-')")
	fs.StringVar(&x.ClusterID, flags.FlagOFASClusterID, x.ClusterID, "cluster id (will be created if empty)")
	fs.StringVar(&x.ClusterName, flags.FlagOFASClusterName, x.ClusterName, "cluster name")
	fs.StringVar(&x.Region, flags.FlagOFASClusterRegion, os.Getenv("AWS_REGION"), "region in which your cluster (control plane and nodes) will be created")
	fs.StringSliceVar(&x.Tags, "tags", x.Tags, "list of K/V pairs used to tag all cloud resources (eg: \"Owner=john@example.com,Team=DevOps\")")
	fs.StringVar(&x.KubernetesVersion, "kubernetes-version", "1.18", "kubernetes version")
}

func (x *CmdSparkCreateClusterOptions) Validate() error {
	if x.ClusterID != "" && x.ClusterName != "" {
		return errors.RequiredXor(flags.FlagOFASClusterID, flags.FlagOFASClusterName)
	}
	if x.ClusterID != "" && x.ConfigFile != "" {
		return errors.RequiredXor(flags.FlagOFASClusterID, flags.FlagOFASConfigFile)
	}
	if x.ClusterName != "" && x.ConfigFile != "" {
		return errors.RequiredXor(flags.FlagOFASClusterName, flags.FlagOFASConfigFile)
	}
	return x.CmdSparkCreateOptions.Validate()
}

func (x *CmdSparkCreateCluster) installDeps(ctx context.Context) error {
	// Initialize a new dependency manager.
	dm, err := x.opts.Clientset.NewDepManager()
	if err != nil {
		return err
	}

	// Install options.
	installOpts := []dep.InstallOption{
		dep.WithInstallPolicy(dep.InstallPolicy(x.opts.InstallPolicy)),
		dep.WithNoninteractive(x.opts.Noninteractive),
		dep.WithDryRun(x.opts.DryRun),
	}

	// Install!
	return dm.InstallBulk(ctx, dep.DefaultDependencyListKubernetes(), installOpts...)
}
