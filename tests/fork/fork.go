package fork

import (
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
Example fork beacon: at the end of creating new beacon block, insert :
if fork.ForkBeaconWithInstruction("forkBSC", blockchain.BeaconChain.multiView, "250", curView.BestBlock, *newBeaconBlock, 4) {
	return nil, errors.New("simulate forkBSC")
}
*/
func ForkBeaconWithInstruction(id string, mv *multiview.MultiView, instType string, blk types.BeaconBlock, newBlock types.BeaconBlock, delayTS uint64) bool {
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

			currentDelayTs := (uint64(newBlock.GetProposeTime()) - uint64(blk.GetProposeTime())) / common.TIMESLOT
			if currentDelayTs < fa.delayTS {
				return true
			}
			if currentDelayTs == fa.delayTS {
				fa.multiView.ClearBranch()
				return true
			}
			// > delayTS, donothing
			return false
		}
	}
	return false
}
