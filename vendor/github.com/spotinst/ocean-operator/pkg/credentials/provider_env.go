// Copyright 2021 NetApp, Inc. All Rights Reserved.

package credentials

import (
	"context"
	"fmt"
	"os"

	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
)

const (
	// EnvCredentialsToken specifies the name of the environment variable points
	// to the Spotinst Token.
	EnvCredentialsToken = credentials.EnvCredentialsVarToken

	// EnvCredentialsAccount specifies the name of the environment variable points
	// to the Spotinst account ID.
	EnvCredentialsAccount = credentials.EnvCredentialsVarAccount
)

// ErrEnvCredentialsNotFound is returned when no credentials can be found in the
// process's environment.
var ErrEnvCredentialsNotFound = fmt.Errorf("credentials: %s and %s not found "+
	"in environment", EnvCredentialsToken, EnvCredentialsAccount)

// EnvProvider retrieves credentials from the environment variables of the process.
type EnvProvider struct{}

// NewEnvProvider returns a new EnvProvider.
func NewEnvProvider() *EnvProvider {
	return new(EnvProvider)
}

// Retrieve retrieves and returns the credentials, or error in case of failure.
func (x *EnvProvider) Retrieve(ctx context.Context) (*Value, error) {
	value := &Value{
		Token:   os.Getenv(EnvCredentialsToken),
		Account: os.Getenv(EnvCredentialsAccount),
	}

	if value.IsEmpty() {
		return value, ErrEnvCredentialsNotFound
	}

	return value, nil
}

// String returns the string representation of the Env provider.
func (x *EnvProvider) String() string {
	return "EnvProvider"
}
