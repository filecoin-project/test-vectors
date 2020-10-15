package main

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/specs-actors/actors/builtin"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func minersAwardedNoPremiums(v *TipsetVectorBuilder) {
	v.SetInitialEpochOffset(1)

	var minerA, minerB, minerC Miner
	cfg := MinerActorCfg{
		SealProofType:  TestSealProofType,
		PeriodBoundary: 1000,
		OwnerBalance:   balance,
	}
	v.Actors.MinerN(cfg, &minerA, &minerB, &minerC)
	v.CommitPreconditions()

	v.StagedMessages.SetDefaults(GasLimit(1_000_000_000), GasPremium(0), GasFeeCap(200))
	transfer1 := v.StagedMessages.Sugar().Transfer(minerA.OwnerAddr.Robust, minerB.OwnerAddr.ID, Value(abi.NewTokenAmount(1)), Nonce(0))
	transfer2 := v.StagedMessages.Sugar().Transfer(minerA.OwnerAddr.Robust, minerB.OwnerAddr.ID, Value(abi.NewTokenAmount(1)), Nonce(1))

	// Add 10 null rounds.
	v.Tipsets.NullRounds(10)

	// EpochOffset 11 -- we'll access the reward actor state at this epoch
	// to get the reward policy.
	ts1 := v.Tipsets.Next(abi.NewTokenAmount(100))
	v.Assert.EqualValues(11, ts1.EpochOffset)
	ts1.Block(minerA, 1, transfer1)
	ts1.Block(minerB, 1, transfer1)
	ts1.Block(minerC, 1, transfer1)

	// EpochOffset 12.
	ts2 := v.Tipsets.Next(abi.NewTokenAmount(100))
	v.Assert.EqualValues(12, ts2.EpochOffset)
	ts2.Block(minerA, 1, transfer2)
	ts2.Block(minerB, 1, transfer2)
	ts2.Block(minerC, 1, transfer2)

	v.CommitApplies()

	// Assert all OK, and that balances of all senders are updated.
	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.EveryMessageSenderSatisfies(BalanceUpdated(big.Zero()))

	// Verify that the reward actor balance has been updated between
	// epoch 11 and epoch 12, based on the wincount=3.
	policy := v.Rewards.ForEpochOffset(ts1.EpochOffset)

	// state at epoch 11.
	state11 := v.StateTracker.Fork(ts1.PostStateRoot)
	prev := state11.Balance(builtin.RewardActorAddr)
	minus := big.Mul(policy.NextPerBlockReward, big.NewInt(3)) // wincount = 3.
	v.Assert.BalanceEq(builtin.RewardActorAddr, big.Sub(prev, minus))

	// TODO add Assert.EveryMinerSatisfies.
	for _, m := range []Miner{minerA, minerB, minerC} {
		addr := m.MinerActorAddr.ID
		exp := big.Add(state11.Balance(addr), policy.NextPerBlockReward)
		v.Assert.BalanceEq(addr, exp)
	}

	// Verify that the burnt gas has been sent to the burnt funds actor.
	v.Assert.BalanceEq(builtin.BurntFundsActorAddr, big.Sum(CalculateBurntGas(transfer1), CalculateBurntGas(transfer2)))
}
