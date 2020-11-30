package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/basemeta"
)

func (blockchain *BlockChain) verifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	txs []basemeta.Transaction,
	shardID byte,
) ([]basemeta.Transaction, error) {

	instUsed := make([]int, len(insts))
	txsUsed := make([]int, len(txs))
	invalidTxs := []basemeta.Transaction{}
	accumulatedValues := &basemeta.AccumulatedValues{
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
