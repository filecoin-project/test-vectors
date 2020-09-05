package builders

import (
	"context"
	"log"

	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/account"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/actors/builtin/power"
	"github.com/filecoin-project/specs-actors/actors/runtime"
	"github.com/filecoin-project/specs-actors/actors/util/adt"

	"github.com/filecoin-project/go-address"

	"github.com/ipfs/go-cid"
)

type Account struct {
	Handle  AddressHandle
	Initial abi.TokenAmount
}

type Miner struct {
	MinerActorAddr, OwnerAddr, WorkerAddr AddressHandle
}

// Actors is an object that manages actors in the test vector.
type Actors struct {
	accounts []Account
	miners   []Miner

	bc *BuilderCommon
	st *StateTracker
}

func NewActors(bc *BuilderCommon, b *StateTracker) *Actors {
	return &Actors{bc: bc, st: b}
}

// Accounts returns all accounts that have been registered through
// Account() / AccountN().
//
// Miner owner and worker accounts, despite being accounts in the strict sense
// of the word, are not returned here. You can get them through Miners().
//
// Similarly, account actors registered through CreateActor() in a bare form are
// not returned here.
func (a *Actors) Accounts() []Account {
	return a.accounts
}

// Miners returns all miners that have been registered through
// Miner() / MinerN(), along with their owner and worker account addresses.
//
// Miners registered through CreateActor() in a bare form are not returned here.
func (a *Actors) Miners() []Miner {
	return a.miners
}

// Count returns the number of accounts and miners registered.
func (a *Actors) Count() int {
	return len(a.accounts) + len(a.miners)
}

// HandleFor gets the canonical handle for a registered address, which can
// appear at either ID or Robust position.
func (a *Actors) HandleFor(addr address.Address) AddressHandle {
	for _, r := range a.accounts {
		if r.Handle.ID == addr || r.Handle.Robust == addr {
			return r.Handle
		}
	}
	for _, r := range a.miners {
		if r.MinerActorAddr.ID == addr || r.MinerActorAddr.Robust == addr {
			return r.MinerActorAddr
		}
		if r.OwnerAddr.ID == addr || r.OwnerAddr.Robust == addr {
			return r.OwnerAddr
		}
		if r.WorkerAddr.ID == addr || r.WorkerAddr.Robust == addr {
			return r.WorkerAddr
		}
	}
	a.bc.Assert.FailNowf("asked for handle of unknown actor", "actor: %s", addr)
	return AddressHandle{} // will never reach here.
}

// InitialBalance returns the initial balance of an account actor that was
// registered during preconditions. It matches against both the ID and Robust
// addresses. It records an assertion failure if the actor is unknown.
func (a *Actors) InitialBalance(addr address.Address) abi.TokenAmount {
	for _, r := range a.accounts {
		if r.Handle.ID == addr || r.Handle.Robust == addr {
			return r.Initial
		}
	}
	a.bc.Assert.FailNowf("asked for initial balance of unknown actor", "actor: %s", addr)
	return big.Zero() // will never reach here.
}

// AccountHandles returns the AddressHandles for all registered accounts.
func (a *Actors) AccountHandles() []AddressHandle {
	ret := make([]AddressHandle, 0, len(a.accounts))
	for _, r := range a.accounts {
		ret = append(ret, r.Handle)
	}
	return ret
}

// AccountN creates many account actors of the specified kind, with the
// specified balance, and places their addresses in the supplied AddressHandles.
func (a *Actors) AccountN(typ address.Protocol, balance abi.TokenAmount, handles ...*AddressHandle) {
	for _, handle := range handles {
		h := a.Account(typ, balance)
		*handle = h
	}
}

// Account creates a single account actor of the specified kind, with the
// specified balance, and returns its AddressHandle.
func (a *Actors) Account(typ address.Protocol, balance abi.TokenAmount) AddressHandle {
	a.bc.Assert.In(typ, address.SECP256K1, address.BLS)

	var addr address.Address
	switch typ {
	case address.SECP256K1:
		addr = a.bc.Wallet.NewSECP256k1Account()
	case address.BLS:
		addr = a.bc.Wallet.NewBLSAccount()
	}

	actorState := &account.State{Address: addr}
	handle := a.CreateActor(builtin.AccountActorCodeID, addr, balance, actorState)

	a.accounts = append(a.accounts, Account{handle, balance})
	return handle
}

