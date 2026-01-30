package parser

import (
	"bytes"
	"math/big"

	"github.com/baking-bad/noble-indexer/internal/storage"
	"github.com/baking-bad/noble-indexer/internal/storage/types"
	dCtx "github.com/baking-bad/noble-indexer/pkg/indexer/decode/context"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/erc1155"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/erc20"
	"github.com/baking-bad/noble-indexer/pkg/indexer/parser/types/erc721"
	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	AddressBytesLength              = 20
	ERC20FirstTopic                 = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	ERC721FirstTopic                = ERC20FirstTopic
	ERC1155TransferSingleFirstTopic = "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62"
	ERC1155TransferBatchFirstTopic  = "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb"
)

func (p *Module) parseTransfers(ctx *dCtx.Context) error {
	for i := range ctx.Block.Txs {
		if ctx.Block.Txs[i].Status == types.TxStatusRevert {
			continue
		}
		ctx.Block.Txs[i].Transfers = make([]*storage.Transfer, 0)

		for j := range ctx.Block.Txs[i].Logs {
			log := ctx.Block.Txs[i].Logs[j]
			transfer := &storage.Transfer{
				Height: ctx.Block.Height,
				Time:   ctx.Block.Time,
				Contract: storage.Contract{
					Address: storage.Address{
						Hash: log.Address.Hash,
					},
				},
			}

			tokenType, transferType := getTokenAndTransferTypes(log.Topics)

			switch tokenType {
			case types.ERC20:
				transferEvent, err := parseLogs[erc20.EventTransfer](p.abi[tokenType], log.Data, log.Topics)
				if err != nil {
					return err
				}
				transferEvent.From = common.BytesToAddress(log.Topics[1])
				transferEvent.To = common.BytesToAddress(log.Topics[2])

				transfer.TokenID = decimal.Zero
				transfer.Amount = decimal.NewFromBigInt(transferEvent.Value, 0)
				setAddresses(
					ctx,
					transferType,
					transfer,
					transferEvent.From.Bytes(),
					transferEvent.To.Bytes(),
					log.Address.Hash,
				)
			case types.ERC721:
				transferEvent, err := parseLogs[erc721.EventTransfer](p.abi[tokenType], log.Data, log.Topics)
				if err != nil {
					return err
				}
				transferEvent.From = common.BytesToAddress(log.Topics[1])
				transferEvent.To = common.BytesToAddress(log.Topics[2])
				transfer.Amount = decimal.NewFromBigInt(new(big.Int).SetBytes(log.Topics[3]), 0)
				setAddresses(
					ctx,
					transferType,
					transfer,
					transferEvent.From.Bytes(),
					transferEvent.To.Bytes(),
					log.Address.Hash,
				)
			case types.ERC1155:
				if log.Topics[0].Hex() == ERC1155TransferSingleFirstTopic {
					transferEvent, err := parseLogs[erc1155.EventTransferSingle](p.abi[tokenType], log.Data, log.Topics)
					if err != nil {
						return err
					}

					transferEvent.From = common.BytesToAddress(log.Topics[2])
					transferEvent.To = common.BytesToAddress(log.Topics[3])
					transfer.TokenID = decimal.NewFromBigInt(transferEvent.Id, 0)
					transfer.Amount = decimal.NewFromBigInt(transferEvent.Value, 0)
					setAddresses(
						ctx,
						transferType,
						transfer,
						transferEvent.From.Bytes(),
						transferEvent.To.Bytes(),
						log.Address.Hash,
					)
				}

				if log.Topics[0].Hex() == ERC1155TransferBatchFirstTopic {
					transferEvent, err := parseLogs[erc1155.EventTransferBatch](p.abi[tokenType], log.Data, log.Topics)
					if err != nil {
						return err
					}

					transferEvent.From = common.BytesToAddress(log.Topics[2])
					transferEvent.To = common.BytesToAddress(log.Topics[3])
					for id := range transferEvent.Ids {
						batchTransfer := &storage.Transfer{
							Height: ctx.Block.Height,
							Time:   ctx.Block.Time,
							Contract: storage.Contract{
								Address: storage.Address{
									Hash: log.Address.Hash,
								},
							},
							TokenID: decimal.NewFromBigInt(transferEvent.Ids[id], 0),
							Amount:  decimal.NewFromBigInt(transferEvent.Values[id], 0),
							Type:    transferType,
						}
						setAddresses(
							ctx,
							transferType,
							batchTransfer,
							transferEvent.From.Bytes(),
							transferEvent.To.Bytes(),
							log.Address.Hash,
						)
						ctx.Block.Txs[i].Transfers = append(ctx.Block.Txs[i].Transfers, batchTransfer)
						ctx.AddToken(&storage.Token{
							TokenID:        batchTransfer.TokenID,
							Height:         ctx.Block.Height,
							LastHeight:     ctx.Block.Height,
							Type:           tokenType,
							Status:         types.Pending,
							Contract:       batchTransfer.Contract,
							TransfersCount: 1,
							Supply:         getSupplyAmount(batchTransfer.Amount, batchTransfer.Type),
						})
						setTokenBalanceUpdates(ctx, batchTransfer)
					}

					continue
				}

			default:
				continue
			}

			transfer.Type = transferType
			ctx.Block.Txs[i].Transfers = append(ctx.Block.Txs[i].Transfers, transfer)
			ctx.AddToken(&storage.Token{
				TokenID:        transfer.TokenID,
				Height:         ctx.Block.Height,
				LastHeight:     ctx.Block.Height,
				Type:           tokenType,
				Status:         types.Pending,
				Contract:       transfer.Contract,
				TransfersCount: 1,
				Supply:         getSupplyAmount(transfer.Amount, transfer.Type),
			})
			setTokenBalanceUpdates(ctx, transfer)
		}
	}

	return nil
}

