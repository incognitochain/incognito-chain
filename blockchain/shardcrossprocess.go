package blockchain

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)


func CreateAllCrossShardBlock(shardBlock *types.ShardBlock, activeShards int) map[byte]*types.CrossShardBlock {
	allCrossShard := make(map[byte]*types.CrossShardBlock)
	if activeShards == 1 {
		return allCrossShard
	}
	for i := 0; i < activeShards; i++ {
		shardID := common.GetShardIDFromLastByte(byte(i))
		if shardID != shardBlock.Header.ShardID {
			crossShard, err := CreateCrossShardBlock(shardBlock, shardID)
			if crossShard != nil {
				Logger.log.Criticalf("Create CrossShardBlock from Shard %+v to Shard %+v: %+v \n", shardBlock.Header.ShardID, shardID, crossShard)
			}
			if crossShard != nil && err == nil {
				allCrossShard[byte(i)] = crossShard
			}
		}
	}
	return allCrossShard
}

func CreateCrossShardBlock(shardBlock *types.ShardBlock, shardID byte) (*types.CrossShardBlock, error) {
	crossShard := &types.CrossShardBlock{}
	crossOutputCoin, crossCustomTokenPrivacyData, err := types.GetCrossShardData(shardBlock.Body.Transactions, shardID)
	if err != nil {
		return nil, errors.New("No cross Outputcoin, Cross Custom Token, Cross Custom Token Privacy")
	}
	// Return nothing if nothing to cross
	if len(crossOutputCoin) == 0 && len(crossCustomTokenPrivacyData) == 0 {
		return nil, errors.New("No cross Outputcoin, Cross Custom Token, Cross Custom Token Privacy")
	}
	merklePathShard, merkleShardRoot := GetMerklePathCrossShard(shardBlock.Body.Transactions, shardID)
	if merkleShardRoot != shardBlock.Header.ShardTxRoot {
		return crossShard, fmt.Errorf("Expect Shard Tx Root To be %+v but get %+v", shardBlock.Header.ShardTxRoot, merkleShardRoot)
	}
	crossShard.ValidationData = shardBlock.ValidationData
	crossShard.Header = shardBlock.Header
	crossShard.MerklePathShard = merklePathShard
	crossShard.CrossOutputCoin = crossOutputCoin
	crossShard.CrossTxTokenPrivacyData = crossCustomTokenPrivacyData
	crossShard.ToShardID = shardID
	return crossShard, nil
}



// VerifyCrossShardBlock Verify CrossShard Block
//- Agg Signature
//- MerklePath
func VerifyCrossShardBlock(crossShardBlock *types.CrossShardBlock, blockchain *BlockChain, committees []incognitokey.CommitteePublicKey) error {
	if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(crossShardBlock, committees); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := VerifyCrossShardBlockUTXO(crossShardBlock, crossShardBlock.MerklePathShard); !ok {
		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
	}
	return nil
}

// VerifyCrossShardBlockUTXO Calculate Final Hash as Hash of:
//	1. CrossTransactionFinalHash
//	2. TxTokenDataVoutFinalHash
//	3. CrossTxTokenPrivacyData
// These hashes will be calculated as comment in getCrossShardDataHash function
func VerifyCrossShardBlockUTXO(block *types.CrossShardBlock, merklePathShard []common.Hash) bool {
	var outputCoinHash common.Hash
	var txTokenDataHash common.Hash
	var txTokenPrivacyDataHash common.Hash
	outCoins := block.CrossOutputCoin
	outputCoinHash = calHashOutCoinCrossShard(outCoins)
	txTokenDataHash = calHashTxTokenDataHashList()
	txTokenPrivacyDataList := block.CrossTxTokenPrivacyData
	txTokenPrivacyDataHash = calHashTxTokenPrivacyDataHashList(txTokenPrivacyDataList)
	tmpByte := append(append(outputCoinHash.GetBytes(), txTokenDataHash.GetBytes()...), txTokenPrivacyDataHash.GetBytes()...)
	finalHash := common.HashH(tmpByte)
	return Merkle{}.VerifyMerkleRootFromMerklePath(finalHash, merklePathShard, block.Header.ShardTxRoot, block.ToShardID)
}