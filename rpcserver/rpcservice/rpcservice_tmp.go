package rpcservice

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver2"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

// GetCoinInfoByHashes returns a list of input coin information for a list of hashes.
func (txService TxService) GetCoinInfoByHashes(txHashList []string, tokenID string) (map[string][][]uint64, *RPCError) {
	res := make(map[string][][]uint64)
	for _, txHashStr := range txHashList {
		if _, ok := res[txHashStr]; ok {
			continue
		}
		txHash, err := common.Hash{}.NewHashFromStr(txHashStr)
		if err != nil {
			return nil, NewRPCError(RPCInvalidParamsError, fmt.Errorf("tx hash %v is invalid", txHashStr))
		}
		Logger.log.Infof("Get Transaction By Hash %+v", *txHash)
		_, _, _, _, tx, err := txService.BlockChain.GetTransactionByHash(*txHash)
		if err != nil {
			return nil, NewRPCError(RPCInternalError, fmt.Errorf("tx hash %v has not been found or has not been confirmed", txHashStr))
		}
		if tx.GetVersion() == 1 ||
			tx.GetType() == common.TxRewardType ||
			tx.GetType() == common.TxReturnStakingType ||
			tx.GetType() == common.TxConversionType {
			res[txHashStr] = make([][]uint64, 0)
			continue
		}
		var sig *tx_ver2.SigPubKey
		if tx.GetTokenID().String() == common.PRVIDStr {
			if tokenID != common.PRVIDStr {
				res[txHashStr] = make([][]uint64, 0)
				continue
			}
			sig = new(tx_ver2.SigPubKey)
			err = sig.SetBytes(tx.GetSigPubKey())
			if err != nil {
				err = fmt.Errorf("parse SigPubKey for PRV error with tx %v, %v: %v", tx.Hash().String(), tx.GetSigPubKey(), err)
				return nil, NewRPCError(RPCInternalError, err)
			}
		} else {
			txToken, ok := tx.(transaction.TransactionToken)
			if !ok {
				err = fmt.Errorf("cannot parse tx %v to transactionToken", txHash.String())
				return nil, NewRPCError(RPCInternalError, err)
			}
			if tokenID == common.PRVIDStr {
				sig = new(tx_ver2.SigPubKey)
				err = sig.SetBytes(txToken.GetTxBase().GetSigPubKey())
				if err != nil {
					err = fmt.Errorf("parse SigPubKey for PRV error with tx %v, %v: %v", tx.Hash().String(), txToken.GetTxBase().GetSigPubKey(), err)
					return nil, NewRPCError(RPCInternalError, err)
				}
			} else {
				if txToken.GetTxTokenData().GetType() == utils.CustomTokenInit || txToken.GetType() == common.TxTokenConversionType{
					res[txHashStr] = make([][]uint64, 0)
					continue
				}
				sig = new(tx_ver2.SigPubKey)
				err = sig.SetBytes(txToken.GetTxNormal().GetSigPubKey())
				if err != nil {
					err = fmt.Errorf("parse SigPubKey for token error with tx %v, %v: %v", tx.Hash().String(), txToken.GetTxNormal().GetSigPubKey(), err)
					return nil, NewRPCError(RPCInternalError, err)
				}
			}
		}

		tmpRes := make([][]uint64, 0)
		for i := range sig.Indexes {
			tmpRes2 := make([]uint64, 0)
			for j := range sig.Indexes[i] {
				tmpRes2 = append(tmpRes2, sig.Indexes[i][j].Uint64())
			}
			tmpRes = append(tmpRes, tmpRes2)
		}
		res[txHashStr] = tmpRes
	}
	return res, nil
}
