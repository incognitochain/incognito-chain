package syncker

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

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
