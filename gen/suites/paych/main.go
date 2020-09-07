package main

import (
	"github.com/filecoin-project/go-state-types/abi"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

var (
	initialBal = abi.NewTokenAmount(1_000_000_000_000)
	toSend     = abi.NewTokenAmount(10_000)
)

func main() {
	g := NewGenerator()

	g.Group("paych",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "create-ok",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: happyPathCreate,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "update-ok",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: happyPathUpdate,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "collect-ok",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: happyPathCollect,
		},
	)

	g.Close()
}
