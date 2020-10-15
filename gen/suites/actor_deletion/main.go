package main

import (
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/specs-actors/actors/builtin"

	"github.com/filecoin-project/lotus/conformance/chaos"

	"github.com/filecoin-project/go-address"

	. "github.com/filecoin-project/test-vectors/gen/builders"
	"github.com/filecoin-project/test-vectors/schema"
)

func main() {
	g := NewGenerator()

	g.Group("no_beneficiary",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "delete-w-zero-balance",
				Version: "v1",
				Desc:    "actor with zero balance is deleted, does not require beneficiary",
			},
			Selector:    schema.Selector{"chaos_actor": "true"},
			MessageFunc: deleteActor,
		},
	)

	g.Group("beneficiary",
		&VectorDef{
			Metadata: &Metadata{
				ID:      "delete-w-balance-and-beneficiary",
				Version: "v1",
				Desc:    "actor with non-zero balance is deleted and sends funds to beneficiary",
			},
			Selector:    schema.Selector{"chaos_actor": "true"},
			MessageFunc: deleteActorWithBeneficiary(big.NewInt(50), address.Undef, exitcode.Ok),
		},
		// TODO: uncomment when merged https://github.com/filecoin-project/lotus/pull/3479
		// It is not marked with HintInvalid because it panics entirely.
		// &VectorDef{
		// 	Metadata: &Metadata{
		// 		ID:      "fail-delete-w-balance-and-unkown-beneficiary",
		// 		Version: "v1",
		// 		Desc:    "fails when actor with non-zero balance is deleted but beneficiary address is unknown",
		// 	},
		// 	Selector:    schema.Selector{"chaos_actor": "true"},
		// 	MessageFunc: deleteActorWithBeneficiary(big.NewInt(50), MustNewSECP256K1Addr("!👹*_👹!"), exitcode.SysErrorIllegalActor),
		// },
		&VectorDef{
			Metadata: &Metadata{
				ID:      "fail-delete-w-balance-and-self-beneficiary",
				Version: "v1",
				Desc:    "fails when actor with non-zero balance is deleted but beneficiary is the calling actor",
				Comment: "should abort with SysErrorIllegalArgument if the beneficiary is the calling actor, will be fixed in https://github.com/filecoin-project/lotus/pull/3478",
			},
			Selector:    schema.Selector{"chaos_actor": "true"},
			Hints:       []string{schema.HintIncorrect, schema.HintNegate},
			MessageFunc: deleteActorWithBeneficiary(big.NewInt(50), chaos.Address, exitcode.Ok),
		},
	)

	g.Close()
}

// deleteActor builds a test vector that tests the simple case of deleting an
// actor with zero balance.
func deleteActor(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	sender := v.Actors.Account(address.SECP256K1, big.NewInt(1_000_000_000_000_000))
	v.CommitPreconditions()

	v.Assert.ActorExists(chaos.Address)
	v.Assert.BalanceEq(chaos.Address, big.Zero())

	v.Messages.Raw(
		sender.ID,
		chaos.Address,
		chaos.MethodDeleteActor,
		MustSerialize(&builtin.BurntFundsActorAddr),
		Value(big.Zero()),
		Nonce(0),
	)
	v.CommitApplies()

	v.Assert.LastMessageResultSatisfies(ExitCode(exitcode.Ok))
	v.Assert.ActorMissing(chaos.Address)
}

// deleteActorWithBeneficiary builds a test vector that tests deleting an actor
// and transfering their funds to another actor. Use address.Undef to have a
// beneficiary created automatically. actorFunds MUST be greater than zero.
func deleteActorWithBeneficiary(actorFunds big.Int, beneficiaryAddr address.Address, expectedCode exitcode.ExitCode) func(v *MessageVectorBuilder) {
	return func(v *MessageVectorBuilder) {
		v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

		sender := v.Actors.Account(address.SECP256K1, big.Add(big.NewInt(1_000_000_000_000_000), actorFunds))

		beneficiaryAddr := beneficiaryAddr // capture
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
		if expectedCode == exitcode.Ok {
			bal = v.StateTracker.Balance(beneficiaryAddr)
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

		v.Assert.LastMessageResultSatisfies(ExitCode(expectedCode))

		// check beneficiary received funds if it succeeded
		if expectedCode == exitcode.Ok && beneficiaryAddr != chaos.Address {
			v.Assert.ActorMissing(chaos.Address)
			v.Assert.BalanceEq(beneficiaryAddr, big.Add(bal, actorFunds))
		}
	}
}
