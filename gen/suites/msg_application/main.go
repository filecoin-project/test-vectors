package main

import (
	"github.com/filecoin-project/go-state-types/abi"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

var (
	unknown      = MustNewIDAddr(10000000)
	balance1T    = abi.NewTokenAmount(1_000_000_000_000)
	transferAmnt = abi.NewTokenAmount(10)
)

func main() {
	g := NewGenerator()

	g.Group("gas_cost",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-ok-secp-bls-gas-costs",
				Version: "v1",
				Desc:    "check the gas cost of secpk (higher) and BLS (lower) messages",
			},
			TipsetFunc: okSecpkBLSCosts,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-apply-fail-receipt-gas",
				Version: "v1",
				Desc:    "fail to cover gas cost for message receipt on chain",
			},
			MessageFunc: failCoverReceiptGasCost,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-apply-fail-onchainsize-gas",
				Version: "v1",
				Desc:    "not enough gas to pay message on-chain-size cost",
			},
			MessageFunc: failCoverOnChainSizeGasCost,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-apply-fail-transfer-accountcreation-gas",
				Version: "v1",
				Desc:    "fail not enough gas to cover account actor creation on transfer",
			},
			MessageFunc: failCoverTransferAccountCreationGasStepwise,
		},
	)

	g.Group("invalid_msgs",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-apply-fail-invalid-nonce",
				Version: "v1",
				Desc:    "invalid actor nonce",
			},
			MessageFunc: failInvalidActorNonce,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-apply-fail-invalid-receiver-method",
				Version: "v1",
				Desc:    "invalid receiver method",
			},
			MessageFunc: failInvalidReceiverMethod,
		},
	)

	g.Group("unknown_actors",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-apply-fail-unknown-sender",
				Version: "v1",
				Desc:    "fail due to lack of gas when sender is unknown",
			},
			MessageFunc: failUnknownSender,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-apply-fail-unknown-receiver",
				Version: "v1",
				Desc:    "inexistent receiver",
				Comment: `Note that this test is not a valid message, since it is using
an unknown actor. However in the event that an invalid message isn't filtered by
block validation we need to ensure behaviour is consistent across VM implementations.`,
			},
			MessageFunc: failUnknownReceiver,
		},
	)

	g.Group("actor_exec",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "msg-apply-fail-actor-execution-illegal-arg",
				Version: "v1",
				Desc:    "abort during actor execution due to illegal argument",
			},
			MessageFunc: failActorExecutionAborted,
		},
	)

	g.Group("duplicates",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "messages-deduplicated",
				Version: "v1",
				Desc:    "duplicated messages in a block are deduplicated",
			},
			TipsetFunc: minerIncludesDuplicateMessages,
		},
	)

	g.Close()
}
