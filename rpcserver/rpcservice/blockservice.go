package rpcservice

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
)

type BlockService struct {
	BlockChain *blockchain.BlockChain
	DB         *database.DatabaseInterface
	TxMemPool  *mempool.TxPool
	MemCache   *memcache.MemoryCache
}

func (blockService BlockService) GetShardBestStates() map[byte]*blockchain.ShardBestState {
	shards := make(map[byte]*blockchain.ShardBestState)
	cacheKey := memcache.GetShardBestStateCachedKey()
	cacheValue, err := blockService.MemCache.Get(cacheKey)
	if err == nil && len(cacheValue) > 0 {
		err1 := json.Unmarshal(cacheValue, &shards)
		if err1 != nil {
			Logger.log.Error("Json Unmarshal cache of shard best state error", err1)
		}
	}
	if len(shards) == 0 {
		shards = blockService.BlockChain.BestState.GetClonedAllShardBestState()
		cacheValue, err := json.Marshal(shards)
		if err == nil {
			err1 := blockService.MemCache.PutExpired(cacheKey, cacheValue, 10000)
			if err1 != nil {
				Logger.log.Error("Cache data of shard best state error", err1)
			}
		}
	}
	return shards
}

func (blockService BlockService) GetShardBestStateByShardID(shardID byte) (*blockchain.ShardBestState, error) {
	if blockService.IsShardBestStateNil() {
		return nil, errors.New("Best State shard not existed")
	}
	shard, err := blockService.BlockChain.BestState.GetClonedAShardBestState(shardID)
	return shard, err
}

func (blockService BlockService) GetShardBestBlocks() map[byte]blockchain.ShardBlock {
	bestBlocks := make(map[byte]blockchain.ShardBlock)
	shards := blockService.BlockChain.BestState.GetClonedAllShardBestState()
	for shardID, best := range shards {
		bestBlocks[shardID] = *best.BestBlock
	}
	return bestBlocks
}

func (blockService BlockService) GetShardBestBlockByShardID(shardID byte) (blockchain.ShardBlock, common.Hash, error) {
	shard, err := blockService.BlockChain.BestState.GetClonedAShardBestState(shardID)
	return *shard.BestBlock, shard.BestBlockHash, err
}

func (blockService BlockService) GetShardBestBlockHashes() map[int]common.Hash {
	bestBlockHashes := make(map[int]common.Hash)
	shards := blockService.BlockChain.BestState.GetClonedAllShardBestState()
	for shardID, best := range shards {
		bestBlockHashes[int(shardID)] = best.BestBlockHash
	}
	return bestBlockHashes
}

func (blockService BlockService) GetShardBestBlockHashByShardID(shardID byte) common.Hash {
	shards := blockService.BlockChain.BestState.GetClonedAllShardBestState()
	return shards[shardID].BestBlockHash
}

func (blockService BlockService) GetBeaconBestState() (*blockchain.BeaconBestState, error) {
	if blockService.IsBeaconBestStateNil() {
		Logger.log.Debugf("handleGetBeaconBestState result: %+v", nil)
		return nil, errors.New("Best State beacon not existed")
	}

	var beacon *blockchain.BeaconBestState

	cachedKey := memcache.GetBeaconBestStateCachedKey()
	cacheValue, err := blockService.MemCache.Get(cachedKey)
	if err == nil && len(cacheValue) > 0 {
		err1 := json.Unmarshal(cacheValue, &beacon)
		if err1 != nil {
			Logger.log.Error("Json Unmarshal cache of shard best state error", err1)
		}
	} else {
		beacon, err = blockService.BlockChain.BestState.GetClonedBeaconBestState()
		cacheValue, err := json.Marshal(beacon)
		if err == nil {
			err1 := blockService.MemCache.PutExpired(cachedKey, cacheValue, 10000)
			if err1 != nil {
				Logger.log.Error("Cache data of beacon best state error", err1)
			}
		}
	}
	return beacon, err
}

