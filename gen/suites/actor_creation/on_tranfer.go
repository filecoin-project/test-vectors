package main

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/exitcode"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

type actorCreationOnTransferParams struct {
	senderType   address.Protocol
	senderBal    abi.TokenAmount
	receiverAddr address.Address
	amount       abi.TokenAmount
	expectedCode exitcode.ExitCode
}

func actorCreationOnTransfer(params actorCreationOnTransferParams) func(v *MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		// Set up sender account.
		sender := v.Actors.Account(params.senderType, params.senderBal)
		v.CommitPreconditions()

		// Perform the transfer.
		v.Messages.Sugar().Transfer(sender.ID, params.receiverAddr, Value(params.amount), Nonce(0))
		v.CommitApplies()

		v.Assert.EveryMessageResultSatisfies(ExitCode(params.expectedCode))
		v.Assert.EveryMessageSenderSatisfies(BalanceUpdated(big.Zero()))

		if params.expectedCode.IsSuccess() {
			v.Assert.EveryMessageSenderSatisfies(NonceUpdated())
			v.Assert.BalanceEq(params.receiverAddr, params.amount)
		}
	}
}
