package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/consensus_multi/signatureschemes"
	"os"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_multi/blsbftv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
	libp2p "github.com/libp2p/go-libp2p-peer"
)

type Node struct {
	id              string
	consensusEngine *blsbftv2.BLSBFT_V2
	chain           *Chain
	nodeList        []*Node
}

type logWriter struct {
	NodeID string
	fd     *os.File
}

func (s logWriter) Write(p []byte) (n int, err error) {
	s.fd.Write(p)
	return len(p), nil
}

func NewNode(committeePkStruct []incognitokey.CommitteePublicKey, miningKey *signatureschemes.MiningKey, committee []string, index int) *Node {
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
	node.chain = NewChain(0, "shard0", committeePkStruct)

	node.consensusEngine = &blsbftv2.BLSBFT_V2{
		Chain:    node.chain,
		Node:     &node,
		ChainKey: "shard",
		PeerID:   name,
		Logger:   consensusLogger,
	}
	node.consensusEngine.LoadUserKeys([]signatureschemes.MiningKey{*miningKey})
	return &node
}

func (s *Node) PushMessageToChain(msg wire.Message, chain common.ChainInterface) error {
	time.AfterFunc(time.Millisecond*100, func() {
		if msg.(*wire.MessageBFT).Type == "propose" {
			timeSlot := uint64(msg.(*wire.MessageBFT).TimeSlot)
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
							c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
						}
					} else {
						s.consensusEngine.Logger.Debug("Send propose to ", c.id)
						c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
					}

				}
			} else {
				for _, c := range s.nodeList {
					if s.id == c.id {
						continue
					}
					s.consensusEngine.Logger.Debug("Send propose to ", c.id)
					c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
				}
			}
			return
		}

		if msg.(*wire.MessageBFT).Type == "vote" {
			vComm := GetSimulation().scenario.voteComm
			timeSlot := uint64(msg.(*wire.MessageBFT).TimeSlot)
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
							c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
						}
					} else {
						s.consensusEngine.Logger.Debug("Send vote to ", c.id)
						c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
					}

				}
			} else {
				for _, c := range s.nodeList {
					if s.id == c.id {
						continue
					}
					s.consensusEngine.Logger.Debug("Send vote to ", c.id)
					c.consensusEngine.ProcessBFTMsg(msg.(*wire.MessageBFT))
				}
			}
			return
		}
	})

	return nil
}

func (s *Node) RequestMissingViewViaStream(peerID string, hashes [][]byte, fromCID int, chainName string) (err error) {
	str := []string{}
	for _, h := range hashes {
		pH, _ := common.Hash{}.NewHash(h)
		str = append(str, pH.String())
	}
	fmt.Println("RequestMissingViewViaStream from ", peerID, strings.Join(str, ","))
	return nil
}

func (s *Node) GetSelfPeerID() libp2p.ID {
	return libp2p.ID(s.id)
}

func (s *Node) Start() {
	fmt.Printf("Node %s log is %s, peerID: %v \n", s.id, fmt.Sprintf("log%s.log", s.id), libp2p.ID(s.id))
	s.consensusEngine.Start()
}
