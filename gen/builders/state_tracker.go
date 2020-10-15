package builders

import (
	"context"
	"fmt"
	"log"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	miner0 "github.com/filecoin-project/specs-actors/actors/builtin/miner"
	adt0 "github.com/filecoin-project/specs-actors/actors/util/adt"
	miner2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	adt2 "github.com/filecoin-project/specs-actors/v2/actors/util/adt"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/test-vectors/schema"

	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/conformance"

	"github.com/filecoin-project/go-state-types/cbor"
)

// StateTracker is an object for tracking state and mutating it by applying
// messages.
type StateTracker struct {
	vector *schema.TestVector
	bc     *BuilderCommon

	StateTreeVersion types.StateTreeVersion
	ActorsVersion    actors.Version

	Stores    *Stores
	StateTree *state.StateTree
	Driver    *conformance.Driver

	CurrRoot cid.Cid

	EmptyObjectCid       cid.Cid
	EmptyArrayCid        cid.Cid
	EmptyMapCid          cid.Cid
	EmptyMultiMapCid     cid.Cid
	EmptyBitfieldCid     cid.Cid
	EmptyDeadlinesCid    cid.Cid
	EmptyVestingFundsCid cid.Cid
}

func NewStateTracker(bc *BuilderCommon, selector schema.Selector, vector *schema.TestVector, stVersion types.StateTreeVersion, actorsVersion actors.Version, zeroFn ActorsZeroStateFn) *StateTracker {
	stores := NewLocalStores(context.Background())

	// create a brand new state tree.
	st, err := state.NewStateTree(stores.CBORStore, stVersion)
	if err != nil {
		panic(err)
	}

	stkr := &StateTracker{
		vector:           vector,
		bc:               bc,
		StateTreeVersion: stVersion,
		ActorsVersion:    actorsVersion,
		Stores:           stores,
		StateTree:        st,
		Driver:           conformance.NewDriver(context.Background(), selector, conformance.DriverOpts{}),
	}

	stkr.initEmptyStructures()

	zeroFn(stkr, selector)

	_ = stkr.Flush()
	return stkr
}

func (st *StateTracker) initEmptyStructures() {
	// empty object
	st.EmptyObjectCid = func() cid.Cid {
		empty, err := st.Stores.ADTStore.Put(context.TODO(), []struct{}{})
		if err != nil {
			panic(err)
		}
		return empty
	}()

	// bitfield -- same no matter actors version.
	st.EmptyBitfieldCid = func() cid.Cid {
		empty, err := st.Stores.ADTStore.Put(context.TODO(), bitfield.NewFromSet(nil))
		if err != nil {
			panic(err)
		}
		return empty
	}()

	// map -- lotus abstraction for version selection EXISTS.
	st.EmptyMapCid = func() cid.Cid {
		empty, err := adt.NewMap(st.Stores.ADTStore, st.ActorsVersion)
		if err != nil {
			panic(err)
		}
		ret, err := empty.Root()
		if err != nil {
			panic(err)
		}
		return ret
	}()

	// array -- lotus abstraction for version selection EXISTS.
	st.EmptyArrayCid = func() cid.Cid {
		empty, err := adt.NewArray(st.Stores.ADTStore, st.ActorsVersion)
		if err != nil {
			panic(err)
		}
		ret, err := empty.Root()
		if err != nil {
			panic(err)
		}
		return ret
	}()

	// multimap -- lotus abstraction for version selection does NOT exist.
	st.EmptyMultiMapCid = func() (ret cid.Cid) {
		var err error
		switch st.ActorsVersion {
		case actors.Version0:
			ret, err = adt0.MakeEmptyMultimap(st.Stores.ADTStore).Root()
			if err != nil {
				panic(err)
			}
		case actors.Version2:
			ret, err = adt2.MakeEmptyMultimap(st.Stores.ADTStore).Root()
			if err != nil {
				panic(err)
			}
		default:
			panic("unknown actors version")
		}
		return ret
	}()

	st.EmptyDeadlinesCid = func() (ret cid.Cid) {
		var obj cbor.Marshaler
		switch st.ActorsVersion {
		case actors.Version0:
			obj = miner0.ConstructDeadline(st.EmptyArrayCid)
		case actors.Version2:
			obj = miner2.ConstructDeadline(st.EmptyArrayCid)
		default:
			panic("unknown actors version")
		}
		ret, err := st.Stores.ADTStore.Put(context.TODO(), obj)
		if err != nil {
			panic(err)
		}
		return ret
	}()

	st.EmptyVestingFundsCid = func() (ret cid.Cid) {
		var obj cbor.Marshaler
		switch st.ActorsVersion {
		case actors.Version0:
			obj = miner0.ConstructVestingFunds()
		case actors.Version2:
			obj = miner2.ConstructVestingFunds()
		default:
			panic("unknown actors version")
		}
		ret, err := st.Stores.ADTStore.Put(context.TODO(), obj)
		if err != nil {
			panic(err)
		}
		return ret
	}()

}

