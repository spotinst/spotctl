// Copyright 2021 NetApp, Inc. All Rights Reserved.

package installer

import (
	"fmt"
	"sync"
)

// Factory is a function that returns an Installer interface. An error is
// returned if the installer fails to initialize, nil otherwise.
type Factory func(options *InstallerOptions) (Installer, error)

// All registered installers.
var (
	installersMutex sync.RWMutex
	installers      = make(map[string]Factory)
)

// MustRegister registers a Factory by name and panics if an error occurs.
func MustRegister(name string, factory Factory) {
	if err := Register(name, factory); err != nil {
		panic(err)
	}
}

// Register registers a Factory by name and returns an error, if any.
func Register(name string, factory Factory) error {
	installersMutex.Lock()
	defer installersMutex.Unlock()

	if name == "" {
		return fmt.Errorf("installer must have a name")
	}

	if _, dup := installers[name]; dup {
		return fmt.Errorf("installer named %q already registered", name)
	}

	installers[name] = factory
	return nil
}

// GetFactory returns a Factory by name.
func GetFactory(name string) (Factory, error) {
	installersMutex.RLock()
	defer installersMutex.RUnlock()

	if factory, ok := installers[name]; ok {
		return factory, nil
	}

	return nil, fmt.Errorf("installer: no factory function found for "+
		"installer %q (missing import?)", name)
}

// GetInstance returns an instance of installer by name.
func GetInstance(name string, options ...InstallerOption) (Installer, error) {
	factory, err := GetFactory(name)
	if err != nil {
		return nil, err
	}

	opts := mutateInstallerOptions(options...)
	return factory(opts)
}
