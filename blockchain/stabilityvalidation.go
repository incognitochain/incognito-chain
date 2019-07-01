package blockchain

import (
	rCommon "github.com/incognitochain/incognito-chain/ethrelaying/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (bc *BlockChain) verifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	txs []metadata.Transaction,
	shardID byte,
) ([]metadata.Transaction, error) {

	instUsed := make([]int, len(insts))
	txsUsed := make([]int, len(txs))
	invalidTxs := []metadata.Transaction{}
	ethTxHashUsed := []rCommon.Hash{}
	for _, tx := range txs {
		ok, err := tx.VerifyMinerCreatedTxBeforeGettingInBlock(txs, txsUsed, insts, instUsed, shardID, bc, ethTxHashUsed)
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
