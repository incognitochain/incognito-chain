package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

// handleGetBestBlock implements the getbestblock command.
func (httpServer *HttpServer) handleGetBestBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBestBlock params: %+v", params)
	result := jsonresult.GetBestBlockResult{
		BestBlocks: make(map[int]jsonresult.GetBestBlockItem),
	}
	for shardID, best := range httpServer.config.BlockChain.BestState.Shard {
		result.BestBlocks[int(shardID)] = jsonresult.GetBestBlockItem{
			Height:   best.BestBlock.Header.Height,
			Hash:     best.BestBlockHash.String(),
			TotalTxs: best.TotalTxns,
		}
	}
	beaconBestState := httpServer.config.BlockChain.BestState.Beacon
	if beaconBestState == nil {
		Logger.log.Infof("handleGetBestBlock result: %+v", result)
		return result, nil
	}
	result.BestBlocks[-1] = jsonresult.GetBestBlockItem{
		Height: beaconBestState.BestBlock.Header.Height,
		Hash:   beaconBestState.BestBlockHash.String(),
	}
	Logger.log.Infof("handleGetBestBlock result: %+v", result)
	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (httpServer *HttpServer) handleGetBestBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockHashResult{
		// BestBlockHashes: make(map[byte]string),
		BestBlockHashes: make(map[int]string),
	}
	for shardID, best := range httpServer.config.BlockChain.BestState.Shard {
		result.BestBlockHashes[int(shardID)] = best.BestBlockHash.String()
	}
	beaconBestState := httpServer.config.BlockChain.BestState.Beacon
	if beaconBestState == nil {
		return result, nil
	}
	result.BestBlockHashes[-1] = beaconBestState.BestBlockHash.String()
	return result, nil
}

