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
	tipsets []*Tipset
	epoch   abi.ChainEpoch

	// msgIdx is an index that stores unique messages enlisted in blocks in this
	// tipset sequence.
	msgIdx map[cid.Cid]*ApplicableMessage
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
		epoch:  initialEpoch,
		msgIdx: make(map[cid.Cid]*ApplicableMessage),
	}
}

// All returns all tipsets that have been registered.
func (tss *TipsetSeq) All() []*Tipset {
	return tss.tipsets
}

// Messages returns all ApplicableMessages that have been included in blocks,
// in no particular order.
func (tss *TipsetSeq) Messages() []*ApplicableMessage {
	msgs := make([]*ApplicableMessage, 0, len(tss.msgIdx))
	for _, msg := range tss.msgIdx {
		msgs = append(msgs, msg)
	}
	return msgs
}

// Next enrols a new Tipset, with the supplied base fee, and advances the epoch
// by 1.
func (tss *TipsetSeq) Next(baseFee abi.TokenAmount) *Tipset {
	ts := &Tipset{
		tss: tss,
		Tipset: schema.Tipset{
			Epoch:   tss.epoch,
			BaseFee: baseFee,
		},
	}
	tss.tipsets = append(tss.tipsets, ts)
	tss.epoch++ // advance the epoch.
	return ts
}

// NullRounds enrols as many null rounds as indicated, advancing the epoch by
// the count.
func (tss *TipsetSeq) NullRounds(count uint64) {
	tss.epoch += abi.ChainEpoch(count)
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
		ts.tss.msgIdx[am.Message.Cid()] = am
	}

	ts.Blocks = append(ts.Blocks, block)
}
