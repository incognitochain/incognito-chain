package fork

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/multiview"
)

type forkAction struct {
	delayTS   uint64
	multiView *multiview.MultiView
}

var forkMap = make(map[string]*forkAction)

/*/
Example fork beacon: when handle propose block, add these block of code :
if a.chain.IsBeaconChain() {
		stateRes := fork.ForkBeaconWithInstruction("forkBSC", a.chain.GetMultiView(), "250", a.chain.GetBestView().GetBlock().(*types.BeaconBlock), blockInfo.(*types.BeaconBlock), 4)
		//within fork TS
		if stateRes == 0 {
			return errors.New("simulate forkBSC")
		}
		//end fork TS -. reset bft + multiview
		if stateRes == 1 {
			a.receiveBlockByHash = make(map[string]*ProposeBlockInfo)
			a.receiveBlockByHeight = make(map[uint64][]*ProposeBlockInfo)
			a.voteHistory = make(map[uint64]types.BlockInterface)
			a.proposeHistory, err = lru.New(1000)
			a.chain.GetMultiView().ClearBranch()
			return errors.New("simulate forkBSC")
		}

	}
*/
func ForkBeaconWithInstruction(id string, mv *multiview.MultiView, instType string, newBlock *types.BeaconBlock, delayTS uint64) int {
	blk := mv.GetBestView().GetBlock()
	finalBllk := mv.GetFinalView().GetBlock()

	instruction := blk.GetInstructions()
	for _, v := range instruction {
		if v[0] == instType {
			fa := forkMap[id]
			if fa == nil {
				fa = &forkAction{
					delayTS:   delayTS,
					multiView: mv,
				}
			}

			currentDelayTs := (uint64(newBlock.GetProposeTime()) - uint64(finalBllk.GetProposeTime())) / common.TIMESLOT
			fmt.Println("debugfork:", currentDelayTs, fa.delayTS)
			if currentDelayTs < fa.delayTS {
				return 0
			}
			if currentDelayTs == fa.delayTS {
				fa.multiView.ClearBranch()
				return 1
			}
			// > delayTS, donothing
			return -1
		}
	}
	return -1
}
