package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/specs-actors/actors/builtin"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

var (
	balance = abi.NewTokenAmount(1_000_000_000_000_000)
)

func main() {
	g := NewGenerator()
	defer g.Close()

	g.Group("reward",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "ok-miners-awarded-no-premiums",
				Version: "v1",
				Desc:    "verifies that miners are awarded for the mining of blocks; no premiums",
			},
			TipsetFunc: minersAwardedNoPremiums,
		},
	)

	g.Group("penalties",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "penalize-inexistent-sender-multiple-miners",
				Version: "v1",
				Desc:    "verifies that a miner including a message from an inexistent sender is penalized; only the first miner is penalized",
				Comment: "https://github.com/filecoin-project/lotus/issues/3491",
			},
			TipsetFunc: minerPenalized(3, func(v *TipsetVectorBuilder) {
				v.StagedMessages.SetDefaults(GasLimit(1_000_000_000), GasPremium(0), GasFeeCap(200))

				from := MustNewBLSAddr(1000)                // random
				to := v.Actors.Miners()[0].OwnerAddr.Robust // sending funds to the first miner's owner account
				v.StagedMessages.Sugar().Transfer(from, to, Value(abi.NewTokenAmount(1)), Nonce(0))
			}, nil),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "penalize-non-account-sender-multiple-miners",
				Version: "v1",
				Desc:    "verifies that a miner including a message from a non-account sender is penalized; only the first miner is penalized",
				Comment: "https://github.com/filecoin-project/lotus/issues/3491",
			},
			TipsetFunc: minerPenalized(3, func(v *TipsetVectorBuilder) {
				v.StagedMessages.SetDefaults(GasLimit(1_000_000_000), GasPremium(0), GasFeeCap(200))

				from := []address.Address{
					v.Actors.Miners()[0].MinerActorAddr.Robust,
					builtin.SystemActorAddr,
					builtin.InitActorAddr,
					builtin.CronActorAddr,
				}

				to := v.Actors.Miners()[0].OwnerAddr.Robust // sending funds to the first miner's owner account
				for i, f := range from {
					v.StagedMessages.Sugar().Transfer(f, to, Value(abi.NewTokenAmount(1)), Nonce(uint64(i)))
				}
			}, nil),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "penalize-bad-nonce-multiple-miners",
				Version: "v1",
				Desc:    "verifies that a miner including a message with a bad nonce is penalized; only the first miner is penalized",
				Comment: "https://github.com/filecoin-project/lotus/issues/3491",
			},
			TipsetFunc: minerPenalized(3, func(v *TipsetVectorBuilder) {
				v.StagedMessages.SetDefaults(GasLimit(1_000_000_000), GasPremium(0), GasFeeCap(200))

				from := v.Actors.Miners()[0].OwnerAddr.Robust
				to := v.Actors.Miners()[0].OwnerAddr.Robust // sending funds to the first miner's owner account
				v.StagedMessages.Sugar().Transfer(from, to, Value(abi.NewTokenAmount(1)), Nonce(100))
			}, nil),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "penalize-insufficient-balance-to-cover-gas",
				Version: "v1",
				Desc:    "verifies that a miner including a message where the sender has insufficient balance to cover gas for chain inclusion is penalized; only the first miner is penalized",
				Comment: "https://github.com/filecoin-project/lotus/issues/3491",
			},
			TipsetFunc: minerPenalized(3, func(v *TipsetVectorBuilder) {
				v.StagedMessages.SetDefaults(GasLimit(1_000_000_000), GasPremium(0), GasFeeCap(200))

				from := v.Actors.Miners()[0].OwnerAddr.Robust
				to := MustNewBLSAddr(1000)

				// create the to account actor by sending it 1 attoFIL.
				// Assuming the gas cost of transfer to the burnt funds actor is > 1, we send a transfer for value 0.
				// This won't trip the balance check for chain inclusion, but will run out of gas during execution.
				v.StagedMessages.Sugar().Transfer(from, to, Value(abi.NewTokenAmount(1)), Nonce(0))
				msg := v.StagedMessages.Sugar().Transfer(to, builtin.BurntFundsActorAddr, Value(big.Zero()), Nonce(0))
				v.StagedMessages.ApplyN(v.StagedMessages.All()...)

				// penalty is non-zero.
				v.Assert.Greater(msg.Result.GasCosts.MinerPenalty.Uint64(), uint64(0))
			}, nil),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "not-penalized-insufficient-balance-to-cover-gas-and-transfer",
				Version: "v1",
				Desc:    "verifies that a miner including transfer with enough balance to cover gas but whose value+gas exceeds sender balance is NOT penalized",
			},
			TipsetFunc: minerPenalized(3, func(v *TipsetVectorBuilder) {
				v.StagedMessages.SetDefaults(GasLimit(1_000_000_000), GasPremium(0), GasFeeCap(200))

				from := v.Actors.Miners()[0].OwnerAddr.Robust
				to := MustNewBLSAddr(1000)

				bal := v.StateTracker.Balance(from)
				v.StagedMessages.Sugar().Transfer(from, to, Value(bal), Nonce(0))
			}, func(v *TipsetVectorBuilder) {
				// penalty is zero.
				v.Assert.Zero(v.Tipsets.Messages()[0].Result.GasCosts.MinerPenalty.Uint64())
			}),
		},
		&VectorDef{
			Metadata: &Metadata{
				ID:      "not-penalized-insufficient-gas-for-return",
				Version: "v1",
				Desc:    "verifies that a miner including a message that winds up failing due to insufficent gas to cover the return value is NOT penalized",
			},
			TipsetFunc: minerPenalized(3, func(v *TipsetVectorBuilder) {
				v.StagedMessages.SetDefaults(Value(big.Zero()), GasLimit(1_000_000_000), GasPremium(0), GasFeeCap(200))

				from := v.Actors.Miners()[0].OwnerAddr.Robust
				to := v.Actors.Miners()[0].MinerActorAddr.Robust

				// discover the cost of a message that returns data.
				msg := v.StagedMessages.Typed(from, to, MinerControlAddresses(nil), Nonce(0))
				v.StagedMessages.ApplyOne(msg)

				// send a message with gas limit = gas used - 1, so that it fails last minute (returning)
				v.StagedMessages.Typed(from, to, MinerControlAddresses(nil), Nonce(1), GasLimit(msg.Result.GasUsed-1))
			}, func(v *TipsetVectorBuilder) {
				msgs := v.Tipsets.Messages()
				v.Assert.Equal(exitcode.Ok, msgs[0].Result.ExitCode)                         // first msg ok
				v.Assert.Equal(exitcode.SysErrOutOfGas, msgs[1].Result.ExitCode)             // second msg fail
				v.Assert.Zero(v.Tipsets.Messages()[0].Result.GasCosts.MinerPenalty.Uint64()) // no penalties levied
			}),
		},
	)
}
