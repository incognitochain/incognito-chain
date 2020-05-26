package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"strconv"
	"time"

	"os"

	blockchainv2 "github.com/incognitochain/incognito-chain/blockchain/v2"
	shardv2 "github.com/incognitochain/incognito-chain/blockchain/v2/shard"
	"github.com/incognitochain/incognito-chain/consensus_v2"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbftv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

type Node struct {
	id              string
	consensusEngine *blsbftv2.BLSBFT
	chain           consensus.ChainViewManagerInterface
	nodeList        []*Node
	startSimulation bool
}

type logWriter struct {
	NodeID string
	fd     *os.File
}

func (s logWriter) Write(p []byte) (n int, err error) {
	s.fd.Write(p)
	return len(p), nil
}

var fullnode *blockchainv2.ChainViewManager

func NewNode(committeePkStruct []incognitokey.CommitteePublicKey, committee []string, index int) *Node {
	name := fmt.Sprintf("log%d", index)
	fd, err := os.OpenFile(fmt.Sprintf("%s.log", name), os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	fd.Truncate(0)
	backendLog := common.NewBackend(logWriter{
		NodeID: name,
		fd:     fd,
	})
	consensusLogger := backendLog.Logger("Consensus", false)
	consensusLogger.SetLevel(1)
	chainViewLogger := backendLog.Logger("ShardView", false)
	chainViewLogger.SetLevel(1)

	node := Node{id: fmt.Sprintf("%d", index)}
	db := &FakeDB{}
	db.genesisBlock = shardv2.CreateShardGenesisBlock(1, blockchain.Testnet, blockchain.TestnetGenesisBlockTime, blockchain.TestnetInitPRV)
	node.chain = blockchainv2.InitNewChainViewManager(fmt.Sprintf("shard0_%d", index), &shardv2.ShardView{
		BC:             &shardv2.FakeBC{},
		Block:          db.genesisBlock.(*shardv2.ShardBlock),
		ShardCommittee: committeePkStruct,
		DB:             db,
		Logger:         chainViewLogger,
	})

	if fullnode == nil {
		db := &FakeDB{}
		db.genesisBlock = shardv2.CreateShardGenesisBlock(1, blockchain.Testnet, blockchain.TestnetGenesisBlockTime, blockchain.TestnetInitPRV)
		backendLog := common.NewBackend(nil).Logger("Fullnode", false)
		fullnode = blockchainv2.InitNewChainViewManager("fullnode", &shardv2.ShardView{
			ShardID:        0,
			Block:          db.genesisBlock.(*shardv2.ShardBlock),
			ShardCommittee: committeePkStruct,
			DB:             db,
			Logger:         backendLog,
		})
	}

	//node.chain.UserPubKey = committeePkStruct[index]
	node.chain.GetBestView()

	node.consensusEngine = &blsbftv2.BLSBFT{
		Chain:    node.chain,
		Node:     &node,
		ChainKey: "shard",
		PeerID:   name,
		Logger:   consensusLogger,
	}

	prvSeed, err := blsbftv2.LoadUserKeyFromIncPrivateKey(committee[index])
	failOnError(err)
	failOnError(node.consensusEngine.LoadUserKey(prvSeed))
	return &node
}

func (s *Node) Start() {
	s.consensusEngine.Start()
}

func (s *Node) GetID() string {
	return s.id
}

func (s *Node) RequestSyncBlock(nodeID string, fromView string, toView string) {
	s.consensusEngine.Logger.Debug("Sync block from node", nodeID)
	time.AfterFunc(time.Millisecond*100, func() {
		nodeIDNumber, _ := strconv.Atoi(nodeID)
		views := GetSimulation().nodeList[nodeIDNumber].chain.GetViewByRange(fromView, toView)
		for _, v := range views {
			s.chain.ConnectBlockAndAddView(v.GetBlock())
		}
		s.consensusEngine.Logger.Debug("Sync block ", len(views))
	})

}

func (s *Node) NotifyOutdatedView(nodeID string, latestView string) {
	time.AfterFunc(time.Millisecond*100, func() {
		nodeIDNumber, _ := strconv.Atoi(nodeID)
		views := s.chain.GetViewByRange("", latestView)
		for _, v := range views {
			GetSimulation().nodeList[nodeIDNumber].chain.ConnectBlockAndAddView(v.GetBlock())
			GetSimulation().nodeList[nodeIDNumber].consensusEngine.Logger.Debug("Sync from notify outdated view, block height", v.GetHeight())
		}
	})

}

func (s *Node) BroadCastBlock(block consensus.BlockInterface) {
	fullnode.ConnectBlockAndAddView(block)
}

func (s *Node) PushMessageToChain(msg interface{}, chain consensus.ChainViewManagerInterface) error {
	time.AfterFunc(time.Millisecond*100, func() {
		if msg.(*wire.MessageBFT).Type == "propose" {
			timeSlot := msg.(*wire.MessageBFT).TimeSlot
			if timeSlot > GetSimulation().maxTimeSlot {
				os.Exit(0)
			}
			pComm := GetSimulation().scenario.proposeComm
			if comm, ok := pComm[timeSlot]; ok {
				for i, c := range s.nodeList {
					if s.id == c.id {
						continue
					}
					if senderComm, ok := comm[s.id]; ok {
						if senderComm[i] == 1 {
							s.consensusEngine.Logger.Debug("Send propose to ", c.id)
							c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT), s)
						}
					} else {
						s.consensusEngine.Logger.Debug("Send propose to ", c.id)
						c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT), s)
					}

				}
			} else {
				for _, c := range s.nodeList {
					if s.id == c.id {
						continue
					}
					s.consensusEngine.Logger.Debug("Send propose to ", c.id)
					c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT), s)
				}
			}
			return
		}

		if msg.(*wire.MessageBFT).Type == "vote" {
			vComm := GetSimulation().scenario.voteComm
			timeSlot := msg.(*wire.MessageBFT).TimeSlot
			if timeSlot > GetSimulation().maxTimeSlot {
				os.Exit(0)
			}
			if comm, ok := vComm[timeSlot]; ok {
				for i, c := range s.nodeList {
					if s.id == c.id {
						continue
					}
					if senderComm, ok := comm[s.id]; ok {
						if senderComm[i] == 1 {
							s.consensusEngine.Logger.Debug("Send vote to ", c.id)
							c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT), s)
						}
					} else {
						s.consensusEngine.Logger.Debug("Send vote to ", c.id)
						c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT), s)
					}

				}
			} else {
				for _, c := range s.nodeList {
					if s.id == c.id {
						continue
					}
					s.consensusEngine.Logger.Debug("Send vote to ", c.id)
					c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT), s)
				}
			}
			return
		}
	})

	return nil
}

func (Node) UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string) {
	//not use in bft
	return
}

func (Node) IsEnableMining() bool {
	//not use in bft
	return true
}

func (Node) GetMiningKeys() string {
	//not use in bft
	panic("implement me")
}

func (Node) GetPrivateKey() string {
	//not use in bft
	panic("implement me")
}

func (Node) DropAllConnections() {
	//not use in bft
	return
}

func (Node) PushMessageToPeer(msg interface{}, peerId libp2p.ID) error {
	return nil
}

func main() {

}
