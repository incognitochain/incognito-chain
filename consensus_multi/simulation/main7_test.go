package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/v2/shard"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbftv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"math/rand"
	"testing"
	"time"
)

func Test_Main7Committee_ScenarioC(t *testing.T) {
	committee := []string{
		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL",
		"112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5",
		"112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR",
		"112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43",
		"112t8rnXBPJQWJTyPdzWsfsUCFTDhcas3y2MYsauKo66euh1udG8dSh2ZszSbfqHwCpYHPRSpFTxYkUcVa619XUM6DjdV7FfUWvYoziWE2Bm",
		"112t8rnXN2SLxQncPYvFdzEivznKjBxK5byYmPbAhnEEv8TderLG7NUD7nwAEDu7DJ7pnCKw9N5PuTuELCHz8qKc7z9S9jF8QG41u7Vomc6L",
		"112t8rnXs5os49h71E7utfHatnWGQnirbVF2b5Ua8h1ttidk1S5AFcUqHCDmpMziiFC15BG8W1LQKK5tYcvr2CM7DyYgsfVmAWYh4kQ6f33T",
	}
	committeePkStruct := []incognitokey.CommitteePublicKey{}
	for _, v := range committee {
		p, _ := blsbftv2.LoadUserKeyFromIncPrivateKey(v)
		m, _ := blsbftv2.GetMiningKeyFromPrivateSeed(p)
		committeePkStruct = append(committeePkStruct, m.GetPublicKey())
	}
	nodeList := []*Node{}
	genesisTime, _ := time.Parse(shard.GENESIS_TIMESTAMP, blockchain.TestnetGenesisBlockTime)
	for {
		if int(common.GetTimeSlot(genesisTime.Unix(), time.Now().Unix(), blsbftv2.TIMESLOT))%len(committee) == len(committee)-1 {
			break
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}

	for i, _ := range committee {
		ni := NewNode(committeePkStruct, committee, i)
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
	startTimeSlot := rootTimeSlot + 1
	fmt.Println("root Time slot", rootTimeSlot)
	GetSimulation().setStartTimeSlot(startTimeSlot)
	var setTimeSlot = func(s int) uint64 {
		return startTimeSlot + uint64(s) - 1
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
	var NO_CONNECT = []int{0, 0, 0, 0, 0, 0, 0}
	for i := 2; i < 100; i++ {
		timeslot = setTimeSlot(i)
		r1, r2 := SelectTwo(len(committee))
		fmt.Println("Timeslot", timeslot, "Disconnect node", r1, r2)
		setProposeCommunication(timeslot, r1, NO_CONNECT)
		setVoteCommunication(timeslot, r1, NO_CONNECT)
		setProposeCommunication(timeslot, r2, NO_CONNECT)
		setVoteCommunication(timeslot, r2, NO_CONNECT)

		connect := []int{1, 1, 1, 1, 1, 1, 1, 1}
		connect[r1] = 0
		connect[r2] = 0
		for i, _ := range committee {
			if i == r1 || i == r2 {
				continue
			}
			setProposeCommunication(timeslot, i, connect)
			setVoteCommunication(timeslot, i, connect)
		}
	}

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

func SelectTwo(n int) (int, int) {
	r1 := rand.Intn(n)
	r2 := rand.Intn(n)
	for {
		if r2 != r1 {
			break
		}
		r2 = rand.Intn(n)
	}
	return r1, r2
}