/*
handleRetrieveBlock RPC return information for block
*/
func (httpServer *HttpServer) handleRetrieveBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleRetrieveBlock params: %+v", params)
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString, ok := paramsT[0].(string)
		if !ok {
			Logger.log.Infof("handleRetrieveBlock result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("hashString is invalid"))
		}
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			Logger.log.Infof("handleRetrieveBlock result: %+v, err: %+v", nil, errH)
			return nil, NewRPCError(ErrUnexpected, errH)
		}
		// block, errD := httpServer.config.BlockChain.GetBlockByHash(hash)
		block, _, errD := httpServer.config.BlockChain.GetShardBlockByHash(*hash)
		if errD != nil {
			Logger.log.Infof("handleRetrieveBlock result: %+v, err: %+v", nil, errD)
			return nil, NewRPCError(ErrUnexpected, errD)
		}
		result := jsonresult.GetBlockResult{}

		verbosity, ok := paramsT[1].(string)
		if !ok {
			Logger.log.Infof("handleRetrieveBlock result: %+v", nil)
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("verbosity is invalid"))
		}

		shardID := block.Header.ShardID

		if verbosity == "0" {
			data, err := json.Marshal(block)
			if err != nil {
				Logger.log.Infof("handleRetrieveBlock result: %+v, err: %+v", nil, err)
				return nil, NewRPCError(ErrUnexpected, err)
			}
			result.Data = hex.EncodeToString(data)
		} else if verbosity == "1" {
			best := httpServer.config.BlockChain.BestState.Shard[shardID].BestBlock

			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			// if blockHeight < best.Header.GetHeight() {
			if blockHeight < best.Header.Height {
				nextHash, err := httpServer.config.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				nextHashString = nextHash.Hash().String()
			}

			result.Hash = block.Hash().String()
			result.Confirmations = int64(1 + best.Header.Height - blockHeight)
			result.Height = block.Header.Height
			result.Version = block.Header.Version
			result.TxRoot = block.Header.TxRoot.String()
			result.Time = block.Header.Timestamp
			result.ShardID = block.Header.ShardID
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.TxHashes = []string{}
			result.BlockProducerSign = block.ProducerSig
			result.BlockProducer = block.Header.ProducerAddress.String()
			result.AggregatedSig = block.AggregatedSig
			result.BeaconHeight = block.Header.BeaconHeight
			result.BeaconBlockHash = block.Header.BeaconHash.String()
			result.R = block.R
			result.Round = block.Header.Round
			result.CrossShards = []int{}
			if len(block.Header.CrossShards) > 0 {
				for _, shardID := range block.Header.CrossShards {
					result.CrossShards = append(result.CrossShards, int(shardID))
				}
			}
			result.Epoch = block.Header.Epoch

			for _, tx := range block.Body.Transactions {
				result.TxHashes = append(result.TxHashes, tx.Hash().String())
			}
		} else if verbosity == "2" {
			best := httpServer.config.BlockChain.BestState.Shard[shardID].BestBlock

			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			if blockHeight < best.Header.Height {
				nextHash, err := httpServer.config.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
				if err != nil {
					Logger.log.Infof("handleRetrieveBlock result: %+v, err: %+v", nil, err)
					return nil, NewRPCError(ErrUnexpected, err)
				}
				nextHashString = nextHash.Hash().String()
			}

			result.Hash = block.Hash().String()
			result.Confirmations = int64(1 + best.Header.Height - blockHeight)
			result.Height = block.Header.Height
			result.Version = block.Header.Version
			result.TxRoot = block.Header.TxRoot.String()
			result.Time = block.Header.Timestamp
			result.ShardID = block.Header.ShardID
			result.PreviousBlockHash = block.Header.PrevBlockHash.String()
			result.NextBlockHash = nextHashString
			result.BlockProducerSign = block.ProducerSig
			result.BlockProducer = block.Header.ProducerAddress.String()
			result.AggregatedSig = block.AggregatedSig
			result.BeaconHeight = block.Header.BeaconHeight
			result.BeaconBlockHash = block.Header.BeaconHash.String()
			result.R = block.R
			result.Round = block.Header.Round
			result.CrossShards = []int{}
			if len(block.Header.CrossShards) > 0 {
				for _, shardID := range block.Header.CrossShards {
					result.CrossShards = append(result.CrossShards, int(shardID))
				}
			}
			result.Epoch = block.Header.Epoch

			result.Txs = make([]jsonresult.GetBlockTxResult, 0)
			for _, tx := range block.Body.Transactions {
				transactionT := jsonresult.GetBlockTxResult{}

				transactionT.Hash = tx.Hash().String()

				switch tx.GetType() {
				case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
					txN := tx.(*transaction.Tx)
					data, err := json.Marshal(txN)
					if err != nil {
						return nil, NewRPCError(ErrUnexpected, err)
					}
					transactionT.HexData = hex.EncodeToString(data)
					transactionT.Locktime = txN.LockTime
				}

				result.Txs = append(result.Txs, transactionT)
			}
		}
		Logger.log.Infof("handleRetrieveBlock result: %+v", result)
		return result, nil
	}
	Logger.log.Infof("handleRetrieveBlock result: %+v", nil)
	return nil, nil
}

