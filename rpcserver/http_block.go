package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
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
	beaconBestState, err := httpServer.blockService.GetBeaconBestStates()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetBeaconBestBlockError, err)
	}
	if beaconBestState == nil {
		Logger.log.Debugf("handleGetBestBlock result: %+v", result)
		return result, nil
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
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString, ok := paramsT[0].(string)
		if !ok {
			Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("hashString is invalid"))
		}
		verbosity, ok := paramsT[1].(string)
		if !ok {
			Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("verbosity is invalid"))
		}

		result, err := httpServer.blockService.RetrieveBlock(hashString, verbosity)
		if err != nil {
			return nil, err
		}
		Logger.log.Debugf("handleRetrieveBlock result: %+v", result)
		return result, nil
	}
	Logger.log.Debugf("handleRetrieveBlock result: %+v", nil)
	return nil, nil
}

/*
handleRetrieveBlock RPC return information for block
*/
func (httpServer *HttpServer) handleRetrieveBeaconBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleRetrieveBeaconBlock params: %+v", params)
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString, ok := paramsT[0].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("hashString is invalid"))
		}
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, errH)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errH)
		}
		block, _, errD := httpServer.config.BlockChain.GetBeaconBlockByHash(*hash)
		if errD != nil {
			Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, errD)
			return nil, rpcservice.NewRPCError(rpcservice.GetBeaconBlockByHashError, errD)
		}

		best := httpServer.config.BlockChain.BestState.Beacon.BestBlock
		blockHeight := block.Header.Height
		// Get next block hash unless there are none.
		var nextHashString string
		// if blockHeight < best.Header.GetHeight() {
		if blockHeight < best.Header.Height {
			nextHash, err := httpServer.config.BlockChain.GetBeaconBlockByHeight(blockHeight + 1)
			if err != nil {
				Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, err)
				return nil, rpcservice.NewRPCError(rpcservice.GetBeaconBlockByHeightError, err)
			}
			nextHashString = nextHash.Hash().String()
		}
		blockBytes, errS := json.Marshal(block)
		if errS != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errS)
		}
		result := jsonresult.NewGetBlocksBeaconResult(block, uint64(len(blockBytes)), nextHashString)
		Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", result, errD)
		return result, nil
	}
	Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, nil)
	return nil, nil
}

// handleGetBlocks - get n top blocks from chain ID
func (httpServer *HttpServer) handleGetBlocks(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlocks params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = append(arrayParams, 0.0, 0.0)
	}
	numBlockTemp, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("numblock is invalid"))
	}
	numBlock := int(numBlockTemp)
	shardIDParamTemp, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardIDParam is invalid"))
	}
	shardIDParam := int(shardIDParamTemp)
	if shardIDParam != -1 {
		result := make([]jsonresult.GetBlockResult, 0)
		shardID := byte(shardIDParam)
		clonedShardBestState, err := httpServer.config.BlockChain.BestState.GetClonedAShardBestState(shardID)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GetClonedShardBestStateError, err)
		}
		bestBlock := clonedShardBestState.BestBlock
		previousHash := bestBlock.Hash()
		for numBlock > 0 {
			numBlock--
			// block, errD := httpServer.config.BlockChain.GetBlockByHash(previousHash)
			block, size, errD := httpServer.config.BlockChain.GetShardBlockByHash(*previousHash)
			if errD != nil {
				Logger.log.Debugf("handleGetBlocks result: %+v, err: %+v", nil, errD)
				return nil, rpcservice.NewRPCError(rpcservice.GetShardBlockByHashError, errD)
			}
			blockResult := jsonresult.NewGetBlockResult(block, size, common.EmptyString)
			result = append(result, *blockResult)
			previousHash = &block.Header.PreviousBlockHash
			if previousHash.String() == (common.Hash{}).String() {
				break
			}
		}
		Logger.log.Debugf("handleGetBlocks result: %+v", result)
		return result, nil
	} else {
		result := make([]jsonresult.GetBlocksBeaconResult, 0)
		clonedBeaconBestState, err := httpServer.config.BlockChain.BestState.GetClonedBeaconBestState()
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
		}
		bestBlock := clonedBeaconBestState.BestBlock
		previousHash := bestBlock.Hash()
		for numBlock > 0 {
			numBlock--
			// block, errD := httpServer.config.BlockChain.GetBlockByHash(previousHash)
			block, size, errD := httpServer.config.BlockChain.GetBeaconBlockByHash(*previousHash)
			if errD != nil {
				return nil, rpcservice.NewRPCError(rpcservice.GetBeaconBlockByHashError, errD)
			}
			blockResult := jsonresult.NewGetBlocksBeaconResult(block, size, common.EmptyString)
			result = append(result, *blockResult)
			previousHash = &block.Header.PreviousBlockHash
			if previousHash.String() == (common.Hash{}).String() {
				break
			}
		}
		Logger.log.Debugf("handleGetBlocks result: %+v", result)
		return result, nil
	}
}

