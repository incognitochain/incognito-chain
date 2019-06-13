package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/incognitochain/incognito-chain/mempool"

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/incognitochain/incognito-chain/wire"
)

//handleListOutputCoins - use readonly key to get all tx which contains output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
//component:
//Parameter #1—the minimum number of confirmations an output must have
//Parameter #2—the maximum number of confirmations an output may have
//Parameter #3—the list paymentaddress-readonlykey which be used to view list outputcoin
//Parameter #4 - optional - token id - default prv coin
func (rpcServer RpcServer) handleListOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleListOutputCoins params: %+v", params)
	result := jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}
	
	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Infof("handleListOutputCoins result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	minTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Infof("handleListOutputCoins result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	min := int(minTemp)
	maxTemp, ok := paramsArray[1].(float64)
	if !ok {
		Logger.log.Infof("handleListOutputCoins result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	max := int(maxTemp)
	_ = min
	_ = max
	//#3: list key component
	listKeyParams := common.InterfaceSlice(paramsArray[2])
	
	//#4: optional token type - default prv coin
	tokenID := &common.Hash{}
	tokenID.SetBytes(common.PRVCoinID[:])
	if len(paramsArray) > 3 {
		var err1 error
		tokenID, err1 = common.Hash{}.NewHashFromStr(paramsArray[3].(string))
		if err1 != nil {
			Logger.log.Infof("handleListOutputCoins result: %+v, err: %+v", nil, err1)
			return nil, NewRPCError(ErrListCustomTokenNotFound, err1)
		}
	}
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})
		
		// get keyset only contain readonly-key by deserializing
		readonlyKeyStr := keys["ReadonlyKey"].(string)
		readonlyKey, err := wallet.Base58CheckDeserialize(readonlyKeyStr)
		if err != nil {
			Logger.log.Infof("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrUnexpected, err)
		}
		
		// get keyset only contain pub-key by deserializing
		pubKeyStr := keys["PaymentAddress"].(string)
		pubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			Logger.log.Infof("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrUnexpected, err)
		}
		
		// create a key set
		keySet := cashec.KeySet{
			ReadonlyKey:    readonlyKey.KeySet.ReadonlyKey,
			PaymentAddress: pubKey.KeySet.PaymentAddress,
		}
		lastByte := keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		outputCoins, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&keySet, shardIDSender, tokenID)
		if err != nil {
			Logger.log.Infof("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrUnexpected, err)
		}
		item := make([]jsonresult.OutCoin, 0)
		
		for _, outCoin := range outputCoins {
			if outCoin.CoinDetails.Value == 0 {
				continue
			}
			item = append(item, jsonresult.OutCoin{
				//SerialNumber:   base58.Base58Check{}.Encode(outCoin.CoinDetails.SerialNumber.Compress(), common.ZeroByte),
				PublicKey:      base58.Base58Check{}.Encode(outCoin.CoinDetails.PublicKey.Compress(), common.ZeroByte),
				Value:          outCoin.CoinDetails.Value,
				Info:           base58.Base58Check{}.Encode(outCoin.CoinDetails.Info[:], common.ZeroByte),
				CoinCommitment: base58.Base58Check{}.Encode(outCoin.CoinDetails.CoinCommitment.Compress(), common.ZeroByte),
				Randomness:     base58.Base58Check{}.Encode(outCoin.CoinDetails.Randomness.Bytes(), common.ZeroByte),
				SNDerivator:    base58.Base58Check{}.Encode(outCoin.CoinDetails.SNDerivator.Bytes(), common.ZeroByte),
			})
		}
		result.Outputs[readonlyKeyStr] = item
	}
	Logger.log.Infof("handleListOutputCoins result: %+v", result)
	return result, nil
}

/*
// handleCreateTransaction handles createtransaction commands.
*/
func (rpcServer RpcServer) handleCreateRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCreateRawTransaction params: %+v", params)
	var err error
	tx, err := rpcServer.buildRawTransaction(params, nil)
	if err.(*RPCError) != nil {
		Logger.log.Critical(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
		ShardID:         txShardID,
	}
	Logger.log.Infof("handleCreateRawTransaction result: %+v", result)
	return result, nil
}

