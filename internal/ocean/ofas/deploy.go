package ofas

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/spotinst/spotctl/internal/kubernetes"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/spotinst/spotctl/internal/ocean/ofas/config"
	"github.com/spotinst/spotctl/internal/uuid"
)

const (
	spotConfigMapNamespace        = metav1.NamespaceSystem
	spotConfigMapName             = "spotinst-kubernetes-cluster-controller-config"
	clusterIdentifierConfigMapKey = "spotinst.cluster-identifier"

	pollInterval = 5 * time.Second
	pollTimeout  = 5 * time.Minute
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

type jobValues struct {
	Name            string
	Namespace       string
	ImagePullSecret string
	ImageDeployer   string
	ImageOperator   string
	ServiceAccount  string
}

func Deploy(ctx context.Context, namespace string) error {
	// TODO This is temporary
	// We should call a /deploy API on the backend. The backend will then run the deployment on the cluster

	client, err := kubernetes.GetClient()
	if err != nil {
		return fmt.Errorf("could not get kubernetes client, %w", err)
	}

	values := jobValues{
		Name:            fmt.Sprintf("ofas-deploy-%s", uuid.NewV4().Short()),
		Namespace:       namespace,
		ImagePullSecret: "bigdata-dev-regcred",
		ImageDeployer:   "598800841386.dkr.ecr.us-east-2.amazonaws.com/private/bigdata-deployer:0.1.1-c31ad4f8",
		ImageOperator:   "598800841386.dkr.ecr.us-east-2.amazonaws.com/private/bigdata-operator:0.1.1-c31ad4f8",
		ServiceAccount:  config.ServiceAccountName,
	}

	jobTemplate, err := template.New("deployJob").Parse(deployJobTemplate)
	if err != nil {
		return fmt.Errorf("could not parse job template, %w", err)
	}

	jobManifestBytes := new(bytes.Buffer)
	err = jobTemplate.Execute(jobManifestBytes, values)
	if err != nil {
		return fmt.Errorf("could not execute job template, %w", err)
	}

	jobManifest := jobManifestBytes.String()

	job := &batchv1.Job{}
	err = yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(jobManifest), len(jobManifest)).Decode(job)
	if err != nil {
		return fmt.Errorf("could not decode job manifest, %w", err)
	}

	log.Debugf("Creating deploy job %s/%s", job.Namespace, job.Name)
	createdJob, err := client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("could not create deploy job, %w", err)
	}

	err = wait.Poll(pollInterval, pollTimeout, func() (bool, error) {
		job, err := client.BatchV1().Jobs(createdJob.Namespace).Get(ctx, createdJob.Name, metav1.GetOptions{})
		if err != nil {
			log.Debugf("Could not get deploy job, err: %s", err.Error())
			return false, nil
		}

		activePods := job.Status.Active
		failedPods := job.Status.Failed
		succeededPods := job.Status.Succeeded
		log.Debugf("Deploy job pods - active: %d, succeeded: %d, failed: %d", activePods, succeededPods, failedPods)

		// TODO Should check conditions instead
		if activePods == 0 && succeededPods > 0 {
			log.Debugf("Deploy job complete")
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return fmt.Errorf("wait for deploy job completion failed, %w", err)
	}

	// TODO Verify that this deletes the job pods
	log.Debugf("Deleting deploy job %s/%s", job.Namespace, job.Name)
	if err := client.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{}); err != nil {
		// Best effort
		log.Warnf("Could not delete deploy job %s/%s", job.Namespace, job.Name)
	}

	return nil
}

const deployJobTemplate = `apiVersion: batch/v1
kind: Job
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  ttlSecondsAfterFinished: 300
  template:
    spec:
      imagePullSecrets:
      - name: {{.ImagePullSecret}}
      containers:
        - image:
            {{.ImageDeployer}}
          name: deployer
          args:
            - install
            - --create-bootstrap-environment
            - --image
            - {{.ImageOperator}}
            - --image-pull-secret
            - {{.ImagePullSecret}}
            - --image-pull-policy
            - Always
          resources: { }
          imagePullPolicy: Always
      serviceAccountName: {{.ServiceAccount}}
      restartPolicy: Never
`
