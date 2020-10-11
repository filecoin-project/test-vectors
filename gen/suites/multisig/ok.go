package main

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"

	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/specs-actors/actors/builtin"
	init0 "github.com/filecoin-project/specs-actors/actors/builtin/init"
	multisig0 "github.com/filecoin-project/specs-actors/actors/builtin/multisig"

	"github.com/filecoin-project/go-address"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func constructor(v *MessageVectorBuilder) {
	var balance = abi.NewTokenAmount(1_000_000_000_000)
	var amount = abi.NewTokenAmount(10)

	v.Messages.SetDefaults(GasLimit(gasLimit), GasPremium(1), GasFeeCap(gasFeeCap))

	// Set up one account.
	alice := v.Actors.Account(address.SECP256K1, balance)
	v.CommitPreconditions()

	createMultisig(v, alice, []address.Address{alice.ID}, 1, Value(amount), Nonce(0))
	v.CommitApplies()
}

func proposeAndCancelOk(v *MessageVectorBuilder) {
	var (
		initial        = abi.NewTokenAmount(1_000_000_000_000)
		amount         = abi.NewTokenAmount(10)
		unlockDuration = abi.ChainEpoch(10)
	)

	v.Messages.SetDefaults(Value(big.Zero()), GasLimit(gasLimit), GasPremium(1), GasFeeCap(gasFeeCap))

	// Set up three accounts: alice and bob (signers), and charlie (outsider).
	var alice, bob, charlie AddressHandle
	v.Actors.AccountN(address.SECP256K1, initial, &alice, &bob, &charlie)
	v.CommitPreconditions()

	// create the multisig actor; created by alice.
	multisigAddr := createMultisig(v, alice, []address.Address{alice.ID, bob.ID}, 2, Value(amount), Nonce(0))

	// alice proposes that charlie should receive 'amount' FIL.
	hash := proposeOk(v, proposeOpts{
		multisigAddr: multisigAddr,
		sender:       alice.ID,
		recipient:    charlie.ID,
		amount:       amount,
	}, Nonce(1))

	// bob cancels alice's transaction. This fails as bob did not create alice's transaction.
	bobCancelMsg := v.Messages.Sugar().MultisigMessage(bob.ID, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Cancel(multisigAddr, 0, hash)
	}, Nonce(0))
	v.Messages.ApplyOne(bobCancelMsg)
	v.Assert.Equal(bobCancelMsg.Result.ExitCode, exitcode.ErrForbidden)

	// alice cancels their transaction; charlie doesn't receive any FIL,
	// the multisig actor's balance is empty, and the transaction is canceled.
	aliceCancelMsg := v.Messages.Sugar().MultisigMessage(alice.ID, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Cancel(multisigAddr, 0, hash)
	}, Nonce(2))
	v.Messages.ApplyOne(aliceCancelMsg)
	v.Assert.Equal(exitcode.Ok, aliceCancelMsg.Result.ExitCode)

	v.CommitApplies()

	// verify balance is untouched.
	v.Assert.BalanceEq(multisigAddr, amount)

	// reload the multisig state and verify
	multisigState, err := multisig.Load(v.StateTracker.Stores.ADTStore, v.StateTracker.Header(multisigAddr))
	v.Assert.NoError(err)

	threshold, _ := multisigState.Threshold()
	v.Assert.Equal(uint64(2), threshold)

	signers, _ := multisigState.Signers()
	v.Assert.Equal([]address.Address{alice.ID, bob.ID}, signers)

	initialBal, _ := multisigState.InitialBalance()
	v.Assert.Equal(amount, initialBal)

	startEpoch, _ := multisigState.StartEpoch()
	v.Assert.Equal(v.ProtocolVersion.FirstEpoch, startEpoch)

	unlockDur, _ := multisigState.UnlockDuration()
	v.Assert.Equal(unlockDuration, unlockDur)

	var pendingTxs int
	_ = multisigState.ForEachPendingTxn(func(_ int64, _ multisig.Transaction) error {
		pendingTxs++
		return nil
	})
	v.Assert.Equal(0, pendingTxs)
}

