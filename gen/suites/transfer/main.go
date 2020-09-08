package main

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/specs-actors/actors/builtin"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

const (
	gasLimit   = 1_000_000_000
	gasFeeCap  = 200
	gasPremium = 1
)

func main() {
	g := NewGenerator()
	defer g.Close()

	g.Group("basic",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok",
				Version: "v1",
				Desc:    "successfully transfer funds from sender to receiver",
			},
			MessageFunc: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(10 * gasLimit * gasFeeCap),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(50),
				expectedCode: exitcode.Ok,
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-zero",
				Version: "v1",
				Desc:    "successfully transfer zero funds from sender to receiver",
			},
			MessageFunc: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(10 * gasFeeCap * gasLimit),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(0),
				expectedCode: exitcode.Ok,
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-exceed-balance",
				Version: "v1",
				Desc:    "fail to transfer more funds than sender balance > 0",
			},
			MessageFunc: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(10 * gasFeeCap * gasLimit),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(10*gasFeeCap*gasLimit - gasFeeCap*gasLimit + 1),
				expectedCode: exitcode.SysErrInsufficientFunds,
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-balance-equal-gas",
				Version: "v1",
				Desc:    "fail to transfer more funds than sender has when sender balance matches gas limit",
			},
			MessageFunc: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(gasFeeCap * gasLimit),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(1),
				expectedCode: exitcode.SysErrInsufficientFunds,
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-balance-under-gaslimit",
				Version: "v1",
				Desc:    "fail to transfer when sender balance under gas limit",
			},
			MessageFunc: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(gasFeeCap*gasLimit - 1),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(0),
				expectedCode: exitcode.SysErrSenderStateInvalid,
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-negative-amount",
				Version: "v1",
				Desc:    "fail to transfer a negative amount",
			},
			MessageFunc: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(10 * gasLimit * gasFeeCap),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(-50),
				expectedCode: exitcode.SysErrForbidden,
			}),
		},
	)

	g.Group("self_transfer",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "secp-to-secp-addresses",
				Version: "v1",
			},
			MessageFunc: selfTransfer(AddressHandle.RobustAddr, AddressHandle.RobustAddr),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "secp-to-id-addresses",
				Version: "v1",
			},
			MessageFunc: selfTransfer(AddressHandle.RobustAddr, AddressHandle.IDAddr),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "id-to-secp-addresses",
				Version: "v1",
			},
			MessageFunc: selfTransfer(AddressHandle.IDAddr, AddressHandle.RobustAddr),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "id-to-id-addresses",
				Version: "v1",
			},
			MessageFunc: selfTransfer(AddressHandle.IDAddr, AddressHandle.IDAddr),
		},
	)

	g.Group("unknown_accounts",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-unknown-sender-known-receiver",
				Version: "v1",
				Desc:    "fail to transfer from unknown account to known address",
			},
			MessageFunc: failTransferUnknownSenderKnownReceiver,
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-unknown-sender-unknown-receiver",
				Version: "v1",
				Desc:    "fail to transfer from unknown address to unknown address",
			},
			Mode:        ModeLenientAssertions,
			MessageFunc: failTransferUnknownSenderUnknownReceiver,
		},
	)

	sysActors := []struct {
		name      string
		addr      address.Address
		extraFunc func(am *ApplicableMessage) big.Int
	}{
		{name: "system", addr: builtin.SystemActorAddr},
		{name: "init", addr: builtin.InitActorAddr},
		{name: "reward", addr: builtin.RewardActorAddr, extraFunc: func(am *ApplicableMessage) big.Int {
			return GetMinerReward(am) // "miner tip"
		}},
		{name: "cron", addr: builtin.CronActorAddr},
		{name: "storage-power", addr: builtin.StoragePowerActorAddr},
		{name: "storage-market", addr: builtin.StorageMarketActorAddr},
		{name: "verified-registry", addr: builtin.VerifiedRegistryActorAddr},
		{name: "burnt-funds", addr: builtin.BurntFundsActorAddr, extraFunc: func(am *ApplicableMessage) big.Int {
			return CalculateBurntGas(am)
		}},
	}

	var sysReceiverItems []*VectorDef
	for _, a := range sysActors {
		sysReceiverItems = append(sysReceiverItems, &VectorDef{
			Metadata: &Metadata{
				ID:      fmt.Sprintf("to-%s-actor", a.name),
				Version: "v1",
				Comment: "May break in the future if send to a system actor becomes" +
					" disallowed: https://github.com/filecoin-project/specs/issues/1069",
			},
			MessageFunc: transferToSystemActor(a.addr, a.extraFunc),
		})
	}

	g.Group("system_receiver", sysReceiverItems...)
}
