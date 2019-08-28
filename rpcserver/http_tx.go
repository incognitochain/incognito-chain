package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/mempool"
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
func (httpServer *HttpServer) handleListOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleListOutputCoins params: %+v", params)
	result := jsonresult.ListOutputCoins{
		Outputs: make(map[string][]jsonresult.OutCoin),
	}

	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Debugf("handleListOutputCoins result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	minTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Debugf("handleListOutputCoins result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	min := int(minTemp)
	maxTemp, ok := paramsArray[1].(float64)
	if !ok {
		Logger.log.Debugf("handleListOutputCoins result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	max := int(maxTemp)
	_ = min
	_ = max
	//#3: list key component
	listKeyParams := common.InterfaceSlice(paramsArray[2])

	//#4: optional token type - default prv coin
	tokenID := &common.Hash{}
	err := tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, NewRPCError(ErrTokenIsInvalid, err)
	}
	if len(paramsArray) > 3 {
		var err1 error
		tokenID, err1 = common.Hash{}.NewHashFromStr(paramsArray[3].(string))
		if err1 != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err1)
			return nil, NewRPCError(ErrListCustomTokenNotFound, err1)
		}
	}
	for _, keyParam := range listKeyParams {
		keys := keyParam.(map[string]interface{})

		// get keyset only contain readonly-key by deserializing
		readonlyKeyStr := keys["ReadonlyKey"].(string)
		readonlyKey, err := wallet.Base58CheckDeserialize(readonlyKeyStr)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrUnexpected, err)
		}

		// get keyset only contain pub-key by deserializing
		pubKeyStr := keys["PaymentAddress"].(string)
		pubKey, err := wallet.Base58CheckDeserialize(pubKeyStr)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrUnexpected, err)
		}

		// create a key set
		keySet := incognitokey.KeySet{
			ReadonlyKey:    readonlyKey.KeySet.ReadonlyKey,
			PaymentAddress: pubKey.KeySet.PaymentAddress,
		}
		lastByte := keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		outputCoins, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&keySet, shardIDSender, tokenID)
		if err != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrUnexpected, err)
		}
		item := make([]jsonresult.OutCoin, 0)

		for _, outCoin := range outputCoins {
			if outCoin.CoinDetails.GetValue() == 0 {
				continue
			}
			item = append(item, jsonresult.NewOutCoin(outCoin))
		}
		result.Outputs[readonlyKeyStr] = item
	}
	Logger.log.Debugf("handleListOutputCoins result: %+v", result)
	return result, nil
}

/*
// handleCreateTransaction handles createtransaction commands.
*/
func (httpServer *HttpServer) handleCreateRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCreateRawTransaction params: %+v", params)
	var err error
	tx, err := httpServer.buildRawTransaction(params, nil)
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
	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, byteArrays, txShardID)
	Logger.log.Debugf("handleCreateRawTransaction result: %+v", result)
	return result, nil
}

/*
// handleSendTransaction implements the sendtransaction command.
Parameter #1—a serialized transaction to broadcast
Parameter #2–whether to allow high fees
Result—a TXID or error Message
*/
func (httpServer *HttpServer) handleSendRawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleSendRawTransaction params: %+v", params)
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

	hash, _, err := httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		mempoolErr, ok := err.(*mempool.MempoolTxError)
		if ok {
			if mempoolErr.Code == mempool.ErrCodeMessage[mempool.RejectInvalidFee].Code {
				Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
				return nil, NewRPCError(ErrRejectInvalidFee, mempoolErr)
			}
		}
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	Logger.log.Debugf("New transaction hash: %+v \n", *hash)

	// broadcast Message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		Logger.log.Errorf("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	err = httpServer.config.Server.PushMessageToAll(txMsg)
	if err == nil {
		Logger.log.Infof("handleSendRawTransaction result: %+v, err: %+v", nil, err)
		httpServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	}

	result := jsonresult.NewCreateTransactionResult(tx.Hash(), common.EmptyString, nil, common.GetShardIDFromLastByte(tx.PubKeyLastByteSender))
	Logger.log.Debugf("\n\n\n\n\n\nhandleSendRawTransaction result: %+v\n\n\n\n\n", result)
	return result, nil
}

/*
handleCreateAndSendTx - RPC creates transaction and send to network
*/
func (httpServer *HttpServer) handleCreateAndSendTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCreateAndSendTx params: %+v", params)
	var err error
	data, err := httpServer.handleCreateRawTransaction(params, closeChan)
	if err.(*RPCError) != nil {
		Logger.log.Debugf("handleCreateAndSendTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err.(*RPCError) != nil {
		Logger.log.Debugf("handleCreateAndSendTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, tx.ShardID)
	Logger.log.Debugf("handleCreateAndSendTx result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetTransactionHashByReceiver(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("key component invalid"))
	}
	paymentAddress := arrayParams[0]

	var keySet *incognitokey.KeySet

	if paymentAddress != "" {
		senderKey, err := wallet.Base58CheckDeserialize(paymentAddress.(string))
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("key component invalid"))
		}

		keySet = &senderKey.KeySet
	} else {
		return nil, NewRPCError(ErrUnexpected, errors.New("key component invalid"))
	}

	result, err := httpServer.config.BlockChain.GetTransactionHashByReceiver(keySet)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	return result, nil
}

