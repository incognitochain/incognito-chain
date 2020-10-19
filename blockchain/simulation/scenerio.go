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

type CreateTxsAndInjectParam struct {
	InjectAt map[int]int
	Txs      []GenerateTxParam
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
	Token      string
	Interval   int
	Until      map[string]int
	IsBlocking bool
}
