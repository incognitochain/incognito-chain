package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_multi"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

var startTimeSlot uint64

func Test_Main4Committee_Case1(t *testing.T) {

	committee := []string{
		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
		"112t8s9hr9GWdfMBwwEGK12wSqvKeqpkw7jHzgHsK47EeTUcpnPAkQuzZa2xYcwHfrWtSZ6QZPeehkuDRN2u4e72HuEj7w6aKBSy4yUAZ2U3",
		"112t8rr9XGZLzuqjU2f59ey8gdyngZS3mWpwgoNzxPmNwAu8xmAQ87nnduVbZmU4Bhqnej4XTLQuS93yaG2iCGq3UXJSbBdZ8chqzhia4UuM",
		"112t8rrASXvBAtZ3dBTwXp6NH8KsX4dgmghUu36HtaPRJGvqeBqSSKb8yi7NUuNwUa58eKcyLGsXWtqfYVTgiPvAZ11GADLRZSHUNb9nssFw",
		"112t8rzc1pPSajQtjYVctFY5MGRgv2tqpRyD5zAZbwGXyh5Fum5Nafkn86iTw9w8RUhRnYH3wFLnFaZpDpb61gi3vBeQFTnzzyEErtd7jiBD",
	}

	committeePkStruct := []incognitokey.CommitteePublicKey{}
	committeePkBytes := [][]byte{}
	for _, v := range committee {
		p, _ := consensus_multi.LoadUserKeyFromIncPrivateKey(v)
		m, _ := consensus_multi.GetMiningKeyFromPrivateSeed(p)
		pk := m.GetPublicKey()
		committeePkStruct = append(committeePkStruct, *pk)
		committeePkBytes = append(committeePkBytes, pk.MiningPubKey["bls"])
	}
	nodeList := []*Node{}

	for i, v := range committee {
		p, _ := consensus_multi.LoadUserKeyFromIncPrivateKey(v)
		m, _ := consensus_multi.GetMiningKeyFromPrivateSeed(p)
		ni := NewNode(committeePkStruct, m, i)
		nodeList = append(nodeList, ni)
	}
	var startNode = func() {
		for i, v := range nodeList {
			v.nodeList = nodeList
			fmt.Println("node start", i)
			go v.Start()
		}
	}

	GetSimulation().nodeList = nodeList
	////simulation
	startTimeSlot = uint64(common.CalculateTimeSlot(time.Now().Unix()))
	GetSimulation().setStartTimeSlot(uint64(startTimeSlot))
	var setTimeSlot = func(s int) uint64 {
		return startTimeSlot + uint64(s) - 1
	}
	var setProposeCommunication = func(timeslot uint64, scenario []int) {
		if GetSimulation().scenario.proposeComm[timeslot] == nil {
			GetSimulation().scenario.proposeComm[timeslot] = []int{}
		}
		GetSimulation().scenario.proposeComm[timeslot] = scenario
	}
	var setVoteCommunication = func(timeslot uint64, nodeID int, scenario []int) {
		if GetSimulation().scenario.voteComm[timeslot] == nil {
			GetSimulation().scenario.voteComm[timeslot] = make(map[string][]int)
		}
		GetSimulation().scenario.voteComm[timeslot][fmt.Sprintf("%d", nodeID)] = scenario
	}

	/*
		START YOUR SIMULATION HERE
	*/

	timeslot := setTimeSlot(1) //normal communication, full connect by default
	beginProducerIdx := GetIndexOfBytes(nodeList[0].chain.GetBestView().GetProposerByTimeSlot(int64(timeslot), 2).MiningPubKey["bls"], committeePkBytes)

	timeslot = setTimeSlot(2) //block 3:2
	ct2 := []int{0, 0, 0, 0, 0, 0, 0, 0}
	ct2[(beginProducerIdx+3)%len(committee)] = 1
	setProposeCommunication(timeslot, ct2)
	setVoteCommunication(timeslot, 0, []int{0, 0, 0, 0, 0, 0, 0, 0})
	setVoteCommunication(timeslot, 1, []int{0, 0, 0, 0, 0, 0, 0, 0})
	setVoteCommunication(timeslot, 2, []int{0, 0, 0, 0, 0, 0, 0, 0})
	setVoteCommunication(timeslot, 3, []int{0, 0, 0, 0, 0, 0, 0, 0})
	setVoteCommunication(timeslot, 4, []int{0, 0, 0, 0, 0, 0, 0, 0})
	setVoteCommunication(timeslot, 5, []int{0, 0, 0, 0, 0, 0, 0, 0})
	setVoteCommunication(timeslot, 6, []int{0, 0, 0, 0, 0, 0, 0, 0})
	setVoteCommunication(timeslot, 7, []int{0, 0, 0, 0, 0, 0, 0, 0})

	//
	timeslot = setTimeSlot(3) //block 3:3
	ct3 := []int{1, 1, 1, 1, 1, 1, 1, 1}
	ct3[(beginProducerIdx+3)%len(committee)] = 0
	setProposeCommunication(timeslot, ct3)
	ct31 := []int{0, 0, 0, 0, 0, 0, 0, 0}
	ct31[(beginProducerIdx+4)%len(committee)] = 1
	setVoteCommunication(timeslot, 0, ct31)
	setVoteCommunication(timeslot, 1, ct31)
	setVoteCommunication(timeslot, 2, ct31)
	setVoteCommunication(timeslot, 3, ct31)
	setVoteCommunication(timeslot, 4, ct31)
	setVoteCommunication(timeslot, 5, ct31)
	setVoteCommunication(timeslot, 6, ct31)
	setVoteCommunication(timeslot, 7, ct31)

	//
	timeslot = setTimeSlot(4) //block 3:4 (3:2)
	ct4 := []int{1, 1, 1, 1, 1, 1, 1, 1}
	ct4[(beginProducerIdx+4)%len(committee)] = 0
	setProposeCommunication(timeslot, ct4)
	setVoteCommunication(timeslot, 0, ct4)
	setVoteCommunication(timeslot, 1, ct4)
	setVoteCommunication(timeslot, 2, ct4)
	setVoteCommunication(timeslot, 3, ct4)
	setVoteCommunication(timeslot, 4, ct4)
	setVoteCommunication(timeslot, 5, ct4)
	setVoteCommunication(timeslot, 6, ct4)
	setVoteCommunication(timeslot, 7, ct4)

	timeslot = setTimeSlot(5) //block 3:3->4:5

	timeslot = setTimeSlot(6)

	timeslot = setTimeSlot(7)
	/*
		END YOUR SIMULATION HERE
	*/
	GetSimulation().setMaxTimeSlot(timeslot)
	startNode()
	lastTimeSlot := uint64(0)
	for {
		curTimeSlot := (uint64(common.CalculateTimeSlot(time.Now().Unix())) - startTimeSlot) + 1
		if lastTimeSlot != curTimeSlot {
			time.AfterFunc(time.Millisecond*1000, func() {
				fmt.Println("==========================")
				fmt.Printf("Timeslot %v:\n", uint64(common.CalculateTimeSlot(time.Now().Unix()))-startTimeSlot+1)
				for i := 0; i < len(committee); i++ {
					fmt.Printf("Node %v: \n -Best view height: %d\n -Final view height: %d\n -View count: %v\n", i, nodeList[i].chain.GetBestView().GetHeight(), nodeList[i].chain.GetFinalView().GetHeight(), len(nodeList[i].chain.multiview.GetAllViewsWithBFS()))
				}
				fmt.Println("==========================")
			})
		}
		lastTimeSlot = curTimeSlot
		time.Sleep(1 * time.Millisecond)
		if curTimeSlot == 7 {
			return
		}
	}
}

