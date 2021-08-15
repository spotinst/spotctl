package ocean

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
	"github.com/spotinst/ocean-operator/pkg/tide"
	"github.com/spotinst/spotctl/internal/flags"
	"github.com/theckman/yacspin"
	ctrl "sigs.k8s.io/controller-runtime"
)

type (
	CmdOperatorUninstall struct {
		cmd  *cobra.Command
		opts CmdOperatorUninstallOptions
	}

	CmdOperatorUninstallOptions struct {
		*CmdOperatorOptions

		ChartName string
		Namespace string
		Purge     bool
	}
)

func NewCmdOperatorUninstall(opts *CmdOperatorOptions) *cobra.Command {
	return newCmdOperatorUninstall(opts).cmd
}

func newCmdOperatorUninstall(opts *CmdOperatorOptions) *CmdOperatorUninstall {
	cmd := &CmdOperatorUninstall{
		opts: CmdOperatorUninstallOptions{
			CmdOperatorOptions: opts,
		},
	}

	cmd.cmd = &cobra.Command{
		Use:           "uninstall",
		Short:         "Uninstall the Ocean Operator",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(*cobra.Command, []string) error {
			return cmd.Run(context.Background())
		},
	}

	cmd.cmd.Flags().StringVar(&cmd.opts.ChartName, "chart-name", tide.OceanOperatorChart, "chart name")
	cmd.cmd.Flags().StringVar(&cmd.opts.Namespace, "namespace", oceanv1alpha1.NamespaceSystem, "namespace where components should be installed")
	cmd.cmd.Flags().BoolVar(&cmd.opts.Purge, "purge", false, "purge all configuration (requires cluster admin access)")

	return cmd
}

func (x *CmdOperatorUninstall) Run(ctx context.Context) error {
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
	spinner.Message("uninstalling ocean-operator")
	logger := newSpinnerLogger("ocean-operator", spinner)

	ctrl.SetLogger(logger)
	config, err := ctrl.GetConfig()
	if err != nil {
		spinner.StopFailMessage("unable to get kubeconfig")
		spinner.StopFail()
		return err
	}

	clientGetter := tide.NewConfigFlags(config, x.opts.Namespace)

	if x.opts.Purge && !x.opts.DryRun {
		manager, err := tide.NewManager(clientGetter, logger)
		if err != nil {
			spinner.StopFailMessage("unable to create tide manager")
			spinner.StopFail()
			return err
		}
		deleteOptions := []tide.DeleteOption{
			tide.WithNamespace(x.opts.Namespace),
		}
		if err = manager.DeleteEnvironment(ctx, deleteOptions...); err != nil {
			return err
		}
	}

	operator := tide.NewOperatorOceanComponent(
		tide.WithChartNamespace(x.opts.Namespace),
		tide.WithChartName(x.opts.ChartName),
	)
	if err = tide.UninstallOperator(ctx, operator, clientGetter,
		x.opts.Wait, x.opts.DryRun, x.opts.Timeout, logger); err != nil {
		spinner.StopFailMessage(fmt.Sprintf("unable to uninstall: %v", err))
		spinner.StopFail()
		return err
	}

	spinner.StopMessage("ocean-operator is now removed")
	spinner.Stop()

	return nil
}
