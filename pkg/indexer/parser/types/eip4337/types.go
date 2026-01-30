package eip4337

import (
	"fmt"
	"math/big"
	"strings"

	pkgTypes "github.com/baking-bad/noble-indexer/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	ABIEntryPointV06          = "EIP4337_entrypoint_v06"
	ABIEntryPointV07          = "EIP4337_entrypoint_v07"
	ABIEntryPointForPaymaster = "EIP4337_entrypoint_for_paymaster"
	ABIAccountV06             = "EIP4337_account_v06"
	ABIAccountV07             = "EIP4337_account_v07"
	ABIPaymasterV06           = "EIP4337_paymaster_v06"
	ABIPaymasterV07           = "EIP4337_paymaster_v07"
)

const UserOperationEventSignature = "0x49628fd1471006c1482da88028e9ce4dbb080b815c9b0344d39e5a8e6ec1419f"

var EntryPointAddresses = map[string]string{
	"0x5ff137d4b0fdcd49dca30c7cf57e578a026d2789": "v0.6",
	"0x0000000071727de22e5e9d8baf0edac6f37da032": "v0.7",
}

var (
	HandleOpsV06Selector = pkgTypes.MustDecodeHex("0x1fad948c")
	HandleOpsV07Selector = pkgTypes.MustDecodeHex("0x765e827f")
)

type Identifiable interface {
	GetUniqueKey() string
}

type CommonUserOperation interface {
	GetSender() common.Address
	GetNonce() *big.Int
	GetInitCode() []byte
	GetCallData() []byte
	GetPreVerificationGas() *big.Int
	GetPaymasterAndData() []byte
	GetSignature() []byte
}

type IdentifiableUserOp interface {
	Identifiable
	UserOperation | PackedUserOperation
}

type UserOperationEvent struct {
	EntryPointV06UserOperationEvent
}

func (op UserOperation) GetUniqueKey() string {
	return strings.ToLower(fmt.Sprintf("%s:%s", op.Sender.Hex(), op.Nonce.String()))
}

func (op PackedUserOperation) GetUniqueKey() string {
	return strings.ToLower(fmt.Sprintf("%s:%s", op.Sender.Hex(), op.Nonce.String()))
}

func (event UserOperationEvent) GetUniqueKey() string {
	return strings.ToLower(fmt.Sprintf("%s:%s", event.Sender.Hex(), event.Nonce.String()))
}
