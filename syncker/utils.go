package syncker

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
)

const RUNNING_SYNC = "running_sync"
const STOP_SYNC = "stop_sync"

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func InsertBatchBlock(chain Chain, blocks []common.BlockInterface) (int, error) {
	curEpoch := chain.GetEpoch()
	sameCommitteeBlock := blocks
	for i, v := range blocks {
		if v.GetCurrentEpoch() == curEpoch+1 {
			sameCommitteeBlock = blocks[:i+1]
			break
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

	for i := len(sameCommitteeBlock) - 1; i >= 0; i-- {
		if err := chain.ValidateBlockSignatures(sameCommitteeBlock[i], chain.GetCommittee()); err != nil {
			sameCommitteeBlock = sameCommitteeBlock[:i]
		} else {
			break
		}
	}

	if len(sameCommitteeBlock) > 0 {
		if sameCommitteeBlock[0].GetHeight()-1 != chain.CurrentHeight() {
			return 0, errors.New(fmt.Sprintf("Not expected height: %d %d", sameCommitteeBlock[0].GetHeight()-1, chain.CurrentHeight()))
		}
	}

	for _, v := range sameCommitteeBlock {
		err := chain.InsertBlk(v)
		if err != nil {
			return 0, err
		}
	}
	return len(sameCommitteeBlock), nil
}

//final block
func GetFinalBlockFromBlockHash_v1(currentFinalHash string, byHash map[string]common.BlockPoolInterface, byPrevHash map[string][]string) (res []common.BlockPoolInterface) {
	var finalBlock common.BlockPoolInterface = nil
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
		res = append([]common.BlockPoolInterface{byHash[finalBlock.Hash().String()]}, res...)
		finalBlock = byHash[finalBlock.GetPrevHash()]
		if finalBlock == nil || finalBlock.Hash().String() == currentFinalHash {
			break
		}
	}
	return res
}

func GetLongestChain(currentFinalHash string, byHash map[string]common.BlockPoolInterface, byPrevHash map[string][]string) (res []common.BlockPoolInterface) {
	var finalBlock common.BlockPoolInterface = nil
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
		res = append([]common.BlockPoolInterface{byHash[finalBlock.Hash().String()]}, res...)
		finalBlock = byHash[finalBlock.GetPrevHash()]
		if finalBlock == nil {
			break
		}
	}
	return res
}
