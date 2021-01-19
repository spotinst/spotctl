package wave

import (
	"context"
	"fmt"

	"github.com/spotinst/wave-operator/api/v1alpha1"
	"github.com/spotinst/wave-operator/catalog"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	spotConfigMapNamespace        = metav1.NamespaceSystem
	spotConfigMapName             = "spotinst-kubernetes-cluster-controller-config"
	clusterIdentifierConfigMapKey = "spotinst.cluster-identifier"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensions.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
}

func getKubernetesClient() (kubernetes.Interface, error) {

	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getControllerRuntimeClient() (ctrlrt.Client, error) {

	conf, err := config.GetConfig()
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

func ListComponents() (*v1alpha1.WaveComponentList, error) {

	ctx := context.TODO()

	rc, err := getControllerRuntimeClient()
	if err != nil {
		return nil, fmt.Errorf("could not get controller runtime client, %w", err)
	}

	components := &v1alpha1.WaveComponentList{}
	options := &ctrlrt.ListOptions{
		Namespace: catalog.SystemNamespace,
	}
	err = rc.List(ctx, components, options)
	if err != nil {
		return nil, fmt.Errorf("could not list wave components, %w", err)
	}

	return components, nil
}

func ValidateClusterContext(clusterIdentifier string) error {

	ctx := context.TODO()

	client, err := getKubernetesClient()
	if err != nil {
		return fmt.Errorf("could not get kubernetes client, %w", err)
	}

	cm, err := client.CoreV1().ConfigMaps(spotConfigMapNamespace).Get(ctx, spotConfigMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not get ocean configuration, %w", err)
	}

	id := cm.Data[clusterIdentifierConfigMapKey]
	if id != clusterIdentifier {
		return fmt.Errorf("current cluster identifier is %q, expected %q", id, clusterIdentifier)
	}

	return nil
}
