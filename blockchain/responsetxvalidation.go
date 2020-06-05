package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) verifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	txs []metadata.Transaction,
	shardID byte,
) ([]metadata.Transaction, error) {

	instUsed := make([]int, len(insts))
	txsUsed := make([]int, len(txs))
	invalidTxs := []metadata.Transaction{}
	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		DBridgeTokenPair: map[string][]byte{},
		CBridgeTokens:    []*common.Hash{},
	}
	for _, tx := range txs {
		ok, err := tx.VerifyMinerCreatedTxBeforeGettingInBlock(txs, txsUsed, insts, instUsed, shardID, blockchain, accumulatedValues, nil, nil)
		if err != nil {
			return nil, err
		}
		if !ok {
			invalidTxs = append(invalidTxs, tx)
		}
	}
	if len(invalidTxs) > 0 {
		return invalidTxs, nil
	}
	return invalidTxs, nil
}
