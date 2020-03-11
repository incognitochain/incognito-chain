package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain_v2/app"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbftv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func Test_Main4BeaconCommittee(t *testing.T) {

	committee := []string{
		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
	}
	committeePkStruct := []incognitokey.CommitteePublicKey{}
	for _, v := range committee {
		p, _ := blsbftv2.LoadUserKeyFromIncPrivateKey(v)
		m, _ := blsbftv2.GetMiningKeyFromPrivateSeed(p)
		committeePkStruct = append(committeePkStruct, m.GetPublicKey())
	}
	nodeList := []*Node{}
	genesisTime, _ := time.Parse(app.GENESIS_TIMESTAMP, blockchain.TestnetGenesisBlockTime)
	for {
		if int(common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT))%len(committee) == len(committee)-1 {
			break
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}

	for i, _ := range committee {
		ni := NewNodeBeacon(committeePkStruct, committee, i)
		nodeList = append(nodeList, ni)
	}
	var startNode = func() {
		for _, v := range nodeList {
			v.nodeList = nodeList
			go v.Start()
		}
	}
	GetSimulation().nodeList = nodeList
	//simulation
	rootTimeSlot := nodeList[0].chain.GetBestView().GetRootTimeSlot()
	currentTimeSlot := common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), 3)
	startTimeSlot := rootTimeSlot + currentTimeSlot
	fmt.Println("root Time slot", rootTimeSlot)
	GetSimulation().setStartTimeSlot(startTimeSlot)
	var setTimeSlot = func(s int) uint64 {
		return startTimeSlot + uint64(s)
	}
	var setProposeCommunication = func(timeslot uint64, nodeID int, scenario []int) {
		if GetSimulation().scenario.proposeComm[timeslot] == nil {
			GetSimulation().scenario.proposeComm[timeslot] = make(map[string][]int)
		}
		GetSimulation().scenario.proposeComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
	}
	var setVoteCommunication = func(timeslot uint64, nodeID int, scenario []int) {
		if GetSimulation().scenario.voteComm[timeslot] == nil {
			GetSimulation().scenario.voteComm[timeslot] = make(map[string][]int)
		}
		GetSimulation().scenario.voteComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
	}

	for _, v := range nodeList {
		v.consensusEngine.Logger.Info("\n\n")
		v.consensusEngine.Logger.Info("===============================")
		v.consensusEngine.Logger.Info("\n\n")
		fmt.Printf("Node %s log is %s\n", v.id, fmt.Sprintf("log%s.log", v.id))
	}

	/*
		START YOUR SIMULATION HERE
	*/
	timeslot := setTimeSlot(1) //normal communication, full connect by default

	timeslot = setTimeSlot(2)
	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 1})
	setVoteCommunication(timeslot, 3, []int{0, 1, 0, 0})
	//
	timeslot = setTimeSlot(3)
	setVoteCommunication(timeslot, 0, []int{0, 0, 0, 0})
	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})
	//
	timeslot = setTimeSlot(4)
	setVoteCommunication(timeslot, 1, []int{0, 1, 1, 1})
	setVoteCommunication(timeslot, 2, []int{0, 1, 1, 1})
	setVoteCommunication(timeslot, 3, []int{0, 1, 1, 1})

	timeslot = setTimeSlot(5) //normal cimmunication

	/*
		END YOUR SIMULATION HERE
	*/
	GetSimulation().setMaxTimeSlot(timeslot)
	startNode()
	go func() {
		lastTimeSlot := uint64(0)
		for {
			curTimeSlot := (common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT) - startTimeSlot) + 1
			if lastTimeSlot != curTimeSlot {
				time.AfterFunc(time.Millisecond*500, func() {
					fmt.Printf("Best view height: %d. Final view height: %d\n", fullnode.GetBestView().GetHeight(), fullnode.GetFinalView().GetHeight())
				})
			}
			for _, v := range nodeList {
				if lastTimeSlot != curTimeSlot && curTimeSlot <= GetSimulation().maxTimeSlot {
					v.consensusEngine.Logger.Info("========================================")
					v.consensusEngine.Logger.Info("SIMULATION NODE", v.id, "TIMESLOT", curTimeSlot)
					v.consensusEngine.Logger.Info("========================================")
				}

			}
			lastTimeSlot = curTimeSlot
			time.Sleep(1 * time.Millisecond)
		}
	}()
	select {}

}

