package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	init_ "github.com/filecoin-project/specs-actors/actors/builtin/init"
	"github.com/filecoin-project/specs-actors/actors/builtin/paych"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func createActorInitExecUnparsableParams(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	// Set up sender and receiver accounts.
	var sender, receiver AddressHandle
	v.Actors.AccountN(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000), &sender, &receiver)
	v.CommitPreconditions()

	// Valid message for construction of a payment channel.
	createMsg := v.Messages.Typed(sender.ID, builtin.InitActorAddr, InitExec(&init_.ExecParams{
		CodeCID:           builtin.PaymentChannelActorCodeID,
		ConstructorParams: MustSerialize(&paych.ConstructorParams{From: sender.ID, To: receiver.ID}),
	}), Value(abi.NewTokenAmount(10_000)), Nonce(0))

	// mangle the InitExec params to form an invalid CBOR payload.
	createMsg.Message.Params = append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, createMsg.Message.Params...)

	v.CommitApplies()

	// make sure that we get ErrSerialization error code -- this assert currently fails.
	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.ErrSerialization))
	// make sure that gas (not value) is deducted from senders's account
	// (the BalanceUpdated predicate omits deducting the value if the exit code != success)
	v.Assert.EveryMessageSenderSatisfies(BalanceUpdated(big.Zero()))
}

func createActorCtorUnparsableParamsViaInitExec(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	// Set up sender and receiver accounts.
	var sender, receiver AddressHandle
	v.Actors.AccountN(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000), &sender, &receiver)
	v.CommitPreconditions()

	ctorparams := MustSerialize(&paych.ConstructorParams{From: sender.ID, To: receiver.ID})
	ctorparams = append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, ctorparams...)

	// Valid message for construction of a payment channel.
	v.Messages.Typed(sender.ID, builtin.InitActorAddr, InitExec(&init_.ExecParams{
		CodeCID:           builtin.PaymentChannelActorCodeID,
		ConstructorParams: ctorparams,
	}), Value(abi.NewTokenAmount(10_000)), Nonce(0))

	v.CommitApplies()

	// make sure that we get ErrSerialization error code -- this assert currently fails.
	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.ErrSerialization))
	// make sure that gas (not value) is deducted from senders's account
	// (the BalanceUpdated predicate omits deducting the value if the exit code != success)
	v.Assert.EveryMessageSenderSatisfies(BalanceUpdated(big.Zero()))
}
