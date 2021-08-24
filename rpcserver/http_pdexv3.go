package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
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
	poolPairs := make(map[string]*pdex.PoolPairState)
	waitingContributions := make(map[string]*rawdbv2.Pdexv3Contribution)
	err = json.Unmarshal(pDexv3State.Reader().WaitingContributions(), &waitingContributions)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	err = json.Unmarshal(pDexv3State.Reader().PoolPairs(), &poolPairs)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}

	beaconBlock := beaconBlocks[0]
	result := jsonresult.Pdexv3State{
		BeaconTimeStamp:      beaconBlock.Header.Timestamp,
		Params:               pDexv3State.Reader().Params(),
		PoolPairs:            poolPairs,
		WaitingContributions: waitingContributions,
		NftIDs:               pDexv3State.Reader().NftIDs(),
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

func (httpServer *HttpServer) handleAddLiquidityV3(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, isPRV, err := httpServer.createRawTxAddLiquidityV3(params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)

	res, err = sendCreatedTransaction(httpServer, newParam, isPRV, closeChan)
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
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("array param is not valid"))
	}
	addLiquidityRequest := Pdexv3AddLiquidityRequest{}
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
	tokenHash, err := common.Hash{}.NewHashFromStr(addLiquidityRequest.TokenID)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	if !tokenHash.IsZeroValue() {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("TokenID can not be empty"))
	}
	nftHash, err := common.Hash{}.NewHashFromStr(addLiquidityRequest.NftID)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	if !nftHash.IsZeroValue() {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NftID can not be empty"))
	}

	otaReceive := privacy.OTAReceiver{}
	otaRefund := privacy.OTAReceiver{}
	err = otaReceive.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	err = otaRefund.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiveStr, err := otaReceive.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaRefundStr, err := otaRefund.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	metaData := metadataPdexv3.NewAddLiquidityRequestWithValue(
		addLiquidityRequest.PoolPairID,
		addLiquidityRequest.PairHash,
		otaReceiveStr, otaRefundStr,
		tokenHash.String(), nftHash.String(),
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
			0,
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
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}

	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3ContributionStatusPrefix(),
		txID[:],
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get contribution data"))
	}

	return string(data), nil
}

