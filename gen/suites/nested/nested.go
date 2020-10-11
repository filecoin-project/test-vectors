package main

import (
	"bytes"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"

	init0 "github.com/filecoin-project/specs-actors/actors/builtin/init"
	paych0 "github.com/filecoin-project/specs-actors/actors/builtin/paych"
	reward0 "github.com/filecoin-project/specs-actors/actors/builtin/reward"
	multisig0 "github.com/filecoin-project/specs-actors/v2/actors/builtin/multisig"

	typegen "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/lotus/conformance/chaos"

	. "github.com/filecoin-project/test-vectors/gen/builders"
)

var (
	acctDefaultBalance = abi.NewTokenAmount(1_000_000_000_000)
	multisigBalance    = abi.NewTokenAmount(1_000_000_000)
	nonce              = uint64(1)
)

func nestedSends_OkBasic(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	balanceBefore := v.StateTracker.Balance(stage.creator)

	// Multisig sends back to the creator.
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(stage.creator, amtSent, builtin.MethodSend, nil, nonce)

	//td.AssertActor(stage.creator, big.Sub(big.Add(balanceBefore, amtSent), result.Result.Receipt.GasUsed.Big()), nonce+1)
	v.Assert.NonceEq(stage.creator, nonce+1)
	v.Assert.BalanceEq(stage.creator, big.Sub(big.Add(balanceBefore, amtSent), CalculateSenderDeduction(result)))
}

func nestedSends_OkToNewActor(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	balanceBefore := v.StateTracker.Balance(stage.creator)

	// Multisig sends to new address.
	newAddr := v.Wallet.NewSECP256k1Account()
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(newAddr, amtSent, builtin.MethodSend, nil, nonce)

	v.Assert.BalanceEq(stage.msAddr, big.Sub(multisigBalance, amtSent))
	v.Assert.BalanceEq(stage.creator, big.Sub(balanceBefore, CalculateSenderDeduction(result)))
	v.Assert.BalanceEq(newAddr, amtSent)
}

func nestedSends_OkToNewActorWithInvoke(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	balanceBefore := v.StateTracker.Balance(stage.creator)

	// Multisig sends to new address and invokes pubkey method at the same time.
	newAddr := v.Wallet.NewSECP256k1Account()
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(newAddr, amtSent, builtin.MethodsAccount.PubkeyAddress, nil, nonce)
	// TODO: use an explicit Approve() and check the return value is the correct pubkey address
	// when the multisig Approve() method plumbs through the inner exit code and value.
	// https://github.com/filecoin-project/specs-actors/issues/113
	//expected := bytes.Buffer{}
	//require.NoError(t, newAddr.MarshalCBOR(&expected))
	//assert.Equal(t, expected.Bytes(), result.Result.Receipt.ReturnValue)

	v.Assert.BalanceEq(stage.msAddr, big.Sub(multisigBalance, amtSent))
	v.Assert.BalanceEq(stage.creator, big.Sub(balanceBefore, CalculateSenderDeduction(result)))
	v.Assert.BalanceEq(newAddr, amtSent)
}

func nestedSends_OkRecursive(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	another := v.Actors.Account(address.SECP256K1, big.Zero())
	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	balanceBefore := v.StateTracker.Balance(stage.creator)

	// Multisig sends to itself.
	params := multisig0.AddSignerParams{
		Signer:   another.ID,
		Increase: false,
	}
	result := stage.sendOk(stage.msAddr, big.Zero(), builtin.MethodsMultisig.AddSigner, &params, nonce)

	v.Assert.BalanceEq(stage.msAddr, multisigBalance)
	v.Assert.Equal(big.Sub(balanceBefore, CalculateSenderDeduction(result)), v.StateTracker.Balance(stage.creator))

	var st multisig0.State
	v.StateTracker.ActorState(stage.msAddr, &st)
	v.Assert.Equal([]address.Address{stage.creator, another.ID}, st.Signers)
}

func nestedSends_OKNonCBORParamsWithTransfer(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)

	newAddr := v.Wallet.NewSECP256k1Account()
	amtSent := abi.NewTokenAmount(1)
	// So long as the parameters are not actually used by the method, a message can carry arbitrary bytes.
	params := typegen.Deferred{Raw: []byte{1, 2, 3, 4}}
	stage.sendOk(newAddr, amtSent, builtin.MethodSend, &params, nonce)

	v.Assert.BalanceEq(stage.msAddr, big.Sub(multisigBalance, amtSent))
	v.Assert.BalanceEq(newAddr, amtSent)
}

