package builders

import (
	"github.com/filecoin-project/lotus/chain/actors/builtin/reward"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/specs-actors/actors/builtin"
)

// RewardSummary holds the state we care about insofar rewards are concerned, at
// every epoch where we perform an observation.
type RewardSummary struct {
	Treasury           abi.TokenAmount
	EpochReward        abi.TokenAmount
	NextPerBlockReward abi.TokenAmount
}

// Rewards tracks RewardSummary objects throughout chain history.
type Rewards struct {
	st *StateTracker
	bc *BuilderCommon
	m  map[abi.ChainEpoch]*RewardSummary
}

// NewRewards creates a new Rewards object.
func NewRewards(bc *BuilderCommon, st *StateTracker) *Rewards {
	return &Rewards{
		st: st,
		bc: bc,
		m:  make(map[abi.ChainEpoch]*RewardSummary),
	}
}

// RecordAt records a RewardSummary associated with the supplied epoch, by
// accessing the latest version of the statetree.
func (r *Rewards) RecordAt(epochOffset int64) {
	actor, err := r.st.StateTree.GetActor(builtin.RewardActorAddr)
	r.bc.Assert.NoError(err)

	state, err := reward.Load(r.st.Stores.ADTStore, actor)
	r.bc.Assert.NoError(err, "failed to load state for reward actor; head=%s", actor.Head)

	rew, err := state.ThisEpochReward()
	r.bc.Assert.NoError(err, "failed to load this epoch reward")

	rs := &RewardSummary{
		Treasury:           actor.Balance,
		EpochReward:        rew,
		NextPerBlockReward: big.Div(rew, big.NewInt(builtin.ExpectedLeadersPerEpoch)),
	}

	r.m[abi.ChainEpoch(epochOffset)] = rs
}

// ForEpochOffset returns the RewardSummary (or nil) associated with the given
// epoch offset.
func (r *Rewards) ForEpochOffset(epochOffset int64) *RewardSummary {
	return r.m[abi.ChainEpoch(epochOffset)]
}