// Get transaction by Hash
func (httpServer *HttpServer) handleGetTransactionByHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetTransactionByHash params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	// param #1: transaction Hash
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Tx hash is empty"))
	}
	Logger.log.Debugf("Get TransactionByHash input Param %+v", arrayParams[0].(string))
	txHashTemp, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Tx hash is invalid"))
	}
	txHash, _ := common.Hash{}.NewHashFromStr(txHashTemp)
	Logger.log.Infof("Get Transaction By Hash %+v", *txHash)
	db := *(httpServer.config.Database)
	shardID, blockHash, index, tx, err := httpServer.config.BlockChain.GetTransactionByHash(*txHash)
	if err != nil {
		// maybe tx is still in tx mempool -> check mempool
		tx, errM := httpServer.config.TxMemPool.GetTx(txHash)
		if errM != nil {
			return nil, NewRPCError(ErrTxNotExistedInMemAndBLock, errors.New("Tx is not existed in block or mempool"))
		}
		shardIDTemp := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		result, errM := jsonresult.NewTransactionDetail(tx, nil, 0, 0, shardIDTemp)
		if errM != nil {
			return nil, NewRPCError(ErrUnexpected, errM)
		}
		result.IsInMempool = true
		return result, nil
	}

	blockHeight, _, err := db.GetIndexOfBlock(blockHash)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result, err := jsonresult.NewTransactionDetail(tx, &blockHash, blockHeight, index, shardID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result.IsInBlock = true
	Logger.log.Debugf("handleGetTransactionByHash result: %+v", result)
	return result, nil
}

// handleCreateRawCustomTokenTransaction - handle create a custom token command and return in hex string format.
func (httpServer *HttpServer) handleCreateRawCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCreateRawCustomTokenTransaction params: %+v", params)
	var err error
	tx, err := httpServer.buildRawCustomTokenTransaction(params, nil)
	if err.(*RPCError) != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	result := jsonresult.CreateTransactionTokenResult{
		ShardID:         common.GetShardIDFromLastByte(tx.Tx.PubKeyLastByteSender),
		TxID:            tx.Hash().String(),
		TokenID:         tx.TxTokenData.PropertyID.String(),
		TokenName:       tx.TxTokenData.PropertyName,
		TokenAmount:     tx.TxTokenData.Amount,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	Logger.log.Debugf("handleCreateRawCustomTokenTransaction result: %+v", result)
	return result, nil
}

// handleSendRawTransaction...
func (httpServer *HttpServer) handleSendRawCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleSendRawCustomTokenTransaction params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckData, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("param is invalid"))
	}
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	tx := transaction.TxNormalToken{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	hash, _, err := httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	Logger.log.Debugf("New Custom Token Transaction: %s\n", hash.String())

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	txMsg.(*wire.MessageTxToken).Transaction = &tx
	err = httpServer.config.Server.PushMessageToAll(txMsg)
	//Mark Fowarded transaction
	if err == nil {
		httpServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	}
	result := jsonresult.CreateTransactionTokenResult{
		TxID:        tx.Hash().String(),
		TokenID:     tx.TxTokenData.PropertyID.String(),
		TokenName:   tx.TxTokenData.PropertyName,
		TokenAmount: tx.TxTokenData.Amount,
		ShardID:     common.GetShardIDFromLastByte(tx.Tx.PubKeyLastByteSender),
	}
	Logger.log.Debugf("handleSendRawCustomTokenTransaction result: %+v", result)
	return result, nil
}