func getTokenAndTransferTypes(topics []pkgTypes.Hex) (types.TokenType, types.TransferType) {
	if isERC20type, transferType := isERC20(topics); isERC20type {
		return types.ERC20, transferType
	}
	if isERC721type, transferType := isERC721(topics); isERC721type {
		return types.ERC721, transferType
	}
	if isERC1155type, transferType := isERC1155(topics); isERC1155type {
		return types.ERC1155, transferType
	}

	return "", ""
}

func parseLogs[T any](contractAbi *abi.ABI, logData []byte, topics []pkgTypes.Hex) (*T, error) {
	topic0 := common.BytesToHash(topics[0])
	event, err := contractAbi.EventByID(topic0)
	if err != nil {
		return nil, errors.Wrapf(err, "could not find event for topic %s", topic0.Hex())
	}

	var data T
	if err = contractAbi.UnpackIntoInterface(&data, event.Name, logData); err != nil {
		return nil, errors.Wrap(err, "failed to unpack log data")
	}

	return &data, nil
}

func isERC20(topics []pkgTypes.Hex) (bool, types.TransferType) {
	if len(topics) != 3 {
		return false, types.Unknown
	}
	if topics[0].Hex() != ERC20FirstTopic {
		return false, types.Unknown
	}
	if !isAddress(topics[1]) {
		return false, types.Unknown
	}
	if !isAddress(topics[2]) {
		return false, types.Unknown
	}

	return true, getTransferType(topics[1][12:], topics[2][12:])
}

func isERC721(topics []pkgTypes.Hex) (bool, types.TransferType) {
	if len(topics) != 4 {
		return false, types.Unknown
	}
	if topics[0].Hex() != ERC721FirstTopic {
		return false, types.Unknown
	}
	if !isAddress(topics[1]) {
		return false, types.Unknown
	}
	if !isAddress(topics[2]) {
		return false, types.Unknown
	}

	return true, getTransferType(topics[1][12:], topics[2][12:])
}

