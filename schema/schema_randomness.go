package schema

import "encoding/json"

// RandomnessKind specifies the type of randomness that is being requested.
type RandomnessKind string

const (
	RandomnessBeacon = RandomnessKind("beacon")
	RandomnessChain  = RandomnessKind("chain")
)

// RandomnessRule represents a rule to evaluate randomness matches against.
// This encodes to JSON as an array. See godocs on the Randomness type for
// more info.
type RandomnessRule struct {
	Kind                RandomnessKind
	DomainSeparationTag int64
	Epoch               int64
	Entropy             Base64EncodedBytes
}

func (rm RandomnessRule) MarshalJSON() ([]byte, error) {
	array := [4]interface{}{
		rm.Kind,
		rm.DomainSeparationTag,
		rm.Epoch,
		rm.Entropy,
	}
	return json.Marshal(array)
}

func (rm *RandomnessRule) UnmarshalJSON(v []byte) error {
	var (
		arr [4]json.RawMessage
		out RandomnessRule
		err error
	)
	if err = json.Unmarshal(v, &arr); err != nil {
		return err
	}
	if err = json.Unmarshal(arr[0], &out.Kind); err != nil {
		return err
	}
	if err = json.Unmarshal(arr[1], &out.DomainSeparationTag); err != nil {
		return err
	}
	if err = json.Unmarshal(arr[2], &out.Epoch); err != nil {
		return err
	}
	if err = json.Unmarshal(arr[3], &out.Entropy); err != nil {
		return err
	}
	*rm = out
	return nil
}

// Randomness encodes randomness the VM runtime should return while executing
// this vector. It is encoded as a list of ordered rules to match on.
//
// The json serialized form is:
//
//  "randomness": [
//    { "on": ["beacon", 12, 49327, "yxpTbzLhr4uaj7bK0Hl4Vw=="], "ret": "iKyZ2N83N8IoiK2tNJ/H9g==" },
//    { "on": ["chain", 8, 61002, "aacQWICNcMJWtuwTnU+1Hg=="], "ret": "M6HqmihwZ5fXcbQQHhbtsg==" }
//  ]
//
// The four positional values of the `on` array field are:
//
//  1. Kind of randomness (json string; values: beacon, chain).
//  2. Domain separation tag (json number).
//  3. EpochOffset (json number).
//  4. Entropy (json string; base64 encoded bytes).
//
// When no rules are matched, the driver should return the raw bytes of
// utf-8 string 'i_am_random_____i_am_random_____'
type Randomness []RandomnessMatch

// RandomnessMatch specifies a randomness match. When the implementation
// requests randomness that matches the RandomnessRule in On, Return will
// be returned.
type RandomnessMatch struct {
	On     RandomnessRule     `json:"on"`
	Return Base64EncodedBytes `json:"ret"`
}
