// Copyright 2021 NetApp, Inc. All Rights Reserved.

package tide

import (
	"context"
	"errors"
	"fmt"
	"time"

	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
	"github.com/spotinst/ocean-operator/pkg/installer"
	"github.com/spotinst/ocean-operator/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// NewOperatorOceanComponent returns an oceanv1alpha1.OceanComponent
// representing the Ocean Operator.
func NewOperatorOceanComponent(options ...ChartOption) *oceanv1alpha1.OceanComponent {
	comp := &oceanv1alpha1.OceanComponent{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: oceanv1alpha1.NamespaceSystem,
			Name:      OceanOperatorChart,
		},
		Spec: oceanv1alpha1.OceanComponentSpec{
			Type:    oceanv1alpha1.OceanComponentTypeHelm,
			State:   oceanv1alpha1.OceanComponentStatePresent,
			Name:    OceanOperatorChart,
			URL:     OceanOperatorRepository,
			Version: OceanOperatorVersion,
		},
	}

	opts := mutateChartOptions(options...)
	comp.Namespace = opts.Namespace
	comp.Spec.Name = oceanv1alpha1.OceanComponentName(opts.Name)
	comp.Spec.URL = opts.URL
	comp.Spec.Version = opts.Version
	comp.Spec.Values = opts.Values

	return comp
}

// InstallOperator installs the Ocean Operator.
func InstallOperator(
	ctx context.Context,
	operator *oceanv1alpha1.OceanComponent,
	clientGetter genericclioptions.RESTClientGetter,
	wait, dryRun bool,
	timeout time.Duration,
	log log.Logger,
) error {
	// install or upgrade
	{
		i, err := installer.GetInstance(
			operator.Spec.Type.String(),
			installer.WithNamespace(operator.Namespace),
			installer.WithClientGetter(clientGetter),
			installer.WithDryRun(dryRun),
			installer.WithLogger(log))
		if err != nil {
			log.Error(err, "unable to create installer")
			return err
		}

		existing, err := i.Get(operator.Spec.Name)
		if err != nil && !installer.IsReleaseNotFound(err) {
			log.Error(err, "error checking ocean operator release")
			return err
		}

		var release *installer.Release
		if existing != nil && i.IsUpgrade(operator, existing) {
			log.Info("upgrading ocean operator")
			release, err = i.Upgrade(operator)
		} else {
			log.Info("installing ocean operator")
			release, err = i.Install(operator)
		}
		if err != nil {
			return fmt.Errorf("cannot release ocean operator: %w", err)
		}
		if dryRun && release != nil {
			log.Info(release.Manifest)
		}
	}

	// validate
	{
		if wait && !dryRun {
			config, err := clientGetter.ToRESTConfig()
			if err != nil {
				return fmt.Errorf("cannot get restconfig: %w", err)
			}

			clientSet, err := kubernetes.NewForConfig(config)
			if err != nil {
				return fmt.Errorf("cannot connect to cluster: %w", err)
			}

			log.Info("waiting for deployment to be ready")
			client := clientSet.AppsV1().Deployments(operator.Namespace)
			err = utilwait.Poll(5*time.Second, timeout, func() (bool, error) {
				dep, err := client.Get(ctx, OceanOperatorDeployment, metav1.GetOptions{})
				if err != nil || dep.Status.AvailableReplicas == 0 || dep.Status.UnavailableReplicas != 0 {
					return false, nil
				}
				log.V(2).Info("polled",
					"deployment", dep.Name,
					"replicas", dep.Status.AvailableReplicas)
				return true, nil
			})
			if err != nil && errors.Is(err, utilwait.ErrWaitTimeout) {
				return fmt.Errorf("timed out waiting for deployment to be ready")
			}
		}
	}

	return nil
}

// UninstallOperator uninstalls the Ocean Operator.
func UninstallOperator(
	ctx context.Context,
	operator *oceanv1alpha1.OceanComponent,
	clientGetter genericclioptions.RESTClientGetter,
	wait, dryRun bool,
	timeout time.Duration,
	log log.Logger,
) error {
	// uninstall
	{
		i, err := installer.GetInstance(
			operator.Spec.Type.String(),
			installer.WithNamespace(operator.Namespace),
			installer.WithClientGetter(clientGetter),
			installer.WithDryRun(dryRun),
			installer.WithLogger(log))
		if err != nil {
			log.Error(err, "unable to create installer")
			return err
		}

		existing, err := i.Get(operator.Spec.Name)
		if err != nil && !installer.IsReleaseNotFound(err) {
			log.Error(err, "error checking ocean operator release")
			return err
		}

		if existing != nil {
			log.Info("uninstalling ocean operator")
			if err = i.Uninstall(operator); err != nil {
				return fmt.Errorf("cannot uninstall ocean operator: %w", err)
			}
		}
	}

	// validate
	{
		if wait && !dryRun {
			config, err := clientGetter.ToRESTConfig()
			if err != nil {
				return fmt.Errorf("cannot get restconfig: %w", err)
			}

			clientSet, err := kubernetes.NewForConfig(config)
			if err != nil {
				return fmt.Errorf("cannot connect to cluster: %w", err)
			}

			log.Info("waiting for deployment to be evicted")
			client := clientSet.AppsV1().Deployments(operator.Namespace)
			err = utilwait.Poll(5*time.Second, timeout, func() (bool, error) {
				dep, err := client.Get(ctx, OceanOperatorDeployment, metav1.GetOptions{})
				if err == nil {
					log.V(2).Info("polled",
						"deployment", dep.Name,
						"replicas", dep.Status.AvailableReplicas)
					return false, nil
				}
				return true, nil
			})
			if err != nil && errors.Is(err, utilwait.ErrWaitTimeout) {
				return fmt.Errorf("timed out waiting for deployment to be evicted")
			}
		}
	}

	return nil
}
