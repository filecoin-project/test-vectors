package builders

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/types"
	lotus "github.com/filecoin-project/lotus/conformance"
	"github.com/filecoin-project/test-vectors/schema"
)

// StateTracker is an object for tracking state and mutating it by applying
// messages.
type StateTracker struct {
	bc *BuilderCommon

	Stores    *Stores
	StateTree *state.StateTree
	Driver    *lotus.Driver

	CurrRoot cid.Cid
}

func NewStateTracker(bc *BuilderCommon, selector schema.Selector) *StateTracker {
	stores := NewLocalStores(context.Background())

	// create a brand new state tree.
	st, err := state.NewStateTree(stores.CBORStore)
	if err != nil {
		panic(err)
	}

	stkr := &StateTracker{
		bc:        bc,
		Stores:    stores,
		StateTree: st,
		Driver:    lotus.NewDriver(context.Background(), selector),
	}

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
		bc:        st.bc,
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
	var err error
	am.Result, st.CurrRoot, err = st.Driver.ExecuteMessage(st.Stores.Blockstore, st.CurrRoot, am.Epoch, am.Message)
	st.bc.Assert.NoError(err)

	// replace the state tree.
	st.StateTree, err = state.LoadStateTree(st.Stores.CBORStore, st.CurrRoot)
	st.bc.Assert.NoError(err)
}

// ActorState retrieves the state of the supplied actor, and sets it in the
// provided object. It also returns the actor's header from the state tree.
func (st *StateTracker) ActorState(addr address.Address, out cbg.CBORUnmarshaler) *types.Actor {
	actor := st.Header(addr)
	err := st.StateTree.Store.Get(context.Background(), actor.Head, out)
	st.bc.Assert.NoError(err, "failed to load state for actorr %s; head=%s", addr, actor.Head)
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
