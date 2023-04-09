package ofas

import (
	"context"
	"fmt"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotinst/spotctl/internal/ocean/ofas/config"
)

const (
	spotConfigMapName             = "spotinst-kubernetes-cluster-controller-config"
	clusterIdentifierConfigMapKey = "spotinst.cluster-identifier"
)

func ValidateClusterContext(ctx context.Context, client kubernetes.Interface, clusterIdentifier string) error {
	fieldSelectorForName := fmt.Sprintf("metadata.name=%s", spotConfigMapName)
	cm, err := client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelectorForName,
	})
	if err != nil {
		return fmt.Errorf("could not get ocean configuration, %w", err)
	}

	switch len(cm.Items) {
	case 0:
		return fmt.Errorf("config map %q not found", spotConfigMapName)
	case 1:
		id := cm.Items[0].Data[clusterIdentifierConfigMapKey]
		if id != clusterIdentifier {
			return fmt.Errorf("current cluster identifier is %q, expected %q", id, clusterIdentifier)
		}
		return nil
	default:
		for i := range cm.Items {
			id := cm.Items[i].Data[clusterIdentifierConfigMapKey]
			if id == clusterIdentifier {
				return nil
			}
		}
		return fmt.Errorf("current cluster identifier is %q, expected %q", cm.Items[0].Data[clusterIdentifierConfigMapKey], clusterIdentifier)
	}
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