func (httpServer *HttpServer) handleWithdrawLiquidityV3(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createRawTxWithdrawLiquidityV3(params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err = sendCreatedTransaction(httpServer, newParam, false, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) createRawTxWithdrawLiquidityV3(
	params interface{},
) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
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

	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}
	withdrawLiquidityParam, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("array param is not valid"))
	}
	withdrawLiquidityRequest := Pdexv3WithdrawLiquidityRequest{}
	// Convert map to json string
	withdrawLiquidityRequestData, err := json.Marshal(withdrawLiquidityParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(withdrawLiquidityRequestData, &withdrawLiquidityRequest)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	tokenAmount, err := common.AssertAndConvertNumber(withdrawLiquidityRequest.TokenAmount)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	token0Amount, err := common.AssertAndConvertNumber(withdrawLiquidityRequest.Token0Amount)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	token1Amount, err := common.AssertAndConvertNumber(withdrawLiquidityRequest.Token1Amount)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	otaReceiveNft := privacy.OTAReceiver{}
	err = otaReceiveNft.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiveNftStr, err := otaReceiveNft.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiveToken0 := privacy.OTAReceiver{}
	err = otaReceiveToken0.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiveToken0Str, err := otaReceiveToken0.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiveToken1 := privacy.OTAReceiver{}
	err = otaReceiveToken1.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiveToken1Str, err := otaReceiveToken1.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	beaconBestView, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	poolPairs := make(map[string]*pdex.PoolPairState)
	err = json.Unmarshal(beaconBestView.PdeState().Reader().PoolPairs(), &poolPairs)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	poolPair, found := poolPairs[withdrawLiquidityRequest.PoolPairID]
	if !found {
		err = fmt.Errorf("Can't find poolPairID %s", withdrawLiquidityRequest.PoolPairID)
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	poolPairState := poolPair.State()

	shareAmount := pdex.CalculateShareAmount(
		poolPairState.Token0RealAmount(), poolPairState.Token1RealAmount(),
		token0Amount, token1Amount, poolPairState.ShareAmount(),
	)
	metaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		withdrawLiquidityRequest.PoolPairID,
		withdrawLiquidityRequest.TokenID,
		otaReceiveNftStr, otaReceiveToken0Str, otaReceiveToken1Str,
		shareAmount,
	)
	receiverAddresses, ok := arrayParams[1].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyTokenTransaction(
		params,
		metaData,
		receiverAddresses,
		withdrawLiquidityRequest.TokenID,
		tokenAmount,
		0,
	)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, rpcErr)
	}
	byteArrays, err := json.Marshal(customTokenTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	txHashStr := customTokenTx.Hash().String()

	res := &jsonresult.CreateTransactionResult{
		TxID:            txHashStr,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return res, nil
}

func (httpServer *HttpServer) handlePdexv3MintNft(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createRawTxPdexv3MintNft(params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)

	res, err = sendCreatedTransaction(httpServer, newParam, true, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) createRawTxPdexv3MintNft(
	params interface{},
) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
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

	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 5, len(arrayParams)))
	}

	param, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	amountParam, ok := param["Amount"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("amount is invalid"))
	}
	amount, err := common.AssertAndConvertNumber(amountParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	otaReceive := privacy.OTAReceiver{}
	err = otaReceive.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiveStr, err := otaReceive.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	metaData := metadataPdexv3.NewUserMintNftRequestWithValue(otaReceiveStr, amount)

	// create new param to build raw tx from param interface
	rawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}
	tx, rpcErr := httpServer.txService.BuildRawTransaction(rawTxParam, metaData)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, rpcErr)
	}
	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	txHashStr := tx.Hash().String()

	res := &jsonresult.CreateTransactionResult{
		TxID:            txHashStr,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetPdexv3WithdrawLiquidityStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawLiquidityStatusPrefix(),
		txID[:],
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get WithdrawLiquidityStatus data"))
	}
	return string(data), nil
}

func (httpServer *HttpServer) handleGetPdexv3MintNftStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3MintNftStatusPrefix(),
		txID[:],
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get MintNftStatus data"))
	}
	return string(data), nil
}

// --- Trade - Order ---

func (httpServer *HttpServer) handlePdexv3TxTradeRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// create tx
	data, isPRV, err := createPdexv3TradeRequestTransaction(httpServer, params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, isPRV, closeChan)
}

func (httpServer *HttpServer) handlePdexv3TxAddOrderRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// create tx
	data, isPRV, err := createPdexv3AddOrderRequestTransaction(httpServer, params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, isPRV, closeChan)
}

func (httpServer *HttpServer) handlePdexv3TxWithdrawOrderRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// create tx
	data, err := createPdexv3WithdrawOrderRequestTransaction(httpServer, params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, false, closeChan)
}

func (httpServer *HttpServer) handlePdexv3GetTradeStatus(params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}

	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3TradeStatusPrefix(),
		txID[:],
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get TradeStatus data"))
	}

	return string(data), nil
}

func (httpServer *HttpServer) handlePdexv3GetAddOrderStatus(params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}

	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3AddOrderStatusPrefix(),
		txID[:],
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get AddOrderStatus data"))
	}

	return string(data), nil
}

func (httpServer *HttpServer) handlePdexv3GetWithdrawOrderStatus(params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawOrderStatusPrefix(),
		txID[:],
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get WithdrawOrderStatus data"))
	}
	return string(data), nil
}

// --- Helpers ---