func Test_Main4BeaconCommittee_ScenarioA(t *testing.T) {
	committee := []string{
		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
	}
	committeePkStruct := []incognitokey.CommitteePublicKey{}
	for _, v := range committee {
		p, _ := blsbftv2.LoadUserKeyFromIncPrivateKey(v)
		m, _ := blsbftv2.GetMiningKeyFromPrivateSeed(p)
		committeePkStruct = append(committeePkStruct, m.GetPublicKey())
	}
	nodeList := []*Node{}
	genesisTime, _ := time.Parse(app.GENESIS_TIMESTAMP, blockchain.TestnetGenesisBlockTime)
	for {
		if int(common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT))%len(committee) == len(committee)-1 {
			break
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}

	for i, _ := range committee {
		ni := NewNodeBeacon(committeePkStruct, committee, i)
		nodeList = append(nodeList, ni)
	}
	var startNode = func() {
		for _, v := range nodeList {
			v.nodeList = nodeList
			go v.Start()
		}
	}
	GetSimulation().nodeList = nodeList
	//simulation
	rootTimeSlot := nodeList[0].chain.GetBestView().GetRootTimeSlot()
	currentTimeSlot := common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), 3)
	startTimeSlot := rootTimeSlot + currentTimeSlot
	fmt.Println("root Time slot", rootTimeSlot)
	GetSimulation().setStartTimeSlot(startTimeSlot)
	var setTimeSlot = func(s int) uint64 {
		return startTimeSlot + uint64(s)
	}
	var setProposeCommunication = func(timeslot uint64, nodeID int, scenario []int) {
		if GetSimulation().scenario.proposeComm[timeslot] == nil {
			GetSimulation().scenario.proposeComm[timeslot] = make(map[string][]int)
		}
		GetSimulation().scenario.proposeComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
	}
	var setVoteCommunication = func(timeslot uint64, nodeID int, scenario []int) {
		if GetSimulation().scenario.voteComm[timeslot] == nil {
			GetSimulation().scenario.voteComm[timeslot] = make(map[string][]int)
		}
		GetSimulation().scenario.voteComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
	}

	for _, v := range nodeList {
		v.consensusEngine.Logger.Info("\n\n")
		v.consensusEngine.Logger.Info("===============================")
		v.consensusEngine.Logger.Info("\n\n")
		fmt.Printf("Node %s log is %s\n", v.id, fmt.Sprintf("log%s.log", v.id))
	}

	/*
		START YOUR SIMULATION HERE
	*/
	timeslot := setTimeSlot(1) //normal communication, full connect by default

	timeslot = setTimeSlot(2)
	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 1})
	setVoteCommunication(timeslot, 3, []int{0, 1, 0, 0})
	//
	timeslot = setTimeSlot(3)
	setVoteCommunication(timeslot, 0, []int{0, 0, 0, 0})
	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})
	//
	timeslot = setTimeSlot(4)
	setVoteCommunication(timeslot, 1, []int{0, 1, 1, 1})
	setVoteCommunication(timeslot, 2, []int{0, 1, 1, 1})
	setVoteCommunication(timeslot, 3, []int{0, 1, 1, 1})

	timeslot = setTimeSlot(5)
	setProposeCommunication(timeslot, 0, []int{0, 0, 1, 0})
	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})

	timeslot = setTimeSlot(6)
	setVoteCommunication(timeslot, 0, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})

	timeslot = setTimeSlot(7)
	setVoteCommunication(timeslot, 1, []int{0, 1, 1, 1})
	setVoteCommunication(timeslot, 2, []int{0, 1, 1, 1})
	setVoteCommunication(timeslot, 3, []int{0, 1, 1, 1})

	timeslot = setTimeSlot(8)
	setProposeCommunication(timeslot, 3, []int{0, 0, 0, 0})

	timeslot = setTimeSlot(9)
	setProposeCommunication(timeslot, 0, []int{0, 0, 1, 0})
	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})

	timeslot = setTimeSlot(10)
	setVoteCommunication(timeslot, 0, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})

	timeslot = setTimeSlot(11)
	timeslot = setTimeSlot(12)

	/*
		END YOUR SIMULATION HERE
	*/
	GetSimulation().setMaxTimeSlot(timeslot)
	startNode()
	go func() {
		lastTimeSlot := uint64(0)
		for {
			curTimeSlot := (common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT) - startTimeSlot) + 1
			if lastTimeSlot != curTimeSlot {
				time.AfterFunc(time.Millisecond*500, func() {
					fmt.Printf("Best view height: %d. Final view height: %d\n", fullnode.GetBestView().GetHeight(), fullnode.GetFinalView().GetHeight())
				})
			}
			for _, v := range nodeList {
				if lastTimeSlot != curTimeSlot && curTimeSlot <= GetSimulation().maxTimeSlot {
					v.consensusEngine.Logger.Info("========================================")
					v.consensusEngine.Logger.Info("SIMULATION NODE", v.id, "TIMESLOT", curTimeSlot)
					v.consensusEngine.Logger.Info("========================================")
				}

			}
			lastTimeSlot = curTimeSlot
			time.Sleep(1 * time.Millisecond)
		}
	}()
	select {}
}

