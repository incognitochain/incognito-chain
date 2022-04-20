package blockchain

import (
	"bytes"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
)

func IsBridgeTxHashUsedInBlock(uniqTx []byte, uniqTxsUsed [][]byte) bool {
	for _, item := range uniqTxsUsed {
		if bytes.Equal(uniqTx, item) {
			return true
		}
	}
	return false
}

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

func SearchUint64(list []uint64, target uint64) (int, bool) {
	// search returns the leftmost position where f returns true, or len(x) if f
	// returns false for all x. This is the insertion position for target in x,
	// and could point to an element that's either == target or not.
	l := 0
	r := len(list) - 1
	for l <= r {
		m := l + (r-l)/2
		if list[m] == target {
			return m, true
		}
		if list[m] > target {
			r = m - 1
		} else {
			l = m + 1
		}
	}
	return l, false
}
