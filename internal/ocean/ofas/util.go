package ofas

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spotinst/spotctl/internal/kubernetes"
	"github.com/spotinst/spotctl/internal/ocean/ofas/config"
)

const (
	spotConfigMapNamespace        = metav1.NamespaceSystem
	spotConfigMapName             = "spotinst-kubernetes-cluster-controller-config"
	clusterIdentifierConfigMapKey = "spotinst.cluster-identifier"
)

func ValidateClusterContext(ctx context.Context, clusterIdentifier string) error {
	client, err := kubernetes.GetClient()
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

func CreateDeployerRBAC(ctx context.Context, namespace string) error {
	client, err := kubernetes.GetClient()
	if err != nil {
		return fmt.Errorf("could not get kubernetes client, %w", err)
	}

	sa, crb, err := config.GetDeployerRBAC(namespace)
	if err != nil {
		return fmt.Errorf("could not get deployer rbac objects, %w", err)
	}

	_, err = client.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("could not create deployer service account, %w", err)
	}

	_, err = client.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("could not create deployer cluster role binding, %w", err)
	}

	return nil
}
