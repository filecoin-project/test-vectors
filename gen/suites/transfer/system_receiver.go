package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func transferToSystemActor(sysAddr address.Address) func(v *Builder) {
	return func(v *Builder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		// Set up sender account.
		sender := v.Actors.Account(address.SECP256K1, initial)
		v.CommitPreconditions()

		// perform the transfer.
		v.Messages.Sugar().Transfer(sender.Robust, sysAddr, transfer, Nonce(0))
		v.CommitApplies()

		v.Assert.BalanceEq(sender.Robust, initial)
		v.Assert.Equal(v.PreRoot.String(), v.PostRoot.String(), "expected pre and post state root to be equal")

		// TODO: These tests may break in the future if sending to a system actor
		// becomes disallowed: https://github.com/filecoin-project/specs/issues/1069
		v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	}
}
