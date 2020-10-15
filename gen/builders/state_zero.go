package builders

import (
	"context"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	account0 "github.com/filecoin-project/specs-actors/actors/builtin/account"
	cron0 "github.com/filecoin-project/specs-actors/actors/builtin/cron"
	init0 "github.com/filecoin-project/specs-actors/actors/builtin/init"
	market0 "github.com/filecoin-project/specs-actors/actors/builtin/market"
	miner0 "github.com/filecoin-project/specs-actors/actors/builtin/miner"
	power0 "github.com/filecoin-project/specs-actors/actors/builtin/power"
	reward0 "github.com/filecoin-project/specs-actors/actors/builtin/reward"
	system0 "github.com/filecoin-project/specs-actors/actors/builtin/system"
	verifreg0 "github.com/filecoin-project/specs-actors/actors/builtin/verifreg"
	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	account2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/account"
	cron2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/cron"
	init2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/init"
	market2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/market"
	miner2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	power2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/power"
	reward2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/reward"
	system2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/system"
	verifreg2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/verifreg"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/lotus/conformance/chaos"

	"github.com/filecoin-project/test-vectors/schema"
)

type (
	// ActorsZeroStateFn is the function the StateTracker will use to initialize
	// the actors zero state.
	ActorsZeroStateFn func(*StateTracker, schema.Selector)
)

var (
	// type checks.
	_ ActorsZeroStateFn = (*StateTracker).ActorsZeroStateV0
	_ ActorsZeroStateFn = (*StateTracker).ActorsZeroStateV2
)

const (
	totalFilecoin     = 2_000_000_000
	filecoinPrecision = 1_000_000_000_000_000_000
)

var (
	TotalNetworkBalance = big.Mul(big.NewInt(totalFilecoin), big.NewInt(filecoinPrecision))
	RootVerifier        = MustNewIDAddr(80)
)

const (
	TestSealProofType = abi.RegisteredSealProof_StackedDrg2KiBV1
)

type ActorState struct {
	Addr    address.Address
	Balance abi.TokenAmount
	Code    cid.Cid
	State   cbor.Marshaler
}

func (st *StateTracker) ActorsZeroStateV0(selector schema.Selector) {
	store := st.Stores.ADTStore

	if _, err := store.Put(context.TODO(), miner0.ConstructDeadline(st.EmptyArrayCid)); err != nil {
		panic(err)
	}

	if _, err := store.Put(context.TODO(), bitfield.NewFromSet(nil)); err != nil {
		panic(err)
	}

	if _, err := store.Put(context.Background(), miner0.ConstructVestingFunds()); err != nil {
		panic(err)
	}

	var actorStates = []ActorState{{
		Addr:    builtin.InitActorAddr,
		Balance: big.Zero(),
		Code:    builtin.InitActorCodeID,
		State:   init0.ConstructState(st.EmptyMapCid, "chain-validation"),
	}, {
		Addr:    builtin.RewardActorAddr,
		Balance: TotalNetworkBalance,
		Code:    builtin.RewardActorCodeID,
		State:   reward0.ConstructState(big.Zero()),
	}, {
		Addr:    builtin.BurntFundsActorAddr,
		Balance: big.Zero(),
		Code:    builtin.AccountActorCodeID,
		State:   &account0.State{Address: builtin.BurntFundsActorAddr},
	}, {
		Addr:    builtin.StoragePowerActorAddr,
		Balance: big.Zero(),
		Code:    builtin.StoragePowerActorCodeID,
		State:   power0.ConstructState(st.EmptyMapCid, st.EmptyMultiMapCid),
	}, {
		Addr:    builtin.StorageMarketActorAddr,
		Balance: big.Zero(),
		Code:    builtin.StorageMarketActorCodeID,
		State:   market0.ConstructState(st.EmptyArrayCid, st.EmptyMapCid, st.EmptyMultiMapCid),
	}, {
		Addr:    builtin.SystemActorAddr,
		Balance: big.Zero(),
		Code:    builtin.SystemActorCodeID,
		State:   new(system0.State),
	}, {
		Addr:    builtin.CronActorAddr,
		Balance: big.Zero(),
		Code:    builtin.CronActorCodeID,
		State:   cron0.ConstructState(cron0.BuiltInEntries()),
	}, {
		Addr:    builtin.VerifiedRegistryActorAddr,
		Balance: big.Zero(),
		Code:    builtin.VerifiedRegistryActorCodeID,
		State:   verifreg0.ConstructState(st.EmptyMapCid, RootVerifier),
	}}

	// Add the chaos actor if this test requires it.
	if chaosOn, ok := selector["chaos_actor"]; ok && chaosOn == "true" {
		actorStates = append(actorStates, ActorState{
			Addr:    chaos.Address,
			Balance: big.Zero(),
			Code:    chaos.ChaosActorCodeCID,
			State:   &chaos.State{},
		})
	}

	for _, act := range actorStates {
		_ = st.CreateActor(act.Code, act.Addr, act.Balance, act.State)
	}
}

