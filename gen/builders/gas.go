package builders

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
)

const (
	overuseNum = 11
	overuseDen = 10
)

// CalculateSenderDeduction returns the balance that shall be deducted from the
// sender's account as a result of applying this message.
func CalculateSenderDeduction(am *ApplicableMessage) big.Int {
	if am.Result.GasUsed == 0 {
		return big.Zero()
	}

	var (
		minerReward = GetMinerReward(am)         // goes to the miner
		burn        = CalculateBurntGas(am)      // vanishes
		deducted    = big.Add(minerReward, burn) // sum of gas accrued
	)
	if am.Result.ExitCode.IsSuccess() {
		deducted = big.Add(deducted, am.Message.Value) // message value
	}
	return deducted
}

// GetMinerReward returns the amount that the miner gets to keep, aka. miner tip.
func GetMinerReward(am *ApplicableMessage) abi.TokenAmount {
	gasLimit := big.NewInt(am.Message.GasLimit)
	gasPremium := am.Message.GasPremium
	return big.Mul(gasLimit, gasPremium)
}

// CalculateBurntGas calculates the amount that will be burnt, a function of the
// gas limit and the gas actually used.
func CalculateBurntGas(am *ApplicableMessage) big.Int {
	gasLimit := am.Message.GasLimit
	gasUsed := am.Result.GasUsed

	over := gasLimit - (overuseNum*gasUsed)/overuseDen
	if over < 0 {
		over = 0
	}
	if over > gasUsed {
		over = gasUsed
	}

	overestimateGas := big.NewInt(gasLimit - gasUsed)
	overestimateGas = big.Mul(overestimateGas, big.NewInt(over))
	if gasUsed != 0 {
		overestimateGas = big.Div(overestimateGas, big.NewInt(gasUsed))
	}

	totalBurnGas := big.Add(overestimateGas, big.NewInt(gasUsed))
	return big.Mul(am.baseFee, totalBurnGas)
}
