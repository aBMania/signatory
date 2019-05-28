package tezos_test

import (
	"fmt"
	"testing"

	"github.com/ecadlabs/signatory/config"
	"github.com/ecadlabs/signatory/tezos"
)

func TestValidateMessage(t *testing.T) {
	type Case struct {
		Name    string
		Message []byte
		Error   error
	}

	cases := []Case{
		Case{Name: "Nil message", Message: nil, Error: tezos.ErrMessageEmpty},
		Case{Name: "Empty message", Message: []byte{}, Error: tezos.ErrMessageEmpty},
		Case{Name: "Generic operation", Message: []byte{0x03, 0x02}, Error: nil},
		Case{Name: "Endorsement operation", Message: []byte{0x02, 0x02}, Error: nil},
		Case{Name: "Block operation", Message: []byte{0x01, 0x02}, Error: nil},
		Case{Name: "Invalid magic byte", Message: []byte{0x00, 0x02}, Error: tezos.ErrInvalidMagicByte},
		Case{Name: "Invalid magic byte", Message: []byte{0x04, 0x02}, Error: tezos.ErrInvalidMagicByte},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			msg := tezos.ParseMessage(c.Message)
			err := msg.Validate()

			if err != c.Error {
				fmt.Printf("Expected %v but got %v\n", c.Error, err)
				t.Fail()
			}
		})
	}
}

func TestGetMessageType(t *testing.T) {
	type Case struct {
		Name    string
		Message []byte
		Type    string
	}

	cases := []Case{
		Case{Name: "Generic operation", Message: []byte{0x03, 0x02}, Type: tezos.OpGeneric},
		Case{Name: "Endorsement operation", Message: []byte{0x02, 0x02}, Type: tezos.OpEndorsement},
		Case{Name: "Block operation", Message: []byte{0x01, 0x02}, Type: tezos.OpBlock},
		Case{Name: "Unknown operation", Message: []byte{0x05, 0x02}, Type: tezos.OpUnknown},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			msg := tezos.ParseMessage(c.Message)
			msgType := msg.Type()

			if msgType != c.Type {
				fmt.Printf("Expected %v but got %v\n", c.Type, msgType)
				t.Fail()
			}
		})
	}
}

func TestFilterMessage(t *testing.T) {
	type Case struct {
		Name    string
		Message []byte
		Config  *config.TezosConfig
		Error   error
	}

	genTezosConfig := func(filters []string) *config.TezosConfig {
		return &config.TezosConfig{
			AllowedOperations: filters,
		}
	}

	cases := []Case{
		Case{Name: "Nil message", Message: nil, Error: tezos.ErrDoNotMatchFilter},
		Case{Name: "Empty message", Message: []byte{}, Error: tezos.ErrDoNotMatchFilter},
		Case{Name: "Generic operation", Message: []byte{0x03, 0x02}, Error: nil, Config: genTezosConfig([]string{tezos.OpGeneric})},
		Case{Name: "Endorsement operation", Message: []byte{0x02, 0x02}, Error: nil, Config: genTezosConfig([]string{tezos.OpEndorsement})},
		Case{Name: "Block operation", Message: []byte{0x01, 0x02}, Error: nil, Config: genTezosConfig([]string{tezos.OpBlock})},
		Case{Name: "Invalid magic byte", Message: []byte{0x00, 0x02}, Error: tezos.ErrDoNotMatchFilter},
		Case{Name: "Invalid magic byte", Message: []byte{0x04, 0x02}, Error: tezos.ErrDoNotMatchFilter},
		Case{Name: "Unsupported operation", Message: []byte{0x03, 0x02}, Error: tezos.ErrDoNotMatchFilter, Config: genTezosConfig([]string{tezos.OpBlock, tezos.OpEndorsement})},
		Case{Name: "Unsupported operation", Message: []byte{0x01, 0x02}, Error: tezos.ErrDoNotMatchFilter, Config: genTezosConfig([]string{tezos.OpGeneric})},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			msg := tezos.ParseMessage(c.Message)
			err := msg.MatchFilter(c.Config)

			if err != c.Error {
				fmt.Printf("Expected %v but got %v\n", c.Error, err)
				t.Fail()
			}
		})
	}
}
