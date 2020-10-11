package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/lotus/conformance/chaos"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func createActor(addressSupplier func(v *MessageVectorBuilder) address.Address, actorCid cid.Cid, expectedCode exitcode.ExitCode) func(v *MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(1e9), GasPremium(1), GasFeeCap(200))

		var alice, bob AddressHandle
		v.Actors.AccountN(address.SECP256K1, abi.NewTokenAmount(1e18), &alice, &bob)
		v.CommitPreconditions()

		params := &chaos.CreateActorArgs{
			ActorCID: actorCid,
			Address:  addressSupplier(v),
		}

		if params.ActorCID == cid.Undef {
			params.ActorCID = builtin.SystemActorCodeID // use a good one, it'll be overridden.
			params.UndefActorCID = true
		}

		if params.Address == address.Undef {
			params.Address = MustNewIDAddr(100) // use a good one, it'll be overridden.
			params.UndefAddress = true
		}

		v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodCreateActor, MustSerialize(params), Nonce(0), Value(big.Zero()))
		v.CommitApplies()

		// make sure that we get the expected error code (usually
		// SysErrorIllegalArgument, but Ok if this is the control case)
		v.Assert.EveryMessageResultSatisfies(ExitCode(expectedCode))
		// make sure that gas is deducted from alice's account
		v.Assert.EveryMessageSenderSatisfies(BalanceUpdated(big.Zero()))
	}
}
