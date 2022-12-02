package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/pkg/errors"
)

func (b *BeaconCommitteeStateV4) InitReputationState() {
	Logger.log.Infof("[curtest] Init reputation")
	bc := b.GetBeaconCommittee()
	if bs := b.GetBeaconSubstitute(); len(bs) != 0 {
		bc = append(bc, bs...)
	}
	for _, v := range bc {
		bcPK, _ := v.ToBase58()
		b.Performance[bcPK] = 500
	}
}

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
		if curRep, ok := b.Performance[bPK]; ok {
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
			b.Performance[bPK] = curRep
		} else {
			return errors.Errorf("Can not found beacon public key %s in list %v, hold list %v", bPK, bCommittee, b.beaconCommittee)
		}
	}
	return nil
}
