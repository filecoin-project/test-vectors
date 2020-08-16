package main

import (
	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func main() {
	g := NewGenerator()

	g.MessageVectorGroup("nested_sends",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "ok-basic",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_OkBasic,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "ok-to-new-actor",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_OkToNewActor,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "ok-to-new-actor-with-invoke",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_OkToNewActorWithInvoke,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "ok-recursive",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_OkRecursive,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "ok-non-cbor-params-with-transfer",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_OKNonCBORParamsWithTransfer,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-non-existent-id-address",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailNonexistentIDAddress,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-non-existent-actor-address",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailNonexistentActorAddress,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-invalid-method-num-new-actor",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailInvalidMethodNumNewActor,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-invalid-method-num-for-actor",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailInvalidMethodNumForActor,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-missing-params",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailMissingParams,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-mismatch-params",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailMismatchParams,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-inner-abort",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailInnerAbort,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-aborted-exec",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailAbortedExec,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-insufficient-funds-for-transfer-in-inner-send",
				Version: "v1",
				Desc:    "",
			},
			Func: nestedSends_FailInsufficientFundsForTransferInInnerSend,
		},
	)

	g.Wait()
}
