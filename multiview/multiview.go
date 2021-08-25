package multiview

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"time"

	"github.com/incognitochain/incognito-chain/common"
)

type MultiView struct {
	viewByHash     map[common.Hash]types.View //viewByPrevHash map[common.Hash][]View
	viewByPrevHash map[common.Hash][]types.View
	actionCh       chan func()

	//state
	finalView types.View
	bestView  types.View
}

func NewMultiView() *MultiView {
	s := &MultiView{
		viewByHash:     make(map[common.Hash]types.View),
		viewByPrevHash: make(map[common.Hash][]types.View),
		actionCh:       make(chan func()),
	}

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case f := <-s.actionCh:
				f()
			case <-ticker.C:
				if len(s.viewByHash) > 100 {
					s.removeOutdatedView()
				}
			}
		}
	}()

	return s

}

func (multiView *MultiView) Reset() {
	multiView.viewByHash = make(map[common.Hash]types.View)
	multiView.viewByPrevHash = make(map[common.Hash][]types.View)
}

func (multiView *MultiView) Clone() *MultiView {
	s := NewMultiView()
	for h, v := range multiView.viewByHash {
		s.viewByHash[h] = v
	}
	for h, v := range multiView.viewByPrevHash {
		s.viewByPrevHash[h] = v
	}
	s.finalView = multiView.finalView
	s.bestView = multiView.bestView
	return s
}

func (multiView *MultiView) removeOutdatedView() {
	for h, v := range multiView.viewByHash {
		if v.GetHeight() < multiView.finalView.GetHeight() {
			delete(multiView.viewByHash, h)
			delete(multiView.viewByPrevHash, h)
			delete(multiView.viewByPrevHash, *v.GetPreviousHash())
		}
	}
}

func (multiView *MultiView) GetViewByHash(hash common.Hash) types.View {
	res := make(chan types.View)
	multiView.actionCh <- func() {
		view, _ := multiView.viewByHash[hash]
		if view == nil || view.GetHeight() < multiView.finalView.GetHeight() {
			res <- nil
		} else {
			res <- view
		}
	}
	return <-res
}

//Only add view if view is validated (at least enough signature)
func (multiView *MultiView) AddView(view types.View) bool {
	res := make(chan bool)
	multiView.actionCh <- func() {
		if len(multiView.viewByHash) == 0 { //if no view in map, this is init view -> always allow
			multiView.viewByHash[*view.GetHash()] = view
			multiView.updateViewState(view)
			res <- true
			return
		} else if _, ok := multiView.viewByHash[*view.GetHash()]; !ok { //otherwise, if view is not yet inserted
			if _, ok := multiView.viewByHash[*view.GetPreviousHash()]; ok { // view must point to previous valid view
				multiView.viewByHash[*view.GetHash()] = view
				multiView.viewByPrevHash[*view.GetPreviousHash()] = append(multiView.viewByPrevHash[*view.GetPreviousHash()], view)
				multiView.updateViewState(view)
				res <- true
				return
			}
		}
		res <- false
	}
	return <-res
}

func (multiView *MultiView) GetBestView() types.View {
	return multiView.bestView
}

func (multiView *MultiView) GetFinalView() types.View {
	return multiView.finalView
}

//update view whenever there is new view insert into system
func (multiView *MultiView) updateViewState(newView types.View) {
	defer func() {
		if multiView.viewByHash[*multiView.finalView.GetPreviousHash()] != nil {
			delete(multiView.viewByHash, *multiView.finalView.GetPreviousHash())
			delete(multiView.viewByPrevHash, *multiView.finalView.GetPreviousHash())
		}
	}()

	if multiView.finalView == nil {
		multiView.bestView = newView
		multiView.finalView = newView
		return
	}

	//update bestView
	if newView.GetHeight() > multiView.bestView.GetHeight() {
		multiView.bestView = newView
	}

	//get bestview with min produce time
	if newView.GetHeight() == multiView.bestView.GetHeight() && newView.GetBlock().GetProduceTime() < multiView.bestView.GetBlock().GetProduceTime() {
		multiView.bestView = newView
	}

	if newView.GetBlock().GetVersion() == types.BFT_VERSION {
		//update finalView: consensus 1
		prev1Hash := multiView.bestView.GetPreviousHash()
		if prev1Hash == nil {
			return
		}
		prev1View := multiView.viewByHash[*prev1Hash]
		if prev1View == nil {
			return
		}
		multiView.finalView = prev1View
	} else if newView.GetBlock().GetVersion() >= types.MULTI_VIEW_VERSION {
		////update finalView: consensus 2
		prev1Hash := multiView.bestView.GetPreviousHash()
		prev1View := multiView.viewByHash[*prev1Hash]
		if prev1View == nil || multiView.finalView.GetHeight() == prev1View.GetHeight() {
			return
		}
		bestViewTimeSlot := common.CalculateTimeSlot(multiView.bestView.GetBlock().GetProposeTime())
		prev1TimeSlot := common.CalculateTimeSlot(prev1View.GetBlock().GetProposeTime())
		if prev1TimeSlot+1 == bestViewTimeSlot { //three sequential time slot
			multiView.finalView = prev1View
			return
		}
		bestViewTimeSlot = common.CalculateTimeSlot(multiView.bestView.GetBlock().GetProduceTime())
		prev1TimeSlot = common.CalculateTimeSlot(prev1View.GetBlock().GetProduceTime())
		if prev1TimeSlot+1 == bestViewTimeSlot { //three sequential time slot
			multiView.finalView = prev1View
		}
	} else {
		fmt.Println("Block version is not correct")
	}

	//fmt.Println("Debug bestview", multiView.bestView.GetHeight())
	return
}

func (multiView *MultiView) GetAllViewsWithBFS() []types.View {
	queue := []types.View{multiView.finalView}
	resCh := make(chan []types.View)

	multiView.actionCh <- func() {
		res := []types.View{}
		for {
			if len(queue) == 0 {
				break
			}
			firstItem := queue[0]
			if firstItem == nil {
				break
			}
			for _, v := range multiView.viewByPrevHash[*firstItem.GetHash()] {
				queue = append(queue, v)
			}
			res = append(res, firstItem)
			queue = queue[1:]
		}
		resCh <- res
	}

	return <-resCh
}
