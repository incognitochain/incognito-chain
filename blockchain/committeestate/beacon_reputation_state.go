package committeestate

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/pkg/errors"
)

type BKPerf struct {
	H    uint64
	CPks []string
	Perf []uint64
}

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

func (b *BeaconCommitteeStateV4) UpdateBeaconPerformanceWithValidationData(valData string) error {
	prevVal, err := consensustypes.DecodeValidationData(valData)
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

func (b *BeaconCommitteeStateV4) BackupPerformance(blkHeight uint64) []byte {
	perfs := []uint64{}
	for _, cPK := range b.beaconCommittee {
		perf, ok := b.Performance[cPK]
		if !ok {
			perf = 500
		}
		perfs = append(perfs, perf)
	}
	bkPerf := BKPerf{
		H:    blkHeight,
		Perf: perfs,
		CPks: b.beaconCommittee,
	}
	res, err := json.Marshal(bkPerf)
	if err != nil {
		panic(err)
	}
	return res
}

func (b *BeaconCommitteeStateV4) RestorePerformance(data []byte, beaconBlocks []types.BeaconBlock) {
	bkPerf := BKPerf{}
	for _, cPK := range b.beaconWaiting {
		b.Performance[cPK] = 500
	}
	for _, cPK := range b.beaconSubstitute {
		b.Performance[cPK] = 500
	}
	if len(data) != 0 {
		err := json.Unmarshal(data, &bkPerf)
		if err != nil {
			panic(err)
		}
		for idx, cPK := range bkPerf.CPks {
			b.Performance[cPK] = bkPerf.Perf[idx]
		}
	}
	for _, blk := range beaconBlocks {
		if blk.Header.Height <= config.Param().ConsensusParam.StakingFlowV4Height {
			continue
		}
		if blk.Header.Height == bkPerf.H {
			break
		}
		err := b.UpdateBeaconPerformanceWithValidationData(blk.Header.PreviousValidationData)
		if err != nil {
			panic(err)
		}
	}
	for pk, _ := range b.bDelegateState.DelegateInfo {
		perf := uint64(0)
		if v, ok := b.Performance[pk]; ok {
			perf = v
		}
		vpow := b.bDelegateState.GetBeaconCandidatePower(pk)
		b.Reputation[pk] = perf * vpow / 1000
	}

}
