package blockchain

import (
	"github.com/constant-money/constant-chain/metadata"
)

func (bc *BlockChain) verifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	txs []metadata.Transaction,
	shardID byte,
) ([]metadata.Transaction, error) {

	instUsed := make([]int, len(insts))
	txsUsed := make([]int, len(txs))
	invalidTxs := []metadata.Transaction{}
	for _, tx := range txs {
		ok, err := tx.VerifyMinerCreatedTxBeforeGettingInBlock(txs, txsUsed, insts, instUsed, shardID, bc)
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
