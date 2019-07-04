package consensus

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/consensus/bft"
	"github.com/incognitochain/incognito-chain/consensus/chain"
	"github.com/incognitochain/incognito-chain/wire"
	"strconv"
	"strings"
	"time"
)

const (
	BEACON_CHAINKEY = "beacon"
	SHARD_CHAINKEY  = "shard"
)

type Engine struct {
	ChainList        map[string]chain.ChainInterface
	Blockchain       *blockchain.BlockChain
	ConsensusOnGoing bool
}

var ConsensusManager = Engine{
	ChainList: make(map[string]chain.ChainInterface),
}

func init() {
	go func() {
		ticker := time.Tick(time.Millisecond * 1000)
		for _ = range ticker {
			if ConsensusManager.Blockchain != nil && ConsensusManager.Blockchain.Synker.IsLatest(false, 0) { //beacon synced
				//TODO: start chain if node is in committee
			}
		}
	}()
}

func (s *Engine) Start(node chain.Node, blockchain *blockchain.BlockChain, blockgen *blockchain.BlkTmplGenerator) {
	//start beacon and run consensus engine
	beaconChain, ok := s.ChainList[BEACON_CHAINKEY]
	if !ok {
		bftcore := &bft.BFTCore{ChainKey: BEACON_CHAINKEY, IsRunning: false, UserKeySet: node.GetUserKeySet()}
		beaconChain = &chain.BeaconChain{Blockchain: blockchain, Node: node, BlockGen: blockgen, ConsensusEngine: bftcore}
		bftcore.Chain = beaconChain
		s.ChainList[BEACON_CHAINKEY] = beaconChain
		bftcore.Start()
	}

	//start all active beacon, but not run
	for i := 0; i < node.GetActiveShardNumber(); i++ {
		shardChain, ok := s.ChainList[SHARD_CHAINKEY+""+strconv.Itoa(i)]
		if !ok {
			bftcore := &bft.BFTCore{ChainKey: SHARD_CHAINKEY + "" + strconv.Itoa(i), IsRunning: false, UserKeySet: node.GetUserKeySet()}
			shardChain = &chain.ShardChain{ShardID: byte(i), Blockchain: blockchain, Node: node, BlockGen: blockgen, ConsensusEngine: bftcore}
			bftcore.Chain = shardChain

			s.ChainList[SHARD_CHAINKEY+""+strconv.Itoa(i)] = shardChain
		}
	}
}

func (s *Engine) Stop(name string) error {
	consensusModule, ok := s.ChainList[name]
	if ok && consensusModule.GetConsensusEngine().IsRun() {
		consensusModule.GetConsensusEngine().Stop()
	}
	return nil
}

func (s *Engine) OnBFTMsg(msg wire.Message) {
	switch msg.MessageType() {
	case wire.CmdBFTPropose:
		rawProposeMsg := msg.(*wire.MessageBFTProposeV2)
		if ConsensusManager.ChainList[rawProposeMsg.ChainKey].GetConsensusEngine().IsRun() {
			ConsensusManager.ChainList[rawProposeMsg.ChainKey].GetConsensusEngine().ReceiveProposeMsg(convertProposeMsg(rawProposeMsg))
		}
	case wire.CmdBFTPrepare:
		rawPrepareMsg := msg.(*wire.MessageBFTPrepareV2)
		if ConsensusManager.ChainList[rawPrepareMsg.ChainKey].GetConsensusEngine().IsRun() {
			ConsensusManager.ChainList[rawPrepareMsg.ChainKey].GetConsensusEngine().ReceivePrepareMsg(convertPrepareMsg(rawPrepareMsg))
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
	if strings.Index(msg.ChainKey, BEACON_CHAINKEY) > -1 { //beacon
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
