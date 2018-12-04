package privacy

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"unsafe"
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
	// for testing
	//point := new(EllipticPoint)
	//if rand{
	//	point.Randomize()
	//	return point
	//}
	//
	//return comZero

	return nil
}

// GetCmIndex returns cmIndex corresponding with cm
func (cmIndex *CMIndex) GetCmIndex(cm *EllipticPoint) {
	// Todo:
	cmIndex.BlockHeight = big.NewInt(10)
	cmIndex.TxIndex = 10
	cmIndex.CmId = 10

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

func (cmIndex CMIndex) Bytes() []byte{
	var bytes []byte
	nBytes := 0

	// convert block height to bytes array
	// because length of block height bytes is not specified
	// need to save length of block height in first byte
	blockHeightBytes := cmIndex.BlockHeight.Bytes()
	// append length of block height bytes
	bytes = append(bytes, byte(len(blockHeightBytes)))
	nBytes += 1
	// append block height bytes
	bytes = append(bytes, blockHeightBytes...)
	nBytes += len(blockHeightBytes)

	// convert tx index to bytes array
	txIndexBytes := (*[4]byte)(unsafe.Pointer(&cmIndex.TxIndex))[:]
	bytes = append(bytes, txIndexBytes...)
	nBytes += 4

	//convert cm id to bytes array
	cmIdBytes := (*[4]byte)(unsafe.Pointer(&cmIndex.CmId))[:]
	bytes = append(bytes, cmIdBytes...)
	nBytes += 4

	fmt.Printf("cmIndex bytes len: %v\n", nBytes)

	return bytes
}

func (cmIndex * CMIndex) SetBytes(bytes []byte) {
	// get len of block height bytes
	lenBlockHeightBytes := bytes[0]

	// get blockHeight
	blockHeightBytes := bytes[1:1+lenBlockHeightBytes]
	cmIndex.BlockHeight = new(big.Int).SetBytes(blockHeightBytes)

	// get tx index
	cmIndex.TxIndex = binary.LittleEndian.Uint32(bytes[1+lenBlockHeightBytes:5+lenBlockHeightBytes])

	// get cm id
	cmIndex.CmId = binary.LittleEndian.Uint32(bytes[5+lenBlockHeightBytes:9+lenBlockHeightBytes])

	fmt.Printf("cmIndex revert: %+v\n", cmIndex)
}

