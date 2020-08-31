package builders

import (
	"bytes"
	"compress/gzip"
	"context"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-car"
)

// EncodeCAR recursively writes the tree referenced by the root in CAR form
// into a gzipped byte buffer, and returns its bytes, ready for embedding in a
// test vector.
func EncodeCAR(dagserv format.DAGService, roots ...cid.Cid) ([]byte, error) {
	carWalkFn := func(nd format.Node) (out []*format.Link, err error) {
		for _, link := range nd.Links() {
			if link.Cid.Prefix().Codec == cid.FilCommitmentSealed || link.Cid.Prefix().Codec == cid.FilCommitmentUnsealed {
				continue
			}
			out = append(out, link)
		}
		return out, nil
	}

	var (
		out = new(bytes.Buffer)
		gw  = gzip.NewWriter(out)
	)

	if err := car.WriteCarWithWalker(context.Background(), dagserv, roots, gw, carWalkFn); err != nil {
		return nil, err
	}
	if err := gw.Flush(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
