package chaos

import (
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/runtime"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
)

//go:generate go run ./gen

// Actor is a chaos actor. It implements a variety of illegal behaviours that
// trigger violations of VM invariants. These behaviours are not found in
// production code, but are important to test that the VM constraints are
// properly enforced.
//
// The chaos actor is being incubated and its behaviour and ABI be standardised
// shortly. Its CID is ChaosActorCodeCID, and its singleton address is 97 (Address).
// It cannot be instantiated via the init actor, and its constructor panics.
//
// Test vectors relying on the chaos actor being deployed will carry selector
// "chaos_actor=true".
type Actor struct{}

type CallerValidationBranch big.Int

var (
	CallerValidationBranchNone                      = big.NewIntUnsigned(0)
	CallerValidationBranchTwice                     = big.NewIntUnsigned(1)
	CallerValidationBranchImmediateCallerAddrNoArgs = big.NewIntUnsigned(2)
	CallerValidationBranchImmediateCallerTypeNoArgs = big.NewIntUnsigned(3)
)

const (
	MethodCallerValidation = builtin.MethodConstructor + 1 + iota
)

func (a Actor) Exports() []interface{} {
	return []interface{}{
		builtin.MethodConstructor: a.Constructor,
		2:                         a.CallerValidation,
	}
}

var _ abi.Invokee = Actor{}

func (a Actor) Constructor(_ runtime.Runtime, _ *adt.EmptyValue) *adt.EmptyValue {
	panic("constructor should not be called; the Chaos actor is a singleton actor")
}

func (a Actor) CallerValidation(rt runtime.Runtime, branch *big.Int) *adt.EmptyValue {
	if branch == nil {
		panic("no branch passed to CallerValidation")
	}

	switch branch.Uint64() {
	case CallerValidationBranchNone.Uint64():
	case CallerValidationBranchTwice.Uint64():
		rt.ValidateImmediateCallerAcceptAny()
		rt.ValidateImmediateCallerAcceptAny()
	case CallerValidationBranchImmediateCallerAddrNoArgs.Uint64():
		rt.ValidateImmediateCallerIs()
	case CallerValidationBranchImmediateCallerTypeNoArgs.Uint64():
		rt.ValidateImmediateCallerType()
	default:
		panic("invalid branch passed to CallerValidation")
	}

	return nil
}
