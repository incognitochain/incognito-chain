package syncker

import (
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

const RUNNING_SYNC = "running_sync"
const STOP_SYNC = "stop_sync"

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func InsertBatchBlock(chain Chain, blocks []types.BlockInterface) (int, error) {
	sameCommitteeBlock := blocks

	containSwap := func(inst [][]string) bool {
		for _, inst := range inst {
			if inst[0] == instruction.SWAP_ACTION {
				return true
			}
		}
		return false
	}

	for i, v := range blocks {
		if chain.CommitteeStateVersion() == committeestate.SELF_SWAP_SHARD_VERSION {
			shouldBreak := false
			switch v.(type) {
			case *types.BeaconBlock:
				// do nothing, beacon committee assume not change
				//if v.GetCurrentEpoch() == curEpoch+1 {
				//	sameCommitteeBlock = blocks[:i+1]
				//	break
				//}
			case *types.ShardBlock:
				//if block contain swap inst,
				if containSwap(v.(*types.ShardBlock).Body.Instructions) {
					sameCommitteeBlock = blocks[:i+1]
					shouldBreak = true
				}
			}
			if shouldBreak {
				break
			}
		} else {
			//TODO: Checking committees for beacon when release beacon
			if i != len(blocks)-1 {
				if v.CommitteeFromBlock().String() != blocks[i+1].CommitteeFromBlock().String() {
					sameCommitteeBlock = blocks[:i+1]
					break
				}
			}
		}
	}

	for i, blk := range sameCommitteeBlock {
		if i == len(sameCommitteeBlock)-1 {
			break
		}
		if blk.GetHeight() != sameCommitteeBlock[i+1].GetHeight()-1 {
			sameCommitteeBlock = blocks[:i+1]
			break
		}
	}

	epochCommittee := []incognitokey.CommitteePublicKey{}

	if len(sameCommitteeBlock) != 0 {
		var err error
		//validate the last block for batching
		epochCommittee, err = chain.GetCommitteeV2(sameCommitteeBlock[0])
		if err != nil {
			return 0, err
		}
	}

	validBlockForInsert := sameCommitteeBlock[:]
	for i := len(sameCommitteeBlock) - 1; i >= 0; i-- {
		if err := chain.ValidateBlockSignatures(sameCommitteeBlock[i], epochCommittee); err != nil {
			validBlockForInsert = sameCommitteeBlock[:i]
		} else {
			break
		}
	}

	batchingValidate := true
	//if no valid block, this could be a fork chain, or the chunks that have old committee (current best block have swap) => try to insert all with full validation
	if len(validBlockForInsert) == 0 {
		validBlockForInsert = sameCommitteeBlock[:]
		batchingValidate = false
	}

	for i, v := range validBlockForInsert {
		if !chain.CheckExistedBlk(v) {
			var err error
			if i == 0 {
				err = chain.InsertBlock(v, true)
			} else {
				err = chain.InsertBlock(v, batchingValidate == false)
			}
			if err != nil {
				committeeStr, _ := incognitokey.CommitteeKeyListToString(epochCommittee)
				Logger.Errorf("Insert block %v hash %v got error %v, Committee of epoch %v", v.GetHeight(), *v.Hash(), err, committeeStr)
				return 0, err
			}
		}
	}
	return len(validBlockForInsert), nil
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
