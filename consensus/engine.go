package consensus

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/consensus/bft"
	"github.com/incognitochain/incognito-chain/wire"
	"strings"
	"time"
)

type ProtocolInterface interface {
	GetInfo() string

	Start()
	Stop()
	IsRun() bool

	ReceiveProposeMsg(interface{})
	ReceivePrepareMsg(interface{})
}

type Engine struct {
	ChainList        map[string]ProtocolInterface
	Blockchain       *blockchain.BlockChain
	ConsensusOnGoing bool
}

var ConsensusEngine = Engine{
	ChainList: make(map[string]ProtocolInterface),
}

func init() {
	go func() {
		ticker := time.Tick(time.Millisecond * 1000)
		for _ = range ticker {
			if ConsensusEngine.Blockchain != nil && ConsensusEngine.Blockchain.Synker.IsLatest(false, 0) { //beacon synced

			}
		}
	}()
}

func (s *Engine) Start(server Node, blockchain blockchain.BlockChain) ProtocolInterface {

	//start beacon

	//start shard
	consensusModule, ok := s.ChainList[name]
	if ok {
		if !consensusModule.IsRun() {
			consensusModule.Start()
		}
		return consensusModule
	}

	bftcore := &bft.BFTCore{Name: name, IsRunning: false}
	s.ChainList[name] = bftcore
	bftcore.Start()
	return bftcore
}

func (s *Engine) Stop(name string) error {
	consensusModule, ok := s.ChainList[name]
	if ok && consensusModule.IsRun() {
		consensusModule.Stop()
	}
	return nil
}

func (s *Engine) OnBFTMsg(msg wire.Message) {
	switch msg.MessageType() {
	case wire.CmdBFTPropose:
		rawProposeMsg := msg.(*wire.MessageBFTProposeV2)
		if ConsensusEngine.ChainList[rawProposeMsg.ChainKey].IsRun() {
			ConsensusEngine.ChainList[rawProposeMsg.ChainKey].ReceiveProposeMsg(convertProposeMsg(rawProposeMsg))
		}
	case wire.CmdBFTPrepare:
		rawPrepareMsg := msg.(*wire.MessageBFTPrepareV2)
		if ConsensusEngine.ChainList[rawPrepareMsg.ChainKey].IsRun() {
			ConsensusEngine.ChainList[rawPrepareMsg.ChainKey].ReceivePrepareMsg(convertPrepareMsg(rawPrepareMsg))
		}
	}

}

func convertProposeMsg(msg *wire.MessageBFTProposeV2) bft.ProposeMsg {
	proposeMsg := bft.ProposeMsg{
		ChainKey:   msg.ChainKey,
		ContentSig: msg.ContentSig,
		Pubkey:     msg.Pubkey,
		Timestamp:  msg.Timestamp,
		RoundKey:   msg.RoundKey,
	}
	if strings.Index(msg.ChainKey, "beacon") > -1 { //beacon
		blk := &blockchain.BeaconBlock{}
		err := json.Unmarshal([]byte(msg.Block), &blk)
		if err != nil {
			fmt.Println("BFT: unmarshal beacon propose msg fail", err)
		}
		proposeMsg.Block = blk
	} else { //shard
		blk := &blockchain.ShardBlock{}
		err := json.Unmarshal([]byte(msg.Block), &blk)
		if err != nil {
			fmt.Println("BFT: unmarshal shard propose msg fail", err)
		}
		proposeMsg.Block = blk
	}
	return proposeMsg
}

func convertPrepareMsg(msg *wire.MessageBFTPrepareV2) bft.PrepareMsg {
	prepareMsg := bft.PrepareMsg{
		ChainKey:   msg.ChainKey,
		ContentSig: msg.ContentSig,
		Pubkey:     msg.Pubkey,
		Timestamp:  msg.Timestamp,
		RoundKey:   msg.RoundKey,
		IsOk:       msg.IsOk,
		BlkHash:    msg.BlkHash,
	}
	return prepareMsg
}
