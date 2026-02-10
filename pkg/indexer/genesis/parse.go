package genesis

import (
	"time"

	"github.com/NobleScope/noble-indexer/internal/storage"
	dCtx "github.com/NobleScope/noble-indexer/pkg/indexer/decode/context"
	"github.com/NobleScope/noble-indexer/pkg/indexer/parser"
	pkgTypes "github.com/NobleScope/noble-indexer/pkg/types"
)

type parsedData struct {
	addresses map[string]*storage.Address
	contracts map[string]*storage.Contract
	chainId   int64
	time      time.Time
}

func newParsedData() parsedData {
	return parsedData{
		addresses: make(map[string]*storage.Address),
		contracts: make(map[string]*storage.Contract),
	}
}

func (module *Module) parse(genesis pkgTypes.Genesis) (parsedData, error) {
	data := newParsedData()
	decodeCtx := dCtx.NewContext()

	coinbase := &storage.Address{
		Hash:        genesis.Coinbase,
		FirstHeight: 0,
		LastHeight:  0,
		Balance:     storage.EmptyBalance(),
	}
	decodeCtx.AddAddress(coinbase)

	for k, v := range genesis.Alloc {
		balance, err := v.Balance.Decimal()
		if err != nil {
			return data, err
		}

		hash, err := pkgTypes.HexFromString(k)
		if err != nil {
			return data, err
		}

		address := &storage.Address{
			Hash:        hash,
			FirstHeight: 0,
			LastHeight:  0,
			Balance: &storage.Balance{
				Value: balance,
			},
		}
		decodeCtx.AddAddress(address)

		if v.Code != nil {
			address.IsContract = true
			contract := &storage.Contract{
				Height: 0,
				Address: storage.Address{
					Hash: hash,
				},
				Code: v.Code,
				TxId: nil,
			}

			err = parser.ParseEvmContractMetadata(contract)
			if err != nil {
				return data, err
			}

			decodeCtx.AddContract(contract)
		}
	}

	for _, addr := range decodeCtx.GetAddresses() {
		data.addresses[addr.String()] = addr
	}

	for _, c := range decodeCtx.GetContracts() {
		data.contracts[c.Address.String()] = c
	}
	data.chainId = genesis.Config.ChainID
	genesisTime, err := genesis.Timestamp.Time()
	if err != nil {
		return data, err
	}
	data.time = genesisTime

	return data, nil
}
