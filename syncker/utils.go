package syncker

import (
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/config"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

const RUNNING_SYNC = "running_sync"
const STOP_SYNC = "stop_sync"

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}
func InsertBatchBlock(chain Chain, blocks []types.BlockInterface, cQuit chan struct{}) (int, error) {
	//if full validation, we dont need batch block
	if config.Config().IsFullValidation {
		for i, v := range blocks {
			if !chain.CheckExistedBlk(v) {
				err := chain.InsertBlock(v, true)
				if err != nil {
					return i, err
				}
			}
		}
		return len(blocks), nil
	}

	//not fullvalidation, we batch block which having same committee
	sameCommitteeBlock := blocks
	validatorRootHashChange := func(blk interface{}) bool {
		switch blk.(type) {
		case *types.BeaconBlock:
			header := blk.(*types.BeaconBlock).Header
			bestHeader := chain.GetBestView().GetBlock().(*types.BeaconBlock).Header
			if header.ShardCandidateRoot != bestHeader.ShardCandidateRoot || header.ShardCommitteeAndValidatorRoot != bestHeader.ShardCommitteeAndValidatorRoot ||
				header.BeaconCommitteeAndValidatorRoot != bestHeader.BeaconCommitteeAndValidatorRoot || header.BeaconCandidateRoot != bestHeader.BeaconCandidateRoot {
				return true
			}
		case *types.ShardBlock:
			header := blk.(*types.ShardBlock).Header
			bestHeader := chain.GetBestView().GetBlock().(*types.ShardBlock).Header
			if header.CommitteeRoot != bestHeader.CommitteeRoot || header.PendingValidatorRoot != header.PendingValidatorRoot {
				return true
			}
		}
		return false
	}

	//if validator root hash change -> break into chunks
	for i, v := range blocks {
		if validatorRootHashChange(v) {
			if i == 0 {
				sameCommitteeBlock = blocks[:1]
			} else {
				sameCommitteeBlock = blocks[:i]
			}
			break
		}

	}

	//check block height is sequential
	for i, blk := range sameCommitteeBlock {
		if i == len(sameCommitteeBlock)-1 {
			break
		}
		if blk.GetHeight() != sameCommitteeBlock[i+1].GetHeight()-1 {
			sameCommitteeBlock = blocks[:i+1]
			break
		}
	}

	//validate the last block for batching
	//get block has same committee
	committees := []incognitokey.CommitteePublicKey{}
	if len(sameCommitteeBlock) != 0 {
		var err error
		committees, err = chain.GetCommitteeV2(sameCommitteeBlock[0])
		if err != nil {
			return 0, err
		}
	}
	validBlockForInsert := sameCommitteeBlock[:]
	for i := len(sameCommitteeBlock) - 1; i >= 0; i-- {
		signingCommittees, err := chain.GetCommitteeV2(sameCommitteeBlock[i])
		if err != nil {
			return 0, err
		}
		if err := chain.ValidateBlockSignatures(sameCommitteeBlock[i], signingCommittees); err != nil {
			sameCommitteeBlock = sameCommitteeBlock[:i]
		} else {
			break
		}
	}

	batchingValidate := true
	if time.Now().Unix()-chain.GetBestView().GetBlock().GetProduceTime() < 24*60*60 { //only batching insert when block is created more than 1 day ago
		batchingValidate = false
	}

	//if no valid block, this could be a fork chain, or the chunks that have old committee (current best block have swap) => try to insert all with full validation
	if len(validBlockForInsert) == 0 {
		validBlockForInsert = sameCommitteeBlock[:]
		batchingValidate = false
	}
	firstInsert := true
	for _, v := range validBlockForInsert {
		if !chain.CheckExistedBlk(v) {
			var err error
			if firstInsert { //always validate the first block even in batch mode
				//Logger.Infof("[config] block height %v proposetime %v", v.GetHeight(), v.GetProposeTime())
				err = chain.InsertBlock(v, true)
				firstInsert = false
			} else {
				err = chain.InsertBlock(v, batchingValidate == false)
			}
			if err != nil {
				committeeStr, _ := incognitokey.CommitteeKeyListToString(committees)
				Logger.Errorf("Insert block %v hash %v got error %v, Committee of epoch %v", v.GetHeight(), *v.Hash(), err, committeeStr)
				return 0, err
			}
		}

	}

	return len(sameCommitteeBlock), nil
}

