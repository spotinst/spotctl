package tide

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/go-logr/logr"
	"github.com/spotinst/wave-operator/api/v1alpha1"
	"github.com/spotinst/wave-operator/catalog"
	"github.com/spotinst/wave-operator/install"
	"github.com/spotinst/wave-operator/internal/version"
	"github.com/spotinst/wave-operator/tide/box"
	v1 "k8s.io/api/core/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	WaveOperatorChart      = "wave-operator"
	WaveOperatorRepository = "https://charts.spot.io"
	WaveOperatorVersion    = "0.1.8"

	CertManagerChart      = "cert-manager"
	CertManagerRepository = "https://charts.jetstack.io"
	CertManagerVersion    = "v1.1.0"
	CertManagerValues     = "installCRDs: true"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensions.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
}

type Manager interface {
	SetConfiguration(k8sProvisioned, oceanClusterProvisioned bool) (*v1alpha1.WaveEnvironment, error)
	GetConfiguration() (*v1alpha1.WaveEnvironment, error)

	Create(env *v1alpha1.WaveEnvironment) error
	Delete() error
}

type manager struct {
	clusterID        string
	log              logr.Logger
	kubeClientGetter genericclioptions.RESTClientGetter
}

func NewManager(log logr.Logger) (Manager, error) {
	conf, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot get cluster configuration, %w", err)
	}

	ctx := context.TODO()
	kc, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to cluster, %w", err)
	}
	cm, err := kc.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(ctx, "spotinst-kubernetes-cluster-controller-config", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error in ocean configuration, %w", err)
	}

	// TODO This is not the Ocean ID
	clusterID := cm.Data["spotinst.cluster-identifier"]
	if clusterID == "" {
		return nil, fmt.Errorf("ocean configuration has no cluster ID")
	}
	log.Info("Reading ocean configuration", "id", clusterID)
	kubeConfig := genericclioptions.NewConfigFlags(false)
	kubeConfig.APIServer = &conf.Host
	kubeConfig.BearerToken = &conf.BearerToken
	kubeConfig.CAFile = &conf.CAFile
	ns := catalog.SystemNamespace
	kubeConfig.Namespace = &ns

	return &manager{
		clusterID:        clusterID,
		log:              log,
		kubeClientGetter: kubeConfig,
	}, nil
}

func (m *manager) getKubernetesClient() (kubernetes.Interface, error) {
	conf, err := m.kubeClientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(conf)
}

func (m *manager) getControllerRuntimeClient() (ctrlrt.Client, error) {
	conf, err := m.kubeClientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	opts := ctrlrt.Options{
		Scheme: scheme,
		Mapper: nil,
	}

	rc, err := ctrlrt.New(conf, opts)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (m *manager) loadCrd(name string) (*apiextensions.CustomResourceDefinition, error) {

	crd := &apiextensions.CustomResourceDefinition{}
	b := box.Boxed.Get(name)
	if b == nil {
		return nil, fmt.Errorf("crd %s not found", name)
	}

	serializer := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := serializer.Decode(b, &schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: runtime.APIVersionInternal,
		Kind:    "CustomResourceDefinition",
	}, crd)
	if err != nil {
		return nil, fmt.Errorf("cannot load crd, %w", err)
	}

	return crd, nil
}

func (m *manager) loadWaveComponents() ([]*v1alpha1.WaveComponent, error) {
	boxed := box.Boxed.List()
	var manifests []string
	for _, n := range boxed {
		// m.log.Info("reading box", "item", n)
		if strings.HasPrefix(n, "/v1alpha1_wavecomponent") {
			manifests = append(manifests, n)
		}
	}
	if len(manifests) == 0 {
		return nil, fmt.Errorf("No wave component manifests found")
	}
	waveComponents := make([]*v1alpha1.WaveComponent, len(manifests))

	for i, mm := range manifests {
		comp := &v1alpha1.WaveComponent{}
		b := box.Boxed.Get(mm)
		serializer := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, _, err := serializer.Decode(b, &schema.GroupVersionKind{
			Group:   "wave.spot.io",
			Version: "v1alpha1",
			Kind:    "WaveComponent",
		}, comp)
		if err != nil {
			return waveComponents, fmt.Errorf("cannot load wave component, %w", err)
		}
		waveComponents[i] = comp
	}
	return waveComponents, nil
}

