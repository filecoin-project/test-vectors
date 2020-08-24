package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"

	"github.com/filecoin-project/test-vectors/chaos"
	. "github.com/filecoin-project/test-vectors/gen/builders"
	. "github.com/filecoin-project/test-vectors/schema"
)

func main() {
	g := NewGenerator()
	defer g.Wait()

	g.MessageVectorGroup("caller_validation",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "none",
				Version: "v1",
				Desc:    "verifies that an actor that performs no caller validation fails",
			},
			Selector: "chaos_actor=true",
			Func:     callerValidation(&chaos.CallerValidationBranchNone, exitcode.SysErrorIllegalActor),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "twice",
				Version: "v1",
				Desc:    "verifies that an actor that validates the caller twice fails",
			},
			Selector: "chaos_actor=true",
			Func:     callerValidation(&chaos.CallerValidationBranchTwice, exitcode.SysErrorIllegalActor),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "nil-allowed-address-set",
				Version: "v1",
				Desc:    "verifies that an actor that validates against a nil allowed address set fails",
			},
			Selector: "chaos_actor=true",
			Func:     callerValidation(&chaos.CallerValidationBranchAddrNilSet, exitcode.SysErrForbidden),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "nil-allowed-type-set",
				Version: "v1",
				Desc:    "verifies that an actor that validates against a nil allowed type set fails",
			},
			Selector: "chaos_actor=true",
			Func:     callerValidation(&chaos.CallerValidationBranchTypeNilSet, exitcode.SysErrForbidden),
		},
	)

	// Build an unknown Actor CID.
	unknownCid, err := cid.V1Builder{Codec: cid.Raw, MhType: multihash.IDENTITY}.Sum([]byte("fil/1/unknown"))
	if err != nil {
		panic(err)
	}

	bobAddr := func(v *Builder) address.Address { return v.Actors.Handles()[1].Robust }
	goodAddr := func(v *Builder) address.Address { return v.Actors.Handles()[1].NextActorAddress(100, 0) }
	undefAddr := func(v *Builder) address.Address { return address.Undef }

	g.MessageVectorGroup("actor_creation",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fails-with-existing-address",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an existing address",
			},
			Selector: "chaos_actor=true",
			Func:     createActor(bobAddr, builtin.AccountActorCodeID),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fails-with-undef-addr",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an address.Undef",
			},
			Selector: "chaos_actor=true",
			Func:     createActor(undefAddr, builtin.AccountActorCodeID),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fails-with-unknown-actor-cid",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an unknown actor code CID",
			},
			Selector: "chaos_actor=true",
			Func:     createActor(goodAddr, unknownCid),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fails-with-unknown-actor-cid-undef-addr",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an unknown actor code CID and an undef address",
			},
			Selector: "chaos_actor=true",
			Func:     createActor(undefAddr, unknownCid),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fails-with-undef-actor-cid-undef-addr",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided an undef actor code CID and an undef address",
			},
			Selector: "chaos_actor=true",
			Func:     createActor(undefAddr, cid.Undef),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fails-with-good-addr-undef-cid",
				Version: "v1",
				Desc:    "verifies that CreateActor aborts when provided a valid address, but an undef CID",
			},
			Selector: "chaos_actor=true",
			Func:     createActor(goodAddr, cid.Undef),
		},
	)
}