func (st *StateTracker) ActorsZeroStateV2(selector schema.Selector) {
	store := st.Stores.ADTStore

	if _, err := store.Put(context.TODO(), miner2.ConstructDeadline(st.EmptyArrayCid)); err != nil {
		panic(err)
	}

	if _, err := store.Put(context.Background(), miner2.ConstructVestingFunds()); err != nil {
		panic(err)
	}

	var actorStates = []ActorState{{
		Addr:    builtin2.InitActorAddr,
		Balance: big.Zero(),
		Code:    builtin2.InitActorCodeID,
		State:   init2.ConstructState(st.EmptyMapCid, "chain-validation"),
	}, {
		Addr:    builtin2.RewardActorAddr,
		Balance: TotalNetworkBalance,
		Code:    builtin2.RewardActorCodeID,
		State:   reward2.ConstructState(big.Zero()),
	}, {
		Addr:    builtin2.BurntFundsActorAddr,
		Balance: big.Zero(),
		Code:    builtin2.AccountActorCodeID,
		State:   &account2.State{Address: builtin2.BurntFundsActorAddr},
	}, {
		Addr:    builtin2.StoragePowerActorAddr,
		Balance: big.Zero(),
		Code:    builtin2.StoragePowerActorCodeID,
		State:   power2.ConstructState(st.EmptyMapCid, st.EmptyMultiMapCid),
	}, {
		Addr:    builtin2.StorageMarketActorAddr,
		Balance: big.Zero(),
		Code:    builtin2.StorageMarketActorCodeID,
		State:   market2.ConstructState(st.EmptyArrayCid, st.EmptyMapCid, st.EmptyMultiMapCid),
	}, {
		Addr:    builtin2.SystemActorAddr,
		Balance: big.Zero(),
		Code:    builtin2.SystemActorCodeID,
		State:   new(system2.State),
	}, {
		Addr:    builtin2.CronActorAddr,
		Balance: big.Zero(),
		Code:    builtin2.CronActorCodeID,
		State:   cron2.ConstructState(cron2.BuiltInEntries()),
	}, {
		Addr:    builtin2.VerifiedRegistryActorAddr,
		Balance: big.Zero(),
		Code:    builtin2.VerifiedRegistryActorCodeID,
		State:   verifreg2.ConstructState(st.EmptyMapCid, RootVerifier),
	}}

	// Add the chaos actor if this test requires it.
	if chaosOn, ok := selector["chaos_actor"]; ok && chaosOn == "true" {
		actorStates = append(actorStates, ActorState{
			Addr:    chaos.Address,
			Balance: big.Zero(),
			Code:    chaos.ChaosActorCodeCID,
			State:   &chaos.State{},
		})
	}

	for _, act := range actorStates {
		_ = st.CreateActor(act.Code, act.Addr, act.Balance, act.State)
	}
}
