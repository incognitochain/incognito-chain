package multiview

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/multiview/mocks"
	"testing"
)

type FakeView struct {
	hash     *common.Hash
	prevHash *common.Hash
	height   uint64
	timeslot uint64
	time     int64
}

func (s FakeView) GetHash() *common.Hash {
	return s.hash
}

func (s *FakeView) GetPreviousHash() *common.Hash {
	return s.prevHash
}

func (s *FakeView) GetHeight() uint64 {
	return s.height
}

func (s *FakeView) GetTimeSlot() uint64 {
	return s.timeslot
}

func (s *FakeView) GetBlockTime() int64 {
	return s.time
}

// TODO: @hung build unit-test
func Test_getAllChains(t *testing.T) {

	views := []*mocks.View{}
	hashs := []common.Hash{}
	viewsByPreHash := make(map[common.Hash][]View)
	for i := 0; i < 10; i++ {
		t := common.Hash{byte(i)}
		hashs = append(hashs, t)
		view := &mocks.View{}
		view.On("GetHash").Return(&t).Times(10)
		views = append(views, view)
	}
	views[0].On("GetHeight").Return(uint64(1)).Times(10)
	views[1].On("GetHeight").Return(uint64(2)).Times(10)
	views[2].On("GetHeight").Return(uint64(2)).Times(10)
	views[3].On("GetHeight").Return(uint64(2)).Times(10)
	views[4].On("GetHeight").Return(uint64(3)).Times(10)
	views[5].On("GetHeight").Return(uint64(3)).Times(10)
	views[6].On("GetHeight").Return(uint64(3)).Times(10)
	views[7].On("GetHeight").Return(uint64(3)).Times(10)
	views[8].On("GetHeight").Return(uint64(4)).Times(10)
	views[9].On("GetHeight").Return(uint64(4)).Times(10)
	/**
	0 -> 1 -> 4
	  -> 2 -> 5
		   -> 6 -> 8
		   -> 7 -> 9
	  -> 3
	*/
	viewsByPreHash[hashs[0]] = []View{views[1], views[2], views[3]}
	viewsByPreHash[hashs[1]] = []View{views[4]}
	viewsByPreHash[hashs[2]] = []View{views[5], views[6], views[7]}
	viewsByPreHash[hashs[6]] = []View{views[8]}
	viewsByPreHash[hashs[7]] = []View{views[9]}

	res := [][]View{}
	getAllChains(hashs[0], viewsByPreHash, &res, []View{views[0]})
	for i, v := range res {
		t.Log(i)
		s1 := ""
		s2 := ""
		for _, v1 := range v {
			s1 += fmt.Sprintf("%d ", v1.GetHeight())
			s2 += fmt.Sprintf("%s ", v1.GetHash().String())
		}
		t.Log(s1)
		t.Log(s2)
	}
}
