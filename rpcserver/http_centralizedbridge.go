package rpcserver

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func (httpServer *HttpServer) handleGetBridgeTokensAmounts(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	db := httpServer.config.BlockChain.GetDatabase()
	tokensAmtsBytesArr, dbErr := db.GetBridgeTokensAmounts()
	if dbErr != nil {
		return nil, NewRPCError(ErrUnexpected, dbErr)
	}

	result := &jsonresult.GetBridgeTokensAmounts{
		BridgeTokensAmounts: make(map[string]jsonresult.GetBridgeTokensAmount),
	}
	for _, tokensAmtsBytes := range tokensAmtsBytesArr {
		var tokenWithAmount lvdb.TokenWithAmount
		err := json.Unmarshal(tokensAmtsBytes, &tokenWithAmount)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, err)
		}
		tokenID := tokenWithAmount.TokenID
		result.BridgeTokensAmounts[tokenID.String()] = jsonresult.GetBridgeTokensAmount{
			TokenID: tokenWithAmount.TokenID,
			Amount:  tokenWithAmount.Amount,
		}
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[createAndSendIssuingRequest]
	return httpServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (httpServer *HttpServer) handleSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return httpServer.sendRawTxWithMetadata(params, closeChan)
}

// handleCreateAndSendIssuingRequest for user to buy PRV (using USD) or BANK token (using USD/ETH) from DCB
func (httpServer *HttpServer) handleCreateAndSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return httpServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		(*HttpServer).handleCreateIssuingRequest,
		(*HttpServer).handleSendIssuingRequest,
	)
}

func (httpServer *HttpServer) handleCreateRawTxWithContractingReq(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	if len(arrayParams) >= 5 {
		hasPrivacyToken := int(arrayParams[5].(float64)) > 0
		if hasPrivacyToken {
			return nil, NewRPCError(ErrUnexpected, errors.New("The privacy mode must be disabled"))
		}
	}

	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	paymentAddr := senderKey.KeySet.PaymentAddress
	tokenParamsRaw := arrayParams[4].(map[string]interface{})
	_, voutsAmount := transaction.CreateCustomTokenReceiverArray(tokenParamsRaw["TokenReceivers"])
	tokenID, err := common.NewHashFromStr(tokenParamsRaw["TokenID"].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	meta, _ := metadata.NewContractingRequest(
		paymentAddr,
		uint64(voutsAmount),
		*tokenID,
		metadata.ContractingRequestMeta,
	)
	customTokenTx, rpcErr := httpServer.buildRawPrivacyCustomTokenTransaction(params, meta)
	// rpcErr := err1.(*RPCError)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	byteArrays, err := json.Marshal(customTokenTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            customTokenTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := httpServer.handleCreateRawTxWithContractingReq(params, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	// sendResult, err1 := httpServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	txID := sendResult.(*common.Hash)
	result := jsonresult.CreateTransactionResult{
		// TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
		TxID: txID.String(),
	}
	return result, nil
}
