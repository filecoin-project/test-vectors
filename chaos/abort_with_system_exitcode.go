package chaos

import (
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/runtime"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
)

func (a Actor) AbortWithSystemExitCode(rt runtime.Runtime, code *big.Int) *adt.EmptyValue {
	rt.Abortf(exitcode.ExitCode(code.Int64()), "aborting with code: %d", code)

	return nil
}
