package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (httpServer *HttpServer) handleGetPdexv3State(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
	beaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(httpServer.config.BlockChain.GetBeaconBestState(), uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}

	if uint64(beaconHeight) < config.Param().PDexParams.Pdexv3BreakPointHeight {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, fmt.Errorf("pDEX v3 is not available"))
	}
	pDexv3State, err := pdex.InitStateFromDB(beaconFeatureStateDB, uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}

	beaconBlocks, err := httpServer.config.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	beaconBlock := beaconBlocks[0]
	result := jsonresult.Pdexv3State{
		BeaconTimeStamp: beaconBlock.Header.Timestamp,
		Params:          pDexv3State.Reader().Params(),
	}
	return result, nil
}

/*
	Params Modifying
*/

func (httpServer *HttpServer) handleCreateAndSendTxWithPdexv3ModifyParams(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPdexv3ModifyParams(params, closeChan)
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

func (httpServer *HttpServer) handleCreateRawTxWithPdexv3ModifyParams(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	newParams, ok := tokenParamsRaw["NewParams"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NewParams is invalid"))
	}

	defaultFeeRateBPS, err := common.AssertAndConvertStrToNumber(newParams["DefaultFeeRateBPS"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("DefaultFeeRateBPS is invalid"))
	}

	feeRateBPSTemp, ok := newParams["FeeRateBPS"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("FeeRateBPS is invalid"))
	}
	feeRateBPS := map[string]uint{}
	for key, feeRatePool := range feeRateBPSTemp {
		value, err := common.AssertAndConvertStrToNumber(feeRatePool)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("FeeRateBPS is invalid"))
		}
		feeRateBPS[key] = uint(value)
	}

	prvDiscountPercent, err := common.AssertAndConvertStrToNumber(newParams["PRVDiscountPercent"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PRVDiscountPercent is invalid"))
	}

	limitProtocolFeePercent, err := common.AssertAndConvertStrToNumber(newParams["LimitProtocolFeePercent"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("LimitProtocolFeePercent is invalid"))
	}

	limitStakingPoolRewardPercent, err := common.AssertAndConvertStrToNumber(newParams["LimitStakingPoolRewardPercent"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("LimitStakingPoolRewardPercent is invalid"))
	}

	tradingProtocolFeePercent, err := common.AssertAndConvertStrToNumber(newParams["TradingProtocolFeePercent"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("TradingProtocolFeePercent is invalid"))
	}

	tradingStakingPoolRewardPercent, err := common.AssertAndConvertStrToNumber(newParams["TradingStakingPoolRewardPercent"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("TradingStakingPoolRewardPercent is invalid"))
	}

	defaultStakingPoolsShare, err := common.AssertAndConvertStrToNumber(newParams["DefaultStakingPoolsShare"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("DefaultStakingPoolsShare is invalid"))
	}

	stakingPoolsShareTemp, ok := newParams["StakingPoolsShare"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingPoolsShare is invalid"))
	}
	stakingPoolsShare := map[string]uint{}
	for key, share := range stakingPoolsShareTemp {
		value, err := common.AssertAndConvertStrToNumber(share)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingPoolsShare is invalid"))
		}
		stakingPoolsShare[key] = uint(value)
	}

	meta, err := metadataPdexv3.NewPdexv3ParamsModifyingRequest(
		metadataCommon.Pdexv3ModifyParamsMeta,
		metadataPdexv3.Pdexv3Params{
			DefaultFeeRateBPS:               uint(defaultFeeRateBPS),
			FeeRateBPS:                      feeRateBPS,
			PRVDiscountPercent:              uint(prvDiscountPercent),
			LimitProtocolFeePercent:         uint(limitProtocolFeePercent),
			LimitStakingPoolRewardPercent:   uint(limitStakingPoolRewardPercent),
			TradingProtocolFeePercent:       uint(tradingProtocolFeePercent),
			TradingStakingPoolRewardPercent: uint(tradingStakingPoolRewardPercent),
			DefaultStakingPoolsShare:        uint(defaultStakingPoolsShare),
			StakingPoolsShare:               stakingPoolsShare,
		},
	)
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

