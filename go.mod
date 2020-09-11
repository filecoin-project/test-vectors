module github.com/filecoin-project/test-vectors

go 1.14

require (
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/filecoin-project/go-address v0.0.3
	github.com/filecoin-project/go-bitfield v0.2.0
	github.com/filecoin-project/go-crypto v0.0.0-20191218222705-effae4ea9f03
	github.com/filecoin-project/go-state-types v0.0.0-20200911004822-964d6c679cfc
	github.com/filecoin-project/lotus v0.7.0
	github.com/filecoin-project/specs-actors v0.9.8
	github.com/filecoin-project/test-vectors/schema v0.0.1
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-blockservice v0.1.4-0.20200624145336-a978cec6e834
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.4
	github.com/ipfs/go-hamt-ipld v0.1.1
	github.com/ipfs/go-ipfs-exchange-interface v0.0.1
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipld-cbor v0.0.5-0.20200428170625-a0bd04d3cbdf
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-merkledag v0.3.2
	github.com/ipld/go-car v0.1.1-0.20200526133713-1c7508d55aae
	github.com/libp2p/go-libp2p-core v0.6.1
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/multiformats/go-multihash v0.0.14
	github.com/multiformats/go-varint v0.0.6
	github.com/stretchr/testify v1.6.1
	github.com/whyrusleeping/cbor-gen v0.0.0-20200814224545-656e08ce49ee
	github.com/xeipuuv/gojsonschema v1.2.0
)

replace github.com/filecoin-project/filecoin-ffi => ./gen/extern/filecoin-ffi

replace github.com/supranational/blst => github.com/supranational/blst v0.1.2-alpha.1

replace github.com/filecoin-project/test-vectors/schema => ./schema
