package responses

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

// DecodedLogTopic represents a single decoded indexed event parameter.
type DecodedLogTopic struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// DecodedLog represents the decoded output for an event log.
type DecodedLog struct {
	Name   string            `json:"name"`
	Topics []DecodedLogTopic `json:"topics"`
	Data   map[string]any    `json:"data"`
}

// DecodedTrace represents the decoded output for a trace input.
type DecodedTrace struct {
	Method string         `json:"method"`
	Args   map[string]any `json:"args"`
}

// parseABI parses a JSON ABI into a go-ethereum ABI struct.
// Returns nil if the input is empty or parsing fails.
func parseABI(abiJSON json.RawMessage) *abi.ABI {
	if len(abiJSON) == 0 {
		return nil
	}
	parsed, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return nil
	}
	return &parsed
}

// decodeLogWithABI decodes a log using an already-parsed ABI.
// Returns nil if abi is nil, topics are empty, or decoding fails.
func decodeLogWithABI(contractABI *abi.ABI, data []byte, topics []pkgTypes.Hex) *DecodedLog {
	if contractABI == nil || len(topics) == 0 {
		return nil
	}

	topic0 := common.BytesToHash(topics[0])
	event, err := contractABI.EventByID(topic0)
	if err != nil {
		return nil
	}

	// Decode indexed parameters from topics[1:]
	indexedArgs := make(abi.Arguments, 0)
	for i := range event.Inputs {
		if event.Inputs[i].Indexed {
			indexedArgs = append(indexedArgs, event.Inputs[i])
		}
	}

	topicHashes := make([]common.Hash, 0, len(topics)-1)
	for _, t := range topics[1:] {
		topicHashes = append(topicHashes, common.BytesToHash(t))
	}

	result := &DecodedLog{
		Name:   event.Name,
		Topics: make([]DecodedLogTopic, 0, len(indexedArgs)),
		Data:   make(map[string]any),
	}

	indexedValues := make(map[string]any, len(indexedArgs))
	if err := abi.ParseTopicsIntoMap(indexedValues, indexedArgs, topicHashes); err == nil {
		for i := range indexedArgs {
			if val, ok := indexedValues[indexedArgs[i].Name]; ok {
				result.Topics = append(result.Topics, DecodedLogTopic{
					Name:  indexedArgs[i].Name,
					Value: formatABIValue(val),
				})
			}
		}
	}

	// Decode non-indexed parameters from data
	if len(data) > 0 {
		if err := contractABI.UnpackIntoMap(result.Data, event.Name, data); err != nil {
			log.Err(err).Msg("unpack into map")
		}
	}

	return result
}

// decodeTxArgs decodes a trace's input data using an already-parsed ABI.
// Returns nil if abi is nil, input is shorter than 4 bytes, or decoding fails.
func decodeTxArgs(contractABI *abi.ABI, input []byte) *DecodedTrace {
	if contractABI == nil || len(input) < 4 {
		return nil
	}

	method, err := contractABI.MethodById(input[:4])
	if err != nil {
		return nil
	}

	args := make(map[string]any, len(method.Inputs))

	if len(input) > 4 {
		values, err := method.Inputs.Unpack(input[4:])
		if err != nil {
			return &DecodedTrace{
				Method: method.Name,
				Args:   args,
			}
		}

		for i, inp := range method.Inputs {
			if i < len(values) {
				args[inp.Name] = formatABIValue(values[i])
			}
		}
	}

	return &DecodedTrace{
		Method: method.Name,
		Args:   args,
	}
}

// formatABIValue converts ABI-decoded Go values to JSON-friendly representations.
func formatABIValue(v any) any {
	switch val := v.(type) {
	case common.Address:
		return val.Hex()
	case *common.Address:
		if val == nil {
			return nil
		}
		return val.Hex()
	case common.Hash:
		return val.Hex()
	case *big.Int:
		if val == nil {
			return "0"
		}
		return val.String()
	case []byte:
		return fmt.Sprintf("0x%x", val)
	case [32]byte:
		return fmt.Sprintf("0x%x", val[:])
	case bool:
		return val
	case string:
		return val
	case uint8:
		return val
	case uint16:
		return val
	case uint32:
		return val
	case uint64:
		return val
	case int8:
		return val
	case int16:
		return val
	case int32:
		return val
	case int64:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}
