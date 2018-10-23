package common

const (
	EmptyString = ""
)

const (
	TxNormalType        = "n" // normal tx(send and receive coin)
	TxActionParamsType  = "a" // action tx to edit params
	TxVotingType        = "v" // voting tx
	IncMerkleTreeHeight = 29
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

const (
	MaxBlockSize            = 5000000 //byte 5MB
	MaxTxsInBlock           = 1000
	MinTxsInBlock           = 10                    // minium txs for block to get immediate process (meaning no wait time)
	MinBlockWaitTime        = 3                     // second
	MaxBlockWaitTime        = 20 - MinBlockWaitTime // second
	MaxSyncChainTime        = 5                     // second
	MaxBlockSigWaitTime     = 5                     // second
	MaxBlockPerTurn         = 100                   // maximum blocks that a validator can create per turn
	TotalValidators         = 20                    // = TOTAL CHAINS
	MinBlockSigs            = (TotalValidators / 2) + 1
	DefaultCoinBaseTxReward = 50
	GetChainStateInterval   = 10 //second
)
