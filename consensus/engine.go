package consensus

import (
	"github.com/incognitochain/incognito-chain/consensus/bft"
	"github.com/incognitochain/incognito-chain/wire"
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
	GetRound() uint64
	GetProducer() string
	Hash() string
}

type ProtocolInterface interface {
	GetInfo() string
	Start()
	Stop()
	IsRun() bool
	ReceiveMsg()
}

type Engine struct {
	ChainList map[string]ProtocolInterface
}

var ConsensusEngine = Engine{
	ChainList: make(map[string]ProtocolInterface),
}

func (s *Engine) Start(name string, chain ChainInterface, protocol string) ProtocolInterface {
	consensusModule, ok := s.ChainList[name]
	if ok {
		if !consensusModule.IsRun() {
			consensusModule.Start()
		}
		return consensusModule
	}

	var bftcore ProtocolInterface
	switch protocol {
	case "BFT":
		bftcore = &bft.BFTCore{Name: name, IsRunning: false}
		s.ChainList[name] = bftcore
		bftcore.Start()
	}

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
		proposeMsg := msg.(*wire.MessageBFTProposeV2)
		if ConsensusEngine.ChainList[proposeMsg.ChainKey].IsRun() {
			ConsensusEngine.ChainList[proposeMsg.ChainKey]
		}
	case wire.CmdBFTPrepare:
		prepareMsg := msg.(*wire.MessageBFTProposeV2)

	}
	return
}