func Test_Main4BeaconCommittee_ScenarioB(t *testing.T) {
	committee := []string{
		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
	}
	committeePkStruct := []incognitokey.CommitteePublicKey{}
	for _, v := range committee {
		p, _ := blsbftv2.LoadUserKeyFromIncPrivateKey(v)
		m, _ := blsbftv2.GetMiningKeyFromPrivateSeed(p)
		committeePkStruct = append(committeePkStruct, m.GetPublicKey())
	}
	nodeList := []*Node{}
	genesisTime, _ := time.Parse(app.GENESIS_TIMESTAMP, blockchain.TestnetGenesisBlockTime)
	for {
		if int(common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT))%len(committee) == len(committee)-1 {
			break
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}

	for i, _ := range committee {
		ni := NewNodeBeacon(committeePkStruct, committee, i)
		nodeList = append(nodeList, ni)
	}
	var startNode = func() {
		for _, v := range nodeList {
			v.nodeList = nodeList
			go v.Start()
		}
	}
	GetSimulation().nodeList = nodeList
	//simulation
	rootTimeSlot := nodeList[0].chain.GetBestView().GetRootTimeSlot()
	currentTimeSlot := common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), 3)
	startTimeSlot := rootTimeSlot + currentTimeSlot
	fmt.Println("root Time slot", rootTimeSlot)
	GetSimulation().setStartTimeSlot(startTimeSlot)
	var setTimeSlot = func(s int) uint64 {
		return startTimeSlot + uint64(s)
	}
	var setProposeCommunication = func(timeslot uint64, nodeID int, scenario []int) {
		if GetSimulation().scenario.proposeComm[timeslot] == nil {
			GetSimulation().scenario.proposeComm[timeslot] = make(map[string][]int)
		}
		GetSimulation().scenario.proposeComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
	}
	var setVoteCommunication = func(timeslot uint64, nodeID int, scenario []int) {
		if GetSimulation().scenario.voteComm[timeslot] == nil {
			GetSimulation().scenario.voteComm[timeslot] = make(map[string][]int)
		}
		GetSimulation().scenario.voteComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
	}

	for _, v := range nodeList {
		v.consensusEngine.Logger.Info("\n\n")
		v.consensusEngine.Logger.Info("===============================")
		v.consensusEngine.Logger.Info("\n\n")
		fmt.Printf("Node %s log is %s\n", v.id, fmt.Sprintf("log%s.log", v.id))
	}

	/*
		START YOUR SIMULATION HERE
	*/
	timeslot := setTimeSlot(1) //normal communication, full connect by default

	timeslot = setTimeSlot(2)
	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 1})
	setVoteCommunication(timeslot, 3, []int{0, 1, 0, 0})
	//
	timeslot = setTimeSlot(3)
	setVoteCommunication(timeslot, 0, []int{0, 0, 0, 0})
	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})
	//
	timeslot = setTimeSlot(4)

	timeslot = setTimeSlot(5)
	setProposeCommunication(timeslot, 0, []int{0, 0, 1, 0})
	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})

	timeslot = setTimeSlot(6)
	setVoteCommunication(timeslot, 0, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 1, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 2, []int{1, 0, 0, 0})
	setVoteCommunication(timeslot, 3, []int{1, 0, 0, 0})

	timeslot = setTimeSlot(7)
	setVoteCommunication(timeslot, 1, []int{0, 1, 1, 1})
	setVoteCommunication(timeslot, 2, []int{0, 1, 1, 1})
	setVoteCommunication(timeslot, 3, []int{0, 1, 1, 1})

	timeslot = setTimeSlot(8)
	setProposeCommunication(timeslot, 3, []int{0, 1, 0, 0})
	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 1})

	timeslot = setTimeSlot(9)
	setProposeCommunication(timeslot, 0, []int{0, 0, 0, 0})

	timeslot = setTimeSlot(10)
	setProposeCommunication(timeslot, 1, []int{0, 1, 1, 1})
	timeslot = setTimeSlot(11)
	setProposeCommunication(timeslot, 2, []int{0, 1, 1, 1})
	timeslot = setTimeSlot(12)
	setProposeCommunication(timeslot, 3, []int{0, 1, 1, 1})

	timeslot = setTimeSlot(13)

	/*
		END YOUR SIMULATION HERE
	*/
	GetSimulation().setMaxTimeSlot(timeslot)
	startNode()
	go func() {
		lastTimeSlot := uint64(0)
		for {
			curTimeSlot := (common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT) - startTimeSlot) + 1
			if lastTimeSlot != curTimeSlot {
				time.AfterFunc(time.Millisecond*500, func() {
					fmt.Printf("Best view height: %d. Final view height: %d\n", fullnode.GetBestView().GetHeight(), fullnode.GetFinalView().GetHeight())
				})
			}
			for _, v := range nodeList {
				if lastTimeSlot != curTimeSlot && curTimeSlot <= GetSimulation().maxTimeSlot {
					v.consensusEngine.Logger.Info("========================================")
					v.consensusEngine.Logger.Info("SIMULATION NODE", v.id, "TIMESLOT", curTimeSlot)
					v.consensusEngine.Logger.Info("========================================")
				}
			}
			lastTimeSlot = curTimeSlot
			time.Sleep(1 * time.Millisecond)
		}
	}()
	select {}
}