// handleCreateAndSendCustomTokenTransaction - create and send a tx which process on a custom token look like erc-20 on eth
func (httpServer *HttpServer) handleCreateAndSendCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCreateAndSendCustomTokenTransaction params: %+v", params)
	data, err := httpServer.handleCreateRawCustomTokenTransaction(params, closeChan)
	if err != nil {
		Logger.log.Debugf("handleCreateAndSendCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionTokenResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txID, err := httpServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
	if err != nil {
		Logger.log.Debugf("handleCreateAndSendCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, err
	}
	Logger.log.Debugf("handleCreateAndSendCustomTokenTransaction result: %+v", txID)
	return tx, nil
}

// handleGetListCustomTokenHolders - return all custom token holder
func (httpServer *HttpServer) handleGetListCustomTokenHolders(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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
	result, err := httpServer.config.BlockChain.GetListTokenHolders(tokenID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return result, nil
}

// handleGetListCustomTokenBalance - return list token + balance for one account payment address
func (httpServer *HttpServer) handleGetListCustomTokenBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetListCustomTokenBalance params: %+v", params)
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	arrayParams := common.InterfaceSlice(params)
	accountParam, ok := arrayParams[0].(string)
	if len(accountParam) == 0 || !ok {
		Logger.log.Debugf("handleGetListCustomTokenBalance result: %+v", nil)
		return result, NewRPCError(ErrRPCInvalidParams, errors.New("Param is invalid"))
	}
	account, err := wallet.Base58CheckDeserialize(accountParam)
	if err != nil {
		Logger.log.Debugf("handleGetListCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, nil
	}
	result.PaymentAddress = accountParam
	accountPaymentAddress := account.KeySet.PaymentAddress
	temps, err := httpServer.config.BlockChain.ListCustomToken()
	if err != nil {
		Logger.log.Debugf("handleGetListCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	for _, tx := range temps {
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxTokenData.PropertyName
		item.Symbol = tx.TxTokenData.PropertySymbol
		item.TokenID = tx.TxTokenData.PropertyID.String()
		item.TokenImage = common.Render([]byte(item.TokenID))
		tokenID := tx.TxTokenData.PropertyID
		res, err := httpServer.config.BlockChain.GetListTokenHolders(&tokenID)
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
	Logger.log.Debugf("handleGetListCustomTokenBalance result: %+v", result)
	return result, nil
}

// handleGetListPrivacyCustomTokenBalance - return list privacy token + balance for one account payment address
func (httpServer *HttpServer) handleGetListPrivacyCustomTokenBalance(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance params: %+v", params)
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	arrayParams := common.InterfaceSlice(params)
	privateKey, ok := arrayParams[0].(string)
	if len(privateKey) == 0 || !ok {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v", nil)
		return result, NewRPCError(ErrRPCInvalidParams, errors.New("Param is invalid"))
	}
	account, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	err = account.KeySet.InitFromPrivateKey(&account.KeySet.PrivateKey)
	if err != nil {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	temps, listCustomTokenCrossShard, err := httpServer.config.BlockChain.ListPrivacyCustomToken()
	if err != nil {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tokenIDs := make(map[common.Hash]interface{})
	for tokenID, tx := range temps {
		tokenIDs[tokenID] = 0
		item := jsonresult.CustomTokenBalance{}
		item.Name = tx.TxPrivacyTokenData.PropertyName
		item.Symbol = tx.TxPrivacyTokenData.PropertySymbol
		item.TokenID = tx.TxPrivacyTokenData.PropertyID.String()
		item.TokenImage = common.Render([]byte(item.TokenID))
		tokenID := tx.TxPrivacyTokenData.PropertyID

		balance := uint64(0)
		// get balance for accountName in wallet
		lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		prvCoinID := &common.Hash{}
		err := prvCoinID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return nil, NewRPCError(ErrTokenIsInvalid, err)
		}
		outcoints, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tokenID)
		if err != nil {
			Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrUnexpected, err)
		}
		for _, out := range outcoints {
			balance += out.CoinDetails.GetValue()
		}

		item.Amount = balance
		if item.Amount == 0 {
			continue
		}
		item.IsPrivacy = true
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	for tokenID, customTokenCrossShard := range listCustomTokenCrossShard {
		if _, ok := tokenIDs[tokenID]; ok {
			continue
		}
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
		err := prvCoinID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return nil, NewRPCError(ErrTokenIsInvalid, err)
		}
		outcoints, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tokenID)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		for _, out := range outcoints {
			balance += out.CoinDetails.GetValue()
		}

		item.Amount = balance
		if item.Amount == 0 {
			continue
		}
		item.IsPrivacy = true
		result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)
		result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	}
	Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v", result)
	return result, nil
}

// handleGetListPrivacyCustomTokenBalance - return list privacy token + balance for one account payment address
func (httpServer *HttpServer) handleGetBalancePrivacyCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetBalancePrivacyCustomToken params: %+v", params)
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 2 {
		Logger.log.Debugf("handleGetBalancePrivacyCustomToken error: Need 2 params but get %+v", len(arrayParams))
	}
	privateKey, ok := arrayParams[0].(string)
	if len(privateKey) == 0 || !ok {
		Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v", nil)
		return result, NewRPCError(ErrRPCInvalidParams, errors.New("Private key is invalid"))
	}
	tokenID, ok := arrayParams[1].(string)
	if len(tokenID) == 0 || !ok {
		Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v", nil)
		return result, NewRPCError(ErrRPCInvalidParams, errors.New("TokenID is invalid"))
	}
	account, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	err = account.KeySet.InitFromPrivateKey(&account.KeySet.PrivateKey)
	if err != nil {
		Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	result.PaymentAddress = account.Base58CheckSerialize(wallet.PaymentAddressType)
	temps, listCustomTokenCrossShard, err := httpServer.config.BlockChain.ListPrivacyCustomToken()
	if err != nil {
		Logger.log.Debugf("handleGetListPrivacyCustomTokenBalance result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	totalValue := uint64(0)
	for tempTokenID := range temps {
		if tokenID == tempTokenID.String() {
			lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			outcoints, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tempTokenID)
			if err != nil {
				Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v, err: %+v", nil, err)
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, out := range outcoints {
				totalValue += out.CoinDetails.GetValue()
			}
		}
	}
	for tempTokenID := range listCustomTokenCrossShard {
		if tokenID == tempTokenID.String() {
			lastByte := account.KeySet.PaymentAddress.Pk[len(account.KeySet.PaymentAddress.Pk)-1]
			shardIDSender := common.GetShardIDFromLastByte(lastByte)
			outcoints, err := httpServer.config.BlockChain.GetListOutputCoinsByKeyset(&account.KeySet, shardIDSender, &tempTokenID)
			if err != nil {
				Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v, err: %+v", nil, err)
				return nil, NewRPCError(ErrUnexpected, err)
			}
			for _, out := range outcoints {
				totalValue += out.CoinDetails.GetValue()
			}
		}
	}
	Logger.log.Debugf("handleGetBalancePrivacyCustomToken result: %+v", totalValue)
	return totalValue, nil
}

// handleCustomTokenDetail - return list tx which relate to custom token by token id
func (httpServer *HttpServer) handleCustomTokenDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCustomTokenDetail params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		Logger.log.Debugf("handleCustomTokenDetail result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
	}
	tokenIDTemp, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleCustomTokenDetail result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
	}
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDTemp)
	if err != nil {
		Logger.log.Debugf("handleCustomTokenDetail result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	txs, _ := httpServer.config.BlockChain.GetCustomTokenTxsHash(tokenID)
	result := jsonresult.CustomToken{
		ListTxs: []string{},
	}
	for _, tx := range txs {
		result.ListTxs = append(result.ListTxs, tx.String())
	}
	Logger.log.Debugf("handleCustomTokenDetail result: %+v", result)
	return result, nil
}

// handlePrivacyCustomTokenDetail - return list tx which relate to privacy custom token by token id
func (httpServer *HttpServer) handlePrivacyCustomTokenDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handlePrivacyCustomTokenDetail params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		Logger.log.Debugf("handlePrivacyCustomTokenDetail result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
	}
	tokenIDTemp, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handlePrivacyCustomTokenDetail result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
	}
	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDTemp)
	if err != nil {
		Logger.log.Debugf("handlePrivacyCustomTokenDetail result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	txs, _ := httpServer.config.BlockChain.GetPrivacyCustomTokenTxsHash(tokenID)
	result := jsonresult.CustomToken{
		ListTxs: []string{},
	}
	for _, tx := range txs {
		result.ListTxs = append(result.ListTxs, tx.String())
	}
	Logger.log.Debugf("handlePrivacyCustomTokenDetail result: %+v", result)
	return result, nil
}

// handleListUnspentCustomToken - return list utxo of custom token
func (httpServer *HttpServer) handleListUnspentCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleListUnspentCustomToken params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 2 {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Not enough params"))
	}
	// param #1: paymentaddress of sender
	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("senderKey is invalid"))
	}
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam)
	if err != nil {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKeyset := senderKey.KeySet

	// param #2: tokenID
	tokenIDParam, ok := arrayParams[1].(string)
	if !ok {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("tokenID is invalid"))
	}
	tokenID, _ := common.Hash{}.NewHashFromStr(tokenIDParam)
	unspentTxTokenOuts, err := httpServer.config.BlockChain.GetUnspentTxCustomTokenVout(senderKeyset, tokenID)

	if err != nil {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := []jsonresult.UnspentCustomToken{}
	for _, temp := range unspentTxTokenOuts {
		item := jsonresult.UnspentCustomToken{
			PaymentAddress: senderKeyParam,
			Index:          temp.GetIndex(),
			TxHash:         temp.GetTxCustomTokenID().String(),
			Value:          temp.Value,
		}
		result = append(result, item)
	}

	Logger.log.Debugf("handleListUnspentCustomToken result: %+v", result)
	return result, nil
}

