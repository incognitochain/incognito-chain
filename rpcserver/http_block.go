package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"log"
)

// handleGetBestBlock implements the getbestblock command.
func (httpServer *HttpServer) handleGetBestBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBestBlock params: %+v", params)
	result := jsonresult.GetBestBlockResult{
		BestBlocks: make(map[int]jsonresult.GetBestBlockItem),
	}

	// for shard
	shardBestStates := httpServer.blockService.GetShardBestStates()
	for shardID, best := range shardBestStates {
		result.BestBlocks[int(shardID)] = jsonresult.GetBestBlockItem{
			Height:   best.BestBlock.Header.Height,
			Hash:     best.BestBlockHash.String(),
			TotalTxs: shardBestStates[shardID].TotalTxns,
		}
	}

	// for beacon
	if httpServer.blockService.IsBeaconBestStateNil() {
		Logger.log.Debugf("handleGetBestBlock result: %+v", result)
		return result, nil
	}

	beaconBestState, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetBeaconBestBlockError, err)
	}
	result.BestBlocks[-1] = jsonresult.GetBestBlockItem{
		Height: beaconBestState.BestBlock.Header.Height,
		Hash:   beaconBestState.BestBlockHash.String(),
	}
	Logger.log.Debugf("handleGetBestBlock result: %+v", result)
	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (httpServer *HttpServer) handleGetBestBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.GetBestBlockHashResult{
		BestBlockHashes: make(map[int]string),
	}

	// get shard
	shardBestBlockHash := httpServer.blockService.GetShardBestBlockHashes()
	for k, v := range shardBestBlockHash {
		result.BestBlockHashes[k] = v.String()
	}

	// get beacon
	beaconBestBlockHash, err := httpServer.blockService.GetBeaconBestBlockHash()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetBeaconBestBlockHashError, err)
	}
	result.BestBlockHashes[-1] = beaconBestBlockHash.String()

	return result, nil
}

/*
handleRetrieveBlock RPC return information for block
*/
func (httpServer *HttpServer) handleRetrieveBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleRetrieveBlock params: %+v", params)
	paramArray, ok := params.([]interface{})
	if ok && len(paramArray) >= 2 {
		hashString, ok := paramArray[0].(string)
		if !ok {
			Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("hashString is invalid"))
		}
		verbosity, ok := paramArray[1].(string)
		if !ok {
			Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("verbosity is invalid"))
		}

		result, err := httpServer.blockService.RetrieveShardBlock(hashString, verbosity)
		if err != nil {
			return nil, err
		}
		Logger.log.Debugf("handleRetrieveBlock result: %+v", result)
		return result, nil
	}
	Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 2 elements"))
}

func (httpServer *HttpServer) handleRetrieveBlockByHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleRetrieveBlock params: %+v", params)
	paramArray, ok := params.([]interface{})
	if ok && len(paramArray) >= 3 {
		blockHeight, ok := paramArray[0].(float64)
		if !ok {
			Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("hashString is invalid"))
		}
		shardID, ok := paramArray[1].(float64)
		if !ok {
			Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardID is invalid"))
		}
		verbosity, ok := paramArray[2].(string)
		if !ok {
			Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("verbosity is invalid"))
		}

		result, err := httpServer.blockService.RetrieveShardBlockByHeight(uint64(blockHeight), int(shardID), verbosity)
		if err != nil {
			return nil, err
		}
		Logger.log.Debugf("handleRetrieveBlock result: %+v", result)
		return result, nil
	}
	Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 2 elements"))
}

/*
handleRetrieveBlock RPC return information for block
*/
func (httpServer *HttpServer) handleRetrieveBeaconBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleRetrieveBeaconBlock params: %+v", params)
	paramArray, ok := params.([]interface{})
	if ok && len(paramArray) >= 1 {
		hashString, ok := paramArray[0].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("hashString is invalid"))
		}
		result, err := httpServer.blockService.RetrieveBeaconBlock(hashString)
		Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", result, err)
		if err != nil {
			return result, err
		}
		return result, nil
	}
	Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, nil)
	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
}