/*
handleRetrieveBlock RPC return information for block
*/
func (httpServer *HttpServer) handleRetrieveBeaconBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleRetrieveBeaconBlock params: %+v", params)
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString, ok := paramsT[0].(string)
		if !ok {
			return nil, NewRPCError(ErrRPCInvalidParams, errors.New("hashString is invalid"))
		}
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			Logger.log.Infof("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, errH)
			return nil, NewRPCError(ErrUnexpected, errH)
		}
		block, _, errD := httpServer.config.BlockChain.GetBeaconBlockByHash(*hash)
		if errD != nil {
			Logger.log.Infof("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, errD)
			return nil, NewRPCError(ErrUnexpected, errD)
		}

		best := httpServer.config.BlockChain.BestState.Beacon.BestBlock
		blockHeight := block.Header.Height
		// Get next block hash unless there are none.
		var nextHashString string
		// if blockHeight < best.Header.GetHeight() {
		if blockHeight < best.Header.Height {
			nextHash, err := httpServer.config.BlockChain.GetBeaconBlockByHeight(blockHeight + 1)
			if err != nil {
				Logger.log.Infof("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, err)
				return nil, NewRPCError(ErrUnexpected, err)
			}
			nextHashString = nextHash.Hash().String()
		}

		result := jsonresult.GetBlocksBeaconResult{
			Hash:              block.Hash().String(),
			Height:            block.Header.Height,
			Instructions:      block.Body.Instructions,
			Time:              block.Header.Timestamp,
			Round:             block.Header.Round,
			Epoch:             block.Header.Epoch,
			Version:           block.Header.Version,
			BlockProducerSign: block.ProducerSig,
			BlockProducer:     block.Header.ProducerAddress.String(),
			AggregatedSig:     block.AggregatedSig,
			R:                 block.R,
			PreviousBlockHash: block.Header.PrevBlockHash.String(),
			NextBlockHash:     nextHashString,
		}
		Logger.log.Infof("handleRetrieveBeaconBlock result: %+v, err: %+v", result, errD)
		return result, nil
	}
	Logger.log.Infof("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, nil)
	return nil, nil
}

// handleGetBlocks - get n top blocks from chain ID
func (httpServer *HttpServer) handleGetBlocks(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBlocks params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = append(arrayParams, 0.0, 0.0)
	}
	numBlockTemp, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("numblock is invalid"))
	}
	numBlock := int(numBlockTemp)
	shardIDParamTemp, ok := arrayParams[1].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("shardIDParam is invalid"))
	}
	shardIDParam := int(shardIDParamTemp)
	if shardIDParam != -1 {
		result := make([]jsonresult.GetBlockResult, 0)
		bestBlock := httpServer.config.BlockChain.BestState.Shard[byte(shardIDParam)].BestBlock
		previousHash := bestBlock.Hash()
		for numBlock > 0 {
			numBlock--
			// block, errD := httpServer.config.BlockChain.GetBlockByHash(previousHash)
			block, size, errD := httpServer.config.BlockChain.GetShardBlockByHash(*previousHash)
			if errD != nil {
				Logger.log.Infof("handleGetBlocks result: %+v, err: %+v", nil, errD)
				return nil, NewRPCError(ErrUnexpected, errD)
			}
			blockResult := jsonresult.GetBlockResult{}
			blockResult.Init(block, size)
			result = append(result, blockResult)
			previousHash = &block.Header.PrevBlockHash
			if previousHash.String() == (common.Hash{}).String() {
				break
			}
		}
		Logger.log.Infof("handleGetBlocks result: %+v", result)
		return result, nil
	} else {
		result := make([]jsonresult.GetBlocksBeaconResult, 0)
		bestBlock := httpServer.config.BlockChain.BestState.Beacon.BestBlock
		previousHash := bestBlock.Hash()
		for numBlock > 0 {
			numBlock--
			// block, errD := httpServer.config.BlockChain.GetBlockByHash(previousHash)
			block, size, errD := httpServer.config.BlockChain.GetBeaconBlockByHash(*previousHash)
			if errD != nil {
				return nil, NewRPCError(ErrUnexpected, errD)
			}
			blockResult := jsonresult.GetBlocksBeaconResult{}
			blockResult.Init(block, size)
			result = append(result, blockResult)
			previousHash = &block.Header.PrevBlockHash
			if previousHash.String() == (common.Hash{}).String() {
				break
			}
		}
		Logger.log.Infof("handleGetBlocks result: %+v", result)
		return result, nil
	}
}