// handleListUnspentCustomToken - return list utxo of custom token
func (httpServer *HttpServer) handleGetBalanceCustomToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetBalanceCustomToken params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 2 {
		Logger.log.Debugf("handleListUnspentCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Not enough params"))
	}
	// param #1: paymentaddress of sender
	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleGetBalanceCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("senderKey is invalid"))
	}
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam)
	if err != nil {
		Logger.log.Debugf("handleGetBalanceCustomToken result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKeyset := senderKey.KeySet

	// param #2: tokenID
	tokenIDParam, ok := arrayParams[1].(string)
	if !ok {
		Logger.log.Debugf("handleGetBalanceCustomToken result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("tokenID is invalid"))
	}
	tokenID, _ := common.Hash{}.NewHashFromStr(tokenIDParam)
	unspentTxTokenOuts, err := httpServer.config.BlockChain.GetUnspentTxCustomTokenVout(senderKeyset, tokenID)

	if err != nil {
		Logger.log.Debugf("handleGetBalanceCustomToken result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	totalValue := uint64(0)
	for _, temp := range unspentTxTokenOuts {
		totalValue += temp.Value
	}

	Logger.log.Debugf("handleGetBalanceCustomToken result: %+v", totalValue)
	return totalValue, nil
}

// handleCreateSignatureOnCustomTokenTx - return a signature which is signed on raw custom token tx
func (httpServer *HttpServer) handleCreateSignatureOnCustomTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCreateSignatureOnCustomTokenTx params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	tx := transaction.TxNormalToken{}
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
	err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		Logger.log.Debugf("handleCreateSignatureOnCustomTokenTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}

	jsSignByteArray, err := tx.GetTxCustomTokenSignature(senderKey.KeySet)
	if err != nil {
		Logger.log.Debugf("handleCreateSignatureOnCustomTokenTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, errors.New("failed to sign the custom token"))
	}
	result := hex.EncodeToString(jsSignByteArray)
	Logger.log.Debugf("handleCreateSignatureOnCustomTokenTx result: %+v", result)
	return result, nil
}

// handleRandomCommitments - from input of outputcoin, random to create data for create new tx
func (httpServer *HttpServer) handleRandomCommitments(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleRandomCommitments params: %+v", params)
	arrayParams := common.InterfaceSlice(params)

	// #1: payment address
	paymentAddressStr, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleRandomCommitments result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("PaymentAddress is invalid"))
	}
	key, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		Logger.log.Debugf("handleRandomCommitments result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	lastByte := key.KeySet.PaymentAddress.Pk[len(key.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)

	// #2: available inputCoin from old outputcoin
	outputs, ok := arrayParams[1].([]interface{})
	if !ok {
		Logger.log.Debugf("handleRandomCommitments result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("outputs is invalid"))
	}
	usableOutputCoins := []*privacy.OutputCoin{}
	for _, item := range outputs {
		out, err1 := jsonresult.NewOutcoinFromInterface(item)
		if err1 != nil {
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New(fmt.Sprint("outputs is invalid", out)))
		}
		temp := big.Int{}
		temp.SetString(out.Value, 10)
		coin := &privacy.Coin{}
		coin.SetValue(temp.Uint64())
		i := &privacy.OutputCoin{
			CoinDetails: coin,
		}
		RandomnessInBytes, _, _ := base58.Base58Check{}.Decode(out.Randomness)
		i.CoinDetails.SetRandomness(new(big.Int).SetBytes(RandomnessInBytes))

		SNDerivatorInBytes, _, _ := base58.Base58Check{}.Decode(out.SNDerivator)
		i.CoinDetails.SetSNDerivator(new(big.Int).SetBytes(SNDerivatorInBytes))

		CoinCommitmentBytes, _, _ := base58.Base58Check{}.Decode(out.CoinCommitment)
		CoinCommitment := &privacy.EllipticPoint{}
		_ = CoinCommitment.Decompress(CoinCommitmentBytes)
		i.CoinDetails.SetCoinCommitment(CoinCommitment)

		PublicKeyBytes, _, _ := base58.Base58Check{}.Decode(out.PublicKey)
		PublicKey := &privacy.EllipticPoint{}
		_ = PublicKey.Decompress(PublicKeyBytes)
		i.CoinDetails.SetPublicKey(PublicKey)

		InfoBytes, _, _ := base58.Base58Check{}.Decode(out.Info)
		i.CoinDetails.SetInfo(InfoBytes)

		usableOutputCoins = append(usableOutputCoins, i)
	}
	usableInputCoins := transaction.ConvertOutputCoinToInputCoin(usableOutputCoins)

	//#3 - tokenID - default PRV
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, NewRPCError(ErrTokenIsInvalid, err)
	}
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[2].(string)
		if !ok {
			Logger.log.Debugf("handleRandomCommitments result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("tokenID is invalid"))
		}
		tokenID, err = common.Hash{}.NewHashFromStr(tokenIDTemp)
		if err != nil {
			Logger.log.Debugf("handleRandomCommitments result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrListCustomTokenNotFound, err)
		}
	}
	commitmentIndexs, myCommitmentIndexs, commitments := httpServer.config.BlockChain.RandomCommitmentsProcess(usableInputCoins, 0, shardIDSender, tokenID)
	result := jsonresult.NewRandomCommitmentResult(commitmentIndexs, myCommitmentIndexs, commitments)
	Logger.log.Debugf("handleRandomCommitments result: %+v", result)
	return result, nil
}

// handleListSerialNumbers - return list all serialnumber in shard for token ID
func (httpServer *HttpServer) handleListSerialNumbers(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	var err error
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:]) // default is PRV coin
	if err != nil {
		return nil, NewRPCError(ErrTokenIsInvalid, err)
	}
	if len(arrayParams) > 0 {
		tokenIDTemp, ok := arrayParams[0].(string)
		if !ok {
			Logger.log.Debugf("handleHasSerialNumbers result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("serialNumbers is invalid"))
		}
		if len(tokenIDTemp) > 0 {
			tokenID, err = (common.Hash{}).NewHashFromStr(tokenIDTemp)
			if err != nil {
				Logger.log.Debugf("handleHasSerialNumbers result: %+v, err: %+v", err)
				return nil, NewRPCError(ErrListCustomTokenNotFound, err)
			}
		}
	}
	shardID := 0
	if len(arrayParams) > 1 {
		shardID = int(arrayParams[1].(float64))
	}
	db := *(httpServer.config.Database)
	result, err := db.ListSerialNumber(*tokenID, byte(shardID))
	if err != nil {
		return nil, NewRPCError(ErrListCustomTokenNotFound, err)
	}
	return result, nil
}

// handleListSerialNumbers - return list all serialnumber in shard for token ID
func (httpServer *HttpServer) handleListSNDerivator(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	var err error
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:]) // default is PRV coin
	if err != nil {
		return nil, NewRPCError(ErrTokenIsInvalid, err)
	}
	if len(arrayParams) > 0 {
		tokenIDTemp, ok := arrayParams[0].(string)
		if !ok {
			Logger.log.Debugf("handleListSNDerivator result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("serialNumbers is invalid"))
		}
		if len(tokenIDTemp) > 0 {
			tokenID, err = (common.Hash{}).NewHashFromStr(tokenIDTemp)
			if err != nil {
				Logger.log.Debugf("handleListSNDerivator result: %+v, err: %+v", err)
				return nil, NewRPCError(ErrListCustomTokenNotFound, err)
			}
		}
	}
	db := *(httpServer.config.Database)
	resultInBytes, err := db.ListSNDerivator(*tokenID)
	result := []big.Int{}
	for _, v := range resultInBytes {
		result = append(result, *(new(big.Int).SetBytes(v)))
	}
	if err != nil {
		return nil, NewRPCError(ErrListCustomTokenNotFound, err)
	}
	return result, nil
}

// handleListCommitments - return list all commitments in shard for token ID
func (httpServer *HttpServer) handleListCommitments(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	var err error
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:]) // default is PRV coin
	if err != nil {
		return nil, NewRPCError(ErrTokenIsInvalid, err)
	}
	if len(arrayParams) > 0 {
		tokenIDTemp, ok := arrayParams[0].(string)
		if !ok {
			Logger.log.Debugf("handleHasSerialNumbers result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("serialNumbers is invalid"))
		}
		if len(tokenIDTemp) > 0 {
			tokenID, err = (common.Hash{}).NewHashFromStr(tokenIDTemp)
			if err != nil {
				Logger.log.Debugf("handleHasSerialNumbers result: %+v, err: %+v", err)
				return nil, NewRPCError(ErrListCustomTokenNotFound, err)
			}
		}
	}
	shardID := 0
	if len(arrayParams) > 1 {
		shardID = int(arrayParams[1].(float64))
	}
	db := *(httpServer.config.Database)
	result, err := db.ListCommitment(*tokenID, byte(shardID))
	if err != nil {
		return nil, NewRPCError(ErrListCustomTokenNotFound, err)
	}
	return result, nil
}

// handleListCommitmentIndices - return list all commitment indices in shard for token ID
func (httpServer *HttpServer) handleListCommitmentIndices(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	var err error
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:]) // default is PRV coin
	if err != nil {
		return nil, NewRPCError(ErrTokenIsInvalid, err)
	}
	if len(arrayParams) > 0 {
		tokenIDTemp, ok := arrayParams[0].(string)
		if !ok {
			Logger.log.Debugf("handleHasSerialNumbers result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("serialNumbers is invalid"))
		}
		if len(tokenIDTemp) > 0 {
			tokenID, err = (common.Hash{}).NewHashFromStr(tokenIDTemp)
			if err != nil {
				Logger.log.Debugf("handleHasSerialNumbers result: %+v, err: %+v", err)
				return nil, NewRPCError(ErrListCustomTokenNotFound, err)
			}
		}
	}
	shardID := 0
	if len(arrayParams) > 1 {
		shardID = int(arrayParams[1].(float64))
	}
	db := *(httpServer.config.Database)
	result, err := db.ListCommitmentIndices(*tokenID, byte(shardID))
	if err != nil {
		return nil, NewRPCError(ErrListCustomTokenNotFound, err)
	}
	return result, nil
}

