package main

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/runtime/exitcode"
	"github.com/filecoin-project/test-vectors/chaos"
	. "github.com/filecoin-project/test-vectors/gen/builders"
	. "github.com/filecoin-project/test-vectors/schema"
)

func main() {
	g := NewGenerator()
	defer g.Wait()

	g.MessageVectorGroup("no_beneficiary",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "delete-w-zero-balance",
				Version: "v1",
				Desc:    "actor with zero balance is deleted, does not require beneficiary",
			},
			Selector: Selector{"chaos_actor": "true"},
			Func:     deleteActor,
		},
	)

	g.MessageVectorGroup("beneficiary",
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "delete-w-balance-and-beneficiary",
				Version: "v1",
				Desc:    "actor with non-zero balance is deleted and sends funds to beneficiary",
			},
			Selector: Selector{"chaos_actor": "true"},
			Func:     deleteActorWithBeneficiary(big.NewInt(50), address.Undef, exitcode.Ok),
		},
		// FIXME: this currently panics
		// &MessageVectorGenItem{
		// 	Metadata: &Metadata{
		// 		ID:      "fail-delete-w-balance-and-unkown-beneficiary",
		// 		Version: "v1",
		// 		Desc:    "fails when actor with non-zero balance is deleted but beneficiary address is unknown",
		// 	},
		// 	Selector: Selector{"chaos_actor": "true"},
		// 	Func:     deleteActorWithBeneficiary(big.NewInt(50), MustNewSECP256K1Addr("!👹*_👹!"), exitcode.SysErrorIllegalActor),
		// },
		&MessageVectorGenItem{
			Metadata: &Metadata{
				ID:      "fail-delete-w-balance-and-self-beneficiary",
				Version: "v1",
				Desc:    "fails when actor with non-zero balance is deleted but beneficiary is the calling actor",
				Comment: "should abort if the beneficiary is the calling actor, see https://github.com/filecoin-project/specs-actors/blob/bcd83e8eb0a98b684851e574a2bd8d4130e21a51/actors/runtime/runtime.go#L117",
			},
			Selector: Selector{"chaos_actor": "true"},
			Hints:    []string{HintIncorrect, HintNegate},
			Func:     deleteActorWithBeneficiary(big.NewInt(50), chaos.Address, exitcode.Ok),
		},
	)
}

// deleteActor builds a test vector that tests the simple case of deleting an
// actor with zero balance.
func deleteActor(v *Builder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	sender := v.Actors.Account(address.SECP256K1, big.NewInt(1_000_000_000_000_000))
	v.CommitPreconditions()

	v.Assert.ActorExists(chaos.Address)
	v.Assert.BalanceEq(chaos.Address, big.Zero())

	delMsg := v.Messages.Raw(
		sender.ID,
		chaos.Address,
		chaos.MethodDeleteActor,
		MustSerialize(&builtin.BurntFundsActorAddr),
		Value(big.Zero()),
		Nonce(0),
	)
	v.CommitApplies()

	v.Assert.NoError(ExitCode(exitcode.Ok)(delMsg.Result))
	v.Assert.ActorMissing(chaos.Address)
}

// deleteActorWithBeneficiary builds a test vector that tests deleting an actor
// and transfering their funds to another actor. Use address.Undef to have a
// beneficiary created automatically. actorFunds MUST be greater than zero.
func deleteActorWithBeneficiary(actorFunds big.Int, beneficiaryAddr address.Address, code exitcode.ExitCode) func(v *Builder) {
	return func(v *Builder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		sender := v.Actors.Account(address.SECP256K1, big.Add(big.NewInt(1_000_000_000_000_000), actorFunds))

		if beneficiaryAddr == address.Undef {
			beneficiaryAddr = v.Actors.Account(address.SECP256K1, big.Zero()).ID
		}

		v.CommitPreconditions()

		v.Assert.ActorExists(chaos.Address)
		v.Assert.BalanceEq(chaos.Address, big.Zero())

		// transfer required funds to the actor that will be deleted
		m := v.Messages.Sugar().Transfer(sender.ID, chaos.Address, Value(actorFunds), Nonce(0))
		v.Messages.ApplyOne(m)
		v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))

		// if this is will succeed, record the current balance so we can check funds
		// were transferred to the beneficiary
		var bal big.Int
		if code == exitcode.Ok {
			bal = v.Actors.Balance(beneficiaryAddr)
		}

		v.Messages.Raw(
			sender.ID,
			chaos.Address,
			chaos.MethodDeleteActor,
			MustSerialize(&beneficiaryAddr),
			Value(big.Zero()),
			Nonce(1),
		)
		v.CommitApplies()

		v.Assert.LastMessageResultSatisfies(ExitCode(code))

		// check beneficiary received funds if it succeeded
		if code == exitcode.Ok && beneficiaryAddr != chaos.Address {
			v.Assert.ActorMissing(chaos.Address)
			v.Assert.BalanceEq(beneficiaryAddr, big.Add(bal, actorFunds))
		}
	}
}
