package main

import (
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/builtin"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

var basefee = abi.NewTokenAmount(100)

func minerPenalized(minerCnt int, messageFn func(v *TipsetVectorBuilder), checksFn func(v *TipsetVectorBuilder)) func(v *TipsetVectorBuilder) {
	return func(v *TipsetVectorBuilder) {
		v.SetInitialEpoch(1)

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

		// Will enroll any messages into StagedMessages.
		messageFn(v)

		ts := v.Tipsets.Next(basefee)
		for _, m := range miners {
			ts.Block(*m, 1, v.StagedMessages.All()...)
		}

		v.CommitApplies()

		var (
			cumPenalty    = big.Zero()
			firstMinerTip = big.Zero()
			burntGas      = big.Zero()
		)

		for _, am := range v.Tipsets.Messages() {
			penalty := am.Result.Penalty
			cumPenalty = big.Sum(cumPenalty, penalty)

			tip := am.Result.MinerTip
			v.Assert.Equal(GetMinerReward(am), tip)
			firstMinerTip = big.Sum(firstMinerTip, tip)

			burntGas = big.Add(burntGas, CalculateBurntGas(am))
		}

		// get the rewards schedule at the starting epoch.
		rewards := v.Rewards.ForEpoch(0)
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
