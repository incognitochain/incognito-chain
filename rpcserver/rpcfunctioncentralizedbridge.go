package rpcserver

import (
	"encoding/json"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/database/lvdb"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
)

func (rpcServer RpcServer) handleGetBridgeTokensAmounts(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	db := rpcServer.config.BlockChain.GetDatabase()
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
			return nil, NewRPCError(ErrUnexpected, dbErr)
		}
		tokenID := tokenWithAmount.TokenID
		result.BridgeTokensAmounts[tokenID.String()] = jsonresult.GetBridgeTokensAmount{
			TokenID: tokenWithAmount.TokenID,
			Amount:  tokenWithAmount.Amount,
		}
	}
	return result, nil
}

func (rpcServer RpcServer) handleCreateIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendIssuingRequest]
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
}

// handleCreateAndSendIssuingRequest for user to buy Constant (using USD) or BANK token (using USD/ETH) from DCB
func (rpcServer RpcServer) handleCreateAndSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateIssuingRequest,
		RpcServer.handleSendIssuingRequest,
	)
}

func (rpcServer RpcServer) handleCreateRawTxWithContractingReq(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

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
	customTokenTx, rpcErr := rpcServer.buildRawCustomTokenTransaction(params, meta)
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

func (rpcServer RpcServer) handleCreateAndSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := rpcServer.handleCreateRawTxWithContractingReq(params, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := rpcServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
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