func (blockService BlockService) GetBeaconBestBlock() (*blockchain.BeaconBlock, error) {
	clonedBeaconBestState, err := blockService.BlockChain.BestState.GetClonedBeaconBestState()
	if err != nil {
		return nil, err
	}
	return &clonedBeaconBestState.BestBlock, nil
}

func (blockService BlockService) GetBeaconBestBlockHash() (*common.Hash, error) {
	clonedBeaconBestState, err := blockService.BlockChain.BestState.GetClonedBeaconBestState()
	if err != nil {
		return nil, err
	}
	return &clonedBeaconBestState.BestBlockHash, nil
}

func (blockService BlockService) RetrieveShardBlock(hashString string, verbosity string) (*jsonresult.GetBlockResult, *RPCError) {
	hash, errH := common.Hash{}.NewHashFromStr(hashString)
	if errH != nil {
		Logger.log.Debugf("handleRetrieveBlock result: %+v, err: %+v", nil, errH)
		return nil, NewRPCError(RPCInvalidParamsError, errH)
	}
	block, _, errD := blockService.BlockChain.GetShardBlockByHash(*hash)
	if errD != nil {
		Logger.log.Debugf("handleRetrieveBlock result: %+v, err: %+v", nil, errD)
		return nil, NewRPCError(GetShardBlockByHashError, errD)
	}
	result := jsonresult.GetBlockResult{}

	shardID := block.Header.ShardID

	if verbosity == "0" {
		data, err := json.Marshal(block)
		if err != nil {
			Logger.log.Debugf("handleRetrieveBlock result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(JsonError, err)
		}
		result.Data = hex.EncodeToString(data)
	} else if verbosity == "1" {
		best := blockService.BlockChain.BestState.Shard[shardID].BestBlock

		blockHeight := block.Header.Height
		// Get next block hash unless there are none.
		var nextHashString string
		// if blockHeight < best.Header.GetHeight() {
		if blockHeight < best.Header.Height {
			nextHash, err := blockService.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
			if err != nil {
				return nil, NewRPCError(GetShardBlockByHeightError, err)
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
		result.PreviousBlockHash = block.Header.PreviousBlockHash.String()
		result.NextBlockHash = nextHashString
		result.TxHashes = []string{}
		// result.BlockProducerSign = block.ProducerSig
		// result.BlockProducer = block.Header.ProducerAddress.String()
		// result.AggregatedSig = block.AggregatedSig
		result.BeaconHeight = block.Header.BeaconHeight
		result.BeaconBlockHash = block.Header.BeaconHash.String()
		// result.R = block.R
		result.ValidationData = block.ValidationData
		result.Round = block.Header.Round
		result.CrossShardBitMap = []int{}
		result.Instruction = block.Body.Instructions
		if len(block.Header.CrossShardBitMap) > 0 {
			for _, shardID := range block.Header.CrossShardBitMap {
				result.CrossShardBitMap = append(result.CrossShardBitMap, int(shardID))
			}
		}
		result.Epoch = block.Header.Epoch

		for _, tx := range block.Body.Transactions {
			result.TxHashes = append(result.TxHashes, tx.Hash().String())
		}
	} else if verbosity == "2" {
		best := blockService.BlockChain.BestState.Shard[shardID].BestBlock

		blockHeight := block.Header.Height
		// Get next block hash unless there are none.
		var nextHashString string
		if blockHeight < best.Header.Height {
			nextHash, err := blockService.BlockChain.GetShardBlockByHeight(blockHeight+1, shardID)
			if err != nil {
				Logger.log.Debugf("handleRetrieveBlock result: %+v, err: %+v", nil, err)
				return nil, NewRPCError(GetShardBlockByHeightError, err)
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
		result.PreviousBlockHash = block.Header.PreviousBlockHash.String()
		result.NextBlockHash = nextHashString
		// result.BlockProducerSign = block.ProducerSig
		// result.BlockProducer = block.Header.ProducerAddress.String()
		// result.AggregatedSig = block.AggregatedSig
		result.BeaconHeight = block.Header.BeaconHeight
		result.BeaconBlockHash = block.Header.BeaconHash.String()
		// result.R = block.R
		result.ValidationData = block.ValidationData
		result.Round = block.Header.Round
		result.CrossShardBitMap = []int{}
		result.Instruction = block.Body.Instructions
		instructions, err := blockchain.CreateShardInstructionsFromTransactionAndInstruction(block.Body.Transactions, blockService.BlockChain, block.Header.ShardID)
		if err == nil {
			result.Instruction = append(result.Instruction, instructions...)
		}
		if len(block.Header.CrossShardBitMap) > 0 {
			for _, shardID := range block.Header.CrossShardBitMap {
				result.CrossShardBitMap = append(result.CrossShardBitMap, int(shardID))
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
					return nil, NewRPCError(JsonError, err)
				}
				transactionT.HexData = hex.EncodeToString(data)
				transactionT.Locktime = txN.LockTime
			}

			result.Txs = append(result.Txs, transactionT)
		}
	}
	return &result, nil
}

func (blockService BlockService) RetrieveBeaconBlock(hashString string) (*jsonresult.GetBlocksBeaconResult, *RPCError) {
	hash, errH := common.Hash{}.NewHashFromStr(hashString)
	if errH != nil {
		Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, errH)
		return nil, NewRPCError(RPCInvalidParamsError, errH)
	}
	block, _, errD := blockService.BlockChain.GetBeaconBlockByHash(*hash)
	if errD != nil {
		Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, errD)
		return nil, NewRPCError(GetBeaconBlockByHashError, errD)
	}

	best := blockService.BlockChain.BestState.Beacon.BestBlock
	blockHeight := block.Header.Height
	// Get next block hash unless there are none.
	var nextHashString string
	// if blockHeight < best.Header.GetHeight() {
	if blockHeight < best.Header.Height {
		nextHash, err := blockService.BlockChain.GetBeaconBlockByHeight(blockHeight + 1)
		if err != nil {
			Logger.log.Debugf("handleRetrieveBeaconBlock result: %+v, err: %+v", nil, err)
			return nil, NewRPCError(GetBeaconBlockByHeightError, err)
		}
		nextHashString = nextHash.Hash().String()
	}
	blockBytes, errS := json.Marshal(block)
	if errS != nil {
		return nil, NewRPCError(UnexpectedError, errS)
	}
	result := jsonresult.NewGetBlocksBeaconResult(block, uint64(len(blockBytes)), nextHashString)
	return result, nil
}

func (blockService BlockService) GetBlocks(shardIDParam int, numBlock int) (interface{}, *RPCError) {
	result := make([]jsonresult.GetBlockResult, 0)
	resultBeacon := make([]jsonresult.GetBlocksBeaconResult, 0)
	var cacheKey []byte
	if shardIDParam != -1 {
		cacheKey = memcache.GetBlocksCachedKey(shardIDParam, numBlock)
		cacheValue, err := blockService.MemCache.Get(cacheKey)
		if err == nil && len(cacheValue) > 0 {
			err1 := json.Unmarshal(cacheValue, &result)
			if err1 != nil {
				Logger.log.Error("Json Unmarshal cache of get shard blocks error", err1)
			} else {
				return result, nil
			}
		}
	} else {
		cacheKey = memcache.GetBlocksCachedKey(shardIDParam, numBlock)
		cacheValue, err := blockService.MemCache.Get(cacheKey)
		if err == nil && len(cacheValue) > 0 {
			err1 := json.Unmarshal(cacheValue, &resultBeacon)
			if err1 != nil {
				Logger.log.Error("Json Unmarshal cache of get beacon blocks error", err1)
			} else {
				return resultBeacon, nil
			}
		}
	}

	if shardIDParam != -1 {
		if len(result) == 0 {
			shardID := byte(shardIDParam)
			clonedShardBestState, err := blockService.BlockChain.BestState.GetClonedAShardBestState(shardID)
			if err != nil {
				return nil, NewRPCError(GetClonedShardBestStateError, err)
			}
			bestBlock := clonedShardBestState.BestBlock
			previousHash := bestBlock.Hash()
			for numBlock > 0 {
				numBlock--
				// block, errD := blockService.BlockChain.GetBlockByHash(previousHash)
				block, size, errD := blockService.BlockChain.GetShardBlockByHash(*previousHash)
				if errD != nil {
					Logger.log.Debugf("handleGetBlocks result: %+v, err: %+v", nil, errD)
					return nil, NewRPCError(GetShardBlockByHashError, errD)
				}
				blockResult := jsonresult.NewGetBlockResult(block, size, common.EmptyString)
				result = append(result, *blockResult)
				previousHash = &block.Header.PreviousBlockHash
				if previousHash.String() == (common.Hash{}).String() {
					break
				}
			}
			Logger.log.Debugf("handleGetBlocks result: %+v", result)
			if len(result) > 0 {
				cacheValue, err := json.Marshal(result)
				if err == nil {
					err1 := blockService.MemCache.PutExpired(cacheKey, cacheValue, 10000)
					if err1 != nil {
						Logger.log.Error("Cache data of shard best state error", err1)
					}
				}
			}
		}
		return result, nil
	} else {
		if len(resultBeacon) == 0 {
			clonedBeaconBestState, err := blockService.BlockChain.BestState.GetClonedBeaconBestState()
			if err != nil {
				return nil, NewRPCError(GetClonedBeaconBestStateError, err)
			}
			bestBlock := clonedBeaconBestState.BestBlock
			previousHash := bestBlock.Hash()
			for numBlock > 0 {
				numBlock--
				// block, errD := blockService.BlockChain.GetBlockByHash(previousHash)
				block, size, errD := blockService.BlockChain.GetBeaconBlockByHash(*previousHash)
				if errD != nil {
					return nil, NewRPCError(GetBeaconBlockByHashError, errD)
				}
				blockResult := jsonresult.NewGetBlocksBeaconResult(block, size, common.EmptyString)
				resultBeacon = append(resultBeacon, *blockResult)
				previousHash = &block.Header.PreviousBlockHash
				if previousHash.String() == (common.Hash{}).String() {
					break
				}
			}
			Logger.log.Debugf("handleGetBlocks result: %+v", resultBeacon)
			if len(resultBeacon) > 0 {
				cacheValue, err := json.Marshal(resultBeacon)
				if err == nil {
					err1 := blockService.MemCache.PutExpired(cacheKey, cacheValue, 10000)
					if err1 != nil {
						Logger.log.Error("Cache data of shard best state error", err1)
					}
				}
			}
		}
		return resultBeacon, nil
	}
}

func (blockService BlockService) GetBeaconBlockByHeight(height uint64) (*blockchain.BeaconBlock, error) {
	return blockService.BlockChain.GetBeaconBlockByHeight(height)
}

func (blockService BlockService) GetShardBlockByHeight(height uint64, shardID byte) (*blockchain.ShardBlock, error) {
	return blockService.BlockChain.GetShardBlockByHeight(height, shardID)
}

func (blockService BlockService) IsBeaconBestStateNil() bool {
	return blockService.BlockChain.BestState == nil || blockService.BlockChain.BestState.Beacon == nil
}

func (blockService BlockService) IsShardBestStateNil() bool {
	return blockService.BlockChain.BestState == nil || blockService.BlockChain.BestState.Shard == nil || len(blockService.BlockChain.BestState.Shard) <= 0
}

func (blockService BlockService) GetValidStakers(publicKeys []string) ([]string, *RPCError) {
	beaconBestState, err := blockService.GetBeaconBestState()
	if err != nil {
		return nil, NewRPCError(GetClonedBeaconBestStateError, err)
	}

	validPublicKeys := beaconBestState.GetValidStakers(publicKeys)

	return validPublicKeys, nil
}

func (blockService BlockService) GetShardBlockByHash(hash common.Hash) (*blockchain.ShardBlock, uint64, error) {
	return blockService.BlockChain.GetShardBlockByHash(hash)
}

func (blockService BlockService) GetBeaconBlockByHash(hash common.Hash) (*blockchain.BeaconBlock, uint64, error) {
	return blockService.BlockChain.GetBeaconBlockByHash(hash)
}

func (blockService BlockService) CheckHashValue(hashStr string) (isTransaction bool, isShardBlock bool, isBeaconBlock bool, err error) {
	isTransaction = false
	isShardBlock = false
	isBeaconBlock = false

	hash, err2 := common.Hash{}.NewHashFromStr(hashStr)
	if err2 != nil {
		err = errors.New("expected hash string value")
		return
	}

	_, _, err = blockService.GetShardBlockByHash(*hash)
	if err == nil {
		isShardBlock = true
		return
	} else {
		_, _, err = blockService.GetBeaconBlockByHash(*hash)
		if err == nil {
			isBeaconBlock = true
			return
		} else {
			_, _, _, _, err = blockService.BlockChain.GetTransactionByHash(*hash)
			if err == nil {
				isTransaction = true
			} else {
				err = nil
			}
		}
	}

	return
}

func (blockService BlockService) GetActiveShards() int {
	return blockService.BlockChain.BestState.Beacon.ActiveShards
}

func (blockService BlockService) ListPrivacyCustomToken() (map[common.Hash]transaction.TxCustomTokenPrivacy, map[common.Hash]blockchain.CrossShardTokenPrivacyMetaData, error) {
	listTxInitPrivacyToken, listTxInitPrivacyTokenCrossShard, err := blockService.BlockChain.ListPrivacyCustomToken()
	return listTxInitPrivacyToken, listTxInitPrivacyTokenCrossShard, err
}

func (blockService BlockService) ListPrivacyCustomTokenCached() (map[common.Hash]transaction.TxCustomTokenPrivacy, map[common.Hash]blockchain.CrossShardTokenPrivacyMetaData, error) {
	listTxInitPrivacyToken := make(map[common.Hash]transaction.TxCustomTokenPrivacy)
	listTxInitPrivacyTokenCrossShard := make(map[common.Hash]blockchain.CrossShardTokenPrivacyMetaData)

	cachedKeyPrivacyToken := memcache.GetListPrivacyTokenCachedKey()
	cachedValuePrivacyToken, err := blockService.MemCache.Get(cachedKeyPrivacyToken)
	if err == nil && len(cachedValuePrivacyToken) > 0 {
		err1 := json.Unmarshal(cachedValuePrivacyToken, &listTxInitPrivacyToken)
		if err1 != nil {
			Logger.log.Error("Json Unmarshal cachedKeyPrivacyToken err", err1)
		}
	}

	cachedKeyPrivacyTokenCrossShard := memcache.GetListPrivacyTokenCrossShardCachedKey()
	cachedValuePrivacyTokenCrossShard, err := blockService.MemCache.Get(cachedKeyPrivacyTokenCrossShard)
	if err == nil && len(cachedValuePrivacyToken) > 0 {
		err1 := json.Unmarshal(cachedValuePrivacyTokenCrossShard, &listTxInitPrivacyTokenCrossShard)
		if err1 != nil {
			Logger.log.Error("Json Unmarshal cachedKeyPrivacyToken err", err1)
		}
	}

	if len(listTxInitPrivacyToken) == 0 || len(listTxInitPrivacyTokenCrossShard) == 0 {
		listTxInitPrivacyToken, listTxInitPrivacyTokenCrossShard, err = blockService.ListPrivacyCustomToken()

		for k, v := range listTxInitPrivacyToken {
			temp := v
			temp.Tx = transaction.Tx{}
			temp.TxPrivacyTokenData.TxNormal = transaction.Tx{}
			listTxInitPrivacyToken[k] = temp
		}
		cachedValuePrivacyToken, err = json.Marshal(listTxInitPrivacyToken)
		if err == nil {
			err1 := blockService.MemCache.PutExpired(cachedKeyPrivacyToken, cachedValuePrivacyToken, 60*1000)
			if err1 != nil {
				Logger.log.Error("Cached error cachedValuePrivacyToken", err1)
			}
		}

		cachedValuePrivacyTokenCrossShard, err = json.Marshal(listTxInitPrivacyTokenCrossShard)
		if err == nil {
			err1 := blockService.MemCache.PutExpired(cachedKeyPrivacyTokenCrossShard, cachedValuePrivacyTokenCrossShard, 60*1000)
			if err1 != nil {
				Logger.log.Error("Cached error cachedValuePrivacyTokenCrossShard", err1)
			}
		}
	}
	return listTxInitPrivacyToken, listTxInitPrivacyTokenCrossShard, err
}

func (blockService BlockService) GetAllCoinID() ([]common.Hash, error) {
	return blockService.BlockChain.GetAllCoinID()
}

func (blockService BlockService) GetMinerRewardFromMiningKey(incPublicKey []byte) (map[string]uint64, error) {
	allCoinIDs, err := blockService.GetAllCoinID()
	if err != nil {
		return nil, err
	}

	rewardAmountResult := make(map[string]uint64)
	rewardAmounts := make(map[common.Hash]uint64)

	for _, coinID := range allCoinIDs {
		amount, err := (*blockService.DB).GetCommitteeReward(incPublicKey, coinID)
		if err != nil {
			return nil, err
		}
		if coinID == common.PRVCoinID {
			rewardAmountResult["PRV"] = amount
		} else {
			rewardAmounts[coinID] = amount
		}
	}

	privateToken, crossPrivateToken, err := blockService.ListPrivacyCustomToken()
	if err != nil {
		return nil, err
	}

	for _, token := range privateToken {
		if rewardAmounts[token.TxPrivacyTokenData.PropertyID] > 0 {
			rewardAmountResult[token.TxPrivacyTokenData.PropertyID.String()] = rewardAmounts[token.TxPrivacyTokenData.PropertyID]
		}
	}

	for _, token := range crossPrivateToken {
		if rewardAmounts[token.TokenID] > 0 {
			rewardAmountResult[token.TokenID.String()] = rewardAmounts[token.TokenID]
		}
	}

	return rewardAmountResult, nil
}

func (blockService BlockService) RevertBeacon() error {
	return blockService.BlockChain.RevertBeaconState()
}

func (blockService BlockService) RevertShard(shardID byte) error {
	return blockService.BlockChain.RevertShardState(shardID)
}

func (blockService BlockService) GetRewardAmount(paymentAddress string) (map[string]uint64, *RPCError) {
	rewardAmountResult := make(map[string]uint64)
	rewardAmounts := make(map[common.Hash]uint64)

	keySet, _, err := GetKeySetFromPaymentAddressParam(paymentAddress)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	publicKey := keySet.PaymentAddress.Pk
	if publicKey == nil {
		return rewardAmountResult, nil
	}

	allCoinIDs, err := blockService.BlockChain.GetAllCoinID()
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}

	for _, coinID := range allCoinIDs {
		amount, err := (*blockService.DB).GetCommitteeReward(publicKey, coinID)
		if err != nil {
			return nil, NewRPCError(UnexpectedError, err)
		}
		if coinID == common.PRVCoinID {
			rewardAmountResult["PRV"] = amount
		} else {
			rewardAmounts[coinID] = amount
		}
	}

	cusPrivTok, crossPrivToken, err := blockService.BlockChain.ListPrivacyCustomToken()

	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}

	for _, token := range cusPrivTok {
		if rewardAmounts[token.TxPrivacyTokenData.PropertyID] > 0 {
			rewardAmountResult[token.TxPrivacyTokenData.PropertyID.String()] = rewardAmounts[token.TxPrivacyTokenData.PropertyID]
		}
	}

	for _, token := range crossPrivToken {
		if rewardAmounts[token.TokenID] > 0 {
			rewardAmountResult[token.TokenID.String()] = rewardAmounts[token.TokenID]
		}
	}

	return rewardAmountResult, nil
}

