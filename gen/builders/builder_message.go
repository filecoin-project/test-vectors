package builders

import (
	"encoding/json"
	"io"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/test-vectors/schema"
)

// MessageVectorBuilder is a helper for building a message-class test vector.
// It enforces a staged process
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
type MessageVectorBuilder struct {
	*BuilderCommon

	StateTracker *StateTracker
	Messages     *Messages

	PreRoot  cid.Cid
	PostRoot cid.Cid
	vector   schema.TestVector
}

// MessageVector creates a builder for a message-class vector.
func MessageVector(metadata *schema.Metadata, selector schema.Selector, mode Mode, hints []string) *MessageVectorBuilder {
	bc := &BuilderCommon{Stage: StagePreconditions}

	st := NewStateTracker(bc, selector)
	bc.Actors = NewActors(bc, st)
	bc.Wallet = NewWallet()

	b := &MessageVectorBuilder{
		BuilderCommon: bc,
		StateTracker:  st,
		Messages:      NewMessages(bc, st),
	}

	b.vector.Class = schema.ClassMessage
	b.vector.Meta = metadata
	b.vector.Pre = &schema.Preconditions{}
	b.vector.Post = &schema.Postconditions{}
	b.vector.Selector = selector
	b.vector.Hints = hints

	bc.Assert = NewAsserter(metadata.ID, mode == ModeLenientAssertions, suppliers{
		messages:     b.Messages.All,
		stateTracker: func() *StateTracker { return st },
		actors:       func() *Actors { return bc.Actors },
		preroot:      func() cid.Cid { return b.PreRoot },
	})

	st.initializeZeroState(selector)

	bc.Assert.enterStage(StagePreconditions)

	return b
}

// CommitPreconditions flushes the state tree, recording the new CID in the
// underlying test vector's precondition.
//
// This method progesses the builder into the "applies" stage and may only be
// called during the "preconditions" stage.
func (b *MessageVectorBuilder) CommitPreconditions() {
	if b.Stage != StagePreconditions {
		panic("called CommitPreconditions at the wrong time")
	}

	// capture the preroot after applying all preconditions.
	preroot := b.StateTracker.Flush()

	b.vector.Pre.Epoch = 0
	b.vector.Pre.StateTree = &schema.StateTree{RootCID: preroot}

	b.PreRoot = preroot
	b.Stage = StageApplies
	b.Assert.enterStage(StageApplies)
}

// CommitApplies applies all accumulated messages. For each message it records
// the new state root, refreshes the state tree, and updates the underlying
// vector with the message and its receipt.
//
// This method progresses the builder into the "checks" stage and may only be
// called during the "applies" stage.
func (b *MessageVectorBuilder) CommitApplies() {
	if b.Stage != StageApplies {
		panic("called CommitApplies at the wrong time")
	}

	for _, am := range b.Messages.All() {
		// apply all messages that are pending application.
		if !am.Applied {
			b.StateTracker.ApplyMessage(am)
		}

		epoch := int64(am.Epoch)
		b.vector.ApplyMessages = append(b.vector.ApplyMessages, schema.Message{
			Bytes: MustSerialize(am.Message),
			Epoch: &epoch,
		})

		// am.Result may still be nil if the message failed to be applied
		if am.Result != nil {
			b.vector.Post.Receipts = append(b.vector.Post.Receipts, &schema.Receipt{
				ExitCode:    int64(am.Result.ExitCode),
				ReturnValue: am.Result.Return,
				GasUsed:     am.Result.GasUsed,
			})
		} else {
			b.vector.Post.Receipts = append(b.vector.Post.Receipts, nil)
		}
	}

	// update the internal state.
	b.PostRoot = b.StateTracker.CurrRoot
	b.vector.Post.StateTree = &schema.StateTree{RootCID: b.PostRoot}
	b.Stage = StageChecks
	b.Assert.enterStage(StageChecks)
}

// Finish signals to the builder that the checks stage is complete and that the
// test vector can be finalized. It creates a CAR from the recorded state root
// in it's pre and post condition and serializes the CAR into the test vector.
//
// This method progresses the builder into the "finished" stage and may only be
// called during the "checks" stage.
func (b *MessageVectorBuilder) Finish(w io.Writer) {
	if b.Stage != StageChecks {
		panic("called Finish at the wrong time")
	}

	car, err := EncodeCAR(b.StateTracker.Stores.DAGService, b.vector.Pre.StateTree.RootCID, b.vector.Post.StateTree.RootCID)
	if err != nil {
		panic(err)
	}
	b.vector.CAR = car

	msgs := b.Messages.All()
	traces := make([]types.ExecutionTrace, 0, len(msgs))
	for _, msgs := range msgs {
		traces = append(traces, msgs.Result.ExecutionTrace)
	}
	b.vector.Diagnostics = EncodeTraces(traces)

	b.Stage = StageFinished
	b.Assert = nil

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(b.vector); err != nil {
		panic(err)
	}
}
