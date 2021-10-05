// Copyright 2021 NetApp, Inc. All Rights Reserved.

package credentials

import (
	"context"
	"sync"
)

// Credentials provides synchronous safe retrieval of credentials.
//
// Credentials object is safe to use across multiple goroutines and will manage
// the synchronous state so the Providers do not need to implement their own
// synchronization.
//
// The first Credentials.Get() will always call Provider.Retrieve() to get the first
// instance of the credentials. All calls to Get() after that will return the
// cached credentials.
type Credentials struct {
	provider     Provider
	mu           sync.Mutex
	forceRefresh bool
	value        *Value
}

// NewCredentials returns a new Credentials with the provider set.
func NewCredentials(provider Provider) *Credentials {
	return &Credentials{
		provider:     provider,
		forceRefresh: true,
	}
}

// Get returns the credentials, or error if the credentials failed to be
// retrieved. Will return the cached credentials. If the credentials are
// empty the Provider's Retrieve() will be called to refresh the credentials.
func (x *Credentials) Get(ctx context.Context) (*Value, error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.value == nil || x.forceRefresh {
		value, err := x.provider.Retrieve(ctx)
		if err != nil {
			return nil, err
		}
		x.value = value
		x.forceRefresh = false
	}

	return x.value, nil
}

// Refresh refreshes the credentials and forces it to be retrieved on the next
// call to Get().
func (x *Credentials) Refresh() *Credentials {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.forceRefresh = true
	return x
}