/*
handleRetrieveBlock RPC return information for block
*/
func (httpServer *HttpServer) handleRetrieveBeaconBlockByHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleRetrieveBeaconBlock params: %+v", params)
	paramArray, ok := params.([]interface{})
	if ok && len(paramArray) >= 1 {
		beaconHeight, ok := paramArray[0].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("hashString is invalid"))
		}
		result, err := httpServer.blockService.RetrieveBeaconBlockByHeigh(uint64(beaconHeight))
		Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", result, err)
		if err != nil {
			return result, err
		}
		return result, nil
	}
	Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, nil)
	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
}

// handleGetBlocks - get n top blocks from chain ID
func (httpServer *HttpServer) handleGetBlocks(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlocks params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = []interface{}{
			0.0,
			0.0,
		}
	}
	numBlockTemp, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("numblock is invalid"))
	}
	numBlock := int(numBlockTemp)

	shardIDParam, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardIDParam is invalid"))
	}
	shardID := int(shardIDParam)

	result, err := httpServer.blockService.GetBlocks(shardID, numBlock)
	if err != nil {
		return nil, err
	}
	return result, nil
}

/*
getblockchaininfo RPC return information for blockchain node
*/
func (httpServer *HttpServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlockChainInfo params: %+v", params)
	result := jsonresult.GetBlockChainInfoResult{
		ChainName:    httpServer.config.ChainParams.Name,
		BestBlocks:   make(map[int]jsonresult.GetBestBlockItem),
		ActiveShards: httpServer.config.ChainParams.ActiveShards,
	}
	shardsBestState := httpServer.blockService.GetShardBestStates()
	for shardID, bestState := range shardsBestState {
		result.BestBlocks[int(shardID)] = *(jsonresult.NewGetBestBlockItemFromShard(bestState))
	}
	beaconBestState, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	bestBlockBeaconItem := jsonresult.NewGetBestBlockItemFromBeacon(beaconBestState)
	bestBlockBeaconItem.RemainingBlockEpoch = (httpServer.config.ChainParams.Epoch * bestBlockBeaconItem.Epoch) - bestBlockBeaconItem.Height
	bestBlockBeaconItem.EpochBlock = httpServer.config.ChainParams.Epoch
	result.BestBlocks[-1] = *bestBlockBeaconItem

	Logger.log.Debugf("handleGetBlockChainInfo result: %+v", result)
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (httpServer *HttpServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlockCount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		Logger.log.Debugf("handleGetBlockChainInfo result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("component empty"))
	}

	paramNumberFloat, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetBlockChainInfo result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Expected get float number component"))
	}
	shardID := byte(int(paramNumberFloat))
	isGetBeacon := int(paramNumberFloat) == -1

	if isGetBeacon {
		beacon, err := httpServer.blockService.GetBeaconBestState()
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
		}

		result := beacon.BestBlock.Header.Height
		Logger.log.Debugf("handleGetBlockChainInfo result: %+v", result)
		return result, nil
	}

	shardById, err := httpServer.blockService.GetShardBestStateByShardID(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedShardBestStateError, err)
	}
	result := shardById.BestBlock.Header.Height + 1
	Logger.log.Debugf("handleGetBlockChainInfo result: %+v", result)
	return result, nil
}

/*
getblockhash RPC return information fo blockchain node
*/
func (httpServer *HttpServer) handleGetBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlockHash params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = []interface{}{
			0.0,
			1.0,
		}
	}

	shardIDParam, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetBlockHash result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardID is invalid"))
	}
	shardID := int(shardIDParam)

	heightParam, ok := arrayParams[1].(float64)
	if !ok {
		Logger.log.Debugf("handleGetBlockHash result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("height is invalid"))
	}
	height := uint64(heightParam)

	result, err := httpServer.blockService.GetBlockHashByHeight(shardID, height)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetShardBlockByHeightError, err)
	}
	Logger.log.Debugf("handleGetBlockHash result: %+v", result)
	return result, nil
}

// handleGetBlockHeader - return block header data
func (httpServer *HttpServer) handleGetBlockHeader(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlockHeader params: %+v", params)

	arrayParams := common.InterfaceSlice(params)
	log.Printf("arrayParams: %+v", arrayParams)
	if arrayParams == nil || len(arrayParams) < 3 {
		arrayParams = []interface{}{"", "", 0.0}
	}
	getBy, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("getBy is invalid"))
	}
	block, ok := arrayParams[1].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("block is invalid"))
	}
	shardID, ok := arrayParams[2].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardID is invalid"))
	}

	blockHeader, blockNum, blockHashStr, err := httpServer.blockService.GetBlockHeader(getBy, block, shardID)
	if err != nil {
		return nil, err
	}
	result := jsonresult.NewHeaderResult(*blockHeader, blockNum, blockHashStr, byte(shardID))
	Logger.log.Debugf("handleGetBlockHeader result: %+v", result)
	return result, nil
}