func createPdexv3TradeRequestTransaction(
	httpServer *HttpServer, params interface{},
) (*jsonresult.CreateTransactionResult, bool, *rpcservice.RPCError) {
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		TradePath           []string
		TokenToSell         common.Hash
		TokenToBuy          common.Hash
		SellAmount          Uint64Reader
		MinAcceptableAmount Uint64Reader
		TradingFee          Uint64Reader
		FeeInPRV            bool
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}
	md, _ := metadataPdexv3.NewTradeRequest(
		mdReader.TradePath, mdReader.TokenToSell, uint64(mdReader.SellAmount),
		uint64(mdReader.MinAcceptableAmount), uint64(mdReader.TradingFee), nil,
		metadataCommon.Pdexv3TradeRequestMeta,
	)

	// set token ID & metadata to paramSelect struct. Generate new OTAReceivers from private key
	paramSelect.SetTokenID(md.TokenToSell)
	isPRV := md.TokenToSell == common.PRVCoinID
	tokenList := []common.Hash{md.TokenToSell, mdReader.TokenToBuy}
	if mdReader.FeeInPRV && !isPRV && mdReader.TokenToBuy != common.PRVCoinID {
		tokenList = append(tokenList, common.PRVCoinID)
	}
	md.Receiver, err = httpServer.pdexTxService.GenerateOTAReceivers(
		tokenList, paramSelect.PRV.SenderKeySet.PaymentAddress)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn selling amount for trade, plus fee
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         md.SellAmount + md.TradingFee,
		},
	}
	if isPRV {
		paramSelect.PRV.PaymentInfos = burnPayments
	} else {
		if mdReader.FeeInPRV {
			// sell amount in token
			paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{
				&privacy.PaymentInfo{
					PaymentAddress: burnAddr,
					Amount:         md.TradingFee,
				},
			}
			// trading fee in PRV
			paramSelect.SetTokenReceivers([]*privacy.PaymentInfo{
				&privacy.PaymentInfo{
					PaymentAddress: burnAddr,
					Amount:         md.SellAmount,
				},
			})
		} else {
			paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
			paramSelect.SetTokenReceivers(burnPayments)
		}
	}

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, isPRV, nil
}

func createPdexv3AddOrderRequestTransaction(
	httpServer *HttpServer, params interface{},
) (*jsonresult.CreateTransactionResult, bool, *rpcservice.RPCError) {
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		TokenToSell         common.Hash
		PoolPairID          string
		SellAmount          Uint64Reader
		MinAcceptableAmount Uint64Reader
		TradingFee          Uint64Reader
		NftID               common.Hash
		FeeInPRV            bool
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters"))
	}
	md, _ := metadataPdexv3.NewAddOrderRequest(
		mdReader.TokenToSell, mdReader.PoolPairID, uint64(mdReader.SellAmount),
		uint64(mdReader.MinAcceptableAmount), uint64(mdReader.TradingFee), nil,
		mdReader.NftID, metadataCommon.Pdexv3AddOrderRequestMeta,
	)

	// set token ID & metadata to paramSelect struct. Generate new OTAReceivers from private key
	paramSelect.SetTokenID(md.TokenToSell)
	isPRV := md.TokenToSell == common.PRVCoinID
	tokenList := []common.Hash{md.TokenToSell}
	if mdReader.FeeInPRV && !isPRV {
		tokenList = append(tokenList, common.PRVCoinID)
	}
	md.Receiver, err = httpServer.pdexTxService.GenerateOTAReceivers(
		tokenList, paramSelect.PRV.SenderKeySet.PaymentAddress)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn selling amount for order, plus fee
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         md.SellAmount + md.TradingFee,
		},
	}
	if isPRV {
		paramSelect.PRV.PaymentInfos = burnPayments
	} else {
		if mdReader.FeeInPRV {
			// sell amount in token
			paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{
				&privacy.PaymentInfo{
					PaymentAddress: burnAddr,
					Amount:         md.TradingFee,
				},
			}
			// trading fee in PRV
			paramSelect.SetTokenReceivers([]*privacy.PaymentInfo{
				&privacy.PaymentInfo{
					PaymentAddress: burnAddr,
					Amount:         md.SellAmount,
				},
			})
		} else {
			paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
			paramSelect.SetTokenReceivers(burnPayments)
		}
	}

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, isPRV, nil
}