type MinerActorCfg struct {
	SealProofType  abi.RegisteredSealProof
	PeriodBoundary abi.ChainEpoch
	OwnerBalance   abi.TokenAmount
}

// Miner creates an owner account, a worker account, and a miner actor with the
// supplied configuration.
func (a *Actors) Miner(cfg MinerActorCfg) Miner {
	owner := a.Account(address.SECP256K1, cfg.OwnerBalance)
	worker := a.Account(address.BLS, big.Zero())

	ss, err := cfg.SealProofType.SectorSize()
	a.bc.Assert.NoError(err, "seal proof sector size")

	ps, err := cfg.SealProofType.WindowPoStPartitionSectors()
	a.bc.Assert.NoError(err, "seal proof window PoSt partition sectors")

	mi := &miner.MinerInfo{
		Owner:                      owner.ID,
		Worker:                     worker.ID,
		PendingWorkerKey:           nil,
		PeerId:                     abi.PeerID("test"),
		Multiaddrs:                 nil,
		SealProofType:              cfg.SealProofType,
		SectorSize:                 ss,
		WindowPoStPartitionSectors: ps,
	}
	infoCid, err := a.st.Stores.CBORStore.Put(context.Background(), mi)
	if err != nil {
		panic(err)
	}

	// create the miner actor s.t. it exists in the init actors map
	minerState, err := miner.ConstructState(infoCid,
		cfg.PeriodBoundary,
		EmptyBitfieldCid,
		EmptyArrayCid,
		EmptyMapCid,
		EmptyDeadlinesCid,
		EmptyVestingFundsCid,
	)
	if err != nil {
		panic(err)
	}

	// TODO allow an address to create multiple miners.
	minerActorAddr := worker.NextActorAddress(0, 0)
	handle := a.CreateActor(builtin.StorageMinerActorCodeID, minerActorAddr, big.Zero(), minerState)

	// assert miner actor has been created, exists in the state tree, and has an entry in the init actor.
	// next update the storage power actor to track the miner

	var spa power.State
	a.st.ActorState(builtin.StoragePowerActorAddr, &spa)

	// set the miners claim
	hm, err := adt.AsMap(adt.WrapStore(context.Background(), a.st.Stores.CBORStore), spa.Claims)
	if err != nil {
		panic(err)
	}

	// add claim for the miner
	// TODO: allow caller to specify.
	err = hm.Put(adt.AddrKey(handle.ID), &power.Claim{
		RawBytePower:    abi.NewStoragePower(0),
		QualityAdjPower: abi.NewStoragePower(0),
	})
	if err != nil {
		panic(err)
	}

	// save the claim
	spa.Claims, err = hm.Root()
	if err != nil {
		panic(err)
	}

	// update miner count
	spa.MinerCount += 1

	// update storage power actor's state in the tree
	_, err = a.st.Stores.CBORStore.Put(context.Background(), &spa)
	if err != nil {
		panic(err)
	}

	m := Miner{
		MinerActorAddr: handle,
		OwnerAddr:      owner,
		WorkerAddr:     worker,
	}
	a.miners = append(a.miners, m)
	return m
}

// MinerN creates many miners with the specified configuration, and places the
// miner objects in the supplied addresses.
//
// It is sugar for calling Miner repeatedly with the same configuration, and
// storing the returned miners in the provided addresses.
func (a *Actors) MinerN(cfg MinerActorCfg, miners ...*Miner) {
	for _, m := range miners {
		*m = a.Miner(cfg)
	}
}

// CreateActor creates an actor in the state tree, of the specified kind, with
// the specified address and balance, and sets its state to the supplied state.
func (a *Actors) CreateActor(code cid.Cid, addr address.Address, balance abi.TokenAmount, state runtime.CBORMarshaler) AddressHandle {
	var id address.Address
	if addr.Protocol() != address.ID {
		var err error
		id, err = a.st.StateTree.RegisterNewAddress(addr)
		if err != nil {
			log.Panicf("register new address for actor: %v", err)
		}
	}

	// Store the new state.
	head, err := a.st.StateTree.Store.Put(context.Background(), state)
	if err != nil {
		panic(err)
	}

	// Set the actor's head to point to that state.
	actr := &types.Actor{
		Code:    code,
		Head:    head,
		Balance: balance,
	}
	if err := a.st.StateTree.SetActor(addr, actr); err != nil {
		log.Panicf("setting new actor for actor: %v", err)
	}

	return AddressHandle{id, addr}
}
