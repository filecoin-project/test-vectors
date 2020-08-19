package chaos

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/runtime"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
)

func (a Actor) CreateAccountActorWithAddr(rt runtime.Runtime, addr *address.Address) *adt.EmptyValue {
	rt.CreateActor(builtin.AccountActorCodeID, *addr)

	return nil
}
