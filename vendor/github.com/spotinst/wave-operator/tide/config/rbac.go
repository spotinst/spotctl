package config

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	ServiceAccountName = "tide"
	RoleBindingName    = "tide-helmadmin"
)

type rbacValues struct {
	ServiceAccountName      string
	ServiceAccountNamespace string
	RoleBindingName         string
}

type RBACManifests struct {
	ServiceAccount     string
	ClusterRoleBinding string
}

func GetRBACManifests(namespace string) (*RBACManifests, error) {

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

	return &RBACManifests{
		ServiceAccount:     saManifest.String(),
		ClusterRoleBinding: rbManifest.String(),
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
