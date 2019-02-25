package blockchain

import (
	"errors"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

//Receive tx list from shard block body, produce merkle path of UTXO CrossShard List from specific shardID
func GetMerklePathCrossShard(txList []metadata.Transaction, shardID byte) (merklePathShard []common.Hash, merkleShardRoot common.Hash) {
	//calculate output coin hash for each shard
	outputCoinHash := getOutCoinHashEachShard(txList)
	// calculate merkel path for a shardID
	// step 1: calculate merkle data : [1, 2, 3, 4, 12, 34, 1234]
	/*
			   	1234=hash(12,34)
			   /			  \
		  12=hash(1,2)	 34=hash(3,4)
			 / \	 		 / \
			1	2			3	4
	*/
	merkleData := outputCoinHash
	// fmt.Println("merkleData 1 ", merkleData)
	// fmt.Println("outputCoinHash 1 ", outputCoinHash)
	cursor := 0
	for {
		v1 := merkleData[cursor]
		v2 := merkleData[cursor+1]
		merkleData = append(merkleData, common.HashH(append(v1.GetBytes(), v2.GetBytes()...)))
		cursor += 2
		if cursor >= len(merkleData)-1 {
			break
		}
	}
	fmt.Println("merkleData 2", merkleData)
	//TODO: @kumi check again (fix infinity loop ->fix by @merman)
	// step 2: get merkle path
	cursor = 0
	lastCursor := 0
	sid := int(shardID)
	i := sid
	time := 0
	for {
		if cursor >= len(merkleData)-2 {
			break
		}
		if i%2 == 0 {
			merklePathShard = append(merklePathShard, merkleData[cursor+i+1])
		} else {
			merklePathShard = append(merklePathShard, merkleData[cursor+i-1])
		}
		i = i / 2

		if time == 0 {
			cursor += len(outputCoinHash)
		} else {
			tmp := cursor
			cursor += (cursor - lastCursor) / 2
			lastCursor = tmp
		}
		time++
	}
	// fmt.Println("merklePathShard", merklePathShard)
	// fmt.Println("merkleShardRoot", merkleShardRoot)
	merkleShardRoot = merkleData[len(merkleData)-1]
	return merklePathShard, merkleShardRoot
}

//Receive a cross shard block and merkle path, verify whether the UTXO list is valid or not
func VerifyCrossShardBlockUTXO(block *CrossShardBlock, merklePathShard []common.Hash) bool {
	outCoins := block.CrossOutputCoin
	tmpByte := []byte{}
	for _, coin := range outCoins {
		tmpByte = append(tmpByte, coin.Bytes()...)
	}
	finalHash := common.HashH(tmpByte)
	for _, hash := range merklePathShard {
		finalHash = common.HashH(append(finalHash.GetBytes(), hash.GetBytes()...))
	}
	return VerifyMerkleTree(finalHash, merklePathShard, block.Header.ShardTxRoot)
}

func VerifyMerkleTree(finalHash common.Hash, merklePath []common.Hash, merkleRoot common.Hash) bool {
	for _, hashPath := range merklePath {
		finalHash = common.HashH(append(finalHash.GetBytes(), hashPath.GetBytes()...))
	}
	if finalHash != merkleRoot {
		return false
	} else {
		return true
	}
}

// helper function to group OutputCoin into shard and get the hash of each group
func getOutCoinHashEachShard(txList []metadata.Transaction) []common.Hash {
	// group transaction by shardID
	// fmt.Println("len of txList", len(txList))
	outCoinEachShard := make([][]*privacy.OutputCoin, common.MAX_SHARD_NUMBER)
	for _, tx := range txList {
		for _, outCoin := range tx.GetProof().OutputCoins {
			lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
			shardID := common.GetShardIDFromLastByte(lastByte)
			// fmt.Printf("ShardID %+v ,outCoin %+v \n", shardID, outCoin)
			outCoinEachShard[shardID] = append(outCoinEachShard[shardID], outCoin)
		}
	}
	// fmt.Println("len of outCoinEachShard", len(outCoinEachShard))
	// fmt.Println("value of outCoinEachShard", outCoinEachShard)
	//calcualte hash for each shard
	outputCoinHash := make([]common.Hash, common.MAX_SHARD_NUMBER)
	for i := 0; i < common.MAX_SHARD_NUMBER; i++ {
		if len(outCoinEachShard[i]) == 0 {
			outputCoinHash[i] = common.HashH([]byte(""))
		} else {
			tmpByte := []byte{}
			for _, coin := range outCoinEachShard[i] {
				tmpByte = append(tmpByte, coin.Bytes()...)
			}
			outputCoinHash[i] = common.HashH(tmpByte)
		}
	}
	// fmt.Println("len of outputCoinHash", len(outputCoinHash))
	return outputCoinHash
}

// helper function to get the hash of OutputCoins (send to a shard) from list of transaction
/*
	Get output coin of transaction
	Check receiver last byte
	Append output coin to corresponding shard
*/
func getOutCoinCrossShard(txList []metadata.Transaction, shardID byte) []privacy.OutputCoin {
	coinList := []privacy.OutputCoin{}
	for _, tx := range txList {
		for _, outCoin := range tx.GetProof().OutputCoins {
			lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
			if lastByte == shardID {
				coinList = append(coinList, *outCoin)
			}
		}
	}
	return coinList
}

/*
	Verify CrossShard Block
	- Agg Signature
	- MerklePath
*/
func (crossShardBlock *CrossShardBlock) VerifyCrossShardBlock(committees []string) error {
	if err := ValidateAggSignature(crossShardBlock.ValidatorsIdx, committees, crossShardBlock.AggregatedSig, crossShardBlock.R, crossShardBlock.Hash()); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := VerifyCrossShardBlockUTXO(crossShardBlock, crossShardBlock.MerklePathShard); !ok {
		return NewBlockChainError(HashError, errors.New("verify Merkle Path Shard"))
	}
	return nil
}

func (self *CrossShardBlock) ShouldStoreBlock() bool {
	// verify block aggregation
	return true
}
