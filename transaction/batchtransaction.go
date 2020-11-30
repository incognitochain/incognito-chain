package transaction

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge/aggregaterange"
)

type batchTransaction struct {
	txs []basemeta.Transaction
}

func NewBatchTransaction(txs []basemeta.Transaction) *batchTransaction {
	return &batchTransaction{txs: txs}
}

func (b *batchTransaction) AddTxs(txs []basemeta.Transaction) {
	b.txs = append(b.txs, txs...)
}

func (b *batchTransaction) Validate(transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, boolParams map[string]bool) (bool, error, int) {
	return b.validateBatchTxsByItself(b.txs, transactionStateDB, bridgeStateDB, boolParams)
}

func (b *batchTransaction) validateBatchTxsByItself(txList []basemeta.Transaction, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, boolParams map[string]bool) (bool, error, int) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err, -1
	}
	bulletProofList := make([]*aggregaterange.AggregatedRangeProof, 0)
	for i, tx := range txList {
		shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		hasPrivacy := tx.IsPrivacy()

		boolParams["hasPrivacy"] = tx.IsPrivacy()
		ok, err := tx.ValidateTransaction(boolParams, transactionStateDB, bridgeStateDB, shardID, prvCoinID)
		if !ok {
			return false, err, i
		}
		if tx.GetMetadata() != nil {
			if hasPrivacy {
				return false, errors.New("Metadata can not exist in privacy tx"), i
			}
			validateMetadata := tx.GetMetadata().ValidateMetadataByItself()
			if !validateMetadata {
				return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid")), i
			}
		}

		if hasPrivacy {
			if bulletProof := tx.GetProof().GetAggregatedRangeProof(); bulletProof != nil {
				bulletProofList = append(bulletProofList, bulletProof)
			}
		}
	}
	//TODO: add go routine
	ok, err, i := aggregaterange.VerifyBatchingAggregatedRangeProofs(bulletProofList)
	if err != nil {
		return false, NewTransactionErr(TxProofVerifyFailError, err), -1
	}
	if !ok {
		Logger.log.Errorf("FAILED VERIFICATION BATCH PAYMENT PROOF %d", i)
		return false, NewTransactionErr(TxProofVerifyFailError, fmt.Errorf("FAILED VERIFICATION BATCH PAYMENT PROOF %d", i)), -1
	}
	return true, nil, -1
}

