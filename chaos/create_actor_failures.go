package chaos

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/runtime"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
	"github.com/ipfs/go-cid"
)

func (a Actor) CreateAccountActorWithAddr(rt runtime.Runtime, addr *address.Address) *adt.EmptyValue {
	rt.CreateActor(builtin.AccountActorCodeID, *addr)

	return nil
}

func (a Actor) CreateUnknownActor(rt runtime.Runtime, addr *address.Address) *adt.EmptyValue {
	var UnknownActorCodeID cid.Cid

	rt.CreateActor(UnknownActorCodeID, *addr)

	return nil
}
