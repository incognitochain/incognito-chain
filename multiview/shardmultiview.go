package multiview

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
)

type ShardMultiView struct {
	*multiView
}

func NewShardMultiView() *ShardMultiView {
	sv := &ShardMultiView{}
	sv.multiView = NewMultiView()
	return sv
}

func (s *ShardMultiView) SimulateAddView(view View) MultiView {
	sv := &ShardMultiView{}
	sv.multiView = s.Clone().(*multiView)
	sv.AddView(view)
	return sv
}

func (s *ShardMultiView) AddView(v View) (res int, err error) {
	added := s.multiView.addView(v)
	shardBlock := v.GetBlock()
	if shardBlock.GetVersion() < types.INSTANT_FINALITY_VERSION {
		err = s.FinalizeView(*s.GetExpectedFinalView().GetHash())
	}
	if added {
		res = 1
	}
	return res, err
}