func (httpServer *HttpServer) handleGetPdexv3ParamsModifyingRequestStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
	status, err := httpServer.blockService.GetPdexv3ParamsModifyingRequestStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3ParamsModyfingStatusError, err)
	}
	return status, nil
}

/*
	Fee Management
*/
func (httpServer *HttpServer) handleGetPdexv3EstimatedLPFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	// data, ok := arrayParams[0].(map[string]interface{})
	// if !ok {
	// 	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	// }
	// pairID, ok := data["PairID"].(string)
	// if !ok {
	// 	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PairID is invalid"))
	// }
	// nfctTokenID, ok := data["NfctTokenID"].(string)
	// if !ok {
	// 	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NfctTokenID is invalid"))
	// }

	// beaconHeight := httpServer.config.BlockChain.GetBeaconBestState().BeaconHeight
	// if uint64(beaconHeight) < config.Param().PDexParams.Pdexv3BreakPointHeight {
	// 	return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, errors.New("pDEX v3 is not available"))
	// }

	// beaconFeatureStateDB := httpServer.config.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()

	// pDexv3State, err := pdex.InitStateFromDB(beaconFeatureStateDB, uint64(beaconHeight))
	// if err != nil {
	// 	return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, err)
	// }

	result := map[string]uint64{}
	// TODO

	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPdexv3WithdrawLPFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPdexv3WithdrawLPFee(params, closeChan)
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

func (httpServer *HttpServer) handleCreateRawTxWithPdexv3WithdrawLPFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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

	pairID, ok := tokenParamsRaw["PairID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PairID is invalid"))
	}

	nfctTokenID, ok := tokenParamsRaw["NfctTokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NfctTokenID is invalid"))
	}

	// payment address v2
	feeReceiver, ok := tokenParamsRaw["FeeReceiver"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("FeeReceiver is invalid"))
	}

	keyWallet, err := wallet.Base58CheckDeserialize(feeReceiver)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Cannot deserialize payment address: %v", err))
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payment address is invalid"))
	}

	nfctReceiverAddress := privacy.OTAReceiver{}
	token0ReceiverAddress := privacy.OTAReceiver{}
	token1ReceiverAddress := privacy.OTAReceiver{}
	prvReceiverAddress := privacy.OTAReceiver{}
	pdexReceiverAddress := privacy.OTAReceiver{}

	err = nfctReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = token0ReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = token1ReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = prvReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = pdexReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	nfctReceiverAddressStr, err := nfctReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	token0ReceiverAddressStr, err := token0ReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	token1ReceiverAddressStr, err := token1ReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	prvReceiverAddressStr, err := prvReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	pdexReceiverAddressStr, err := pdexReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	meta, err := metadataPdexv3.NewPdexv3WithdrawalLPFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		pairID,
		nfctTokenID,
		nfctReceiverAddressStr,
		metadataPdexv3.FeeReceiverAddress{
			Token0ReceiverAddress: token0ReceiverAddressStr,
			Token1ReceiverAddress: token1ReceiverAddressStr,
			PRVReceiverAddress:    prvReceiverAddressStr,
			PDEXReceiverAddress:   pdexReceiverAddressStr,
		},
	)
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

func (httpServer *HttpServer) handleGetPdexv3WithdrawalLPFeeStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
	status, err := httpServer.blockService.GetPdexv3WithdrawalLPFeeStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3WithdrawlLPFeeStatusError, err)
	}
	return status, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPdexv3WithdrawProtocolFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPdexv3WithdrawProtocolFee(params, closeChan)
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

