package thirdparty

import (
	"fmt"
	"log"
	"sync"
)

// All registered commands.
var (
	commandsMutex sync.RWMutex
	commands      = make(map[CommandName]Factory)
)

// Register registers a command factory by name.
//
// The command MUST have a name: lower case and one word.
func Register(name CommandName, factory Factory) {
	commandsMutex.Lock()
	defer commandsMutex.Unlock()

	if name == "" {
		log.Fatalf("Command must have a name")
	}

	if _, dup := commands[name]; dup {
		log.Fatalf("Command named %q already registered", name)
	}

	commands[name] = factory
}

// GetFactory returns a factory of command by name.
func GetFactory(name CommandName) (Factory, error) {
	commandsMutex.RLock()
	defer commandsMutex.RUnlock()

	if factory, ok := commands[name]; ok {
		return factory, nil
	}

	return nil, fmt.Errorf("thirdparty: no factory function found for "+
		"command %q (missing import?)", name)
}

// GetInstance returns an instance of command by name.
func GetInstance(name CommandName, options ...CommandOption) (Command, error) {
	factory, err := GetFactory(name)
	if err != nil {
		return nil, err
	}

	opts := initDefaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	return factory(opts)
}
