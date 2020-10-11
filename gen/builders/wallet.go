package builders

import (
	"fmt"
	"math/rand"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet"
	"github.com/filecoin-project/lotus/lib/sigs"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-crypto"

	acrypto "github.com/filecoin-project/go-state-types/crypto"
)

type Wallet struct {
	// Private keys by address
	keys map[address.Address]*wallet.Key
	// Seed for deterministic secp key generation.
	secpSeed int64
	// Seed for deterministic bls key generation.
	blsSeed int64 // nolint: structcheck
}

func NewWallet() *Wallet {
	return &Wallet{
		keys:     make(map[address.Address]*wallet.Key),
		secpSeed: 0,
		blsSeed:  1,
	}
}

func (w *Wallet) NewSECP256k1Account() address.Address {
	secpKey := w.newSecp256k1Key()
	w.keys[secpKey.Address] = secpKey
	return secpKey.Address
}

func (w *Wallet) NewBLSAccount() address.Address {
	blsKey := w.newBLSKey()
	w.keys[blsKey.Address] = blsKey
	return blsKey.Address
}

func (w *Wallet) Sign(addr address.Address, data []byte) (*acrypto.Signature, error) {
	ki, ok := w.keys[addr]
	if !ok {
		return nil, fmt.Errorf("unknown address %v", addr)
	}

	return sigs.Sign(wallet.ActSigType(ki.Type), ki.PrivateKey, data)
}

func (w *Wallet) newSecp256k1Key() *wallet.Key {
	randSrc := rand.New(rand.NewSource(w.secpSeed))
	prv, err := crypto.GenerateKeyFromSeed(randSrc)
	if err != nil {
		panic(err)
	}
	w.secpSeed++
	key, err := wallet.NewKey(types.KeyInfo{
		Type:       types.KTSecp256k1,
		PrivateKey: prv,
	})
	if err != nil {
		panic(err)
	}
	return key
}

func (w *Wallet) newBLSKey() *wallet.Key {
	// FIXME: bls needs deterministic key generation
	//sk := ffi.PrivateKeyGenerate(s.blsSeed)
	// s.blsSeed++
	var sk [32]byte
	sk[0] = uint8(w.blsSeed) // hack to keep gas values determinist
	w.blsSeed++
	key, err := wallet.NewKey(types.KeyInfo{
		Type:       types.KTBLS,
		PrivateKey: sk[:],
	})
	if err != nil {
		panic(err)
	}
	return key
}