// handleHasSerialNumbers - check list serial numbers existed in db of node
func (httpServer *HttpServer) handleHasSerialNumbers(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleHasSerialNumbers params: %+v", params)
	arrayParams := common.InterfaceSlice(params)

	// #1: payment address
	paymentAddressStr, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleHasSerialNumbers result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("PaymentAddress is invalid"))
	}
	key, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	lastByte := key.KeySet.PaymentAddress.Pk[len(key.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	//#2: list serialnumbers in base58check encode string
	serialNumbersStr, ok := arrayParams[1].([]interface{})
	if !ok {
		Logger.log.Debugf("handleHasSerialNumbers result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("serialNumbers is invalid"))
	}

	// #3: optional - token ID - default is prv coin
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:]) // default is PRV coin
	if err != nil {
		return nil, NewRPCError(ErrTokenIsInvalid, err)
	}
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[2].(string)
		if !ok {
			Logger.log.Debugf("handleHasSerialNumbers result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("serialNumbers is invalid"))
		}
		tokenID, err = (common.Hash{}).NewHashFromStr(tokenIDTemp)
		if err != nil {
			Logger.log.Debugf("handleHasSerialNumbers result: %+v, err: %+v", err)
			return nil, NewRPCError(ErrListCustomTokenNotFound, err)
		}
	}
	result := make([]bool, 0)
	for _, item := range serialNumbersStr {
		serialNumber, _, _ := base58.Base58Check{}.Decode(item.(string))
		db := *(httpServer.config.Database)
		ok, _ := db.HasSerialNumber(*tokenID, serialNumber, shardIDSender)
		if ok {
			// serial number in db
			result = append(result, true)
		} else {
			// serial number not in db
			result = append(result, false)
		}
	}
	Logger.log.Debugf("handleHasSerialNumbers result: %+v", result)
	return result, nil
}

