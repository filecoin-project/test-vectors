package builders

import (
	"log"
	"os"

	"github.com/filecoin-project/lotus/chain/vm"

	"github.com/filecoin-project/test-vectors/schema"
)

func init() {
	// disable logs, as we need a clean stdout output.
	log.SetOutput(os.Stderr)
	log.SetPrefix(">>> ")

	_ = os.Setenv("LOTUS_DISABLE_VM_BUF", "iknowitsabadidea")

	// enable gas tracing in execution traces.
	vm.EnableGasTracing = true
}

// Builder is a vector builder.
//
// It enforces a staged process for constructing a vector, starting in a
// "preconditions" stage, where the state tree is manipulated into the desired
// shape for the test to begin.
//
// Next, the "applies" stage serves as a time to accumulate messages that will
// mutate the state.
//
// Transitioning to the "checks" stage applies the messages and transitions the
// builder to a period where any final assertions can be made on
// the receipts or the state of the system.
//
// Finally, transitioning to the "finished" stage finalizes the test vector,
// serializing the pre and post state trees and completing the stages.
//
// From here the test vector can be serialized into a JSON file.
type Builder interface {
	// CommitPreconditions transitions the vector from the preconditions stage
	// to the applies stage.
	CommitPreconditions()

	// CommitApplies transitions the vector from the applies stage to the
	// checks stage.
	CommitApplies()

	// Finish closes this test vector (transioning it to the terminal finished
	// stage), and returns it.
	Finish() *schema.TestVector
}

// BuilderCommon bundles common services and state fields that are available in
// all vector builder types.
type BuilderCommon struct {
	Stage  Stage
	Actors *Actors
	Assert *Asserter
	Wallet *Wallet

	// ProtocolVersion this vector is being built against.
	ProtocolVersion ProtocolVersion
}

// Stage is an identifier for the current stage a MessageVectorBuilder is in.
type Stage string

const (
	// StagePreconditions is where the state tree is manipulated into the desired shape.
	StagePreconditions = Stage("preconditions")
	// StageApplies is where messages are accumulated.
	StageApplies = Stage("applies")
	// StageChecks is where assertions are made on receipts and state.
	StageChecks = Stage("checks")
	// StageFinished is where the test vector is finalized, ready for serialization.
	StageFinished = Stage("finished")
)

// Mode tunes certain elements of how the generation and assertion of
// a test vector will be conducted, such as being lenient to assertion
// failures when a vector is knowingly incorrect. Refer to the Mode* constants
// for further information.
type Mode int

const (
	// ModeStandard is the implicit mode (0). In this mode, assertion failures
	// cause the vector generation to abort.
	ModeStandard Mode = iota

	// ModeLenientAssertions allows generation to proceed even if assertions
	// fail. Use it when you know that a test vector is broken in the reference
	// implementation, but want to generate it anyway.
	//
	// Consider using Hints to convey to drivers how they should treat this
	// test vector.
	ModeLenientAssertions
)
