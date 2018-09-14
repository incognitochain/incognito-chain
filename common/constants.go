package common

const (
	TxNormalType        = "n" // normal tx(send and receive coin)
	TxActionParamsType  = "a" // action tx to edit params
	IncMerkleTreeHeight = 29
)

// unit type use in tx
const (
	TxOutCoinType = "c" // coin
	TxOutBondType = "b" // bond
)
