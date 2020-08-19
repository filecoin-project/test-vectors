package builders

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"

	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/test-vectors/schema"
)

const LotusExecutionTraceV1 = "Lotus-ExecutionTrace-V1"

// EncodeTraces takes a set of serialized lotus ExecutionTraces and writes them
// to the test vector serialized diagnostic format.
func EncodeTraces(traces []types.ExecutionTrace) *schema.Diagnostics {
	d := schema.Diagnostics{Format: LotusExecutionTraceV1}
	serialized, err := json.Marshal(traces)
	if err != nil {
		panic(err)
	}

	data := bytes.NewBuffer(nil)
	formatter := base64.NewEncoder(base64.StdEncoding, data)
	compressor := gzip.NewWriter(formatter)
	_, err = compressor.Write(serialized)
	if err != nil {
		panic(err)
	}
	if err := compressor.Close(); err != nil {
		panic(err)
	}

	d.Data = data.Bytes()
	return &d
}
