package mock

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type TxPool struct{}

func (tp *TxPool) HaveTransaction(hash *common.Hash) bool {
	return false
}

func (tp *TxPool) RemoveTx(txs []metadata.Transaction, isInBlock bool) {

}
func (tp *TxPool) RemoveCandidateList([]string) {

}

func (tp *TxPool) EmptyPool() bool {
	return true
}

func (tp *TxPool) MaybeAcceptTransactionForBlockProducing(metadata.Transaction, int64, *blockchain.ShardBestState) (*metadata.TxDesc, error) {
	return nil, nil
}

func (tp *TxPool) MaybeAcceptBatchTransactionForBlockProducing(byte, []metadata.Transaction, int64, *blockchain.ShardBestState) ([]*metadata.TxDesc, error) {
	return nil, nil
}
