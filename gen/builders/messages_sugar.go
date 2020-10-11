package builders

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/actors/builtin/paych"
)

type sugarMsg struct{ m *Messages }

// Transfer enlists a value transfer message.
func (s *sugarMsg) Transfer(from, to address.Address, opts ...MsgOpt) *ApplicableMessage {
	return s.m.Typed(from, to, Transfer(), opts...)
}

func (s *sugarMsg) PaychMessage(from address.Address, builderFn func(paych.MessageBuilder) (*types.Message, error), opts ...MsgOpt) *ApplicableMessage {
	builder := paych.Message(s.m.st.ActorsVersion, from)
	msg, err := builderFn(builder)
	s.m.bc.Assert.NoError(err, "failed to create paych message")
	return s.m.Message(msg, opts...)
}

func (s *sugarMsg) MultisigMessage(from address.Address, builderFn func(multisig.MessageBuilder) (*types.Message, error), opts ...MsgOpt) *ApplicableMessage {
	builder := multisig.Message(s.m.st.ActorsVersion, from)
	msg, err := builderFn(builder)
	s.m.bc.Assert.NoError(err, "failed to create multisig message")
	return s.m.Message(msg, opts...)
}