func (m *manager) SetConfiguration(k8sProvisioned, oceanClusterProvisioned bool) (*v1alpha1.WaveEnvironment, error) {
	ctx := context.TODO()

	kc, err := m.getKubernetesClient()
	if err != nil {
		return nil, err
	}
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: catalog.SystemNamespace,
		},
	}
	_, err = kc.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return nil, err
	}

	certManagerExists, err := m.checkCertManagerPreinstallation()
	if err != nil {
		return nil, fmt.Errorf("can't determine state of certificate manager before installation, %w", err)
	}

	crd, err := m.loadCrd("/wave.spot.io_waveenvironments.yaml")
	if err != nil {
		return nil, err
	}
	ucrd := &unstructured.Unstructured{}
	gv := schema.GroupVersion{
		Group:   "apiextensions.k8s.io",
		Version: runtime.APIVersionInternal,
	}
	if err := scheme.Convert(crd, ucrd, gv); err != nil {
		return nil, fmt.Errorf("failed to convert, %w", err)
	}
	rc, err := m.getControllerRuntimeClient()
	if err != nil {
		return nil, err
	}

	err = rc.Create(ctx, crd, &ctrlrt.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("failed to create crd, %w", err)

	}

	env := &v1alpha1.WaveEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.clusterID,
			Namespace: catalog.SystemNamespace,
		},
		Spec: v1alpha1.WaveEnvironmentSpec{
			OceanClusterId:          m.clusterID,
			OperatorVersion:         version.BuildVersion,
			CertManagerDeployed:     !certManagerExists,
			K8sClusterProvisioned:   k8sProvisioned,
			OceanClusterProvisioned: oceanClusterProvisioned,
		},
	}
	uenv := &unstructured.Unstructured{}
	if err := scheme.Convert(env, uenv, nil); err != nil {
		return nil, err
	}

	err = rc.Create(ctx, uenv)
	return env, nil
}

func (m *manager) GetConfiguration() (*v1alpha1.WaveEnvironment, error) {
	client, err := m.getControllerRuntimeClient()
	if err != nil {
		return nil, err
	}
	env := &v1alpha1.WaveEnvironment{}
	ctx := context.TODO()
	key := ctrlrt.ObjectKey{Name: m.clusterID, Namespace: catalog.SystemNamespace}
	err = client.Get(ctx, key, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (m *manager) Create(env *v1alpha1.WaveEnvironment) error {
	ctx := context.TODO()

	waveComponents, err := m.loadWaveComponents()
	if err != nil {
		return err
	}

	if env.Spec.CertManagerDeployed {
		err = m.installCertManager(ctx)
		if err != nil {
			return err
		}
	}

	err = m.installWaveOperator(ctx)
	if err != nil {
		return err
	}

	rc, err := m.getControllerRuntimeClient()
	if err != nil {
		return fmt.Errorf("kubernetes config error, %w", err)
	}

	for _, wc := range waveComponents {
		m.log.Info("installing wave component", "name", wc.Name)
		wc.Namespace = catalog.SystemNamespace
		err = rc.Create(ctx, wc)
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				m.log.Info("wave component already exists", "name", wc.Name)
			} else {
				return fmt.Errorf("cannot install component %s, %w", wc.Name, err)
			}
		}
	}

	return nil
}

func (m *manager) Delete() error {

	ctx := context.TODO()

	rc, err := m.getControllerRuntimeClient()
	if err != nil {
		return fmt.Errorf("kubernetes config error, %w", err)
	}

	components := &v1alpha1.WaveComponentList{}
	err = rc.List(ctx, components)
	if err != nil {
		crdGone, ok := err.(*apimeta.NoKindMatchError)
		if ok {
			m.log.Info("WaveComponent CRD is not present", "message", crdGone.Error())
		} else {
			return err
		}
	} else {
		for _, wc := range components.Items {
			rc.Delete(ctx, &wc)
		}
	}

	err = wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		for _, wc := range components.Items {
			obj := &v1alpha1.WaveComponent{}
			key := types.NamespacedName{
				Namespace: wc.Namespace,
				Name:      wc.Name,
			}
			// wait for IsNotFound on all wavecomponents
			err := rc.Get(ctx, key, obj)
			if err == nil {
				return false, nil
			} else if !k8serrors.IsNotFound(err) {
				return false, err
			}
		}
		return true, nil
	})

	err = m.deleteWaveOperator(ctx)
	if err != nil {
		return err
	}

	env, err := m.GetConfiguration()
	if err != nil {
		return fmt.Errorf("unable to read wave environment, %w", err)
	}

	if env.Spec.CertManagerDeployed {
		err = m.deleteCertManager(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) installCertManager(ctx context.Context) error {
	kc, err := m.getKubernetesClient()
	if err != nil {
		return err
	}
	certNS := CertManagerChart // chart name == namespace
	_, _ = kc.CoreV1().Namespaces().Create(
		ctx,
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: certNS}},
		metav1.CreateOptions{},
	)
	installer := install.GetHelm("", m.kubeClientGetter, m.log)
	installer.SetNamespace(certNS)
	err = installer.Install(CertManagerChart, CertManagerRepository, CertManagerVersion, CertManagerValues)
	if err != nil {
		return fmt.Errorf("cannot install cert manager, %w", err)
	}

	// webhook must have cert and endpoint before we can proceed
	// Exited with error: cannot install wave operator, installation error, Internal error occurred: failed calling webhook "webhook.cert-manager.io": Post https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s: no endpoints available for service "cert-manager-webhook"

	err = wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		wh, err := kc.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, "cert-manager-webhook", metav1.GetOptions{})
		if err != nil || wh.Webhooks[0].ClientConfig.CABundle == nil {
			return false, nil
		}
		ep, err := kc.CoreV1().Endpoints(certNS).Get(ctx, "cert-manager-webhook", metav1.GetOptions{})
		if err != nil || len(ep.Subsets) == 0 || len(ep.Subsets[0].Addresses) == 0 {
			return false, nil
		}
		m.log.Info("polled", "webhook", "cert-manager-webhook", "name", wh.Webhooks[0].Name)

		return true, nil
	})
	return err
}

