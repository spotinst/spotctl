package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GetClient() (kubernetes.Interface, error) {
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

// EnsureNamespace checks for the existence of a namespace with the given name.
// If the namespace does not exist, it is created.
func EnsureNamespace(ctx context.Context, name string) error {
	client, err := GetClient()
	if err != nil {
		return fmt.Errorf("could not get client, %w", err)
	}

	namespace := &corev1.Namespace{}
	namespace.Name = name

	_, err = client.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("could not create namespace %q, %w", name, err)
	}

	return nil
}
