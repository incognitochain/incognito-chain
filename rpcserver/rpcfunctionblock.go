package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
)

// handleGetBestBlock implements the getbestblock command.
func (rpcServer RpcServer) handleGetBestBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockResult{
		// BestBlocks: make(map[string]jsonresult.GetBestBlockItem),
		BestBlocks: make(map[int]jsonresult.GetBestBlockItem),
	}
	for shardID, best := range rpcServer.config.BlockChain.BestState.Shard {
		// result.BestBlocks[strconv.Itoa(int(shardID))] = jsonresult.GetBestBlockItem{
		result.BestBlocks[int(shardID)] = jsonresult.GetBestBlockItem{
			Height:   best.BestBlock.Header.Height,
			Hash:     best.BestBlockHash.String(),
			TotalTxs: best.TotalTxns,
		}
	}
	beaconBestState := rpcServer.config.BlockChain.BestState.Beacon
	if beaconBestState == nil {
		return result, nil
	}
	result.BestBlocks[-1] = jsonresult.GetBestBlockItem{
		Height: beaconBestState.BestBlock.Header.Height,
		Hash:   beaconBestState.BestBlockHash.String(),
	}

	return result, nil
}

// handleGetBestBlock implements the getbestblock command.
func (rpcServer RpcServer) handleGetBestBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBestBlockHashResult{
		// BestBlockHashes: make(map[byte]string),
		BestBlockHashes: make(map[int]string),
	}
	for shardID, best := range rpcServer.config.BlockChain.BestState.Shard {
		result.BestBlockHashes[int(shardID)] = best.BestBlockHash.String()
	}
	beaconBestState := rpcServer.config.BlockChain.BestState.Beacon
	if beaconBestState == nil {
		return result, nil
	}
	result.BestBlockHashes[-1] = beaconBestState.BestBlockHash.String()
	return result, nil
}