/*
getblockchaininfo RPC return information fo blockchain node
*/
func (httpServer *HttpServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBlockChainInfo params: %+v", params)
	result := jsonresult.GetBlockChainInfoResult{
		ChainName:    httpServer.config.ChainParams.Name,
		BestBlocks:   make(map[int]jsonresult.GetBestBlockItem),
		ActiveShards: httpServer.config.ChainParams.ActiveShards,
	}
	beaconBestState := httpServer.config.BlockChain.BestState.Beacon
	for shardID, bestState := range httpServer.config.BlockChain.BestState.Shard {
		result.BestBlocks[int(shardID)] = jsonresult.GetBestBlockItem{
			Height:           bestState.BestBlock.Header.Height,
			Hash:             bestState.BestBlockHash.String(),
			TotalTxs:         bestState.TotalTxns,
			BlockProducer:    bestState.BestBlock.Header.ProducerAddress.String(),
			BlockProducerSig: bestState.BestBlock.ProducerSig,
		}
	}

	result.BestBlocks[-1] = jsonresult.GetBestBlockItem{
		Height:           beaconBestState.BestBlock.Header.Height,
		Hash:             beaconBestState.BestBlock.Hash().String(),
		BlockProducer:    beaconBestState.BestBlock.Header.ProducerAddress.String(),
		BlockProducerSig: beaconBestState.BestBlock.ProducerSig,
		Epoch:            beaconBestState.Epoch,
	}
	Logger.log.Infof("handleGetBlockChainInfo result: %+v", result)
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (httpServer *HttpServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBlockCount params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		Logger.log.Infof("handleGetBlockChainInfo result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("component empty"))
	}
	params, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetBlockChainInfo result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Expected get float number component"))
	}
	paramNumber := int(params.(float64))
	shardID := byte(paramNumber)
	isGetBeacon := paramNumber == -1
	if isGetBeacon {
		if httpServer.config.BlockChain.BestState != nil && httpServer.config.BlockChain.BestState.Beacon != nil {
			result := httpServer.config.BlockChain.BestState.Beacon.BestBlock.Header.Height
			Logger.log.Infof("handleGetBlockChainInfo result: %+v", result)
			return result, nil
		}
	}

	if httpServer.config.BlockChain.BestState != nil && httpServer.config.BlockChain.BestState.Shard[shardID] != nil && httpServer.config.BlockChain.BestState.Shard[shardID].BestBlock != nil {
		result := httpServer.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Height + 1
		Logger.log.Infof("handleGetBlockChainInfo result: %+v", result)
		return result, nil
	}
	result := 0
	Logger.log.Infof("handleGetBlockChainInfo result: %+v", result)
	return 0, nil
}