//This function return the result of cross shard block of a specific block in shard
func (httpServer *HttpServer) handleGetCrossShardBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetCrossShardBlock params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	// Logger.log.Info(arrayParams)
	log.Printf("arrayParams: %+v", arrayParams)
	if arrayParams == nil || len(arrayParams) != 2 {
		Logger.log.Debugf("handleGetCrossShardBlock result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("wrong request format"))
	}
	// #param1: shardID
	shardIDParam, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetCrossShardBlock result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardID is invalid"))
	}
	shardID := int(shardIDParam)
	// #param2: shard block height
	blockHeightParam, ok := arrayParams[1].(float64)
	if !ok {
		Logger.log.Debugf("handleGetCrossShardBlock result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("blockHeight is invalid"))
	}
	blockHeight := uint64(blockHeightParam)

	shardBlock, err := httpServer.config.BlockChain.GetShardBlockByHeight(blockHeight, byte(shardID))
	if err != nil {
		Logger.log.Debugf("handleGetCrossShardBlock result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.GetShardBlockByHeightError, err)
	}

	result := jsonresult.CrossShardDataResult{HasCrossShard: false}
	flag := false
	for _, crossTransactions := range shardBlock.Body.CrossTransactions {
		if !flag {
			flag = true //has cross shard block
		}
		for _, crossTransaction := range crossTransactions {
			for _, outputCoin := range crossTransaction.OutputCoin {
				pubkey := outputCoin.CoinDetails.GetPublicKey().ToBytesS()
				pubkeyStr := base58.Base58Check{}.Encode(pubkey, common.ZeroByte)
				if outputCoin.CoinDetailsEncrypted == nil {
					crossShardPRVResult := jsonresult.CrossShardPRVResult{
						PublicKey: pubkeyStr,
						Value:     outputCoin.CoinDetails.GetValue(),
					}
					result.CrossShardPRVResultList = append(result.CrossShardPRVResultList, crossShardPRVResult)
				} else {
					crossShardPRVPrivacyResult := jsonresult.CrossShardPRVPrivacyResult{
						PublicKey: pubkeyStr,
					}
					result.CrossShardPRVPrivacyResultList = append(result.CrossShardPRVPrivacyResultList, crossShardPRVPrivacyResult)
				}
			}
			for _, tokenPrivacyData := range crossTransaction.TokenPrivacyData {
				crossShardCSTokenResult := jsonresult.CrossShardCSTokenResult{
					Name:                               tokenPrivacyData.PropertyName,
					Symbol:                             tokenPrivacyData.PropertySymbol,
					TokenID:                            tokenPrivacyData.PropertyID.String(),
					Amount:                             tokenPrivacyData.Amount,
					IsPrivacy:                          true,
					CrossShardPrivacyCSTokenResultList: []jsonresult.CrossShardPrivacyCSTokenResult{},
				}
				for _, outputCoin := range tokenPrivacyData.OutputCoin {
					pubkey := outputCoin.CoinDetails.GetPublicKey().ToBytesS()
					pubkeyStr := base58.Base58Check{}.Encode(pubkey, common.ZeroByte)
					crossShardPrivacyCSTokenResult := jsonresult.CrossShardPrivacyCSTokenResult{
						PublicKey: pubkeyStr,
					}
					crossShardCSTokenResult.CrossShardPrivacyCSTokenResultList = append(crossShardCSTokenResult.CrossShardPrivacyCSTokenResultList, crossShardPrivacyCSTokenResult)
				}
				result.CrossShardCSTokenResultList = append(result.CrossShardCSTokenResultList, crossShardCSTokenResult)
			}
		}
	}
	if flag {
		result.HasCrossShard = flag
	}
	Logger.log.Debugf("handleGetCrossShardBlock result: %+v", result)
	return result, nil
}
