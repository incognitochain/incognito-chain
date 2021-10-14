package rpcservice

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver2"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"github.com/incognitochain/incognito-chain/wallet"
	"math"
	"strings"
)

type CoinInfo struct {
	PrvDecoys    [][]uint64 `json:"prv_decoys"`
	PrvOutputs   []string   `json:"prv_outputs"`
	TokenDecoys  [][]uint64 `json:"token_decoys,omitempty"`
	TokenOutputs []string   `json:"token_outputs,omitempty"`
}

func (coinService CoinService) GetCoinsInfoFromTx(tx metadata.Transaction) (*CoinInfo, *RPCError) {
	if tx.GetVersion() != 2 {
		return nil, NewRPCError(RPCInternalError, fmt.Errorf("must be a tx ver 2"))
	}

	res := new(CoinInfo)
	var prvTx, tokenTx metadata.Transaction
	var err *RPCError
	if tx.GetTokenID().String() == common.PRVIDStr {
		prvTx = tx
	} else {
		tmpTxToken := tx.(transaction.TransactionToken)
		prvTx = tmpTxToken.GetTxBase()
		tokenTx = tmpTxToken.GetTxNormal()
	}

	if prvTx != nil {
		res.PrvDecoys, err = getDecoysFromTx(prvTx, common.PRVIDStr)
		if err != nil && !strings.Contains(err.Error(), "parse SigPubKey for"){
			return nil, NewRPCError(RPCInternalError, err)
		}
		res.PrvOutputs, err = getOutputCoinsFromTx(prvTx, common.PRVIDStr)
		if err != nil {
			return nil, NewRPCError(RPCInternalError, err)
		}
	}
	if tokenTx != nil {
		res.TokenDecoys, err = getDecoysFromTx(tokenTx, common.ConfidentialAssetID.String())
		if err != nil && !strings.Contains(err.Error(), "parse SigPubKey for"){
			return nil, NewRPCError(RPCInternalError, err)
		}
		res.TokenOutputs, err = getOutputCoinsFromTx(prvTx, common.ConfidentialAssetID.String())
		if err != nil {
			return nil, NewRPCError(RPCInternalError, err)
		}
	}

	return res, nil
}

// getDecoysFromTx returns a list of decoys from the tx.
func getDecoysFromTx(tx metadata.Transaction, tokenID string) ([][]uint64, *RPCError) {
	res := make([][]uint64, 0)
	var err error
	if tx.GetVersion() == 1 ||
		tx.GetType() == common.TxRewardType ||
		tx.GetType() == common.TxReturnStakingType ||
		tx.GetType() == common.TxConversionType {
		return res, nil
	}
	var sig *tx_ver2.SigPubKey
	if tx.GetTokenID().String() == common.PRVIDStr {
		if tokenID != common.PRVIDStr {
			return res, nil
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
			err = fmt.Errorf("cannot parse tx %v to transactionToken", tx.Hash().String())
			return nil, NewRPCError(RPCInternalError, err)
		}
		if tokenID == common.PRVIDStr {
			sig = new(tx_ver2.SigPubKey)
			err = sig.SetBytes(txToken.GetTxBase().GetSigPubKey())
			if err != nil {
				err = fmt.Errorf("parse SigPubKey for token error with tx %v, %v: %v", tx.Hash().String(), txToken.GetTxBase().GetSigPubKey(), err)
				return nil, NewRPCError(RPCInternalError, err)
			}
		} else {
			if txToken.GetTxTokenData().GetType() == utils.CustomTokenInit || txToken.GetType() == common.TxTokenConversionType{
				return res, nil
			}
			sig = new(tx_ver2.SigPubKey)
			err = sig.SetBytes(txToken.GetTxNormal().GetSigPubKey())
			if err != nil {
				err = fmt.Errorf("parse SigPubKey for token error with tx %v, %v: %v", tx.Hash().String(), txToken.GetTxNormal().GetSigPubKey(), err)
				return nil, NewRPCError(RPCInternalError, err)
			}
		}
	}

	for i := range sig.Indexes {
		tmpRes := make([]uint64, 0)
		for j := range sig.Indexes[i] {
			tmpRes = append(tmpRes, sig.Indexes[i][j].Uint64())
		}
		res = append(res, tmpRes)
	}

	return res, nil
}

// getOutputCoinsFromTx returns a list of output coin information from the tx.
func getOutputCoinsFromTx(tx metadata.Transaction, tokenID string) ([]string, *RPCError) {
	res := make([]string, 0)
	var err error
	if tx.GetVersion() == 1 {
		return res, nil
	}
	var proof privacy.Proof
	if tx.GetTokenID().String() == common.PRVIDStr {
		if tokenID != common.PRVIDStr {
			return res, nil
		}
		proof = tx.GetProof()
	} else {
		txToken, ok := tx.(transaction.TransactionToken)
		if !ok {
			err = fmt.Errorf("cannot parse tx %v to transactionToken", tx.Hash().String())
			return nil, NewRPCError(RPCInternalError, err)
		}
		if tokenID == common.PRVIDStr {
			proof = txToken.GetTxBase().GetProof()
		} else {
			proof = txToken.GetTxNormal().GetProof()
		}
	}

	if proof == nil || proof.GetOutputCoins() == nil {
		return res, nil
	}

	if err != nil {
		return nil, NewRPCError(RPCInvalidParamsError, err)
	}
	for _, outCoin := range proof.GetOutputCoins() {
		pk := outCoin.GetPublicKey().ToBytesS()
		pkStr := base58.Base58Check{}.Encode(pk, 0)
		res = append(res, pkStr)
	}
	return res, nil
}

