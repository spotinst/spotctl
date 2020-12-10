package wave

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/go-logr/logr"
	"github.com/spotinst/spotctl/internal/wave/box"
	"github.com/spotinst/wave-operator/api/v1alpha1"
	"github.com/spotinst/wave-operator/catalog"
	"github.com/spotinst/wave-operator/install"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	WaveOperatorChart      = "wave-operator"
	WaveOperatorRepository = "https://charts.spot.io"
	WaveOperatorVersion    = "0.1.7"

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
	Create() error
	Describe() error
	// create, get, describe, delete
}

type manager struct {
	clusterID        string
	log              logr.Logger
	kubeClientGetter genericclioptions.RESTClientGetter
}

func NewManager(clusterID string, log logr.Logger) (Manager, error) {
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

	id := cm.Data["spotinst.cluster-identifier"]
	if id != clusterID {
		return nil, fmt.Errorf("error in ocean configuration, cluster id %s != %s", clusterID, id)
	}
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

func (m *manager) loadWaveComponents() ([]*v1alpha1.WaveComponent, error) {
	manifests := box.Boxed.List()
	waveComponents := make([]*v1alpha1.WaveComponent, len(manifests))

	for i, mm := range manifests {
		m.log.Info("loading wave component", "manifest", mm)
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
		m.log.Info("loaded wave component", "name", comp.Name)
	}
	return waveComponents, nil
}

func (m *manager) Create() error {

	waveComponents, err := m.loadWaveComponents()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	err = m.installCertManager(ctx)
	if err != nil {
		return err
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
		wc.Namespace = catalog.SystemNamespace
		err = rc.Create(ctx, wc)
		if err != nil {
			return fmt.Errorf("cannot install component %s, %w", wc.Name, err)
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
	err = wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		wh, err := kc.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, "cert-manager-webhook", metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if wh.Webhooks[0].ClientConfig.CABundle == nil {
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

	installer := install.GetHelm("spotctl", m.kubeClientGetter, m.log)
	err = installer.Install(WaveOperatorChart, WaveOperatorRepository, WaveOperatorVersion, "")
	if err != nil {
		return fmt.Errorf("cannot install wave operator, %w", err)
	}

	err = wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		dep, err := kc.AppsV1().Deployments(catalog.SystemNamespace).Get(ctx, "spotctl-wave-operator", metav1.GetOptions{})

		if err != nil {
			return false, nil
		}
		if dep.Status.AvailableReplicas == 0 {
			return false, nil
		}
		m.log.Info("polled", "deployment", "spotctl-wave-operator", "replicas", dep.Status.AvailableReplicas)

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
