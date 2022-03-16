package rpcserver

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/btcsuite/btcd/wire"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

// PortalDepositParams consists of parameters for creating a portal shielding transaction.
// A PortalDepositParams is valid if at least one of the following conditions hold:
//	- Signature is not empty
//		- Receiver and DepositPubKey must not be empty
//	- Signature is empty
//		- If Receiver is empty, it will be generated from the sender's privateKey
//		- If DepositPrivateKey is empty, it will be derived from the DepositKeyIndex
//		- DepositPubKey is derived from DepositPrivateKey.
type PortalDepositParams struct {
	// TokenID is the shielding asset ID.
	TokenID string

	// ShieldingProof is a merkel proof for the shielding request.
	ShieldingProof string

	// DepositPrivateKey is a base58-encoded deposit privateKey used to sign the request.
	// If set empty, it will be derived from the DepositKeyIndex.
	DepositPrivateKey string

	// DepositPubKey is a base58-encoded deposit publicKey. If Signature is not provided, DepositPubKey will be derived from the DepositPrivateKey.
	DepositPubKey string

	// DepositKeyIndex is the index of the OTDepositKey. It is used to generate DepositPrivateKey and DepositPubKey when the DepositPrivateKey is not supply.
	DepositKeyIndex uint64

	// Receiver is a base58-encoded OTAReceiver. If set empty, it will be generated from the sender's privateKey.
	Receiver string

	// Signature is a valid signature signed by the owner of the shielding asset.
	// If Signature is not empty, DepositPubKey and Receiver must not be empty.
	Signature string
}

// IsValid checks if a PortalDepositParams is valid.
func (dp PortalDepositParams) IsValid() (bool, error) {
	var err error

	_, err = common.Hash{}.NewHashFromStr(dp.TokenID)
	if err != nil || dp.TokenID == "" {
		return false, fmt.Errorf("invalid tokenID %v", dp.TokenID)
	}

	if dp.Signature != "" {
		_, _, err = base58.Base58Check{}.Decode(dp.Signature)
		if err != nil {
			return false, fmt.Errorf("invalid signature")
		}
		if dp.DepositPubKey == "" || dp.Receiver == "" {
			return false, fmt.Errorf("must have both `DepositPubKey` and `Receiver`")
		}
	} else {
		if dp.DepositPrivateKey != "" {
			_, _, err = base58.Base58Check{}.Decode(dp.DepositPrivateKey)
			if err != nil {
				return false, fmt.Errorf("invalid DepositPrivateKey")
			}
		}
	}

	if dp.DepositPubKey != "" {
		_, _, err = base58.Base58Check{}.Decode(dp.DepositPubKey)
		if err != nil {
			return false, fmt.Errorf("invalid DepositPubKey")
		}
	}

	if dp.Receiver != "" {
		otaReceiver := new(privacy.OTAReceiver)
		err = otaReceiver.FromString(dp.Receiver)
		if err != nil {
			return false, fmt.Errorf("invalid receiver: %v", err)
		}
	}

	return true, nil
}

/*
===== Get Portal State
*/
func (httpServer *HttpServer) handleGetPortalV4State(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// parse params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one element"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	beaconHeight, err := common.AssertAndConvertStrToNumber(data["BeaconHeight"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	// get PortalStateDB
	beaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(httpServer.config.BlockChain.GetBeaconBestState(), uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4StateError, fmt.Errorf("Can't found FeatureStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetBeaconChainDatabase()))

	// init Portal State from PortalStateDB
	portalParamV4 := httpServer.config.BlockChain.GetPortalParamsV4(beaconHeight)
	portalState, err := portalprocessv4.InitCurrentPortalStateV4FromDB(beaconFeatureStateDB, nil, portalParamV4)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4StateError, err)
	}

	// get beacon block to get timestamp
	beaconBlocks, err := httpServer.config.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4StateError, err)
	}
	beaconBlock := beaconBlocks[0]

	type CurrentPortalState struct {
		UTXOs                     map[string]map[string]*statedb.UTXO
		ShieldingExternalTx       map[string]map[string]*statedb.ShieldingRequest
		WaitingUnshieldRequests   map[string]map[string]*statedb.WaitingUnshieldRequest
		ProcessedUnshieldRequests map[string]map[string]*statedb.ProcessedUnshieldRequestBatch
		BeaconTimeStamp           int64
	}

	result := CurrentPortalState{
		BeaconTimeStamp:           beaconBlock.Header.Timestamp,
		UTXOs:                     portalState.UTXOs,
		ShieldingExternalTx:       portalState.ShieldingExternalTx,
		WaitingUnshieldRequests:   portalState.WaitingUnshieldRequests,
		ProcessedUnshieldRequests: portalState.ProcessedUnshieldRequests,
	}
	return result, nil
}

