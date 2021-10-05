// Copyright 2021 NetApp, Inc. All Rights Reserved.

package credentials

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrNoValidProvidersFoundInChain Is returned when there are no valid credentials
// providers in the ChainProvider.
var ErrNoValidProvidersFoundInChain = errors.New("credentials: no valid " +
	"credentials providers in chain")

// ChainProvider will search for a provider which returns credentials and cache
// that provider until Retrieve is called again.
//
// The ChainProvider provides a way of chaining multiple providers together which
// will pick the first available using priority order of the Providers in the list.
//
// If none of the Providers retrieve valid credentials, Retrieve() will return
// the error ErrNoValidProvidersFoundInChain.
//
// If a Provider is found which returns valid credentials, ChainProvider will
// cache that Provider for all calls until Retrieve is called again.
type ChainProvider struct {
	Providers []Provider
}

// NewChainProvider returns a new ChainProvider.
func NewChainProvider(providers ...Provider) *ChainProvider {
	return &ChainProvider{
		Providers: providers,
	}
}

// Retrieve retrieves and returns the credentials, or error in case of failure.
func (x *ChainProvider) Retrieve(ctx context.Context) (*Value, error) {
	value := new(Value)
	var errs errorList

	if len(x.Providers) > 0 {
		for _, p := range x.Providers {
			v, err := p.Retrieve(ctx)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if value.Merge(v).IsComplete() {
				break
			}
		}
	}

	if value.IsEmpty() {
		err := ErrNoValidProvidersFoundInChain
		if len(errs) > 0 {
			err = errs
		}
		return nil, err
	}

	return value, nil
}

// String returns the string representation of the Chain provider.
func (x *ChainProvider) String() string {
	providers := make([]string, len(x.Providers))
	if len(x.Providers) > 0 {
		for i, provider := range x.Providers {
			providers[i] = provider.String()
		}
	}
	return fmt.Sprintf("ChainProvider(%s)", strings.Join(providers, ","))
}

// An error list that satisfies the error interface.
type errorList []error

// Error returns the string representation of the error list.
func (e errorList) Error() string {
	msg := ""
	if size := len(e); size > 0 {
		for i := 0; i < size; i++ {
			msg += fmt.Sprintf("%s", e[i].Error())

			// Check the next index to see if it is within the slice. If it is,
			// append a newline. We do this, because unit tests could be broken
			// with the additional '\n'.
			if i+1 < size {
				msg += "\n"
			}
		}
	}
	return msg
}
