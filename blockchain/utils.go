package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) GetTransactionsByHashesWithShardID(
	hashes []common.Hash,
	shardID byte,
) ([]metadata.Transaction, error) {
	res := []metadata.Transaction{}
	blocks := make(map[string]*types.ShardBlock)
	for _, hash := range hashes {
		blockHash, index, err := rawdbv2.GetTransactionByHash(
			blockchain.GetShardChainDatabase(shardID),
			hash,
		)
		if err != nil {
			return nil, NewBlockChainError(GetTransactionFromDatabaseError, fmt.Errorf("Not found transaction with tx hash %+v", hash))
		}
		block, ok := blocks[blockHash.String()]
		if ok && block != nil {
			res = append(res, block.Body.Transactions[index])
			continue
		}

		// error is nil
		shardBlock, _, err := blockchain.GetShardBlockByHashWithShardID(blockHash, shardID)
		if err != nil {
			return nil, NewBlockChainError(GetTransactionFromDatabaseError, fmt.Errorf("Not found transaction with tx hash %+v", hash))
		}
		res = append(res, shardBlock.Body.Transactions[index])
	}
	return res, nil
}
