package privacy

import (
	"math/big"
)

type CMIndex struct{
	BlockHeight *big.Int
	TxIndex     uint32
	CmId        uint32
}

// Randomize generate cmIndex with blockHeightMax is current block height
func (cmIndex * CMIndex) Randomize(blockHeightMax *big.Int){
}

// RandTxID generates TxID in blockHeight, numTx is number of txs in that blockHeight
func (cmIndex * CMIndex)  RandTxID(blockHeight *big.Int, numTxs uint64){
}

// RandTxID generates CmId in txID in blockHeight, numCms is number of commitment in that txID
func (cmIndex * CMIndex)  RandCmId(blockHeight *big.Int, txIndex uint64, numCms uint64){
}

// GetCommitment returns commitment with cmIndex
func (cmIndex CMIndex) GetCommitment() *EllipticPoint{
	return nil
}

// GetCmIndex returns cmIndex corresponding with cm
func (cmIndex *CMIndex) GetCmIndex(cm *EllipticPoint) {

}

// IsEqual returns true if two cmIndexs is the same
func (cmIndex CMIndex) IsEqual(target *CMIndex) bool {
	if cmIndex.BlockHeight.Cmp(target.BlockHeight) != 0 {
		return false
	}
	if cmIndex.TxIndex != target.TxIndex {
		return false
	}
	if cmIndex.CmId != target.CmId {
		return false
	}

	return true
}
