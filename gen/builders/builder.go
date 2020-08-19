package builders

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-car"

	lotus "github.com/filecoin-project/lotus/conformance"
)

// Stage is an identifier for the current stage a Builder is in.
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

func init() {
	// disable logs, as we need a clean stdout output.
	log.SetOutput(os.Stderr)
	log.SetPrefix(">>> ")

	_ = os.Setenv("LOTUS_DISABLE_VM_BUF", "iknowitsabadidea")
}

// Builder is a helper for building a test vector. It enforces a staged process
// for constructing a vector, starting in a "preconditions" stage, where the
// state tree is manipulated into the desired shape for the test to begin. Next,
// the "applies" stage serves as a time to accumulate messages that will mutate
// the state. Transitioning to the "checks" stage applies the messages and
// transitions the builder to a period where any final assertions can be made on
// the receipts or the state of the system. Finally, transitioning to the
// "finished" stage finalizes the test vector, serializing the pre and post
// state trees and completing the stages. From here the test vector can be
// serialized into a JSON file.
//
// TODO use stage.Surgeon with non-proxying blockstore.
type Builder struct {
	Actors    *Actors
	Assert    *Asserter
	Messages  *Messages
	Traces    []types.ExecutionTrace
	Driver    *lotus.Driver
	PreRoot   cid.Cid
	PostRoot  cid.Cid
	CurrRoot  cid.Cid
	Wallet    *Wallet
	StateTree *state.StateTree
	Stores    *Stores

	vector TestVector
	stage  Stage
}

// MessageVector creates a builder for a message-class vector.
func MessageVector(metadata *Metadata) *Builder {
	stores := NewLocalStores(context.Background())

	// Create a brand new state tree.
	st, err := state.NewStateTree(stores.CBORStore)
	if err != nil {
		panic(err)
	}

	b := &Builder{
		stage:     StagePreconditions,
		Stores:    stores,
		StateTree: st,
		PreRoot:   cid.Undef,
		Driver:    lotus.NewDriver(context.Background()),
	}

	b.Wallet = newWallet()
	b.Assert = newAsserter(b, StagePreconditions)
	b.Actors = newActors(b)
	b.Messages = &Messages{b: b}

	b.vector.Class = ClassMessage
	b.vector.Meta = metadata
	b.vector.Pre = &Preconditions{}
	b.vector.Post = &Postconditions{}

	b.initializeZeroState()

	return b
}

// CommitPreconditions flushes the state tree, recording the new CID in the
// underlying test vector's precondition.
//
// This method progesses the builder into the "applies" stage and may only be
// called during the "preconditions" stage.
func (b *Builder) CommitPreconditions() {
	if b.stage != StagePreconditions {
		panic("called CommitPreconditions at the wrong time")
	}

	// capture the preroot after applying all preconditions.
	preroot := b.FlushState()

	b.vector.Pre.Epoch = 0
	b.vector.Pre.StateTree = &StateTree{RootCID: preroot}

	b.CurrRoot, b.PreRoot = preroot, preroot
	b.stage = StageApplies
	b.Assert = newAsserter(b, StageApplies)
}

// CommitApplies applies all accumulated messages. For each message it records
// the new state root, refreshes the state tree, and updates the underlying
// vector with the message and its receipt.
//
// This method progresses the builder into the "checks" stage and may only be
// called during the "applies" stage.
func (b *Builder) CommitApplies() {
	if b.stage != StageApplies {
		panic("called CommitApplies at the wrong time")
	}

	for _, am := range b.Messages.All() {
		// apply all messages that are pending application.
		if am.Result == nil {
			b.applyMessage(am)
		}
	}

	b.PostRoot = b.CurrRoot
	b.vector.Post.StateTree = &StateTree{RootCID: b.CurrRoot}
	b.stage = StageChecks
	b.Assert = newAsserter(b, StageChecks)
}

// applyMessage executes the provided message via the driver, records the new
// root, refreshes the state tree, and updates the underlying vector with the
// message and its receipt.
func (b *Builder) applyMessage(am *ApplicableMessage) {
	var err error
	am.Result, b.CurrRoot, err = b.Driver.ExecuteMessage(am.Message, b.CurrRoot, b.Stores.Blockstore, am.Epoch)
	b.Assert.NoError(err)

	// replace the state tree.
	b.StateTree, err = state.LoadStateTree(b.Stores.CBORStore, b.CurrRoot)
	b.Assert.NoError(err)

	b.vector.ApplyMessages = append(b.vector.ApplyMessages, Message{
		Bytes: MustSerialize(am.Message),
		Epoch: &am.Epoch,
	})
	b.vector.Post.Receipts = append(b.vector.Post.Receipts, &Receipt{
		ExitCode:    am.Result.ExitCode,
		ReturnValue: am.Result.Return,
		GasUsed:     am.Result.GasUsed,
	})
	b.Traces = append(b.Traces, am.Result.ExecutionTrace)
}

// Finish signals to the builder that the checks stage is complete and that the
// test vector can be finialized. It creates a CAR from the recorded state root
// in it's pre and post condition and serializes the CAR into the test vector.
//
// This method progresses the builder into the "finished" stage and may only be
// called during the "checks" stage.
func (b *Builder) Finish(w io.Writer) {
	if b.stage != StageChecks {
		panic("called Finish at the wrong time")
	}

	out := new(bytes.Buffer)
	gw := gzip.NewWriter(out)
	if err := b.WriteCAR(gw, b.vector.Pre.StateTree.RootCID, b.vector.Post.StateTree.RootCID); err != nil {
		panic(err)
	}
	if err := gw.Flush(); err != nil {
		panic(err)
	}
	if err := gw.Close(); err != nil {
		panic(err)
	}

	b.vector.CAR = out.Bytes()
	b.vector.Diagnostics = EncodeTraces(b.Traces)

	b.stage = StageFinished
	b.Assert = nil

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(b.vector); err != nil {
		panic(err)
	}
}

// WriteCAR recursively writes the tree referenced by the root as assert CAR
// into the supplied io.Writer.
//
// TODO use state.Surgeon instead. (This is assert copy of Surgeon#WriteCAR).
func (b *Builder) WriteCAR(w io.Writer, roots ...cid.Cid) error {
	carWalkFn := func(nd format.Node) (out []*format.Link, err error) {
		for _, link := range nd.Links() {
			if link.Cid.Prefix().Codec == cid.FilCommitmentSealed || link.Cid.Prefix().Codec == cid.FilCommitmentUnsealed {
				continue
			}
			out = append(out, link)
		}
		return out, nil
	}

	return car.WriteCarWithWalker(context.Background(), b.Stores.DAGService, roots, w, carWalkFn)
}

// FlushState calls Flush on the builder's StateTree.
func (b *Builder) FlushState() cid.Cid {
	preroot, err := b.StateTree.Flush(context.Background())
	if err != nil {
		panic(err)
	}
	return preroot
}
