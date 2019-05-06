package mempool

import (
	"encoding/json"
	"strings"
	"time"
	
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/databasemp/lvdb"
)
type TempDesc struct {
	StartTime time.Time
	IsPushMessage bool
	Height uint64
	Fee uint64
	FeePerKB int32
}
func (tp *TxPool) AddTransactionToDatabaseMP(txHash *common.Hash, txDesc TxDesc) error {
	tx := txDesc.Desc.Tx
	tempDesc := TempDesc{
		StartTime: txDesc.StartTime,
		IsPushMessage:txDesc.IsFowardMessage,
		Height:txDesc.Desc.Height,
		Fee:txDesc.Desc.Fee,
		FeePerKB: txDesc.Desc.FeePerKB,
	}
	switch tx.GetType() {
	//==================For Constant Transfer Only
	case common.TxNormalType:
		{
			normalTx := tx.(*transaction.Tx)
			valueTx, err := json.Marshal(normalTx)
			if err != nil {
				return err
			}
			valueDesc, err := json.Marshal(tempDesc)
			if err != nil {
				return err
			}
			err = tp.config.DataBaseMempool.AddTransaction(txHash, common.TxNormalType, valueTx, valueDesc)
			if err != nil {
				return err
			}
		}
	//==================For Constant & TxCustomToken Transfer
	case common.TxCustomTokenType:
		{
			customTokenTx := tx.(*transaction.TxCustomToken)
			valueTx, err := json.Marshal(customTokenTx)
			if err != nil {
				return err
			}
			valueDesc, err := json.Marshal(tempDesc)
			if err != nil {
				return err
			}
			err = tp.config.DataBaseMempool.AddTransaction(txHash, common.TxCustomTokenType, valueTx, valueDesc)
			if err != nil {
				return err
			}
		}
	case common.TxCustomTokenPrivacyType:
		{
			customTokenPrivacyTx := tx.(*transaction.TxCustomTokenPrivacy)
			valueTx, err := json.Marshal(customTokenPrivacyTx)
			if err != nil {
				return err
			}
			valueDesc, err := json.Marshal(tempDesc)
			if err != nil {
				return err
			}
			err = tp.config.DataBaseMempool.AddTransaction(txHash, common.TxCustomTokenPrivacyType, valueTx, valueDesc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tp *TxPool) GetTransactionFromDatabaseMP(txHash *common.Hash) (TxDesc, error) {
	value, err := tp.config.DataBaseMempool.GetTransaction(txHash)
	values := strings.Split(string(value),string(lvdb.Splitter))
	txDesc, err := UmmarshallTxDescFromDatabase(values[0], []byte(values[1]), []byte(values[2]))
	if err != nil {
		return TxDesc{}, err
	}
	return txDesc, nil
}
func (tp *TxPool) ResetDatabaseMP() error {
	return tp.config.DataBaseMempool.Reset()
}
func (tp *TxPool) LoadDatabaseMP() ([]TxDesc,error) {
	txDescs := []TxDesc{}
	allTxHashes, allTxs, err := tp.config.DataBaseMempool.Load()
	ttl := time.Duration(tp.TxLifeTime) * time.Second
	if err != nil {
		return txDescs, err
	}
	for index, tx := range allTxs {
		values := strings.Split(string(tx),string(lvdb.Splitter))
		txDesc, err := UmmarshallTxDescFromDatabase(values[0], []byte(values[1]), []byte(values[2]))
		if err != nil {
			txHash,err := common.NewHash(allTxHashes[index][3:])
			if err != nil {
				continue
			}
			// fail to ummarshall transaction then remove
			tp.RemoveTransactionFromDatabaseMP(txHash)
			continue
		}
		//if transaction is timeout then remove
		if time.Since(txDesc.StartTime) > ttl {
			tp.RemoveTransactionFromDatabaseMP(txDesc.Desc.Tx.Hash())
		}
		//if not validated by current blockchain db then remove
		err = tp.ValidateTransaction(txDesc.Desc.Tx)
		if err != nil {
			tp.RemoveTransactionFromDatabaseMP(txDesc.Desc.Tx.Hash())
			continue
		}
		
		tp.addTx(&txDesc, false)
		txDescs = append(txDescs, txDesc)
	}
	return txDescs, nil
}
func (tp *TxPool) RemoveTransactionFromDatabaseMP(txHash *common.Hash) (error){
	if has, _ := tp.config.DataBaseMempool.HasTransaction(txHash); has {
		err := tp.config.DataBaseMempool.RemoveTransaction(txHash)
		return err
	}
	return nil
}

func UmmarshallTxDescFromDatabase(txType string, valueTx []byte, valueDesc []byte) (TxDesc, error) {
	txDesc := TxDesc{}
	switch txType {
	case common.TxNormalType:
		{
		tx := transaction.Tx{}
		err := json.Unmarshal(valueTx, &tx)
		if err != nil {
			return txDesc, err
		}
		
		txDesc.Desc.Tx = &tx
		}
	case common.TxCustomTokenType:
		{
			customTokenTx := transaction.TxCustomToken{}
			err := json.Unmarshal(valueTx, &customTokenTx)
			if err != nil {
				return txDesc, err
			}
			txDesc.Desc.Tx = &customTokenTx
		}
	case common.TxCustomTokenPrivacyType:
		{
			customTokenPrivacyTx := transaction.TxCustomTokenPrivacy{}
			err := json.Unmarshal(valueTx, &customTokenPrivacyTx)
			if err != nil {
				return txDesc, err
			}
			txDesc.Desc.Tx = &customTokenPrivacyTx
		}
	}
	tempDesc := TempDesc{}
	err := json.Unmarshal(valueDesc, &tempDesc)
	if err != nil {
		return txDesc, err
	}
	txDesc.IsFowardMessage = tempDesc.IsPushMessage
	txDesc.StartTime = tempDesc.StartTime
	txDesc.Desc.Height = tempDesc.Height
	txDesc.Desc.Fee = tempDesc.Fee
	txDesc.Desc.FeePerKB = tempDesc.FeePerKB
	return txDesc, nil
}
