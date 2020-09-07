package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/exitcode"

	. "github.com/filecoin-project/test-vectors/gen/builders"
	"github.com/filecoin-project/test-vectors/schema"
)

func main() {
	g := NewGenerator()
	defer g.Close()

	g.Group("addresses",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "sequential-10",
				Version: "v1",
				Desc:    "actor addresses are sequential",
			},
			MessageFunc: sequentialAddresses,
		},
	)

	g.Group("on_transfer",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-create-secp256k1",
				Version: "v1",
			},
			MessageFunc: actorCreationOnTransfer(actorCreationOnTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(1_000_000_000_000_000),
				receiverAddr: MustNewSECP256K1Addr("publickeyfoo"),
				amount:       abi.NewTokenAmount(10_000),
				expectedCode: exitcode.Ok,
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-create-bls",
				Version: "v1",
			},
			MessageFunc: actorCreationOnTransfer(actorCreationOnTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(1_000_000_000_000_000),
				receiverAddr: MustNewBLSAddr(1),
				amount:       abi.NewTokenAmount(10_000),
				expectedCode: exitcode.Ok,
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-secp256k1-insufficient-balance",
				Version: "v1",
			},
			MessageFunc: actorCreationOnTransfer(actorCreationOnTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(9_999),
				receiverAddr: MustNewSECP256K1Addr("publickeyfoo"),
				amount:       abi.NewTokenAmount(10_000),
				expectedCode: exitcode.SysErrSenderStateInvalid,
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-bls-insufficient-balance",
				Version: "v1",
			},
			MessageFunc: actorCreationOnTransfer(actorCreationOnTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(9_999),
				receiverAddr: MustNewBLSAddr(1),
				amount:       abi.NewTokenAmount(10_000),
				expectedCode: exitcode.SysErrSenderStateInvalid,
			}),
		},
	)

	g.Group("params",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-unparsable-init-actor-exec-msg",
				Version: "v1",
				Desc:    "verifies that actor creation fails and gas is deducted when passing unparsable init exec message",
				Comment: "this should not return SysErrSenderInvalid; it should return something else, likely an SysErrSerialization",
			},
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: createActorInitExecUnparsableParams,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fails-unparsable-constructor-params-via-init-actor",
				Version: "v1",
				Desc:    "verifies that actor creation fails and gas is deducted when passing unparsable constructor params via init actor",
				Comment: "this should not return SysErrSenderInvalid; it should return something else, likely an ErrSerialization because the error is in actor space",
			},
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: createActorCtorUnparsableParamsViaInitExec,
		},
	)
}
