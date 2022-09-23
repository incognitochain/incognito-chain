package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/pkg/errors"
)

// type BeaconReputationState struct {
// 	Reputation uint64
// 	TotalVotingPower
// }

func (b *BeaconCommitteeStateV4) UpdateBeaconReputationWithBlock(bBlock *types.BeaconBlock) error {
	prevVal, err := consensustypes.DecodeValidationData(bBlock.Header.PreviousValidationData)
	if err != nil {
		return err
	}
	bCommittee := b.beaconCommittee
	listVoted := prevVal.ValidatiorsIdx
	return b.updateBeaconReputation(bCommittee, listVoted)
}

func (b *BeaconCommitteeStateV4) updateBeaconReputation(bCommittee []string, listVotes []int) error {
	votedMap := map[int]interface{}{}
	for _, votedIdx := range listVotes {
		votedMap[votedIdx] = nil
	}
	for idx, bPK := range bCommittee {
		if curRep, ok := b.bReputation[bPK]; ok {
			if _, voted := votedMap[idx]; voted {
				curRep = curRep * 1015 / 1000
			} else {
				curRep = curRep * 965 / 1000
			}
			if curRep < 100 {
				curRep = 100
			}
			if curRep > 1000 {
				curRep = 1000
			}
			b.bReputation[bPK] = curRep
		} else {
			return errors.Errorf("Can not found beacon public key %s in list %v, hold list %v", bPK, bCommittee, b.beaconCommittee)
		}
	}
	return nil
}
