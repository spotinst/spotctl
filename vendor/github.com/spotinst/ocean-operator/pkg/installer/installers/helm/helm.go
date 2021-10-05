// Copyright 2021 NetApp, Inc. All Rights Reserved.

package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
	"github.com/spotinst/ocean-operator/pkg/installer"
	"github.com/spotinst/ocean-operator/pkg/log"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	_ "helm.sh/helm/v3/pkg/downloader"
	_ "helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	installer.MustRegister(oceanv1alpha1.OceanComponentTypeHelm.String(),
		func(options *installer.InstallerOptions) (installer.Installer, error) {
			return NewInstaller(options), nil
		})
}

type Installer struct {
	ClientGetter genericclioptions.RESTClientGetter
	Namespace    string
	DryRun       bool
	Log          log.Logger
}

// NewInstaller returns a Installer.
func NewInstaller(options *installer.InstallerOptions) *Installer {
	return &Installer{
		ClientGetter: options.ClientGetter,
		Namespace:    options.Namespace,
		DryRun:       options.DryRun,
		Log:          options.Log,
	}
}

func (i *Installer) Get(name oceanv1alpha1.OceanComponentName) (*installer.Release, error) {
	config, err := i.getActionConfig(i.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action configuration: %w", err)
	}

	rel, err := action.NewGet(config).Run(name.String())
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil, installer.ErrReleaseNotFound
		}
		return nil, err
	}

	values := make(map[string]interface{})
	if rel.Config != nil {
		values = rel.Config
	}

	return i.translateRelease(rel, values), nil
}

func (i *Installer) Install(component *oceanv1alpha1.OceanComponent) (*installer.Release, error) {
	values := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(component.Spec.Values), &values); err != nil {
		return nil, fmt.Errorf("invalid values configuration: %w", err)
	}
	i.Log.V(5).Info("install values configuration", "values", values)

	config, err := i.getActionConfig(i.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action configuration: %w", err)
	}

	chartName := component.Spec.Name.String()
	rel, err := action.NewGet(config).Run(chartName)
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return nil, fmt.Errorf("existing release check failed: %w", err)
	} else if rel != nil {
		i.Log.Info("release already exists", "name", chartName)
		return i.translateRelease(rel, values), nil
	}

	act := action.NewInstall(config)
	act.ReleaseName = chartName
	act.Namespace = i.Namespace
	act.DryRun = i.DryRun
	act.ChartPathOptions.RepoURL = component.Spec.URL
	act.ChartPathOptions.Version = component.Spec.Version
	act.CreateNamespace = true

	settings := new(cli.EnvSettings)
	cache, err := ioutil.TempDir(os.TempDir(), "oceancache-")
	if err != nil {
		return nil, fmt.Errorf("unable to create cache directory: %w", err)
	}
	defer func() {
		err := os.RemoveAll(cache)
		if err != nil {
			i.Log.Error(err, "could not delete cache directory", "path", cache)
		}
	}()
	settings.RepositoryCache = cache
	settings.Debug = i.DryRun // renders out invalid yaml

	// Check for the existence of a file called 'chartName' in the current directory.
	// If it exists, it will assume that is the chart and it won't download the chart.
	cp, err := act.ChartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart %s: %w", chartName, err)
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart %s: %w", cp, err)
	}

	rel, err = act.Run(chart, values)
	if err != nil {
		return nil, fmt.Errorf("installation error: %w", err)
	}

	i.Log.Info("installed", "name", rel.Name)
	return i.translateRelease(rel, values), nil
}

func (i *Installer) Uninstall(component *oceanv1alpha1.OceanComponent) error {
	config, err := i.getActionConfig(i.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get action configuration: %w", err)
	}

	act := action.NewUninstall(config)
	act.DryRun = i.DryRun

	chartName := component.Spec.Name.String()
	_, err = act.Run(chartName)
	if err != nil {
		i.Log.Error(err, fmt.Sprintf("ignoring deletion error: %v", err))
	} else {
		i.Log.Info("uninstalled", "name", chartName)
	}

	return nil
}

