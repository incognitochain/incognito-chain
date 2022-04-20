package multiview

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
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
	CommitteeFromBlock() common.Hash
	GetBeaconHeight() uint64
	GetProposerByTimeSlot(ts int64, version int) (incognitokey.CommitteePublicKey, int)
	GetProposerLength() int
}

type BlockChain interface {
	GetBeaconBlockByHash(common.Hash) (*types.BeaconBlock, uint64, error)
}

// beacon BS and shard BS not share all the function
// maintain all old functions, only add new function, do not change interface because it impact many other package
// BBS: add view, finalize block = old finality rule, finalize block = new rule, get best confirmable chains, get confirmable chain
// SBS: add view, finalize block = old finality rule, finalize block = new rule
//

type MultiView struct {
	viewByHash     map[common.Hash]View
	viewByPrevHash map[common.Hash][]View
	finalityProof  map[common.Hash]*ReProposeProof
	actionCh       chan func()
	bc             BlockChain
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
	for h, v := range multiView.finalityProof {
		s.finalityProof[h] = v
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
		if v.GetHeight() < multiView.finalView.GetHeight() ||
			(v.GetBlock().GetVersion() == types.INSTANT_FINALITY_VERSION && v.GetHeight() == multiView.finalView.GetHeight()) {
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

func (multiView *MultiView) AddViewAndFinalizeV1(view View) (View, bool) {

	type result struct {
		v     View
		added bool
	}

	ch := make(chan result)

	multiView.actionCh <- func() {

		added := multiView.addView(view, nil)
		if !added {
			ch <- result{v: nil, added: false}
			return
		}
		ch <- result{v: multiView.finalizeViewV1(view), added: true}
	}

	res := <-ch
	return res.v, res.added
}

func (multiView *MultiView) SimulateAddViewV1(view View) (*MultiView, bool) {

	type result struct {
		v     *MultiView
		added bool
	}

	ch := make(chan result)

	multiView.actionCh <- func() {

		clonedView := multiView.Clone()

		added := clonedView.addView(view, nil)
		if !added {
			ch <- result{v: nil, added: false}
			return
		}

		clonedView.finalizeViewV1(view)

		ch <- result{v: clonedView, added: true}
		return
	}

	res := <-ch
	return res.v, res.added
}

//Only add view if view is validated (at least enough signature)
func (multiView *MultiView) addView(view View, proof *ReProposeProof) bool {

	if len(multiView.viewByHash) == 0 { //if no view in map, this is init view -> always allow
		multiView.viewByHash[*view.GetHash()] = view
		if view.GetBlock().GetVersion() == types.INSTANT_FINALITY_VERSION && proof != nil {
			multiView.finalityProof[*view.GetHash()] = proof
		}
		return true
	} else if _, ok := multiView.viewByHash[*view.GetHash()]; !ok { //otherwise, if view is not yet inserted
		if _, ok := multiView.viewByHash[*view.GetPreviousHash()]; ok { // view must point to previous valid view
			multiView.viewByHash[*view.GetHash()] = view
			multiView.viewByPrevHash[*view.GetPreviousHash()] = append(multiView.viewByPrevHash[*view.GetPreviousHash()], view)
			if view.GetBlock().GetVersion() == types.INSTANT_FINALITY_VERSION && proof != nil {
				multiView.finalityProof[*view.GetHash()] = proof
			}
			return true
		}
	}

	return false
}

func (multiView *MultiView) GetBestView() View {
	return multiView.bestView
}

func (multiView *MultiView) GetFinalView() View {
	return multiView.finalView
}

// finalizeViewV1 try to finalize a view, return new finalize view
// this view must be added to multiview before
func (multiView *MultiView) finalizeViewV1(newView View) View {

	defer func() {
		if multiView.viewByHash[*multiView.finalView.GetPreviousHash()] != nil {
			delete(multiView.viewByHash, *multiView.finalView.GetPreviousHash())
			delete(multiView.viewByPrevHash, *multiView.finalView.GetPreviousHash())
		}
	}()

	if multiView.finalView == nil {
		multiView.bestView = newView
		multiView.finalView = newView
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
		}
		if newView.GetBlock().GetVersion() >= types.LEMMA2_VERSION {
			// update final view lemma 2
			if newView.GetBlock().GetHeight()-1 == newView.GetBlock().GetFinalityHeight() {
				multiView.finalView = preView
			}
		}
	} else {
		fmt.Println("Block version is not correct")
	}

	return multiView.finalView
}

func (multiView *MultiView) BeaconSimulateAddViewV2(view View, proof *ReProposeProof) (*MultiView, error) {

	type result struct {
		m   *MultiView
		err error
	}

	ch := make(chan result)

	multiView.actionCh <- func() {
		cloneView := multiView.Clone()

		isAdded := cloneView.addView(view, proof)
		if !isAdded {
			ch <- result{m: nil, err: fmt.Errorf("add view hash %+v, height %+v failed", view.GetHash().String(), view.GetHeight())}
			return
		}
		_, err := cloneView.tryInstantFinalizeBeacon(view)
		if err != nil {
			ch <- result{m: nil, err: err}
			return
		} else {
			ch <- result{m: cloneView, err: nil}
			return
		}
	}

	res := <-ch

	return res.m, res.err
}

// TODO: @hung review handle add view when init view or restore view
func (multiView *MultiView) BeaconAddViewAndFinalizeV2(view View, proof *ReProposeProof) (View, error) {
	type result struct {
		v   View
		err error
	}

	ch := make(chan result)

	multiView.actionCh <- func() {
		if view.GetBlock().GetVersion() < types.INSTANT_FINALITY_VERSION {
			err := fmt.Errorf("Instant finality do not support this block version %+v", view.GetBlock().GetVersion())
			ch <- result{v: nil, err: err}
			return
		}
		multiView.addView(view, proof)
		v, err := multiView.tryInstantFinalizeBeacon(view)
		ch <- result{v: v, err: err}
		return
	}

	res := <-ch
	return res.v, res.err
}

func (multiView *MultiView) tryInstantFinalizeBeacon(newView View) (View, error) {

	if multiView.finalView == nil {
		multiView.bestView = newView
		multiView.finalView = newView
		return multiView.finalView, nil
	}

	//update bestView
	if newView.GetHeight() > multiView.bestView.GetHeight() {
		multiView.bestView = newView
	}

	//get bestview with min produce time
	if newView.GetHeight() == multiView.bestView.GetHeight() && newView.GetBlock().GetProduceTime() < multiView.bestView.GetBlock().GetProduceTime() {
		multiView.bestView = newView
	}

	childBlock := multiView.bestView.GetBlock()
	parentView := multiView.viewByHash[*multiView.bestView.GetPreviousHash()]
	parentBlock := parentView.GetBlock()
	proof, ok := multiView.finalityProof[*childBlock.Hash()]
	if !ok {
		proof = nil
	}

	res := isFirstProposeNextHeight(childBlock, parentBlock) || isValidReProposeBlock(childBlock, parentBlock, proof)
	if res {

		multiView.finalView = multiView.bestView

		deletedView := parentView
		for deletedView != nil {
			delete(multiView.viewByHash, *deletedView.GetHash())
			delete(multiView.viewByPrevHash, *deletedView.GetHash())
			delete(multiView.finalityProof, *deletedView.GetHash())
			tempView, ok := multiView.viewByHash[*deletedView.GetPreviousHash()]
			if !ok {
				break
			} else {
				deletedView = tempView
			}
		}

		return multiView.finalView, nil
	}

	return nil, nil
}

func (multiView *MultiView) ShardSimulateAddViewV2(view View, proof *ReProposeProof, finalizedIndex map[uint64]common.Hash) (*MultiView, error) {
	type result struct {
		m   *MultiView
		err error
	}

	ch := make(chan result)

	multiView.actionCh <- func() {
		cloneView := multiView.Clone()

		isAdded := cloneView.addView(view, proof)
		if !isAdded {
			ch <- result{m: nil, err: fmt.Errorf("add view hash %+v, height %+v failed", view.GetHash().String(), view.GetHeight())}
			return
		}
		err := cloneView.tryInstantFinalizeShard(finalizedIndex)
		if err != nil {
			ch <- result{m: nil, err: err}
			return
		} else {
			ch <- result{m: cloneView, err: nil}
			return
		}
	}

	res := <-ch

	return res.m, res.err
}
func (multiView *MultiView) ShardRestoreViewV2(view View) error {

	ch := make(chan error)
	multiView.actionCh <- func() {
		var err error
		added := multiView.addView(view, nil)
		if !added {
			ch <- fmt.Errorf("failed to add view %+v, %+v to multiview", view.GetHeight(), *view.GetHash())
			return
		}
		if multiView.finalView == nil {
			multiView.finalView = view
			multiView.bestView = view
		} else {
			multiView.bestView, err = multiView.calculateShardBestView()
		}
		ch <- err
	}

	return <-ch
}

// TODO: @hung review handle add view when init view or restore view
func (multiView *MultiView) ShardAddViewAndFinalizeV2(view View, proof *ReProposeProof, finalizedIndex map[uint64]common.Hash) (View, error) {

	type result struct {
		v   View
		err error
	}

	ch := make(chan result)

	multiView.actionCh <- func() {
		if view.GetBlock().GetVersion() < types.INSTANT_FINALITY_VERSION {
			err := errors.New("Instant finality do not support this view version")
			ch <- result{v: nil, err: err}
			return
		}

		multiView.addView(view, proof)
		err := multiView.tryInstantFinalizeShard(finalizedIndex)
		ch <- result{v: multiView.finalView, err: err}
	}

	res := <-ch
	return res.v, res.err
}

func (multiView *MultiView) tryInstantFinalizeShard(finalizedIndex map[uint64]common.Hash) error {

	finalHeight := multiView.finalView.GetHeight()
	deletedView := []common.Hash{}

	for i := 0; i < len(finalizedIndex); i++ {
		nextHeight := finalHeight + 1

		hash, ok := finalizedIndex[nextHeight]
		if !ok {
			break
		}
		finalView, ok := multiView.viewByHash[hash]
		if !ok {
			break
		}
		deletedView = append(deletedView, *finalView.GetPreviousHash())
		multiView.finalView = finalView

		finalHeight = nextHeight
	}

	for _, hash := range deletedView {
		delete(multiView.viewByHash, hash)
		delete(multiView.viewByPrevHash, hash)
		delete(multiView.finalityProof, hash)
	}

	bestView, err := multiView.calculateShardBestView()
	if err != nil {
		return err
	}
	multiView.bestView = bestView

	return nil
}

func (multiView *MultiView) calculateShardBestView() (View, error) {

	// get all chains from final view including the final view
	chains := [][]View{}
	getAllBranches(*multiView.finalView.GetHash(), multiView.viewByPrevHash, &chains, []View{multiView.finalView})

	//confirmableChains := [][]View{}
	//for _, v := range chains {
	//	_, isConfirmable := isConfirmableChain(v, multiView.finalityProof)
	//	if isConfirmable {
	//		confirmableChains = append(confirmableChains, v)
	//	}
	//}

	bestChain, err := multiView.shardForkChoiceRule(chains)
	if err != nil {
		return nil, err
	}

	return bestChain[len(bestChain)-1], nil
}

func (multiView *MultiView) GetBestConfirmableShardBranch() ([]types.BlockInterface, error) {

	type result struct {
		blocks []types.BlockInterface
		err    error
	}

	ch := make(chan result)

	multiView.actionCh <- func() {
		// get all branches from final view including the final view
		branches := [][]View{}
		getAllBranches(*multiView.finalView.GetHash(), multiView.viewByPrevHash, &branches, []View{multiView.finalView})

		confirmableChains := [][]View{}
		for _, v := range branches {
			finalizedIndex, isConfirmable := isConfirmableChain(v, multiView.finalityProof)
			if isConfirmable {
				confirmableChains = append(confirmableChains, v[:finalizedIndex+1])
			}
		}

		bestBranch, err := multiView.shardForkChoiceRule(confirmableChains)
		if err != nil {
			ch <- result{blocks: nil, err: err}
			return
		}

		res := []types.BlockInterface{}
		for i, v := range bestBranch {
			// discard final view
			if i == 0 {
				continue
			}
			res = append(res, v.GetBlock())
		}

		ch <- result{blocks: res, err: nil}
		return
	}

	res := <-ch

	return res.blocks, res.err
}

// shardForkChoiceRule output the best chain from multiple confirmable chain
// The longest chain (higher than any other chains)
// If chains share the same height
// 	choose chains with the newer committee
// 	if chains share the same committee
//		choose chains that have smaller producer time
func (multiView *MultiView) shardForkChoiceRule(confirmableBranches [][]View) ([]View, error) {

	if len(confirmableBranches) == 0 {
		return []View{}, errors.New("find no confirmable chain")
	}

	if len(confirmableBranches) == 1 {
		return confirmableBranches[0], nil
	}

	sameHeightBranches := [][]View{}
	maxHeight := uint64(0)

	for _, chain := range confirmableBranches {
		height := chain[len(chain)-1].GetHeight()
		if height > maxHeight {
			maxHeight = height
		}
	}

	for _, chain := range confirmableBranches {
		height := chain[len(chain)-1].GetHeight()
		if height == maxHeight {
			sameHeightBranches = append(sameHeightBranches, chain)
		}
	}

	if len(sameHeightBranches) == 1 {
		return sameHeightBranches[0], nil
	}

	sameCommitteeBranches := [][]View{}
	bestCommitteeHash := common.Hash{}
	bestBeaconHeight := uint64(0)

	for i := 0; i < len(sameHeightBranches); i++ {
		committeeHash := sameHeightBranches[i][len(sameHeightBranches)-1].CommitteeFromBlock()
		_, blockHeight, err := multiView.bc.GetBeaconBlockByHash(committeeHash)
		if err != nil {
			return []View{}, err
		}
		if blockHeight > bestBeaconHeight {
			bestBeaconHeight = blockHeight
			bestCommitteeHash = committeeHash
		}
	}

	for i := 0; i < len(sameHeightBranches); i++ {
		committeeHash := sameHeightBranches[i][len(sameHeightBranches)-1].CommitteeFromBlock()
		if bestCommitteeHash == committeeHash {
			sameCommitteeBranches = append(sameCommitteeBranches, sameHeightBranches[i])
		}
	}

	if len(sameCommitteeBranches) == 1 {
		return sameCommitteeBranches[0], nil
	}

	smallerProduceTimeBranches := 0
	minProduceTime := int64(math.MaxInt64)
	for i, v := range sameCommitteeBranches {
		if minProduceTime < v[len(v)-1].GetBlock().GetProduceTime() {
			smallerProduceTimeBranches = i
		}
	}

	return sameCommitteeBranches[smallerProduceTimeBranches], nil
}

func (multiView *MultiView) IsBelongToConfirmableChain(branch []types.BlockInterface) (bool, error) {

	if len(branch) == 0 {
		return false, errors.New("empty branch")
	}

	type result struct {
		confirmable bool
		reason      error
	}

	ch := make(chan result)

	multiView.actionCh <- func() {
		firstBlock := branch[0]
		if firstBlock.GetPrevHash() != *multiView.finalView.GetHash() {
			ch <- result{confirmable: false, reason: fmt.Errorf("first block %+v %+v, previous %+v, do not connect to final view %+v %+v",
				*firstBlock.Hash(), firstBlock.GetHeight(), firstBlock.GetPrevHash(), *multiView.finalView.GetHash(), multiView.finalView.GetHeight())}
			return
		}

		childBlock := branch[len(branch)-1]
		var parentBlock types.BlockInterface
		if len(branch) == 1 {
			parentBlock = multiView.finalView.GetBlock()
		} else {
			parentBlock = branch[len(branch)-2]
			for i := len(branch) - 1; i >= 1; i-- {
				if branch[i].GetPrevHash() != *branch[i-1].Hash() {
					ch <- result{confirmable: false, reason: fmt.Errorf("block %+v %+v, previous %+v, do not connect to block %+v %+v",
						*branch[i].Hash(), branch[i].GetHeight(), branch[i].GetPrevHash(), *branch[i-1].Hash(), branch[i-1].GetHeight())}
				}
			}
		}

		proof, ok := multiView.finalityProof[*childBlock.Hash()]
		if !ok {
			proof = nil
		}

		isValid := isFirstProposeNextHeight(childBlock, parentBlock) || isValidReProposeBlock(childBlock, parentBlock, proof)
		if isValid {
			ch <- result{confirmable: true, reason: nil}
			return
		}

		// this branch must belong to a finalized branch
		// calculate if the chain from child block to the latest view contain at least a finalized chain
		v := multiView.viewByHash[*childBlock.Hash()]
		branches := [][]View{}
		getAllBranches(*childBlock.Hash(), multiView.viewByPrevHash, &branches, []View{v})
		for _, branch := range branches {
			_, confirm := isConfirmableChain(branch, multiView.finalityProof)
			if confirm {
				ch <- result{confirmable: true, reason: nil}
				return
			}
		}

		ch <- result{confirmable: false, reason: errors.New("branch do not belong to any confirmable chain")}
		return
	}

	res := <-ch
	return res.confirmable, res.reason
}

// isConfirmableChain iterator from the last element of the list to find to confirmable view
func isConfirmableChain(chain []View, proof map[common.Hash]*ReProposeProof) (int, bool) {

	if len(chain) == 1 || len(chain) == 0 {
		return -1, false
	}

	for i := len(chain) - 1; i >= 1; i-- {
		childBlock := chain[i].GetBlock()
		parentBlock := chain[i-1].GetBlock()
		isValid := isFirstProposeNextHeight(childBlock, parentBlock)
		if isValid {
			return i, true
		}
		if proof, ok := proof[*childBlock.Hash()]; ok {
			isValid := isValidReProposeBlock(childBlock, parentBlock, proof)
			if isValid {
				return i, true
			}
		}
	}

	return -1, false
}

// isFirstProposeNextHeight check if block is first propose next height, satisfy all three condition
// ProduceTimeSlot - PreviousProposeTimeSlot = 1
// Producer = Proposer
// ProduceTimeSlot = ProposeTimeSlot
func isFirstProposeNextHeight(childBlock, parentBlock types.BlockInterface) bool {

	childProduceTimeSlot := common.CalculateTimeSlot(childBlock.GetProduceTime())
	childProposeTimeSlot := common.CalculateTimeSlot(childBlock.GetProposeTime())
	parentProduceTimeSlot := common.CalculateTimeSlot(parentBlock.GetProduceTime())

	if childProduceTimeSlot-parentProduceTimeSlot != 1 {
		return false
	}

	if childProposeTimeSlot != childProduceTimeSlot {
		return false
	}

	if childBlock.GetProposer() != childBlock.GetProducer() {
		return false
	}

	return true
}

// isValidReProposeBlock check if block is first propose next height, satisfy all four condition
// ProduceTimeSlot - PreviousProposeTimeSlot = 1
// Verify proposer index with producer + gap timeslot
// Verify re-propose hash signature
// Get Finality Proof and Verify Proof corresponding to the gap between ProposeTimeSlot and PropduceTimeSlot
func isValidReProposeBlock(childBlock, parentBlock types.BlockInterface, reProposeProof *ReProposeProof) bool {

	if reProposeProof == nil {
		return false
	}

	if reProposeProof.cacheValid {
		return true
	}

	childProduceTimeSlot := common.CalculateTimeSlot(childBlock.GetProduceTime())
	parentProduceTimeSlot := common.CalculateTimeSlot(parentBlock.GetProduceTime())

	if childProduceTimeSlot-parentProduceTimeSlot != 1 {
		return false
	}

	err := reProposeProof.Proof.Verify(
		*parentBlock.Hash(),
		childBlock.GetProducer(),
		childProduceTimeSlot,
		reProposeProof.Proposer,
		childBlock.GetAggregateRootHash(),
	)

	return err == nil
}

// getAllBranches by using depth first search to find all path from final view, consider views are directed acyclic graph
func getAllBranches(curView common.Hash, views map[common.Hash][]View, chains *[][]View, curChain []View) {

	childViews, exist := views[curView]
	// end of the tree
	if !exist {
		tempChain := make([]View, len(curChain))
		copy(tempChain, curChain)
		*chains = append(*chains, tempChain)
		return
	}
	// continue to traverse
	for _, childView := range childViews {
		curChain = append(curChain, childView)
		getAllBranches(*childView.GetHash(), views, chains, curChain)
		curChain = curChain[:len(curChain)-1]
	}

	return
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
