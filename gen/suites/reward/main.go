package main

import (
	"github.com/filecoin-project/specs-actors/actors/abi"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

var (
	balance = abi.NewTokenAmount(1_000_000_000_000_000)
)

func main() {
	g := NewGenerator()
	defer g.Close()

	g.Group("reward",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-miners-awarded-no-premiums",
				Version: "v1",
				Desc:    "verifies that miners are awarded for the mining of blocks; no premiums",
			},
			TipsetFunc: minersAwardedNoPremiums,
		},
	)
}
