package chaos

import (
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// ChaosActorCodeCID is the CID by which this kind of actor will be identified.
var ChaosActorCodeCID = func() cid.Cid {
	builder := cid.V1Builder{Codec: cid.Raw, MhType: multihash.IDENTITY}
	c, err := builder.Sum([]byte("fil/1/chaos"))
	if err != nil {
		panic(err)
	}
	return c
}()

// Address is the singleton address of this actor. It is value 97
// (builtin.FirstNonSingletonActorId - 3), as 99 is reserved for the burnt funds
// singleton, and 98 is the puppet actor.
var Address = func() address.Address {
	// the address before the burnt funds address (99) and the puppet actor (98)
	addr, err := address.NewIDAddress(97)
	if err != nil {
		panic(err)
	}
	return addr
}()
