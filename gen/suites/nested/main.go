package main

import (
	. "github.com/filecoin-project/test-vectors/gen/builders"
	"github.com/filecoin-project/test-vectors/schema"
)

func main() {
	g := NewGenerator()

	g.Group("nested_sends",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-basic",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_OkBasic,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-to-new-actor",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_OkToNewActor,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-to-new-actor-with-invoke",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_OkToNewActorWithInvoke,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-recursive",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_OkRecursive,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-non-cbor-params-with-transfer",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_OKNonCBORParamsWithTransfer,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-non-existent-id-address",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_FailNonexistentIDAddress,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-non-existent-actor-address",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_FailNonexistentActorAddress,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-invalid-method-num-new-actor",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_FailInvalidMethodNumNewActor,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-invalid-method-num-for-actor",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_FailInvalidMethodNumForActor,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-missing-params",
				Version: "v1",
				Desc:    "",
				Comment: "nested message exit code should be ErrSerialization see https://github.com/filecoin-project/test-vectors/issues/93#issuecomment-689593946",
			},
			MessageFunc: nestedSends_FailMissingParams,
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-mismatch-params",
				Version: "v1",
				Desc:    "",
				Comment: "nested message exit code should be ErrSerialization see https://github.com/filecoin-project/test-vectors/issues/93#issuecomment-689593946",
			},
			MessageFunc: nestedSends_FailMismatchParams,
			Mode:        ModeLenientAssertions,
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-inner-abort",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_FailInnerAbort,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-aborted-exec",
				Version: "v1",
				Desc:    "",
			},
			MessageFunc: nestedSends_FailAbortedExec,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-insufficient-funds-for-transfer-in-inner-send",
				Version: "v1",
				Desc:    "",
			},
			Selector:    map[string]string{"chaos_actor": "true"},
			MessageFunc: nestedSends_FailInsufficientFundsForTransferInInnerSend,
		},
	)

	g.Close()
}
