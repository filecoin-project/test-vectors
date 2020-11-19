module github.com/filecoin-project/test-vectors

go 1.14

require (
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/filecoin-project/go-address v0.0.5-0.20201103152444-f2023ef3f5bb
	github.com/filecoin-project/go-bitfield v0.2.3-0.20201110211213-fe2c1862e816
	github.com/filecoin-project/go-crypto v0.0.0-20191218222705-effae4ea9f03
	github.com/filecoin-project/go-state-types v0.0.0-20201102161440-c8033295a1fc
	github.com/filecoin-project/lotus v1.1.4-0.20201117223607-62594a97a5b5
	github.com/filecoin-project/specs-actors v0.9.13
	github.com/filecoin-project/specs-actors/v2 v2.3.2
	github.com/filecoin-project/test-vectors/schema v0.0.5
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-blockservice v0.1.4
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.5
	github.com/ipfs/go-hamt-ipld v0.1.1
	github.com/ipfs/go-ipfs-exchange-interface v0.0.1
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipld-cbor v0.0.5
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-merkledag v0.3.2
	github.com/ipld/go-car v0.1.1-0.20200923150018-8cdef32e2da4
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/multiformats/go-multihash v0.0.14
	github.com/multiformats/go-varint v0.0.6
	github.com/stretchr/testify v1.6.1
	github.com/whyrusleeping/cbor-gen v0.0.0-20200826160007-0b9f6c5fb163
	github.com/xeipuuv/gojsonschema v1.2.0
)

replace github.com/filecoin-project/filecoin-ffi => ./gen/extern/filecoin-ffi

replace github.com/supranational/blst => ./gen/extern/fil-blst/blst

replace github.com/filecoin-project/fil-blst => ./gen/extern/fil-blst

replace github.com/filecoin-project/test-vectors/schema => ./schema
