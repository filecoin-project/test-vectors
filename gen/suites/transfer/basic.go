package main

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/exitcode"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

type basicTransferParams struct {
	senderType   address.Protocol
	senderBal    abi.TokenAmount
	receiverType address.Protocol
	amount       abi.TokenAmount
	expectedCode exitcode.ExitCode
}

func basicTransfer(params basicTransferParams) func(v *MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(gasLimit), GasPremium(1), GasFeeCap(gasFeeCap))

		// Set up sender and receiver accounts.
		var sender, receiver AddressHandle
		sender = v.Actors.Account(params.senderType, params.senderBal)
		receiver = v.Actors.Account(params.receiverType, big.Zero())
		v.CommitPreconditions()

		// Perform the transfer.
		v.Messages.Sugar().Transfer(sender.ID, receiver.ID, Value(params.amount), Nonce(0))
		v.CommitApplies()

		v.Assert.EveryMessageResultSatisfies(ExitCode(params.expectedCode))
		v.Assert.EveryMessageSenderSatisfies(BalanceUpdated(big.Zero()))

		if params.expectedCode.IsSuccess() {
			v.Assert.EveryMessageSenderSatisfies(NonceUpdated())
			v.Assert.BalanceEq(receiver.ID, params.amount)
		}
	}
}
