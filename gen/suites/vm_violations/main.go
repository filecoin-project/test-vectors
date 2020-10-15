package main

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/exitcode"
	builtin0 "github.com/filecoin-project/specs-actors/actors/builtin"
	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"

	"github.com/filecoin-project/lotus/conformance/chaos"

	. "github.com/filecoin-project/test-vectors/gen/builders"
	"github.com/filecoin-project/test-vectors/schema"
)

func main() {
	g := NewGenerator()
	defer g.Close()

	g.Group("caller_validation",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "none",
				Version: "v1",
				Desc:    "verifies that an actor that performs no caller validation fails",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: callerValidation(chaos.CallerValidationArgs{Branch: chaos.CallerValidationBranchNone}, exitcode.SysErrorIllegalActor),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "twice",
				Version: "v1",
				Desc:    "verifies that an actor that validates the caller twice fails",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: callerValidation(chaos.CallerValidationArgs{Branch: chaos.CallerValidationBranchTwice}, exitcode.SysErrorIllegalActor),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "nil-allowed-address-set",
				Version: "v1",
				Desc:    "verifies that an actor that validates against a nil allowed address set fails",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: callerValidation(chaos.CallerValidationArgs{Branch: chaos.CallerValidationBranchIsAddress}, exitcode.SysErrForbidden),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-incorrect-caller-addr",
				Version: "v1",
				Desc:    "verifies that an actor that validates against an address set that does not include the caller addr fails",
			},
			Selector: map[string]string{"chaos_actor": "true"},
			MessageFunc: callerValidation(chaos.CallerValidationArgs{
				Branch: chaos.CallerValidationBranchIsAddress,
				// caller address will be a brand new account NOT the system actor address
				Addrs: []address.Address{builtin0.SystemActorAddr},
			}, exitcode.SysErrForbidden),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "nil-allowed-type-set",
				Version: "v1",
				Desc:    "verifies that an actor that validates against a nil allowed type set fails",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: callerValidation(chaos.CallerValidationArgs{Branch: chaos.CallerValidationBranchIsType}, exitcode.SysErrForbidden),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-incorrect-caller-type",
				Version: "v1",
				Desc:    "verifies that an actor that validates against a actor type set that does not include the caller type",
			},
			Selector: map[string]string{"chaos_actor": "true"},
			MessageFunc: callerValidation(chaos.CallerValidationArgs{
				Branch: chaos.CallerValidationBranchIsType,
				// caller will be of type account actor NOT system actor
				Types: []cid.Cid{builtin0.SystemActorCodeID},
			}, exitcode.SysErrForbidden),
		},
	)

	// Build an unknown Actor CID.
	unknownCid, err := cid.V1Builder{Codec: cid.Raw, MhType: multihash.IDENTITY}.Sum([]byte("fil/1/unknown"))
	if err != nil {
		panic(err)
	}

	// CreateActor requires ID addresses; if it receives a Robust address, it'll
	// try to resolve the ID address from the init actor. But we're not
	// adding a mapping to the init actor here, so that would've failed for a
	// different reason (red herring).
	bobAddr := func(v *MessageVectorBuilder) address.Address { return v.Actors.AccountHandles()[1].ID }
	goodAddr := func(v *MessageVectorBuilder) address.Address { return MustNextIDAddr(v.Actors.AccountHandles()[1].ID) }
	undefAddr := func(v *MessageVectorBuilder) address.Address { return address.Undef }

	g.Group("actor_creation",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "control-ok-with-good-address-good-cid",
				Version: "v1",
				Desc:    "control test case to verify that correct actor creation messages do indeed succeed",
			},
			SupportedVersions: KnownProtocolVersionsBefore("actorsv2"),
			Selector:          map[string]string{"chaos_actor": "true"},
			MessageFunc:       createActor(goodAddr, builtin0.AccountActorCodeID, exitcode.Ok),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "control-ok-with-good-address-good-cid",
				Version: "v2",
				Desc:    "control test case to verify that correct actor creation messages do indeed succeed",
			},
			SupportedVersions: KnownProtocolVersionsFrom("actorsv2"),
			Selector:          map[string]string{"chaos_actor": "true"},
			MessageFunc:       createActor(goodAddr, builtin2.AccountActorCodeID, exitcode.Ok),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-with-existing-address",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an existing address",
			},
			SupportedVersions: KnownProtocolVersionsBefore("actorsv2"),
			Selector:          map[string]string{"chaos_actor": "true"},
			MessageFunc:       createActor(bobAddr, builtin0.AccountActorCodeID, exitcode.SysErrorIllegalArgument),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-with-existing-address",
				Version: "v2",
				Desc:    "verifies that CreateActor aborts when provided an existing address",
			},
			SupportedVersions: KnownProtocolVersionsFrom("actorsv2"),
			Selector:          map[string]string{"chaos_actor": "true"},
			MessageFunc:       createActor(bobAddr, builtin2.AccountActorCodeID, exitcode.SysErrorIllegalArgument),
		},
		//
		// TODO this is commented because it causes an uncontrolled VM error
		//  with no Result or post root whatsoever. We do not support such
		//  failure modes in ModeLenientAssertions. This needs to be fixed
		//  upstream and then enabled.
		//
		// &VectorDef{
		// 	Metadata: &Metadata{
		// 		ID:      "fails-with-undef-addr",
		// 		Version: "v1",
		// 		Desc:    "verifies that CreateActor aborts when provided an address.Undef",
		// 	},
		// 	Mode:     ModeLenientAssertions,
		// 	Hints:    []string{HintIncorrect, HintNegate},
		// 	Selector: map[string]string{"chaos_actor":"true"},
		// 	MessageFunc:     createActor(undefAddr, builtin.AccountActorCodeID, exitcode.SysErrorIllegalArgument),
		// },
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-with-unknown-actor-cid",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an unknown actor code CID",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: createActor(goodAddr, unknownCid, exitcode.SysErrorIllegalArgument),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-with-unknown-actor-cid-undef-addr",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an unknown actor code CID and an undef address",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: createActor(undefAddr, unknownCid, exitcode.SysErrorIllegalArgument),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-with-undef-actor-cid-undef-addr",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an undef actor code CID and an undef address",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: createActor(undefAddr, cid.Undef, exitcode.SysErrorIllegalArgument),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-with-good-addr-undef-cid",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided a valid address, but an undef CID",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: createActor(goodAddr, cid.Undef, exitcode.SysErrorIllegalArgument),
		},
	)

	g.Group("address_resolution",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "resolve-address-id-identity",
				Version: "v1",
				Desc:    "verifies that runtime.ResolveAddress is an identity function for ID type addresses",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: actorResolutionIDIdentity,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "resolve-address-bad-id-identity",
				Version: "v1",
				Desc:    "verifies that runtime.ResolveAddress is an identity function for ID type addresses",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: actorResolutionInvalidIdentity,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "resolve-address-nonexistant",
				Version: "v1",
				Desc:    "verifies that runtime.ResolveAddress on non-existant addresses are undefined",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: actorResolutionNonexistant,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "resolve-address-bls-lookup",
				Version: "v1",
				Desc:    "verifies that runtime.ResolveAddress on known addresses are resolved",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: actorResolutionBlsExistant,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "resolve-address-secp-lookup",
				Version: "v1",
				Desc:    "verifies that runtime.ResolveAddress on known addresses are resolved",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: actorResolutionSecpExistant,
		},
	)

	valPfx := "vm_violations/state_mutation/"

	g.Group("state_mutation",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "in-transaction",
				Version: "v1",
				Desc:    "test an actor can mutate state within a transaction",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: mutateState(valPfx+"in-transaction", chaos.MutateInTransaction, exitcode.Ok),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "readonly",
				Version: "v1",
				Desc:    "test an actor cannot ILLEGALLY mutate readonly state",
				Comment: "should abort with SysErrorIllegalActor, not succeed with Ok, see https://github.com/filecoin-project/lotus/issues/3545",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: mutateState(valPfx+"readonly", chaos.MutateReadonly, exitcode.SysErrorIllegalActor),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "after-transaction",
				Version: "v1",
				Desc:    "test an actor cannot ILLEGALLY mutate state acquired for transaction but used after the transaction has ended",
				Comment: "should abort with SysErrorIllegalActor, not succeed with Ok, see https://github.com/filecoin-project/lotus/issues/3545",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: mutateState(valPfx+"after-transaction", chaos.MutateAfterTransaction, exitcode.SysErrorIllegalActor),
		},
	)

	actorAbortVectors := []*VectorDef{
		{
			Metadata: &Metadata{
				ID:      "custom-exit-code",
				Version: "v1",
				Desc:    "actors can abort with custom exit codes",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: actorAbort(exitcode.FirstActorSpecificExitCode, "custom exit code abort", exitcode.FirstActorSpecificExitCode),
		},
		{
			Metadata: &Metadata{
				ID:      "negative-exit-code",
				Version: "v1",
				Desc:    "actors should not abort with negative exit codes",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: actorAbort(-1, "negative exit code abort", exitcode.SysErrorIllegalActor),
		},
		{
			Metadata: &Metadata{
				ID:      "no-exit-code",
				Version: "v1",
				Desc:    "actor failure, a panic with no associated exit code",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: actorPanic("no exit code abort"),
		},
	}

	sysExitCodes := []exitcode.ExitCode{
		exitcode.SysErrSenderInvalid,
		exitcode.SysErrSenderStateInvalid,
		exitcode.SysErrInvalidMethod,
		exitcode.SysErrReserved1,
		exitcode.SysErrInvalidReceiver,
		exitcode.SysErrInsufficientFunds,
		exitcode.SysErrOutOfGas,
		exitcode.SysErrForbidden,
		exitcode.SysErrorIllegalActor,
		exitcode.SysErrorIllegalArgument,
		exitcode.SysErrReserved2,
		exitcode.SysErrReserved3,
		exitcode.SysErrReserved4,
		exitcode.SysErrReserved5,
		exitcode.SysErrReserved6,
	}

	for _, xc := range sysExitCodes {
		actorAbortVectors = append(actorAbortVectors, &VectorDef{
			Metadata: &Metadata{
				ID:      fmt.Sprintf("system-exit-code-%d", xc),
				Version: "v1",
				Desc:    fmt.Sprintf("actors should not abort with %s", xc),
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: actorAbort(xc, fmt.Sprintf("%s abort", xc), exitcode.SysErrorIllegalActor),
		})
	}

	g.Group("actor_abort", actorAbortVectors...)

	g.Group("runtime_inspections",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "caller-always-id-address",
				Version: "v1",
				Desc:    "verify Runtime.Caller() returns an ID address even when message is sent from a robust address",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: callerAlwaysIDAddress,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "receiver-always-id-address",
				Version: "v1",
				Desc:    "verify Runtime.Receiver() returns an ID address even when message is sent to a robust address",
				Comment: "the call to Runtime.Receiver() should return an ID address but returns the robust address that the message was sent to, fixed by https://github.com/filecoin-project/lotus/pull/3589",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: receiverAlwaysIDAddress,
		},
	)
}
