package builders

import (
	"github.com/filecoin-project/go-state-types/abi"
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

	vector schema.TestVector
}

// MessageVector creates a builder for a message-class vector.
func MessageVector(metadata *schema.Metadata, selector schema.Selector, mode Mode, hints []string, pv ProtocolVersion) *MessageVectorBuilder {
	bc := &BuilderCommon{
		Stage:           StagePreconditions,
		ProtocolVersion: pv,
	}
	bc.Wallet = NewWallet()

	b := &MessageVectorBuilder{
		BuilderCommon: bc,
	}

	b.StateTracker = NewStateTracker(bc, selector, &b.vector, pv.StateTree, pv.Actors, pv.ZeroStateTree)
	b.Messages = NewMessages(bc, b.StateTracker)
	bc.Actors = NewActors(bc, b.StateTracker)

	b.vector.Class = schema.ClassMessage
	b.vector.Meta = metadata
	b.vector.Pre = &schema.Preconditions{}
	b.vector.Post = &schema.Postconditions{}
	b.vector.Selector = selector
	b.vector.Hints = hints

	bc.Assert = NewAsserter(metadata.ID, pv, mode == ModeLenientAssertions, suppliers{
		messages:     b.Messages.All,
		stateTracker: func() *StateTracker { return b.StateTracker },
		actors:       func() *Actors { return bc.Actors },
		preroot:      func() cid.Cid { return b.PreRoot },
	})

	bc.Assert.enterStage(StagePreconditions)

	return b
}

// SetCirculatingSupply sets the circulating supply for this vector. If not set,
// the driver should use the total maximum supply of Filecoin as specified in
// the protocol when executing these messages.
func (b *MessageVectorBuilder) SetCirculatingSupply(supply abi.TokenAmount) {
	b.vector.Pre.CircSupply = supply.Int
}

// SetBaseFee sets the base fee for this vector. If not set, the driver should
// use 100 attoFIL as the base fee when executing this vector.
func (b *MessageVectorBuilder) SetBaseFee(basefee abi.TokenAmount) {
	b.vector.Pre.BaseFee = basefee.Int
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
	b.vector.Pre.StateTree = &schema.StateTree{RootCID: preroot}
	b.PreRoot = preroot

	b.vector.Pre.Variants = []schema.Variant{{
		ID:             b.ProtocolVersion.ID,
		Epoch:          int64(b.ProtocolVersion.FirstEpoch),
		NetworkVersion: uint(b.ProtocolVersion.Network),
	}}

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

	for i, am := range b.Messages.All() {
		// apply all messages that are pending application.
		if !am.Applied {
			b.StateTracker.ApplyMessage(am)
		}

		epoch := int64(am.EpochOffset)
		b.vector.ApplyMessages = append(b.vector.ApplyMessages, schema.Message{
			Bytes:       MustSerialize(am.Message),
			EpochOffset: &epoch,
		})

		if am.Failed {
			b.vector.Post.ApplyMessageFailures = append(b.vector.Post.ApplyMessageFailures, i)
		}

		var receipt *schema.Receipt
		if !am.Failed {
			receipt = &schema.Receipt{
				ExitCode:    int64(am.Result.ExitCode),
				ReturnValue: am.Result.Return,
				GasUsed:     am.Result.GasUsed,
			}
		}
		b.vector.Post.Receipts = append(b.vector.Post.Receipts, receipt)
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
func (b *MessageVectorBuilder) Finish() *schema.TestVector {
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
	for _, msg := range msgs {
		if !msg.Failed {
			traces = append(traces, msg.Result.ExecutionTrace)
		}
	}
	b.vector.Diagnostics = EncodeTraces(traces)

	b.Stage = StageFinished
	b.Assert = nil

	return &b.vector
}
