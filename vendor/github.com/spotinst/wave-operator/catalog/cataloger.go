package catalog

import (
	"context"

	"github.com/spotinst/wave-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SystemNamespace    = "spot-system"
	SparkJobsNamespace = "spark-jobs"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
}

type Cataloger interface {

	// List returns the set of WaveComponents that are in the system
	List() (*v1alpha1.WaveComponentList, error)

	// Get returns a WaveComponent by name
	Get(name string) (*v1alpha1.WaveComponent, error)

	// Update applies changes to the spec of a WaveComponent
	Update(component *v1alpha1.WaveComponent) error
}

func NewCataloger() (Cataloger, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	client, err := ctrlclient.New(
		config,
		ctrlclient.Options{
			Scheme: scheme,
		},
	)

	return &cataloger{
		client: client,
	}, nil
}

type cataloger struct {
	client ctrlclient.Client
}

func (c cataloger) List() (*v1alpha1.WaveComponentList, error) {
	ctx := context.TODO()
	list := v1alpha1.WaveComponentList{}
	options := &ctrlclient.ListOptions{
		LabelSelector: nil,
		FieldSelector: nil,
		Namespace:     SystemNamespace,
	}
	err := c.client.List(ctx, &list, options)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (c cataloger) Get(name string) (*v1alpha1.WaveComponent, error) {
	ctx := context.TODO()
	comp := v1alpha1.WaveComponent{}
	err := c.client.Get(ctx, ctrlclient.ObjectKey{Name: name, Namespace: SystemNamespace}, &comp)
	if err != nil {
		return nil, err
	}
	// update status? probably not, read should not affect the system
	// operator should be keeping the status up to date
	return &comp, nil
}

func (c cataloger) Update(component *v1alpha1.WaveComponent) error {
	panic("implement me")
}
