package mock

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
)

type TxPool struct {
	BlockChain *blockchain.BlockChain
}

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

func (tp *TxPool) MaybeAcceptTransactionForBlockProducing(tx metadata.Transaction, beaconHeight int64, state *blockchain.ShardBestState) (*metadata.TxDesc, error) {
	var txDesc metadata.TxDesc
	txDesc.Tx = tx
	return &txDesc, nil
}

func (tp *TxPool) MaybeAcceptBatchTransactionForBlockProducing(shardID byte, txs []metadata.Transaction, beaconHeight int64, shardView *blockchain.ShardBestState) ([]*metadata.TxDesc, error) {

	// bHeight := shardView.BestBlock.Header.BeaconHeight
	beaconBlockHash := shardView.BestBlock.Header.BeaconHash
	beaconView, err := tp.BlockChain.GetBeaconViewStateDataFromBlockHash(beaconBlockHash)
	if err != nil {
		return nil, err
	}
	txDescs := []*metadata.TxDesc{}
	// txHashes := []common.Hash{}
	batch := transaction.NewBatchTransaction(txs)
	boolParams := make(map[string]bool)
	boolParams["isNewTransaction"] = false
	boolParams["isBatch"] = true
	boolParams["isNewZKP"] = tp.BlockChain.IsAfterNewZKPCheckPoint(uint64(beaconHeight))
	ok, err, _ := batch.Validate(shardView.GetCopiedTransactionStateDB(), beaconView.GetBeaconFeatureStateDB(), boolParams)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Verify Batch Transaction failed %+v", txs)
	}
	for _, tx := range txs {
		// validate tx
		// err := tp.validateTransaction(shardView, beaconView, tx, beaconHeight, true, false)
		// if err != nil {
		// 	return nil, err
		// }
		shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		bestHeight := tp.BlockChain.GetBestStateShard(byte(shardID)).BestBlock.Header.Height
		txFee := tx.GetTxFee()
		txFeeToken := tx.GetTxFeeToken()
		txD := createTxDescMempool(tx, bestHeight, txFee, txFeeToken)
		// err = tp.addTx(txD, false)
		// if err != nil {
		// 	return nil, nil, err
		// }
		txDescs = append(txDescs, &txD.Desc)
		// txHashes = append(txHashes, *tx.Hash())
	}
	return txDescs, nil
}

func (tp *TxPool) ValidateSerialNumberHashH(serialNumber []byte) error {
	return nil
}
func (tp *TxPool) MaybeAcceptTransaction(tx metadata.Transaction, beaconHeight int64) (*common.Hash, *mempool.TxDesc, error) {
	return nil, nil, nil
}
func (tp *TxPool) GetTx(txHash *common.Hash) (metadata.Transaction, error) {
	return nil, nil
}
func (tp *TxPool) GetClonedPoolCandidate() map[common.Hash]string {
	return nil
}
func (tp *TxPool) ListTxs() []string {
	return nil
}
func (tp *TxPool) TriggerCRemoveTxs(tx metadata.Transaction) {
	return
}
func (tp *TxPool) MarkForwardedTransaction(txHash common.Hash) {
	return
}
func (tp *TxPool) MaxFee() uint64 {
	return 0
}
func (tp *TxPool) ListTxsDetail() []metadata.Transaction {
	return nil
}
func (tp *TxPool) Count() int {
	return 0
}
func (tp *TxPool) Size() uint64 {
	return 0
}

func (tp *TxPool) SendTransactionToBlockGen() {
	return
}

// createTxDescMempool - return an object TxDesc for mempool from original Tx
func createTxDescMempool(tx metadata.Transaction, height uint64, fee uint64, feeToken uint64) *mempool.TxDesc {
	txDesc := &mempool.TxDesc{
		Desc: metadata.TxDesc{
			Tx:       tx,
			Height:   height,
			Fee:      fee,
			FeeToken: feeToken,
		},
		StartTime:       time.Now(),
		IsFowardMessage: false,
	}
	return txDesc
}
