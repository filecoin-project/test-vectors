package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/lotus/chain/actors/builtin/paych"
	"github.com/filecoin-project/lotus/chain/types"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func failActorExecutionAborted(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	// Set up sender and receiver accounts.
	var sender, receiver AddressHandle
	var paychAddr AddressHandle

	v.Actors.AccountN(address.SECP256K1, balance1T, &sender, &receiver)
	paychAddr = AddressHandle{
		ID:     MustNewIDAddr(MustIDFromAddress(receiver.ID) + 1),
		Robust: sender.NextActorAddress(0, 0),
	}
	v.CommitPreconditions()

	// Construct the payment channel.
	createMsg := v.Messages.Sugar().PaychMessage(sender.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Create(receiver.Robust, abi.NewTokenAmount(10_000))
	}, Value(abi.NewTokenAmount(10_000)))

	// Update the payment channel.
	updateMsg := v.Messages.Sugar().PaychMessage(sender.Robust, func(b paych.MessageBuilder) (*types.Message, error) {
		return b.Update(paychAddr.Robust, &paych.SignedVoucher{
			ChannelAddr: paychAddr.Robust,
			TimeLockMin: abi.ChainEpoch(10),
			Lane:        123,
			Nonce:       1,
			Amount:      big.NewInt(10),
			Signature: &crypto.Signature{
				Type: crypto.SigTypeBLS,
				Data: []byte("Grrr im an invalid signature, I cause panics in the payment channel actor"),
			},
		}, nil)
	}, Nonce(1), Value(big.Zero()))

	v.CommitApplies()

	v.Assert.Equal(exitcode.Ok, createMsg.Result.ExitCode)
	v.Assert.Equal(exitcode.ErrIllegalArgument, updateMsg.Result.ExitCode)
}
