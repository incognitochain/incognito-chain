package committeestate

import (
	"encoding/json"

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
	if bw := b.GetBeaconWaiting(); len(bw) != 0 {
		bc = append(bc, bw...)
	}
	b.Reputation = map[string]uint64{}
	for _, v := range bc {
		bcPK, _ := v.ToBase58()
		b.Performance[bcPK] = 500
	}
}

func (b *BeaconCommitteeStateV4) UpdateBeaconPerformanceWithBlock(bBlock *types.BeaconBlock) error {
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

func (b *BeaconCommitteeStateV4) BackupPerformance() []byte {
	perfs := []uint64{}
	for _, cPK := range b.beaconCommittee {
		perf, ok := b.Performance[cPK]
		if !ok {
			perf = 500
		}
		perfs = append(perfs, perf)
	}
	res, err := json.Marshal(perfs)
	if err != nil {
		panic(err)
	}
	return res
}

func (b *BeaconCommitteeStateV4) RestorePerformance(data []byte, beaconBlocks []*types.BeaconBlock) {
	perfs := []uint64{}
	err := json.Unmarshal(data, &perfs)
	if err != nil {
		panic(err)
	}
	for idx, cPK := range b.beaconCommittee {
		b.Performance[cPK] = perfs[idx]
	}
	for _, cPK := range b.beaconWaiting {
		b.Performance[cPK] = 500
	}
	for _, cPK := range b.beaconSubstitute {
		b.Performance[cPK] = 500
	}
	for _, blk := range beaconBlocks {
		err := b.UpdateBeaconPerformanceWithBlock(blk)
		if err != nil {
			panic(err)
		}
	}
}