// Fork forks this state tracker into a new one, using the provided cid.Cid as
// the root CID.
func (st *StateTracker) Fork(root cid.Cid) *StateTracker {
	tree, err := state.LoadStateTree(st.Stores.CBORStore, root)
	if err != nil {
		panic(err)
	}

	cpy := *st
	cpy.StateTree = tree
	cpy.CurrRoot = root
	return &cpy
}

// Load sets the state tree to the one indicated by this root.
func (st *StateTracker) Load(root cid.Cid) {
	tree, err := state.LoadStateTree(st.Stores.CBORStore, root)
	if err != nil {
		panic(err)
	}
	st.StateTree = tree
}

// Flush calls Flush on the StateTree, and records the CurrRoot.
func (st *StateTracker) Flush() cid.Cid {
	root, err := st.StateTree.Flush(context.Background())
	if err != nil {
		panic(err)
	}
	st.CurrRoot = root
	return root
}

// ApplyMessage executes the provided message via the driver, records the new
// root, refreshes the state tree, and updates the underlying vector with the
// message and its receipt.
func (st *StateTracker) ApplyMessage(am *ApplicableMessage) {
	var postRoot cid.Cid
	var err error

	am.baseFee = conformance.BaseFeeOrDefault(st.vector.Pre.BaseFee)
	am.Applied = true
	am.Result, postRoot, err = st.Driver.ExecuteMessage(st.Stores.Blockstore, conformance.ExecuteMessageParams{
		Preroot:    st.CurrRoot,
		Epoch:      st.bc.ProtocolVersion.FirstEpoch + am.EpochOffset,
		Message:    am.Message,
		BaseFee:    conformance.BaseFeeOrDefault(st.vector.Pre.BaseFee),
		CircSupply: conformance.CircSupplyOrDefault(st.vector.Pre.CircSupply),
	})
	if err != nil {
		am.Failed = true
		return
	}

	st.CurrRoot = postRoot
	// replace the state tree.
	st.StateTree, err = state.LoadStateTree(st.Stores.CBORStore, st.CurrRoot)
	if err != nil {
		panic(fmt.Sprintf("failed reload state tree after applying message: %s", err))
	}
}

// CreateActor creates an actor in the state tree, of the specified kind, with
// the specified address and balance, and sets its state to the supplied state.
func (st *StateTracker) CreateActor(code cid.Cid, addr address.Address, balance abi.TokenAmount, state cbor.Marshaler) AddressHandle {
	var id address.Address
	if addr.Protocol() != address.ID {
		var err error
		id, err = st.StateTree.RegisterNewAddress(addr)
		if err != nil {
			log.Panicf("register new address for actor: %v", err)
		}
	}

	// Store the new state.
	head, err := st.StateTree.Store.Put(context.Background(), state)
	if err != nil {
		panic(err)
	}

	// Set the actor's head to point to that state.
	actr := &types.Actor{
		Code:    code,
		Head:    head,
		Balance: balance,
	}
	if err := st.StateTree.SetActor(addr, actr); err != nil {
		panic(fmt.Sprintf("setting new actor for actor: %v", err))
	}

	return AddressHandle{id, addr}
}

// ActorState retrieves the state of the supplied actor, and sets it in the
// provided object. It also returns the actor's header from the state tree.
func (st *StateTracker) ActorState(addr address.Address, out cbg.CBORUnmarshaler) *types.Actor {
	actor := st.Header(addr)
	err := st.StateTree.Store.Get(context.Background(), actor.Head, out)
	st.bc.Assert.NoError(err, "failed to load state for actor %s (head=%s)", addr, actor.Head)
	return actor
}

// Header returns the actor's header from the state tree.
func (st *StateTracker) Header(addr address.Address) *types.Actor {
	actor, err := st.StateTree.GetActor(addr)
	st.bc.Assert.NoError(err, "failed to fetch actor %s from state", addr)
	return actor
}

// Balance is a shortcut for Header(addr).Balance.
func (st *StateTracker) Balance(addr address.Address) abi.TokenAmount {
	return st.Header(addr).Balance
}

// Head is a shortcut for Header(addr).Head.
func (st *StateTracker) Head(addr address.Address) cid.Cid {
	return st.Header(addr).Head
}

// Nonce is a shortcut for Header(addr).Nonce.
func (st *StateTracker) Nonce(addr address.Address) uint64 {
	return st.Header(addr).Nonce
}

// Code is a shortcut for Header(addr).Code.
func (st *StateTracker) Code(addr address.Address) cid.Cid {
	return st.Header(addr).Code
}