// handleHasSerialNumbers - check list serial numbers existed in db of node
func (httpServer *HttpServer) handleHasSnDerivators(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleHasSnDerivators params: %+v", params)
	arrayParams := common.InterfaceSlice(params)

	// #1: payment address
	paymentAddressStr, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleHasSnDerivators result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("paymentAddress is invalid"))
	}
	key, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		Logger.log.Debugf("handleHasSnDerivators result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	lastByte := key.KeySet.PaymentAddress.Pk[len(key.KeySet.PaymentAddress.Pk)-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	_ = shardIDSender
	//#2: list serialnumbers in base58check encode string
	snDerivatorStr, ok := arrayParams[1].([]interface{})
	if !ok {
		Logger.log.Debugf("handleHasSnDerivators result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("snDerivatorStr is invalid"))
	}

	// #3: optional - token ID - default is prv coin
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:]) // default is PRV coin
	if err != nil {
		return nil, NewRPCError(ErrTokenIsInvalid, err)
	}
	if len(arrayParams) > 2 {
		tokenIDTemp, ok := arrayParams[1].(string)
		if !ok {
			Logger.log.Debugf("handleHasSnDerivators result: %+v", nil)
			return nil, NewRPCError(ErrUnexpected, errors.New("tokenID is invalid"))
		}
		tokenID, err = (common.Hash{}).NewHashFromStr(tokenIDTemp)
		if err != nil {
			Logger.log.Debugf("handleHasSnDerivators result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(ErrListCustomTokenNotFound, err)
		}
	}
	result := make([]bool, 0)
	for _, item := range snDerivatorStr {
		snderivator, _, _ := base58.Base58Check{}.Decode(item.(string))
		db := *(httpServer.config.Database)
		ok, err := db.HasSNDerivator(*tokenID, common.AddPaddingBigInt(new(big.Int).SetBytes(snderivator), common.BigIntSize))
		if ok && err == nil {
			// SnD in db
			result = append(result, true)
		} else {
			// SnD not in db
			result = append(result, false)
		}
	}
	Logger.log.Debugf("handleHasSnDerivators result: %+v", result)
	return result, nil
}

