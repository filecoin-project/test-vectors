package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func minerIncludesDuplicateMessages(v *TipsetVectorBuilder) {
	var (
		alice = v.Actors.Account(address.SECP256K1, balance1T)
		bob   = v.Actors.Account(address.BLS, balance1T)
	)

	var miner1, miner2, miner3 Miner
	v.Actors.MinerN(MinerActorCfg{
		SealProofType:  TestSealProofType,
		PeriodBoundary: 0,
		OwnerBalance:   big.Zero(),
	}, &miner1, &miner2, &miner3)

	v.CommitPreconditions()

	transferAmt := abi.NewTokenAmount(100)

	v.StagedMessages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))
	v.StagedMessages.Sugar().Transfer(alice.Robust, alice.Robust, Value(transferAmt), Nonce(0))
	v.StagedMessages.Sugar().Transfer(alice.Robust, alice.Robust, Value(transferAmt), Nonce(0))
	v.StagedMessages.Sugar().Transfer(alice.Robust, alice.Robust, Value(transferAmt), Nonce(0))

	v.StagedMessages.Sugar().Transfer(bob.Robust, bob.Robust, Value(transferAmt), Nonce(0))
	v.StagedMessages.Sugar().Transfer(bob.Robust, bob.Robust, Value(transferAmt), Nonce(0))
	v.StagedMessages.Sugar().Transfer(bob.Robust, bob.Robust, Value(transferAmt), Nonce(0))

	ts := v.Tipsets.Next(abi.NewTokenAmount(100))
	ts.Block(miner1, 1, v.StagedMessages.All()...)
	ts.Block(miner2, 1, v.StagedMessages.All()...)
	ts.Block(miner3, 1, v.StagedMessages.All()...)

	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.Len(v.Tipsets.Messages(), 2)
}