/*
getblockchaininfo RPC return information fo blockchain node
*/
func (httpServer *HttpServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlockChainInfo params: %+v", params)
	result := jsonresult.GetBlockChainInfoResult{
		ChainName:    httpServer.config.ChainParams.Name,
		BestBlocks:   make(map[int]jsonresult.GetBestBlockItem),
		ActiveShards: httpServer.config.ChainParams.ActiveShards,
	}
	shards := httpServer.config.BlockChain.BestState.GetClonedAllShardBestState()
	for shardID, bestState := range shards {
		result.BestBlocks[int(shardID)] = jsonresult.GetBestBlockItem{
			Height:           bestState.BestBlock.Header.Height,
			Hash:             bestState.BestBlockHash.String(),
			TotalTxs:         bestState.TotalTxns,
			BlockProducer:    bestState.BestBlock.Header.ProducerAddress.String(),
			BlockProducerSig: bestState.BestBlock.ProducerSig,
			Time:             bestState.BestBlock.Header.Timestamp,
		}
	}
	clonedBeaconBestState, err := httpServer.config.BlockChain.BestState.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	result.BestBlocks[-1] = jsonresult.GetBestBlockItem{
		Height:           clonedBeaconBestState.BestBlock.Header.Height,
		Hash:             clonedBeaconBestState.BestBlock.Hash().String(),
		BlockProducer:    clonedBeaconBestState.BestBlock.Header.ProducerAddress.String(),
		BlockProducerSig: clonedBeaconBestState.BestBlock.ProducerSig,
		Epoch:            clonedBeaconBestState.Epoch,
		Time:             clonedBeaconBestState.BestBlock.Header.Timestamp,
	}
	Logger.log.Debugf("handleGetBlockChainInfo result: %+v", result)
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (httpServer *HttpServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlockCount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		Logger.log.Debugf("handleGetBlockChainInfo result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("component empty"))
	}
	params, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetBlockChainInfo result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Expected get float number component"))
	}
	paramNumber := int(params.(float64))
	shardID := byte(paramNumber)
	isGetBeacon := paramNumber == -1
	if isGetBeacon {
		beacon, err := httpServer.config.BlockChain.BestState.GetClonedBeaconBestState()
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
		}
		if httpServer.config.BlockChain.BestState != nil && beacon != nil {
			result := beacon.BestBlock.Header.Height
			Logger.log.Debugf("handleGetBlockChainInfo result: %+v", result)
			return result, nil
		}
	}
	shardById, err := httpServer.config.BlockChain.BestState.GetClonedAShardBestState(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedShardBestStateError, err)
	}
	if httpServer.config.BlockChain.BestState != nil && shardById != nil &&
		shardById.BestBlock != nil {
		result := shardById.BestBlock.Header.Height + 1
		Logger.log.Debugf("handleGetBlockChainInfo result: %+v", result)
		return result, nil
	}
	result := 0
	Logger.log.Debugf("handleGetBlockChainInfo result: %+v", result)
	return 0, nil
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

	shardIDTemp, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetBlockHash result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardID is invalid"))
	}
	shardID := int(shardIDTemp)
	heightTemp, ok := arrayParams[1].(float64)
	if !ok {
		Logger.log.Debugf("handleGetBlockHash result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("height is invalid"))
	}
	height := uint64(heightTemp)

	var hash *common.Hash
	var err error
	var beaconBlock *blockchain.BeaconBlock
	var shardBlock *blockchain.ShardBlock

	isGetBeacon := shardID == -1

	if isGetBeacon {
		beaconBlock, err = httpServer.config.BlockChain.GetBeaconBlockByHeight(height)
	} else {
		shardBlock, err = httpServer.config.BlockChain.GetShardBlockByHeight(height, byte(shardID))
	}

	if err != nil {
		Logger.log.Debugf("handleGetBlockHash result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.GetShardBlockByHeightError, err)
	}

	if isGetBeacon {
		hash = beaconBlock.Hash()
	} else {
		hash = shardBlock.Hash()
	}
	result := hash.String()
	Logger.log.Debugf("handleGetBlockHash result: %+v", result)
	return result, nil
}

// handleGetBlockHeader - return block header data
func (httpServer *HttpServer) handleGetBlockHeader(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetBlockHeader params: %+v", params)
	result := jsonresult.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	log.Printf("arrayParams: %+v", arrayParams)
	if arrayParams == nil || len(arrayParams) == 0 || len(arrayParams) <= 3 {
		arrayParams = append(arrayParams, "", "", 0.0)
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
	switch getBy {
	case "blockhash":
		hash := common.Hash{}
		err := hash.Decode(&hash, block)
		// Logger.log.Info(bhash)
		log.Printf("%+v", hash)
		if err != nil {
			Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("invalid blockhash format"))
		}
		// block, err := httpServer.config.BlockChain.GetBlockByHash(&bhash)
		block, _, err := httpServer.config.BlockChain.GetShardBlockByHash(hash)
		if err != nil {
			Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.GetShardBlockByHashError, errors.New("block not exist"))
		}
		result.Header = block.Header
		result.BlockNum = int(block.Header.Height) + 1
		result.ShardID = uint8(shardID)
		result.BlockHash = hash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(block)
		if err != nil {
			Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("invalid blocknum format"))
		}
		fmt.Println(shardID)
		shard, err := httpServer.config.BlockChain.BestState.GetClonedAShardBestState(uint8(shardID))
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GetClonedShardBestStateError, err)
		}
		if uint64(bnum-1) > shard.BestBlock.Header.Height || bnum <= 0 {
			Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
			return nil, rpcservice.NewRPCError(rpcservice.GetShardBestBlockError, errors.New("Block not exist"))
		}
		block, _ := httpServer.config.BlockChain.GetShardBlockByHeight(uint64(bnum-1), uint8(shardID))

		if block != nil {
			result.Header = block.Header
			result.BlockHash = block.Hash().String()
		}
		result.BlockNum = bnum
		result.ShardID = uint8(shardID)
	default:
		Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("wrong request format"))
	}

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
	shardIDtemp, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetCrossShardBlock result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardID is invalid"))
	}
	shardID := int(shardIDtemp)
	// #param2: shard block height
	blockHeightTemp, ok := arrayParams[1].(float64)
	if !ok {
		Logger.log.Debugf("handleGetCrossShardBlock result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("blockHeight is invalid"))
	}
	blockHeight := uint64(blockHeightTemp)
	shardBlock, err := httpServer.config.BlockChain.GetShardBlockByHeight(blockHeight, byte(shardID))
	if err != nil {
		Logger.log.Debugf("handleGetCrossShardBlock result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.GetShardBlockByHeightError, err)
	}
	result := jsonresult.CrossShardDataResult{HasCrossShard: false}
	flag := false
	for _, tx := range shardBlock.Body.Transactions {
		if tx.GetType() == common.TxCustomTokenType {
			customTokenTx := tx.(*transaction.TxNormalToken)
			if customTokenTx.TxTokenData.Type == transaction.CustomTokenCrossShard {
				if !flag {
					flag = true //has cross shard block
				}
				crossShardCSTokenResult := jsonresult.CrossShardCSTokenResult{
					Name:                               customTokenTx.TxTokenData.PropertyName,
					Symbol:                             customTokenTx.TxTokenData.PropertySymbol,
					TokenID:                            customTokenTx.TxTokenData.PropertyID.String(),
					Amount:                             customTokenTx.TxTokenData.Amount,
					IsPrivacy:                          false,
					CrossShardCSTokenBalanceResultList: []jsonresult.CrossShardCSTokenBalanceResult{},
					CrossShardPrivacyCSTokenResultList: []jsonresult.CrossShardPrivacyCSTokenResult{},
				}
				crossShardCSTokenBalanceResultList := []jsonresult.CrossShardCSTokenBalanceResult{}
				for _, vout := range customTokenTx.TxTokenData.Vouts {
					paymentAddressWallet := wallet.KeyWallet{
						KeySet: incognitokey.KeySet{
							PaymentAddress: vout.PaymentAddress,
						},
					}
					paymentAddress := paymentAddressWallet.Base58CheckSerialize(wallet.PaymentAddressType)
					crossShardCSTokenBalanceResult := jsonresult.CrossShardCSTokenBalanceResult{
						PaymentAddress: paymentAddress,
						Value:          vout.Value,
					}
					crossShardCSTokenBalanceResultList = append(crossShardCSTokenBalanceResultList, crossShardCSTokenBalanceResult)
				}
				crossShardCSTokenResult.CrossShardCSTokenBalanceResultList = crossShardCSTokenBalanceResultList
				result.CrossShardCSTokenResultList = append(result.CrossShardCSTokenResultList, crossShardCSTokenResult)
			}
		}
	}
	for _, crossTransactions := range shardBlock.Body.CrossTransactions {
		if !flag {
			flag = true //has cross shard block
		}
		for _, crossTransaction := range crossTransactions {
			for _, outputCoin := range crossTransaction.OutputCoin {
				pubkey := outputCoin.CoinDetails.GetPublicKey().Compress()
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
					pubkey := outputCoin.CoinDetails.GetPublicKey().Compress()
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
