package ofas

import (
	"context"
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/spotinst/spotctl/internal/kubernetes"
	"github.com/spotinst/spotctl/internal/ocean/ofas/config"
	"github.com/spotinst/spotctl/internal/uuid"
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

func Deploy(ctx context.Context, namespace string) error {
	// TODO This is temporary
	// We should call a /deploy API on the backend. The backend will then run the deployment on the cluster

	client, err := kubernetes.GetClient()
	if err != nil {
		return fmt.Errorf("could not get kubernetes client, %w", err)
	}

	job := &batchv1.Job{}
	err = yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(deployJob), len(deployJob)).Decode(job)
	if err != nil {
		return fmt.Errorf("could not decode job yaml, %w", err)
	}

	job.Name = fmt.Sprintf("ofas-deployer-install-%s", uuid.NewV4().Short())
	job.Namespace = namespace
	job.Spec.Template.Spec.ServiceAccountName = config.ServiceAccountName

	createdJob, err := client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("could not create deploy job, %w", err)
	}

	fmt.Println(createdJob.Name)

	return nil
}

const deployJob = `apiVersion: batch/v1
kind: Job
metadata:
  name: job-name
  namespace: job-ns
spec:
  template:
    spec:
      imagePullSecrets:
      - name: bigdata-dev-regcred
      containers:
        - image:
            598800841386.dkr.ecr.us-east-2.amazonaws.com/private/bigdata-deployer:0.1.1-c31ad4f8
          name: deployer
          args:
            - install
            - --create-bootstrap-environment
            - --image
            - 598800841386.dkr.ecr.us-east-2.amazonaws.com/private/bigdata-operator:0.1.1-c31ad4f8
            - --image-pull-secret
            - bigdata-dev-regcred
            - --image-pull-policy
            - Always
          resources: { }
          imagePullPolicy: Always
      serviceAccountName: job-sa
      restartPolicy: Never
`
