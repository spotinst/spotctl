package writer

import (
	"fmt"
	"io"
	"log"
	"sync"
)

// All registered writers.
var (
	writersMutex sync.RWMutex
	writers      = make(map[Format]Factory)
)

// Register registers a writer factory by format.
//
// The writer MUST have a named format: lower case and one word.
func Register(format Format, factory Factory) {
	writersMutex.Lock()
	defer writersMutex.Unlock()

	if format == "" {
		log.Fatalf("Writer must have a name")
	}

	if _, dup := writers[format]; dup {
		log.Fatalf("Writer named %q already registered", format)
	}

	writers[format] = factory
}

// GetFactory returns a factory of writer by format.
func GetFactory(format Format) (Factory, error) {
	writersMutex.RLock()
	defer writersMutex.RUnlock()

	if factory, ok := writers[format]; ok {
		return factory, nil
	}

	return nil, fmt.Errorf("writer: no factory function found for "+
		"writer %q (missing import?)", format)
}

// GetInstance returns an instance of writer by format.
func GetInstance(format Format, w io.Writer) (Writer, error) {
	factory, err := GetFactory(format)
	if err != nil {
		return nil, err
	}

	return factory(w)
}