/*
// handleSendTransaction implements the sendtransaction command.
Parameter #1—a serialized transaction to broadcast
Parameter #2–whether to allow high fees
Result—a TXID or error Message
*/
func (rpcServer RpcServer) handleSendRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleSendRawTransaction params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckData := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	var tx transaction.Tx
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	hash, _, err := rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	//rpcServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		mempoolErr, ok := err.(mempool.MempoolTxError)
		if ok {
			if mempoolErr.Code == mempool.ErrCodeMessage[mempool.RejectInvalidFee].Code {
				Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
				return nil, NewRPCError(ErrRejectInvalidFee, mempoolErr)
			}
		}
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	Logger.log.Infof("New transaction hash: %+v \n", *hash)
	
	// broadcast Message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	txMsg.(*wire.MessageTx).Transaction = &tx
	err = rpcServer.config.Server.PushMessageToAll(txMsg)
	if err == nil {
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		rpcServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	}
	
	txID := tx.Hash().String()
	result := jsonresult.CreateTransactionResult{
		TxID: txID,
	}
	Logger.log.Infof("\n\n\n\n\n\nhandleSendRawTransaction result: %+v\n\n\n\n\n", result)
	return result, nil
}

/*
handleCreateAndSendTx - RPC creates transaction and send to network
*/
func (rpcServer RpcServer) handleCreateAndSendTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCreateAndSendTx params: %+v", params)
	var err error
	data, err := rpcServer.handleCreateRawTransaction(params, closeChan)
	if err.(*RPCError) != nil {
		Logger.log.Infof("handleCreateAndSendTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	if err.(*RPCError) != nil {
		Logger.log.Infof("handleCreateAndSendTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:    sendResult.(jsonresult.CreateTransactionResult).TxID,
		ShardID: tx.ShardID,
	}
	Logger.log.Infof("handleCreateAndSendTx result: %+v", result)
	return result, nil
}

/*
handleGetMempoolInfo - RPC returns information about the node's current txs memory pool
*/
func (rpcServer RpcServer) handleGetMempoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetMempoolInfo params: %+v", params)
	result := jsonresult.GetMempoolInfo{}
	result.Size = rpcServer.config.TxMemPool.Count()
	result.Bytes = rpcServer.config.TxMemPool.Size()
	result.MempoolMaxFee = rpcServer.config.TxMemPool.MaxFee()
	listTxsDetail := rpcServer.config.TxMemPool.ListTxsDetail()
	if len(listTxsDetail) > 0 {
		result.ListTxs = make([]jsonresult.GetMempoolInfoTx, 0)
		for _, tx := range listTxsDetail {
			item := jsonresult.GetMempoolInfoTx{
				LockTime: tx.GetLockTime(),
				TxID:     tx.Hash().String(),
			}
			result.ListTxs = append(result.ListTxs, item)
		}
	}
	Logger.log.Infof("handleGetMempoolInfo result: %+v", result)
	return result, nil
}

