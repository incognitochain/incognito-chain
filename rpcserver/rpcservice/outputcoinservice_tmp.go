package rpcservice

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver2"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"github.com/incognitochain/incognito-chain/wallet"
	"math"
)

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
func (coinService CoinService) GetOutputCoinInfoByHashes(txHashList []string, tokenID string) (map[string][]uint64, *RPCError) {
	res := make(map[string][]uint64)
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
			res[txHashStr] = make([]uint64, 0)
			continue
		}
		var proof privacy.Proof
		if tx.GetTokenID().String() == common.PRVIDStr {
			if tokenID != common.PRVIDStr {
				res[txHashStr] = make([]uint64, 0)
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
			res[txHashStr] = make([]uint64, 0)
			continue
		}

		tmpRes := make([]uint64, 0)
		burningPK := wallet.GetBurningPublicKey()
		tokenIDHash, err := new(common.Hash).NewHashFromStr(tokenID)
		if err != nil {
			return nil, NewRPCError(RPCInvalidParamsError, err)
		}
		for _, outCoin := range proof.GetOutputCoins() {
			pk := outCoin.GetPublicKey().ToBytesS()
			if bytes.Equal(pk, burningPK) {
				tmpRes = append(tmpRes, math.MaxUint64)
				continue
			}
			shardID := common.GetShardIDFromLastByte(pk[len(pk)-1])
			idx, err := coinService.GetOTAIdxByPublicKey(shardID, *tokenIDHash, pk)
			if err != nil {
				return nil, NewRPCError(RPCInternalError, err)
			}
			tmpRes = append(tmpRes, idx)
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
