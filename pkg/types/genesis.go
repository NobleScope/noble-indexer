package types

type Genesis struct {
	Config        Config                  `json:"config"`
	Nonce         Hex                     `json:"nonce"`
	Timestamp     Hex                     `json:"timestamp"`
	ExtraData     Hex                     `json:"extraData"`
	GasLimit      Hex                     `json:"gasLimit"`
	Difficulty    Hex                     `json:"difficulty"`
	MixHash       Hex                     `json:"mixHash"`
	Coinbase      Hex                     `json:"coinbase"`
	Alloc         map[string]GenesisAlloc `json:"alloc"`
	Number        Hex                     `json:"number"`
	GasUsed       Hex                     `json:"gasUsed"`
	ParentHash    Hex                     `json:"parentHash"`
	BaseFeePerGas *Hex                    `json:"baseFeePerGas,omitempty"`
	ExcessBlobGas *Hex                    `json:"excessBlobGas,omitempty"`
	BlobGasUsed   *Hex                    `json:"blobGasUsed,omitempty"`
}

type Config struct {
	ChainID                 int64        `json:"chainId"`
	HomesteadBlock          int64        `json:"homesteadBlock"`
	DaoForkSupport          bool         `json:"daoForkSupport"`
	Eip150Block             int64        `json:"eip150Block"`
	Eip155Block             int64        `json:"eip155Block"`
	Eip158Block             int64        `json:"eip158Block"`
	ByzantiumBlock          int64        `json:"byzantiumBlock"`
	ConstantinopleBlock     int64        `json:"constantinopleBlock"`
	PetersburgBlock         int64        `json:"petersburgBlock"`
	IstanbulBlock           int64        `json:"istanbulBlock"`
	MuirGlacierBlock        int64        `json:"muirGlacierBlock"`
	BerlinBlock             int64        `json:"berlinBlock"`
	LondonBlock             int64        `json:"londonBlock"`
	MergeNetsplitBlock      int64        `json:"mergeNetsplitBlock"`
	ShanghaiTime            int64        `json:"shanghaiTime"`
	CancunTime              int64        `json:"cancunTime"`
	PragueTime              int64        `json:"pragueTime"`
	TerminalTotalDifficulty int64        `json:"terminalTotalDifficulty"`
	DepositContractAddress  Hex          `json:"depositContractAddress"`
	BlobSchedule            BlobSchedule `json:"blobSchedule"`
}

type BlobSchedule struct {
	Cancun BlobParams `json:"cancun"`
	Prague BlobParams `json:"prague"`
}

type BlobParams struct {
	Target                int64 `json:"target"`
	Max                   int64 `json:"max"`
	BaseFeeUpdateFraction int64 `json:"baseFeeUpdateFraction"`
}

type GenesisAlloc struct {
	Code    Hex            `json:"code,omitempty"`
	Storage map[string]Hex `json:"storage,omitempty"`
	Balance Hex            `json:"balance"`
}
