package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func okSecpkBLSCosts(v *TipsetVectorBuilder) {
	var (
		alice = v.Actors.Account(address.SECP256K1, balance1T)
		bob   = v.Actors.Account(address.BLS, balance1T)
	)

	miner := v.Actors.Miner(MinerActorCfg{
		SealProofType:  TestSealProofType,
		PeriodBoundary: 0,
		OwnerBalance:   big.Zero(),
	})

	v.CommitPreconditions()

	transferAmt := abi.NewTokenAmount(100)

	v.StagedMessages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))
	secp := v.StagedMessages.Sugar().Transfer(alice.Robust, alice.Robust, Value(transferAmt), Nonce(0))
	bls := v.StagedMessages.Sugar().Transfer(bob.Robust, bob.Robust, Value(transferAmt), Nonce(0))

	ts := v.Tipsets.Next(abi.NewTokenAmount(100))
	ts.Block(miner, 1, secp, bls)

	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.Greater(secp.Result.GasUsed, bls.Result.GasUsed)
}

func failCoverReceiptGasCost(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, balance1T)
	v.CommitPreconditions()

	v.Messages.Sugar().Transfer(alice.ID, alice.ID, Value(transferAmnt), Nonce(0), GasPremium(1), GasLimit(8))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.SysErrOutOfGas))
}

func failCoverOnChainSizeGasCost(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, balance1T)
	v.CommitPreconditions()

	v.Messages.Sugar().Transfer(alice.ID, alice.ID, Value(transferAmnt), Nonce(0), GasPremium(10), GasLimit(1))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.SysErrOutOfGas))
}

func failCoverTransferAccountCreationGasStepwise(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	var alice, bob, charlie AddressHandle
	alice = v.Actors.Account(address.SECP256K1, balance1T)
	bob.Robust, charlie.Robust = MustNewSECP256K1Addr("1"), MustNewSECP256K1Addr("2")
	v.CommitPreconditions()

	var nonce uint64
	ref := v.Messages.Sugar().Transfer(alice.Robust, bob.Robust, Value(transferAmnt), Nonce(nonce))
	nonce++
	v.Messages.ApplyOne(ref)
	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))

	// decrease the gas cost by `gasStep` for each apply and ensure `SysErrOutOfGas` is always returned.
	trueGas := ref.Result.GasUsed
	gasStep := trueGas / 100
	for tryGas := trueGas - gasStep; tryGas > 0; tryGas -= gasStep {
		v.Messages.Sugar().Transfer(alice.Robust, charlie.Robust, Value(transferAmnt), Nonce(nonce), GasPremium(1), GasLimit(tryGas))
		nonce++
	}
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.SysErrOutOfGas), ref)
}
