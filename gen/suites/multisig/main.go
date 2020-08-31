package main

import (
	. "github.com/filecoin-project/test-vectors/gen/builders"
)

const (
	gasLimit  = 1_000_000_000
	gasFeeCap = 200
)

func main() {
	g := NewGenerator()
	defer g.Close()

	g.Group("basic",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-create",
				Version: "v1",
				Desc:    "multisig actor constructor ok",
			},
			MessageFunc: constructor,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-propose-and-cancel",
				Version: "v1",
				Desc:    "multisig actor propose and cancel ok",
			},
			MessageFunc: proposeAndCancelOk,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-propose-and-approve",
				Version: "v1",
				Desc:    "multisig actor propose, unauthorized proposals+approval, and approval ok",
			},
			MessageFunc: proposeAndApprove,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-add-signer",
				Version: "v1",
				Desc:    "multisig actor accepts only AddSigner messages that go through a reflexive flow",
			},
			MessageFunc: addSigner,
		},
	)
}
