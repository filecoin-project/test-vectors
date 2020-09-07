package main

import (
	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/exitcode"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

// transferToSystemActor tests that an amount can be successfully transferred
// to a system actor. The (optional) calcExtra parameter calculates an amount to
// expect to see in the receiver's account due to the fact that the receiver is
// a special actor that is sent funds during the process of a transfer.
//
// e.g. the reward actor receives the miner tip (gas limit * gas premium), which
// is held for the block miner.
//
// e.g. The burnt funds actor receives the gas burn base fee.
//
// TODO: These tests may break in the future if sending to a system actor
// becomes disallowed: https://github.com/filecoin-project/specs/issues/1069
func transferToSystemActor(sysAddr address.Address, calcExtra func(am *ApplicableMessage) big.Int) func(v *MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(gasLimit), GasPremium(gasPremium), GasFeeCap(gasFeeCap))
		initial := abi.NewTokenAmount(1_000_000_000_000)
		transfer := abi.NewTokenAmount(123)

		// Set up sender account.
		sender := v.Actors.Account(address.SECP256K1, initial)
		v.CommitPreconditions()

		sysActor := v.StateTracker.Header(sysAddr)

		// Calculate the end balance.
		endBal := big.Add(sysActor.Balance, transfer)

		ref := v.Messages.Sugar().Transfer(sender.Robust, sysAddr, Value(transfer), Nonce(0))
		v.Messages.ApplyOne(ref)
		v.CommitApplies()

		// Add any extra
		if calcExtra != nil {
			endBal = big.Add(endBal, calcExtra(ref))
		}

		// System actor received the funds.
		v.Assert.BalanceEq(sysAddr, endBal)
		// Sender sent the funds + gas.
		v.Assert.EveryMessageSenderSatisfies(BalanceUpdated(big.Zero()))
		// Everything is great.
		v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	}
}