//final block
func GetFinalBlockFromBlockHash_v1(currentFinalHash string, byHash map[string]types.BlockPoolInterface, byPrevHash map[string][]string) (res []types.BlockPoolInterface) {
	var finalBlock types.BlockPoolInterface = nil
	var traverse func(currentHash string)
	traverse = func(currentHash string) {
		if byPrevHash[currentHash] == nil {
			return
		} else {
			if finalBlock == nil {
				finalBlock = byHash[currentHash]
			} else if finalBlock.GetHeight() < byHash[currentHash].GetHeight() {
				finalBlock = byHash[currentHash]
			}
			for _, nextHash := range byPrevHash[currentHash] {
				traverse(nextHash)
			}
		}
	}
	traverse(currentFinalHash)

	if finalBlock == nil {
		return nil
	}

	for {
		if currentFinalHash == finalBlock.Hash().String() {
			return
		}
		res = append([]types.BlockPoolInterface{byHash[finalBlock.Hash().String()]}, res...)
		finalBlock = byHash[finalBlock.GetPrevHash().String()]
		if finalBlock == nil || finalBlock.Hash().String() == currentFinalHash {
			break
		}
	}
	return res
}

func GetLongestChain(currentFinalHash string, byHash map[string]types.BlockPoolInterface, byPrevHash map[string][]string) (res []types.BlockPoolInterface) {
	var finalBlock types.BlockPoolInterface = nil
	var traverse func(currentHash string)
	traverse = func(currentHash string) {
		if byPrevHash[currentHash] == nil {
			if finalBlock == nil {
				finalBlock = byHash[currentHash]
			} else if finalBlock.GetHeight() < byHash[currentHash].GetHeight() {
				finalBlock = byHash[currentHash]
			}
			return
		} else {

			for _, nextHash := range byPrevHash[currentHash] {
				traverse(nextHash)
			}
		}
	}
	traverse(currentFinalHash)

	if finalBlock == nil {
		return nil
	}

	for {
		res = append([]types.BlockPoolInterface{byHash[finalBlock.Hash().String()]}, res...)
		finalBlock = byHash[finalBlock.GetPrevHash().String()]
		if finalBlock == nil {
			break
		}
	}
	return res
}

func GetPoolInfo(byHash map[string]types.BlockPoolInterface) (res []types.BlockPoolInterface) {
	for _, v := range byHash {
		res = append(res, v)
	}
	return res
}

func compareLists(poolList map[byte][]interface{}, hashList map[byte][]common.Hash) (diffHashes map[byte][]common.Hash) {
	diffHashes = make(map[byte][]common.Hash)
	poolListsHash := make(map[byte][]common.Hash)
	for shardID, blkList := range poolList {
		for _, blk := range blkList {
			blkHash := blk.(types.BlockPoolInterface).Hash()
			poolListsHash[shardID] = append(poolListsHash[shardID], *blkHash)
		}
	}

	for shardID, blockHashes := range hashList {
		if blockList, ok := poolListsHash[shardID]; ok {
			for _, blockHash := range blockHashes {
				if exist, _ := common.SliceExists(blockList, blockHash); !exist {
					diffHashes[shardID] = append(diffHashes[shardID], blockHash)
				}
			}
		} else {
			diffHashes[shardID] = blockHashes
		}
	}
	return diffHashes
}

func compareListsByHeight(poolList map[byte][]interface{}, heightList map[byte][]uint64) (diffHeights map[byte][]uint64) {
	diffHeights = make(map[byte][]uint64)
	poolListsHeight := make(map[byte][]uint64)
	for shardID, blkList := range poolList {
		for _, blk := range blkList {
			blkHeight := blk.(types.BlockPoolInterface).GetHeight()
			poolListsHeight[shardID] = append(poolListsHeight[shardID], blkHeight)
		}
	}

	for shardID, blockHeights := range heightList {
		if blockList, ok := poolListsHeight[shardID]; ok {
			for _, height := range blockHeights {
				if exist, _ := common.SliceExists(blockList, height); !exist {
					diffHeights[shardID] = append(diffHeights[shardID], height)
				}
			}
		} else {
			diffHeights[shardID] = blockHeights
		}
	}
	return diffHeights
}

func GetBlksByPrevHash(
	prevHash string,
	byHash map[string]types.BlockPoolInterface,
	byPrevHash map[string][]string,
) (
	res []types.BlockPoolInterface,
) {
	if hashes, ok := byPrevHash[prevHash]; ok {
		for _, hash := range hashes {
			if blk, exist := byHash[hash]; exist {
				res = append(res, blk)
			}
		}
	}
	return res
}

func GetAllViewFromHash(
	rHash string,
	byHash map[string]types.BlockPoolInterface,
	byPrevHash map[string][]string,
) (
	res []types.BlockPoolInterface,
) {
	hashes := []string{rHash}
	for {
		if len(hashes) == 0 {
			return res
		}
		hash := hashes[0]
		hashes = hashes[1:]
		for h, blk := range byHash {
			if blk.GetPrevHash().String() == hash {
				hashes = append(hashes, h)
				res = append(res, blk)
			}
		}
	}
}
