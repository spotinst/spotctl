package install

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/spotinst/wave-operator/api/v1alpha1"
	"github.com/spotinst/wave-operator/catalog"
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

var (
	// ErrReleaseNotFound duplicates the helm error from the driver package
	ErrReleaseNotFound = errors.New("release: not found")
)

const (
	Failed      string = "failed"
	Progressing string = "progressing"
	Uninstalled string = "uninstalled"
	Deployed    string = "deployed"
	Unknown     string = "unknown"
)

type Installer interface {

	// Namespace where components are installed, the namespace must exist before Install() is called
	GetNamespace() string

	// Set the namespace where components are installed
	SetNamespace(namespace string)

	// returns the release name for a given chart
	GetReleaseName(chartName string) string

	// Install applies a helm chart to a cluster.  It's a lightweight wrapper around the helm 3 code
	Install(name string, repository string, version string, values string) error

	// Get a details of an installation
	Get(name string) (*Installation, error)

	// Upgrade applies a helm chart to a cluster.  It's a lightweight wrapper around the helm 3 code
	Upgrade(name string, repository string, version string, values string) error
	//Upgrade(rlsName string, chart *chart.Chart, vals map[string]string, namespace string, opts ...bool) (*release.Release, error)

	IsUpgrade(comp *v1alpha1.WaveComponent, i *Installation) bool

	// Delete removes a helm release from a cluster.  It's a lightweight wrapper around the helm 3 code
	Delete(name string, repository string, version string, values string) error
}

// Installation represents a helm release, that is, a chart installed into a cluster
type Installation struct {
	Name        string
	Version     string
	Values      map[string]interface{}
	Status      string
	Description string
}

func NewInstallation(name, version, status, description string, vals map[string]interface{}) *Installation {
	if vals == nil {
		vals = map[string]interface{}{}
	}
	return &Installation{
		Name:        name,
		Version:     version,
		Status:      status,
		Values:      vals,
		Description: description,
	}
}

type HelmInstaller struct {
	prefix       string
	namespace    string
	ClientGetter genericclioptions.RESTClientGetter
	Log          logr.Logger
}

// helm libraries require a logger function of particular form, different from zap's Info()
func (i *HelmInstaller) logForHelm(format string, v ...interface{}) {
	i.Log.Info(fmt.Sprintf(format, v...))
}

// GetHelm matches type controller.InstallerGetter, returns a HelmInstaller
var GetHelm = func(name string, getter genericclioptions.RESTClientGetter, log logr.Logger) Installer {
	return &HelmInstaller{
		prefix:       name,
		namespace:    catalog.SystemNamespace,
		ClientGetter: getter,
		Log:          log,
	}
}

func (i *HelmInstaller) GetNamespace() string {
	return i.namespace
}

func (i *HelmInstaller) SetNamespace(namespace string) {
	i.namespace = namespace
}

func (i *HelmInstaller) Get(name string) (*Installation, error) {
	cfg, err := i.getActionConfig(i.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action config, %w", err)
	}
	act := action.NewGet(cfg)
	rel, err := act.Run(name)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return nil, ErrReleaseNotFound
		}
		return nil, err
	}
	return NewInstallation(
		rel.Name,
		rel.Chart.Metadata.Version,
		translateStatus(rel.Info.Status),
		rel.Info.Description,
		rel.Config,
	), nil
}

func translateStatus(status release.Status) string {
	switch status {
	case release.StatusFailed, release.StatusSuperseded:
		return Failed
	case release.StatusPendingInstall, release.StatusPendingRollback, release.StatusPendingUpgrade, release.StatusUninstalling:
		return Progressing
	case release.StatusUninstalled:
		return Uninstalled
	case release.StatusDeployed:
		return Deployed
	case release.StatusUnknown:
		return Unknown
	default:
		return ""
	}
}

// https://stackoverflow.com/questions/59782217/run-helm3-client-from-in-cluster
func (i *HelmInstaller) getActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(i.ClientGetter, namespace, "secret", i.logForHelm); err != nil {
		return nil, err
	}
	return actionConfig, nil
}

