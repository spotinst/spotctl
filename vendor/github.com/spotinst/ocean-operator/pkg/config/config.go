// Copyright 2021 NetApp, Inc. All Rights Reserved.

package config

import (
	"context"
	"sync"
)

// Config provides synchronous safe retrieval of configuration.
//
// Config object is safe to use across multiple goroutines and will manage
// the synchronous state so the Providers do not need to implement their own
// synchronization.
//
// The first Config.Get() will always call Provider.Retrieve() to get the first
// instance of the configuration. All calls to Get() after that will return the
// cached configuration.
type Config struct {
	provider     Provider
	mu           sync.Mutex
	forceRefresh bool
	value        *Value
}

// NewConfig returns a new Config with the provider set.
func NewConfig(provider Provider) *Config {
	return &Config{
		provider:     provider,
		forceRefresh: true,
	}
}

// Get returns the configuration, or error if the configuration failed to be
// retrieved. Will return the cached configuration. If the configuration is
// empty the Provider's Retrieve() will be called to refresh the configuration.
func (x *Config) Get(ctx context.Context) (*Value, error) {
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

// Refresh refreshes the configuration and forces it to be retrieved on the next
// call to Get().
func (x *Config) Refresh() *Config {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.forceRefresh = true
	return x
}
