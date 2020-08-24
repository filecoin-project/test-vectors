package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"

	"github.com/filecoin-project/test-vectors/chaos"
	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func actorResolutionIdentity(v *Builder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&builtin.SystemActorAddr), Nonce(0), Value(big.Zero()))

	invalidIDAddr, _ := address.NewIDAddress(77)
	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&invalidIDAddr), Nonce(1), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
}

func actorResolutionNonexistant(v *Builder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	invalidActorAddr, _ := address.NewActorAddress([]byte("invalid"))
	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&invalidActorAddr), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.EveryMessageResultSatisfies(MessageReturns(&builtin.SystemActorAddr))
}

func actorResolutionExistant(v *Builder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&alice.ID), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.EveryMessageResultSatisfies(MessageReturns(&alice.ID))
}

