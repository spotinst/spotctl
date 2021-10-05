// Copyright 2021 NetApp, Inc. All Rights Reserved.

package credentials

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ErrSecretCredentialsNotFound is returned when no credentials can be found in Secret.
var ErrSecretCredentialsNotFound = fmt.Errorf("credentials: %s and %s not found "+
	"in Secret", EnvCredentialsToken, EnvCredentialsAccount)

// SecretProvider retrieves credentials from a Secret.
type SecretProvider struct {
	Client          client.Client
	Name, Namespace string
}

// NewSecretProvider returns a new SecretProvider.
func NewSecretProvider(client client.Client, name, namespace string) *SecretProvider {
	return &SecretProvider{
		Client:    client,
		Name:      name,
		Namespace: namespace,
	}
}

// Retrieve retrieves and returns the credentials, or error in case of failure.
func (x *SecretProvider) Retrieve(ctx context.Context) (*Value, error) {
	secret, err := getSecret(ctx, x.Client, x.Name, x.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error retrieving secret %q from "+
			"namespace %q: %w", x.Name, x.Namespace, err)
	}

	value, err := decodeSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("error decoding secret %q from "+
			"namespace %q: %w", x.Name, x.Namespace, err)
	}

	if value.IsEmpty() {
		return value, ErrSecretCredentialsNotFound
	}

	return value, nil
}

// String returns the string representation of the Secret provider.
func (x *SecretProvider) String() string {
	return "SecretProvider"
}

func getSecret(ctx context.Context, client client.Client,
	name, namespace string) (*corev1.Secret, error) {
	obj := new(corev1.Secret)
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := client.Get(ctx, key, obj)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	return obj, nil
}

func decodeSecret(secret *corev1.Secret) (*Value, error) {
	data := make(map[string]string)
	value := new(Value)

	if secret != nil {
		// Copy all non-binary secret data.
		if secret.StringData != nil {
			for k, v := range secret.StringData {
				data[k] = v
			}
		}

		// Copy all binary secret data.
		if secret.Data != nil {
			for k, v := range secret.Data {
				data[k] = string(v)
			}
		}
	}

	return value, mapstructure.Decode(data, value)
}
