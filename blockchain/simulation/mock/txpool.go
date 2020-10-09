package mock

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

type TxPool struct{}

func (tp *TxPool) HaveTransaction(hash *common.Hash) bool {}

func (tp *TxPool) RemoveTx(txs []metadata.Transaction, isInBlock bool) {}
func (tp *TxPool) RemoveCandidateList([]string)                        {}
func (tp *TxPool) EmptyPool() bool                                     {}
func (tp *TxPool) MaybeAcceptTransactionForBlockProducing(metadata.Transaction, int64, *ShardBestState) (*metadata.TxDesc, error) {
}
func (tp *TxPool) MaybeAcceptBatchTransactionForBlockProducing(byte, []metadata.Transaction, int64, *ShardBestState) ([]*metadata.TxDesc, error) {
}
