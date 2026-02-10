package parser

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/NobleScope/noble-indexer/internal/storage/types"
	"github.com/NobleScope/noble-indexer/pkg/indexer/config"
	dCtx "github.com/NobleScope/noble-indexer/pkg/indexer/decode/context"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
	// ERC20 Transfer event topic
	erc20TransferTopic = pkgTypes.MustDecodeHex("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	// Test addresses (32-byte padded)
	zeroAddressPadded    = pkgTypes.MustDecodeHex("0x0000000000000000000000000000000000000000000000000000000000000000")
	fromAddressPadded    = pkgTypes.MustDecodeHex("0x000000000000000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	toAddressPadded      = pkgTypes.MustDecodeHex("0x000000000000000000000000bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	contractAddressBytes = pkgTypes.MustDecodeHex("0xcccccccccccccccccccccccccccccccccccccccc")

	// Precompiled contract address
	precompiledAddressBytes = pkgTypes.MustDecodeHex("0xdddddddddddddddddddddddddddddddddddddddd")
	precompiledAddressHex   = "0xdddddddddddddddddddddddddddddddddddddddd"
)

func createTestModule(t *testing.T, precompiledAddresses []string) *Module {
	t.Helper()

	erc20ABI, err := abi.JSON(bytes.NewReader([]byte(`[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`)))
	require.NoError(t, err)

	erc721ABI, err := abi.JSON(bytes.NewReader([]byte(`[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":true,"name":"tokenId","type":"uint256"}],"name":"Transfer","type":"event"}]`)))
	require.NoError(t, err)

	erc1155ABI, err := abi.JSON(bytes.NewReader([]byte(`[{"anonymous":false,"inputs":[{"indexed":true,"name":"operator","type":"address"},{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"id","type":"uint256"},{"indexed":false,"name":"value","type":"uint256"}],"name":"TransferSingle","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"operator","type":"address"},{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"ids","type":"uint256[]"},{"indexed":false,"name":"values","type":"uint256[]"}],"name":"TransferBatch","type":"event"}]`)))
	require.NoError(t, err)

	precompiled := make(map[string]struct{}, len(precompiledAddresses))
	for _, addr := range precompiledAddresses {
		precompiled[addr] = struct{}{}
	}

	return &Module{
		cfg:                  config.Indexer{},
		networkConfig:        config.Network{PrecompiledContracts: precompiledAddresses},
		precompiledContracts: precompiled,
		abi: map[types.TokenType]*abi.ABI{
			types.ERC20:   &erc20ABI,
			types.ERC721:  &erc721ABI,
			types.ERC1155: &erc1155ABI,
		},
	}
}

func encodeUint256(val *big.Int) []byte {
	b := val.Bytes()
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

func createERC20TransferLog(contractAddr pkgTypes.Hex, from, to pkgTypes.Hex, amount *big.Int) *storage.Log {
	return &storage.Log{
		Address: storage.Address{Hash: contractAddr},
		Topics:  []pkgTypes.Hex{erc20TransferTopic, from, to},
		Data:    encodeUint256(amount),
	}
}

func TestParseTransfers_RevertedTxSkipped(t *testing.T) {
	module := createTestModule(t, nil)

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusRevert,
				Logs: []*storage.Log{
					createERC20TransferLog(
						contractAddressBytes,
						fromAddressPadded,
						toAddressPadded,
						big.NewInt(1000),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	// Transfers slice should not be initialized for reverted tx
	require.Nil(t, ctx.Block.Txs[0].Transfers)
	// No tokens should be added
	require.Empty(t, ctx.GetTokens())
}

func TestParseTransfers_ERC20Transfer_RegularContract(t *testing.T) {
	module := createTestModule(t, nil)

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					createERC20TransferLog(
						contractAddressBytes,
						fromAddressPadded,
						toAddressPadded,
						big.NewInt(1000),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	// Should have 1 transfer
	require.Len(t, ctx.Block.Txs[0].Transfers, 1)

	transfer := ctx.Block.Txs[0].Transfers[0]
	require.Equal(t, types.Transfer, transfer.Type)
	require.Equal(t, decimal.NewFromInt(1000), transfer.Amount)
	require.Equal(t, decimal.Zero, transfer.TokenID)
	require.NotNil(t, transfer.FromAddress)
	require.NotNil(t, transfer.ToAddress)

	// Should add token for regular contract
	tokens := ctx.GetTokens()
	require.Len(t, tokens, 1)
	require.Equal(t, types.ERC20, tokens[0].Type)

	// Should add token balances for regular contract
	tokenBalances := ctx.GetTokenBalances()
	require.Len(t, tokenBalances, 2) // from and to
}

func TestParseTransfers_ERC20Transfer_PrecompiledContract(t *testing.T) {
	module := createTestModule(t, []string{precompiledAddressHex})

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					createERC20TransferLog(
						precompiledAddressBytes,
						fromAddressPadded,
						toAddressPadded,
						big.NewInt(1000),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	// Should have 1 transfer
	require.Len(t, ctx.Block.Txs[0].Transfers, 1)

	transfer := ctx.Block.Txs[0].Transfers[0]
	require.Equal(t, types.Transfer, transfer.Type)
	require.Equal(t, decimal.NewFromInt(1000), transfer.Amount)

	// Precompiled contracts should NOT add tokens
	tokens := ctx.GetTokens()
	require.Empty(t, tokens)

	// Precompiled contracts should NOT add token balances
	tokenBalances := ctx.GetTokenBalances()
	require.Empty(t, tokenBalances)

	// Verify addresses are added and balances are updated
	addresses := ctx.GetAddresses()
	require.NotEmpty(t, addresses)

	// Find from and to addresses and verify balance updates
	var fromAddr, toAddr *storage.Address
	for _, addr := range addresses {
		fromHex := pkgTypes.Hex(fromAddressPadded[12:]).String()
		toHex := pkgTypes.Hex(toAddressPadded[12:]).String()
		if addr.Hash.String() == fromHex {
			fromAddr = addr
		}
		if addr.Hash.String() == toHex {
			toAddr = addr
		}
	}

	require.NotNil(t, fromAddr, "from address should be added")
	require.NotNil(t, toAddr, "to address should be added")

	// Check balances were updated for precompiled contract
	require.NotNil(t, fromAddr.Balance, "from address balance should be set")
	require.NotNil(t, toAddr.Balance, "to address balance should be set")

	// From address should have negative balance (subtracted)
	require.True(t, fromAddr.Balance.Value.Equal(decimal.NewFromInt(-1000)),
		"from address balance should be -1000, got %s", fromAddr.Balance.Value)

	// To address should have positive balance (added)
	require.True(t, toAddr.Balance.Value.Equal(decimal.NewFromInt(1000)),
		"to address balance should be 1000, got %s", toAddr.Balance.Value)
}

func TestParseTransfers_ERC20Mint_PrecompiledContract(t *testing.T) {
	module := createTestModule(t, []string{precompiledAddressHex})

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					createERC20TransferLog(
						precompiledAddressBytes,
						zeroAddressPadded, // from zero address = mint
						toAddressPadded,
						big.NewInt(5000),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	require.Len(t, ctx.Block.Txs[0].Transfers, 1)
	transfer := ctx.Block.Txs[0].Transfers[0]
	require.Equal(t, types.Mint, transfer.Type)
	require.Equal(t, decimal.NewFromInt(5000), transfer.Amount)

	// For mint, only ToAddress should be set
	require.Nil(t, transfer.FromAddress)
	require.NotNil(t, transfer.ToAddress)

	// No tokens for precompiled
	require.Empty(t, ctx.GetTokens())

	// Find to address and verify balance
	addresses := ctx.GetAddresses()
	var toAddr *storage.Address
	for _, addr := range addresses {
		toHex := pkgTypes.Hex(toAddressPadded[12:]).String()
		if addr.Hash.String() == toHex {
			toAddr = addr
		}
	}

	require.NotNil(t, toAddr)
	require.NotNil(t, toAddr.Balance)
	require.True(t, toAddr.Balance.Value.Equal(decimal.NewFromInt(5000)))
}

func TestParseTransfers_ERC20Burn_PrecompiledContract(t *testing.T) {
	module := createTestModule(t, []string{precompiledAddressHex})

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					createERC20TransferLog(
						precompiledAddressBytes,
						fromAddressPadded,
						zeroAddressPadded, // to zero address = burn
						big.NewInt(2000),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	require.Len(t, ctx.Block.Txs[0].Transfers, 1)
	transfer := ctx.Block.Txs[0].Transfers[0]
	require.Equal(t, types.Burn, transfer.Type)
	require.Equal(t, decimal.NewFromInt(2000), transfer.Amount)

	// For burn, only FromAddress should be set
	require.NotNil(t, transfer.FromAddress)
	require.Nil(t, transfer.ToAddress)

	// No tokens for precompiled
	require.Empty(t, ctx.GetTokens())

	// Find from address and verify balance
	addresses := ctx.GetAddresses()
	var fromAddr *storage.Address
	for _, addr := range addresses {
		fromHex := pkgTypes.Hex(fromAddressPadded[12:]).String()
		if addr.Hash.String() == fromHex {
			fromAddr = addr
		}
	}

	require.NotNil(t, fromAddr)
	require.NotNil(t, fromAddr.Balance)
	require.True(t, fromAddr.Balance.Value.Equal(decimal.NewFromInt(-2000)))
}

func TestParseTransfers_ERC20Mint_RegularContract(t *testing.T) {
	module := createTestModule(t, nil)

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					createERC20TransferLog(
						contractAddressBytes,
						zeroAddressPadded,
						toAddressPadded,
						big.NewInt(5000),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	require.Len(t, ctx.Block.Txs[0].Transfers, 1)
	transfer := ctx.Block.Txs[0].Transfers[0]
	require.Equal(t, types.Mint, transfer.Type)

	// Regular contract should add token
	tokens := ctx.GetTokens()
	require.Len(t, tokens, 1)
	require.True(t, tokens[0].Supply.Equal(decimal.NewFromInt(5000)))

	// Regular contract should add token balance (only for to address on mint)
	tokenBalances := ctx.GetTokenBalances()
	require.Len(t, tokenBalances, 1)
	require.True(t, tokenBalances[0].Balance.Equal(decimal.NewFromInt(5000)))
}

func TestParseTransfers_ERC20Burn_RegularContract(t *testing.T) {
	module := createTestModule(t, nil)

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					createERC20TransferLog(
						contractAddressBytes,
						fromAddressPadded,
						zeroAddressPadded,
						big.NewInt(3000),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	require.Len(t, ctx.Block.Txs[0].Transfers, 1)
	transfer := ctx.Block.Txs[0].Transfers[0]
	require.Equal(t, types.Burn, transfer.Type)

	// Regular contract should add token with negative supply for burn
	tokens := ctx.GetTokens()
	require.Len(t, tokens, 1)
	require.True(t, tokens[0].Supply.Equal(decimal.NewFromInt(-3000)))

	// Regular contract should add negative token balance for burn
	tokenBalances := ctx.GetTokenBalances()
	require.Len(t, tokenBalances, 1)
	require.True(t, tokenBalances[0].Balance.Equal(decimal.NewFromInt(-3000)))
}

func TestParseTransfers_MultipleTransfersInTx(t *testing.T) {
	module := createTestModule(t, []string{precompiledAddressHex})

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					// First transfer: regular contract
					createERC20TransferLog(
						contractAddressBytes,
						fromAddressPadded,
						toAddressPadded,
						big.NewInt(100),
					),
					// Second transfer: precompiled contract
					createERC20TransferLog(
						precompiledAddressBytes,
						fromAddressPadded,
						toAddressPadded,
						big.NewInt(200),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	// Should have 2 transfers
	require.Len(t, ctx.Block.Txs[0].Transfers, 2)

	// Only regular contract should add token
	tokens := ctx.GetTokens()
	require.Len(t, tokens, 1)

	// Only regular contract should add token balances
	tokenBalances := ctx.GetTokenBalances()
	require.Len(t, tokenBalances, 2) // from and to for regular contract only
}

func TestParseTransfers_MultipleTxs(t *testing.T) {
	module := createTestModule(t, nil)

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					createERC20TransferLog(
						contractAddressBytes,
						fromAddressPadded,
						toAddressPadded,
						big.NewInt(100),
					),
				},
			},
			{
				Status: types.TxStatusRevert,
				Logs: []*storage.Log{
					createERC20TransferLog(
						contractAddressBytes,
						fromAddressPadded,
						toAddressPadded,
						big.NewInt(200),
					),
				},
			},
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					createERC20TransferLog(
						contractAddressBytes,
						fromAddressPadded,
						toAddressPadded,
						big.NewInt(300),
					),
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	// First tx should have transfer
	require.Len(t, ctx.Block.Txs[0].Transfers, 1)
	require.Equal(t, decimal.NewFromInt(100), ctx.Block.Txs[0].Transfers[0].Amount)

	// Second tx (reverted) should not have transfers
	require.Nil(t, ctx.Block.Txs[1].Transfers)

	// Third tx should have transfer
	require.Len(t, ctx.Block.Txs[2].Transfers, 1)
	require.Equal(t, decimal.NewFromInt(300), ctx.Block.Txs[2].Transfers[0].Amount)
}

func TestParseTransfers_NonTransferLogIgnored(t *testing.T) {
	module := createTestModule(t, nil)

	// Non-transfer topic (e.g., Approval)
	approvalTopic := pkgTypes.MustDecodeHex("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925")

	ctx := dCtx.NewContext()
	ctx.Block = &storage.Block{
		Height: 100,
		Time:   time.Now(),
		Txs: []*storage.Tx{
			{
				Status: types.TxStatusSuccess,
				Logs: []*storage.Log{
					{
						Address: storage.Address{Hash: contractAddressBytes},
						Topics:  []pkgTypes.Hex{approvalTopic, fromAddressPadded, toAddressPadded},
						Data:    encodeUint256(big.NewInt(1000)),
					},
				},
			},
		},
	}

	err := module.parseTransfers(ctx)
	require.NoError(t, err)

	// Should have no transfers
	require.Empty(t, ctx.Block.Txs[0].Transfers)
	require.Empty(t, ctx.GetTokens())
}

func TestGetTransferType(t *testing.T) {
	tests := []struct {
		name     string
		from     pkgTypes.Hex
		to       pkgTypes.Hex
		expected types.TransferType
	}{
		{
			name:     "transfer",
			from:     fromAddressPadded[12:],
			to:       toAddressPadded[12:],
			expected: types.Transfer,
		},
		{
			name:     "mint",
			from:     zeroAddressPadded[12:],
			to:       toAddressPadded[12:],
			expected: types.Mint,
		},
		{
			name:     "burn",
			from:     fromAddressPadded[12:],
			to:       zeroAddressPadded[12:],
			expected: types.Burn,
		},
		{
			name:     "unknown - both zero",
			from:     zeroAddressPadded[12:],
			to:       zeroAddressPadded[12:],
			expected: types.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTransferType(tt.from, tt.to)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSupplyAmount(t *testing.T) {
	amount := decimal.NewFromInt(1000)

	tests := []struct {
		name         string
		transferType types.TransferType
		expected     decimal.Decimal
	}{
		{
			name:         "mint adds to supply",
			transferType: types.Mint,
			expected:     decimal.NewFromInt(1000),
		},
		{
			name:         "burn subtracts from supply",
			transferType: types.Burn,
			expected:     decimal.NewFromInt(-1000),
		},
		{
			name:         "transfer does not affect supply",
			transferType: types.Transfer,
			expected:     decimal.Zero,
		},
		{
			name:         "unknown does not affect supply",
			transferType: types.Unknown,
			expected:     decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSupplyAmount(amount, tt.transferType)
			require.True(t, result.Equal(tt.expected), "expected %s, got %s", tt.expected, result)
		})
	}
}

func TestIsERC20(t *testing.T) {
	tests := []struct {
		name         string
		topics       []pkgTypes.Hex
		isERC20      bool
		transferType types.TransferType
	}{
		{
			name:         "valid ERC20 transfer",
			topics:       []pkgTypes.Hex{erc20TransferTopic, fromAddressPadded, toAddressPadded},
			isERC20:      true,
			transferType: types.Transfer,
		},
		{
			name:         "wrong number of topics",
			topics:       []pkgTypes.Hex{erc20TransferTopic, fromAddressPadded},
			isERC20:      false,
			transferType: types.Unknown,
		},
		{
			name:         "wrong topic",
			topics:       []pkgTypes.Hex{fromAddressPadded, fromAddressPadded, toAddressPadded},
			isERC20:      false,
			transferType: types.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isERC20, transferType := isERC20(tt.topics)
			require.Equal(t, tt.isERC20, isERC20)
			if tt.isERC20 {
				require.Equal(t, tt.transferType, transferType)
			}
		})
	}
}

func TestIsAddress(t *testing.T) {
	tests := []struct {
		name     string
		data     pkgTypes.Hex
		expected bool
	}{
		{
			name:     "valid padded address",
			data:     fromAddressPadded,
			expected: true,
		},
		{
			name:     "zero address",
			data:     zeroAddressPadded,
			expected: true,
		},
		{
			name:     "20 byte address",
			data:     contractAddressBytes,
			expected: true,
		},
		{
			name:     "too many non-zero bytes",
			data:     pkgTypes.MustDecodeHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAddress(tt.data)
			require.Equal(t, tt.expected, result)
		})
	}
}