func (rpcServer RpcServer) revertTxToResponseObject(tx metadata.Transaction, blockHash *common.Hash, blockHeight uint64, index int, shardID byte) (*jsonresult.TransactionDetail, *RPCError) {
	var result *jsonresult.TransactionDetail
	blockHashStr := ""
	if blockHash != nil {
		blockHashStr = blockHash.String()
	}
	switch tx.GetType() {
	case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
		{
			tempTx := tx.(*transaction.Tx)
			result = &jsonresult.TransactionDetail{
				BlockHash:   blockHashStr,
				BlockHeight: blockHeight,
				Index:       uint64(index),
				ShardID:     shardID,
				Hash:        tx.Hash().String(),
				Version:     tempTx.Version,
				Type:        tempTx.Type,
				LockTime:    time.Unix(tempTx.LockTime, 0).Format(common.DateOutputFormat),
				Fee:         tempTx.Fee,
				IsPrivacy:   tempTx.IsPrivacy(),
				Proof:       tempTx.Proof,
				SigPubKey:   tempTx.SigPubKey,
				Sig:         tempTx.Sig,
			}
			if result.Proof != nil && len(result.Proof.InputCoins) > 0 && result.Proof.InputCoins[0].CoinDetails.PublicKey != nil {
				result.InputCoinPubKey = base58.Base58Check{}.Encode(result.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), common.ZeroByte)
			}
			if tempTx.Metadata != nil {
				metaData, _ := json.MarshalIndent(tempTx.Metadata, "", "\t")
				result.Metadata = string(metaData)
			}
			if result.Proof != nil {
				result.ProofDetail.ConvertFromProof(result.Proof)
			}
		}
	case common.TxCustomTokenType:
		{
			tempTx := tx.(*transaction.TxCustomToken)
			result = &jsonresult.TransactionDetail{
				BlockHash:   blockHashStr,
				BlockHeight: blockHeight,
				Index:       uint64(index),
				ShardID:     shardID,
				Hash:        tx.Hash().String(),
				Version:     tempTx.Version,
				Type:        tempTx.Type,
				LockTime:    time.Unix(tempTx.LockTime, 0).Format(common.DateOutputFormat),
				Fee:         tempTx.Fee,
				Proof:       tempTx.Proof,
				SigPubKey:   tempTx.SigPubKey,
				Sig:         tempTx.Sig,
			}
			txCustomData, _ := json.MarshalIndent(tempTx.TxTokenData, "", "\t")
			result.CustomTokenData = string(txCustomData)
			if result.Proof != nil && len(result.Proof.InputCoins) > 0 && result.Proof.InputCoins[0].CoinDetails.PublicKey != nil {
				result.InputCoinPubKey = base58.Base58Check{}.Encode(result.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), common.ZeroByte)
			}
			if tempTx.Metadata != nil {
				metaData, _ := json.MarshalIndent(tempTx.Metadata, "", "\t")
				result.Metadata = string(metaData)
			}
			if result.Proof != nil {
				result.ProofDetail.ConvertFromProof(result.Proof)
			}
		}
	case common.TxCustomTokenPrivacyType:
		{
			tempTx := tx.(*transaction.TxCustomTokenPrivacy)
			result = &jsonresult.TransactionDetail{
				BlockHash:   blockHashStr,
				BlockHeight: blockHeight,
				Index:       uint64(index),
				ShardID:     shardID,
				Hash:        tx.Hash().String(),
				Version:     tempTx.Version,
				Type:        tempTx.Type,
				LockTime:    time.Unix(tempTx.LockTime, 0).Format(common.DateOutputFormat),
				Fee:         tempTx.Fee,
				Proof:       tempTx.Proof,
				SigPubKey:   tempTx.SigPubKey,
				Sig:         tempTx.Sig,
			}
			if result.Proof != nil && len(result.Proof.InputCoins) > 0 && result.Proof.InputCoins[0].CoinDetails.PublicKey != nil {
				result.InputCoinPubKey = base58.Base58Check{}.Encode(result.Proof.InputCoins[0].CoinDetails.PublicKey.Compress(), common.ZeroByte)
			}
			tokenData, _ := json.MarshalIndent(tempTx.TxTokenPrivacyData, "", "\t")
			result.PrivacyCustomTokenData = string(tokenData)
			if tempTx.Metadata != nil {
				metaData, _ := json.MarshalIndent(tempTx.Metadata, "", "\t")
				result.Metadata = string(metaData)
			}
			if result.Proof != nil {
				result.ProofDetail.ConvertFromProof(result.Proof)
			}
		}
	default:
		{
			return nil, NewRPCError(ErrTxTypeInvalid, errors.New("Tx type is invalid"))
		}
	}
	return result, nil
}

// Get transaction by Hash
func (rpcServer RpcServer) handleGetTransactionByHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetTransactionByHash params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	// param #1: transaction Hash
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Tx hash is empty"))
	}
	Logger.log.Infof("Get TransactionByHash input Param %+v", arrayParams[0].(string))
	txHashTemp, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Tx hash is invalid"))
	}
	txHash, _ := common.Hash{}.NewHashFromStr(txHashTemp)
	Logger.log.Infof("Get Transaction By Hash %+v", txHash)
	db := *(rpcServer.config.Database)
	shardID, blockHash, index, tx, err := rpcServer.config.BlockChain.GetTransactionByHash(*txHash)
	if err != nil {
		// maybe tx is still in tx mempool -> check mempool
		tx, errM := rpcServer.config.TxMemPool.GetTx(txHash)
		if errM != nil {
			return nil, NewRPCError(ErrTxNotExistedInMemAndBLock, errors.New("Tx is not existed in block or mempool"))
		}
		result, errM := rpcServer.revertTxToResponseObject(tx, nil, 0, 0, byte(0))
		if errM.(*RPCError) != nil {
			return nil, errM.(*RPCError)
		}
		result.IsInMempool = true
		return result, nil
	}
	
	blockHeight, _, err := db.GetIndexOfBlock(blockHash)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result, err := rpcServer.revertTxToResponseObject(tx, &blockHash, blockHeight, index, shardID)
	if err.(*RPCError) != nil {
		return nil, err.(*RPCError)
	}
	result.IsInBlock = true
	Logger.log.Infof("handleGetTransactionByHash result: %+v", result)
	return result, nil
}

func (self RpcServer) handleGetBlockProducerList(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := make(map[string]string)
	// for shardID, bestState := range self.config.BlockChain.BestState {
	// 	if bestState.BestBlock.BlockProducer != "" {
	// 		result[strconv.Itoa(shardID)] = bestState.BestBlock.BlockProducer
	// 	} else {
	// 		result[strconv.Itoa(shardID)] = self.config.ChainParams.GenesisBlock.Header.Committee[shardID]
	// 	}
	// }
	return result, nil
}