func createPdexv3WithdrawOrderRequestTransaction(
	httpServer *HttpServer, params interface{},
) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		PoolPairID string
		OrderID    string
		TokenID    common.Hash
		Amount     Uint64Reader
		NftID      common.Hash
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters"))
	}
	md, _ := metadataPdexv3.NewWithdrawOrderRequest(
		mdReader.PoolPairID, mdReader.OrderID, mdReader.TokenID, uint64(mdReader.Amount),
		privacy.OTAReceiver{}, mdReader.NftID, metadataCommon.Pdexv3WithdrawOrderRequestMeta)

	// set token ID & metadata to paramSelect struct. Generate new OTAReceivers from private key
	if md.NftID == common.PRVCoinID {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("Cannot use PRV for withdrawOrder TX"))
	}
	paramSelect.SetTokenID(md.NftID)
	tokenList := []common.Hash{md.TokenID}
	recv, err := httpServer.pdexTxService.GenerateOTAReceivers(
		tokenList, paramSelect.PRV.SenderKeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	md.Receiver = recv[md.TokenID]
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn 1 governance-NFT to withdraw order
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         1,
		},
	}
	paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
	paramSelect.SetTokenReceivers(burnPayments)

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, nil
}

func sendCreatedTransaction(httpServer *HttpServer, params interface{}, isPRV bool, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var sendTxResult interface{}
	var err *rpcservice.RPCError
	if isPRV {
		sendTxResult, err = httpServer.handleSendRawTransaction(params, closeChan)
	} else {
		sendTxResult, err = httpServer.handleSendRawPrivacyCustomTokenTransaction(params, closeChan)
	}
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return sendTxResult, nil
}

