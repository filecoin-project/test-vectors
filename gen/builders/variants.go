package builders

import (
	"fmt"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/network"

	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/test-vectors/schema"
)

// ProtocolVersion represents a protocol upgrade we track.
type ProtocolVersion struct {
	// ID is the code name of the version. It is output as a selector
	// on vectors.
	ID string

	// Height is the height at which the version activates. It is used to
	// calculate the variant's epoch.
	FirstEpoch abi.ChainEpoch

	// Network is the network version. It is output on the variant's epoch.
	Network network.Version

	// StateTree is the state tree version.
	StateTree types.StateTreeVersion

	// Actors is the actors version.
	Actors actors.Version

	// ZeroStateTree is the constructor of the initial state tree, including
	// singleton system actors.
	ZeroStateTree func(*StateTracker, schema.Selector)
}

// KnownProtocolVersions enumerates the protocol versions we're capable of
// generating test vectors against.
var KnownProtocolVersions = []ProtocolVersion{
	{
		ID:            "genesis",
		FirstEpoch:    1,
		StateTree:     types.StateTreeVersion0,
		Network:       network.Version0,
		Actors:        actors.Version0,
		ZeroStateTree: (*StateTracker).ActorsZeroStateV0,
	},
	{
		ID:            "breeze",
		FirstEpoch:    build.UpgradeBreezeHeight + 1,
		StateTree:     types.StateTreeVersion0,
		Network:       network.Version1,
		Actors:        actors.Version0,
		ZeroStateTree: (*StateTracker).ActorsZeroStateV0,
	}, {
		ID:            "smoke",
		FirstEpoch:    build.UpgradeSmokeHeight + 1,
		StateTree:     types.StateTreeVersion0,
		Network:       network.Version2,
		Actors:        actors.Version0,
		ZeroStateTree: (*StateTracker).ActorsZeroStateV0,
	}, {
		ID:            "ignition",
		FirstEpoch:    build.UpgradeIgnitionHeight + 1,
		StateTree:     types.StateTreeVersion0,
		Network:       network.Version3,
		Actors:        actors.Version0,
		ZeroStateTree: (*StateTracker).ActorsZeroStateV0,
	}, {
		ID:            "actorsv2",
		FirstEpoch:    build.UpgradeActorsV2Height + 1,
		StateTree:     types.StateTreeVersion1,
		Network:       network.Version4,
		Actors:        actors.Version2,
		ZeroStateTree: (*StateTracker).ActorsZeroStateV2,
	},
	{
		ID:            "tape",
		FirstEpoch:    build.UpgradeTapeHeight + 1,
		StateTree:     types.StateTreeVersion1,
		Network:       network.Version5,
		Actors:        actors.Version2,
		ZeroStateTree: (*StateTracker).ActorsZeroStateV2,
	},
	{
		ID:            "liftoff",
		FirstEpoch:    build.UpgradeLiftoffHeight + 1,
		StateTree:     types.StateTreeVersion1,
		Network:       network.Version5,
		Actors:        actors.Version2,
		ZeroStateTree: (*StateTracker).ActorsZeroStateV2,
	},
}

// KnownProtocolVersionsFrom returns all protocol versions known starting from
// the protocol version with the supplied ID, it inclusive.
func KnownProtocolVersionsFrom(id string) []ProtocolVersion {
	for i, pv := range KnownProtocolVersions {
		if pv.ID == id {
			return KnownProtocolVersions[i:]
		}
	}
	panic(fmt.Sprintf("unknown protocol version: %s", id))
}

// KnownProtocolVersionsBefore returns all protocol versions known until the
// protocol version with the supplied ID, it exclusive.
func KnownProtocolVersionsBefore(id string) []ProtocolVersion {
	for i, pv := range KnownProtocolVersions {
		if pv.ID == id {
			return KnownProtocolVersions[0:i]
		}
	}
	panic(fmt.Sprintf("unknown protocol version: %s", id))
}

// KnownProtocolVersionsBetween returns all protocol versions known between the
// supplied range, both inclusive.
func KnownProtocolVersionsBetween(from, to string) []ProtocolVersion {
	start, end := -1, -1
	for i, pv := range KnownProtocolVersions {
		if pv.ID == from {
			start = i
		}
		if pv.ID == to {
			end = i
		}
	}
	if start == -1 || end == -1 {
		panic(fmt.Sprintf("at least one unknown protocol version: %s, %s", from, to))
	}
	return KnownProtocolVersions[start : end+1]
}
