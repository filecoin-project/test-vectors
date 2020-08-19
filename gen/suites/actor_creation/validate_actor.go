package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"

	"github.com/filecoin-project/test-vectors/chaos"
	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func createAccountActorWithExistingAddr(v *Builder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	bob := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	balanceBefore := v.Actors.Balance(alice.Robust)
	msg := v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodCreateAccountActorWithAddr, MustSerialize(&bob.Robust), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.SysErrorIllegalArgument))  // make sure that we get SysErrorIllegalArgument error code
	v.Assert.BalanceEq(alice.Robust, big.Sub(balanceBefore, CalculateDeduction(msg))) // make sure that gas is deducted from alice's account
}
