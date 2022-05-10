package multiview

import (
	"errors"
	"fmt"
	"sync"
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

type MultiView interface {
	GetViewByHash(hash common.Hash) View
	SimulateAddView(view View) (cloneMultiview MultiView)
	GetBestView() View
	GetFinalView() View
	FinalizeView(hashToFinalize common.Hash) error
	GetExpectedFinalView() View
	GetAllViewsWithBFS() []View
	RunCleanProcess()
	Clone() MultiView
	Reset()
	AddView(v View) (int, error)
}

type multiView struct {
	viewByHash     map[common.Hash]View //viewByPrevHash map[common.Hash][]View
	viewByPrevHash map[common.Hash][]View
	lock           *sync.RWMutex

	//state
	finalView         View //view that must not be revert
	expectedFinalView View //view at this time seen as final view (shardchain could revert from beacon view)
	bestView          View // best view from final view
}

func NewMultiView() *multiView {
	s := &multiView{
		viewByHash:     make(map[common.Hash]View),
		viewByPrevHash: make(map[common.Hash][]View),
		lock:           new(sync.RWMutex),
	}

	return s
}

func (s *multiView) RunCleanProcess() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				if len(s.viewByHash) > 1 {
					s.removeOutdatedView()
				}
			}
		}
	}()
}

func (s *multiView) Clone() MultiView {
	s.lock.RLock()
	defer s.lock.RUnlock()
	cloneMV := NewMultiView()
	for h, v := range s.viewByHash {
		cloneMV.viewByHash[h] = v
	}
	for h, v := range s.viewByPrevHash {
		cloneMV.viewByPrevHash[h] = v
	}
	cloneMV.expectedFinalView = s.expectedFinalView
	cloneMV.bestView = s.bestView
	cloneMV.finalView = s.finalView
	return cloneMV
}

func (s *multiView) Reset() {
	s.viewByHash = make(map[common.Hash]View)
	s.viewByPrevHash = make(map[common.Hash][]View)
}

func (s *multiView) ClearBranch() {
	s.bestView = s.expectedFinalView
}

func (s *multiView) removeOutdatedView() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for h, v := range s.viewByHash {
		//buffer 1 views in the mem, so that GetPreviousHash work for bestview (also is final view)
		if v.GetHeight() < s.finalView.GetHeight()-1 {
			delete(s.viewByHash, h)
			delete(s.viewByPrevHash, h)
			delete(s.viewByPrevHash, *v.GetPreviousHash())
		}
	}
}

func (s *multiView) GetViewByHash(hash common.Hash) View {
	s.lock.RLock()
	defer s.lock.RUnlock()
	view, _ := s.viewByHash[hash]
	if view == nil || view.GetHeight() < s.finalView.GetHeight() {
		return nil
	} else {
		return view
	}

}

//forward final view to specific view
//Instant finality: Beacon chain will forward to expected final view. Shard chain will forward to shard view that beacon confirm
//need to check if it is still compatible with best view
func (s *multiView) FinalizeView(hashToFinalize common.Hash) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	viewToFinalize, ok := s.viewByHash[hashToFinalize]
	if !ok {
		return errors.New("Cannot find view by hash " + hashToFinalize.String())
	}

	//recheck hashToFinalize is on the same branch with bestview view
	notLink := true
	prevView := s.bestView
	for {
		if hashToFinalize.String() == prevView.GetHash().String() {
			notLink = false
			break
		}
		prevView = s.viewByHash[*prevView.GetPreviousHash()]
		if prevView == nil {
			break
		}
	}

	//if not link on the same branch with final view, reorg the multiview
	//this must be not happen with our flow
	if notLink {

		panic("This must not happen!")
		newOrgMultiView := s.getAllViewsWithBFS(viewToFinalize)
		s.Reset()
		for _, view := range newOrgMultiView {
			s.updateViewState(view)
		}
	} else {
		//if current final view is less than the specified view
		if s.finalView.GetHeight() < viewToFinalize.GetHeight() {
			s.finalView = viewToFinalize
		}

	}

	return nil
}

func (s *multiView) SimulateAddView(view View) (cloneMultiview MultiView) {
	cloneMultiView := s.Clone()
	cloneMultiView.(*multiView).updateViewState(view)
	return cloneMultiView
}

//Only add view if view is validated (at least enough signature)
func (s *multiView) addView(view View) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.viewByHash) == 0 { //if no view in map, this is init view -> always allow
		s.viewByHash[*view.GetHash()] = view
		s.updateViewState(view)
		return true
	} else if _, ok := s.viewByHash[*view.GetHash()]; !ok { //otherwise, if view is not yet inserted
		if _, ok := s.viewByHash[*view.GetPreviousHash()]; ok { // view must point to previous valid view
			s.viewByHash[*view.GetHash()] = view
			s.viewByPrevHash[*view.GetPreviousHash()] = append(s.viewByPrevHash[*view.GetPreviousHash()], view)
			s.updateViewState(view)

			return true
		}
	}
	return false

}