func (httpServer *HttpServer) handleGetPortalV4Params(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// parse params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one element"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	beaconHeight, err := common.AssertAndConvertStrToNumber(data["BeaconHeight"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	portalParamV4 := httpServer.config.BlockChain.GetPortalParamsV4(beaconHeight)
	portalParamV4.PortalTokens = nil

	return portalParamV4, nil
}

/*
===== Shielding request
*/

// DEPRECATED: consider using handleCreatePortalTxDepositReqWithDepositKey instead.
func (httpServer *HttpServer) handleCreateRawTxWithShieldingReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least 5"))
	}

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata param is invalid"))
	}
	tokenID, ok := data["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenID is invalid"))
	}

	incognitoAddress, ok := data["IncogAddressStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata IncogAddressStr is invalid"))
	}

	shieldingProof, ok := data["ShieldingProof"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata ShieldingProof param is invalid"))
	}

	meta, _ := metadata.NewPortalShieldingRequest(
		metadataCommon.PortalV4ShieldingRequestMeta,
		tokenID,
		incognitoAddress,
		shieldingProof,
		"",
		nil,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}
	// HasPrivacyCoin param is always false
	createRawTxParam.HasPrivacyCoin = false

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreatePortalTxDepositReqWithDepositKey(params interface{}, _ <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("must have at lease 5 parameters"))
	}

	privateKeyStr := arrayParams[0].(string)
	w, err := wallet.Base58CheckDeserialize(privateKeyStr)
	if err != nil || len(w.KeySet.PrivateKey[:]) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("invalid privateKey: %v", privateKeyStr))
	}
	privateKey := w.KeySet.PrivateKey[:]

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("metadata param is invalid"))
	}
	jsb, _ := json.Marshal(data)

	var dp *PortalDepositParams
	err = json.Unmarshal(jsb, &dp)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot unmarshal depositParams"))
	}
	if _, err = dp.IsValid(); err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("invalid PortalDepositParams: %v", err))
	}

	receiver, depositPubKey := dp.Receiver, dp.DepositPubKey
	var sig []byte
	if dp.Signature != "" {
		sig, _, _ = base58.Base58Check{}.Decode(dp.Signature)
	} else {
		if receiver == "" {
			otaReceiver := new(privacy.OTAReceiver)
			err = otaReceiver.FromAddress(w.KeySet.PaymentAddress)
			if err != nil {
				return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("generate OTAReceiver error: %v", err))
			}
			receiver, _ = otaReceiver.String()
		}
		otaReceiver := new(coin.OTAReceiver)
		_ = otaReceiver.FromString(receiver)

		var depositPrivateKey *privacy.Scalar
		if dp.DepositPrivateKey != "" {
			tmp, _, _ := base58.Base58Check{}.Decode(dp.DepositPrivateKey)
			depositPrivateKey = new(privacy.Scalar).FromBytesS(tmp)
		} else {
			depositKey, err := incognitokey.GenerateOTDepositKeyFromPrivateKey(privateKey, dp.TokenID, dp.DepositKeyIndex)
			if err != nil {
				return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("GenerateOTDepositKeyFromPrivateKey error: %v", err))
			}
			depositPrivateKey = new(privacy.Scalar).FromBytesS(depositKey.PrivateKey)
		}
		depositPubKeyBytes := new(privacy.Point).ScalarMultBase(depositPrivateKey).ToBytesS()
		depositPubKey = base58.Base58Check{}.NewEncode(depositPubKeyBytes, 0)

		schnorrPrivateKey := new(privacy.SchnorrPrivateKey)
		r := new(privacy.Scalar).FromUint64(0) // must use r = 0
		schnorrPrivateKey.Set(depositPrivateKey, r)
		metaDataBytes, _ := otaReceiver.Bytes()
		tmpSig, err := schnorrPrivateKey.Sign(common.HashB(metaDataBytes))
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
		}
		sig = tmpSig.Bytes()
	}

	md, _ := metadata.NewPortalShieldingRequest(
		metadataCommon.PortalV4ShieldingRequestMeta,
		dp.TokenID,
		receiver,
		dp.ShieldingProof,
		depositPubKey,
		sig,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}
	// HasPrivacyCoin param is always false
	createRawTxParam.HasPrivacyCoin = false

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, md)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

