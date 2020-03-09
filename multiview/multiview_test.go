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
	time     int
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

func (s *FakeView) GetBlockTime() int {
	return s.time
}

func TestNewMultiView(t *testing.T) {

	initView := []View{&FakeView{
		&common.Hash{1},
		nil,
		1,
		1,
		1,
	}}
	multiView := NewMultiView(initView)
	var print = func() {
		fmt.Println("best", multiView.bestView.GetHash().String())
		fmt.Println("final", multiView.finalView.GetHash().String())
		fmt.Println(" ")
	}

	print()

	multiView.AddView(&FakeView{
		&common.Hash{2},
		&common.Hash{1},
		2,
		2,
		2,
	})

	print()
	multiView.AddView(&FakeView{
		&common.Hash{3},
		&common.Hash{2},
		3,
		3,
		3,
	})

	print()
	multiView.AddView(&FakeView{
		&common.Hash{4},
		&common.Hash{3},
		4,
		4,
		4,
	})

	print()

	if multiView.bestView.GetHash().String() != "0000000000000000000000000000000000000000000000000000000000000004" || multiView.finalView.GetHash().String() != "0000000000000000000000000000000000000000000000000000000000000003" {
		panic("Wrong")
	}
}
