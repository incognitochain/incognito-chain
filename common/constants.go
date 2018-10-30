package common

const (
	EmptyString         = ""
	MiliConstant        = 3 // 1 constant = 10^3 mili constant, we will use 1 miliconstant as minimum unit constant in tx
	IncMerkleTreeHeight = 29
)

const (
	TxNormalType       = "n" // normal tx(send and receive coin)
	TxSalaryType       = "s" // salary tx(gov pay salary for block producer)
	TxActionParamsType = "a" // action tx to edit params
	TxVotingType       = "v" // voting tx
	TxCustomTokenType  = "t" // token  tx
)

// unit type use in tx
// coin or token or bond
const (
	AssetTypeCoin     = "c" // 'constant' coin
	AssetTypeBond     = "b" // bond
	AssetTypeGovToken = "g" // government token
	AssetTypeDcbToken = "d" // decentralized central bank token
)

var ListAsset = []string{AssetTypeCoin, AssetTypeBond, AssetTypeGovToken, AssetTypeDcbToken}

// for mining consensus
const (
	MaxBlockSize          = 5000000 //byte 5MB
	MaxTxsInBlock         = 1000
	MinTxsInBlock         = 10                    // minium txs for block to get immediate process (meaning no wait time)
	MinBlockWaitTime      = 3                     // second
	MaxBlockWaitTime      = 20 - MinBlockWaitTime // second
	MaxSyncChainTime      = 5                     // second
	MaxBlockSigWaitTime   = 5                     // second
	MaxBlockPerTurn       = 100                   // maximum blocks that a validator can create per turn
	TotalValidators       = 20                    // = TOTAL CHAINS
	MinBlockSigs          = (TotalValidators / 2) + 1
	GetChainStateInterval = 10 //second
)
