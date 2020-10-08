package schema

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRandomnessCircularSerde(t *testing.T) {
	tv1 := TestVector{
		Randomness: Randomness{
			{
				On: RandomnessRule{
					Kind:                RandomnessBeacon,
					DomainSeparationTag: 5,
					Epoch:               10,
					Entropy:             []byte("hello world!"),
				},
				Return: []byte("super random"),
			},
			{
				On: RandomnessRule{
					Kind:                RandomnessChain,
					DomainSeparationTag: 99,
					Epoch:               68592,
					Entropy:             nil, // no entropy
				},
				Return: []byte("another random value"),
			},
		},
	}

	serialized, err := json.Marshal(tv1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(serialized))

	var tv2 TestVector
	err = json.Unmarshal(serialized, &tv2)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(tv1.Randomness, tv2.Randomness) {
		t.Fatal("values not equal")
	}

}