func (blockService BlockService) CanPubkeyStake(publicKey string) (bool, error) {
	canStake := true
	validStakers, err := blockService.GetValidStakers([]string{publicKey})
	if err != nil {
		return false, err
	}

	if len(validStakers) == 0 {
		canStake = false
	} else {
		// get pool candidate
		poolCandidate := blockService.TxMemPool.GetClonedPoolCandidate()
		if common.IndexOfStrInHashMap(publicKey, poolCandidate) > 0 {
			canStake = false
		}
	}

	return canStake, nil
}

func (blockService BlockService) GetBlockHashByHeight(shardID int, height uint64) (string, error) {
	var hash *common.Hash
	var err error
	var beaconBlock *blockchain.BeaconBlock
	var shardBlock *blockchain.ShardBlock

	isGetBeacon := shardID == -1

	if isGetBeacon {
		beaconBlock, err = blockService.GetBeaconBlockByHeight(height)
	} else {
		shardBlock, err = blockService.GetShardBlockByHeight(height, byte(shardID))
	}

	if err != nil {
		Logger.log.Debugf("handleGetBlockHash result: %+v", nil)
		return "", err
	}

	if isGetBeacon {
		hash = beaconBlock.Hash()
	} else {
		hash = shardBlock.Hash()
	}

	result := hash.String()
	return result, nil
}