// DEPRECATED: consider using RPC handleCreateAndSendPortalDepositTxWithDepositKey instead.
func (httpServer *HttpServer) handleCreateAndSendTxWithShieldingReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithShieldingReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendPortalDepositTxWithDepositKey(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreatePortalTxDepositReqWithDepositKey(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleGetPortalShieldingRequestStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	reqTxID, ok := data["ReqTxID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param ReqTxID is invalid"))
	}
	status, err := httpServer.blockService.GetPortalShieldingRequestStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4ShieldReqStatusError, err)
	}
	return status, nil
}

/*
====== Unshielding request - Burn Ptoken
*/
func (httpServer *HttpServer) handleCreateRawTxWithPortalV4UnshieldRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// parse params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) >= 7 {
		hasPrivacyTokenParam, ok := arrayParams[6].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("HasPrivacyToken is invalid"))
		}
		hasPrivacyToken := int(hasPrivacyTokenParam) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("The privacy mode must be disabled"))
		}
	}
	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	portalTokenID, ok := tokenParamsRaw["PortalTokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PortalTokenID is invalid"))
	}

	unshieldAmount, err := common.AssertAndConvertStrToNumber(tokenParamsRaw["UnshieldAmount"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	// payment address v2
	incAddressStr, ok := tokenParamsRaw["IncAddressStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("IncAddressStr is invalid"))
	}

	// generate ota corresponding the payment address
	otaPublicKeyStr, otaTxRandomStr, err := httpServer.txService.GenerateOTAFromPaymentAddress(incAddressStr)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	remoteAddress, ok := tokenParamsRaw["RemoteAddress"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("RemoteAddress is invalid"))
	}

	meta, err := metadata.NewPortalUnshieldRequest(metadataCommon.PortalV4UnshieldingRequestMeta,
		otaPublicKeyStr, otaTxRandomStr, portalTokenID, remoteAddress, unshieldAmount)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	byteArrays, err2 := json.Marshal(customTokenTx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            customTokenTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPortalV4UnshieldRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// create raw transaction
	data, err := httpServer.handleCreateRawTxWithPortalV4UnshieldRequest(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)

	// send raw transaction
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleGetPortalUnshieldingRequestStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// parse params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one element"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	unshieldID, ok := data["UnshieldID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param UnshieldID is invalid"))
	}

	// get status of unshielding request
	status, err := httpServer.blockService.GetPortalUnshieldingRequestStatus(unshieldID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4UnshieldReqStatusError, err)
	}
	return status, nil
}

func (httpServer *HttpServer) handleGetPortalBatchUnshieldingRequestStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// parse params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	batchID, ok := data["BatchID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param BatchID is invalid"))
	}
	// get status of batch unshield process
	status, err := httpServer.blockService.GetPortalBatchUnshieldingRequestStatus(batchID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4BatchUnshieldReqStatusError, err)
	}
	return status, nil
}

/*
====== Get raw signed tx
*/
func (httpServer *HttpServer) handleGetPortalSignedExtTxWithBatchID(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// parse params
	listParams, ok := params.([]interface{})
	if !ok || len(listParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param array must be an array at least one element"))
	}
	data, ok := listParams[0].(map[string]interface{})
	if !ok || len(listParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an map[string]interface{}"))
	}
	batchIDParam, ok := data["BatchID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param TxID should be a string"))
	}

	// get beacon block height from txID
	unshieldBatch, err := httpServer.blockService.GetPortalBatchUnshieldingRequestStatus(batchIDParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	// get signed transaction
	return getRawSignedTxByHeight(httpServer, unshieldBatch.BeaconHeight, unshieldBatch.RawExternalTx, unshieldBatch.UTXOs)
}

type getSignedTxResult struct {
	SignedTx     string
	BeaconHeight uint64
	TxID         string
}

