package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/lotus/conformance/chaos"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func callerAlwaysIDAddress(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	msg := v.Messages.Raw(alice.Robust, chaos.Address, chaos.MethodInspectRuntime, []byte{}, Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	var rt chaos.InspectRuntimeReturn
	MustDeserialize(msg.Result.Return, &rt)

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.Equal(alice.ID, rt.Caller)
}

func receiverAlwaysIDAddress(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))

	// setup a robust address for chaos
	chaosRobustAddr := v.Wallet.NewSECP256k1Account()
	chaosIDAddr, err := v.StateTracker.StateTree.RegisterNewAddress(chaosRobustAddr)
	v.Assert.NoError(err)

	err = v.StateTracker.StateTree.SetActor(chaosIDAddr, v.StateTracker.Header(chaos.Address))
	v.Assert.NoError(err)

	v.CommitPreconditions()

	msg := v.Messages.Raw(alice.ID, chaosRobustAddr, chaos.MethodInspectRuntime, []byte{}, Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	var rt chaos.InspectRuntimeReturn
	MustDeserialize(msg.Result.Return, &rt)

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.Equal(chaosIDAddr, rt.Receiver)
}
