// Copyright 2021 NetApp, Inc. All Rights Reserved.

package tide

import (
	"context"
	"fmt"

	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
	"github.com/spotinst/ocean-operator/pkg/config"
	"github.com/spotinst/ocean-operator/pkg/credentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewConfigFlags(config *rest.Config, namespace string) *genericclioptions.ConfigFlags {
	cf := genericclioptions.NewConfigFlags(true)
	cf.APIServer = &config.Host
	cf.BearerToken = &config.BearerToken
	cf.CAFile = &config.CAFile
	cf.Namespace = &namespace
	return cf
}

func NewControllerRuntimeClient(config *rest.Config, scheme *runtime.Scheme) (client.Client, error) {
	return client.New(config, client.Options{
		Scheme: scheme,
		Mapper: nil,
	})
}

func LoadConfig(ctx context.Context, client client.Client) (*config.Value, error) {
	providers := []config.Provider{
		&config.ConfigMapProvider{
			Client:    client,
			Name:      OceanOperatorConfigMap,
			Namespace: oceanv1alpha1.NamespaceSystem,
		},
		&config.ConfigMapProvider{
			Client:    client,
			Name:      OceanOperatorConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		&config.ConfigMapProvider{
			Client:    client,
			Name:      OceanOperatorConfigMap,
			Namespace: metav1.NamespaceDefault,
		},
		&config.ConfigMapProvider{
			Client:    client,
			Name:      LegacyOceanControllerConfigMap,
			Namespace: metav1.NamespaceDefault,
		},
		&config.EnvProvider{},
	}

	value, err := config.NewConfig(config.NewChainProvider(providers...)).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return value, nil
}

func LoadCredentials(ctx context.Context, client client.Client) (*credentials.Value, error) {
	providers := []credentials.Provider{
		&credentials.SecretProvider{
			Client:    client,
			Name:      OceanOperatorSecret,
			Namespace: oceanv1alpha1.NamespaceSystem,
		},
		&credentials.SecretProvider{
			Client:    client,
			Name:      OceanOperatorSecret,
			Namespace: metav1.NamespaceSystem,
		},
		&credentials.SecretProvider{
			Client:    client,
			Name:      OceanOperatorSecret,
			Namespace: metav1.NamespaceDefault,
		},
		&credentials.SecretProvider{
			Client:    client,
			Name:      LegacyOceanControllerSecret,
			Namespace: metav1.NamespaceSystem,
		},
		&credentials.EnvProvider{},
	}

	value, err := credentials.NewCredentials(credentials.NewChainProvider(providers...)).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	return value, nil
}
