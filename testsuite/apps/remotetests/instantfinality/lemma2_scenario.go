package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/apps/remotetests"
	"time"
)

type ChainCommittee struct {
	beacon []string
	shard  map[byte][]string
}

func GetChainCommittee(beaconNode remotetests.NodeClient) ChainCommittee {
	bs, _ := beaconNode.RPCClient.GetBeaconBestState()
	return ChainCommittee{
		beacon: bs.BeaconCommittee,
		shard:  bs.ShardCommittee,
	}
}

func Lemma2ScenarioTest(nodeManager remotetests.NodeManager) {
	chainCommittee := GetChainCommittee(nodeManager.BeaconNode[0])
	finish := false

	//beacon re-propose
	fmt.Println("Set beacon consensus rule to no vote ...")
	for _, beaconCpk := range chainCommittee.beacon {
		client := nodeManager.CommitteePublicKeys[beaconCpk].RPCClient
		client.SetConsensusRule(devframework.ConsensusRule{VoteRule: blsbft.VOTE_RULE_NO_VOTE})
	}
	go func() {
		beaconNode := nodeManager.BeaconNode[0]
		shardNode0 := nodeManager.ShardFixNode[0][0]
		shardNode1 := nodeManager.ShardFixNode[1][0]
		for !finish {
			go func() {
				v, _ := beaconNode.RPCClient.GetAllViewDetail(-1)
				if len(v) > 1 {

					panic("More than 1 view")
				}
				v, _ = shardNode0.RPCClient.GetAllViewDetail(-1)
				if len(v) > 1 {
					panic("More than 1 view")
				}
				v, _ = shardNode1.RPCClient.GetAllViewDetail(-1)
				if len(v) > 1 {
					panic("More than 1 view")
				}
			}()
			time.Sleep(time.Second * 2)
		}
	}()
	time.Sleep(time.Second * 60)
	fmt.Println("Set beacon consensus rule to normal ...")
	for _, beaconCpk := range chainCommittee.beacon {
		client := nodeManager.CommitteePublicKeys[beaconCpk].RPCClient
		client.SetConsensusRule(devframework.ConsensusRule{VoteRule: blsbft.VOTE_RULE_VOTE})
	}
	time.Sleep(time.Second * 30)

	//shard re-propose
	fmt.Println("Set shard consensus rule to no vote ...")
	for _, shardCpk := range chainCommittee.shard[0] {
		client := nodeManager.CommitteePublicKeys[shardCpk].RPCClient
		client.SetConsensusRule(devframework.ConsensusRule{VoteRule: blsbft.VOTE_RULE_NO_VOTE})
	}
	go func() {
		beaconNode := nodeManager.BeaconNode[0]
		shardNode0 := nodeManager.ShardFixNode[0][0]
		for !finish {
			go func() {
				v, _ := beaconNode.RPCClient.IsInstantFinality(0)
				if !v {
					panic("Not instant finality")
				}
				v, _ = shardNode0.RPCClient.IsInstantFinality(0)
				if !v {
					panic("Not instant finality")
				}
			}()
			time.Sleep(time.Second * 2)
		}
	}()
	time.Sleep(time.Second * 60)
	fmt.Println("Set shard consensus rule to normal ...")
	for _, shardCpk := range chainCommittee.shard[0] {
		client := nodeManager.CommitteePublicKeys[shardCpk].RPCClient
		client.SetConsensusRule(devframework.ConsensusRule{VoteRule: blsbft.VOTE_RULE_VOTE})
	}
	time.Sleep(time.Second * 30)
	finish = true
	time.Sleep(time.Second * 50)

}
