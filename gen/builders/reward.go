package builders

import (
	"context"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/reward"
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
func (r *Rewards) RecordAt(epoch abi.ChainEpoch) {
	actor, err := r.st.StateTree.GetActor(builtin.RewardActorAddr)
	r.bc.Assert.NoError(err)

	var state reward.State
	err = r.st.Stores.CBORStore.Get(context.Background(), actor.Head, &state)
	r.bc.Assert.NoError(err, "failed to load state for reward actor; head=%s", actor.Head)

	rs := &RewardSummary{
		Treasury:           actor.Balance,
		EpochReward:        state.ThisEpochReward,
		NextPerBlockReward: big.Div(state.ThisEpochReward, big.NewInt(builtin.ExpectedLeadersPerEpoch)),
	}

	r.m[epoch] = rs
}

// ForEpoch returns the RewardSummary (or nil) associated with the given epoch.
func (r *Rewards) ForEpoch(epoch abi.ChainEpoch) *RewardSummary {
	return r.m[epoch]
}