func (httpServer *HttpServer) handleCreateRawTxWithPdexv3WithdrawProtocolFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	pairID, ok := tokenParamsRaw["PairID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PairID is invalid"))
	}

	// payment address v2
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().PDexParams.AdminAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Cannot deserialize paymentAddress: %v", err))
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("pDEX v3 admin payment address is invalid"))
	}

	token0ReceiverAddress := privacy.OTAReceiver{}
	token1ReceiverAddress := privacy.OTAReceiver{}
	prvReceiverAddress := privacy.OTAReceiver{}
	pdexReceiverAddress := privacy.OTAReceiver{}

	err = token0ReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = token1ReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = prvReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = pdexReceiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	token0ReceiverAddressStr, err := token0ReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	token1ReceiverAddressStr, err := token1ReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	prvReceiverAddressStr, err := prvReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	pdexReceiverAddressStr, err := pdexReceiverAddress.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	meta, err := metadataPdexv3.NewPdexv3WithdrawalProtocolFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		pairID,
		metadataPdexv3.FeeReceiverAddress{
			Token0ReceiverAddress: token0ReceiverAddressStr,
			Token1ReceiverAddress: token1ReceiverAddressStr,
			PRVReceiverAddress:    prvReceiverAddressStr,
			PDEXReceiverAddress:   pdexReceiverAddressStr,
		},
	)
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

func (httpServer *HttpServer) handleGetPdexv3WithdrawalProtocolFeeStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
	status, err := httpServer.blockService.GetPdexv3WithdrawalProtocolFeeStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3WithdrawlProtocolFeeStatusError, err)
	}
	return status, nil
}

func (httpServer *HttpServer) handleAddLiquidityV3(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, isPRV, err := httpServer.createRawTxAddLiquidityV3(params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)

	if isPRV {
		res, err = httpServer.handleSendRawTransaction(newParam, closeChan)
	} else {
		res, err = httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	}
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) createRawTxAddLiquidityV3(
	params interface{},
) (*jsonresult.CreateTransactionResult, bool, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	isPRV := false
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}

	if len(arrayParams) != 5 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}
	addLiquidityParam, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata type is invalid"))
	}
	addLiquidityRequest := PDEAddLiquidityV3Request{}
	// Convert map to json string
	addLiquidityParamData, err := json.Marshal(addLiquidityParam)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(addLiquidityParamData, &addLiquidityRequest)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	tokenAmount, err := common.AssertAndConvertNumber(addLiquidityRequest.TokenAmount)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	amplifier, err := common.AssertAndConvertNumber(addLiquidityRequest.Amplifier)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	tokenFee, err := common.AssertAndConvertNumber(addLiquidityRequest.Fee)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	tokenHash, err := common.Hash{}.NewHashFromStr(addLiquidityRequest.TokenID)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	receiverAddress := privacy.OTAReceiver{}
	refundAddress := privacy.OTAReceiver{}
	err = receiverAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = refundAddress.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	receiverAddressStr, err := receiverAddress.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	refundAddressStr, err := refundAddress.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	metaData := metadataPdexv3.NewAddLiquidityWithValue(
		addLiquidityRequest.PoolPairID,
		addLiquidityRequest.PairHash,
		receiverAddressStr, refundAddressStr,
		tokenHash.String(),
		tokenAmount,
		uint(amplifier),
	)

	if addLiquidityRequest.TokenID == common.PRVIDStr {
		isPRV = true
	}

	var byteArrays []byte
	var txHashStr string
	if isPRV {
		// create new param to build raw tx from param interface
		rawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
		if errNewParam != nil {
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
		}
		tx, rpcErr := httpServer.txService.BuildRawTransaction(rawTxParam, metaData)
		if rpcErr != nil {
			Logger.log.Error(rpcErr)
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.UnexpectedError, rpcErr)
		}
		byteArrays, err = json.Marshal(tx)
		if err != nil {
			Logger.log.Error(err)
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		txHashStr = tx.Hash().String()
	} else {
		receiverAddresses, ok := arrayParams[1].(map[string]interface{})
		if !ok {
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
		}

		customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyTokenTransaction(
			params,
			metaData,
			receiverAddresses,
			addLiquidityRequest.TokenID,
			tokenAmount,
			tokenFee,
		)
		if rpcErr != nil {
			Logger.log.Error(rpcErr)
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.UnexpectedError, rpcErr)
		}
		byteArrays, err = json.Marshal(customTokenTx)
		if err != nil {
			Logger.log.Error(err)
			return nil, isPRV, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		txHashStr = customTokenTx.Hash().String()
	}

	res := &jsonresult.CreateTransactionResult{
		TxID:            txHashStr,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return res, isPRV, nil
}

func (httpServer *HttpServer) handleGetPdexv3ContributionStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleGetPDEContributionStatusV2(params, closeChan)
}
