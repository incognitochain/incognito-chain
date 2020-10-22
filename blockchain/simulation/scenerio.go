package main

type ScenerioAction struct {
	Action Action
	Params interface{}
}

type Action string

const (
	GENERATEBLOCKS     Action = "GENERATEBLOCKS"
	AUTOGENERATEBLOCKS Action = "AUTOGENERATEBLOCKS"
	GENERATETXS        Action = "GENERATETXS"
	CREATETXSANDINJECT Action = "CREATETXSANDINJECT"
	CREATESTAKINGTX    Action = "CREATESTAKINGTX"
	CHECKBESTSTATES    Action = "BESTSTATES"
	CHECKBALANCES      Action = "CHECKBALANCES"
	SWITCHTOMANUAL     Action = "SWITCHTOMANUAL"
)

type GenerateBlocksParam struct {
	ChainID    int
	Blocks     int
	IsBlocking bool
}

type AutoGenerateBlocks struct {
	ChainID int
	Enable  bool
}
type GenerateTxParam struct {
	SenderPrK string
	Receivers map[string]int
}

type CreateStakingTx struct {
	SenderPrk   string
	MinerPrk    string
	RewardAddr  string
	StakeShard  bool
	AutoRestake bool
}

type CreateTxsAndInjectParam struct {
	InjectAt struct {
		ChainID int
		Height  uint64
	}
	Txs []GenerateTxParam
}

type CheckBestStateParam struct {
	ChainID           int
	AtHeight          uint64
	CheckLengthFields map[string]int
	CheckDataFields   map[string]string
	IsBlocking        bool
}

type CheckBalanceParam struct {
	PrivateKey string
	Tokens     map[string]uint64
	Interval   int
	Until      struct {
		ChainID int
		Height  uint64
	}
	IsBlocking bool
}
