// Copyright 2021 NetApp, Inc. All Rights Reserved.

package config

import (
	"context"
	"fmt"
)

// Provider defines the interface for any component which will provide configuration.
// The Provider should not need to implement its own mutexes, because that will
// be managed by Config.
type Provider interface {
	fmt.Stringer

	// Retrieve retrieves and returns the configuration, or error in case of failure.
	Retrieve(ctx context.Context) (*Value, error)
}

// Value represents the operator configuration.
type Value struct {
	// ClusterIdentifier represents the cluster identifier that should be used by the Ocean Controller.
	ClusterIdentifier string `json:"clusterIdentifier" yaml:"clusterIdentifier"`
	// ACDIdentifier represents the ACDIdentifier identifier that should be used by the Ocean AKS Connector.
	ACDIdentifier string `json:"acdIdentifier" yaml:"acdIdentifier"`
}

// IsEmpty if all fields of a Value are empty.
func (v *Value) IsEmpty() bool { return v == nil || v.ClusterIdentifier == "" }

// IsComplete if all fields of a Value are set.
func (v *Value) IsComplete() bool { return v != nil && v.ClusterIdentifier != "" }

// Merge merges the passed in Value into the existing Value object.
func (v *Value) Merge(v2 *Value) *Value {
	if v != nil && v2 != nil {
		if v.ClusterIdentifier == "" {
			v.ClusterIdentifier = v2.ClusterIdentifier
		}
		if v.ACDIdentifier == "" {
			v.ACDIdentifier = v2.ACDIdentifier
		}
	}
	return v
}
