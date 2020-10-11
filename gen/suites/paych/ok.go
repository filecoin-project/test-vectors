package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/lotus/chain/actors/builtin/paych"
	"github.com/filecoin-project/lotus/chain/types"

	init_ "github.com/filecoin-project/specs-actors/actors/builtin/init"
	paych0 "github.com/filecoin-project/specs-actors/actors/builtin/paych"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func happyPathCreate(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	// Set up sender and receiver accounts.
	var sender, receiver AddressHandle
	v.Actors.AccountN(address.SECP256K1, initialBal, &sender, &receiver)
	v.CommitPreconditions()

	// Add the constructor message.
	createMsg := v.Messages.Sugar().PaychMessage(sender.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Create(receiver.Robust, toSend)
	}, Value(toSend))
	v.CommitApplies()

	expectedActorAddr := AddressHandle{
		ID:     MustNewIDAddr(MustIDFromAddress(receiver.ID) + 1),
		Robust: sender.NextActorAddress(0, 0),
	}

	// Verify init actor return.
	// TODO no abstraction for init messages.
	var ret init_.ExecReturn
	MustDeserialize(createMsg.Result.Return, &ret)
	v.Assert.Equal(expectedActorAddr.Robust, ret.RobustAddress)
	v.Assert.Equal(expectedActorAddr.ID, ret.IDAddress)

	// Verify the paych state.
	head := v.StateTracker.Header(ret.IDAddress)
	state, err := paych.Load(v.StateTracker.Stores.ADTStore, head)
	v.Assert.NoError(err)

	from, _ := state.From()
	to, _ := state.To()
	v.Assert.Equal(sender.ID, from)
	v.Assert.Equal(receiver.ID, to)
	v.Assert.Equal(toSend, head.Balance)

	v.Assert.EveryMessageSenderSatisfies(NonceUpdated())
}

func happyPathUpdate(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	var (
		timelock = abi.ChainEpoch(0)
		lane     = uint64(123)
		nonce    = uint64(1)
		amount   = big.NewInt(10)
	)

	// Set up sender and receiver accounts.
	var sender, receiver AddressHandle
	var paychAddr AddressHandle

	v.Actors.AccountN(address.SECP256K1, initialBal, &sender, &receiver)
	paychAddr = AddressHandle{
		ID:     MustNewIDAddr(MustIDFromAddress(receiver.ID) + 1),
		Robust: sender.NextActorAddress(0, 0),
	}
	v.CommitPreconditions()

	// Construct the payment channel.
	createMsg := v.Messages.Sugar().PaychMessage(sender.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Create(receiver.Robust, toSend)
	}, Value(toSend))

	sv := paych.SignedVoucher{
		ChannelAddr:     paychAddr.Robust,
		TimeLockMin:     timelock,
		TimeLockMax:     0, // TimeLockMax set to 0 means no timeout
		Lane:            lane,
		Nonce:           nonce,
		Amount:          amount,
		MinSettleHeight: 0,
	}
	sb, err := sv.SigningBytes()
	v.Assert.NoError(err)
	sig, err := v.Wallet.Sign(receiver.Robust, sb)
	v.Assert.NoError(err)
	sv.Signature = sig

	// Update the payment channel.
	v.Messages.Sugar().PaychMessage(sender.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Update(paychAddr.Robust, &sv, nil)
	}, Nonce(1), Value(big.Zero()))

	v.CommitApplies()

	// all messages succeeded.
	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))

	// Verify init actor return.
	var ret init_.ExecReturn
	MustDeserialize(createMsg.Result.Return, &ret)

	// Verify the paych state.
	head := v.StateTracker.Header(ret.IDAddress)
	state, err := paych.Load(v.StateTracker.Stores.ADTStore, head)
	v.Assert.NoError(err)

	laneCnt, _ := state.LaneCount()
	v.Assert.EqualValues(1, laneCnt)

	_ = state.ForEachLaneState(func(idx uint64, dl paych.LaneState) error {
		redeemed, _ := dl.Redeemed()
		nonce, _ := dl.Nonce()

		v.Assert.Equal(amount, redeemed)
		v.Assert.Equal(nonce, nonce)
		return nil
	})

	v.Assert.EveryMessageSenderSatisfies(NonceUpdated())
}

func happyPathCollect(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	// Set up sender and receiver accounts.
	var sender, receiver AddressHandle
	var paychAddr AddressHandle
	v.Actors.AccountN(address.SECP256K1, initialBal, &sender, &receiver)
	paychAddr = AddressHandle{
		ID:     MustNewIDAddr(MustIDFromAddress(receiver.ID) + 1),
		Robust: sender.NextActorAddress(0, 0),
	}

	v.CommitPreconditions()

	// Construct the payment channel.
	createMsg := v.Messages.Sugar().PaychMessage(sender.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Create(receiver.Robust, toSend)
	}, Value(toSend))

	sv := paych.SignedVoucher{
		ChannelAddr:     paychAddr.Robust,
		TimeLockMin:     0,
		TimeLockMax:     0, // TimeLockMax set to 0 means no timeout
		Lane:            1,
		Nonce:           1,
		Amount:          toSend,
		MinSettleHeight: 0,
	}
	sb, err := sv.SigningBytes()
	v.Assert.NoError(err)
	sig, err := v.Wallet.Sign(receiver.Robust, sb)
	v.Assert.NoError(err)
	sv.Signature = sig

	// Update the payment channel.
	updateMsg := v.Messages.Sugar().PaychMessage(sender.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Update(paychAddr.Robust, &sv, nil)
	}, Nonce(1), Value(big.Zero()))

	settleMsg := v.Messages.Sugar().PaychMessage(receiver.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Settle(paychAddr.Robust)
	}, Value(big.Zero()), Nonce(0))

	// advance the epoch so the funds may be redeemed.
	collectMsg := v.Messages.Sugar().PaychMessage(receiver.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Collect(paychAddr.Robust)
	}, Value(big.Zero()), Nonce(1), EpochOffset(paych0.SettleDelay)) // TODO accessing specs-actors directly, SettleDelay is not exposed

	v.CommitApplies()

	// all messages succeeded.
	v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))

	v.Assert.MessageSendersSatisfy(BalanceUpdated(big.Zero()), createMsg, updateMsg)
	v.Assert.MessageSendersSatisfy(BalanceUpdated(toSend), settleMsg, collectMsg)
	v.Assert.EveryMessageSenderSatisfies(NonceUpdated())

	// the paych actor should have been deleted after the collect
	v.Assert.ActorMissing(paychAddr.Robust)
	v.Assert.ActorMissing(paychAddr.ID)
}
