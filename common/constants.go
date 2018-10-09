package common

const (
	TxNormalType        = "n" // normal tx(send and receive coin)
	TxActionParamsType  = "a" // action tx to edit params
	TxVotingType        = "v" // voting tx
	IncMerkleTreeHeight = 29
)

// unit type use in tx
const (
	TxOutCoinType = "c" // coin
	TxOutBondType = "b" // bond
)
