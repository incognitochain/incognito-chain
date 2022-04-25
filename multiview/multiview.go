package multiview

import (
	"errors"
	"fmt"
	"log"
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
	CompareCommitteeFromBlock(View) int
}

type MultiView struct {
	viewByHash     map[common.Hash]View //viewByPrevHash map[common.Hash][]View
	viewByPrevHash map[common.Hash][]View
	actionCh       chan func()

	//state
	rootView  View //view that must not be revert
	finalView View //view at this time seen as final view (shardchain could revert from beacon view)
	bestView  View // best view from final view
}

func NewMultiView() *MultiView {
	s := &MultiView{
		viewByHash:     make(map[common.Hash]View),
		viewByPrevHash: make(map[common.Hash][]View),
		actionCh:       make(chan func()),
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
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
	res := make(chan *MultiView)
	multiView.actionCh <- func() {
		s := NewMultiView()
		for h, v := range multiView.viewByHash {
			s.viewByHash[h] = v
		}
		for h, v := range multiView.viewByPrevHash {
			s.viewByPrevHash[h] = v
		}
		s.finalView = multiView.finalView
		s.bestView = multiView.bestView
		res <- s
	}
	return <-res
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
		if v.GetHeight() < multiView.rootView.GetHeight() {
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
		if view == nil || view.GetHeight() < multiView.rootView.GetHeight() {
			res <- nil
		} else {
			res <- view
		}
	}
	return <-res
}

//forward root to specific view
//Instant finality: call after insert beacon block. Beacon chain will forward to final view. Shard chain will forward to shard view that beacon confirm
//need to check if it is still compatible with final&best view
func (multiView *MultiView) ForwardRoot(rootHash common.Hash) error {
	rootView, ok := multiView.viewByHash[rootHash]
	if !ok {
		return errors.New("Cannot find view by hash " + rootHash.String())
	}

	//recheck rootView is on the same view with final view
	notLink := true
	for {
		if rootView.GetHash().String() == multiView.finalView.GetHash().String() {
			notLink = false
			break
		}
		prevView := *multiView.finalView.GetPreviousHash()
		if _, ok := multiView.viewByHash[prevView]; !ok {
			break
		}
	}

	//if not link on the same branch with final view, reorg the multiview
	if notLink {
		newOrgMultiView := multiView.GetAllViewsWithBFS()
		multiView.Reset()
		for _, view := range newOrgMultiView {
			multiView.updateViewState(view)
		}
	}

	multiView.rootView = rootView
	return nil
}

func (multiView *MultiView) SimulateAddView(view View) (cloneMultiview *MultiView) {
	cloneMultiView := multiView.Clone()
	cloneMultiView.updateViewState(view)
	return cloneMultiView
}

//Only add view if view is validated (at least enough signature)
func (multiView *MultiView) AddView(view View) bool {
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

func (multiView *MultiView) GetBestView() View {
	return multiView.bestView
}

func (multiView *MultiView) GetFinalView() View {
	return multiView.finalView
}

//update view whenever there is new view insert into system
func (multiView *MultiView) updateViewState(newView View) {

	if multiView.finalView == nil {
		multiView.rootView = newView
		multiView.bestView = newView
		multiView.finalView = newView
		return
	}

	//update bestView
	if newView.GetHeight() > multiView.bestView.GetHeight() {
		multiView.bestView = newView
	}

	//get bestview with min produce time or better committee from block
	if newView.GetHeight() == multiView.bestView.GetHeight() {
		switch newView.CompareCommitteeFromBlock(multiView.bestView) {
		case 0:
			if newView.GetBlock().GetProduceTime() < multiView.bestView.GetBlock().GetProduceTime() {
				multiView.bestView = newView
			}
		case 1:
			multiView.bestView = newView
		case -1:
		}
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
	} else if newView.GetBlock().GetVersion() >= types.INSTANT_FINALITY_VERSION {
		////update finalView: consensus 2
		prev1Hash := multiView.bestView.GetPreviousHash()
		prev1View := multiView.viewByHash[*prev1Hash]

		//if no prev1View, return, something wrong, add view need to link to
		if prev1View == nil {
			log.Println("Previous view is nil, something wrong")
			return
		}

		bestViewTimeSlot := common.CalculateTimeSlot(multiView.bestView.GetBlock().GetProposeTime())
		prev1TimeSlot := common.CalculateTimeSlot(prev1View.GetBlock().GetProposeTime())
		if prev1TimeSlot+1 == bestViewTimeSlot { //two sequential time slot
			multiView.finalView = multiView.bestView
		}

		if multiView.bestView.GetBlock().GetFinalityHeight() != 0 { //this version, finality height mean this block having repropose proof of missing TS
			multiView.finalView = multiView.bestView
		}

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
		}
		if newView.GetBlock().GetVersion() >= types.LEMMA2_VERSION {
			// update final view lemma 2
			if newView.GetBlock().GetHeight()-1 == newView.GetBlock().GetFinalityHeight() {
				multiView.finalView = prev1View
			}
		}
	} else {
		fmt.Println("Block version is not correct")
	}

	//fmt.Println("Debug bestview", multiView.bestView.GetHeight())
	return
}

func (multiView *MultiView) GetAllViewsWithBFS() []View {
	queue := []View{multiView.rootView}
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
