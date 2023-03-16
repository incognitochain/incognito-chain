package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	metadataBridgeHub "github.com/incognitochain/incognito-chain/metadata/bridgehub"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (httpServer *HttpServer) handleGetBridgeHubState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Beacon height is invalid"))
	}
	result, err := httpServer.blockService.GetBridgeHubState(uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetBridgeHubStateError, err)
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendBridgeHubRegisterBridgeTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createBridgeHubRegisterBridgeTx(params)
	if err != nil {
		return nil, err
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return res, nil
}

func (httpServer *HttpServer) createBridgeHubRegisterBridgeTx(
	params interface{},
) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("expect length of param at least %v but get %v", 5, len(arrayParams)))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		ExtChainID       string   `json:"ExtChainID"`
		BridgePoolPubKey string   `json:"BridgePoolPubKey"`
		ValidatorPubKeys []string `json:"ValidatorPubKeys"`
		VaultAddress     string   `json:"VaultAddress"`
		Signature        string   `json:"Signature"`
	}{}
	// parse params & metadata
	_, err = httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}

	md, _ := metadataBridgeHub.NewRegisterBridgeRequest(
		mdReader.ExtChainID,
		mdReader.BridgePoolPubKey,
		mdReader.ValidatorPubKeys,
		mdReader.VaultAddress,
		mdReader.Signature,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, md)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

// staking request
func (httpServer *HttpServer) handleCreateAndSendBridgeHubStakeTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createBridgeHubStakeTx(params)
	if err != nil {
		return nil, err
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return res, nil
}

func (httpServer *HttpServer) createBridgeHubStakeTx(
	params interface{},
) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("expect length of param at least %v but get %v", 5, len(arrayParams)))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		ExtChainID       string      `json:"ExtChainID"`
		BridgePoolPubKey string      `json:"BridgePoolPubKey"`
		ValidatorPubKey  string      `json:"ValidatorPubKey"`
		TokenID          common.Hash `json:"TokenID"`
		BurnAmount       uint64      `json:"BurnAmount"`
	}{}
	// parse params & metadata
	_, err = httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}

	md, _ := metadataBridgeHub.NewStakePRVRequest(
		mdReader.ExtChainID,
		mdReader.BridgePoolPubKey,
		mdReader.ValidatorPubKey,
		mdReader.BurnAmount,
		mdReader.TokenID,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, md)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

// shielding request
func (httpServer *HttpServer) handleCreateAndSendBridgeShieldingTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createBridgeHubShieldingTx(params)
	if err != nil {
		return nil, err
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return res, nil
}

func (httpServer *HttpServer) createBridgeHubShieldingTx(
	params interface{},
) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("expect length of param at least %v but get %v", 5, len(arrayParams)))
	}
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		Amount     uint64      `json:"Amount"`
		Signature  []byte      `json:"Signature"`
		BtcTx      common.Hash `json:"BtcTx"`
		IncTokenID common.Hash `json:"IncTokenID"`
		Receiver   string      `json:"Receiver"`
		ExtChainID string      `json:"ExtChainID"`
	}{}
	// parse params & metadata
	_, err = httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}

	md, _ := metadataBridgeHub.NewShieldingBTCRequest(
		mdReader.Amount,
		mdReader.BtcTx,
		mdReader.IncTokenID,
		mdReader.Receiver,
		mdReader.Signature,
		mdReader.ExtChainID,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, md)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

// unshield bridge token
func processBurningBridgeHubReq(
	params interface{},
	closeChan <-chan struct{},
	httpServer *HttpServer,
) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	if len(arrayParams) > 6 {
		privacyTemp, ok := arrayParams[6].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("The privacy mode must be valid "))
		}
		hasPrivacyToken := int(privacyTemp) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be disabled"))
		}
	}

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token param is invalid"))
	}

	tokenReceivers, ok := tokenParamsRaw["TokenReceivers"].(interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token receivers is invalid"))
	}

	tokenID, ok := tokenParamsRaw["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token ID is invalid"))
	}

	remoteAddress, ok := tokenParamsRaw["RemoteAddress"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("remote address is invalid"))
	}

	extChainID, ok := tokenParamsRaw["ExtChainID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("external chain id is invalid"))
	}

	_, voutsAmount, err := rpcservice.CreateCustomTokenPrivacyBurningReceiverArray(tokenReceivers, httpServer.GetBlockchain(), httpServer.GetBlockchain().BeaconChain.CurrentHeight())
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("can not extract burn amount"))
	}
	tokenIDHash, err := common.Hash{}.NewHashFromStr(tokenID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("invalid token id"))
	}

	meta, err := metadataBridgeHub.NewBridgeHubUnshieldRequest(
		uint64(voutsAmount),
		*tokenIDHash,
		remoteAddress,
		extChainID,
	)
	if err != nil {
		Logger.log.Error(err)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	var byteArrays []byte
	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}
	byteArrays, err = json.Marshal(customTokenTx)
	txHash := customTokenTx.Hash().String()

	if err != nil {
		Logger.log.Error(err)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            txHash,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendBurningBridgeHubReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := processBurningBridgeHubReq(params, closeChan, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}
