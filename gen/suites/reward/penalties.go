package main

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	reward2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/reward"

	"github.com/filecoin-project/lotus/chain/actors"
	. "github.com/filecoin-project/test-vectors/gen/builders"
)

var basefee = abi.NewTokenAmount(100)

// minerPenalized is a factory for tipset vectors that test whether miners were
// penalized properly.
//
// It takes the number of miners, a messages builder that'll execute in applies
// stage, and a checker function that'll execute in checks stage.
//
// Only a single tipset is generated, with ALL miners including ALL messages
// that have been staged.
//
// Note that, currently, penalties are only levied against the first miner
// to present a block with the penalisable message.
// See https://github.com/filecoin-project/lotus/issues/3491. In real networks,
// the first block will be the one with the lowest ticket, thus that miner will
// be the one to swallow the penalty.
//
// We also check that gas and penalties are properly sent to the
// burnt funds actor.
func minerPenalized(minerCnt int, messageFn func(v *TipsetVectorBuilder), checksFn func(v *TipsetVectorBuilder)) func(v *TipsetVectorBuilder) {
	return func(v *TipsetVectorBuilder) {
		v.SetInitialEpochOffset(1)

		miners := make([]*Miner, minerCnt)
		for i := range miners {
			miners[i] = new(Miner)
		}
		cfg := MinerActorCfg{
			SealProofType:  TestSealProofType,
			PeriodBoundary: 1000,
			OwnerBalance:   balance,
		}
		v.Actors.MinerN(cfg, miners...)
		v.CommitPreconditions()

		messageFn(v)

		// All StagedMessages are enrolled in all blocks in the singleton tipset.
		ts := v.Tipsets.Next(basefee)
		for _, m := range miners {
			ts.Block(*m, 1, v.StagedMessages.All()...)
		}

		v.CommitApplies()

		// calculate cumulative penalties, tips, and burnt gas.
		var (
			cumPenalty    = big.Zero()
			firstMinerTip = big.Zero()
			burntGas      = big.Zero()
		)
		for _, am := range v.Tipsets.Messages() {
			penalty := am.Result.GasCosts.MinerPenalty
			cumPenalty = big.Sum(cumPenalty, penalty)

			tip := am.Result.GasCosts.MinerTip
			v.Assert.Equal(GetMinerReward(am), tip)
			firstMinerTip = big.Sum(firstMinerTip, tip)

			burntGas = big.Add(burntGas, CalculateBurntGas(am))
		}

		if v.ProtocolVersion.Actors >= actors.Version2 {
			// actorsv2 introduced a penalty multiplier.
			cumPenalty = big.Mul(cumPenalty, big.NewInt(reward2.PenaltyMultiplier))
		}

		// get the rewards schedule at the starting epoch.
		rewards := v.Rewards.ForEpochOffset(0)
		for i, m := range miners {
			expected := rewards.NextPerBlockReward
			if i == 0 {
				// if this is the first miner, add the miner tips and deduct the penalties.
				expected = big.Sub(expected, cumPenalty)
				expected = big.Add(expected, firstMinerTip)
			}
			v.Assert.BalanceEq(m.MinerActorAddr.ID, expected)
		}

		// Verify that the accumulated penalties were sent to the
		// burnt funds actor.
		v.Assert.BalanceEq(builtin.BurntFundsActorAddr, big.Add(cumPenalty, burntGas))

		// Invoke checks function.
		if checksFn != nil {
			checksFn(v)
		}
	}
}
