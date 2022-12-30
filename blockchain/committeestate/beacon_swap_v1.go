package committeestate

import (
	"math"
	"sort"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type CandidateInfo struct {
	cpk         incognitokey.CommitteePublicKey
	cpkStr      string
	score       uint64
	votingPower int64
	currentRole string
}

func (s *BeaconCommitteeStateV4) beaconRemoveAndSwap(env *BeaconCommitteeStateEnvironment) (
	map[string]bool, map[string]bool,
	map[string]incognitokey.CommitteePublicKey, map[string]incognitokey.CommitteePublicKey,
	error) {

	slashCpk, unstakeCpk, committeeList, pendingList := beaconRemove(s.config.MIN_PERFORMANCE, s.config.DEFAULT_PERFORMING, s.beaconCommittee, s.beaconPending, s.beaconWaiting)

	//swap pending <-> committee
	newBeaconCommittee := map[string]incognitokey.CommitteePublicKey{}
	newBeaconPending := map[string]incognitokey.CommitteePublicKey{}
	fixNodeVotingPower := int64(0)
	for cpk, stakerInfo := range s.beaconCommittee {
		if stakerInfo.FixedNode {
			newBeaconCommittee[cpk] = stakerInfo.cpkStr
			fixNodeVotingPower += int64(math.Sqrt(float64(stakerInfo.StakingAmount)))
		}

	}

	pendingList, committeeList = beaconSwap(pendingList, committeeList, fixNodeVotingPower, env.MaxBeaconCommitteeSize-len(newBeaconCommittee))
	//other candidate
	for _, candidate := range committeeList {
		newBeaconCommittee[candidate.cpkStr] = candidate.cpk
	}
	for _, candidate := range pendingList {
		newBeaconPending[candidate.cpkStr] = candidate.cpk
	}
	return slashCpk, unstakeCpk, newBeaconCommittee, newBeaconPending, nil
}

func fillCandidate(
	pendingList []string,
	pendingVotePower []int64,
	sumCommitteeVotePower int64,
	totalSwapInSlot int64,
) (
	swapIn map[string]interface{},
	swapInVotePower int64,
) {
	swapIn = map[string]interface{}{}
	swapInVotePower = int64(0)
	for pI, pPK := range pendingList {
		totalVotePowerTmp := sumCommitteeVotePower + swapInVotePower + pendingVotePower[pI]
		if sumCommitteeVotePower > totalVotePowerTmp*2/3 {
			swapIn[pPK] = nil
			swapInVotePower += pendingVotePower[pI]
		}
	}
	return swapIn, swapInVotePower
}

func swapCandidate(
	pendingList []string,
	pendingVotePower []int64,
	pendingScore []uint64,
	committeeList []string,
	sumCommitteeVotePower []int64,
	committeeScore []uint64,
	swapInVotePower int64,
) (
	swapIn map[string]interface{},
	swapOut map[string]interface{},
) {
	swapIn = map[string]interface{}{}
	swapOut = map[string]interface{}{}
	cI := 0
	for pI, pPK := range pendingList {
		if cI < len(committeeList) {
			if pendingScore[pI] > committeeScore[cI] {
				totalVotePowerTmp := sumCommitteeVotePower[cI+1] + swapInVotePower + pendingVotePower[pI]
				if sumCommitteeVotePower[cI+1] > totalVotePowerTmp*2/3 {
					swapIn[pPK] = nil
					swapOut[committeeList[cI]] = nil
					swapInVotePower += pendingVotePower[pI]
				}
			}
		} else {
			return
		}
	}
	return swapIn, swapOut
}

func beaconRemove(
	cfgMinPerf uint64,
	cfgDefPerf uint64,
	beaconCommittee map[string]*StakerInfo,
	beaconPending map[string]*StakerInfo,
	beaconWaiting map[string]*StakerInfo,
) (
	map[string]bool,
	map[string]bool,
	[]CandidateInfo,
	[]CandidateInfo,
) {
	// slash
	slashCpk := map[string]bool{}
	for cpk, stakerInfo := range beaconCommittee {
		if stakerInfo.Performance < cfgMinPerf && !stakerInfo.FixedNode {
			slashCpk[cpk] = true
		}
	}

	// unstake
	unstakeCpk := map[string]bool{}
	for cpk, stakerInfo := range beaconCommittee {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = true
		}
	}
	for cpk, stakerInfo := range beaconPending {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = true
		}
	}
	for cpk, stakerInfo := range beaconWaiting {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = true
		}
	}
	pendingList := []CandidateInfo{}
	for cpk, stakerInfo := range beaconPending {
		if !slashCpk[cpk] && !unstakeCpk[cpk] {
			score := cfgDefPerf * stakerInfo.StakingAmount
			pendingList = append(pendingList, CandidateInfo{stakerInfo.cpkStr, cpk, score, int64(math.Sqrt(float64(stakerInfo.StakingAmount))), "pending"})
		}
	}

	committeeList := []CandidateInfo{}
	for cpk, stakerInfo := range beaconCommittee {
		if !slashCpk[cpk] && !unstakeCpk[cpk] {
			score := stakerInfo.Performance * stakerInfo.StakingAmount
			if !stakerInfo.FixedNode {
				committeeList = append(committeeList, CandidateInfo{stakerInfo.cpkStr, cpk, score, int64(math.Floor(math.Sqrt(float64(stakerInfo.StakingAmount)))), "committee"})
			}
		}
	}
	return slashCpk, unstakeCpk, committeeList, pendingList

}