// handleCreateRawCustomTokenTransaction - handle create a custom token command and return in hex string format.
func (httpServer *HttpServer) handleCreateRawPrivacyCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCreateRawPrivacyCustomTokenTransaction params: %+v", params)
	var err error
	tx, err := httpServer.buildRawPrivacyCustomTokenTransaction(params, nil)
	if err.(*RPCError) != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	result := jsonresult.CreateTransactionTokenResult{
		ShardID:         common.GetShardIDFromLastByte(tx.Tx.PubKeyLastByteSender),
		TxID:            tx.Hash().String(),
		TokenID:         tx.TxPrivacyTokenData.PropertyID.String(),
		TokenName:       tx.TxPrivacyTokenData.PropertyName,
		TokenAmount:     tx.TxPrivacyTokenData.Amount,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	Logger.log.Debugf("handleCreateRawPrivacyCustomTokenTransaction result: %+v", result)
	return result, nil
}

// handleSendRawTransaction...
func (httpServer *HttpServer) handleSendRawPrivacyCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v", nil)
		return nil, NewRPCError(ErrSendTxData, errors.New("Param is invalid"))
	}
	base58CheckData, ok := arrayParams[0].(string)
	if !ok {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v", nil)
		return nil, NewRPCError(ErrSendTxData, errors.New("Param is invalid"))
	}
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)
	if err != nil {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	tx := transaction.TxCustomTokenPrivacy{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	hash, _, err := httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
	if err != nil {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	Logger.log.Debugf("there is hash of transaction: %s\n", hash.String())

	txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
	if err != nil {
		Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}

	txMsg.(*wire.MessageTxPrivacyToken).Transaction = &tx
	err = httpServer.config.Server.PushMessageToAll(txMsg)
	//Mark forwarded message
	if err == nil {
		httpServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	}
	result := jsonresult.CreateTransactionTokenResult{
		TxID:        tx.Hash().String(),
		TokenID:     tx.TxPrivacyTokenData.PropertyID.String(),
		TokenName:   tx.TxPrivacyTokenData.PropertyName,
		TokenAmount: tx.TxPrivacyTokenData.Amount,
		ShardID:     common.GetShardIDFromLastByte(tx.Tx.PubKeyLastByteSender),
	}
	Logger.log.Debugf("handleSendRawPrivacyCustomTokenTransaction result: %+v", result)
	return result, nil
}

// handleCreateAndSendCustomTokenTransaction - create and send a tx which process on a custom token look like erc-20 on eth
func (httpServer *HttpServer) handleCreateAndSendPrivacyCustomTokenTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCreateAndSendPrivacyCustomTokenTransaction params: %+v", params)
	data, err := httpServer.handleCreateRawPrivacyCustomTokenTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionTokenResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err == nil {
		Logger.log.Debugf("handleCreateAndSendPrivacyCustomTokenTransaction result: %+v, err: %+v", nil, err)
		return tx, nil
	}
	Logger.log.Debugf("handleCreateAndSendPrivacyCustomTokenTransaction result: %+v", txId)
	return tx, nil
}

/*
// handleCreateRawStakingTransaction handles create staking
*/
func (httpServer *HttpServer) handleCreateRawStakingTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// get component
	Logger.log.Debugf("handleCreateRawStakingTransaction params: %+v", params)
	paramsArray := common.InterfaceSlice(params)
	//var err error
	if len(paramsArray) != 7 {
		return nil, NewRPCError(ErrRPCInvalidParams, fmt.Errorf("Empty Params For Staking Transaction %+v", paramsArray))
	}
	stakingType, ok := paramsArray[4].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, fmt.Errorf("Invalid Staking Type For Staking Transaction %+v", paramsArray[4]))
	}
	candidatePaymentAddress, ok := paramsArray[5].(string)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, fmt.Errorf("Invalid Producer Payment Address for Staking Transaction %+v", paramsArray[5]))
	}
	isRewardFunder, ok := paramsArray[6].(bool)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, fmt.Errorf("Invalid Producer Payment Address for Staking Transaction %+v", paramsArray[5]))
	}
	senderKeyParam := paramsArray[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		Logger.log.Critical(err)
		Logger.log.Debugf("handleCreateRawStakingTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Cannot get payment address"))
	}
	err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Cannot import key set"))
	}
	paymentAddress, _ := senderKey.Serialize(wallet.PaymentAddressType)
	fmt.Println("SA: staking from", base58.Base58Check{}.Encode(paymentAddress, common.ZeroByte))

	stakingMetadata, err := metadata.NewStakingMetadata(int(stakingType), base58.Base58Check{}.Encode(paymentAddress, common.ZeroByte), candidatePaymentAddress, httpServer.config.ChainParams.StakingAmountShard, isRewardFunder)
	tx, err := httpServer.buildRawTransaction(params, stakingMetadata)
	if err.(*RPCError) != nil {
		Logger.log.Critical(err)
		Logger.log.Debugf("handleCreateRawStakingTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		// return hex for a new tx
		Logger.log.Debugf("handleCreateRawStakingTransaction result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, common.ZeroByte),
		ShardID:         txShardID,
	}
	Logger.log.Debugf("handleCreateRawStakingTransaction result: %+v", result)
	return result, nil
}

/*
handleCreateAndSendStakingTx - RPC creates staking transaction and send to network
*/
func (httpServer *HttpServer) handleCreateAndSendStakingTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleCreateAndSendStakingTx params: %+v", params)
	var err error
	data, err := httpServer.handleCreateRawStakingTransaction(params, closeChan)
	if err.(*RPCError) != nil {
		return nil, NewRPCError(ErrCreateTxData, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData

	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err.(*RPCError) != nil {
		Logger.log.Debugf("handleCreateAndSendStakingTx result: %+v, err: %+v", nil, err)
		return nil, NewRPCError(ErrSendTxData, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, tx.ShardID)
	Logger.log.Debugf("handleCreateAndSendStakingTx result: %+v", result)
	return result, nil
}