func nestedSends_FailNonexistentIDAddress(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)

	newAddr := MustNewIDAddr(1234)
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(newAddr, amtSent, builtin.MethodSend, nil, nonce)

	var ret multisig0.ProposeReturn
	MustDeserialize(result.Result.Return, &ret)
	v.Assert.ExitCodeEq(ret.Code, exitcode.SysErrInvalidReceiver)

	v.Assert.BalanceEq(stage.msAddr, multisigBalance) // No change.
	v.Assert.ActorMissing(newAddr)
}

func nestedSends_FailNonexistentActorAddress(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)

	newAddr := MustNewActorAddr("1234")
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(newAddr, amtSent, builtin.MethodSend, nil, nonce)

	var ret multisig0.ProposeReturn
	MustDeserialize(result.Result.Return, &ret)
	v.Assert.ExitCodeEq(ret.Code, exitcode.SysErrInvalidReceiver)

	v.Assert.BalanceEq(stage.msAddr, multisigBalance) // No change.
	v.Assert.ActorMissing(newAddr)
}

func nestedSends_FailInvalidMethodNumNewActor(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)

	newAddr := v.Wallet.NewSECP256k1Account()
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(newAddr, amtSent, abi.MethodNum(99), nil, nonce)

	var ret multisig0.ProposeReturn
	MustDeserialize(result.Result.Return, &ret)
	v.Assert.ExitCodeEq(ret.Code, exitcode.SysErrInvalidMethod)

	v.Assert.BalanceEq(stage.msAddr, multisigBalance) // No change.
	v.Assert.ActorMissing(newAddr)
}

func nestedSends_FailInvalidMethodNumForActor(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	balanceBefore := v.StateTracker.Balance(stage.creator)

	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(stage.creator, amtSent, abi.MethodNum(99), nil, nonce)

	var ret multisig0.ProposeReturn
	MustDeserialize(result.Result.Return, &ret)
	v.Assert.ExitCodeEq(ret.Code, exitcode.SysErrInvalidMethod)

	v.Assert.BalanceEq(stage.msAddr, multisigBalance)                                           // No change.
	v.Assert.BalanceEq(stage.creator, big.Sub(balanceBefore, CalculateSenderDeduction(result))) // Pay gas, don't receive funds.
}

func nestedSends_FailMissingParams(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	balanceBefore := v.StateTracker.Balance(stage.creator)

	params := abi.Empty // Missing params required by AddSigner
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(stage.msAddr, amtSent, builtin.MethodsMultisig.AddSigner, params, nonce)

	var ret multisig0.ProposeReturn
	MustDeserialize(result.Result.Return, &ret)
	v.Assert.ExitCodeEq(ret.Code, exitcode.ErrSerialization)

	v.Assert.BalanceEq(stage.creator, big.Sub(balanceBefore, CalculateSenderDeduction(result)))
	v.Assert.BalanceEq(stage.msAddr, multisigBalance) // No change.
	signers, _ := stage.state().Signers()
	v.Assert.Equal(1, len(signers)) // No new signers
}

func nestedSends_FailMismatchParams(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	balanceBefore := v.StateTracker.Balance(stage.creator)

	// Wrong params for AddSigner
	params := multisig0.ProposeParams{
		To:     stage.creator,
		Value:  big.Zero(),
		Method: builtin.MethodSend,
		Params: nil,
	}
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(stage.msAddr, amtSent, builtin.MethodsMultisig.AddSigner, &params, nonce)

	var ret multisig0.ProposeReturn
	MustDeserialize(result.Result.Return, &ret)
	v.Assert.ExitCodeEq(ret.Code, exitcode.ErrSerialization)

	v.Assert.BalanceEq(stage.creator, big.Sub(balanceBefore, CalculateSenderDeduction(result)))
	v.Assert.BalanceEq(stage.msAddr, multisigBalance) // No change.
	signers, _ := stage.state().Signers()
	v.Assert.Equal(1, len(signers)) // No new signers
}

func nestedSends_FailInnerAbort(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	prevHead := v.StateTracker.Head(builtin.RewardActorAddr)

	// AwardBlockReward will abort unless invoked by the system actor
	params := reward0.AwardBlockRewardParams{
		Miner:     stage.creator,
		Penalty:   big.Zero(),
		GasReward: big.Zero(),
	}
	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(builtin.RewardActorAddr, amtSent, builtin.MethodsReward.AwardBlockReward, &params, nonce)

	var ret multisig0.ProposeReturn
	MustDeserialize(result.Result.Return, &ret)
	v.Assert.ExitCodeEq(ret.Code, exitcode.SysErrForbidden)

	v.Assert.BalanceEq(stage.msAddr, multisigBalance) // No change.
	v.Assert.HeadEq(builtin.RewardActorAddr, prevHead)
}