func (i *Installer) Upgrade(component *oceanv1alpha1.OceanComponent) (*installer.Release, error) {
	values := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(component.Spec.Values), &values); err != nil {
		return nil, fmt.Errorf("invalid values configuration: %w", err)
	}
	i.Log.V(5).Info("upgrade values configuration", "values", values)

	config, err := i.getActionConfig(i.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action configuration: %w", err)
	}

	act := action.NewUpgrade(config)
	act.Namespace = i.Namespace
	act.DryRun = i.DryRun
	act.ChartPathOptions.RepoURL = component.Spec.URL
	act.ChartPathOptions.Version = component.Spec.Version
	act.ReuseValues = true

	settings := new(cli.EnvSettings)
	cacheDir, err := ioutil.TempDir(os.TempDir(), "oceancache-")
	if err != nil {
		return nil, fmt.Errorf("unable to create cache directory: %w", err)
	}
	defer func() {
		err := os.RemoveAll(cacheDir)
		if err != nil {
			i.Log.Error(err, "could not delete cache directory", "path", cacheDir)
		}
	}()
	settings.RepositoryCache = cacheDir
	settings.Debug = i.DryRun // renders out invalid yaml

	chartName := component.Spec.Name.String()
	cp, err := act.ChartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart %s: %w", chartName, err)
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart %s: %w", cp, err)
	}

	rel, err := act.Run(chartName, chart, values)
	if err != nil {
		return nil, fmt.Errorf("installation error: %w", err)
	}

	i.Log.Info("upgraded", "name", rel.Name)
	return i.translateRelease(rel, values), nil
}

func (i *Installer) IsUpgrade(component *oceanv1alpha1.OceanComponent, release *installer.Release) bool {
	if component.Spec.Version != release.Version {
		return true
	}

	newValues := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(component.Spec.Values), &newValues); err != nil {
		i.Log.Error(err, "failed to unmarshal values")
		return true // fail properly later
	}
	if newValues == nil {
		newValues = make(map[string]interface{})
	}

	oldValues := make(map[string]interface{})
	if release.Values != nil {
		oldValues = release.Values
	}

	if diff := strings.TrimSpace(cmp.Diff(newValues, oldValues)); diff != "" {
		i.Log.V(5).Info("upgrade is required", "diff", diff)
		return true
	}

	return false
}

// https://stackoverflow.com/questions/59782217/run-helm3-client-from-in-cluster
func (i *Installer) getActionConfig(namespace string) (*action.Configuration, error) {
	config := new(action.Configuration)
	if err := config.Init(i.ClientGetter, namespace, "secret", i.actionLogger); err != nil {
		return nil, err
	}
	return config, nil
}

// actionLogger returns an action.DebugLog that uses Zap to log.
func (i *Installer) actionLogger(format string, v ...interface{}) {
	i.Log.Info(fmt.Sprintf(format, v...))
}

func (i *Installer) translateRelease(rel *release.Release, values map[string]interface{}) *installer.Release {
	return &installer.Release{
		Name:        rel.Name,
		Version:     rel.Chart.Metadata.Version,
		AppVersion:  rel.Chart.Metadata.AppVersion,
		Description: rel.Info.Description,
		Manifest:    rel.Manifest,
		Status:      i.translateReleaseStatus(rel.Info.Status),
		Values:      values,
	}
}

func (i *Installer) translateReleaseStatus(status release.Status) installer.ReleaseStatus {
	switch status {
	case release.StatusFailed, release.StatusSuperseded:
		return installer.ReleaseStatusFailed
	case release.StatusPendingInstall, release.StatusPendingRollback, release.StatusPendingUpgrade, release.StatusUninstalling:
		return installer.ReleaseStatusProgressing
	case release.StatusUninstalled:
		return installer.ReleaseStatusUninstalled
	case release.StatusDeployed:
		return installer.ReleaseStatusDeployed
	case release.StatusUnknown:
		return installer.ReleaseStatusUnknown
	default:
		return ""
	}
}