func beaconSwap(
	pendingList []CandidateInfo,
	committeeList []CandidateInfo,
	fixNodeVotingPower int64,
	committeeSlot int,
) (
	[]CandidateInfo,
	[]CandidateInfo,
) {
	//sort candidate list
	sort.Slice(committeeList, func(i, j int) bool {
		return committeeList[i].score < committeeList[j].score
	})
	sort.Slice(pendingList, func(i, j int) bool {
		return pendingList[i].score > pendingList[j].score
	})
	//Pre-process to get candidate info
	pendingKeyList := []string{}
	pendingVotePower := []int64{}
	pendingScore := []uint64{}
	for _, v := range pendingList {
		pendingKeyList = append(pendingKeyList, v.cpkStr)
		pendingVotePower = append(pendingVotePower, v.votingPower)
		pendingScore = append(pendingScore, v.score)
	}
	committeeKeyList := []string{}
	committeeScore := []uint64{}
	sumCommitteeVotePower := fixNodeVotingPower
	for _, v := range committeeList {
		sumCommitteeVotePower += v.votingPower
		committeeKeyList = append(committeeKeyList, v.cpkStr)
		committeeScore = append(committeeScore, v.score)
	}
	listSumCommitteeVotePower := []int64{sumCommitteeVotePower}
	for i, v := range committeeList {
		listSumCommitteeVotePower = append(listSumCommitteeVotePower, listSumCommitteeVotePower[i]-v.votingPower)
	}
	swapInList := []CandidateInfo{}

	//fill pending key in empty slot
	fillInM, swapInVotePower := fillCandidate(pendingKeyList, pendingVotePower, sumCommitteeVotePower, int64(committeeSlot))

	//process swap in info
	newIdx := -1
	for _, v := range pendingList {
		if _, ok := fillInM[v.cpkStr]; !ok {
			newIdx++
			pendingList[newIdx] = v
			pendingKeyList[newIdx] = v.cpkStr
			pendingVotePower[newIdx] = v.votingPower
			pendingScore[newIdx] = v.score
		} else {
			swapInList = append(swapInList, v)
		}
	}
	pendingList = pendingList[:newIdx+1]
	pendingKeyList = pendingKeyList[:newIdx+1]
	pendingVotePower = pendingVotePower[:newIdx+1]
	pendingScore = pendingScore[:newIdx+1]

	// Start swap beacon
	swapInM, swapOutM := swapCandidate(pendingKeyList, pendingVotePower, pendingScore, committeeKeyList, listSumCommitteeVotePower, committeeScore, swapInVotePower)

	// Process swap in and swap out keys
	newIdx = -1
	for _, v := range pendingList {
		if _, ok := swapInM[v.cpkStr]; !ok {
			newIdx++
			pendingList[newIdx] = v
		} else {
			swapInList = append(swapInList, v)
		}
	}
	pendingList = pendingList[:newIdx+1]

	newIdx = -1
	for _, v := range committeeList {
		if _, ok := swapOutM[v.cpkStr]; !ok {
			newIdx++
			committeeList[newIdx] = v
		} else {
			pendingList = append(pendingList, v)
		}
	}
	committeeList = append(committeeList[:newIdx+1], swapInList...)

	// Sort again and return
	sort.Slice(committeeList, func(i, j int) bool {
		return committeeList[i].score < committeeList[j].score
	})
	sort.Slice(pendingList, func(i, j int) bool {
		return pendingList[i].score > pendingList[j].score
	})

	return pendingList, committeeList
}

