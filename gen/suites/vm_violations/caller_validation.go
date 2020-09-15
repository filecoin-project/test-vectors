package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/lotus/conformance/chaos"
	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func callerValidation(buildArgs func() chaos.CallerValidationArgs, expectedCode exitcode.ExitCode) func(v *MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
		v.CommitPreconditions()

		args := buildArgs()
		v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodCallerValidation, MustSerialize(&args), Nonce(0), Value(big.Zero()))
		v.CommitApplies()

		v.Assert.EveryMessageResultSatisfies(ExitCode(expectedCode))
	}
}

func robustCallerValidation(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	// send from robust address, allowing ID address - should succeed
	from := alice.Robust
	args := chaos.CallerValidationArgs{Branch: chaos.CallerValidationBranchIs, Addrs: []address.Address{alice.ID}}

	v.Messages.Raw(from, chaos.Address, chaos.MethodCallerValidation, MustSerialize(&args), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
}

func robustReceiverValidation(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	initial := int64(1_000_000_000_000)
	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(initial))

	// setup a robust address for chaos
	chaosRobustAddr := v.Wallet.NewSECP256k1Account()
	chaosIDAddr, err := v.StateTracker.StateTree.RegisterNewAddress(chaosRobustAddr)
	v.Assert.NoError(err)

	err = v.StateTracker.StateTree.SetActor(chaosIDAddr, v.StateTracker.Header(chaos.Address))
	v.Assert.NoError(err)

	v.CommitPreconditions()

	// Verify that Runtime.Message().Caller() returns ID address:
	// Send from alice's robust address, allowing her ID address. Should succeed
	// if Runtime.Message().Caller() returns ID addresses.
	args := chaos.CallerValidationArgs{Branch: chaos.CallerValidationBranchIs, Addrs: []address.Address{alice.ID}}
	// ...also send some fil here to allow chaos to send a message to itself afterwards.
	m0 := v.Messages.Raw(alice.Robust, chaosIDAddr, chaos.MethodCallerValidation, MustSerialize(&args), Nonce(0), Value(big.NewInt(initial/2)))

	// Verify that Runtime.Message().Receiver() returns an ID address:
	// Tell chaos to send a message to its (robust) self, calling
	// CallerValidation method, using the branch that validates that the caller
	// IS the receiver. We've already verified above that
	// Runtime.Message().Caller() returns ID addresses, so provided
	// Runtime.Message().Receiver() returns an ID address, not a robust
	// address, then this should work.
	m1 := v.Messages.Typed(alice.ID, chaosIDAddr, ChaosSend(&chaos.SendArgs{
		To:     chaosRobustAddr,
		Value:  big.Zero(),
		Method: chaos.MethodCallerValidation,
		Params: MustSerialize(&chaos.CallerValidationArgs{Branch: chaos.CallerValidationBranchIsReceiver}),
	}), Nonce(1), Value(big.Zero()))

	v.CommitApplies()

	v.Assert.ExitCodeEq(m0.Result.ExitCode, exitcode.Ok)
	v.Assert.ExitCodeEq(m1.Result.ExitCode, exitcode.Ok)

	var chaosRet chaos.SendReturn
	MustDeserialize(m1.Result.MessageReceipt.Return, &chaosRet)

	// The inner message should succeed.
	v.Assert.ExitCodeEq(chaosRet.Code, exitcode.Ok)
}
