//go:generate go run generator.go

package box

import (
	"sort"
)

type Box interface{

	// Add a file content to box
	Add(file string, content []byte)

	// Get a file from box
	Get(file string) []byte

	// Has a file in box
	Has(file string) bool

	// List  files in box
	List() []string

}

type embedBox struct {
	storage map[string][]byte
}

// Embed box expose
var Boxed Box = newEmbedBox()


// Create new box for embed files
func newEmbedBox() *embedBox {
	return &embedBox{storage: make(map[string][]byte)}
}

// Add a file to box
func (e *embedBox) Add(file string, content []byte) {
	e.storage[file] = content
}

// Get file's content
// Always use / for looking up
// For example: /init/README.md is actually configs/init/README.md
func (e *embedBox) Get(file string) []byte {
	if f, ok := e.storage[file]; ok {
		return f
	}
	return nil
}

// Find for a file
func (e *embedBox) Has(file string) bool {
	if _, ok := e.storage[file]; ok {
		return true
	}
	return false
}

func (e *embedBox) List() []string {
	filenames := make([]string, len(e.storage))
	i:=0
	for key, _ := range e.storage {
		filenames[i] = key
		i++
	}
	sort.Strings(filenames)
	return filenames
}
