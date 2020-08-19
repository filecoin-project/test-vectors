package chaos

import (
	"fmt"
	"io"
)

type State struct {
	// Unmarshallable is a sentinel value. If the slice contains no values, the
	// State struct will encode as CBOR without issue. If the slice is non-nil,
	// CBOR encoding will fail.
	Unmarshallable []*UnmarshallableCBOR
}

type UnmarshallableCBOR struct{}

func (t *UnmarshallableCBOR) UnmarshalCBOR(io.Reader) error {
	return fmt.Errorf("failed to unmarshal cbor")
}

func (t *UnmarshallableCBOR) MarshalCBOR(w io.Writer) error {
	return fmt.Errorf("failed to marshal cbor")
}