func (blockService BlockService) GetBlockHeader(getBy string, blockParam string, shardID float64) (
	*blockchain.ShardHeader, int, string, *RPCError) {
	switch getBy {
	case "blockhash":
		hash := common.Hash{}
		err := hash.Decode(&hash, blockParam)
		log.Printf("%+v", hash)
		if err != nil {
			Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
			return nil, 0, "", NewRPCError(RPCInvalidParamsError, errors.New("invalid blockhash format"))
		}
		// block, err := httpServer.config.BlockChain.GetBlockByHash(&bhash)
		block, _, err := blockService.BlockChain.GetShardBlockByHash(hash)
		if err != nil {
			Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
			return nil, 0, "", NewRPCError(GetShardBlockByHashError, errors.New("blockParam not exist"))
		}

		blockNum := int(block.Header.Height) + 1

		return &block.Header, blockNum, hash.String(), nil
	case "blocknum":
		blockNum, err := strconv.Atoi(blockParam)
		if err != nil {
			Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
			return nil, 0, "", NewRPCError(RPCInvalidParamsError, errors.New("invalid blocknum format"))
		}

		shard, err := blockService.GetShardBestStateByShardID(byte(shardID))
		if err != nil {
			return nil, 0, "", NewRPCError(GetClonedShardBestStateError, err)
		}

		var blockHeader = blockchain.ShardHeader{}
		var blockHash = ""
		if uint64(blockNum-1) > shard.BestBlock.Header.Height || blockNum <= 0 {
			Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
			return nil, 0, "", NewRPCError(GetShardBestBlockError, errors.New("Block not exist"))
		}
		block, _ := blockService.GetShardBlockByHeight(uint64(blockNum-1), uint8(shardID))
		if block != nil {
			blockHeader = block.Header
			blockHash = block.Hash().String()
		}

		return &blockHeader, blockNum, blockHash, nil
	default:
		Logger.log.Debugf("handleGetBlockHeader result: %+v", nil)
		return nil, 0, "", NewRPCError(RPCInvalidParamsError, errors.New("wrong request format"))
	}
}
