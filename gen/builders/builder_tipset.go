package builders

import (
	"context"

	"github.com/filecoin-project/test-vectors/schema"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/conformance"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/ipfs/go-cid"
)

// TipsetVectorBuilder builds a tipset-class vector. It follows the same staged
// approach as MessageVectorBuilder.
//
// During the precondition stage, the user sets up the pre-existing state of
// the system, including the miners that are going to be producing the blocks
// comprising the vector. If the initial epoch is different to 0, the user must
// also set it during the precondition stage via SetInitialEpochOffset.
//
// During the application stage, the user stages the messages they later want to
// incorporate to a block in the StagedMessages object. Messages applied have no
// effect on the vector itself. The entire state is scrapped.
//
// Tipsets are registered in the Tipsets object, by calling Tipset#Next. This
// "opens" a new tipset at the next epoch, starting at the initial epoch set by
// SetInitialEpochOffset, and internally incrementing the counter by one.
//
// To register blocks on a tipset, call Tipset#Block, supplying the miner, win
// count, and the messages to enroll. Messages need to have been staged
// previously.
type TipsetVectorBuilder struct {
	*BuilderCommon

	// StagedMessages is a staging area for messages.
	StagedMessages *Messages
	// StateTracker is used for staging messages, and it's scrapped and replaced
	// by a fork at the PreRoot when committing applies.
	StateTracker *StateTracker

	InitialEpochOffset abi.ChainEpoch

	Tipsets *TipsetSeq
	Rewards *Rewards

	PreRoot  cid.Cid
	PostRoot cid.Cid
	vector   schema.TestVector
}

var _ Builder = (*TipsetVectorBuilder)(nil)

// TipsetVector creates a new TipsetVectorBuilder. For usage details, read the
// godocs on that type.
func TipsetVector(metadata *schema.Metadata, selector schema.Selector, mode Mode, hints []string, pv ProtocolVersion) *TipsetVectorBuilder {
	bc := &BuilderCommon{
		Stage:           StagePreconditions,
		ProtocolVersion: pv,
	}
	bc.Wallet = NewWallet()

	b := &TipsetVectorBuilder{
		BuilderCommon: bc,
	}

	b.StateTracker = NewStateTracker(bc, selector, &b.vector, pv.StateTree, pv.Actors, pv.ZeroStateTree)
	bc.Actors = NewActors(bc, b.StateTracker)

	b.vector.Class = schema.ClassTipset
	b.vector.Meta = metadata
	b.vector.Pre = &schema.Preconditions{}
	b.vector.Selector = selector
	b.vector.Hints = hints

	bc.Assert = NewAsserter(metadata.ID, pv, mode == ModeLenientAssertions, suppliers{
		messages: func() []*ApplicableMessage {
			// only return messages that have actually been enrolled on tipsets.
			return b.Tipsets.Messages()
		},
		stateTracker: func() *StateTracker {
			return b.StateTracker
		},
		actors:  func() *Actors { return bc.Actors },
		preroot: func() cid.Cid { return b.PreRoot },
	})

	bc.Assert.enterStage(StagePreconditions)

	return b
}

// SetInitialEpochOffset sets the initial epoch offset of this tipset-class
// vector. It MUST be called during the preconditions stage.
func (b *TipsetVectorBuilder) SetInitialEpochOffset(epoch abi.ChainEpoch) {
	if b.Stage != StagePreconditions {
		panic("you can only call SetInitialEpochOffset at preconditions stage")
	}
	b.InitialEpochOffset = epoch
}

// CommitPreconditions flushes the state tree, recording the new CID in the
// underlying test vector's precondition. It creates the StagedMessages and
// Tipsets object where messages will be staged, and tipsets will be registered.
//
// This method progesses the builder into the "applies" stage and may only be
// called during the "preconditions" stage.
func (b *TipsetVectorBuilder) CommitPreconditions() {
	if b.Stage != StagePreconditions {
		panic("called CommitPreconditions at the wrong time")
	}

	// capture the preroot after applying all preconditions.
	preroot := b.StateTracker.Flush()
	b.PreRoot = preroot

	// update the vector.
	b.vector.Pre.Variants = []schema.Variant{{
		ID:             b.ProtocolVersion.ID,
		Epoch:          int64(b.InitialEpochOffset + b.ProtocolVersion.FirstEpoch),
		NetworkVersion: uint(b.ProtocolVersion.Network),
	}}
	b.vector.Pre.StateTree = &schema.StateTree{RootCID: preroot}

	// initialize the Tipsets object.
	b.Tipsets = NewTipsetSeq(b.InitialEpochOffset)

	// update the internal state.
	// create a staging state tracker that will be used during applies.
	// create the message staging area, linked to the temporary state tracker.
	b.StagedMessages = NewMessages(b.BuilderCommon, b.StateTracker)

	b.Stage = StageApplies
	b.Assert.enterStage(StageApplies)
}

