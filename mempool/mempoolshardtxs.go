package mempool

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
)

/*
	Verify Transaction with these condition:
	1. Validate with current mempool:
	2. Validate Init Custom Token
	3. Check tx existed in block
	4. Check duplicate staker public key in block
	5. Check duplicate Init Custom Token in block
*/
func (tp *TxPool) ValidateTxList(txs []metadata.Transaction) error {
	var errCh chan error
	errCh = make(chan error)
	validTxCount := 0
	// salaryTxCount := 0
	//validate individual tx
	for _, tx := range txs {
		go func(tx metadata.Transaction) {
			err := tp.validateTxIndependentProperties(tx)
			errCh <- err
		}(tx)
	}

	for {
		err := <-errCh
		if err != nil {
			return errors.New("tx in new block error:" + err.Error())
		}
		validTxCount++
		if validTxCount == len(txs) {
			break
		}
	}
	//validate txs list
	for _, tx := range txs {
		txHash := tx.Hash()
		// Don't accept the transaction if it already exists in the pool.
		if tp.isTxInPool(txHash) {
			return NewMempoolTxError(RejectDuplicateTx, fmt.Errorf("already have transaction %+v", txHash.String()))
		}

		// check tx with all txs in current mempool
		err := tx.ValidateTxWithCurrentMempool(tp)
		if err != nil {
			return err
		}

		// check duplicate stake public key ONLY with staking transaction
		if tx.GetMetadata() != nil {
			if tx.GetMetadata().GetType() == metadata.ShardStakingMeta || tx.GetMetadata().GetType() == metadata.BeaconStakingMeta {
				pubkey := base58.Base58Check{}.Encode(tx.GetSigPubKey(), common.ZeroByte)
				found := common.IndexOfStrInHashMap(pubkey, tp.poolCandidate)
				if found > 0 {
					return NewMempoolTxError(RejectDuplicateStakePubkey, fmt.Errorf("This public key already stake and still in pool %+v", pubkey))
				}
			}
		}

		shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		bestHeight := tp.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Height
		txFee := tx.GetTxFee()
		txFeeToken := tx.GetTxFeeToken()
		txD := createTxDescMempool(tx, bestHeight, txFee, txFeeToken)
		tp.addTx(txD, false)
	}

	return nil
}

/*
SKIP salary transaction
Verify Transaction with these condition:
	1. Validate tx version
	2. Validate fee with tx size
	3. Validate type of tx
	4. Validate sanity data of tx
	5. Validate By it self (data in tx): privacy proof, metadata,...
	6. Validate tx with blockchain: douple spend, ...
*/
func (tp *TxPool) validateTxIndependentProperties(tx metadata.Transaction) error {
	var shardID byte
	var err error
	txHash := tx.Hash()

	if tx.IsSalaryTx() {
		return nil
	}
	// check version
	ok := tx.CheckTxVersion(maxVersion)
	if !ok {
		return NewMempoolTxError(RejectVersion, fmt.Errorf("transaction %+v's version is invalid", txHash.String()))
	}

	// check actual size
	actualSize := tx.GetTxActualSize()
	Logger.log.Debugf("Transaction %+v 's size %+v \n", *txHash, actualSize)
	if actualSize >= common.MaxBlockSize || actualSize >= common.MaxTxSize {
		return NewMempoolTxError(RejectInvalidSize, fmt.Errorf("transaction %+v's size is invalid, more than %+v Kilobyte", txHash.String(), common.MaxBlockSize))
	}

	// check fee of tx
	// limitFee := tp.config.FeeEstimator[shardID].limitFee
	// txFee := tx.GetTxFee()
	// ok = tx.CheckTransactionFee(limitFee)
	// if !ok {
	// 	return NewMempoolTxError(RejectInvalidFee, fmt.Errorf("transaction %+v has %d fees which is under the required amount of %d", tx.Hash().String(), txFee, limitFee*tx.GetTxActualSize()))
	// }
	// end check with policy

	ok = tx.ValidateType()
	if !ok {
		return err
	}

	// sanity data
	if validated, errS := tx.ValidateSanityData(tp.config.BlockChain, 0); !validated {
		return NewMempoolTxError(RejectSanityTx, fmt.Errorf("transaction's sansity %v is error %v", txHash.String(), errS.Error()))
	}

	// ValidateTransaction tx by it self
	shardID = common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	validated, _ := tx.ValidateTxByItself(tx.IsPrivacy(), tp.config.BlockChain.GetDatabase(), tp.config.BlockChain, shardID, false)
	if !validated {
		return NewMempoolTxError(RejectInvalidTx, errors.New("invalid tx"))
	}
	// validate tx with data of blockchain
	err = tx.ValidateTxWithBlockChain(tp.config.BlockChain, shardID, tp.config.BlockChain.GetDatabase())
	if err != nil {
		return err
	}
	return nil
}
