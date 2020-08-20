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

// CallerValidationBranch is an enum used to select a branch in the
// CallerValidation method.
type CallerValidationBranch big.Int

var (
	CallerValidationBranchNone       = big.NewIntUnsigned(0)
	CallerValidationBranchTwice      = big.NewIntUnsigned(1)
	CallerValidationBranchAddrNilSet = big.NewIntUnsigned(2)
	CallerValidationBranchTypeNilSet = big.NewIntUnsigned(3)
)

const (
	_                      = 0 // skip zero iota value; first usage of iota gets 1.
	MethodCallerValidation = builtin.MethodConstructor + iota
	MethodCreateAccountActorWithAddr
	MethodCreateUnknownActor
	MethodAbortWithSystemExitCode
)

func (a Actor) Exports() []interface{} {
	return []interface{}{
		builtin.MethodConstructor:        a.Constructor,
		MethodCallerValidation:           a.CallerValidation,
		MethodCreateAccountActorWithAddr: a.CreateAccountActorWithAddr,
		MethodCreateUnknownActor:         a.CreateUnknownActor,
		MethodAbortWithSystemExitCode:    a.AbortWithSystemExitCode,
	}
}

var _ abi.Invokee = Actor{}

func (a Actor) Constructor(_ runtime.Runtime, _ *adt.EmptyValue) *adt.EmptyValue {
	panic("constructor should not be called; the Chaos actor is a singleton actor")
}

// CallerValidation violates VM call validation constraints.
//
//  CallerValidationBranchNone performs no validation.
//  CallerValidationBranchTwice validates twice.
//  CallerValidationBranchAddrNilSet validates against an empty caller
//  address set.
//  CallerValidationBranchTypeNilSet validates against an empty caller type set.
func (a Actor) CallerValidation(rt runtime.Runtime, branch *big.Int) *adt.EmptyValue {
	if branch == nil {
		panic("no branch passed to CallerValidation")
	}

	switch branch.Uint64() {
	case CallerValidationBranchNone.Uint64():
	case CallerValidationBranchTwice.Uint64():
		rt.ValidateImmediateCallerAcceptAny()
		rt.ValidateImmediateCallerAcceptAny()
	case CallerValidationBranchAddrNilSet.Uint64():
		rt.ValidateImmediateCallerIs()
	case CallerValidationBranchTypeNilSet.Uint64():
		rt.ValidateImmediateCallerType()
	default:
		panic("invalid branch passed to CallerValidation")
	}

	return nil
}
