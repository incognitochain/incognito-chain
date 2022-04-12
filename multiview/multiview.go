package multiview

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incdb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type View interface {
	GetHash() *common.Hash
	GetPreviousHash() *common.Hash
	GetHeight() uint64
	GetCommittee() []incognitokey.CommitteePublicKey
	GetPreviousBlockCommittee(db incdb.Database) ([]incognitokey.CommitteePublicKey, error)
	CommitteeStateVersion() int
	GetBlock() types.BlockInterface
	GetBeaconHeight() uint64
	GetProposerByTimeSlot(ts int64, version int) (incognitokey.CommitteePublicKey, int)
	GetProposerLength() int
}

// beacon BS and shard BS not share all the function
// maintain all old functions, only add new function, do not change interface because it impact many other package
// BBS: add view, finalize block = old finality rule, finalize block = new rule, get best confirmable chains, get confirmable chain
// SBS: add view, finalize block = old finality rule, finalize block = new rule
//

type MultiView struct {
	viewByHash     map[common.Hash]View
	viewByPrevHash map[common.Hash][]View
	actionCh       chan func()

	//state
	finalView View
	bestView  View
}

type Context struct {
}

type IMultiView interface {
	// add a new view to a tree
	AddView(View) // same with all
	// get best view => Rule
	// 1. View with max depth
	// 2. View with equal depth but smallest block producing time
	// 3. View with equal depth but newer committee
	GetBestView() // same with all
	// Capture Final View
	GetFinalView() // same with all
	// Finalize view based on version and shard or beacon
	TryFinalizeView(View) View // 3 version now based on shard or beacon
	// Use breadth first search rule
	GetAllViewsWithBFS() // same with all
}

func NewMultiView() *MultiView {
	s := &MultiView{
		viewByHash:     make(map[common.Hash]View),
		viewByPrevHash: make(map[common.Hash][]View),
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

func (multiView *MultiView) Reset() {
	multiView.viewByHash = make(map[common.Hash]View)
	multiView.viewByPrevHash = make(map[common.Hash][]View)
}

func (multiView *MultiView) ClearBranch() {
	multiView.bestView = multiView.finalView
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

func (multiView *MultiView) GetViewByHash(hash common.Hash) View {
	res := make(chan View)
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
func (multiView *MultiView) AddView(view View) bool {
	res := make(chan bool)
	multiView.actionCh <- func() {
		if len(multiView.viewByHash) == 0 { //if no view in map, this is init view -> always allow
			multiView.viewByHash[*view.GetHash()] = view
			res <- true
			return
		} else if _, ok := multiView.viewByHash[*view.GetHash()]; !ok { //otherwise, if view is not yet inserted
			if _, ok := multiView.viewByHash[*view.GetPreviousHash()]; ok { // view must point to previous valid view
				multiView.viewByHash[*view.GetHash()] = view
				multiView.viewByPrevHash[*view.GetPreviousHash()] = append(multiView.viewByPrevHash[*view.GetPreviousHash()], view)
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

// TryFinalizeView try to finalize a view, return new finalize view
// this view must be added to multiview before
func (multiView *MultiView) TryFinalizeView(newView View) View {

	defer func() {
		if multiView.viewByHash[*multiView.finalView.GetPreviousHash()] != nil {
			delete(multiView.viewByHash, *multiView.finalView.GetPreviousHash())
			delete(multiView.viewByPrevHash, *multiView.finalView.GetPreviousHash())
		}
	}()

	if multiView.finalView == nil {
		multiView.bestView = newView
		multiView.finalView = newView
		return multiView.finalView
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
			return nil
		}
		prev1View := multiView.viewByHash[*prev1Hash]
		if prev1View == nil {
			return nil
		}

		multiView.finalView = prev1View
		return multiView.finalView

	} else if newView.GetBlock().GetVersion() >= types.MULTI_VIEW_VERSION && newView.GetBlock().GetVersion() <= types.BLOCK_PRODUCINGV3_VERSION {
		// update finalView: consensus 2
		preHash := multiView.bestView.GetPreviousHash()
		preView := multiView.viewByHash[*preHash]
		if preView == nil || multiView.finalView.GetHeight() == preView.GetHeight() {
			return nil
		}
		bestViewTimeSlot := common.CalculateTimeSlot(multiView.bestView.GetBlock().GetProposeTime())
		prev1TimeSlot := common.CalculateTimeSlot(preView.GetBlock().GetProposeTime())
		if prev1TimeSlot+1 == bestViewTimeSlot { //three sequential time slot
			multiView.finalView = preView
			return multiView.finalView
		}
		if newView.GetBlock().GetVersion() >= types.LEMMA2_VERSION {
			// update final view lemma 2
			if newView.GetBlock().GetHeight()-1 == newView.GetBlock().GetFinalityHeight() {
				multiView.finalView = preView
				return multiView.finalView
			}
		}
	} else {
		fmt.Println("Block version is not correct")
	}

	return nil
}

func (multiView *MultiView) BeaconInstantFinalize(view View, reProposeHash []string, committees []incognitokey.CommitteePublicKey) {

}

func (multiView *MultiView) GetAllViewsWithBFS() []View {
	queue := []View{multiView.finalView}
	resCh := make(chan []View)

	multiView.actionCh <- func() {
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
		resCh <- res
	}

	return <-resCh
}