func getRawSignedTxByHeight(
	httpServer *HttpServer,
	height uint64,
	rawTx string,
	utxos []*statedb.UTXO,
) (interface{}, *rpcservice.RPCError) {
	// get portal params v4
	portalParamsv4 := httpServer.config.BlockChain.GetPortalParamsV4(height)

	// get beacon block
	beaconBlockQueried, err := getSingleBeaconBlockByHeight(httpServer.GetBlockchain(), height)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	block := &beaconBlock{BeaconBlock: beaconBlockQueried}

	// get portal v4 sigs from beacon block
	portalV4Sig, err := block.PortalV4Sigs(httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// parse rawTx to get externalTxHash
	hexRawTx, err := hex.DecodeString(rawTx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	buffer := bytes.NewReader(hexRawTx)
	externalTx := wire.NewMsgTx(wire.TxVersion)
	err = externalTx.Deserialize(buffer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	externalTxHash := externalTx.TxHash().String()

	// get poral v4 sigs of externalTxHash
	var tokenID string
	numSig := uint(0)
	sigs := make([][][]byte, len(externalTx.TxIn))
	for i := range sigs {
		sigs[i] = make([][]byte, 1)
	}

	for _, v := range portalV4Sig {
		if v.RawTxHash != externalTxHash {
			continue
		}

		if tokenID == "" {
			tokenID = v.TokenID
		}
		for i, v2 := range v.Sigs {
			if i >= len(sigs) {
				return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Invalid length of portal sigs"))
			}
			sigs[i] = append(sigs[i], v2)
		}
		numSig++
		if numSig == portalParamsv4.NumRequiredSigs {
			break
		}
	}
	if tokenID == "" {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Not found portal sigs for batchID"))
	}

	// append multisig script hex into each sigs
	// then, attach sig into TxIn in externalTx
	portalParams := httpServer.blockService.BlockChain.GetPortalParamsV4(height)
	for i, v := range sigs {
		multisigScriptHex, _, err := portalParams.PortalTokens[tokenID].GenerateOTMultisigAddress(portalParams.MasterPubKeys[tokenID], int(portalParams.NumRequiredSigs), utxos[i].GetChainCodeSeed())
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		v = append(v, multisigScriptHex)
		externalTx.TxIn[i].Witness = v
	}

	// hex-encoding signed external tx
	var signedTx bytes.Buffer
	err = externalTx.Serialize(&signedTx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	hexSignedTx := hex.EncodeToString(signedTx.Bytes())
	return getSignedTxResult{
		SignedTx:     hexSignedTx,
		BeaconHeight: height,
		TxID:         externalTx.TxHash().String(),
	}, nil
}

/*
 ====== Replace-By-Fee for Unshield Requests
*/
func (httpServer *HttpServer) handleCreateRawTxWithPortalReplaceUnshieldFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	portalTokenID, ok := tokenParamsRaw["PortalTokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PortalTokenID is invalid"))
	}

	replacementFee, err := common.AssertAndConvertStrToNumber(tokenParamsRaw["ReplacementFee"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	batchID, ok := tokenParamsRaw["BatchID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("BatchID is invalid"))
	}

	meta, err := metadata.NewPortalReplacementFeeRequest(metadataCommon.PortalV4FeeReplacementRequestMeta, portalTokenID, batchID, uint(replacementFee))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
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

func (httpServer *HttpServer) handleCreateAndSendTxWithPortalReplaceUnshieldFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPortalReplaceUnshieldFee(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleGetPortalReplacementFeeRequestStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	reqTxID, ok := data["ReqTxID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param ReqTxID is invalid"))
	}
	status, err := httpServer.blockService.GetPortalReqReplacementFeeStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4FeeReplacementReqStatusError, err)
	}
	return status, nil
}

/*
 ====== Submit confirmed external tx
*/
func (httpServer *HttpServer) handleCreateRawTxWithPortalSubmitConfirmedTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	portalTokenID, ok := tokenParamsRaw["PortalTokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PortalTokenID is invalid"))
	}

	unshieldProof, ok := tokenParamsRaw["UnshieldProof"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("UnshieldProof is invalid"))
	}

	batchID, ok := tokenParamsRaw["BatchID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("BatchID is invalid"))
	}

	meta, err := metadata.NewPortalSubmitConfirmedTxRequest(metadataCommon.PortalV4SubmitConfirmedTxMeta, unshieldProof, portalTokenID, batchID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPortalSubmitConfirmedTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPortalSubmitConfirmedTx(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleGetPortalPortalSubmitConfirmedTxStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	reqTxID, ok := data["ReqTxID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param ReqTxID is invalid"))
	}
	status, err := httpServer.blockService.GetPortalSubmitConfirmedTxStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4SubmitConfirmedTxStatusError, err)
	}
	return status, nil
}

func (httpServer *HttpServer) handleGetPortalTransactionSignedWithFeeReplacementTx(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	listParams, ok := params.([]interface{})
	if !ok || len(listParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data, ok := listParams[0].(map[string]interface{})
	if !ok || len(listParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an map[string]interface{}"))
	}
	txIDParam, ok := data["TxID"].(string)
	if !ok || txIDParam == "" {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param TxID should be a string and not empty"))
	}

	replaceFeeStatus, err := httpServer.blockService.GetPortalReqReplacementFeeStatus(txIDParam)
	if err != nil || replaceFeeStatus == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Get portal btc signed tx error: %v", err))
	}

	unshieldBatch, err := httpServer.blockService.GetPortalBatchUnshieldingRequestStatus(replaceFeeStatus.BatchID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Get portal proof error: %v", err))
	}

	// get signed transaction
	return getRawSignedTxByHeight(httpServer, replaceFeeStatus.BeaconHeight, replaceFeeStatus.ExternalRawTx, unshieldBatch.UTXOs)
}