func beacon_swap_v1(pendingList []CandidateInfo, committeeList []CandidateInfo, fixNodeVotingPower int64, committeeSlot int) ([]CandidateInfo, []CandidateInfo) {
	//sort candidate list
	sort.Slice(committeeList, func(i, j int) bool {
		return committeeList[i].score < committeeList[j].score
	})
	sort.Slice(pendingList, func(i, j int) bool {
		return pendingList[i].score > pendingList[j].score
	})

	//add to committeeSlot
	swapInVotingPower := func(candidates []CandidateInfo) (int64, int64) {
		swapIn := int64(0)
		total := fixNodeVotingPower
		for _, c := range candidates {
			if c.currentRole == "pending" {
				swapIn += c.votingPower
			}
			total += c.votingPower
		}
		return swapIn, total
	}

	for j := 0; j < len(pendingList); j++ {
		swapIn, total := swapInVotingPower(append(committeeList, pendingList[j]))
		if len(committeeList) < committeeSlot && swapIn < total/3 {
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
			if swapIn < total/3 {
				swapCommittee := committeeList[0]
				swapPending := pendingList[j]
				committeeList = append(committeeList[1:], swapPending)
				newPendingList := []CandidateInfo{}
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

func (s *BeaconCommitteeStateV4) beacon_swap_v1(env *BeaconCommitteeStateEnvironment) (
	map[string]bool, map[string]bool,
	map[string]incognitokey.CommitteePublicKey, map[string]incognitokey.CommitteePublicKey,
	error) {

	//slash
	slashCpk := map[string]bool{}
	for cpk, stakerInfo := range s.beaconCommittee {
		if stakerInfo.Performance < s.config.MIN_PERFORMANCE && !stakerInfo.FixedNode {
			slashCpk[cpk] = true
		}
	}

	//unstake
	unstakeCpk := map[string]bool{}
	for cpk, stakerInfo := range s.beaconCommittee {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = true
		}
	}
	for cpk, stakerInfo := range s.beaconPending {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = true
		}
	}
	for cpk, stakerInfo := range s.beaconWaiting {
		if stakerInfo.Unstake && !stakerInfo.FixedNode {
			unstakeCpk[cpk] = true
		}
	}

	//swap pending <-> committee
	newBeaconCommittee := map[string]incognitokey.CommitteePublicKey{}
	newBeaconPending := map[string]incognitokey.CommitteePublicKey{}
	pendingList := []CandidateInfo{}
	for cpk, stakerInfo := range s.beaconPending {
		if !slashCpk[cpk] && !unstakeCpk[cpk] {
			score := s.config.DEFAULT_PERFORMING * stakerInfo.StakingAmount
			pendingList = append(pendingList, CandidateInfo{stakerInfo.cpkStr, cpk, score, int64(math.Sqrt(float64(stakerInfo.StakingAmount))), "pending"})
		}
	}

	committeeList := []CandidateInfo{}
	fixNodeVotingPower := int64(0)
	for cpk, stakerInfo := range s.beaconCommittee {
		if !slashCpk[cpk] && !unstakeCpk[cpk] {
			score := stakerInfo.Performance * stakerInfo.StakingAmount
			if !stakerInfo.FixedNode {
				committeeList = append(committeeList, CandidateInfo{stakerInfo.cpkStr, cpk, score, int64(math.Floor(math.Sqrt(float64(stakerInfo.StakingAmount)))), "committee"})
			} else {
				newBeaconCommittee[cpk] = stakerInfo.cpkStr
				fixNodeVotingPower += int64(math.Sqrt(float64(stakerInfo.StakingAmount)))
			}
		}
	}

	pendingList, committeeList = beacon_swap_v1(pendingList, committeeList, fixNodeVotingPower, env.MaxBeaconCommitteeSize-len(newBeaconCommittee))

	//other candidate
	for _, candidate := range committeeList {
		newBeaconCommittee[candidate.cpkStr] = candidate.cpk
	}
	for _, candidate := range pendingList {
		newBeaconPending[candidate.cpkStr] = candidate.cpk
	}

	return slashCpk, unstakeCpk, newBeaconCommittee, newBeaconPending, nil
}
