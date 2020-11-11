package wave

import (
	"context"
	"fmt"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"github.com/go-logr/logr"
	"github.com/spotinst/spotctl/internal/wave/box"
	"github.com/spotinst/wave-operator/api/v1alpha1"
	"github.com/spotinst/wave-operator/catalog"
	"github.com/spotinst/wave-operator/install"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	// "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	WaveOperatorChart      = "wave-operator"
	WaveOperatorRepository = "https://ntfrnzn.github.io/charts/"
	WaveOperatorVersion    = "0.1.1"
)

var (
	scheme   = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensions.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
}

type Manager interface {
	Create() error
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
		return nil, err
	}

	ctx := context.TODO()
	kc, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}
	cm, err := kc.CoreV1().ConfigMaps("kube-system").Get(ctx, "spotinst-kubernetes-cluster-controller-config", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	id := cm.Data["spotinst.cluster-identifier"]
	if id != clusterID {
		return nil, fmt.Errorf("mismatch in cluster id, %s != %s", clusterID, id)
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

func (m *manager) Create() error {

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
			return err
		}
		waveComponents[i] = comp
		m.log.Info("loaded wave component", "name", comp.Name)
	}


	installer := install.GetHelm("spotctl", m.kubeClientGetter, m.log)
	err := installer.Install(WaveOperatorChart, WaveOperatorRepository, WaveOperatorVersion, "")
	if err != nil {
		return err
	}

	conf, err := m.kubeClientGetter.ToRESTConfig()
	if err != nil {
		return err
	}

	opts :=ctrlrt.Options{
		Scheme: scheme,
		Mapper: nil,
	}

	rc, err := ctrlrt.New(conf, opts)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	for _, wc := range waveComponents {
		wc.Namespace = catalog.SystemNamespace
		err = rc.Create(ctx, wc)
		if err != nil {
			return err
		}
	}

	//m.kubeClientGetter.

	return nil
}
