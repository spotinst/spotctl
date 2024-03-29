package ofas

import (
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	rbakv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"

	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/ocean/ofas/config"
)

const (
	spotConfigMapNamespace        = metav1.NamespaceSystem
	spotConfigMapName             = "spotinst-kubernetes-cluster-controller-config"
	clusterIdentifierConfigMapKey = "spotinst.cluster-identifier"
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

func CreateDeployerRBAC(ctx context.Context, client kubernetes.Interface) error {
	deploymentYaml, err := config.GetDeploymentYaml()
	if err != nil {
		return fmt.Errorf("could not get deployment yaml, %w", err)
	}
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(deploymentYaml), 4096)

	for {
		// Decode the YAML into an unstructured object
		var obj unstructured.Unstructured
		if err := decoder.Decode(&obj); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("could not decode yaml, %w", err)
		}

		// Handle each resource type
		switch obj.GetKind() {
		case "Namespace":
			ns := &corev1.Namespace{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, ns); err != nil {
				return fmt.Errorf("could not decode namespace yaml, %w", err)
			}
			_, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
			if err != nil && !k8serrors.IsAlreadyExists(err) {
				return fmt.Errorf("could not create namespace, %w", err)
			}

		case "ServiceAccount":
			sa := &corev1.ServiceAccount{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, sa); err != nil {
				return fmt.Errorf("could not decode service account yaml, %w", err)
			}
			_, err := client.CoreV1().ServiceAccounts(sa.Namespace).Create(ctx, sa, metav1.CreateOptions{})
			if err != nil && !k8serrors.IsAlreadyExists(err) {
				return fmt.Errorf("could not create service account, %w", err)
			}

		case "RoleBinding":
			rb := &rbakv1.RoleBinding{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, rb); err != nil {
				return fmt.Errorf("could not decode role binding yaml, %w", err)
			}
			_, err := client.RbacV1().RoleBindings(rb.Namespace).Create(ctx, rb, metav1.CreateOptions{})
			if err != nil && !k8serrors.IsAlreadyExists(err) {
				return fmt.Errorf("could not create role binding, %w", err)
			}

		case "ClusterRole":
			cr := &rbakv1.ClusterRole{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, cr); err != nil {
				return fmt.Errorf("could not decode cluster role yaml, %w", err)
			}
			_, err := client.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
			if err != nil && !k8serrors.IsAlreadyExists(err) {
				return fmt.Errorf("could not create cluster role, %w", err)
			}

		case "ClusterRoleBinding":
			crb := &rbakv1.ClusterRoleBinding{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, crb); err != nil {
				return fmt.Errorf("could not decode cluster role binding yaml, %w", err)
			}
			_, err := client.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
			if err != nil && !k8serrors.IsAlreadyExists(err) {
				return fmt.Errorf("could not create cluster role binding, %w", err)
			}

		default:
			log.Warnf("Skipping unsupported resource type %q", obj.GetKind())
		}
	}

	return nil
}
