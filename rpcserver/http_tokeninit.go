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
	shardID := common.GetShardIDFromLastByte(pkBytes[len(pkBytes)-1])

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

	txID, txBytes, txShardID, err1 := httpServer.txService.CreateRawTransaction(rawTxParam, tokenInitMeta)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	tokenID := metadata.GenTokenIDFromRequest(txID.String(), txShardID)

	Logger.log.Infof("creating token init transaction: txHash = %v, shardID = %v, tokenID = %v\n", txID.String(), txShardID, tokenID.String())
	//although the tx has type `n`, the returned result must include the generated tokenID. Therefore, we use CreateTransactionTokenResult here
	result := jsonresult.CreateTransactionTokenResult{
		TxID:            txID.String(),
		ShardID:         txShardID,
		TokenName:       tokenInitParam.TokenName,
		TokenID:         tokenID.String(),
		TokenAmount:     tokenInitParam.Amount,
		Base58CheckData: base58.Base58Check{}.Encode(txBytes, common.ZeroByte),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTokenInitTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var err error
	data, err := httpServer.handleCreateRawTokenInitTx(params, closeChan)
	if err.(*rpcservice.RPCError) != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	tx := data.(jsonresult.CreateTransactionTokenResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	_, err = httpServer.handleSendRawTransaction(newParam, closeChan)
	if err.(*rpcservice.RPCError) != nil {
		return nil, rpcservice.NewRPCError(rpcservice.SendTxDataError, err)
	}

	return data, nil
}

//func (httpServer *HttpServer) handleGetTokenInitStatus(params interface{}, _ <-chan struct{}) (interface{}, *rpcservice.RPCError) {
//	arrayParams := common.InterfaceSlice(params)
//	data, ok := arrayParams[0].(map[string]interface{})
//	if !ok {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("params are invalid: %v", arrayParams))
//	}
//	txReq, ok := data["TxRequestID"]
//	if !ok {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("TxRequestID not found in %v", data))
//	}
//	txReqStr, ok := txReq.(string)
//	if !ok {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse %v as a string", txReq))
//	}
//	txIDHash, err := common.Hash{}.NewHashFromStr(txReqStr)
//	if err != nil {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("txHash %v is invalid: %v", txReqStr, err))
//	}
//
//	tx, err := httpServer.txService.GetTransactionByHash(txReqStr)
//	if err != nil {
//		return nil, rpcservice.NewRPCError(0, fmt.Errorf("txHash %v not found in blocks or mempool", txReqStr))
//	}
//
//	shardID := tx.ShardID
//	record := tx.Hash
//	record += strconv.FormatUint(uint64(shardID), 10)
//
//	tokenID := common.HashH([]byte(record))
//
//	if tx.IsInMempool {
//		return TokenInitStatus{Status: 0, TokenID: tokenID.String()}, nil
//	}
//
//	beaconFeatureDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
//
//	return TokenInitStatus{Status: 1, TokenID: tokenID.String()}, nil
//}
