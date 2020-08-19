package main

import (
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"

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
}
