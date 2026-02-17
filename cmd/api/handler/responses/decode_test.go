package responses

import (
	"encoding/json"
	"testing"

	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/stretchr/testify/require"
)

// ERC20 transfer(address,uint256) ABI
var erc20ABI = json.RawMessage(`[{"inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`)

func TestParseABI(t *testing.T) {
	t.Run("valid ABI", func(t *testing.T) {
		parsed := parseABI(erc20ABI)
		require.NotNil(t, parsed)
		require.Contains(t, parsed.Methods, "transfer")
	})

	t.Run("nil ABI", func(t *testing.T) {
		parsed := parseABI(nil)
		require.Nil(t, parsed)
	})

	t.Run("empty ABI", func(t *testing.T) {
		parsed := parseABI(json.RawMessage{})
		require.Nil(t, parsed)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		parsed := parseABI(json.RawMessage(`invalid`))
		require.Nil(t, parsed)
	})
}

func TestDecodeTxArgs(t *testing.T) {
	parsed := parseABI(erc20ABI)
	require.NotNil(t, parsed)

	t.Run("valid transfer input", func(t *testing.T) {
		// transfer(address,uint256) selector = 0xa9059cbb
		// to = 0x000000000000000000000000dead000000000000000000000000000000000000
		// amount = 1000 (0x3e8)
		input := pkgTypes.MustDecodeHex(
			"a9059cbb" +
				"000000000000000000000000dead000000000000000000000000000000000000" +
				"00000000000000000000000000000000000000000000000000000000000003e8",
		)
		decoded := decodeTxArgs(parsed, input)
		require.NotNil(t, decoded)
		require.Equal(t, "transfer", decoded.Method)
		require.Contains(t, decoded.Args, "to")
		require.Contains(t, decoded.Args, "amount")
		require.Equal(t, "0xdEad000000000000000000000000000000000000", decoded.Args["to"])
		require.Equal(t, "1000", decoded.Args["amount"])
	})

	t.Run("nil ABI", func(t *testing.T) {
		input := []byte{0xa9, 0x05, 0x9c, 0xbb}
		decoded := decodeTxArgs(nil, input)
		require.Nil(t, decoded)
	})

	t.Run("input shorter than 4 bytes", func(t *testing.T) {
		decoded := decodeTxArgs(parsed, []byte{0xa9, 0x05})
		require.Nil(t, decoded)
	})

	t.Run("unknown selector", func(t *testing.T) {
		decoded := decodeTxArgs(parsed, []byte{0x00, 0x00, 0x00, 0x00})
		require.Nil(t, decoded)
	})

	t.Run("only selector without args", func(t *testing.T) {
		decoded := decodeTxArgs(parsed, []byte{0xa9, 0x05, 0x9c, 0xbb})
		require.NotNil(t, decoded)
		require.Equal(t, "transfer", decoded.Method)
		require.Empty(t, decoded.Args)
	})
}

func TestDecodeLogWithABI(t *testing.T) {
	transferEventABI := json.RawMessage(`[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`)

	parsed := parseABI(transferEventABI)
	require.NotNil(t, parsed)

	t.Run("nil ABI", func(t *testing.T) {
		decoded := decodeLogWithABI(nil, []byte{}, []pkgTypes.Hex{{0x01}})
		require.Nil(t, decoded)
	})

	t.Run("empty topics", func(t *testing.T) {
		decoded := decodeLogWithABI(parsed, []byte{}, nil)
		require.Nil(t, decoded)
	})
}