func isERC1155(topics []pkgTypes.Hex) (bool, types.TransferType) {
	if len(topics) != 4 {
		return false, types.Unknown
	}
	if topics[0].Hex() != ERC1155TransferSingleFirstTopic && topics[0].Hex() != ERC1155TransferBatchFirstTopic {
		return false, types.Unknown
	}
	if !isAddress(topics[1]) {
		return false, types.Unknown
	}
	if !isAddress(topics[2]) {
		return false, types.Unknown
	}
	if !isAddress(topics[3]) {
		return false, types.Unknown
	}

	return true, getTransferType(topics[1][12:], topics[2][12:])
}

func getTransferType(from, to pkgTypes.Hex) types.TransferType {
	fromIsZero := true
	toIsZero := true

	for i := range AddressBytesLength {
		if from[i] != 0 {
			fromIsZero = false
			break
		}
	}
	for i := range AddressBytesLength {
		if to[i] != 0 {
			toIsZero = false
			break
		}
	}

	if !fromIsZero && !toIsZero {
		return types.Transfer
	}
	if fromIsZero && !toIsZero {
		return types.Mint
	}
	if !fromIsZero {
		return types.Burn
	}

	return types.Unknown
}

func setAddresses(ctx *dCtx.Context, transferType types.TransferType, transfer *storage.Transfer, from, to, contract pkgTypes.Hex) {
	fromAddress := &storage.Address{
		Hash:        from,
		FirstHeight: ctx.Block.Height,
		LastHeight:  ctx.Block.Height,
		Balance:     storage.EmptyBalance(),
	}
	toAddress := &storage.Address{
		Hash:        to,
		FirstHeight: ctx.Block.Height,
		LastHeight:  ctx.Block.Height,
		Balance:     storage.EmptyBalance(),
	}
	contractAddress := &storage.Address{
		Hash:        contract,
		FirstHeight: ctx.Block.Height,
		LastHeight:  ctx.Block.Height,
		IsContract:  true,
	}

	storageContract := &storage.Contract{
		Address: storage.Address{
			Hash: contractAddress.Hash,
		},
		Height: ctx.Block.Height,
	}

	ctx.AddContract(storageContract)
	ctx.AddAddress(contractAddress)

	if transferType == types.Burn || transferType == types.Transfer {
		transfer.FromAddress = fromAddress
		ctx.AddAddress(fromAddress)
	}
	if transferType == types.Mint || transferType == types.Transfer {
		transfer.ToAddress = toAddress
		ctx.AddAddress(toAddress)
	}
}

func isAddress(data pkgTypes.Hex) bool {
	return len(bytes.TrimLeft(data, "\x00")) <= AddressBytesLength
}

func getSupplyAmount(amount decimal.Decimal, transferType types.TransferType) decimal.Decimal {
	switch transferType {
	case types.Mint:
		return amount
	case types.Burn:
		return amount.Neg()
	default:
		return decimal.Zero
	}
}

func setTokenBalanceUpdates(ctx *dCtx.Context, transfer *storage.Transfer) {
	switch transfer.Type {
	case types.Mint:
		ctx.AddTokenBalance(&storage.TokenBalance{
			TokenID:  transfer.TokenID,
			Balance:  transfer.Amount,
			Contract: transfer.Contract,
			Address:  *transfer.ToAddress,
		})
	case types.Burn:
		ctx.AddTokenBalance(&storage.TokenBalance{
			TokenID:  transfer.TokenID,
			Balance:  transfer.Amount.Neg(),
			Contract: transfer.Contract,
			Address:  *transfer.FromAddress,
		})
	case types.Transfer:
		ctx.AddTokenBalance(&storage.TokenBalance{
			TokenID:  transfer.TokenID,
			Balance:  transfer.Amount.Neg(),
			Contract: transfer.Contract,
			Address:  *transfer.FromAddress,
		})
		ctx.AddTokenBalance(&storage.TokenBalance{
			TokenID:  transfer.TokenID,
			Balance:  transfer.Amount,
			Contract: transfer.Contract,
			Address:  *transfer.ToAddress,
		})
	}
}
