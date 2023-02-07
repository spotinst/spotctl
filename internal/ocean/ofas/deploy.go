package ofas

import (
	"context"
	"fmt"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotinst/spotctl/internal/ocean/ofas/config"
)

const (
	spotConfigMapNamespace        = metav1.NamespaceSystem
	spotConfigMapName             = "spotinst-kubernetes-cluster-controller-config"
	clusterIdentifierConfigMapKey = "spotinst.cluster-identifier"

	pollInterval = 5 * time.Second
	pollTimeout  = 5 * time.Minute
)

func ValidateClusterContext(ctx context.Context, client kubernetes.Interface, clusterIdentifier string) error {
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

func CreateDeployerRBAC(ctx context.Context, client kubernetes.Interface, namespace string) error {
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
