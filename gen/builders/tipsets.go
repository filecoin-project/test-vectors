package builders

import (
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/test-vectors/schema"

	"github.com/filecoin-project/go-state-types/abi"
)

// TipsetSeq is a sequence of tipsets to be applied during the test.
// TipsetSeq provides methods to build a sequence of tipsets, such as
// adding new tipsets and null rounds.
type TipsetSeq struct {
	tipsets     []*Tipset
	epochOffset abi.ChainEpoch

	// msgIdx is an index that stores unique messages enlisted in blocks in this
	// tipset sequence.
	msgIdx map[cid.Cid]*ApplicableMessage

	// orderedMsgs keeps track of the order of messages as they appear when new blocks
	// are added to a tipset.
	orderedMsgs []*ApplicableMessage
}

type Tipset struct {
	tss *TipsetSeq

	// PostStateRoot stores the state root CID after applying this tipset.
	// It can be used with Asserter#AtState to obtain an asserter against that
	// state root.
	PostStateRoot cid.Cid

	schema.Tipset
}

type Block = schema.Block

// NewTipsetSeq returns a new TipSetSeq object initialized at the provided
// epoch.
func NewTipsetSeq(initialEpoch abi.ChainEpoch) *TipsetSeq {
	return &TipsetSeq{
		epochOffset: initialEpoch,
		msgIdx:      make(map[cid.Cid]*ApplicableMessage),
	}
}

// All returns all tipsets that have been registered.
func (tss *TipsetSeq) All() []*Tipset {
	return tss.tipsets
}

// Messages returns all ApplicableMessages that have been included in blocks,
// ordered based on inclusion in the tipset.
func (tss *TipsetSeq) Messages() []*ApplicableMessage {
	msgs := make([]*ApplicableMessage, len(tss.orderedMsgs))
	copy(msgs, tss.orderedMsgs)
	return msgs
}

// Next enrols a new Tipset, with the supplied base fee, and advances the epoch
// by 1.
func (tss *TipsetSeq) Next(baseFee abi.TokenAmount) *Tipset {
	ts := &Tipset{
		tss: tss,
		Tipset: schema.Tipset{
			EpochOffset: int64(tss.epochOffset),
			BaseFee:     *baseFee.Int,
		},
	}
	tss.tipsets = append(tss.tipsets, ts)
	tss.epochOffset++ // advance the epoch.
	return ts
}

// NullRounds enrols as many null rounds as indicated, advancing the epoch by
// the count.
func (tss *TipsetSeq) NullRounds(count uint64) {
	tss.epochOffset += abi.ChainEpoch(count)
}

// Block adds a new block to this tipset, produced by the indicated miner, with
// the supplied wincount, and containing the listed (and previously staged)
// msgIdx.
func (ts *Tipset) Block(miner Miner, winCount int64, msgs ...*ApplicableMessage) {
	block := Block{
		MinerAddr: miner.MinerActorAddr.ID,
		WinCount:  winCount,
	}
	for _, am := range msgs {
		block.Messages = append(block.Messages, MustSerialize(am.Message))

		// if we see this message for the first time, add it to the `msgIdx` map and to the `orderMsgs` slice.
		if _, ok := ts.tss.msgIdx[am.Message.Cid()]; !ok {
			ts.tss.msgIdx[am.Message.Cid()] = am
			ts.tss.orderedMsgs = append(ts.tss.orderedMsgs, am)
		}
	}

	ts.Blocks = append(ts.Blocks, block)
}
