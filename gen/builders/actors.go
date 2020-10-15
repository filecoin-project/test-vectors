package builders

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/cbor"

	"github.com/filecoin-project/lotus/chain/actors"

	builtin0 "github.com/filecoin-project/specs-actors/actors/builtin"
	account0 "github.com/filecoin-project/specs-actors/actors/builtin/account"
	miner0 "github.com/filecoin-project/specs-actors/actors/builtin/miner"
	power0 "github.com/filecoin-project/specs-actors/actors/builtin/power"

	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	account2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/account"
	miner2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	power2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/power"

	adt0 "github.com/filecoin-project/specs-actors/actors/util/adt"
	adt2 "github.com/filecoin-project/specs-actors/v2/actors/util/adt"
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

	var state cbor.Marshaler
	var handle AddressHandle
	switch a.st.ActorsVersion {
	case actors.Version0:
		state = &account0.State{Address: addr}
		handle = a.st.CreateActor(builtin0.AccountActorCodeID, addr, balance, state)
	case actors.Version2:
		state = &account2.State{Address: addr}
		handle = a.st.CreateActor(builtin2.AccountActorCodeID, addr, balance, state)
	default:
		panic("unknown actors version")
	}

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
	var (
		owner  = a.Account(address.SECP256K1, cfg.OwnerBalance)
		worker = a.Account(address.BLS, big.Zero())

		minerInfo cbor.Marshaler
		state     cbor.Marshaler
		err       error
	)

	switch a.st.ActorsVersion {
	case actors.Version0:
		minerInfo, err = miner0.ConstructMinerInfo(owner.ID, worker.ID, nil, []byte("test"), nil, cfg.SealProofType)
	case actors.Version2:
		minerInfo, err = miner2.ConstructMinerInfo(owner.ID, worker.ID, nil, []byte("test"), nil, cfg.SealProofType)
	default:
		panic("unknown actors version")
	}

	a.bc.Assert.NoError(err, "failed to construct miner info")

	infoCid, err := a.st.Stores.CBORStore.Put(context.Background(), minerInfo)
	if err != nil {
		panic(err)
	}

	// TODO allow an address to create multiple miners.
	minerActorAddr := worker.NextActorAddress(0, 0)
	var minerHandle AddressHandle
	switch a.st.ActorsVersion {
	case actors.Version0:
		state, err = miner0.ConstructState(infoCid,
			cfg.PeriodBoundary,
			a.st.EmptyBitfieldCid,
			a.st.EmptyArrayCid,
			a.st.EmptyMapCid,
			a.st.EmptyDeadlinesCid,
			a.st.EmptyVestingFundsCid,
		)
		a.bc.Assert.NoError(err)
		minerHandle = a.st.CreateActor(builtin0.StorageMinerActorCodeID, minerActorAddr, big.Zero(), state)

		// next update the storage power actor to track the miner
		var spa power0.State
		a.st.ActorState(builtin0.StoragePowerActorAddr, &spa)

		// set the miners claim
		hm, err := adt0.AsMap(a.st.Stores.ADTStore, spa.Claims)
		a.bc.Assert.NoError(err)

		// add claim for the miner TODO: allow caller to specify.
		err = hm.Put(abi.AddrKey(minerHandle.ID), &power0.Claim{
			RawBytePower:    abi.NewStoragePower(0),
			QualityAdjPower: abi.NewStoragePower(0),
		})

		// save the claim, update miner count
		spa.Claims, err = hm.Root()
		spa.MinerCount += 1

		// update storage power actor's state in the tree
		spaCid, err := a.st.Stores.CBORStore.Put(context.Background(), &spa)
		a.bc.Assert.NoError(err)

		// update spa header.
		spaHeader := a.st.Header(builtin0.StoragePowerActorAddr)
		spaHeader.Head = spaCid
		err = a.st.StateTree.SetActor(builtin0.StoragePowerActorAddr, spaHeader)
		a.bc.Assert.NoError(err)

	case actors.Version2:
		state, err = miner2.ConstructState(infoCid,
			cfg.PeriodBoundary,
			0, // different
			a.st.EmptyBitfieldCid,
			a.st.EmptyArrayCid,
			a.st.EmptyMapCid,
			a.st.EmptyDeadlinesCid,
			a.st.EmptyVestingFundsCid,
		)
		a.bc.Assert.NoError(err)

		minerHandle = a.st.CreateActor(builtin2.StorageMinerActorCodeID, minerActorAddr, big.Zero(), state)

		// next update the storage power actor to track the miner
		var spa power2.State
		a.st.ActorState(builtin2.StoragePowerActorAddr, &spa)

		// set the miners claim
		hm, err := adt2.AsMap(a.st.Stores.ADTStore, spa.Claims)
		a.bc.Assert.NoError(err)

		// add claim for the miner TODO: allow caller to specify.
		err = hm.Put(abi.AddrKey(minerHandle.ID), &power2.Claim{
			RawBytePower:    abi.NewStoragePower(0),
			QualityAdjPower: abi.NewStoragePower(0),
		})

		// save the claim, update miner count
		spa.Claims, err = hm.Root()
		spa.MinerCount += 1

		// update storage power actor's state in the tree
		spaCid, err := a.st.Stores.CBORStore.Put(context.Background(), &spa)
		a.bc.Assert.NoError(err)

		// update spa header.
		spaHeader := a.st.Header(builtin2.StoragePowerActorAddr)
		spaHeader.Head = spaCid
		err = a.st.StateTree.SetActor(builtin2.StoragePowerActorAddr, spaHeader)
		a.bc.Assert.NoError(err)

	default:
		panic("unknown actors version")
	}

	m := Miner{
		MinerActorAddr: minerHandle,
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
