package consensus

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/bft"
	"github.com/incognitochain/incognito-chain/wire"
	"strings"
	"time"
)

type ChainInterface interface {
	PushMessageToValidator(wire.Message) error
	GetLastBlockTimeStamp() uint64
	GetBlkMinTime() time.Duration
	IsReady() bool
	GetHeight() uint64
	GetCommitteeSize() int
	GetNodePubKeyIndex() int
	GetLastProposerIndex() int
	GetNodePubKey() string
	CreateNewBlock() BlockInterface
	ValidateBlock(interface{}) bool
	ValidateSignature(interface{}, string) bool
	InsertBlk(interface{}, bool)
}

type BlockInterface interface {
	GetHeight() uint64
	GetProducerPubKey() string
	Hash() *common.Hash
}

type ProtocolInterface interface {
	GetInfo() string

	Start()
	Stop()
	IsRun() bool

	ReceiveProposeMsg(ProposeMsg)
	ReceivePrepareMsg(PrepareMsg)
}

type ProposeMsg struct {
	ChainKey   string
	Block      BlockInterface
	ContentSig string
	Pubkey     string
	Timestamp  int64
	RoundKey   string
}

type PrepareMsg struct {
	ChainKey   string
	IsOk       bool
	Pubkey     string
	ContentSig string
	BlkHash    string
	RoundKey   string
	Timestamp  int64
}

type Engine struct {
	ChainList map[string]ProtocolInterface
}

var ConsensusEngine = Engine{
	ChainList: make(map[string]ProtocolInterface),
}

func (s *Engine) Start(name string, chain ChainInterface) ProtocolInterface {
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

func convertProposeMsg(msg *wire.MessageBFTProposeV2) ProposeMsg {
	proposeMsg := ProposeMsg{
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

func convertPrepareMsg(msg *wire.MessageBFTPrepareV2) PrepareMsg {
	prepareMsg := PrepareMsg{
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
