package kubernetes

import (
	"context"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetDefaultKubeConfigPath() string {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

func GetConfig(kubeConfigPath string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("could not build kube config, %w", err)
	}
	return config, nil
}

func GetClient(config *rest.Config) (kubernetes.Interface, error) {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create client, %w", err)
	}
	return client, nil
}

// EnsureNamespace checks for the existence of a namespace with the given name.
// If the namespace does not exist, it is created.
func EnsureNamespace(ctx context.Context, client kubernetes.Interface, name string) error {
	namespace := &corev1.Namespace{}
	namespace.Name = name

	_, err := client.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("could not create namespace %q, %w", name, err)
	}

	return nil
}