/*
getblockhash RPC return information fo blockchain node
*/
func (httpServer *HttpServer) handleGetBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBlockHash params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = []interface{}{
			0.0,
			1.0,
		}
	}

	shardIDTemp, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetBlockHash result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("shardID is invalid"))
	}
	shardID := int(shardIDTemp)
	heightTemp, ok := arrayParams[1].(float64)
	if !ok {
		Logger.log.Infof("handleGetBlockHash result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("height is invalid"))
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
		Logger.log.Infof("handleGetBlockHash result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	if isGetBeacon {
		hash = beaconBlock.Hash()
	} else {
		hash = shardBlock.Hash()
	}
	result := hash.String()
	Logger.log.Infof("handleGetBlockHash result: %+v", result)
	return result, nil
}

// handleGetBlockHeader - return block header data
func (httpServer *HttpServer) handleGetBlockHeader(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBlockHeader params: %+v", params)
	result := jsonresult.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	log.Printf("arrayParams: %+v", arrayParams)
	if arrayParams == nil || len(arrayParams) == 0 || len(arrayParams) <= 3 {
		arrayParams = append(arrayParams, "", "", 0.0)
	}
	getBy, ok := arrayParams[0].(string)
	if !ok {
		return nil, NewRPCError(ErrUnexpected, errors.New("getBy is invalid"))
	}
	block, ok := arrayParams[1].(string)
	if !ok {
		return nil, NewRPCError(ErrUnexpected, errors.New("block is invalid"))
	}
	shardID, ok := arrayParams[2].(float64)
	if !ok {
		return nil, NewRPCError(ErrUnexpected, errors.New("shardID is invalid"))
	}
	switch getBy {
	case "blockhash":
		hash := common.Hash{}
		err := hash.Decode(&hash, block)
		// Logger.log.Info(bhash)
		log.Printf("%+v", hash)
		if err != nil {
			Logger.log.Infof("handleGetBlockHeader result: %+v", nil)
			return nil, NewRPCError(ErrUnexpected, errors.New("invalid blockhash format"))
		}
		// block, err := httpServer.config.BlockChain.GetBlockByHash(&bhash)
		block, _, err := httpServer.config.BlockChain.GetShardBlockByHash(hash)
		if err != nil {
			Logger.log.Infof("handleGetBlockHeader result: %+v", nil)
			return nil, NewRPCError(ErrUnexpected, errors.New("block not exist"))
		}
		result.Header = block.Header
		// result.BlockNum = int(block.Header.GetHeight()) + 1
		result.BlockNum = int(block.Header.Height) + 1
		result.ShardID = uint8(shardID)
		result.BlockHash = hash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(block)
		if err != nil {
			Logger.log.Infof("handleGetBlockHeader result: %+v", nil)
			return nil, NewRPCError(ErrUnexpected, errors.New("invalid blocknum format"))
		}
		fmt.Println(shardID)
		// if uint64(bnum-1) > httpServer.config.BlockChain.BestState.Shard[uint8(shardID)].BestBlock.Header.GetHeight() || bnum <= 0 {
		if uint64(bnum-1) > httpServer.config.BlockChain.BestState.Shard[uint8(shardID)].BestBlock.Header.Height || bnum <= 0 {
			Logger.log.Infof("handleGetBlockHeader result: %+v", nil)
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		block, _ := httpServer.config.BlockChain.GetShardBlockByHeight(uint64(bnum-1), uint8(shardID))

		if block != nil {
			result.Header = block.Header
			result.BlockHash = block.Hash().String()
		}
		result.BlockNum = bnum
		result.ShardID = uint8(shardID)
	default:
		Logger.log.Infof("handleGetBlockHeader result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("wrong request format"))
	}

	Logger.log.Infof("handleGetBlockHeader result: %+v", result)
	return result, nil
}

//This function return the result of cross shard block of a specific block in shard
func (httpServer *HttpServer) handleGetCrossShardBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetCrossShardBlock params: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	// Logger.log.Info(arrayParams)
	log.Printf("arrayParams: %+v", arrayParams)
	if arrayParams == nil || len(arrayParams) != 2 {
		Logger.log.Infof("handleGetCrossShardBlock result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("wrong request format"))
	}
	// #param1: shardID
	shardIDtemp, ok := arrayParams[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetCrossShardBlock result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("shardID is invalid"))
	}
	shardID := int(shardIDtemp)
	// #param2: shard block height
	blockHeightTemp, ok := arrayParams[1].(float64)
	if !ok {
		Logger.log.Infof("handleGetCrossShardBlock result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("blockHeight is invalid"))
	}
	blockHeight := uint64(blockHeightTemp)
	shardBlock, err := httpServer.config.BlockChain.GetShardBlockByHeight(blockHeight, byte(shardID))
	if err != nil {
		Logger.log.Infof("handleGetCrossShardBlock result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CrossShardDataResult{HasCrossShard: false}
	flag := false
	for _, tx := range shardBlock.Body.Transactions {
		if tx.GetType() == common.TxCustomTokenType {
			customTokenTx := tx.(*transaction.TxCustomToken)
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
						KeySet: cashec.KeySet{
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
				pubkey := outputCoin.CoinDetails.PublicKey.Compress()
				pubkeyStr := base58.Base58Check{}.Encode(pubkey, common.ZeroByte)
				if outputCoin.CoinDetailsEncrypted == nil {
					crossShardConstantResult := jsonresult.CrossShardConstantResult{
						PublicKey: pubkeyStr,
						Value:     outputCoin.CoinDetails.Value,
					}
					result.CrossShardConstantResultList = append(result.CrossShardConstantResultList, crossShardConstantResult)
				} else {
					crossShardConstantPrivacyResult := jsonresult.CrossShardConstantPrivacyResult{
						PublicKey: pubkeyStr,
					}
					result.CrossShardConstantPrivacyResultList = append(result.CrossShardConstantPrivacyResultList, crossShardConstantPrivacyResult)
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
					pubkey := outputCoin.CoinDetails.PublicKey.Compress()
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
	Logger.log.Infof("handleGetCrossShardBlock result: %+v", result)
	return result, nil
}