// handleCreateRawCustomTokenTransaction - handle create a custom token command and return in hex string format.
func (rpcServer RpcServer) handleCreateRawCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCreateRawCustomTokenTransaction params: %+v", params)
	var err error
	tx, err := rpcServer.buildRawCustomTokenTransaction(params, nil)
	if err.(*RPCError) != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	result := jsonresult.CreateTransactionCustomTokenResult{
		ShardID:         tx.Tx.PubKeyLastByteSender,
		TxID:            tx.Hash().String(),
		TokenID:         tx.TxTokenData.PropertyID.String(),
		TokenName:       tx.TxTokenData.PropertyName,
		TokenAmount:     tx.TxTokenData.Amount,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	Logger.log.Infof("handleCreateRawCustomTokenTransaction result: %+v", result)
	return result, nil
}

// handleSendRawTransaction...
func (rpcServer RpcServer) handleSendRawCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleSendRawCustomTokenTransaction params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckData, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleSendRawCustomTokenTransaction result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("param is invalid"))
	}
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		Logger.log.Infof("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	tx := transaction.TxCustomToken{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		Logger.log.Infof("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	hash, _, err := rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	//rpcServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		Logger.log.Infof("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	Logger.log.Infof("New Custom Token Transaction: %s\n", hash.String())
	
	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		Logger.log.Infof("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	txMsg.(*wire.MessageTxToken).Transaction = &tx
	err = rpcServer.config.Server.PushMessageToAll(txMsg)
	//Mark Fowarded transaction
	if err == nil {
		rpcServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	}
	result := tx.Hash()
	Logger.log.Infof("handleSendRawCustomTokenTransaction result: %+v", result)
	return result, nil
}

// handleCreateAndSendCustomTokenTransaction - create and send a tx which process on a custom token look like erc-20 on eth
func (rpcServer RpcServer) handleCreateAndSendCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCreateAndSendCustomTokenTransaction params: %+v", params)
	data, err := rpcServer.handleCreateRawCustomTokenTransaction(params, closeChan)
	if err != nil {
		Logger.log.Infof("handleCreateAndSendCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionCustomTokenResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		Logger.log.Infof("handleCreateAndSendCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txID, err := rpcServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
	if err != nil {
		Logger.log.Infof("handleCreateAndSendCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, err
	}
	Logger.log.Infof("handleCreateAndSendCustomTokenTransaction result: %+v", txID)
	return txID, nil
}

// handleGetListCustomTokenHolders - return all custom token holder
func (rpcServer RpcServer) handleGetListCustomTokenHolders(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("TokenID is invalid"))
	}
	tokenIDStr := arrayParams[0].(string)
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
	if err != nil {
		if len(arrayParams) < 1 {
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("TokenID is invalid"))
		}
	}
	result, err := rpcServer.config.BlockChain.GetListTokenHolders(tokenID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return result, nil
}

// handleGetListCustomTokenBalance - return list token + balance for one account payment address
func (rpcServer RpcServer) handleGetListCustomTokenBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetListCustomTokenBalance params: %+v", params)
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	arrayParams := common.InterfaceSlice(params)
	accountParam, ok := arrayParams[0].(string)
	if len(accountParam) == 0 || !ok {
		Logger.log.Infof("handleGetListCustomTokenBalance result: %+v", nil)
		return result, NewRPCError(ErrRPCInvalidParams, errors.New("Param is invalid"))
	}
	account, err := wallet.Base58CheckDeserialize(accountParam)
	if err != nil {
		Logger.log.Infof("handleGetListCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, nil
	}
	result.PaymentAddress = accountParam
	accountPaymentAddress := account.KeySet.PaymentAddress
	temps, err := rpcServer.config.BlockChain.ListCustomToken()
	if err != nil {
		Logger.log.Infof("handleGetListCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, tx := range temps {
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxTokenData.PropertyName
		item.Symbol = tx.TxTokenData.PropertySymbol
		item.TokenID = tx.TxTokenData.PropertyID.String()
		item.TokenImage = common.Render([]byte(item.TokenID))
		tokenID := tx.TxTokenData.PropertyID
		res, err := rpcServer.config.BlockChain.GetListTokenHolders(&tokenID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		paymentAddressInStr := base58.Base58Check{}.Encode(accountPaymentAddress.Bytes(), 0x00)
		item.Amount = res[paymentAddressInStr]
		if item.Amount == 0 {
			continue
		}
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	Logger.log.Infof("handleGetListCustomTokenBalance result: %+v", result)
	return result, nil
}

// handleGetListPrivacyCustomTokenBalance - return list privacy token + balance for one account payment address
func (rpcServer RpcServer) handleGetListPrivacyCustomTokenBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetListPrivacyCustomTokenBalance params: %+v", params)
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	arrayParams := common.InterfaceSlice(params)
	privateKey, ok := arrayParams[0].(string)
	if len(privateKey) == 0 || !ok {
		Logger.log.Infof("handleGetListPrivacyCustomTokenBalance result: %+v", nil)
		return result, NewRPCError(ErrRPCInvalidParams, errors.New("Param is invalid"))
	}
	account, err := wallet.Base58CheckDeserialize(privateKey)
	account.KeySet.ImportFromPrivateKey(&account.KeySet.PrivateKey)
	if err != nil {
		Logger.log.Infof("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	temps, listCustomTokenCrossShard, err := rpcServer.config.BlockChain.ListPrivacyCustomToken()
	if err != nil {
		Logger.log.Infof("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, tx := range temps {
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxTokenPrivacyData.PropertyName
		item.Symbol = tx.TxTokenPrivacyData.PropertySymbol
		item.TokenID = tx.TxTokenPrivacyData.PropertyID.String()
		item.TokenImage = common.Render([]byte(item.TokenID))
		tokenID := tx.TxTokenPrivacyData.PropertyID
		
		balance := uint64(0)
		// get balance for accountName in wallet
		lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		prvCoinID := &common.Hash{}
		prvCoinID.SetBytes(common.PRVCoinID[:])
		outcoints, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tokenID)
		if err != nil {
			Logger.log.Infof("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrUnexpected, err)
		}
		for _, out := range outcoints {
			balance += out.CoinDetails.Value
		}
		
		item.Amount = balance
		if item.Amount == 0 {
			continue
		}
		item.IsPrivacy = true
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	for _, customTokenCrossShard := range listCustomTokenCrossShard {
		item := jsonresult.CustomTokenBalance{}
		item.Name = customTokenCrossShard.PropertyName
		item.Symbol = customTokenCrossShard.PropertySymbol
		item.TokenID = customTokenCrossShard.TokenID.String()
		item.TokenImage = common.Render([]byte(item.TokenID))
		tokenID := customTokenCrossShard.TokenID
		
		balance := uint64(0)
		// get balance for accountName in wallet
		lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		prvCoinID := &common.Hash{}
		prvCoinID.SetBytes(common.PRVCoinID[:])
		outcoints, err := rpcServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tokenID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		for _, out := range outcoints {
			balance += out.CoinDetails.Value
		}
		
		item.Amount = balance
		if item.Amount == 0 {
			continue
		}
		item.IsPrivacy = true
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	Logger.log.Infof("handleGetListPrivacyCustomTokenBalance result: %+v", result)
	return result, nil
}

// handleCustomTokenDetail - return list tx which relate to custom token by token id
func (rpcServer RpcServer) handleCustomTokenDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCustomTokenDetail params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		Logger.log.Infof("handleCustomTokenDetail result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
	}
	tokenIDTemp, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleCustomTokenDetail result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
	}
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDTemp)
	if err != nil {
		Logger.log.Infof("handleCustomTokenDetail result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	txs, _ := rpcServer.config.BlockChain.GetCustomTokenTxsHash(tokenID)
	result := jsonresult.CustomToken{
		ListTxs: []string{},
	}
	for _, tx := range txs {
		result.ListTxs = append(result.ListTxs, tx.String())
	}
	Logger.log.Infof("handleCustomTokenDetail result: %+v", result)
	return result, nil
}

// handlePrivacyCustomTokenDetail - return list tx which relate to privacy custom token by token id
func (rpcServer RpcServer) handlePrivacyCustomTokenDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handlePrivacyCustomTokenDetail params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		Logger.log.Infof("handlePrivacyCustomTokenDetail result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
	}
	tokenIDTemp, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handlePrivacyCustomTokenDetail result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
	}
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDTemp)
	if err != nil {
		Logger.log.Infof("handlePrivacyCustomTokenDetail result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	txs, _ := rpcServer.config.BlockChain.GetPrivacyCustomTokenTxsHash(tokenID)
	result := jsonresult.CustomToken{
		ListTxs: []string{},
	}
	for _, tx := range txs {
		result.ListTxs = append(result.ListTxs, tx.String())
	}
	Logger.log.Infof("handlePrivacyCustomTokenDetail result: %+v", result)
	return result, nil
}

// handleListUnspentCustomToken - return list utxo of custom token
func (rpcServer RpcServer) handleListUnspentCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleListUnspentCustomToken params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 2 {
		Logger.log.Infof("handleListUnspentCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Not enough params"))
	}
	// param #1: paymentaddress of sender
	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleListUnspentCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("senderKey is invalid"))
	}
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam)
	if err != nil {
		Logger.log.Infof("handleListUnspentCustomToken result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKeyset := senderKey.KeySet
	
	// param #2: tokenID
	tokenIDParam, ok := arrayParams[1].(string)
	if !ok {
		Logger.log.Infof("handleListUnspentCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("tokenID is invalid"))
	}
	tokenID, _ := common.Hash{}.NewHashFromStr(tokenIDParam)
	unspentTxTokenOuts, err := rpcServer.config.BlockChain.GetUnspentTxCustomTokenVout(senderKeyset, tokenID)
	
	if err != nil {
		Logger.log.Infof("handleListUnspentCustomToken result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := []jsonresult.UnspentCustomToken{}
	for _, temp := range unspentTxTokenOuts {
		item := jsonresult.UnspentCustomToken{
			PaymentAddress:  senderKeyParam,
			Index:           temp.GetIndex(),
			TxCustomTokenID: temp.GetTxCustomTokenID().String(),
			Value:           temp.Value,
		}
		result = append(result, item)
	}
	
	Logger.log.Infof("handleListUnspentCustomToken result: %+v", result)
	return result, nil
}

// handleCreateSignatureOnCustomTokenTx - return a signature which is signed on raw custom token tx
func (rpcServer RpcServer) handleCreateSignatureOnCustomTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCreateSignatureOnCustomTokenTx params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)
	
	if err != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	senderKeyParam := arrayParams[1]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		Logger.log.Infof("handleCreateSignatureOnCustomTokenTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	
	jsSignByteArray, err := tx.GetTxCustomTokenSignature(senderKey.KeySet)
	if err != nil {
		Logger.log.Infof("handleCreateSignatureOnCustomTokenTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, errors.New("failed to sign the custom token"))
	}
	result := hex.EncodeToString(jsSignByteArray)
	Logger.log.Infof("handleCreateSignatureOnCustomTokenTx result: %+v", result)
	return result, nil
}

// handleRandomCommitments - from input of outputcoin, random to create data for create new tx
func (rpcServer RpcServer) handleRandomCommitments(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleRandomCommitments params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	
	// #1: payment address
	paymentAddressStr, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleRandomCommitments result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("PaymentAddress is invalid"))
	}
	key, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		Logger.log.Infof("handleRandomCommitments result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	lastByte := key.KeySet.PaymentAddress.Pk[len(key.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	
	// #2: available inputCoin from old outputcoin
	outputs, ok := arrayParams[1].([]interface{})
	if !ok {
		Logger.log.Infof("handleRandomCommitments result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("outputs is invalid"))
	}
	usableOutputCoins := []*privacy.OutputCoin{}
	for _, item := range outputs {
		out := jsonresult.OutCoin{}
		out.Init(item)
		i := &privacy.OutputCoin{
			CoinDetails: &privacy.Coin{
				Value: out.Value,
			},
		}
		RandomnessInBytes, _, _ := base58.Base58Check{}.Decode(out.Randomness)
		i.CoinDetails.Randomness = new(big.Int).SetBytes(RandomnessInBytes)
		
		SNDerivatorInBytes, _, _ := base58.Base58Check{}.Decode(out.SNDerivator)
		i.CoinDetails.SNDerivator = new(big.Int).SetBytes(SNDerivatorInBytes)
		
		CoinCommitmentBytes, _, _ := base58.Base58Check{}.Decode(out.CoinCommitment)
		CoinCommitment := &privacy.EllipticPoint{}
		_ = CoinCommitment.Decompress(CoinCommitmentBytes)
		i.CoinDetails.CoinCommitment = CoinCommitment
		
		PublicKeyBytes, _, _ := base58.Base58Check{}.Decode(out.PublicKey)
		PublicKey := &privacy.EllipticPoint{}
		_ = PublicKey.Decompress(PublicKeyBytes)
		i.CoinDetails.PublicKey = PublicKey
		
		InfoBytes, _, _ := base58.Base58Check{}.Decode(out.Info)
		i.CoinDetails.Info = InfoBytes
		
		usableOutputCoins = append(usableOutputCoins, i)
	}
	usableInputCoins := transaction.ConvertOutputCoinToInputCoin(usableOutputCoins)
	
	//#3 - tokenID - default constant
	tokenID := &common.Hash{}
	tokenID.SetBytes(common.PRVCoinID[:])
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[2].(string)
		if !ok {
			Logger.log.Infof("handleRandomCommitments result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("tokenID is invalid"))
		}
		tokenID, err = common.Hash{}.NewHashFromStr(tokenIDTemp)
		if err != nil {
			Logger.log.Infof("handleRandomCommitments result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrListCustomTokenNotFound, err)
		}
	}
	commitmentIndexs, myCommitmentIndexs, commitments := rpcServer.config.BlockChain.RandomCommitmentsProcess(usableInputCoins, 0, shardIDSender, tokenID)
	result := make(map[string]interface{})
	result["CommitmentIndices"] = commitmentIndexs
	result["MyCommitmentIndexs"] = myCommitmentIndexs
	temp := []string{}
	for _, commitment := range commitments {
		temp = append(temp, base58.Base58Check{}.Encode(commitment, common.ZeroByte))
	}
	result["Commitments"] = temp
	Logger.log.Infof("handleRandomCommitments result: %+v", result)
	return result, nil
}

// handleHasSerialNumbers - check list serial numbers existed in db of node
func (rpcServer RpcServer) handleHasSerialNumbers(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleHasSerialNumbers params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	
	// #1: payment address
	paymentAddressStr, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleHasSerialNumbers result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("PaymentAddress is invalid"))
	}
	key, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	lastByte := key.KeySet.PaymentAddress.Pk[len(key.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	//#2: list serialnumbers in base58check encode string
	serialNumbersStr, ok := arrayParams[1].([]interface{})
	if !ok {
		Logger.log.Infof("handleHasSerialNumbers result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("serialNumbers is invalid"))
	}
	
	// #3: optional - token ID - default is prv coin
	tokenID := &common.Hash{}
	tokenID.SetBytes(common.PRVCoinID[:]) // default is constant
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[2].(string)
		if !ok {
			Logger.log.Infof("handleHasSerialNumbers result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("serialNumbers is invalid"))
		}
		tokenID, err = (common.Hash{}).NewHashFromStr(tokenIDTemp)
		if err != nil {
			Logger.log.Infof("handleHasSerialNumbers result: %+v, err: %+v", err)
			return nil, NewRPCError(ErrListCustomTokenNotFound, err)
		}
	}
	result := make([]bool, 0)
	for _, item := range serialNumbersStr {
		serialNumber, _, _ := base58.Base58Check{}.Decode(item.(string))
		db := *(rpcServer.config.Database)
		ok, _ := db.HasSerialNumber(*tokenID, serialNumber, shardIDSender)
		if ok || err != nil {
			// serial number in db
			result = append(result, true)
		} else {
			// serial number not in db
			result = append(result, false)
		}
	}
	Logger.log.Infof("handleHasSerialNumbers result: %+v", result)
	return result, nil
}

// handleHasSerialNumbers - check list serial numbers existed in db of node
func (rpcServer RpcServer) handleHasSnDerivators(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleHasSnDerivators params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	
	// #1: payment address
	paymentAddressStr, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleHasSnDerivators result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("paymentAddress is invalid"))
	}
	key, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		Logger.log.Infof("handleHasSnDerivators result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	lastByte := key.KeySet.PaymentAddress.Pk[len(key.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	if err != nil {
		Logger.log.Infof("handleHasSnDerivators result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	//#2: list serialnumbers in base58check encode string
	snDerivatorStr, ok := arrayParams[1].([]interface{})
	if !ok {
		Logger.log.Infof("handleHasSnDerivators result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("snDerivatorStr is invalid"))
	}
	
	// #3: optional - token ID - default is prv coin
	tokenID := &common.Hash{}
	tokenID.SetBytes(common.PRVCoinID[:]) // default is constant
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[1].(string)
		if !ok {
			Logger.log.Infof("handleHasSnDerivators result: %+v", nil)
			return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
		}
		tokenID, err = (common.Hash{}).NewHashFromStr(tokenIDTemp)
		if err != nil {
			Logger.log.Infof("handleHasSnDerivators result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrListCustomTokenNotFound, err)
		}
	}
	result := make([]bool, 0)
	for _, item := range snDerivatorStr {
		snderivator, _, _ := base58.Base58Check{}.Decode(item.(string))
		db := *(rpcServer.config.Database)
		ok, err := db.HasSNDerivator(*tokenID, privacy.AddPaddingBigInt(new(big.Int).SetBytes(snderivator), privacy.BigIntSize), shardIDSender)
		if ok || err != nil {
			// serial number in db
			result = append(result, true)
		} else {
			// serial number not in db
			result = append(result, false)
		}
	}
	Logger.log.Infof("handleHasSnDerivators result: %+v", result)
	return result, nil
}

// handleCreateRawCustomTokenTransaction - handle create a custom token command and return in hex string format.
func (rpcServer RpcServer) handleCreateRawPrivacyCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCreateRawPrivacyCustomTokenTransaction params: %+v", params)
	var err error
	tx, err := rpcServer.buildRawPrivacyCustomTokenTransaction(params, nil)
	if err.(*RPCError) != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	result := jsonresult.CreateTransactionCustomTokenResult{
		ShardID:         tx.Tx.PubKeyLastByteSender,
		TxID:            tx.Hash().String(),
		TokenID:         tx.TxTokenPrivacyData.PropertyID.String(),
		TokenName:       tx.TxTokenPrivacyData.PropertyName,
		TokenAmount:     tx.TxTokenPrivacyData.Amount,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	Logger.log.Infof("handleCreateRawPrivacyCustomTokenTransaction result: %+v", result)
	return result, nil
}

// handleSendRawTransaction...
func (rpcServer RpcServer) handleSendRawPrivacyCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleSendRawPrivacyCustomTokenTransaction params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		Logger.log.Infof("handleSendRawPrivacyCustomTokenTransaction result: %+v", nil)
		return nil, NewRPCError(ErrSendTxData, errors.New("Param is invalid"))
	}
	base58CheckData, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Infof("handleSendRawPrivacyCustomTokenTransaction result: %+v", nil)
		return nil, NewRPCError(ErrSendTxData, errors.New("Param is invalid"))
	}
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		Logger.log.Infof("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	tx := transaction.TxCustomTokenPrivacy{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		Logger.log.Infof("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	hash, _, err := rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	//rpcServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		Logger.log.Infof("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	
	txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
	if err != nil {
		Logger.log.Infof("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	
	txMsg.(*wire.MessageTxPrivacyToken).Transaction = &tx
	err = rpcServer.config.Server.PushMessageToAll(txMsg)
	//Mark forwarded message
	if err == nil {
		rpcServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	}
	result := tx.Hash()
	Logger.log.Infof("handleSendRawPrivacyCustomTokenTransaction result: %+v", result)
	return result, nil
}

// handleCreateAndSendCustomTokenTransaction - create and send a tx which process on a custom token look like erc-20 on eth
func (rpcServer RpcServer) handleCreateAndSendPrivacyCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCreateAndSendPrivacyCustomTokenTransaction params: %+v", params)
	data, err := rpcServer.handleCreateRawPrivacyCustomTokenTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionCustomTokenResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		Logger.log.Infof("handleCreateAndSendPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := rpcServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err == nil {
		Logger.log.Infof("handleCreateAndSendPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return tx, nil
	}
	Logger.log.Infof("handleCreateAndSendPrivacyCustomTokenTransaction result: %+v", txId)
	return txId, nil
}

/*
// handleCreateRawStakingTransaction handles create staking
*/
func (rpcServer RpcServer) handleCreateRawStakingTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// get component
	Logger.log.Infof("handleCreateRawStakingTransaction params: %+v", params)
	paramsArray := common.InterfaceSlice(params)
	//var err error
	if len(paramsArray) < 5 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Empty staking type component"))
	}
	stakingType, ok := paramsArray[4].(float64)
	
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Invalid staking type component"))
	}
	
	senderKeyParam := paramsArray[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		Logger.log.Critical(err)
		Logger.log.Infof("handleCreateRawStakingTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Cannot get payment address"))
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	paymentAddress, _ := senderKey.Serialize(wallet.PaymentAddressType)
	fmt.Println("SA: staking from", base58.Base58Check{}.Encode(paymentAddress, common.ZeroByte))

	metadata, err := metadata.NewStakingMetadata(int(stakingType), base58.Base58Check{}.Encode(paymentAddress, common.ZeroByte), rpcServer.config.ChainParams.StakingAmountShard)

	tx, err := rpcServer.buildRawTransaction(params, metadata)
	if err.(*RPCError) != nil {
		Logger.log.Critical(err)
		Logger.log.Infof("handleCreateRawStakingTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		Logger.log.Infof("handleCreateRawStakingTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, common.ZeroByte),
		ShardID:         txShardID,
	}
	Logger.log.Infof("handleCreateRawStakingTransaction result: %+v", result)
	return result, nil
}

/*
handleCreateAndSendStakingTx - RPC creates staking transaction and send to network
*/
func (rpcServer RpcServer) handleCreateAndSendStakingTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleCreateAndSendStakingTx params: %+v", params)
	var err error
	data, err := rpcServer.handleCreateRawStakingTransaction(params, closeChan)
	if err.(*RPCError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := rpcServer.handleSendRawTransaction(newParam, closeChan)
	if err.(*RPCError) != nil {
		Logger.log.Infof("handleCreateAndSendStakingTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:    sendResult.(jsonresult.CreateTransactionResult).TxID,
		ShardID: tx.ShardID,
	}
	Logger.log.Infof("handleCreateAndSendStakingTx result: %+v", result)
	return result, nil
}
