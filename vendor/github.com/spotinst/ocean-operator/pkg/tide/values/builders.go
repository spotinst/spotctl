// Copyright 2021 NetApp, Inc. All Rights Reserved.

package values

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spotinst/ocean-operator/pkg/config"
	"github.com/spotinst/ocean-operator/pkg/credentials"
	"github.com/spotinst/ocean-operator/pkg/tide"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder defines the interface used by chart builders.
type Builder interface {
	// Build builds chart values.
	Build(ctx context.Context) (string, error)
}

// Blank assignments to verify that all builders implement the Builder interface.
var (
	_ Builder = new(OceanOperatorBuilder)
	_ Builder = new(OceanControllerBuilder)
)

// region Base Builder

// OceanBaseBuilder builds values passed to Helm charts.
type OceanBaseBuilder struct {
	credentials *credentials.Value
	config      *config.Value
	client      client.Client
}

// NewOceanBaseBuilder returns a new OceanBaseBuilder.
func NewOceanBaseBuilder() *OceanBaseBuilder {
	return new(OceanBaseBuilder)
}

// WithCredentials sets the credentials that should be used.
func (b *OceanBaseBuilder) WithCredentials(value *credentials.Value) *OceanBaseBuilder {
	b.credentials = value
	return b
}

// WithToken sets the value for `spotinst.token`.
// It's a shorthand for WithCredentials(&Value{Token:"redacted"}).
func (b *OceanBaseBuilder) WithToken(value string) *OceanBaseBuilder {
	if b.credentials == nil {
		b.credentials = new(credentials.Value)
	}
	b.credentials.Token = value
	return b
}

// WithAccount sets the value for `spotinst.account`.
// It's a shorthand for WithCredentials(&Value{Account:"redacted"}).
func (b *OceanBaseBuilder) WithAccount(value string) *OceanBaseBuilder {
	if b.credentials == nil {
		b.credentials = new(credentials.Value)
	}
	b.credentials.Account = value
	return b
}

// WithConfig sets the config that should be used.
func (b *OceanBaseBuilder) WithConfig(value *config.Value) *OceanBaseBuilder {
	b.config = value
	return b
}

// WithClusterIdentifier sets the value for `oceanController.clusterIdentifier`.
// It's a shorthand for WithConfig(&Value{ClusterIdentifier:"redacted"}).
func (b *OceanBaseBuilder) WithClusterIdentifier(value string) *OceanBaseBuilder {
	if b.config == nil {
		b.config = new(config.Value)
	}
	b.config.ClusterIdentifier = value
	return b
}

// WithACDIdentifier sets the value for `oceanController.acdIdentifier`.
// It's a shorthand for WithConfig(&Value{ACDIdentifier:"redacted"}).
func (b *OceanBaseBuilder) WithACDIdentifier(value string) *OceanBaseBuilder {
	if b.config == nil {
		b.config = new(config.Value)
	}
	b.config.ACDIdentifier = value
	return b
}

// WithClient sets the client that should be used to fetch in-cluster config/credentials.
func (b *OceanBaseBuilder) WithClient(client client.Client) *OceanBaseBuilder {
	b.client = client
	return b
}

// Complete completes the setup of the builder.
func (b *OceanBaseBuilder) Complete(ctx context.Context) error {
	var err error

	if b.credentials == nil && b.client != nil {
		b.credentials, err = tide.LoadCredentials(ctx, b.client)
		if err != nil {
			return err
		}
	}

	if b.config == nil && b.client != nil {
		b.config, err = tide.LoadConfig(ctx, b.client)
		if err != nil {
			return err
		}
	}

	return nil
}

// endregion

// region Ocean Operator Builder

type OceanOperatorBuilder struct {
	*OceanBaseBuilder
	components []string
}

func NewOceanOperatorBuilder(base *OceanBaseBuilder) *OceanOperatorBuilder {
	return &OceanOperatorBuilder{
		OceanBaseBuilder: base,
	}
}

// WithComponents sets the value for `bootstrap.components`.
func (b *OceanOperatorBuilder) WithComponents(components []string) *OceanOperatorBuilder {
	b.components = components
	return b
}

func (b *OceanOperatorBuilder) Build(ctx context.Context) (string, error) {
	if err := b.Complete(ctx); err != nil {
		return "", err
	}

	values := &valuesOceanOperator{
		Spotinst: &valuesOceanOperatorSpotinst{
			Token:             b.credentials.Token,
			Account:           b.credentials.Account,
			ClusterIdentifier: b.config.ClusterIdentifier,
			ACDIdentifier:     b.config.ACDIdentifier,
		},
		Bootstrap: &valuesOceanOperatorBootstrap{
			Components: b.components,
		},
	}

	o, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("failed to marshal values: %w", err)
	}

	return string(o), nil
}

// endregion

// region Ocean Controller Builder

type OceanControllerBuilder struct {
	*OceanBaseBuilder
}

func NewOceanControllerBuilder(base *OceanBaseBuilder) *OceanControllerBuilder {
	return &OceanControllerBuilder{
		OceanBaseBuilder: base,
	}
}

func (b *OceanControllerBuilder) Build(ctx context.Context) (string, error) {
	if err := b.Complete(ctx); err != nil {
		return "", err
	}

	values := &valuesOceanController{
		Spotinst: &valuesOceanControllerSpotinst{
			Token:             b.credentials.Token,
			Account:           b.credentials.Account,
			ClusterIdentifier: b.config.ClusterIdentifier,
		},
		Connector: &valuesOceanControllerConnector{
			ACDIdentifier: b.config.ACDIdentifier,
		},
	}

	o, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("failed to marshal values: %w", err)
	}

	return string(o), nil
}

// endregion

// region Types

type (
	valuesOceanOperatorSpotinst struct {
		Token             string `json:"token" yaml:"token"`
		Account           string `json:"account" yaml:"account"`
		ClusterIdentifier string `json:"clusterIdentifier" yaml:"clusterIdentifier"`
		ACDIdentifier     string `json:"acdIdentifier" yaml:"acdIdentifier"`
	}

	valuesOceanOperatorBootstrap struct {
		Components []string `json:"components" yaml:"components"`
	}

	valuesOceanOperator struct {
		Spotinst  *valuesOceanOperatorSpotinst  `json:"spotinst" yaml:"spotinst"`
		Bootstrap *valuesOceanOperatorBootstrap `json:"bootstrap" yaml:"bootstrap"`
	}
)

func (v *valuesOceanOperator) Valid() bool {
	return v.Spotinst != nil &&
		v.Spotinst.Token != "" &&
		v.Spotinst.Account != "" &&
		v.Spotinst.ClusterIdentifier != ""
}

type (
	valuesOceanControllerSpotinst struct {
		Token             string `json:"token" yaml:"token"`
		Account           string `json:"account" yaml:"account"`
		ClusterIdentifier string `json:"clusterIdentifier" yaml:"clusterIdentifier"`
	}

	valuesOceanControllerConnector struct {
		ACDIdentifier string `json:"acdIdentifier" yaml:"acdIdentifier"`
	}

	valuesOceanController struct {
		Spotinst  *valuesOceanControllerSpotinst  `json:"spotinst" yaml:"spotinst"`
		Connector *valuesOceanControllerConnector `json:"aksConnector" yaml:"aksConnector"`
	}
)

func (v *valuesOceanController) Valid() bool {
	return v.Spotinst != nil &&
		v.Spotinst.Token != "" &&
		v.Spotinst.Account != "" &&
		v.Spotinst.ClusterIdentifier != ""
}

// endregion