func (m *manager) installWaveOperator(ctx context.Context) error {
	kc, err := m.getKubernetesClient()
	if err != nil {
		return err
	}

	_, _ = kc.CoreV1().Namespaces().Create(
		ctx,
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: catalog.SystemNamespace}},
		metav1.CreateOptions{},
	)

	installer := install.GetHelm("", m.kubeClientGetter, m.log)
	err = installer.Install(WaveOperatorChart, WaveOperatorRepository, WaveOperatorVersion, "")
	if err != nil {
		return fmt.Errorf("cannot install wave operator, %w", err)
	}

	err = wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		dep, err := kc.AppsV1().Deployments(catalog.SystemNamespace).Get(ctx, "wave-operator", metav1.GetOptions{})
		if err != nil || dep.Status.AvailableReplicas == 0 {
			return false, nil
		}
		m.log.Info("polled", "deployment", "wave-operator", "replicas", dep.Status.AvailableReplicas)

		return true, nil
	})
	return err
}

func (m *manager) Describe() error {

	rc, err := m.getControllerRuntimeClient()
	if err != nil {
		return fmt.Errorf("kubernetes config error, %w", err)
	}
	ctx := context.TODO()
	components := &v1alpha1.WaveComponentList{}
	err = rc.List(ctx, components)
	if err != nil {
		return fmt.Errorf("cannot list wave components, %w", err)
	}

	width := 20
	writer := tabwriter.NewWriter(os.Stdout, width, 8, 1, '\t', tabwriter.AlignRight)
	bar := strings.Repeat("-", width)
	boundary := bar + "\t" + bar + "\t" + bar + "\t" + bar
	fmt.Fprintln(writer, "component\tcondition\tproperty\tvalue")
	fmt.Fprintln(writer, boundary)
	for _, wc := range components.Items {
		sort.Slice(wc.Status.Conditions, func(i, j int) bool {
			return wc.Status.Conditions[i].LastUpdateTime.Time.After(wc.Status.Conditions[j].LastUpdateTime.Time)
		})
		condition := "Unknown"
		if len(wc.Status.Conditions) > 0 {
			condition = fmt.Sprintf("%s=%s", wc.Status.Conditions[0].Type, wc.Status.Conditions[0].Status)
			// m.log.Info("         ", "condition", fmt.Sprintf("%s=%s", wc.Status.Conditions[0].Type, wc.Status.Conditions[0].Status))
		}
		if len(wc.Status.Properties) == 0 {
			fmt.Fprintln(writer, wc.Name+"\t"+condition+"\t\t")
		} else {
			h := wc.Name + "\t" + condition
			for k, v := range wc.Status.Properties {
				fmt.Fprintln(writer, h+"\t"+k+"\t"+v)
				h = "\t"
			}
		}
		fmt.Fprintln(writer, boundary)
	}
	writer.Flush()
	return nil
}

func (m *manager) deleteWaveOperator(ctx context.Context) error {
	kc, err := m.getKubernetesClient()
	if err != nil {
		return err
	}

	installer := install.GetHelm("", m.kubeClientGetter, m.log)
	err = installer.Delete(WaveOperatorChart, WaveOperatorRepository, WaveOperatorVersion, "")
	if err != nil {
		return fmt.Errorf("cannot delete wave operator, %w", err)
	}

	err = wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		_, err := kc.AppsV1().Deployments(catalog.SystemNamespace).Get(ctx, "spotctl-wave-operator", metav1.GetOptions{})
		if err == nil {
			return false, nil
		} else if !k8serrors.IsNotFound(err) {
			return false, err
		}
		return true, nil
	})
	return err
}

func (m *manager) deleteCertManager(ctx context.Context) error {
	kc, err := m.getKubernetesClient()
	if err != nil {
		return err
	}
	certNS := CertManagerChart // chart name == namespace

	installer := install.GetHelm("", m.kubeClientGetter, m.log)
	installer.SetNamespace(certNS)
	err = installer.Delete(CertManagerChart, CertManagerRepository, CertManagerVersion, CertManagerValues)
	if err != nil {
		return fmt.Errorf("cannot delete wave operator, %w", err)
	}

	err = wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		_, err := kc.AppsV1().Deployments(certNS).Get(ctx, "cert-manager", metav1.GetOptions{})
		if err == nil {
			return false, nil
		} else if !k8serrors.IsNotFound(err) {
			return false, err
		}
		return true, nil
	})
	return err
}
