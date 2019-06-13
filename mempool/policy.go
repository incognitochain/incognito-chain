package mempool

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/transaction"
)

// Policy houses the policy (configuration parameters) which is used to control the mempool.
type Policy struct {
	// MaxTxVersion is the transaction version that the mempool should
	// accept.  All transactions above this version are rejected as
	// non-standard.
	MaxTxVersion int8

	BlockChain *blockchain.BlockChain
}

/*

 */
// func (self *Policy) CheckTxVersion(tx metadata.Transaction) bool {
// 	txType := tx.GetType()
// 	switch txType {
// 	case common.TxSalaryType, common.TxNormalType:
// 		{
// 			temp := (*tx).(*transaction.Tx)
// 			if temp.Version > self.MaxTxVersion {
// 				return false
// 			}
// 		}
// 	}
// 	return true
// }

// return min transacton fee required for a transaction that we accepted into the memmory pool and replayed.
func (self *Policy) calcMinFeeTxCustomTokenAccepted(tx *transaction.TxCustomToken) uint64 {
	// return self.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.TxFee
	return 1
}

// return min transacton fee required for a transaction that we accepted into the memmory pool and replayed.
func (self *Policy) calcMinFeeTxAccepted(tx *transaction.Tx) uint64 {
	// return self.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.TxFee
	return 1
}

/*

 */
func (self *Policy) CheckTransactionFee(tx *transaction.Tx) error {
	minFee := self.calcMinFeeTxAccepted(tx)
	if tx.Fee < minFee {
		str := fmt.Sprintf("transaction %+v has %d fees which is under the required amount of %d", tx.Hash().String(), tx.Fee, minFee)
		err := MempoolTxError{}
		err.Init(RejectInvalidFee, errors.New(str))
		return err
	}
	return nil
}

func (self *Policy) CheckCustomTokenTransactionFee(tx *transaction.TxCustomToken) error {
	minFee := self.calcMinFeeTxCustomTokenAccepted(tx)
	if tx.Fee < minFee {
		str := fmt.Sprintf("transaction %+v has %d fees which is under the required amount of %d", tx.Hash().String(), tx.Fee, minFee)
		err := MempoolTxError{}
		err.Init(RejectInvalidFee, errors.New(str))
		return err
	}
	return nil
}