func Test_Main4BeaconCommittee_ScenarioC(t *testing.T) {
	committee := []string{
		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
	}
	committeePkStruct := []incognitokey.CommitteePublicKey{}
	for _, v := range committee {
		p, _ := blsbftv2.LoadUserKeyFromIncPrivateKey(v)
		m, _ := blsbftv2.GetMiningKeyFromPrivateSeed(p)
		committeePkStruct = append(committeePkStruct, m.GetPublicKey())
	}
	nodeList := []*Node{}
	genesisTime, _ := time.Parse(app.GENESIS_TIMESTAMP, blockchain.TestnetGenesisBlockTime)
	for {
		if int(common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT))%len(committee) == len(committee)-1 {
			break
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}

	for i, _ := range committee {
		ni := NewNodeBeacon(committeePkStruct, committee, i)
		nodeList = append(nodeList, ni)
	}
	var startNode = func() {
		for _, v := range nodeList {
			v.nodeList = nodeList
			go v.Start()
		}
	}
	GetSimulation().nodeList = nodeList
	//simulation
	rootTimeSlot := nodeList[0].chain.GetBestView().GetRootTimeSlot()
	currentTimeSlot := common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), 3)
	startTimeSlot := rootTimeSlot + currentTimeSlot
	fmt.Println("root Time slot", rootTimeSlot)
	GetSimulation().setStartTimeSlot(startTimeSlot)
	var setTimeSlot = func(s int) uint64 {
		return startTimeSlot + uint64(s)
	}
	var setProposeCommunication = func(timeslot uint64, nodeID int, scenario []int) {
		if GetSimulation().scenario.proposeComm[timeslot] == nil {
			GetSimulation().scenario.proposeComm[timeslot] = make(map[string][]int)
		}
		GetSimulation().scenario.proposeComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
	}
	//var setVoteCommunication = func(timeslot uint64, nodeID int, scenario []int) {
	//	if GetSimulation().scenario.voteComm[timeslot] == nil {
	//		GetSimulation().scenario.voteComm[timeslot] = make(map[string][]int)
	//	}
	//	GetSimulation().scenario.voteComm[timeslot][fmt.Sprintf("%d", int(int(startTimeSlot)+nodeID)%len(committee))] = scenario
	//}

	for _, v := range nodeList {
		v.consensusEngine.Logger.Info("\n\n")
		v.consensusEngine.Logger.Info("===============================")
		v.consensusEngine.Logger.Info("\n\n")
		fmt.Printf("Node %s log is %s\n", v.id, fmt.Sprintf("log%s.log", v.id))
	}

	/*
		START YOUR SIMULATION HERE
	*/
	timeslot := setTimeSlot(1)
	setProposeCommunication(timeslot, 0, []int{0, 0, 0, 0})
	setProposeCommunication(timeslot, 1, []int{0, 0, 0, 0})
	setProposeCommunication(timeslot, 2, []int{0, 0, 0, 0})
	setProposeCommunication(timeslot, 3, []int{0, 0, 0, 0})

	timeslot = setTimeSlot(100) //normal communication, full connect by default

	/*
		END YOUR SIMULATION HERE
	*/
	GetSimulation().setMaxTimeSlot(timeslot)
	startNode()
	go func() {
		lastTimeSlot := uint64(0)
		for {
			curTimeSlot := (common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT) - startTimeSlot) + 1
			if lastTimeSlot != curTimeSlot {
				time.AfterFunc(time.Millisecond*500, func() {
					fmt.Printf("Best view height: %d. Final view height: %d\n", fullnode.GetBestView().GetHeight(), fullnode.GetFinalView().GetHeight())
				})
			}
			for _, v := range nodeList {
				if lastTimeSlot != curTimeSlot && curTimeSlot <= GetSimulation().maxTimeSlot {
					v.consensusEngine.Logger.Info("========================================")
					v.consensusEngine.Logger.Info("SIMULATION NODE", v.id, "TIMESLOT", curTimeSlot)
					v.consensusEngine.Logger.Info("========================================")
				}
			}
			lastTimeSlot = curTimeSlot
			time.Sleep(1 * time.Millisecond)
		}
	}()
	select {}
}