func nestedSends_FailAbortedExec(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	stage := prepareStage(v, acctDefaultBalance, multisigBalance)
	prevHead := v.StateTracker.Head(builtin.InitActorAddr)

	// Illegal paych constructor params (addresses are not accounts)
	ctorParams := paych0.ConstructorParams{
		From: builtin.SystemActorAddr,
		To:   builtin.SystemActorAddr,
	}
	execParams := init0.ExecParams{
		CodeCID:           builtin.PaymentChannelActorCodeID,
		ConstructorParams: MustSerialize(&ctorParams),
	}

	amtSent := abi.NewTokenAmount(1)
	result := stage.sendOk(builtin.InitActorAddr, amtSent, builtin.MethodsInit.Exec, &execParams, nonce)

	var ret multisig0.ProposeReturn
	MustDeserialize(result.Result.Return, &ret)
	v.Assert.ExitCodeEq(ret.Code, exitcode.ErrForbidden)

	v.Assert.BalanceEq(stage.msAddr, multisigBalance) // No change.
	v.Assert.HeadEq(builtin.InitActorAddr, prevHead)  // Init state unchanged.
}

func nestedSends_FailInsufficientFundsForTransferInInnerSend(v *MessageVectorBuilder) {
	v.Messages.SetDefaults(GasLimit(1_000_000_000), GasPremium(1), GasFeeCap(200))

	alice := v.Actors.Account(address.SECP256K1, acctDefaultBalance)
	bob := v.Actors.Account(address.SECP256K1, big.Zero())

	v.CommitPreconditions()

	// alice tells the chaos actor to send funds to bob, the chaos actor has 0 balance so the inner send will fail,
	// and alice will pay the gas cost.
	amtSent := abi.NewTokenAmount(1)
	msg := v.Messages.Typed(alice.ID, chaos.Address, ChaosSend(&chaos.SendArgs{
		To:     bob.ID,
		Value:  amtSent,
		Method: builtin.MethodSend,
		Params: nil,
	}), Nonce(0), Value(big.Zero()))

	v.Messages.ApplyOne(msg)

	v.CommitApplies()

	// the outer message should be applied successfully
	v.Assert.Equal(exitcode.Ok, msg.Result.ExitCode)

	var chaosRet chaos.SendReturn
	MustDeserialize(msg.Result.MessageReceipt.Return, &chaosRet)

	// the inner message should fail
	v.Assert.Equal(exitcode.SysErrInsufficientFunds, chaosRet.Code)

	// alice should be charged for the gas cost and bob should have not received any funds.
	v.Assert.MessageSendersSatisfy(BalanceUpdated(big.Zero()), msg)
	v.Assert.BalanceEq(bob.ID, big.Zero())
}

type msStage struct {
	v       *MessageVectorBuilder
	creator address.Address // Address of the creator and sole signer of the multisig.
	msAddr  address.Address // Address of the multisig actor from which nested messages are sent.
}

// Creates a multisig actor with its creator as sole approver.
func prepareStage(v *MessageVectorBuilder, creatorBalance, msBalance abi.TokenAmount) *msStage {
	// Set up sender and receiver accounts.
	creator := v.Actors.Account(address.SECP256K1, creatorBalance)
	v.CommitPreconditions()

	msg := v.Messages.Sugar().MultisigMessage(creator.ID, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Create([]address.Address{creator.ID}, 1, 0, 0, msBalance)
	}, Value(msBalance), Nonce(0))
	v.Messages.ApplyOne(msg)

	v.Assert.Equal(msg.Result.ExitCode, exitcode.Ok)

	// Verify init actor return.
	var ret init0.ExecReturn
	MustDeserialize(msg.Result.Return, &ret)

	return &msStage{
		v:       v,
		creator: creator.ID,
		msAddr:  ret.IDAddress,
	}
}

func (s *msStage) sendOk(to address.Address, value abi.TokenAmount, method abi.MethodNum, params cbor.Marshaler, approverNonce uint64) *ApplicableMessage {
	buf := bytes.Buffer{}
	if params != nil {
		err := params.MarshalCBOR(&buf)
		if err != nil {
			panic(err)
		}
	}

	msg := s.v.Messages.Sugar().MultisigMessage(s.creator, func(b multisig.MessageBuilder) (*types.Message, error) {
		return b.Propose(s.msAddr, to, value, method, buf.Bytes())
	}, Nonce(approverNonce), Value(big.Zero()))
	s.v.CommitApplies()

	// all messages succeeded.
	s.v.Assert.EveryMessageResultSatisfies(ExitCode(exitcode.Ok))

	return msg
}

func (s *msStage) state() multisig.State {
	state, err := multisig.Load(s.v.StateTracker.Stores.ADTStore, s.v.StateTracker.Header(s.msAddr))
	s.v.Assert.NoError(err)
	return state
}
