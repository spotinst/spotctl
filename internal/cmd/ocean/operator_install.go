package ocean

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
	"github.com/spotinst/ocean-operator/pkg/tide"
	"github.com/spotinst/ocean-operator/pkg/tide/values"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/ocean"
	"github.com/spotinst/spotctl/internal/uuid"
	"github.com/theckman/yacspin"
	ctrl "sigs.k8s.io/controller-runtime"
)

type (
	CmdOperatorInstall struct {
		cmd  *cobra.Command
		opts CmdOperatorInstallOptions
	}

	CmdOperatorInstallOptions struct {
		*CmdOperatorOptions

		// chart
		ChartName       string
		ChartVersion    string
		ChartURL        string
		ChartValuesJSON string

		// shorthands for --chart-values='{...}'
		Components        *ocean.ComponentsFlag
		Namespace         string
		ClusterIdentifier string
		ACDIdentifier     string
	}
)

func NewCmdOperatorInstall(opts *CmdOperatorOptions) *cobra.Command {
	return newCmdOperatorInstall(opts).cmd
}

func newCmdOperatorInstall(opts *CmdOperatorOptions) *CmdOperatorInstall {
	cmd := &CmdOperatorInstall{
		opts: CmdOperatorInstallOptions{
			CmdOperatorOptions: opts,
			Components:         ocean.NewDefaultComponentsFlag(log.DefaultLogger()),
			ClusterIdentifier:  fmt.Sprintf("ocean-%s", uuid.NewV4().Short()),
		},
	}

	cmd.cmd = &cobra.Command{
		Use:           "install",
		Short:         "Install the Ocean Operator",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	// chart
	cmd.cmd.Flags().StringVar(&cmd.opts.ChartName, "chart-name", tide.OceanOperatorChart, "chart name")
	cmd.cmd.Flags().StringVar(&cmd.opts.ChartVersion, "chart-version", tide.OceanOperatorVersion, "chart version")
	cmd.cmd.Flags().StringVar(&cmd.opts.ChartURL, "chart-url", tide.OceanOperatorRepository, "chart repository url")
	cmd.cmd.Flags().StringVar(&cmd.opts.ChartValuesJSON, "chart-values", tide.OceanOperatorValues, "chart values (json)")

	// shorthands for --chart-values='{...}'
	cmd.cmd.Flags().Var(cmd.opts.Components, "components", "list of components to install")
	cmd.cmd.Flags().StringVar(&cmd.opts.Namespace, "namespace", oceanv1alpha1.NamespaceSystem, "namespace where components should be installed")
	cmd.cmd.Flags().StringVar(&cmd.opts.ClusterIdentifier, "cluster-identifier", cmd.opts.ClusterIdentifier, "unique identifier used by the ocean-controller to connect between the backend and the cluster")
	cmd.cmd.Flags().StringVar(&cmd.opts.ACDIdentifier, "acd-identifier", "", "unique identifier used by the ocean-aks-connector when importing an aks cluster")

	return cmd
}

func (x *CmdOperatorInstall) Run(ctx context.Context) error {
	flags.Log(x.cmd)

	spinner, err := yacspin.New(yacspin.Config{
		Frequency:         200 * time.Millisecond,
		CharSet:           yacspin.CharSets[69],
		Suffix:            " tide",
		SuffixAutoColon:   true,
		ColorAll:          true,
		HideCursor:        true,
		Message:           "start",
		StopCharacter:     "●",
		StopFailCharacter: "●",
		StopColors:        []string{"fgGreen"},
		StopFailColors:    []string{"fgRed"},
	})
	if err != nil {
		return err
	}

	spinner.Start()
	spinner.Message("installing ocean-operator")
	logger := newSpinnerLogger("ocean-operator", spinner)

	ctrl.SetLogger(logger)
	config, err := ctrl.GetConfig()
	if err != nil {
		spinner.StopFailMessage("unable to get kubeconfig")
		spinner.StopFail()
		return err
	}

	clientRuntime, err := tide.NewControllerRuntimeClient(config, tide.DefaultScheme())
	if err != nil {
		spinner.StopFailMessage("unable to get runtime client")
		spinner.StopFail()
		return err
	}

	spinner.Message("initializing chart values builder")
	valuesBuilder := values.NewOceanOperatorBuilder(values.NewOceanBaseBuilder().
		WithClient(clientRuntime). // fetch values from configmap/secret
		WithClusterIdentifier(x.opts.ClusterIdentifier).
		WithACDIdentifier(x.opts.ACDIdentifier)).
		WithComponents(x.opts.Components.StringSlice())

	spinner.Message("completing chart values")
	chartValues, err := values.ForOceanOperator(ctx, x.opts.ChartValuesJSON, valuesBuilder)
	if err != nil {
		spinner.StopFailMessage("unable to complete missing chart values")
		spinner.StopFail()
		return err
	}

	operator := tide.NewOperatorOceanComponent(
		tide.WithChartNamespace(x.opts.Namespace),
		tide.WithChartName(x.opts.ChartName),
		tide.WithChartURL(x.opts.ChartURL),
		tide.WithChartVersion(x.opts.ChartVersion),
		tide.WithChartValues(chartValues),
	)

	clientGetter := tide.NewConfigFlags(config, x.opts.Namespace)
	if err = tide.InstallOperator(ctx, operator, clientGetter,
		x.opts.Wait, x.opts.DryRun, x.opts.Timeout, logger); err != nil {
		spinner.StopFailMessage(fmt.Sprintf("unable to install: %v", err))
		spinner.StopFail()
		return err
	}

	if x.opts.DryRun {
		spinner.StopMessage("would install ocean-operator and its components")
	} else {
		spinner.StopMessage("ocean-operator is now installed and managing components")
	}

	spinner.Stop()
	return nil
}