//func Test_Main4BeaconCommittee_ScenarioA(t *testing.T) {
//	committee := []string{
//		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
//		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
//		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
//		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
//	}
//	committeePkStruct := []incognitokey.CommitteePublicKey{}
//	for _, v := range committee {
//		p, _ := blsbftv2.LoadUserKeyFromIncPrivateKey(v)
//		m, _ := blsbftv2.GetMiningKeyFromPrivateSeed(p)
//		committeePkStruct = append(committeePkStruct, m.GetPublicKey())
//	}
//	nodeList := []*Node{}
//	genesisTime, _ := time.Parse(app.GENESIS_TIMESTAMP, blockchain.TestnetGenesisBlockTime)
//	for {
//		if int(common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT))%len(committee) == len(committee)-1 {
//			break
//		} else {
//			time.Sleep(1 * time.Millisecond)
//		}
//	}
//
//	for i, _ := range committee {
//		ni := NewNodeBeacon(committeePkStruct, committee, i)
//		nodeList = append(nodeList, ni)
//	}
//	var startNode = func() {
//		for _, v := range nodeList {
//			v.nodeList = nodeList
//			go v.Start()
//		}
//	}
//	GetSimulation().nodeList = nodeList
//	//simulation
//	rootTimeSlot := nodeList[0].chain.GetBestView().GetRootTimeSlot()
//	currentTimeSlot := common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), 3)
//	startTimeSlot := rootTimeSlot + currentTimeSlot
//	fmt.Println("root Time slot", rootTimeSlot)
//	GetSimulation().setStartTimeSlot(startTimeSlot)
//	var setTimeSlot = func(s int) uint64 {
//		return startTimeSlot + uint64(s)
//	}
//	var setProposeCommunication = func(timeslot uint64, nodeID int, scenario []int) {
//		if GetSimulation().scenario.proposeComm[timeslot] == nil {
//			GetSimulation().scenario.proposeComm[timeslot] = make(map[string][]int)
//		}
//		GetSimulation().scenario.proposeComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
//	}
//	var setVoteCommunication = func(timeslot uint64, nodeID int, scenario []int) {
//		if GetSimulation().scenario.voteComm[timeslot] == nil {
//			GetSimulation().scenario.voteComm[timeslot] = make(map[string][]int)
//		}
//		GetSimulation().scenario.voteComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
//	}
//
//	for _, v := range nodeList {
//		v.consensusEngine.Logger.Info("\n\n")
//		v.consensusEngine.Logger.Info("===============================")
//		v.consensusEngine.Logger.Info("\n\n")
//		fmt.Printf("Node %s log is %s\n", v.id, fmt.Sprintf("log%s.log", v.id))
//	}
//
//	/*
//		START YOUR SIMULATION HERE
//	*/
//	timeslot := setTimeSlot(1) //normal communication, full connect by default
//
//	timeslot = setTimeSlot(2)
//	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 1})
//	setVoteCommunication(timeslot, 3, []int{0, 1, 0, 0})
//	//
//	timeslot = setTimeSlot(3)
//	setVoteCommunication(timeslot, 0, []int{0, 0, 0, 0})
//	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})
//	//
//	timeslot = setTimeSlot(4)
//	setVoteCommunication(timeslot, 1, []int{0, 1, 1, 1})
//	setVoteCommunication(timeslot, 2, []int{0, 1, 1, 1})
//	setVoteCommunication(timeslot, 3, []int{0, 1, 1, 1})
//
//	timeslot = setTimeSlot(5)
//	setProposeCommunication(timeslot, 0, []int{0, 0, 1, 0})
//	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
//
//	timeslot = setTimeSlot(6)
//	setVoteCommunication(timeslot, 0, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})
//
//	timeslot = setTimeSlot(7)
//	setVoteCommunication(timeslot, 1, []int{0, 1, 1, 1})
//	setVoteCommunication(timeslot, 2, []int{0, 1, 1, 1})
//	setVoteCommunication(timeslot, 3, []int{0, 1, 1, 1})
//
//	timeslot = setTimeSlot(8)
//	setProposeCommunication(timeslot, 3, []int{0, 0, 0, 0})
//
//	timeslot = setTimeSlot(9)
//	setProposeCommunication(timeslot, 0, []int{0, 0, 1, 0})
//	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
//
//	timeslot = setTimeSlot(10)
//	setVoteCommunication(timeslot, 0, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})
//
//	timeslot = setTimeSlot(11)
//	timeslot = setTimeSlot(12)
//
//	/*
//		END YOUR SIMULATION HERE
//	*/
//	GetSimulation().setMaxTimeSlot(timeslot)
//	startNode()
//	go func() {
//		lastTimeSlot := uint64(0)
//		for {
//			curTimeSlot := (common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT) - startTimeSlot) + 1
//			if lastTimeSlot != curTimeSlot {
//				time.AfterFunc(time.Millisecond*500, func() {
//					fmt.Printf("Best view height: %d. Final view height: %d\n", fullnode.GetBestView().GetHeight(), fullnode.GetFinalView().GetHeight())
//				})
//			}
//			for _, v := range nodeList {
//				if lastTimeSlot != curTimeSlot && curTimeSlot <= GetSimulation().maxTimeSlot {
//					v.consensusEngine.Logger.Info("========================================")
//					v.consensusEngine.Logger.Info("SIMULATION NODE", v.id, "TIMESLOT", curTimeSlot)
//					v.consensusEngine.Logger.Info("========================================")
//				}
//
//			}
//			lastTimeSlot = curTimeSlot
//			time.Sleep(1 * time.Millisecond)
//		}
//	}()
//	select {}
//}
//
//func Test_Main4BeaconCommittee_ScenarioB(t *testing.T) {
//	committee := []string{
//		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
//		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
//		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
//		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
//	}
//	committeePkStruct := []incognitokey.CommitteePublicKey{}
//	for _, v := range committee {
//		p, _ := blsbftv2.LoadUserKeyFromIncPrivateKey(v)
//		m, _ := blsbftv2.GetMiningKeyFromPrivateSeed(p)
//		committeePkStruct = append(committeePkStruct, m.GetPublicKey())
//	}
//	nodeList := []*Node{}
//	genesisTime, _ := time.Parse(app.GENESIS_TIMESTAMP, blockchain.TestnetGenesisBlockTime)
//	for {
//		if int(common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT))%len(committee) == len(committee)-1 {
//			break
//		} else {
//			time.Sleep(1 * time.Millisecond)
//		}
//	}
//
//	for i, _ := range committee {
//		ni := NewNodeBeacon(committeePkStruct, committee, i)
//		nodeList = append(nodeList, ni)
//	}
//	var startNode = func() {
//		for _, v := range nodeList {
//			v.nodeList = nodeList
//			go v.Start()
//		}
//	}
//	GetSimulation().nodeList = nodeList
//	//simulation
//	rootTimeSlot := nodeList[0].chain.GetBestView().GetRootTimeSlot()
//	currentTimeSlot := common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), 3)
//	startTimeSlot := rootTimeSlot + currentTimeSlot
//	fmt.Println("root Time slot", rootTimeSlot)
//	GetSimulation().setStartTimeSlot(startTimeSlot)
//	var setTimeSlot = func(s int) uint64 {
//		return startTimeSlot + uint64(s)
//	}
//	var setProposeCommunication = func(timeslot uint64, nodeID int, scenario []int) {
//		if GetSimulation().scenario.proposeComm[timeslot] == nil {
//			GetSimulation().scenario.proposeComm[timeslot] = make(map[string][]int)
//		}
//		GetSimulation().scenario.proposeComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
//	}
//	var setVoteCommunication = func(timeslot uint64, nodeID int, scenario []int) {
//		if GetSimulation().scenario.voteComm[timeslot] == nil {
//			GetSimulation().scenario.voteComm[timeslot] = make(map[string][]int)
//		}
//		GetSimulation().scenario.voteComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
//	}
//
//	for _, v := range nodeList {
//		v.consensusEngine.Logger.Info("\n\n")
//		v.consensusEngine.Logger.Info("===============================")
//		v.consensusEngine.Logger.Info("\n\n")
//		fmt.Printf("Node %s log is %s\n", v.id, fmt.Sprintf("log%s.log", v.id))
//	}
//
//	/*
//		START YOUR SIMULATION HERE
//	*/
//	timeslot := setTimeSlot(1) //normal communication, full connect by default
//
//	timeslot = setTimeSlot(2)
//	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 1})
//	setVoteCommunication(timeslot, 3, []int{0, 1, 0, 0})
//	//
//	timeslot = setTimeSlot(3)
//	setVoteCommunication(timeslot, 0, []int{0, 0, 0, 0})
//	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})
//	//
//	timeslot = setTimeSlot(4)
//
//	timeslot = setTimeSlot(5)
//	setProposeCommunication(timeslot, 0, []int{0, 0, 1, 0})
//	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
//
//	timeslot = setTimeSlot(6)
//	setVoteCommunication(timeslot, 0, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
//	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})
//
//	timeslot = setTimeSlot(7)
//	setVoteCommunication(timeslot, 1, []int{0, 1, 1, 1})
//	setVoteCommunication(timeslot, 2, []int{0, 1, 1, 1})
//	setVoteCommunication(timeslot, 3, []int{0, 1, 1, 1})
//
//	timeslot = setTimeSlot(8)
//	setProposeCommunication(timeslot, 3, []int{0, 1, 0, 0})
//	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 1})
//
//	timeslot = setTimeSlot(9)
//	setProposeCommunication(timeslot, 0, []int{0, 0, 0, 0})
//
//	timeslot = setTimeSlot(10)
//	setProposeCommunication(timeslot, 1, []int{0, 1, 1, 1})
//	timeslot = setTimeSlot(11)
//	setProposeCommunication(timeslot, 2, []int{0, 1, 1, 1})
//	timeslot = setTimeSlot(12)
//	setProposeCommunication(timeslot, 3, []int{0, 1, 1, 1})
//
//	timeslot = setTimeSlot(13)
//
//	/*
//		END YOUR SIMULATION HERE
//	*/
//	GetSimulation().setMaxTimeSlot(timeslot)
//	startNode()
//	go func() {
//		lastTimeSlot := uint64(0)
//		for {
//			curTimeSlot := (common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT) - startTimeSlot) + 1
//			if lastTimeSlot != curTimeSlot {
//				time.AfterFunc(time.Millisecond*500, func() {
//					fmt.Printf("Best view height: %d. Final view height: %d\n", fullnode.GetBestView().GetHeight(), fullnode.GetFinalView().GetHeight())
//				})
//			}
//			for _, v := range nodeList {
//				if lastTimeSlot != curTimeSlot && curTimeSlot <= GetSimulation().maxTimeSlot {
//					v.consensusEngine.Logger.Info("========================================")
//					v.consensusEngine.Logger.Info("SIMULATION NODE", v.id, "TIMESLOT", curTimeSlot)
//					v.consensusEngine.Logger.Info("========================================")
//				}
//			}
//			lastTimeSlot = curTimeSlot
//			time.Sleep(1 * time.Millisecond)
//		}
//	}()
//	select {}
//}
//
//func Test_Main4BeaconCommittee_ScenarioC(t *testing.T) {
//	committee := []string{
//		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
//		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
//		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
//		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
//	}
//	committeePkStruct := []incognitokey.CommitteePublicKey{}
//	for _, v := range committee {
//		p, _ := blsbftv2.LoadUserKeyFromIncPrivateKey(v)
//		m, _ := blsbftv2.GetMiningKeyFromPrivateSeed(p)
//		committeePkStruct = append(committeePkStruct, m.GetPublicKey())
//	}
//	nodeList := []*Node{}
//	genesisTime, _ := time.Parse(app.GENESIS_TIMESTAMP, blockchain.TestnetGenesisBlockTime)
//	for {
//		if int(common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT))%len(committee) == len(committee)-1 {
//			break
//		} else {
//			time.Sleep(1 * time.Millisecond)
//		}
//	}
//
//	for i, _ := range committee {
//		ni := NewNodeBeacon(committeePkStruct, committee, i)
//		nodeList = append(nodeList, ni)
//	}
//	var startNode = func() {
//		for _, v := range nodeList {
//			v.nodeList = nodeList
//			go v.Start()
//		}
//	}
//	GetSimulation().nodeList = nodeList
//	//simulation
//	rootTimeSlot := nodeList[0].chain.GetBestView().GetRootTimeSlot()
//	currentTimeSlot := common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), 3)
//	startTimeSlot := rootTimeSlot + currentTimeSlot
//	fmt.Println("root Time slot", rootTimeSlot)
//	GetSimulation().setStartTimeSlot(startTimeSlot)
//	var setTimeSlot = func(s int) uint64 {
//		return startTimeSlot + uint64(s)
//	}
//	var setProposeCommunication = func(timeslot uint64, nodeID int, scenario []int) {
//		if GetSimulation().scenario.proposeComm[timeslot] == nil {
//			GetSimulation().scenario.proposeComm[timeslot] = make(map[string][]int)
//		}
//		GetSimulation().scenario.proposeComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
//	}
//	//var setVoteCommunication = func(timeslot uint64, nodeID int, scenario []int) {
//	//	if GetSimulation().scenario.voteComm[timeslot] == nil {
//	//		GetSimulation().scenario.voteComm[timeslot] = make(map[string][]int)
//	//	}
//	//	GetSimulation().scenario.voteComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
//	//}
//
//	for _, v := range nodeList {
//		v.consensusEngine.Logger.Info("\n\n")
//		v.consensusEngine.Logger.Info("===============================")
//		v.consensusEngine.Logger.Info("\n\n")
//		fmt.Printf("Node %s log is %s\n", v.id, fmt.Sprintf("log%s.log", v.id))
//	}
//
//	/*
//		START YOUR SIMULATION HERE
//	*/
//	timeslot := setTimeSlot(1)
//	setProposeCommunication(timeslot, 0, []int{0, 0, 0, 0})
//	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 0})
//	setProposeCommunication(timeslot, 2, []int{0, 0, 0, 0})
//	setProposeCommunication(timeslot, 3, []int{0, 0, 0, 0})
//
//	timeslot = setTimeSlot(100) //normal communication, full connect by default
//
//	/*
//		END YOUR SIMULATION HERE
//	*/
//	GetSimulation().setMaxTimeSlot(timeslot)
//	startNode()
//	go func() {
//		lastTimeSlot := uint64(0)
//		for {
//			curTimeSlot := (common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT) - startTimeSlot) + 1
//			if lastTimeSlot != curTimeSlot {
//				time.AfterFunc(time.Millisecond*500, func() {
//					fmt.Printf("Best view height: %d. Final view height: %d\n", fullnode.GetBestView().GetHeight(), fullnode.GetFinalView().GetHeight())
//				})
//			}
//			for _, v := range nodeList {
//				if lastTimeSlot != curTimeSlot && curTimeSlot <= GetSimulation().maxTimeSlot {
//					v.consensusEngine.Logger.Info("========================================")
//					v.consensusEngine.Logger.Info("SIMULATION NODE", v.id, "TIMESLOT", curTimeSlot)
//					v.consensusEngine.Logger.Info("========================================")
//				}
//			}
//			lastTimeSlot = curTimeSlot
//			time.Sleep(1 * time.Millisecond)
//		}
//	}()
//	select {}
//}
