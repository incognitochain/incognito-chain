package rpcservice

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

type TxMemPoolService struct {
	TxMemPool MempoolInterface
}

func (txMemPoolService TxMemPoolService) GetPoolCandidate() map[common.Hash]string {
	return txMemPoolService.TxMemPool.GetClonedPoolCandidate()
}

func (txMemPoolService TxMemPoolService) FilterMemPoolOutcoinsToSpent(outCoins []*privacy.OutputCoin) ([]*privacy.OutputCoin, error) {
	remainOutputCoins := make([]*privacy.OutputCoin, 0)

	for _, outCoin := range outCoins {
		if txMemPoolService.TxMemPool.ValidateSerialNumberHashH(outCoin.CoinDetails.GetSerialNumber().ToBytesS()) == nil {
			remainOutputCoins = append(remainOutputCoins, outCoin)
		}
	}
	return remainOutputCoins, nil
}

func (txMemPoolService TxMemPoolService) GetNumberOfTxsInMempool() int {
	return len(txMemPoolService.TxMemPool.ListTxs())
}

func (txMemPoolService TxMemPoolService) MempoolEntry(txIDString string) (metadata.Transaction, byte, *RPCError) {
	txID, err := common.Hash{}.NewHashFromStr(txIDString)
	if err != nil {
		Logger.log.Debugf("handleMempoolEntry result: nil %+v", err)
		return nil, byte(0), NewRPCError(RPCInvalidParamsError, err)
	}

	txInPool, err := txMemPoolService.TxMemPool.GetTx(txID)
	if err != nil {
		Logger.log.Error(err)
		return nil, byte(0), NewRPCError(GeTxFromPoolError, err)
	}
	shardIDTemp := common.GetShardIDFromLastByte(txInPool.GetSenderAddrLastByte())

	return txInPool, shardIDTemp, nil
}

func (txMemPoolService *TxMemPoolService) RemoveTxInMempool(txIDString string) (bool, *RPCError) {
	txID, err := common.Hash{}.NewHashFromStr(txIDString)
	if err != nil {
		Logger.log.Debugf("RemoveTxInMempool result: nil %+v", err)
		return false, NewRPCError(RPCInvalidParamsError, err)
	}

	tempTx, err := txMemPoolService.TxMemPool.GetTx(txID)
	if err != nil {
		return false, NewRPCError(GeTxFromPoolError, err)
	}
	txMemPoolService.TxMemPool.RemoveTx([]metadata.Transaction{tempTx}, false)
	txMemPoolService.TxMemPool.TriggerCRemoveTxs(tempTx)

	return true, nil
}

// CheckListSerialNumbersExistedInMempool checks list of serial numbers (base58 encoded) are existed in mempool or not
func (txMemPoolService TxMemPoolService) CheckListSerialNumbersExistedInMempool(serialNumbers []interface{}) ([]bool, *RPCError) {
	isExisteds := []bool{}
	for _, item := range serialNumbers {
		snStr, okParam := item.(string)
		if !okParam {
			return nil, NewRPCError(RPCInvalidParamsError, fmt.Errorf("Serial number must be a string, %+v", item))
		}
		snBytes, version, err := base58.Base58Check{}.Decode(snStr)
		if err != nil || version != common.ZeroByte || len(snBytes) == 0 {
			return []bool{}, NewRPCError(RPCInvalidParamsError, fmt.Errorf("Serial number %v is invalid format", snStr))
		}
		isExisted := txMemPoolService.TxMemPool.ValidateSerialNumberHashH(snBytes) != nil
		isExisteds = append(isExisteds, isExisted)
	}
	return isExisteds, nil
}