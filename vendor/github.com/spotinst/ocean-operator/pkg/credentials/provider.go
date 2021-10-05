// Copyright 2021 NetApp, Inc. All Rights Reserved.

package credentials

import (
	"context"
	"fmt"
)

// Provider defines the interface for any component which will provide credentials.
// The Provider should not need to implement its own mutexes, because that will
// be managed by config.
type Provider interface {
	fmt.Stringer

	// Retrieve retrieves and returns the credentials, or error in case of failure.
	Retrieve(ctx context.Context) (*Value, error)
}

// Value represents the operator credentials.
type Value struct {
	// Token represents the token that should be used by the Ocean Controller.
	Token string `json:"token" yaml:"token"`
	// Account represents the account that should be used by the Ocean Controller.
	Account string `json:"account" yaml:"account"`
}

// IsEmpty if all fields of a Value are empty.
func (v *Value) IsEmpty() bool { return v == nil || (v.Token == "" && v.Account == "") }

// IsComplete if all fields of a Value are set.
func (v *Value) IsComplete() bool { return v != nil && v.Token != "" && v.Account != "" }

// Merge merges the passed in Value into the existing Value object.
func (v *Value) Merge(v2 *Value) *Value {
	if v != nil && v2 != nil {
		if v.Token == "" {
			v.Token = v2.Token
		}
		if v.Account == "" {
			v.Account = v2.Account
		}
	}
	return v
}
