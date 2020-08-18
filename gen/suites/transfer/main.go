package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"

	. "github.com/filecoin-project/test-vectors/gen/builders"
	. "github.com/filecoin-project/test-vectors/schema"
)

const (
	gasLimit  = 1_000_000_000
	gasFeeCap = 200
)

func main() {
	g := NewGenerator()
	defer g.Wait()

	g.MessageVectorGroup("basic",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "ok",
				Version: "v1",
				Desc:    "successfully transfer funds from sender to receiver",
			},
			Func: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(10 * gasLimit * gasFeeCap),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(50),
				exitCode:     exitcode.Ok,
			}),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "ok-zero",
				Version: "v1",
				Desc:    "successfully transfer zero funds from sender to receiver",
			},
			Func: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(10 * gasFeeCap * gasLimit),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(0),
				exitCode:     exitcode.Ok,
			}),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-exceed-balance",
				Version: "v1",
				Desc:    "fail to transfer more funds than sender balance > 0",
			},
			Func: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(10 * gasFeeCap * gasLimit),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(10*gasFeeCap*gasLimit - gasFeeCap*gasLimit + 1),
				exitCode:     exitcode.SysErrInsufficientFunds,
			}),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-balance-equal-gas",
				Version: "v1",
				Desc:    "fail to transfer more funds than sender has when sender balance matches gas limit",
			},
			Func: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(gasFeeCap * gasLimit),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(1),
				exitCode:     exitcode.SysErrInsufficientFunds,
			}),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-balance-under-gaslimit",
				Version: "v1",
				Desc:    "fail to transfer when sender balance under gas limit",
			},
			Func: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(gasFeeCap*gasLimit - 1),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(0),
				exitCode:     exitcode.SysErrSenderStateInvalid,
			}),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-negative-amount",
				Version: "v1",
				Desc:    "fail to transfer a negative amount",
			},
			Func: basicTransfer(basicTransferParams{
				senderType:   address.SECP256K1,
				senderBal:    abi.NewTokenAmount(10 * gasLimit * gasFeeCap),
				receiverType: address.SECP256K1,
				amount:       abi.NewTokenAmount(-50),
				exitCode:     exitcode.SysErrForbidden,
			}),
		},
	)

	g.MessageVectorGroup("self_transfer",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "secp-to-secp-addresses",
				Version: "v1",
			},
			Func: selfTransfer(AddressHandle.RobustAddr, AddressHandle.RobustAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "secp-to-id-addresses",
				Version: "v1",
			},
			Func: selfTransfer(AddressHandle.RobustAddr, AddressHandle.IDAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "id-to-secp-addresses",
				Version: "v1",
			},
			Func: selfTransfer(AddressHandle.IDAddr, AddressHandle.RobustAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "id-to-id-addresses",
				Version: "v1",
			},
			Func: selfTransfer(AddressHandle.IDAddr, AddressHandle.IDAddr),
		},
	)

	g.MessageVectorGroup("unknown_accounts",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-unknown-sender-known-receiver",
				Version: "v1",
				Desc:    "fail to transfer from unknown account to known address",
			},
			Func: failTransferUnknownSenderKnownReceiver,
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-unknown-sender-unknown-receiver",
				Version: "v1",
				Desc:    "fail to transfer from unknown address to unknown address",
			},
			Mode: ModeLenientAssertions,
			Func: failTransferUnknownSenderUnknownReceiver,
		},
	)

	g.MessageVectorGroup("invalid_receiver",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-system-actor",
				Version: "v1",
			},
			Func: failTransferToSystemActor(builtin.SystemActorAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-init-actor",
				Version: "v1",
			},
			Func: failTransferToSystemActor(builtin.InitActorAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-reward-actor",
				Version: "v1",
			},
			Func: failTransferToSystemActor(builtin.RewardActorAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-cron-actor",
				Version: "v1",
			},
			Func: failTransferToSystemActor(builtin.CronActorAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-storage-power-actor",
				Version: "v1",
			},
			Func: failTransferToSystemActor(builtin.StoragePowerActorAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-storage-market-actor",
				Version: "v1",
			},
			Func: failTransferToSystemActor(builtin.StorageMarketActorAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-verified-registry-actor",
				Version: "v1",
			},
			Func: failTransferToSystemActor(builtin.VerifiedRegistryActorAddr),
		},
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-burnt-funds-actor",
				Version: "v1",
			},
			Func: failTransferToSystemActor(builtin.BurntFundsActorAddr),
		},
	)
}
