package builders

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"path"
	"time"

	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/test-vectors/schema"
)

const LotusExecutionTraceV1 = "Lotus-ExecutionTrace-V1"

// EncodeTraces takes a set of serialized lotus ExecutionTraces and writes them
// to the test vector serialized diagnostic format.
func EncodeTraces(traces []types.ExecutionTrace) *schema.Diagnostics {
	if len(traces) == 0 {
		return nil
	}

	d := schema.Diagnostics{Format: LotusExecutionTraceV1}
	serialized, err := json.Marshal(cleanTraces(traces))
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

// cleanTraces recursively strips variable/volatile fields from execution traces,
// e.g. TimeTaken, in order to remove noise and facilitate comparison and diffing.
func cleanTraces(t []types.ExecutionTrace) []types.ExecutionTrace {
	for i := range t {
		t[i].Duration = time.Duration(0)
		t[i].Subcalls = cleanTraces(t[i].Subcalls)
		for j := range t[i].GasCharges {
			for k := range t[i].GasCharges[j].Location {
				_, file := path.Split(t[i].GasCharges[j].Location[k].File)
				t[i].GasCharges[j].Location[k].File = file
			}
			t[i].GasCharges[j].TimeTaken = time.Duration(0)
			t[i].GasCharges[j].Callers = nil
		}
	}
	return t
}