func (s *multiView) GetBestView() View {
	return s.bestView
}

func (s *multiView) GetFinalView() View {
	return s.finalView
}

func (s *multiView) GetExpectedFinalView() View {
	return s.expectedFinalView
}

//update view whenever there is new view insert into system
func (s *multiView) updateViewState(newView View) {

	if s.expectedFinalView == nil {
		s.finalView = newView
		s.bestView = newView
		s.expectedFinalView = newView
		return
	}

	//update bestView
	if newView.GetHeight() > s.bestView.GetHeight() {
		s.bestView = newView
	}

	//get bestview with min produce time or better committee from block
	if newView.GetHeight() == s.bestView.GetHeight() {
		switch newView.CompareCommitteeFromBlock(s.bestView) {
		case 0:
			if newView.GetBlock().GetProduceTime() < s.bestView.GetBlock().GetProduceTime() {
				s.bestView = newView
			}
		case 1:
			s.bestView = newView
		case -1:
		}
	}

	if newView.GetBlock().GetVersion() == types.BFT_VERSION {
		//update expectedFinalView: consensus 1
		prev1Hash := s.bestView.GetPreviousHash()
		if prev1Hash == nil {
			return
		}
		prev1View := s.viewByHash[*prev1Hash]
		if prev1View == nil {
			return
		}
		s.expectedFinalView = prev1View
	} else if newView.GetBlock().GetVersion() >= types.INSTANT_FINALITY_VERSION {
		//we traverse backward to update expected final view (in case bestview change branch)
		currentViewPoint := s.bestView
		for {
			prev1Hash := currentViewPoint.GetPreviousHash()
			prev1View := s.viewByHash[*prev1Hash]

			if prev1View == nil {
				s.expectedFinalView = currentViewPoint
				return
			}

			bestViewTimeSlot := common.CalculateTimeSlot(currentViewPoint.GetBlock().GetProposeTime())
			prev1TimeSlot := common.CalculateTimeSlot(prev1View.GetBlock().GetProposeTime())

			if prev1TimeSlot+1 == bestViewTimeSlot { //two sequential time slot
				s.expectedFinalView = currentViewPoint
				break
			} else if currentViewPoint.GetBlock().GetFinalityHeight() != 0 { //this version, finality height mean this block having repropose proof of missing TS
				s.expectedFinalView = currentViewPoint
				break
			}
			currentViewPoint = prev1View
		}

	} else if newView.GetBlock().GetVersion() >= types.MULTI_VIEW_VERSION {
		////update expectedFinalView: consensus 2
		prev1Hash := s.bestView.GetPreviousHash()
		prev1View := s.viewByHash[*prev1Hash]
		if prev1View == nil || s.expectedFinalView.GetHeight() == prev1View.GetHeight() {
			return
		}
		bestViewTimeSlot := common.CalculateTimeSlot(s.bestView.GetBlock().GetProposeTime())
		prev1TimeSlot := common.CalculateTimeSlot(prev1View.GetBlock().GetProposeTime())
		if prev1TimeSlot+1 == bestViewTimeSlot { //three sequential time slot
			s.expectedFinalView = prev1View
		}
		if newView.GetBlock().GetVersion() >= types.LEMMA2_VERSION {
			// update final view lemma 2
			if newView.GetBlock().GetHeight()-1 == newView.GetBlock().GetFinalityHeight() {
				s.expectedFinalView = prev1View
			}
		}
	} else {
		fmt.Println("Block version is not correct")
	}

	//fmt.Println("Debug bestview", s.bestView.GetHeight())
	return
}

func (s *multiView) getAllViewsWithBFS(rootView View) []View {
	queue := []View{rootView}

	s.lock.RLock()
	defer s.lock.RUnlock()

	res := []View{}
	for {
		if len(queue) == 0 {
			break
		}
		firstItem := queue[0]
		if firstItem == nil {
			break
		}
		for _, v := range s.viewByPrevHash[*firstItem.GetHash()] {
			queue = append(queue, v)
		}
		res = append(res, firstItem)
		queue = queue[1:]
	}
	return res
}

func (s *multiView) GetAllViewsWithBFS() []View {
	return s.getAllViewsWithBFS(s.finalView)
}

//this is just for interface compatible, we dont expect running this function
func (s *multiView) AddView(v View) (int, error) {
	panic("must not use this")
}