// CommitApplies applies all accumulated tipsets. It updates the vector after
// every tipset application, and records the Rewards state existing at that
// epoch under the Rewards object.
//
// It also sets the PostStateRoot on each applied Tipset, which can be used in
// combination with StateTracker#Fork and Asserter#AtState to load and assert
// on state at an interim tipset.
//
// This method progresses the builder into the "checks" stage and may only be
// called during the "applies" stage.
func (b *TipsetVectorBuilder) CommitApplies() {
	if b.Stage != StageApplies {
		panic("called CommitApplies at the wrong time")
	}

	// discard the temporary state, and fork at the preroot.
	b.StateTracker = b.StateTracker.Fork(b.PreRoot)

	var (
		ds = b.StateTracker.Stores.Datastore
		bs = b.StateTracker.Stores.Blockstore
	)

	// instantiate the reward tracker
	// record a rewards observation at the initial epoch.
	b.Rewards = NewRewards(b.BuilderCommon, b.StateTracker)
	b.Rewards.RecordAt(0)

	// Initialize Postconditions on the vector; set the preroot as the temporary
	// postcondition root.
	b.vector.Post = &schema.Postconditions{
		StateTree: &schema.StateTree{RootCID: b.PreRoot},
	}

	var traces []types.ExecutionTrace
	var prevEpoch = b.ProtocolVersion.FirstEpoch + b.InitialEpochOffset
	driver := conformance.NewDriver(context.Background(), b.vector.Selector, conformance.DriverOpts{})
	for _, ts := range b.Tipsets.All() {
		// Store the tipset in the vector.
		b.vector.ApplyTipsets = append(b.vector.ApplyTipsets, ts.Tipset)

		// Execute the tipset via the driver.
		root := b.vector.Post.StateTree.RootCID
		execEpoch := b.ProtocolVersion.FirstEpoch + b.InitialEpochOffset + abi.ChainEpoch(ts.EpochOffset)
		ret, err := driver.ExecuteTipset(bs, ds, root, prevEpoch, &ts.Tipset, execEpoch)
		b.Assert.NoError(err, "failed to apply tipset at epoch: %d", ts.EpochOffset)

		ts.PostStateRoot = ret.PostStateRoot

		for i, res := range ret.AppliedResults {
			// store the receipt in the vector.
			b.vector.Post.Receipts = append(b.vector.Post.Receipts, &schema.Receipt{
				ExitCode:    int64(res.ExitCode),
				ReturnValue: res.Return,
				GasUsed:     res.GasUsed,
			})

			// store the trace.
			traces = append(traces, res.ExecutionTrace)

			// store the result and basefee in the original message being
			// tracked by the TipsetSeq, so we can do asserts.
			// this is inefficient, but this is not production code.
			mcid := ret.AppliedMessages[i].Cid()
			for _, m := range b.Tipsets.Messages() {
				if m.Message.Cid() == mcid {
					m.baseFee = conformance.BaseFeeOrDefault(b.vector.Pre.BaseFee)
					m.Result = res
					break
				}
			}
		}

		// Update the state and receipts root in the vector.
		b.vector.Post.StateTree.RootCID = ret.PostStateRoot
		b.vector.Post.ReceiptsRoots = append(b.vector.Post.ReceiptsRoots, ret.ReceiptsRoot)
		prevEpoch = execEpoch

		// Update the state tree.
		b.PostRoot = b.vector.Post.StateTree.RootCID
		b.StateTracker.Load(b.PostRoot)

		// record a rewards observation.
		// TODO this is incomplete because it only records non-null rounds, but
		//  the rewards policy can change during null rounds. Unfortunately
		//  we don't get the intermediate roots through null rounds, so this
		//  is not straightforward.
		//  https://github.com/filecoin-project/test-vectors/issues/91
		b.Rewards.RecordAt(ts.EpochOffset)
	}

	// Update the vector diagnostics.
	b.vector.Diagnostics = EncodeTraces(traces)

	// Advance the stage to checks.
	b.Stage = StageChecks
	b.Assert.enterStage(StageChecks)
}

// Finish signals to the builder that the checks stage is complete and that the
// test vector can be finalized. It writes the test vector to the supplied
// io.Writer.
//
// This method progresses the builder into the "finished" stage and may only be
// called during the "checks" stage.
func (b *TipsetVectorBuilder) Finish() *schema.TestVector {
	if b.Stage != StageChecks {
		panic("called Finish at the wrong time")
	}

	car, err := EncodeCAR(b.StateTracker.Stores.DAGService, b.vector.Pre.StateTree.RootCID, b.vector.Post.StateTree.RootCID)
	if err != nil {
		panic(err)
	}
	b.vector.CAR = car

	b.Stage = StageFinished
	b.Assert = nil

	return &b.vector
}
