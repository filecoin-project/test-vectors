package builders

import (
	"context"
	"fmt"
	"log"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
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

	Stores    *Stores
	StateTree *state.StateTree
	Driver    *conformance.Driver

	CurrRoot cid.Cid
}

func NewStateTracker(selector schema.Selector, vector *schema.TestVector) *StateTracker {
	stores := NewLocalStores(context.Background())

	// create a brand new state tree.
	// TODO: specify network version in vectors.
	st, err := state.NewStateTree(stores.CBORStore, types.StateTreeVersion0)
	if err != nil {
		panic(err)
	}

	stkr := &StateTracker{
		vector:    vector,
		Stores:    stores,
		StateTree: st,
		Driver:    conformance.NewDriver(context.Background(), selector, conformance.DriverOpts{}),
	}

	stkr.initializeZeroState(selector)

	_ = stkr.Flush()
	return stkr
}

// Fork forks this state tracker into a new one, using the provided cid.Cid as
// the root CID.
func (st *StateTracker) Fork(root cid.Cid) *StateTracker {
	tree, err := state.LoadStateTree(st.Stores.CBORStore, root)
	if err != nil {
		panic(err)
	}

	return &StateTracker{
		Stores:    st.Stores,
		StateTree: tree,
		Driver:    st.Driver,
		CurrRoot:  root,
	}
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
		Epoch:      am.Epoch,
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
		log.Panicf("setting new actor for actor: %v", err)
	}

	return AddressHandle{id, addr}
}

// ActorState retrieves the state of the supplied actor, and sets it in the
// provided object. It also returns the actor's header from the state tree.
func (st *StateTracker) ActorState(addr address.Address, out cbg.CBORUnmarshaler) *types.Actor {
	actor := st.Header(addr)
	err := st.StateTree.Store.Get(context.Background(), actor.Head, out)
	if err != nil {
		panic(fmt.Sprintf("failed to load state for actor %s (head=%s): %s", addr, actor.Head, err))
	}
	return actor
}

// Header returns the actor's header from the state tree.
func (st *StateTracker) Header(addr address.Address) *types.Actor {
	actor, err := st.StateTree.GetActor(addr)
	if err != nil {
		panic(fmt.Sprintf("failed to fetch actor %s from state: %s", addr, err))
	}
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
