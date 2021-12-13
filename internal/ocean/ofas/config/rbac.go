package config

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	ServiceAccountName = "bigdata-deployer"
	RoleBindingName    = "bigdata-deployer-admin"
)

type rbacValues struct {
	ServiceAccountName      string
	ServiceAccountNamespace string
	RoleBindingName         string
}

type rbacManifests struct {
	serviceAccount     string
	clusterRoleBinding string
}

func GetDeployerRBAC(namespace string) (*corev1.ServiceAccount, *rbacv1.ClusterRoleBinding, error) {
	manifests, err := getRBACManifests(namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get rbac manifests, %w", err)
	}

	sa := &corev1.ServiceAccount{}
	err = yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(manifests.serviceAccount), len(manifests.serviceAccount)).Decode(sa)
	if err != nil {
		return nil, nil, fmt.Errorf("could not decode service account yaml, %w", err)
	}

	crb := &rbacv1.ClusterRoleBinding{}
	err = yamlutil.NewYAMLOrJSONDecoder(strings.NewReader(manifests.clusterRoleBinding), len(manifests.clusterRoleBinding)).Decode(crb)
	if err != nil {
		return nil, nil, fmt.Errorf("could not decode cluster role binding yaml, %w", err)
	}

	return sa, crb, nil
}

func getRBACManifests(namespace string) (*rbacManifests, error) {
	values := rbacValues{
		ServiceAccountName:      ServiceAccountName,
		ServiceAccountNamespace: namespace,
		RoleBindingName:         RoleBindingName,
	}

	saTemplate, err := template.New("sa").Parse(serviceAccountTemplate)
	if err != nil {
		return nil, fmt.Errorf("could not parse service account template, %w", err)
	}

	saManifest := new(bytes.Buffer)
	err = saTemplate.Execute(saManifest, values)
	if err != nil {
		return nil, fmt.Errorf("could not execute service account template, %w", err)
	}

	rbTemplate, err := template.New("roleBinding").Parse(roleBindingTemplate)
	if err != nil {
		return nil, fmt.Errorf("could not parse role binding template, %w", err)
	}

	rbManifest := new(bytes.Buffer)
	err = rbTemplate.Execute(rbManifest, values)
	if err != nil {
		return nil, fmt.Errorf("could not execute role binding template, %w", err)
	}

	return &rbacManifests{
		serviceAccount:     saManifest.String(),
		clusterRoleBinding: rbManifest.String(),
	}, nil
}

const serviceAccountTemplate = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.ServiceAccountName}}
  namespace: {{.ServiceAccountNamespace}}
`

const roleBindingTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.RoleBindingName}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: {{.ServiceAccountName}}
    namespace: {{.ServiceAccountNamespace}}
`
