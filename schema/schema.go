package schema

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/ipfs/go-cid"
)

// Class represents the type of test vector this instance is.
type Class string

const (
	// ClassMessage tests the VM behaviour and resulting state over one or
	// many messages.
	ClassMessage Class = "message"
	// ClassTipset tests the VM behaviour and resulting state over one or many
	// tipsets and/or null rounds.
	ClassTipset Class = "tipset"
	// ClassBlockSeq tests the state of the system after the arrival of
	// particular blocks at concrete points in time.
	ClassBlockSeq Class = "blockseq"
)

const (
	// HintIncorrect is a standard hint to convey that a vector is knowingly
	// incorrect. Drivers may choose to skip over these vectors, or if it's
	// accompanied by HintNegate, they may perform the assertions as explained
	// in its godoc.
	HintIncorrect = "incorrect"

	// HintNegate is a standard hint to convey to drivers that, if this vector
	// is run, they should negate the postcondition checks (i.e. check that the
	// postcondition state is expressly NOT the one encoded in this vector).
	HintNegate = "negate"
)

// Selector is a predicate the driver can use to determine if this test vector
// is relevant given the capabilities/features of the underlying implementation
// and/or test environment.
type Selector map[string]string

// Metadata provides information on the generation of this test case
type Metadata struct {
	ID      string           `json:"id"`
	Version string           `json:"version,omitempty"`
	Desc    string           `json:"description,omitempty"`
	Comment string           `json:"comment,omitempty"`
	Gen     []GenerationData `json:"gen"`
	Tags    []string         `json:"tags,omitempty"`
}

// GenerationData tags the source of this test case.
type GenerationData struct {
	Source  string `json:"source,omitempty"`
	Version string `json:"version,omitempty"`
}

// StateTree represents a state tree within preconditions and postconditions.
type StateTree struct {
	RootCID cid.Cid `json:"root_cid"`
}

// Base64EncodedBytes is a base64-encoded binary value.
type Base64EncodedBytes []byte

// Preconditions contain a representation of VM state at the beginning of the test
type Preconditions struct {
	Epoch     abi.ChainEpoch `json:"epoch"`
	StateTree *StateTree     `json:"state_tree"`
}

// Receipt represents a receipt to match against.
type Receipt struct {
	ExitCode    exitcode.ExitCode  `json:"exit_code"`
	ReturnValue Base64EncodedBytes `json:"return"`
	GasUsed     int64              `json:"gas_used"`
}

// Postconditions contain a representation of VM state at th end of the test
type Postconditions struct {
	StateTree     *StateTree `json:"state_tree"`
	Receipts      []*Receipt `json:"receipts"`
	ReceiptsRoots []cid.Cid  `json:"receipts_roots,omitempty"`
}

// MarshalJSON implements json.Marshal for Base64EncodedBytes
func (beb Base64EncodedBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(beb))
}

// UnmarshalJSON implements json.Unmarshal for Base64EncodedBytes
func (beb *Base64EncodedBytes) UnmarshalJSON(v []byte) error {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}

	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	*beb = bytes
	return nil
}

// Diagnostics contain a representation of VM diagnostics
type Diagnostics struct {
	Format string             `json:"format"`
	Data   Base64EncodedBytes `json:"data"`
}

// TestVector is a single test case
type TestVector struct {
	Class    `json:"class"`
	Selector `json:"selector,omitempty"`

	// Hints are arbitrary flags that convey information to the driver.
	// Use hints to express facts like this vector is knowingly incorrect
	// (e.g. when the reference implementation is broken), or that drivers
	// should negate the postconditions (i.e. test that they are NOT the ones
	// expressed in the vector), etc.
	//
	// Refer to the Hint* constants for common hints.
	Hints []string `json:"hints,omitempty"`

	Meta *Metadata `json:"_meta"`

	// CAR binary data to be loaded into the test environment, usually a CAR
	// containing multiple state trees, addressed by root CID from the relevant
	// objects.
	CAR Base64EncodedBytes `json:"car"`

	Pre *Preconditions `json:"preconditions"`

	ApplyMessages []Message `json:"apply_messages,omitempty"`
	ApplyTipsets  []Tipset  `json:"apply_tipsets,omitempty"`

	Post        *Postconditions `json:"postconditions"`
	Diagnostics *Diagnostics    `json:"diagnostics,omitempty"`
}

type Message struct {
	Bytes Base64EncodedBytes `json:"bytes"`
	Epoch *abi.ChainEpoch    `json:"epoch,omitempty"`
}

type Tipset struct {
	Epoch   abi.ChainEpoch  `json:"epoch"`
	BaseFee abi.TokenAmount `json:"basefee"`
	Blocks  []Block         `json:"blocks,omitempty"`
}

type Block struct {
	MinerAddr address.Address      `json:"miner_addr"`
	WinCount  int64                `json:"win_count"`
	Messages  []Base64EncodedBytes `json:"messages"`
}

// Validate validates this test vector against the JSON schema, and applies
// further validation rules that cannot be enforced through JSON Schema.
func (tv TestVector) Validate() error {
	if tv.Class == ClassMessage {
		if len(tv.Post.Receipts) != len(tv.ApplyMessages) {
			return fmt.Errorf("length of postcondition receipts must match length of messages to apply")
		}
	}
	if tv.Class == ClassTipset {

	}
	return nil
}

// MustMarshalJSON encodes the test vector to JSON and panics if it errors.
func (tv TestVector) MustMarshalJSON() []byte {
	b, err := json.Marshal(&tv)
	if err != nil {
		panic(err)
	}
	return b
}
