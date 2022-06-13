package remotetests

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"github.com/spf13/viper"
	"time"
)

type HostConfig struct {
	Hosts []string `mapstructure:"hosts"`
	Seed  []string `mapstructure:"miningkeys"`
}

type NodeClient struct {
	Miningkey signatureschemes.MiningKey
	RPCClient devframework.RemoteRPCClient
}

type NodeManager struct {
	Config              HostConfig            `json:"Config"`
	BeaconNode          []NodeClient          `json:"Beacon"`
	ShardFixNode        map[int][]NodeClient  `json:"ShardFixNode"`
	CommitteePublicKeys map[string]NodeClient `json:"CPK"`
}

func readFileConfig() HostConfig {
	viper.SetConfigName("hosts") // name of config file (without extension)
	viper.SetConfigType("yaml")  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("./")    // path to look for the config file in
	err := viper.ReadInConfig()  // Find and read the config file
	if err != nil {              // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
	c := &HostConfig{}
	viper.Unmarshal(c)
	return *c
}

func (s *NodeManager) initNodeManager() {

	miningKeyMap := map[string]*signatureschemes.MiningKey{}
	nodeMap := map[string]NodeClient{}
	committeePublicKeys := map[string]NodeClient{}
	for _, seed := range s.Config.Seed {
		miningKey, err := consensus_v2.GetMiningKeyFromPrivateSeed(seed)
		if err != nil {
			panic(err)
		}
		miningKeyMap[miningKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus)] = miningKey
	}

	beaconHosts := []NodeClient{}
	for _, host := range s.Config.Hosts {
		rpcClient := devframework.RemoteRPCClient{host}
		node := NodeClient{
			RPCClient: rpcClient,
		}
		miningInfo, err := rpcClient.GetMiningInfo()
		if err != nil {
			fmt.Println("Error get mining info with host", host)
			panic(err)
		}
		if _, ok := miningKeyMap[miningInfo.MiningPublickey]; !ok {
			fmt.Println("Cannot miningkey for host", host)
			time.Sleep(time.Second)
		} else {
			node.Miningkey = *miningKeyMap[miningInfo.MiningPublickey]
		}
		if miningInfo.Layer == "beacon" {
			beaconHosts = append(beaconHosts, node)
		}
		nodeMap[miningInfo.MiningPublickey] = node
	}

	beaconView, err := beaconHosts[0].RPCClient.GetBeaconBestState()
	if err != nil {
		panic(err)
	}

	for _, cpkStr := range beaconView.BeaconCommittee {
		cpk, _ := account.ParseCommitteePubkey(cpkStr)
		committeePublicKeys[cpkStr] = nodeMap[cpk.GetMiningKeyBase58(common.BlsConsensus)]
		s.BeaconNode = append(s.BeaconNode, nodeMap[cpk.GetMiningKeyBase58(common.BlsConsensus)])
	}
	for sid, committees := range beaconView.ShardCommittee {
		for id, cpkStr := range committees {
			cpk, _ := account.ParseCommitteePubkey(cpkStr)
			committeePublicKeys[cpkStr] = nodeMap[cpk.GetMiningKeyBase58(common.BlsConsensus)]
			if id < beaconView.MinShardCommitteeSize {
				s.ShardFixNode[int(sid)] = append(s.ShardFixNode[int(sid)], nodeMap[cpk.GetMiningKeyBase58(common.BlsConsensus)])
			}
		}
	}
	s.CommitteePublicKeys = committeePublicKeys
	return
}

func NewRemoteNodeManager() NodeManager {
	fmt.Println("reading config ...")
	config := readFileConfig()
	fmt.Println("map mining key and host ...")
	nodeManager := NodeManager{
		Config:              config,
		CommitteePublicKeys: make(map[string]NodeClient),
		ShardFixNode:        make(map[int][]NodeClient),
	}
	nodeManager.initNodeManager()
	return nodeManager
}