/*
===== Converting vault request
*/

func (httpServer *HttpServer) handleCreateRawTxWithPortalConvertVault(
	params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least 5"))
	}

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata param is invalid"))
	}
	tokenID, ok := data["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenID is invalid"))
	}

	convertingProof, ok := data["ConvertingProof"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata ConvertingProof param is invalid"))
	}

	meta, _ := metadata.NewPortalConvertVaultRequest(
		metadataCommon.PortalV4ConvertVaultRequestMeta,
		tokenID,
		convertingProof,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}
	// HasPrivacyCoin param is always false
	createRawTxParam.HasPrivacyCoin = false

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPortalConvertVault(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPortalConvertVault(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

func (httpServer *HttpServer) handleGetPortalConvertVaultTxStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	reqTxID, ok := data["ReqTxID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param ReqTxID is invalid"))
	}
	status, err := httpServer.blockService.GetPortalConvertVaultTxStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalV4ConvertVaultTxStatusError, err)
	}
	return status, nil
}

// handleGenerateShieldingMultisigAddress returns the multi-sig shielding address for a given payment address and tokenID.
// DEPRECATED: use handleGenerateDepositAddress.
func (httpServer *HttpServer) handleGenerateShieldingMultisigAddress(
	params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// parse params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one element"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	incAddressStr, ok := data["IncAddressStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("IncAddressStr is invalid"))
	}
	// validate incAddressStr must be a version 2 one.
	if _, err := metadata.AssertPaymentAddressAndTxVersion(incAddressStr, 2); err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("IncAddressStr must be a valid payment address version 2"))
	}
	tokenID, ok := data["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("TokenID is invalid"))
	}

	// get portal params with the latest beacon height
	latestBeaconHeight := httpServer.config.BlockChain.GetBeaconBestState().BeaconHeight
	portalParamV4 := httpServer.config.BlockChain.GetPortalParamsV4(latestBeaconHeight)

	// check is Portal token
	if !portalParamV4.IsPortalToken(tokenID) {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("TokenID is not a portal token"))
	}

	// generate shielding multisig address
	_, shieldingAddress, err := portalParamV4.PortalTokens[tokenID].GenerateOTMultisigAddress(
		portalParamV4.MasterPubKeys[tokenID], int(portalParamV4.NumRequiredSigs), incAddressStr)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("Error when generating multisig address %v\n", err))
	}

	return shieldingAddress, nil
}

func (httpServer *HttpServer) handleGenerateDepositAddress(
	params interface{}, _ <-chan struct{},
) (interface{}, *rpcservice.RPCError) {

	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("param array must be at least one element"))
	}
	paramList, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("payload data is invalid"))
	}

	type TmpParam struct {
		IncAddressStr string `json:"IncAddressStr,omitempty"`
		DepositPubKey string `json:"DepositPubKey,omitempty"`
		TokenID       string
	}
	var tmpParam *TmpParam
	jsb, _ := json.Marshal(paramList)
	err := json.Unmarshal(jsb, &tmpParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("unmarshal parameters error: %v", err))
	}
	fmt.Println("BUGLOG12", tmpParam)

	var chainCodeSeed string
	if tmpParam.IncAddressStr != "" {
		// validate incAddressStr must be a version 2 one.
		if _, err := metadata.AssertPaymentAddressAndTxVersion(tmpParam.IncAddressStr, 2); err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
				fmt.Errorf("IncAddressStr must be a valid payment address version 2"))
		}
		chainCodeSeed = tmpParam.IncAddressStr
	} else {
		chainCodeSeed = tmpParam.DepositPubKey
	}
	if chainCodeSeed == "" {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("no valid chainCodeSeed found"))
	}

	// get portal params with the latest beacon height
	latestBeaconHeight := httpServer.config.BlockChain.GetBeaconBestState().BeaconHeight
	portalParamV4 := httpServer.config.BlockChain.GetPortalParamsV4(latestBeaconHeight)
	if !portalParamV4.IsPortalToken(tmpParam.TokenID) {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("TokenID is not a portal token"))
	}

	_, shieldingAddress, err := portalParamV4.PortalTokens[tmpParam.TokenID].GenerateOTMultisigAddress(
		portalParamV4.MasterPubKeys[tmpParam.TokenID], int(portalParamV4.NumRequiredSigs), chainCodeSeed)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("error when generating shielding address: %v\n", err))
	}

	return shieldingAddress, nil
}