/*
handleRetrieveBlock RPC return information for block
*/
func (rpcServer RpcServer) handleRetrieveBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString := paramsT[0].(string)
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			return nil, NewRPCError(ErrUnexpected, errH)
		}
		// block, errD := rpcServer.config.BlockChain.GetBlockByHash(hash)
		block, errD, _ := rpcServer.config.BlockChain.GetShardBlockByHash(hash)
		if errD != nil {
			return nil, NewRPCError(ErrUnexpected, errD)
		}
		result := jsonresult.GetBlockResult{}

		verbosity := paramsT[1].(string)

		shardID := block.Header.ShardID

		if verbosity == "0" {
			data, err := json.Marshal(block)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			result.Data = hex.EncodeToString(data)
		} else if verbosity == "1" {
			best := rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock

			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			// if blockHeight < best.Header.GetHeight() {
			if blockHeight < best.Header.Height {
				nextHash, err := rpcServer.config.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
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
			result.BlockProducer = block.Header.Producer
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
			best := rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock

			blockHeight := block.Header.Height
			// Get next block hash unless there are none.
			var nextHashString string
			if blockHeight < best.Header.Height {
				nextHash, err := rpcServer.config.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
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
			result.BlockProducerSign = block.ProducerSig
			result.BlockProducer = block.Header.Producer
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
				if tx.GetType() == common.TxNormalType || tx.GetType() == common.TxSalaryType {
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

		return result, nil
	}
	return nil, nil
}

/*
handleRetrieveBlock RPC return information for block
*/
func (rpcServer RpcServer) handleRetrieveBeaconBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramsT, ok := params.([]interface{})
	if ok && len(paramsT) >= 2 {
		hashString := paramsT[0].(string)
		hash, errH := common.Hash{}.NewHashFromStr(hashString)
		if errH != nil {
			return nil, NewRPCError(ErrUnexpected, errH)
		}
		block, errD, _ := rpcServer.config.BlockChain.GetBeaconBlockByHash(hash)
		if errD != nil {
			return nil, NewRPCError(ErrUnexpected, errD)
		}

		best := rpcServer.config.BlockChain.BestState.Beacon.BestBlock
		blockHeight := block.Header.Height
		// Get next block hash unless there are none.
		var nextHashString string
		// if blockHeight < best.Header.GetHeight() {
		if blockHeight < best.Header.Height {
			nextHash, err := rpcServer.config.BlockChain.GetBeaconBlockByHeight(blockHeight + 1)
			if err != nil {
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
			BlockProducer:     block.Header.Producer,
			AggregatedSig:     block.AggregatedSig,
			R:                 block.R,
			PreviousBlockHash: block.Header.PrevBlockHash.String(),
			NextBlockHash:     nextHashString,
		}

		return result, nil
	}
	return nil, nil
}

// handleGetBlocks - get n top blocks from chain ID
func (rpcServer RpcServer) handleGetBlocks(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = append(arrayParams, 0.0, 0.0)
	}
	numBlock := int(arrayParams[0].(float64))
	shardIDParam := int(arrayParams[1].(float64))
	if shardIDParam != -1 {
		result := make([]jsonresult.GetBlockResult, 0)
		bestBlock := rpcServer.config.BlockChain.BestState.Shard[byte(shardIDParam)].BestBlock
		previousHash := bestBlock.Hash()
		for numBlock > 0 {
			numBlock--
			// block, errD := rpcServer.config.BlockChain.GetBlockByHash(previousHash)
			block, errD, size := rpcServer.config.BlockChain.GetShardBlockByHash(previousHash)
			if errD != nil {
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
		return result, nil
	} else {
		result := make([]jsonresult.GetBlocksBeaconResult, 0)
		bestBlock := rpcServer.config.BlockChain.BestState.Beacon.BestBlock
		previousHash := bestBlock.Hash()
		for numBlock > 0 {
			numBlock--
			// block, errD := rpcServer.config.BlockChain.GetBlockByHash(previousHash)
			block, errD, size := rpcServer.config.BlockChain.GetBeaconBlockByHash(previousHash)
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
		return result, nil
	}
}

/*
getblockchaininfo RPC return information fo blockchain node
*/
func (rpcServer RpcServer) handleGetBlockChainInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	result := jsonresult.GetBlockChainInfoResult{
		ChainName:    rpcServer.config.ChainParams.Name,
		BestBlocks:   make(map[int]jsonresult.GetBestBlockItem),
		ActiveShards: rpcServer.config.ChainParams.ActiveShards,
	}
	beaconBestState := rpcServer.config.BlockChain.BestState.Beacon
	for shardID, bestState := range rpcServer.config.BlockChain.BestState.Shard {
		result.BestBlocks[int(shardID)] = jsonresult.GetBestBlockItem{
			Height:           bestState.BestBlock.Header.Height,
			Hash:             bestState.BestBlockHash.String(),
			SalaryPerTx:      beaconBestState.StabilityInfo.GOVConstitution.GOVParams.SalaryPerTx,
			BasicSalary:      beaconBestState.StabilityInfo.GOVConstitution.GOVParams.BasicSalary,
			TotalTxs:         bestState.TotalTxns,
			SalaryFund:       beaconBestState.StabilityInfo.SalaryFund,
			BlockProducer:    bestState.BestBlock.Header.Producer,
			BlockProducerSig: bestState.BestBlock.ProducerSig,
		}
	}

	result.BestBlocks[-1] = jsonresult.GetBestBlockItem{
		Height:           beaconBestState.BestBlock.Header.Height,
		Hash:             beaconBestState.BestBlock.Hash().String(),
		BlockProducer:    beaconBestState.BestBlock.Header.Producer,
		BlockProducerSig: beaconBestState.BestBlock.ProducerSig,
		SalaryFund:       beaconBestState.StabilityInfo.SalaryFund,
	}
	return result, nil
}

/*
getblockcount RPC return information fo blockchain node
*/
func (rpcServer RpcServer) handleGetBlockCount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("component empty"))
	}
	params, ok := arrayParams[0].(float64)
	// component, ok := component.(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Expected get float number component"))
	}
	paramNumber := int(params.(float64))
	shardID := byte(paramNumber)
	isGetBeacon := paramNumber == -1
	if isGetBeacon {
		if rpcServer.config.BlockChain.BestState != nil && rpcServer.config.BlockChain.BestState.Beacon != nil && rpcServer.config.BlockChain.BestState.Beacon.BestBlock != nil {
			return rpcServer.config.BlockChain.BestState.Beacon.BestBlock.Header.Height, nil
		}
	}

	if rpcServer.config.BlockChain.BestState != nil && rpcServer.config.BlockChain.BestState.Shard[shardID] != nil && rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock != nil {
		return rpcServer.config.BlockChain.BestState.Shard[shardID].BestBlock.Header.Height + 1, nil
	}
	return 0, nil
}

