package multiview

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
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

func TestNewMultiView(t *testing.T) {

	multiView := NewMultiView()
	multiView.AddView(&FakeView{
		&common.Hash{1},
		&common.Hash{0},
		1,
		1,
		1,
	})

	var print = func() {
		fmt.Println("best", multiView.bestView.GetHash().String())
		fmt.Println("final", multiView.finalView.GetHash().String())
		fmt.Println(" ")
	}

	multiView.AddView(&FakeView{
		&common.Hash{2},
		&common.Hash{1},
		2,
		2,
		2,
	})

	multiView.AddView(&FakeView{
		&common.Hash{3},
		&common.Hash{2},
		3,
		3,
		3,
	})

	multiView.AddView(&FakeView{
		&common.Hash{4},
		&common.Hash{3},
		4,
		4,
		4,
	})

	multiView.AddView(&FakeView{
		&common.Hash{5},
		&common.Hash{3},
		4,
		5,
		5,
	})

	print()

	for _, v := range multiView.GetAllViewsWithBFS() {
		fmt.Println("node", v.GetHash().String())
		fmt.Println("->")
	}

	if multiView.bestView.GetHash().String() != "0000000000000000000000000000000000000000000000000000000000000004" || multiView.finalView.GetHash().String() != "0000000000000000000000000000000000000000000000000000000000000003" {
		panic("Wrong")
	}
}
