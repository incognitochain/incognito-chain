package rpcserver

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
)

type TokenInitParam struct {
	PrivateKey  string `json:"PrivateKey"`
	TokenName   string `json:"TokenName"`
	TokenSymbol string `json:"TokenSymbol"`
	Amount      uint64 `json:"Amount"`
}

func (httpServer *HttpServer) handleCreateRawTokenInitTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("no param found"))
	}

	tmpBytes, err := json.Marshal(arrayParams[0])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse %v init a tokenInit param: %v", arrayParams[0], err))
	}

	var tokenInitParam TokenInitParam
	err = json.Unmarshal(tmpBytes, &tokenInitParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse %v init a tokenInit param: %v", arrayParams[0], err))
	}

	keyWallet, err := wallet.Base58CheckDeserialize(tokenInitParam.PrivateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private key %v: %v", tokenInitParam.PrivateKey, err))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("private key length not valid: %v", keyWallet.KeySet.PrivateKey))
	}
	senderAddr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	pkBytes := keyWallet.KeySet.PaymentAddress.GetPublicSpend().ToBytesS()
	shardID := common.GetShardIDFromLastByte(pkBytes[len(pkBytes) - 1])


	otaStr, txRandomStr, err := httpServer.txService.GenerateOTAFromPaymentAddress(senderAddr)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	tokenInitMeta, _ := metadata.NewInitTokenRequest(
		otaStr,
		txRandomStr,
		tokenInitParam.Amount,
		tokenInitParam.TokenName,
		tokenInitParam.TokenSymbol,
		metadata.InitTokenRequestMeta,
	)

	rawTxParam := &bean.CreateRawTxParam{
		SenderKeySet:         &keyWallet.KeySet,
		ShardIDSender:        shardID,
		PaymentInfos:         []*privacy.PaymentInfo{},
		EstimateFeeCoinPerKb: 1,
		HasPrivacyCoin:       false,
		Info:                 []byte{},
	}

	txID, txBytes, txShardID, err := httpServer.txService.CreateRawTransaction(rawTxParam, tokenInitMeta)
	if err.(*rpcservice.RPCError) != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}

	Logger.log.Infof("creating staking transaction: txHash = %v, shardID = %v, stakingMeta = %v", txID.String(), txShardID, *tokenInitMeta)
	result := jsonresult.CreateTransactionResult{
		TxID:            txID.String(),
		Base58CheckData: base58.Base58Check{}.Encode(txBytes, common.ZeroByte),
		ShardID:         txShardID,
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTokenInitTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var err error
	data, err := httpServer.handleCreateRawTokenInitTx(params, closeChan)
	if err.(*rpcservice.RPCError) != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err.(*rpcservice.RPCError) != nil {
		return nil, rpcservice.NewRPCError(rpcservice.SendTxDataError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, tx.ShardID)
	return result, nil
}