/*
getblockhash RPC return information fo blockchain node
*/
func (rpcServer RpcServer) handleGetBlockHash(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		arrayParams = []interface{}{
			0.0,
			1.0,
		}
	}

	shardID := int(arrayParams[0].(float64))
	height := uint64(arrayParams[1].(float64))

	var hash *common.Hash
	var err error
	var beaconBlock *blockchain.BeaconBlock
	var shardBlock *blockchain.ShardBlock

	isGetBeacon := shardID == -1

	if isGetBeacon {
		beaconBlock, err = rpcServer.config.BlockChain.GetBeaconBlockByHeight(height)
	} else {
		shardBlock, err = rpcServer.config.BlockChain.GetShardBlockByHeight(height, byte(shardID))
	}

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	if isGetBeacon {
		hash = beaconBlock.Hash()
	} else {
		hash = shardBlock.Hash()
	}
	// return hash.Hash().String(), nil
	return hash.String(), nil
}

// handleGetBlockHeader - return block header data
func (rpcServer RpcServer) handleGetBlockHeader(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// Logger.log.Info(component)
	log.Printf("%+v", params)
	result := jsonresult.GetHeaderResult{}

	arrayParams := common.InterfaceSlice(params)
	// Logger.log.Info(arrayParams)
	log.Printf("arrayParams: %+v", arrayParams)
	if arrayParams == nil || len(arrayParams) == 0 || len(arrayParams) <= 3 {
		arrayParams = append(arrayParams, "", "", 0.0)
	}
	getBy := arrayParams[0].(string)
	block := arrayParams[1].(string)
	shardID := arrayParams[2].(float64)
	switch getBy {
	case "blockhash":
		bhash := common.Hash{}
		err := bhash.Decode(&bhash, block)
		// Logger.log.Info(bhash)
		log.Printf("%+v", bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("invalid blockhash format"))
		}
		// block, err := rpcServer.config.BlockChain.GetBlockByHash(&bhash)
		block, err, _ := rpcServer.config.BlockChain.GetShardBlockByHash(&bhash)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("block not exist"))
		}
		result.Header = block.Header
		// result.BlockNum = int(block.Header.GetHeight()) + 1
		result.BlockNum = int(block.Header.Height) + 1
		result.ShardID = uint8(shardID)
		result.BlockHash = bhash.String()
	case "blocknum":
		bnum, err := strconv.Atoi(block)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, errors.New("invalid blocknum format"))
		}
		fmt.Println(shardID)
		// if uint64(bnum-1) > rpcServer.config.BlockChain.BestState.Shard[uint8(shardID)].BestBlock.Header.GetHeight() || bnum <= 0 {
		if uint64(bnum-1) > rpcServer.config.BlockChain.BestState.Shard[uint8(shardID)].BestBlock.Header.Height || bnum <= 0 {
			return nil, NewRPCError(ErrUnexpected, errors.New("Block not exist"))
		}
		block, _ := rpcServer.config.BlockChain.GetShardBlockByHeight(uint64(bnum-1), uint8(shardID))

		if block != nil {
			result.Header = block.Header
			result.BlockHash = block.Hash().String()
		}
		result.BlockNum = bnum
		result.ShardID = uint8(shardID)
	default:
		return nil, NewRPCError(ErrUnexpected, errors.New("wrong request format"))
	}

	return result, nil
}

//This function return the result of cross shard block of a specific block in shard
func (rpcServer RpcServer) handleGetCrossShardBlock(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// Logger.log.Info(arrayParams)
	log.Printf("arrayParams: %+v", arrayParams)
	if arrayParams == nil || len(arrayParams) != 2 {
		return nil, NewRPCError(ErrUnexpected, errors.New("wrong request format"))
	}
	// #param1: shardID
	// #param2: shard block height
	shardID := int(arrayParams[0].(float64))
	blockHeight := uint64(arrayParams[1].(float64))
	shardBlock, err := rpcServer.config.BlockChain.GetShardBlockByHeight(blockHeight, byte(shardID))
	if err != nil {
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
	return result, nil
}