// GetInputCoinInfoByHashes returns a list of input coin information for a list of hashes.
func (coinService CoinService) GetInputCoinInfoByHashes(txHashList []string, tokenID string) (map[string][][]uint64, *RPCError) {
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
		_, _, _, _, tx, err := coinService.BlockChain.GetTransactionByHash(*txHash)
		if err != nil {
			return nil, NewRPCError(RPCInternalError, fmt.Errorf("tx hash %v has not been found or has not been confirmed", txHashStr))
		}
		fmt.Println("IIII", tx.String())
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

// GetOutputCoinInfoByHashes returns a list of output coin information for a list of hashes.
func (coinService CoinService) GetOutputCoinInfoByHashes(txHashList []string, tokenID string) (map[string][]string, *RPCError) {
	res := make(map[string][]string)
	for _, txHashStr := range txHashList {
		if _, ok := res[txHashStr]; ok {
			continue
		}
		txHash, err := common.Hash{}.NewHashFromStr(txHashStr)
		if err != nil {
			return nil, NewRPCError(RPCInvalidParamsError, fmt.Errorf("tx hash %v is invalid", txHashStr))
		}
		Logger.log.Infof("Get Transaction By Hash %+v", *txHash)
		_, _, _, _, tx, err := coinService.BlockChain.GetTransactionByHash(*txHash)
		if err != nil {
			return nil, NewRPCError(RPCInternalError, fmt.Errorf("tx hash %v has not been found or has not been confirmed", txHashStr))
		}
		if tx.GetVersion() == 1 {
			res[txHashStr] = make([]string, 0)
			continue
		}
		var proof privacy.Proof
		if tx.GetTokenID().String() == common.PRVIDStr {
			if tokenID != common.PRVIDStr {
				res[txHashStr] = make([]string, 0)
				continue
			}
			proof = tx.GetProof()
		} else {
			txToken, ok := tx.(transaction.TransactionToken)
			if !ok {
				err = fmt.Errorf("cannot parse tx %v to transactionToken", txHash.String())
				return nil, NewRPCError(RPCInternalError, err)
			}
			if tokenID == common.PRVIDStr {
				proof = txToken.GetTxBase().GetProof()
			} else {
				proof = txToken.GetTxNormal().GetProof()
			}
		}

		if proof == nil || proof.GetOutputCoins() == nil {
			res[txHashStr] = make([]string, 0)
			continue
		}

		tmpRes := make([]string, 0)
		//burningPK := wallet.GetBurningPublicKey()
		//tokenIDHash, err := new(common.Hash).NewHashFromStr(tokenID)
		if err != nil {
			return nil, NewRPCError(RPCInvalidParamsError, err)
		}
		for _, outCoin := range proof.GetOutputCoins() {
			pk := outCoin.GetPublicKey().ToBytesS()
			pkStr := base58.Base58Check{}.Encode(pk, 0)
			//if bytes.Equal(pk, burningPK) {
			//	tmpRes = append(tmpRes, math.MaxInt64)
			//	continue
			//}
			//shardID := common.GetShardIDFromLastByte(pk[len(pk)-1])
			//idx, err := coinService.GetOTAIdxByPublicKey(shardID, *tokenIDHash, pk)
			//if err != nil {
			//	return nil, NewRPCError(RPCInternalError, err)
			//}
			tmpRes = append(tmpRes, pkStr)
		}
		res[txHashStr] = tmpRes
	}
	return res, nil
}

// GetOTACoinByPublicKey returns an OutCoin given its public key.
func (coinService CoinService) GetOTACoinByPublicKey(shardID byte, tokenID common.Hash, pkByte []byte) (*jsonresult.OutCoin, *RPCError) {
	burningPK := wallet.GetBurningPublicKey()
	if bytes.Equal(pkByte, burningPK) {
		return nil, NewRPCError(RPCInternalError, fmt.Errorf("cannot get burning coin"))
	}
	db := coinService.BlockChain.GetBestStateShard(shardID).GetCopiedTransactionStateDB()
	idxBig, err := statedb.GetOTACoinIndex(db, tokenID, pkByte)
	if err != nil {
		return nil, NewRPCError(RPCInternalError, err)
	}
	otaCoinBytes, err := statedb.GetOTACoinByIndex(db, tokenID, idxBig.Uint64(), shardID)
	if err != nil {
		return nil, NewRPCError(RPCInternalError, err)
	}

	otaCoin := new(privacy.CoinV2)
	if err := otaCoin.SetBytes(otaCoinBytes); err != nil {
		return nil, NewRPCError(RPCInternalError, fmt.Errorf("internal error happened when parsing coin"))
	}

	outCoin := jsonresult.NewOutCoin(otaCoin)
	return &outCoin, nil
}

// GetOTAIdxByPublicKey returns an index given its public key.
func (coinService CoinService) GetOTAIdxByPublicKey(shardID byte, tokenID common.Hash, pkByte []byte) (uint64, *RPCError) {
	burningPK := wallet.GetBurningPublicKey()
	if bytes.Equal(pkByte, burningPK) {
		return math.MaxUint64, NewRPCError(RPCInternalError, fmt.Errorf("cannot get burning coin"))
	}
	db := coinService.BlockChain.GetBestStateShard(shardID).GetCopiedTransactionStateDB()
	idxBig, err := statedb.GetOTACoinIndex(db, tokenID, pkByte)
	if err != nil {
		return math.MaxUint64, NewRPCError(RPCInternalError, err)
	}
	return idxBig.Uint64(), nil
}
