package multiview

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

type View interface {
	GetHash() *common.Hash
	GetPreviousHash() *common.Hash
	GetHeight() uint64
	//GetTimeSlot() uint64
	GetBlockTime() int64
}

type MultiView struct {
	viewByHash     map[common.Hash]View //viewByPrevHash map[common.Hash][]View
	viewByPrevHash map[common.Hash][]View
	actionCh       chan func()

	//state
	finalView View
	bestView  View
}

func NewMultiView() *MultiView {
	s := &MultiView{
		viewByHash:     make(map[common.Hash]View),
		viewByPrevHash: make(map[common.Hash][]View),
		actionCh:       make(chan func()),
	}

	go func() {
		for {
			f := <-s.actionCh
			f()
		}
	}()

	return s

}

//Only add view if view is validated (at least enough signature)
func (multiView *MultiView) AddView(view View) bool {
	res := make(chan bool)
	multiView.actionCh <- func() {
		if len(multiView.viewByHash) == 0 { //if no view in map, this is init view -> always allow
			multiView.viewByHash[*view.GetHash()] = view
			multiView.viewByPrevHash[*view.GetPreviousHash()] = append(multiView.viewByPrevHash[*view.GetPreviousHash()], view)
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

func (multiView *MultiView) GetBestView() View {
	return multiView.bestView
}

func (multiView *MultiView) GetFinalView() View {
	return multiView.finalView
}

//update view whenever there is new view insert into system
func (multiView *MultiView) updateViewState(newView View) {
	//update bestView
	if multiView.bestView == nil || newView.GetHeight() > multiView.bestView.GetHeight() {
		multiView.bestView = newView
	}
	if multiView.finalView == nil || newView.GetHeight() == multiView.bestView.GetHeight() && newView.GetBlockTime() < multiView.bestView.GetBlockTime() {
		multiView.bestView = newView
	}

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

	////update finalView: consensus 2
	//prev1Hash := multiView.bestView.GetPreviousHash()
	//prev1View := multiView.viewByHash[*prev1Hash]
	//if prev1View == nil {
	//	return
	//}
	//
	//prev2Hash := prev1View.GetPreviousHash()
	//prev2View := multiView.viewByHash[*prev2Hash]
	//if prev2View == nil {
	//	return
	//}
	//
	//if prev1View.GetTimeSlot()+1 == multiView.bestView.GetTimeSlot() && prev2View.GetTimeSlot()+2 == multiView.bestView.GetTimeSlot() {
	//	multiView.finalView = prev2View
	//}
	fmt.Println("Debug bestview", multiView.bestView.GetHeight())
	return
}

//update view when
//func (multiView *MultiView) SerializeViews() []byte {
//
//}
//
//func (multiView *MultiView) DeSerializeViews() []byte {
//
//}

//
func (multiView *MultiView) GetAllViewsWithBFS() []View {
	queue := []View{multiView.finalView}
	res := []View{}
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
	return res
}
