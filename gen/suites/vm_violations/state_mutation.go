package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/lotus/conformance/chaos"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func mutateState(value string, mutBranch chaos.MutateStateBranch, expectedCode exitcode.ExitCode) func(*MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		sender := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
		v.CommitPreconditions()

		v.Messages.Raw(
			sender.ID,
			chaos.Address,
			chaos.MethodMutateState,
			MustSerialize(&chaos.MutateStateArgs{Branch: mutBranch, Value: value}),
			Value(big.Zero()),
			Nonce(0),
		)
		v.CommitApplies()

		v.Assert.LastMessageResultSatisfies(ExitCode(expectedCode))

		var st chaos.State
		v.StateTracker.ActorState(chaos.Address, &st)

		// verify the state was/wasn't updated
		if expectedCode == exitcode.Ok {
			v.Assert.Equal(value, st.Value)
		} else {
			v.Assert.NotEqual(value, st.Value)
		}
	}
}
