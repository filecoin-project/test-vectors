package builders

import (
	"fmt"
	"os"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/lotus/chain/state"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

// Asserter offers useful assertions to verify outcomes at various stages of
// the test vector creation.
type Asserter struct {
	*require.Assertions

	// id is the vector ID, for logging purposes.
	id string

	pv ProtocolVersion

	// stage is the builder state we're at.
	stage Stage

	// lenient, when enabled, records assertions without aborting.
	lenient bool

	// suppliers contains functions that the Asserter uses to obtain values from
	// the builder that may vary during construction.
	suppliers suppliers
}

// suppliers is a struct containing functions that the Asserter will use to
// obtain values it needs from the builder.
type suppliers struct {
	messages     func() []*ApplicableMessage
	stateTracker func() *StateTracker
	actors       func() *Actors
	preroot      func() cid.Cid
}

var _ require.TestingT = &Asserter{}

func NewAsserter(id string, pv ProtocolVersion, lenient bool, suppliers suppliers) *Asserter {
	a := &Asserter{id: id, pv: pv, lenient: lenient, suppliers: suppliers}
	a.Assertions = require.New(a)
	return a
}

// AtState forks this Asserter making it point to the specified state root.
func (a *Asserter) AtState(root cid.Cid) *Asserter {
	// fork the state tracker at the specified root.
	st := a.suppliers.stateTracker().Fork(root)

	// clone the asserter, replacing the stateTracker function.
	cpy := *a
	cpy.suppliers.stateTracker = func() *StateTracker {
		return st
	}
	return &cpy
}

// enterStage sets a new stage in the Asserter.
func (a *Asserter) enterStage(stage Stage) {
	a.stage = stage
}

// In is assert fluid version of require.Contains. It inverts the argument order,
// such that the admissible set can be supplied through assert variadic argument.
func (a *Asserter) In(v interface{}, set ...interface{}) {
	a.Contains(set, v, "set %v does not contain element %v", set, v)
}

// BalanceEq verifies that the balance of the address equals the expected one.
func (a *Asserter) BalanceEq(addr address.Address, expected abi.TokenAmount) {
	st := a.suppliers.stateTracker()
	actor, err := st.StateTree.GetActor(addr)
	a.NoError(err, "failed to fetch actor %s from state", addr)
	a.Equal(expected, actor.Balance, "balances mismatch for address %s", addr)
}

// NonceEq verifies that the nonce of the actor equals the expected one.
func (a *Asserter) NonceEq(addr address.Address, expected uint64) {
	st := a.suppliers.stateTracker()
	actor, err := st.StateTree.GetActor(addr)
	a.NoError(err, "failed to fetch actor %s from state", addr)
	a.Equal(expected, actor.Nonce, "expected actor %s nonce: %d, got: %d", addr, expected, actor.Nonce)
}

// HeadEq verifies that the head of the actor equals the expected one.
func (a *Asserter) HeadEq(addr address.Address, expected cid.Cid) {
	st := a.suppliers.stateTracker()
	actor, err := st.StateTree.GetActor(addr)
	a.NoError(err, "failed to fetch actor %s from state", addr)
	a.Equal(expected, actor.Head, "expected actor %s head: %v, got: %v", addr, expected, actor.Head)
}

// ExitCodeEq verifies two exit codes are the same (and prints system codes nicely).
func (a *Asserter) ExitCodeEq(actual exitcode.ExitCode, expected exitcode.ExitCode) {
	a.Equal(expected, actual, "expected exit code: %s, got: %s", expected, actual)
}

// ActorExists verifies that the actor exists in the state tree.
func (a *Asserter) ActorExists(addr address.Address) {
	st := a.suppliers.stateTracker()
	_, err := st.StateTree.GetActor(addr)
	a.NoError(err, "expected no error while looking up actor %s", addr)
}

// ActorExists verifies that the actor is absent from the state tree.
func (a *Asserter) ActorMissing(addr address.Address) {
	st := a.suppliers.stateTracker()
	_, err := st.StateTree.GetActor(addr)
	a.Error(err, "expected error while looking up actor %s", addr)
}

// LastMessageResultSatisfies verifies that the last applied message result
// satisfies the provided predicate.
func (a *Asserter) LastMessageResultSatisfies(predicate ApplyRetPredicate) {
	msgs := a.suppliers.messages()
	except := msgs[0 : len(msgs)-1]
	a.EveryMessageResultSatisfies(predicate, except...)
}

// EveryMessageResultSatisfies verifies that every message result satisfies the
// provided predicate.
func (a *Asserter) EveryMessageResultSatisfies(predicate ApplyRetPredicate, except ...*ApplicableMessage) {
	exceptm := make(map[*ApplicableMessage]struct{}, len(except))
	for _, am := range except {
		exceptm[am] = struct{}{}
	}
	msgs := a.suppliers.messages()
	for i, m := range msgs {
		if _, ok := exceptm[m]; ok {
			continue
		}
		err := predicate(m.Result)
		a.NoError(err, "message result predicate failed on message %d: %s", i, err)
	}
}

// EveryMessageSenderSatisfies verifies that the sender actors of the supplied
// messages match a condition.
//
// This function groups ApplicableMessages by sender actor, and calls the
// predicate for each unique sender, passing in the initial state (when
// preconditions were committed), the final state (could be nil), and the
// ApplicableMessages themselves.
func (a *Asserter) MessageSendersSatisfy(predicate ActorPredicate, ams ...*ApplicableMessage) {
	st := a.suppliers.stateTracker()
	actors := a.suppliers.actors()
	preroot := a.suppliers.preroot()

	bysender := make(map[AddressHandle][]*ApplicableMessage, len(ams))
	for _, am := range ams {
		h := actors.HandleFor(am.Message.From)
		bysender[h] = append(bysender[h], am)
	}
	// we now have messages organized by unique senders.
	for sender, amss := range bysender {
		// get precondition state
		pretree, err := state.LoadStateTree(st.Stores.CBORStore, preroot)
		a.NoError(err)
		prestate, err := pretree.GetActor(sender.Robust)
		a.NoError(err)

		// get postcondition state; if actor has been deleted, we store a nil.
		poststate, _ := st.StateTree.GetActor(sender.Robust)

		// invoke predicate.
		err = predicate(sender, prestate, poststate, amss)
		a.NoError(err, "'every sender actor' predicate failed for sender %s: %s", sender, err)
	}
}

// EveryMessageSenderSatisfies is sugar for MessageSendersSatisfy(predicate, Messages.All()),
// but supports an exclusion set to restrict the messages that will actually be asserted.
func (a *Asserter) EveryMessageSenderSatisfies(predicate ActorPredicate, except ...*ApplicableMessage) {
	ams := a.suppliers.messages()
	if len(except) > 0 {
		filtered := ams[:0]
		for _, ex := range except {
			for _, am := range ams {
				if am == ex {
					continue
				}
				filtered = append(filtered, am)
			}
		}
		ams = filtered
	}
	a.MessageSendersSatisfy(predicate, ams...)
}

func (a *Asserter) FailNow() {
	if !a.lenient {
		os.Exit(1)
	}
	fmt.Println("⏩  ignoring assertion failure in lenient mode")
}

func (a *Asserter) Errorf(format string, args ...interface{}) {
	stage := a.stage
	fmt.Printf("❌  id: %s, pv: %s, stage: %s:"+format, append([]interface{}{a.id, a.pv.ID, stage}, args...)...)
}
