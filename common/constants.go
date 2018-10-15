package common

const (
	TxNormalType        = "n" // normal tx(send and receive coin)
	TxActionParamsType  = "a" // action tx to edit params
	IncMerkleTreeHeight = 29
)

// unit type use in tx
// coin or token or bond
const (
	TxOutCoinType  = "x" // coin x
	TxOutBondType  = "b" // bond
	TxGovTokenType = "g" // government token
	TxDcbTokenType = "d" // decentralized central bank token
	TxCmbTokenType = "c" // commercial bank token
)

const (
	MaxBlockSize            = 5000000 //byte 5MB
	MaxTxsInBlock           = 1000
	MinTxsInBlock           = 10                    // minium txs for block to get immediate process (meaning no wait time)
	MinBlockWaitTime        = 3                     // second
	MaxBlockWaitTime        = 20 - MinBlockWaitTime // second
	MaxSyncChainTime        = 5                     // second
	MaxBlockSigWaitTime     = 20                    // second
	TotalValidators         = 20                    // = TOTAL CHAINS
	MinBlockSigs            = (TotalValidators / 2) + 1
	DefaultCoinBaseTxReward = 50
	GetChainStateInterval   = 10 //second
)
