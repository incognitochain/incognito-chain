package rpcserver

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

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
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalStateError, fmt.Errorf("Can't found FeatureStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetBeaconChainDatabase()))

	// init Portal State from PortalStateDB
	portalState, err := portalprocessv4.InitCurrentPortalStateV4FromDB(beaconFeatureStateDB, nil)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalStateError, err)
	}

	// get beacon block to get timestamp
	beaconBlocks, err := httpServer.config.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPortalStateError, err)
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

/*
===== Shielding request
*/

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
		metadata.PortalV4ShieldingRequestMeta,
		tokenID,
		incognitoAddress,
		shieldingProof,
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

	// TODO: remove after testing
	// hard code the fixed ota for testing
	//otaPublicKeyStr = "1297N3bbmtDQVEX5mxVm7efobwQxT8XyL4xFrkqMN2dfhKCzHk6"
	//otaTxRandomStr = "1crTvXc9VtocvtFwD4HoD4Fs7GejVhnWnKXF6GyL1i1x5cMpPXjjvBa8ZxvRpfQ7RrLyBdk4eMHAiyoJsBfFK4YmDfkgkb9awg7"

	remoteAddress, ok := tokenParamsRaw["RemoteAddress"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("RemoteAddress is invalid"))
	}

	meta, err := metadata.NewPortalUnshieldRequest(metadata.PortalV4UnshieldingRequestMeta,
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

	meta, err := metadata.NewPortalReplacementFeeRequest(metadata.PortalV4FeeReplacementRequestMeta, portalTokenID, batchID, uint(replacementFee))
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

	meta, err := metadata.NewPortalSubmitConfirmedTxRequest(metadata.PortalV4SubmitConfirmedTxMeta, unshieldProof, portalTokenID, batchID)
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
		metadata.PortalV4ConvertVaultRequestMeta,
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
