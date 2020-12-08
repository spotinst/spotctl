package cloud

import (
	"fmt"
	"log"
	"sync"
)

// All registered cloud providers.
var (
	cloudsMutex sync.RWMutex
	clouds      = make(map[ProviderName]Factory)
)

// Register registers a cloud provider factory by name.
//
// The cloud MUST have a name: lower case and one word.
func Register(name ProviderName, factory Factory) {
	cloudsMutex.Lock()
	defer cloudsMutex.Unlock()

	if name == "" {
		log.Fatalf("Cloud provider must have a name")
	}

	if _, dup := clouds[name]; dup {
		log.Fatalf("Cloud provider named %q already registered", name)
	}

	clouds[name] = factory
}

// GetFactory returns a factory of cloud provider by name.
func GetFactory(name ProviderName) (Factory, error) {
	cloudsMutex.RLock()
	defer cloudsMutex.RUnlock()

	if factory, ok := clouds[name]; ok {
		return factory, nil
	}

	return nil, fmt.Errorf("cloud: no factory function found for "+
		"cloud %q (missing import?)", name)
}

// GetInstance returns an instance of cloud provider by name.
func GetInstance(name ProviderName, options ...ProviderOption) (Provider, error) {
	factory, err := GetFactory(name)
	if err != nil {
		return nil, err
	}

	return factory(options...)
}
