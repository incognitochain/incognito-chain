package privacy

import (
	"math/big"
	"github.com/ninjadotorg/constant/common"
)

type CMIndex struct{
	BlockHeight *big.Int
	TxId *common.Hash
	CmId uint64
}

// Randomize generate cmIndex with blockHeightMax is current block height
func (cmIndex * CMIndex) Randomize(blockHeightMax *big.Int){
}

// RandTxID generates TxID in blockHeight, numTx is number of txs in that blockHeight
func (cmIndex * CMIndex)  RandTxID(blockHeight *big.Int, numTxs *common.Hash){
}

// RandTxID generates CmId in txID in blockHeight, numCms is number of commitment in that txID
func (cmIndex * CMIndex)  RandCmId(blockHeight *big.Int, txID *common.Hash, numCms uint64){
}

// GetCommitment returns commitment with cmIndex
func (cmIndex CMIndex) GetCommitment() *EllipticPoint{
	return nil
}
