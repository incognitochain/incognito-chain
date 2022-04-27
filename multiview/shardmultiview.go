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
func (s *ShardMultiView) AddViewWithFinalizedHash(v View, newFinalizedHash *common.Hash) (int, error) {
	s.multiView.addView(v)
	shardBlock := v.GetBlock()

	if shardBlock.GetVersion() >= types.INSTANT_FINALITY_VERSION {
		if newFinalizedHash != nil {
			s.FinalizeView(*newFinalizedHash)
		}
	} else {
		s.FinalizeView(*s.GetExpectedFinalView().GetHash())
	}

	return 0, nil
}
