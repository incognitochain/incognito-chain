package committeestate

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"math"
	"sort"
)

type CandidateInfoV1 struct {
	cpk         incognitokey.CommitteePublicKey
	cpkStr      string
	score       uint64
	currentRole string
}

//swap in to empty committee slot (new committee < 1/3 new total size)
//swap lowest score in committee (new committee < 1/3 new total size)
func beacon_swap_v1(pendingList []CandidateInfoV1, committeeList []CandidateInfoV1, numberOfFixNode int, committeeSlot int) ([]CandidateInfoV1, []CandidateInfoV1) {
	//sort candidate list
	sort.Slice(committeeList, func(i, j int) bool {
		return committeeList[i].score < committeeList[j].score
	})
	sort.Slice(pendingList, func(i, j int) bool {
		return pendingList[i].score > pendingList[j].score
	})

	//add to committeeSlot
	swapInVotingPower := func(candidates []CandidateInfoV1) (int, int) {
		swapIn := int(0)
		total := numberOfFixNode
		for _, c := range candidates {
			if c.currentRole == "pending" {
				swapIn++
			}
			total++
		}
		return swapIn, total
	}

	for j := 0; j < len(pendingList); j++ {
		swapIn, total := swapInVotingPower(append(committeeList, pendingList[j]))
		if len(committeeList) < committeeSlot && swapIn*3 < total {
			committeeList = append(committeeList, pendingList[j])        //append pending j
			pendingList = append(pendingList[0:j], pendingList[j+1:]...) //remove pending j
			j--
		}
	}
	//find pending candidate to replace committee with the smallest score
	if len(pendingList) > 0 {

		for j := 0; j < len(pendingList); j++ {
			//no candidate in committee
			if len(committeeList) == 0 {

				break
			}

			//if we swap all old committee list
			if committeeList[0].currentRole != "committee" {
				break
			}

			//if we swap all old pending list
			if pendingList[j].currentRole != "pending" {
				break
			}

			//if commitee[0] score is better the best pending candidate score
			if committeeList[0].score >= pendingList[j].score {
				break
			}

			//check we can swap
			swapIn, total := swapInVotingPower(append(committeeList[1:], pendingList[j]))
			if swapIn*3 < total {
				swapCommittee := committeeList[0]
				swapPending := pendingList[j]
				committeeList = append(committeeList[1:], swapPending)
				newPendingList := []CandidateInfoV1{}
				for k, p := range pendingList {
					if j != k {
						newPendingList = append(newPendingList, p)
					}
				}
				pendingList = append(newPendingList, swapCommittee)
				j--
			}
		}
	}

	sort.Slice(committeeList, func(i, j int) bool {
		return committeeList[i].score > committeeList[j].score
	})
	sort.Slice(pendingList, func(i, j int) bool {
		return pendingList[i].score > pendingList[j].score
	})

	return pendingList, committeeList
}

func (s *BeaconCommitteeStateV4) beacon_swap_v1(env ProcessContext) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey) {

	//swap pending <-> committee
	newBeaconCommittee := []incognitokey.CommitteePublicKey{}
	newBeaconPending := []incognitokey.CommitteePublicKey{}
	pendingList := []CandidateInfoV1{}
	for _, cpk := range s.GetBeaconSubstitute() {
		cpkString, _ := cpk.ToBase58()
		stakerInfo := s.getBeaconStakerInfo(cpkString)
		score := s.config.DEFAULT_PERFORMING * stakerInfo.StakingAmount
		pendingList = append(pendingList, CandidateInfoV1{stakerInfo.cpkStruct, cpkString, score, "pending"})

	}

	committeeList := []CandidateInfoV1{}
	fixNodeVotingPower := int64(0)
	for _, cpk := range s.GetBeaconCommittee() {
		cpkString, _ := cpk.ToBase58()
		stakerInfo := s.getBeaconStakerInfo(cpkString)
		score := stakerInfo.Performance * stakerInfo.StakingAmount
		if !stakerInfo.FixedNode {
			committeeList = append(committeeList, CandidateInfoV1{stakerInfo.cpkStruct, cpkString, score, "committee"})
		} else {
			newBeaconCommittee = append(newBeaconCommittee, stakerInfo.cpkStruct)
			fixNodeVotingPower += int64(math.Sqrt(float64(stakerInfo.StakingAmount)))
		}
	}
	fixNode := len(newBeaconCommittee)
	pendingList, committeeList = beacon_swap_v1(pendingList, committeeList, fixNode, s.config.BEACON_COMMITTEE_SIZE-len(newBeaconCommittee))

	//other candidate
	for _, candidate := range committeeList {
		newBeaconCommittee = append(newBeaconCommittee, candidate.cpk)
	}
	for _, candidate := range pendingList {
		newBeaconPending = append(newBeaconPending, candidate.cpk)
	}

	return newBeaconCommittee, newBeaconPending
}
