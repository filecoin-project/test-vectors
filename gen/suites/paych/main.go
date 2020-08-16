package main

import (
	"github.com/filecoin-project/specs-actors/actors/abi"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

var (
	initialBal = abi.NewTokenAmount(1_000_000_000_000)
	toSend     = abi.NewTokenAmount(10_000)
)

func main() {
	g := NewGenerator()

	g.MessageVectorGroup("paych",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "create-ok",
				Version: "v1",
				Desc:    "",
			},
			Func: happyPathCreate,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "update-ok",
				Version: "v1",
				Desc:    "",
			},
			Func: happyPathUpdate,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "collect-ok",
				Version: "v1",
				Desc:    "",
			},
			Func: happyPathCollect,
		},
	)

	g.Wait()
}
