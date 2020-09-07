package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"
	typegen "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/test-vectors/chaos"
	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func callerValidation(branch chaos.CallerValidationBranch, expectedCode exitcode.ExitCode) func(v *MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
		v.CommitPreconditions()

		cborBranch := typegen.CborInt(branch)
		v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodCallerValidation, MustSerialize(&cborBranch), Nonce(0), Value(big.Zero()))
		v.CommitApplies()

		v.Assert.EveryMessageResultSatisfies(ExitCode(expectedCode))
	}
}