func (httpServer *HttpServer) handlePdexv3Staking(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, isPRV, err := httpServer.createPdexv3StakingRawTx(params)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err = sendCreatedTransaction(httpServer, newParam, isPRV, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) createPdexv3StakingRawTx(
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
	stakingParam, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("array param is not valid"))
	}
	stakingRequest := Pdexv3StakingRequest{}
	// Convert map to json string
	stakingParamData, err := json.Marshal(stakingParam)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(stakingParamData, &stakingRequest)
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

	tokenAmount, err := common.AssertAndConvertNumber(stakingRequest.TokenAmount)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	tokenHash, err := common.Hash{}.NewHashFromStr(stakingRequest.TokenID)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	if !tokenHash.IsZeroValue() {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("TokenID can not be empty"))
	}
	nftHash, err := common.Hash{}.NewHashFromStr(stakingRequest.NftID)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	if !nftHash.IsZeroValue() {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NftID can not be empty"))
	}
	otaReceivers := make(map[string]string)
	otaReceiverStakingToken := privacy.OTAReceiver{}
	err = otaReceiverStakingToken.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverStakingTokenStr, err := otaReceiverStakingToken.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceivers[stakingRequest.TokenID] = otaReceiverStakingTokenStr

	metaData := metadataPdexv3.NewStakingRequestWithValue(
		tokenHash.String(),
		nftHash.String(),
		tokenAmount,
		otaReceivers,
	)

	if stakingRequest.TokenID == common.PRVIDStr {
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
			stakingRequest.TokenID,
			tokenAmount,
			0,
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

/*func (httpServer *HttpServer) handlePdexv3Unstaking(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {*/
//var res interface{}
//data, err := httpServer.createPdexv3RawTxUnstaking(params)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
//}
//base58CheckData := data.Base58CheckData
//newParam := make([]interface{}, 0)
//newParam = append(newParam, base58CheckData)

//res, err = httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
//}
//return res, nil
//}

//func (httpServer *HttpServer) createPdexv3RawTxUnstaking(
//params interface{},
//) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
//arrayParams := common.InterfaceSlice(params)
//privateKey, ok := arrayParams[0].(string)
//if !ok {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
//}
//privacyDetect, ok := arrayParams[3].(float64)
//if !ok {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
//}
//if int(privacyDetect) <= 0 {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
//}
//keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
//}
//if len(keyWallet.KeySet.PrivateKey) == 0 {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
//}

//if len(arrayParams) != 5 {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
//}
//withdrawLiquidityParam, ok := arrayParams[4].(map[string]interface{})
//if !ok {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("array param is not valid"))
//}
//withdrawLiquidityRequest := Pdexv3WithdrawLiquidityRequest{}
//// Convert map to json string
//withdrawLiquidityRequestData, err := json.Marshal(withdrawLiquidityParam)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
//}
//err = json.Unmarshal(withdrawLiquidityRequestData, &withdrawLiquidityRequest)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
//}

//tokenAmount, err := common.AssertAndConvertNumber(withdrawLiquidityRequest.TokenAmount)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
//}
//token0Amount, err := common.AssertAndConvertNumber(withdrawLiquidityRequest.Token0Amount)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
//}
//token1Amount, err := common.AssertAndConvertNumber(withdrawLiquidityRequest.Token1Amount)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
//}

//otaReceiveNft := privacy.OTAReceiver{}
//err = otaReceiveNft.FromAddress(keyWallet.KeySet.PaymentAddress)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
//}
//otaReceiveNftStr, err := otaReceiveNft.String()
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
//}
//otaReceiveToken0 := privacy.OTAReceiver{}
//err = otaReceiveToken0.FromAddress(keyWallet.KeySet.PaymentAddress)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
//}
//otaReceiveToken0Str, err := otaReceiveToken0.String()
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
//}
//otaReceiveToken1 := privacy.OTAReceiver{}
//err = otaReceiveToken1.FromAddress(keyWallet.KeySet.PaymentAddress)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
//}
//otaReceiveToken1Str, err := otaReceiveToken1.String()
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
//}
//beaconBestView, err := httpServer.blockService.GetBeaconBestState()
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
//}
//poolPairs := make(map[string]*pdex.PoolPairState)
//err = json.Unmarshal(beaconBestView.PdeState().Reader().PoolPairs(), &poolPairs)
//if err != nil {
//return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
//}
//poolPair, found := poolPairs[withdrawLiquidityRequest.PoolPairID]
//if !found {
//err = fmt.Errorf("Can't find poolPairID %s", withdrawLiquidityRequest.PoolPairID)
//return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
//}
//poolPairState := poolPair.State()

//shareAmount := pdex.CalculateShareAmount(
//poolPairState.Token0RealAmount(), poolPairState.Token1RealAmount(),
//token0Amount, token1Amount, poolPairState.ShareAmount(),
//)
//metaData := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
//withdrawLiquidityRequest.PoolPairID,
//withdrawLiquidityRequest.TokenID,
//otaReceiveNftStr, otaReceiveToken0Str, otaReceiveToken1Str,
//shareAmount,
//)
//receiverAddresses, ok := arrayParams[1].(map[string]interface{})
//if !ok {
//return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
//}
//customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyTokenTransaction(
//params,
//metaData,
//receiverAddresses,
//withdrawLiquidityRequest.TokenID,
//tokenAmount,
//0,
//)
//if rpcErr != nil {
//Logger.log.Error(rpcErr)
//return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, rpcErr)
//}
//byteArrays, err := json.Marshal(customTokenTx)
//if err != nil {
//Logger.log.Error(err)
//return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
//}
//txHashStr := customTokenTx.Hash().String()

//res := &jsonresult.CreateTransactionResult{
//TxID:            txHashStr,
//Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
//}
//return res, nil
/*}*/