func (i *HelmInstaller) Upgrade(chartName string, repository string, version string, values string) error {

	var vals map[string]interface{}
	err := yaml.Unmarshal([]byte(values), &vals)
	if err != nil {
		return fmt.Errorf("invalid values configuration, %w", err)
	}

	for k, v := range vals {
		i.Log.Info("values", k, v)
	}

	cfg, err := i.getActionConfig(i.namespace)
	if err != nil {
		return fmt.Errorf("failed to get action config, %w", err)
	}

	upgradeAction := action.NewUpgrade(cfg)

	releaseName := i.GetReleaseName(chartName)
	upgradeAction.Namespace = i.namespace
	upgradeAction.ChartPathOptions.RepoURL = repository
	upgradeAction.ChartPathOptions.Version = version

	settings := &cli.EnvSettings{}
	cache, err := ioutil.TempDir(os.TempDir(), "wavecache-")
	if err != nil {
		return fmt.Errorf("unable to create cache directory, %s", err)
	}
	defer os.RemoveAll(cache)
	settings.RepositoryCache = os.TempDir()

	cp, err := upgradeAction.ChartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart %s, %w", chartName, err)
	}

	chrt, err := loader.Load(cp)
	if err != nil {
		return fmt.Errorf("failed to load chart %s, %w", cp, err)

	}

	rel, err := upgradeAction.Run(releaseName, chrt, vals)
	if err != nil {
		return fmt.Errorf("installation error, %w", err)

	}
	i.Log.Info("installed", "release", rel.Name)
	return nil
}

func (i *HelmInstaller) Install(chartName string, repository string, version string, values string) error {

	var vals map[string]interface{}
	err := yaml.Unmarshal([]byte(values), &vals)
	if err != nil {
		return fmt.Errorf("invalid values configuration, %w", err)
	}

	releaseName := i.GetReleaseName(chartName)

	cfg, err := i.getActionConfig(i.namespace)
	if err != nil {
		return fmt.Errorf("failed to get action config, %w", err)
	}

	getAction := action.NewGet(cfg)
	rel, err := getAction.Run(releaseName)

	if err != nil && err != driver.ErrReleaseNotFound {
		return fmt.Errorf("existing release check failed, %w", err)
	} else if rel != nil {
		i.Log.Info("release already exists")
		return nil
	}

	//repo := fmt.Sprintf("%s/%s-%s.tgz", repository, chartName, version)
	installAction := action.NewInstall(cfg)

	installAction.ReleaseName = releaseName
	installAction.Namespace = i.namespace
	installAction.ChartPathOptions.RepoURL = repository
	installAction.ChartPathOptions.Version = version

	settings := &cli.EnvSettings{}
	cache, err := ioutil.TempDir(os.TempDir(), "wavecache-")
	if err != nil {
		return fmt.Errorf("unable to create cache directory, %s", err)
	}
	defer os.RemoveAll(cache)
	settings.RepositoryCache = os.TempDir()

	cp, err := installAction.ChartPathOptions.LocateChart(chartName, settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart %s, %w", chartName, err)
	}

	chrt, err := loader.Load(cp)
	if err != nil {
		return fmt.Errorf("failed to load chart %s, %w", cp, err)
	}

	rel, err = installAction.Run(chrt, vals)
	if err != nil {
		return fmt.Errorf("installation error, %w", err)
	}
	i.Log.Info("installed", "release", rel.Name)
	return nil
}

func (i *HelmInstaller) GetReleaseName(name string) string {
	if i.prefix != "" {
		return i.prefix + "-" + name
	}
	return name
}

func (i *HelmInstaller) Delete(chartName string, repository string, version string, values string) error {

	var vals map[string]interface{}
	err := yaml.Unmarshal([]byte(values), &vals)
	if err != nil {
		return fmt.Errorf("invalid values configuration, %w", err)
	}

	releaseName := i.GetReleaseName(chartName)

	cfg, err := i.getActionConfig(i.namespace)
	if err != nil {
		return fmt.Errorf("failed to get action config, %w", err)
	}

	getAction := action.NewUninstall(cfg)
	rel, err := getAction.Run(releaseName)
	if err != nil {
		i.Log.Error(err, "ignoring deletion error")
	}
	//
	// if err != nil && err != driver.ErrReleaseNotFound {
	// 	return fmt.Errorf("existing release check failed, %w", err)
	// } else if rel != nil {
	// 	i.Log.Info("release already exists")
	// 	return nil
	// }

	i.Log.Info("removed", "release", rel.Release.Name)
	return nil
}

func (i *HelmInstaller) IsUpgrade(comp *v1alpha1.WaveComponent, inst *Installation) bool {
	if comp.Spec.Version != inst.Version {
		return true
	}
	var newVals map[string]interface{}
	err := yaml.Unmarshal([]byte(comp.Spec.ValuesConfiguration), &newVals)
	if err != nil {
		return true // fail properly later
	}
	if newVals == nil {
		newVals = map[string]interface{}{}
	}
	if !reflect.DeepEqual(newVals, inst.Values) {
		i.Log.Info("upgrade required", "diff", cmp.Diff(newVals, inst.Values))
		return true
	}
	return false
}
