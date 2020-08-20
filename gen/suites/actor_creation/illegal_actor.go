package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"

	"github.com/filecoin-project/test-vectors/chaos"
	. "github.com/filecoin-project/test-vectors/gen/builders"
)

func actorAbortWithSystemExitCode(i int) func(v *Builder) {
	return func(v *Builder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
		v.CommitPreconditions()

		code := big.NewInt(int64(i))

		msg := v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodAbortWithSystemExitCode, MustSerialize(&code), Nonce(0), Value(big.Zero()))
		v.CommitApplies()

		v.Assert.Equal(exitcode.ExitCode(i), msg.Result.ExitCode)
	}
}

func actorAbortWithSystemExitCodeSingle(v *Builder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, abi.NewTokenAmount(1_000_000_000_000))
	v.CommitPreconditions()

	var msgs []*ApplicableMessage
	//for i := 0; i <= 15; i++ {
	for i := 1; i <= 15; i++ {
		code := big.NewInt(int64(i))

		msg := v.Messages.Raw(alice.ID, chaos.Address, chaos.MethodAbortWithSystemExitCode, MustSerialize(&code), Nonce(uint64(i-1)), Value(big.Zero()))
		msgs = append(msgs, msg)
	}

	v.CommitApplies()

	v.Assert.Equal(exitcode.ExitCode(1), msgs[0].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(2), msgs[1].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(3), msgs[2].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(4), msgs[3].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(5), msgs[4].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(6), msgs[5].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(7), msgs[6].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(8), msgs[7].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(2), msgs[8].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(2), msgs[9].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(2), msgs[10].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(2), msgs[11].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(2), msgs[12].Result.ExitCode)
	v.Assert.Equal(exitcode.ExitCode(2), msgs[13].Result.ExitCode)
}