func proposeAndApprove(v *MessageVectorBuilder) {
	var (
		initial        = abi.NewTokenAmount(1_000_000_000_000)
		amount         = abi.NewTokenAmount(10)
		unlockDuration = abi.ChainEpoch(10)
	)

	v.Messages.SetDefaults(Value(big.Zero()), GasLimit(gasLimit), GasPremium(1), GasFeeCap(gasFeeCap))

	// Set up three accounts: alice and bob (signers), and charlie (outsider).
	var alice, bob, charlie AddressHandle
	v.Actors.AccountN(address.SECP256K1, initial, &alice, &bob, &charlie)
	v.CommitPreconditions()

	// create the multisig actor; created by alice.
	multisigAddr := createMultisig(v, alice, []address.Address{alice.ID, bob.ID}, 2, Value(amount), Nonce(0))

	// alice proposes that charlie should receive 'amount' FIL.
	hash := proposeOk(v, proposeOpts{
		multisigAddr: multisigAddr,
		sender:       alice.ID,
		recipient:    charlie.ID,
		amount:       amount,
	}, Nonce(1))

	// charlie proposes himself -> fails.
	charliePropose := v.Messages.Sugar().MultisigMessage(charlie.ID, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Propose(multisigAddr, charlie.ID, amount, builtin.MethodSend, nil)
	}, Nonce(0))
	v.Messages.ApplyOne(charliePropose)
	v.Assert.Equal(exitcode.ErrForbidden, charliePropose.Result.ExitCode)

	// charlie attempts to accept the pending transaction -> fails.
	charlieApprove := v.Messages.Sugar().MultisigMessage(charlie.ID, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Approve(multisigAddr, 0, hash)
	}, Nonce(1))
	v.Messages.ApplyOne(charlieApprove)
	v.Assert.Equal(exitcode.ErrForbidden, charlieApprove.Result.ExitCode)

	// bob approves transfer of 'amount' FIL to charlie.
	// epoch is unlockDuration + 1
	bobApprove := v.Messages.Sugar().MultisigMessage(bob.ID, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Approve(multisigAddr, 0, hash)
	}, Nonce(0), EpochOffset(unlockDuration+1))
	v.Messages.ApplyOne(bobApprove)
	v.Assert.Equal(exitcode.Ok, bobApprove.Result.ExitCode)

	v.CommitApplies()

	var approveRet multisig0.ApproveReturn
	MustDeserialize(bobApprove.Result.Return, &approveRet)
	v.Assert.Equal(multisig0.ApproveReturn{
		Applied: true,
		Code:    0,
		Ret:     nil,
	}, approveRet)

	// assert that the multisig balance has been drained, and charlie's incremented.
	v.Assert.BalanceEq(multisigAddr, big.Zero())
	v.Assert.MessageSendersSatisfy(BalanceUpdated(amount), charliePropose, charlieApprove)

	// reload the multisig state and verify
	state, err := multisig.Load(v.StateTracker.Stores.ADTStore, v.StateTracker.Header(multisigAddr))
	v.Assert.NoError(err)

	threshold, _ := state.Threshold()
	v.Assert.Equal(uint64(2), threshold)

	signers, _ := state.Signers()
	v.Assert.Equal([]address.Address{alice.ID, bob.ID}, signers)

	initialBal, _ := state.InitialBalance()
	v.Assert.Equal(amount, initialBal)

	startEpoch, _ := state.StartEpoch()
	v.Assert.Equal(v.ProtocolVersion.FirstEpoch, startEpoch)

	unlockDur, _ := state.UnlockDuration()
	v.Assert.Equal(unlockDuration, unlockDur)

	var pendingTxs int
	_ = state.ForEachPendingTxn(func(_ int64, _ multisig.Transaction) error {
		pendingTxs++
		return nil
	})
	v.Assert.Equal(0, pendingTxs)
}

