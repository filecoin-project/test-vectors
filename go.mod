module github.com/filecoin-project/test-vectors

go 1.14

require (
	github.com/AndreasBriese/bbloom v0.0.0-20190825152654-46b345b51c96 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/buger/goterm v0.0.0-20200322175922-2f3e71b85129 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/drand/drand v1.0.3-0.20200714175734-29705eaf09d4 // indirect
	github.com/filecoin-project/go-address v0.0.3
	github.com/filecoin-project/go-amt-ipld v0.0.0-20191205011053-79efc22d6cdc // indirect
	github.com/filecoin-project/go-amt-ipld/v2 v2.1.1-0.20200731171407-e559a0579161 // indirect
	github.com/filecoin-project/go-bitfield v0.2.0
	github.com/filecoin-project/go-crypto v0.0.0-20191218222705-effae4ea9f03
	github.com/filecoin-project/go-fil-markets v0.5.8 // indirect
	github.com/filecoin-project/go-jsonrpc v0.1.2-0.20200822201400-474f4fdccc52 // indirect
	github.com/filecoin-project/lotus v0.4.1
	github.com/filecoin-project/sector-storage v0.0.0-20200810171746-eac70842d8e0 // indirect
	github.com/filecoin-project/specs-actors v0.9.1-0.20200903020352-42ad3e9fbfa9
	github.com/filecoin-project/specs-storage v0.1.1-0.20200730063404-f7db367e9401 // indirect
	github.com/filecoin-project/storage-fsm v0.0.0-20200805013058-9d9ea4e6331f // indirect
	github.com/ipfs/go-bitswap v0.2.20 // indirect
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-blockservice v0.1.4-0.20200624145336-a978cec6e834
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.4
	github.com/ipfs/go-ds-badger2 v0.1.1-0.20200708190120-187fc06f714e // indirect
	github.com/ipfs/go-fs-lock v0.0.6 // indirect
	github.com/ipfs/go-hamt-ipld v0.1.1
	github.com/ipfs/go-ipfs-exchange-interface v0.0.1
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipld-cbor v0.0.5-0.20200428170625-a0bd04d3cbdf
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-merkledag v0.3.2
	github.com/ipld/go-car v0.1.1-0.20200526133713-1c7508d55aae
	github.com/lib/pq v1.7.0 // indirect
	github.com/libp2p/go-libp2p v0.11.0 // indirect
	github.com/libp2p/go-libp2p-core v0.6.1
	github.com/libp2p/go-libp2p-kad-dht v0.8.3 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.3.6-0.20200901174250-06a12f17b7de // indirect
	github.com/libp2p/go-libp2p-quic-transport v0.8.0 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/multiformats/go-multihash v0.0.14
	github.com/multiformats/go-varint v0.0.6
	github.com/raulk/clock v1.1.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/supranational/blst v0.1.1 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20200814224545-656e08ce49ee
	github.com/willscott/go-cmp v0.5.2-0.20200812183318-8affb9542345 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xorcare/golden v0.6.1-0.20191112154924-b87f686d7542 // indirect
	go.uber.org/dig v1.10.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/filecoin-project/filecoin-ffi => ./gen/extern/filecoin-ffi

replace github.com/supranational/blst => github.com/supranational/blst v0.1.2-alpha.1
