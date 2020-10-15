package main

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/specs-actors/actors/builtin"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/lotus/conformance/chaos"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func actorResolutionIDIdentity(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&builtin.SystemActorAddr), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.EveryMessageResultSatisfies(MessageReturns(&chaos.ResolveAddressResponse{builtin.SystemActorAddr, true}))
}

func actorResolutionInvalidIdentity(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	invalidIDAddr, _ := address.NewIDAddress(77)
	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&invalidIDAddr), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.EveryMessageResultSatisfies(MessageReturns(&chaos.ResolveAddressResponse{invalidIDAddr, true}))
}

func actorResolutionNonexistant(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	invalidActorAddr, _ := address.NewActorAddress([]byte("invalid"))
	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&invalidActorAddr), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.EveryMessageResultSatisfies(MessageReturns(&chaos.ResolveAddressResponse{builtin.SystemActorAddr, false}))
}

func actorResolutionSecpExistant(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&alice.ID), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.EveryMessageResultSatisfies(MessageReturns(&chaos.ResolveAddressResponse{alice.ID, true}))
}

func actorResolutionBlsExistant(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.BLS, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodResolveAddress, MustSerialize(&alice.ID), Nonce(0), Value(big.Zero()))
	v.CommitApplies()

	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.EveryMessageResultSatisfies(MessageReturns(&chaos.ResolveAddressResponse{alice.ID, true}))
}
