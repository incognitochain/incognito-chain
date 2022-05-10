package multiview

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
)

type ShardMultiView struct {
	*multiView
}

func NewShardMultiView() *ShardMultiView {
	sv := &ShardMultiView{}
	sv.multiView = NewMultiView()
	return sv
}

func (s *ShardMultiView) AddView(v View) (int, error) {
	added := s.multiView.addView(v)
	res := 0
	if added {
		res = 1
	}
	return res, nil
}

func (s *ShardMultiView) AddViewWithFinalizedHash(v View, newFinalizedHash *common.Hash) (res int, err error) {
	added := s.multiView.addView(v)
	shardBlock := v.GetBlock()

	if shardBlock.GetVersion() >= types.INSTANT_FINALITY_VERSION {
		if newFinalizedHash != nil {
			err = s.FinalizeView(*newFinalizedHash)
		}
	} else {
		err = s.FinalizeView(*s.GetExpectedFinalView().GetHash())
	}
	if added {
		res = 1
	}
	return res, err
}
