package uuid

import (
	"github.com/satori/go.uuid"
)

// UUID is a thin wrapper around "github.com/satori/go.uuid".UUID.
type UUID struct {
	uuid.UUID
}

// String returns the canonical representation of the UUID.
func (u *UUID) String() string {
	if u == nil {
		return "<nil>"
	}

	return u.UUID.String()
}

// Short returns the first eight characters of the output of String().
func (u *UUID) Short() string {
	if u == nil {
		return u.String()
	}

	return u.String()[:8]
}

// Size returns the marshalled size of u, in bytes.
func (u *UUID) Size() int {
	return len(u.UUID)
}

// MarshalTo marshals u to data.
func (u *UUID) MarshalTo(data []byte) (int, error) {
	return copy(data, u.UUID.Bytes()), nil
}

// Unmarshal unmarshals data to u.
func (u *UUID) Unmarshal(data []byte) error {
	return u.UUID.UnmarshalBinary(data)
}

// NewV4 delegates to "github.com/satori/go.uuid".NewV4 and wraps the result in
// a UUID.
func NewV4() *UUID {
	return &UUID{uuid.NewV4()}
}

// FromBytes delegates to "github.com/satori/go.uuid".FromBytes and wraps the
// result in a UUID.
func FromBytes(input []byte) (*UUID, error) {
	u, err := uuid.FromBytes(input)
	return &UUID{u}, err
}

// FromString delegates to "github.com/satori/go.uuid".FromString and wraps the
// result in a UUID.
func FromString(input string) (*UUID, error) {
	u, err := uuid.FromString(input)
	return &UUID{u}, err
}