func addSigner(v *MessageVectorBuilder) {
	var (
		initial = abi.NewTokenAmount(1_000_000_000_000)
		amount  = abi.NewTokenAmount(10)
	)

	v.Messages.SetDefaults(Value(big.Zero()), GasLimit(gasLimit), GasPremium(1), GasFeeCap(gasFeeCap))

	// Set up three accounts: alice and bob (signers), and charlie (outsider).
	var alice, bob, charlie AddressHandle
	v.Actors.AccountN(address.SECP256K1, initial, &alice, &bob, &charlie)
	v.CommitPreconditions()

	// create the multisig actor; created by alice.
	multisigAddr := createMultisig(v, alice, []address.Address{alice.ID}, 1, Value(amount), Nonce(0))

	addParams := &multisig0.AddSignerParams{
		Signer:   bob.ID,
		Increase: false,
	}

	// attempt to add bob as a signer; this fails because the addition needs to go through
	// the multisig flow, as it is subject to the same approval policy.
	v.Messages.Typed(alice.ID, multisigAddr, MultisigAddSigner(addParams), Nonce(1))

	// go through the multisig wallet.
	// since approvals = 1, this auto-approves the transaction.
	v.Messages.Sugar().MultisigMessage(alice.ID, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Propose(multisigAddr, multisigAddr, big.Zero(), builtin.MethodsMultisig.AddSigner, MustSerialize(addParams))
	}, Nonce(2))

	// TODO also exercise the approvals = 2 case with explicit approval.

	v.CommitApplies()

	// reload the multisig state and verify that bob is now a signer.
	state, err := multisig.Load(v.StateTracker.Stores.ADTStore, v.StateTracker.Header(multisigAddr))
	v.Assert.NoError(err)

	threshold, _ := state.Threshold()
	v.Assert.Equal(uint64(1), threshold)

	signers, _ := state.Signers()
	v.Assert.Equal([]address.Address{alice.ID, bob.ID}, signers)

	initialBal, _ := state.InitialBalance()
	v.Assert.Equal(amount, initialBal)

	startEpoch, _ := state.StartEpoch()
	v.Assert.Equal(v.ProtocolVersion.FirstEpoch, startEpoch)

	unlockDur, _ := state.UnlockDuration()
	v.Assert.Equal(abi.ChainEpoch(10), unlockDur)

	var pendingTxs int
	_ = state.ForEachPendingTxn(func(_ int64, _ multisig.Transaction) error {
		pendingTxs++
		return nil
	})
	v.Assert.Equal(0, pendingTxs)
}

type proposeOpts struct {
	multisigAddr address.Address
	sender       address.Address
	recipient    address.Address
	amount       abi.TokenAmount
}

func proposeOk(v *MessageVectorBuilder, proposeOpts proposeOpts, opts ...MsgOpt) *multisig.ProposalHashData {
	proposeMsg := v.Messages.Sugar().MultisigMessage(proposeOpts.sender, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Propose(proposeOpts.multisigAddr, proposeOpts.recipient, proposeOpts.amount, builtin.MethodSend, nil)
	}, opts...)

	v.Messages.ApplyOne(proposeMsg)

	// verify that the multisig state contains the outstanding TX.
	state, err := multisig.Load(v.StateTracker.Stores.ADTStore, v.StateTracker.Header(proposeOpts.multisigAddr))
	v.Assert.NoError(err)

	// load the transaction.
	var tx *multisig.Transaction
	_ = state.ForEachPendingTxn(func(id int64, txn multisig.Transaction) error {
		tx = &txn
		return nil
	})

	v.Assert.Equal(&multisig.Transaction{
		To:       proposeOpts.recipient,
		Value:    proposeOpts.amount,
		Method:   builtin.MethodSend,
		Approved: []address.Address{proposeOpts.sender},
	}, tx)

	return &multisig.ProposalHashData{
		Requester: proposeOpts.sender,
		To:        proposeOpts.recipient,
		Value:     proposeOpts.amount,
		Method:    builtin.MethodSend,
	}
}

func createMultisig(v *MessageVectorBuilder, creator AddressHandle, approvers []address.Address, threshold uint64, opts ...MsgOpt) address.Address {
	const unlockDuration = abi.ChainEpoch(10)

	vestingStart := abi.ChainEpoch(0) // actors v1 only supports 0 for vesting start.
	if v.ProtocolVersion.Actors >= actors.Version2 {
		// use the protocol version's initial epoch.
		vestingStart = v.ProtocolVersion.FirstEpoch
	}

	// create the multisig actor.
	msg := v.Messages.Sugar().MultisigMessage(creator.ID, func(b multisig.MessageBuilder) (*types.Message, error) {
		// TODO this is ugly -- initial amount will be overriden by value MsgOpt.
		return b.Create(approvers, threshold, vestingStart, unlockDuration, big.Zero())
	}, opts...)

	v.Messages.ApplyOne(msg)

	// verify ok
	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))

	// verify the assigned addess is as expected.
	var ret init0.ExecReturn
	MustDeserialize(msg.Result.Return, &ret)
	v.Assert.Equal(creator.NextActorAddress(msg.Message.Nonce, 0), ret.RobustAddress)
	handles := v.Actors.AccountHandles()
	v.Assert.Equal(MustNewIDAddr(MustIDFromAddress(handles[len(handles)-1].ID)+1), ret.IDAddress)

	// the multisig address's balance is incremented by the value sent to it.
	v.Assert.BalanceEq(ret.IDAddress, msg.Message.Value)

	return ret.IDAddress
}
