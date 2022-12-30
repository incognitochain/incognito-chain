package committeestate

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"log"
	"math"
	"sort"
)

type CandidateInfo struct {
	cpk          incognitokey.CommitteePublicKey
	cpkStr       string
	score        uint64
	stakingPower int64
	currentRole  string
}

func beacon_swap_v1(pendingList []CandidateInfo, committeeList []CandidateInfo, fixNodeStakingPower int64, committeeSlot int) ([]CandidateInfo, []CandidateInfo) {
	//sort candidate list
	sort.Slice(committeeList, func(i, j int) bool {
		return committeeList[i].score < committeeList[j].score
	})
	sort.Slice(pendingList, func(i, j int) bool {
		return pendingList[i].score > pendingList[j].score
	})

	//add to committeeSlot
	swapInStakingPower := func(candidates []CandidateInfo) (int64, int64) {
		swapIn := int64(0)
		total := fixNodeStakingPower
		for _, c := range candidates {
			if c.currentRole == "pending" {
				swapIn += c.stakingPower
			}
			total += c.stakingPower
		}
		return swapIn, total
	}

	for j := 0; j < len(pendingList); j++ {
		swapIn, total := swapInStakingPower(append(committeeList, pendingList[j]))
		if len(committeeList) < committeeSlot && swapIn < total/3 {
			committeeList = append(committeeList, pendingList[j])        //append pending j
			pendingList = append(pendingList[0:j], pendingList[j+1:]...) //remove pending j
			j--
		}
	}

	endSwapIn := false
	//find pending candidate to replace committee with the smallest score
	if len(pendingList) > 0 {
		for true {
			if endSwapIn {
				break
			}
			for j := 0; j < len(pendingList); j++ {
				//no candidate in committee
				if len(committeeList) == 0 {
					endSwapIn = true
					break
				}

				//if we swap all old committee list
				if committeeList[0].currentRole != "committee" {
					endSwapIn = true
					break
				}

				//if we swap all old pending list
				if pendingList[j].currentRole != "pending" {
					endSwapIn = true
					break
				}

				//if commitee[0] score is better the best pending candidate score
				if committeeList[0].score >= pendingList[j].score {
					endSwapIn = true
					break
				}

				//check we can swap
				swapIn, total := swapInStakingPower(append(committeeList[1:], pendingList[j]))
				execSwap := false
				if swapIn < total/3 {
					execSwap = true
					committeeList = append(committeeList[1:], pendingList[j])    //append pending j
					pendingList = append(pendingList[0:j], pendingList[j+1:]...) //remove pending j
					j--
					pendingList = append(pendingList, committeeList[0])
				}

				//if cannot find any pending candidate that can swap committee[0]
				if !execSwap && j == len(pendingList)-1 {
					endSwapIn = true
					break
				}
			}
		}
	}

	//re-sort the swap list
	sort.Slice(committeeList, func(i, j int) bool {
		return committeeList[i].score < committeeList[j].score
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
	fixNodeStakingPower := int64(0)
	for cpk, stakerInfo := range s.beaconCommittee {
		if !slashCpk[cpk] && !unstakeCpk[cpk] {
			score := stakerInfo.Performance * stakerInfo.StakingAmount
			if !stakerInfo.FixedNode {
				committeeList = append(committeeList, CandidateInfo{stakerInfo.cpkStr, cpk, score, int64(math.Sqrt(float64(stakerInfo.StakingAmount))), "committee"})
			} else {
				newBeaconCommittee[cpk] = stakerInfo.cpkStr
				fixNodeStakingPower += int64(math.Sqrt(float64(stakerInfo.StakingAmount)))
			}
		}
	}

	pendingList, committeeList = beacon_swap_v1(pendingList, committeeList, fixNodeStakingPower, env.MaxBeaconCommitteeSize-len(newBeaconCommittee))

	//other candidate
	for _, candidate := range committeeList {
		newBeaconCommittee[candidate.cpkStr] = candidate.cpk
	}
	for _, candidate := range pendingList {
		newBeaconPending[candidate.cpkStr] = candidate.cpk
	}

	log.Println("newBeaconCommittee", len(newBeaconCommittee))
	log.Println("newBeaconPending", len(newBeaconPending))
	return slashCpk, unstakeCpk, newBeaconCommittee, newBeaconPending, nil
}
