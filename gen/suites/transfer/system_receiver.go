package main

import (
	"github.com/filecoin-project/specs-actors/actors/abi/big"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

// TODO: These tests may break in the future if sending to a system actor
// becomes disallowed: https://github.com/filecoin-project/specs/issues/1069
func transferToSystemActor(sysAddr address.Address) func(v *Builder) {
	return func(v *Builder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))
		initial := abi.NewTokenAmount(1_000_000_000_000)
		transfer := abi.NewTokenAmount(10)

		// Set up sender account.
		sender := v.Actors.Account(address.SECP256K1, initial)
		v.CommitPreconditions()

		sysActor, err := v.StateTree.GetActor(sysAddr)
		v.Assert.NoError(err, "failed to fetch actor %s from state", sysAddr)

		// Calculate the end balance.
		endBal := big.Add(sysActor.Balance, transfer)

		v.Messages.Sugar().Transfer(sender.Robust, sysAddr, Value(transfer), Nonce(0))
		v.CommitApplies()

		// System actor received the funds.
		v.Assert.BalanceEq(sysAddr, endBal)
		// Sender sent the funds + gas.
		v.Assert.EveryMessageSenderSatisfies(BalanceUpdated(big.Zero()))
		// Everything is great.
		v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	}
}
