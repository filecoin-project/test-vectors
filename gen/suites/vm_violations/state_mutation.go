package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/test-vectors/chaos"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func mutateState(value string, mutType chaos.MutateStateType, checksFn func(*MessageVectorBuilder, string)) func(*MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		sender := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
		v.CommitPreconditions()

		v.Messages.Raw(
			sender.ID,
			chaos.Address,
			chaos.MethodMutateState,
			MustSerialize(&chaos.MutateStateArgs{Type: mutType, Value: value}),
			Value(big.Zero()),
			Nonce(0),
		)
		v.CommitApplies()

		checksFn(v, value)
	}
}
