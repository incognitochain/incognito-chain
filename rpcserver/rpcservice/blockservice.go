package rpcservice

import (
	"encoding/hex"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"errors"
)

type BlockService struct {
	BlockChain *blockchain.BlockChain
	DB  *database.DatabaseInterface
}

func (blockService BlockService) GetShardBestStates() map[byte]*blockchain.ShardBestState {
	shards := blockService.BlockChain.BestState.GetClonedAllShardBestState()
	return shards
}

func (blockService BlockService) GetShardBestStateByShardID(shardID byte) (*blockchain.ShardBestState, error) {
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
	beacon, err := blockService.BlockChain.BestState.GetClonedBeaconBestState()
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
	if shardIDParam != -1 {
		result := make([]jsonresult.GetBlockResult, 0)
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
		return result, nil
	} else {
		result := make([]jsonresult.GetBlocksBeaconResult, 0)
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

func (blockService BlockService) GetBeaconBlockByHeight(height uint64) (*blockchain.BeaconBlock, error) {
	return blockService.BlockChain.GetBeaconBlockByHeight(height)
}

func (blockService BlockService) GetShardBlockByHeight(height uint64, shardID byte) (*blockchain.ShardBlock, error) {
	return blockService.BlockChain.GetShardBlockByHeight(height, shardID)
}

func (blockService BlockService) IsBeaconBestStateNil() (bool) {
	return blockService.BlockChain.BestState.Beacon == nil
}

func (blockService BlockService) IsShardBestStateNil() (bool) {
	return blockService.BlockChain.BestState.Shard == nil || len(blockService.BlockChain.BestState.Shard) <= 0
}

func (blockService BlockService) GetValidStakers(publicKeys []string) ([]string,  *RPCError) {
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

func (blockService BlockService) CheckHashValue(hashStr string) (isTransaction bool, isShardBlock bool, isBeaconBlock bool, err error){
	isTransaction = false
	isShardBlock = false
	isBeaconBlock = false

	hash, err2 := common.Hash{}.NewHashFromStr(hashStr)
	if err2 != nil {
		err = errors.New("expected hash string value")
		return
	}

	_, _, err = blockService.GetShardBlockByHash(*hash)
	if err == nil{
		isShardBlock = true
		return
	} else {
		_, _, err = blockService.GetBeaconBlockByHash(*hash)
		if err == nil{
			isBeaconBlock = true
			return
		} else{
			_, _, _, _, err = blockService.BlockChain.GetTransactionByHash(*hash)
			if err == nil{
				isTransaction = true
			} else {
				err = nil
			}
		}
	}

	return
}

func (blockService BlockService) GetActiveShards() (int) {
	return blockService.BlockChain.BestState.Beacon.ActiveShards
}

func (blockService BlockService) ListPrivacyCustomToken() (map[common.Hash]transaction.TxCustomTokenPrivacy, map[common.Hash]blockchain.CrossShardTokenPrivacyMetaData, error){
	return blockService.BlockChain.ListPrivacyCustomToken()
}


func (blockService BlockService) GetAllCoinID() ([]common.Hash, error){
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



