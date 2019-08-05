package consensus

import (
	"errors"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/consensus/chain"
	"github.com/incognitochain/incognito-chain/wire"
)

const (
	BEACON_CHAINKEY = "beacon"
	SHARD_CHAINKEY  = "shard"
)

var AvailableConsensus map[string]chain.ConsensusInterface

type Engine struct {
	sync.Mutex
	cQuit              chan struct{}
	started            bool
	Node               chain.Node
	ChainConsensusList map[string]chain.ConsensusInterface
	Blockchain         *blockchain.BlockChain
	BlockGen           *blockchain.BlockGenerator
}

func New(node chain.Node, blockchain *blockchain.BlockChain, blockgen *blockchain.BlockGenerator) *Engine {
	engine := Engine{
		Node:       node,
		Blockchain: blockchain,
		BlockGen:   blockgen,
	}
	return &engine
}

//watchConsensusState will watch MiningKey Role as well as chain consensus type
func (engine *Engine) watchConsensusState() {

}

func (engine *Engine) Start() error {
	engine.Lock()
	defer engine.Unlock()
	if engine.started {
		return errors.New("Consensus engine is already started")
	}
	if engine.Node.GetMiningKeys() == "" {
		return errors.New("MiningKeys can't be empty")
	}
	engine.cQuit = make(chan struct{})
	go func() {
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				time.Sleep(time.Millisecond * 100)

				// if !engine.config.BlockChain.Synker.IsLatest(false, 0) {
				// 	userRole, shardID := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.userPk, 0)
				// 	if userRole == common.SHARD_ROLE {
				// 		go engine.NotifyShardRole(int(shardID))
				// 		go engine.NotifyBeaconRole(false)
				// 	} else {
				// 		if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
				// 			go engine.NotifyBeaconRole(true)
				// 			go engine.NotifyShardRole(-1)
				// 		}
				// 	}
				// } else {
				// 	if !engine.config.Server.IsEnableMining() {
				// 		time.Sleep(time.Second * 1)
				// 		continue
				// 	}
				// 	userRole, shardID := engine.config.BlockChain.BestState.Beacon.GetPubkeyRole(engine.userPk, engine.currentBFTRound)
				// 	if engine.config.NodeMode == common.NODEMODE_BEACON && userRole == common.SHARD_ROLE {
				// 		userRole = common.EmptyString
				// 	}
				// 	if engine.config.NodeMode == common.NODEMODE_SHARD && userRole != common.SHARD_ROLE {
				// 		userRole = common.EmptyString
				// 	}
				// 	engine.userLayer = userRole
				// 	switch userRole {
				// 	case common.VALIDATOR_ROLE, common.PROPOSER_ROLE:
				// 		engine.userLayer = common.BEACON_ROLE
				// 	}
				// 	engine.config.Server.UpdateConsensusState(engine.userLayer, engine.userPk, nil, engine.config.BlockChain.BestState.Beacon.BeaconCommittee, engine.config.BlockChain.BestState.Beacon.GetShardCommittee())
				// 	switch engine.userLayer {
				// 	case common.BEACON_ROLE:
				// 		if engine.config.NodeMode == common.NODEMODE_BEACON || engine.config.NodeMode == common.NODEMODE_AUTO {
				// 			engine.config.BlockChain.ConsensusOngoing = true
				// 			engine.execBeaconRole()
				// 			engine.config.BlockChain.ConsensusOngoing = false
				// 		}
				// 	case common.SHARD_ROLE:
				// 		if engine.config.NodeMode == common.NODEMODE_SHARD || engine.config.NodeMode == common.NODEMODE_AUTO {
				// 			if engine.config.BlockChain.Synker.IsLatest(true, shardID) {
				// 				engine.config.BlockChain.ConsensusOngoing = true
				// 				engine.execShardRole(shardID)
				// 				engine.config.BlockChain.ConsensusOngoing = false
				// 			}
				// 		}
				// 	case common.EmptyString:
				// 		time.Sleep(time.Second * 1)
				// 	}
				// }
			}
		}
	}()
	return nil
}

func (engine *Engine) Stop(name string) error {
	engine.Lock()
	defer engine.Unlock()
	if !engine.started {
		return errors.New("Consensus engine is already stopped")
	}
	engine.started = false
	close(engine.cQuit)
	return nil
}

func (engine *Engine) SwitchConsensus(chainkey string, consensus string) error {
	if engine.ChainConsensusList[BEACON_CHAINKEY].GetConsensusName() != engine.Blockchain.BestState.Beacon.ConsensusAlgorithm {
		consensus, ok := AvailableConsensus[engine.ChainConsensusList[BEACON_CHAINKEY].GetConsensusName()]
		if ok {
			engine.ChainConsensusList[BEACON_CHAINKEY] = consensus.NewInstance()
		} else {
			panic("Update code please")
		}
	}
	for idx := 0; idx < engine.Blockchain.BestState.Beacon.ActiveShards; idx++ {
		shard, ok := engine.Blockchain.BestState.Shard[byte(idx)]
		if ok {
			chainKey := GetShardChainKey(byte(idx))
			if shard.ConsensusAlgorithm != engine.ChainConsensusList[chainKey].GetConsensusName() {
				consensus, ok := AvailableConsensus[engine.ChainConsensusList[chainKey].GetConsensusName()]
				if ok {
					engine.ChainConsensusList[chainKey] = consensus.NewInstance()
				} else {
					panic("Update code please")
				}
			}
		} else {
			panic("Oops... Maybe a bug cause this, please update code")
		}
	}
	return nil
}

func GetShardChainKey(shardID byte) string {
	return SHARD_CHAINKEY + "-" + string(shardID)
}

func RegisterConsensus(name string, consensus chain.ConsensusInterface) error {
	AvailableConsensus[name] = consensus
	return nil
}

func (engine *Engine) ValidateBlockWithConsensus(block chain.BlockInterface, chainName string, consensusType string) error {
	consensusModule, ok := engine.ChainConsensusList[chainName]
	if ok && !consensusModule.IsOngoing() {
		consensusModule.ValidateBlock(block)
	}
	return nil
}

func (engine *Engine) IsOngoing(chainName string) bool {
	consensusModule, ok := engine.ChainConsensusList[chainName]
	if ok {
		return consensusModule.IsOngoing()
	}
	return false
}

func (s *Engine) OnBFTMsg(msg wire.Message) {
	// switch msg.MessageType() {
	// case wire.CmdBFTPropose:
	// 	rawProposeMsg := msg.(*wire.MessageBFTProposeV2)
	// 	if ConsensusManager.ChainList[rawProposeMsg.ChainKey].GetConsensusEngine().IsRun() {
	// 		ConsensusManager.ChainList[rawProposeMsg.ChainKey].GetConsensusEngine().ReceiveProposeMsg(convertProposeMsg(rawProposeMsg))
	// 	}
	// case wire.CmdBFTPrepare:
	// 	rawPrepareMsg := msg.(*wire.MessageBFTPrepareV2)
	// 	if ConsensusManager.ChainList[rawPrepareMsg.ChainKey].GetConsensusEngine().IsRun() {
	// 		ConsensusManager.ChainList[rawPrepareMsg.ChainKey].GetConsensusEngine().ReceivePrepareMsg(convertPrepareMsg(rawPrepareMsg))
	// 	}
	// }
}

func (s *Engine) GetUserRole() (string, int) {
	return "", 0
}

// func convertProposeMsg(msg *wire.MessageBFTProposeV2) bft.ProposeMsg {
// 	proposeMsg := bft.ProposeMsg{
// 		ChainKey:   msg.ChainKey,
// 		ContentSig: msg.ContentSig,
// 		Pubkey:     msg.Pubkey,
// 		Timestamp:  msg.Timestamp,
// 		RoundKey:   msg.RoundKey,
// 	}
// 	if strings.Index(msg.ChainKey, BEACON_CHAINKEY) > -1 { //beacon
// 		blk := &blockchain.BeaconBlock{}
// 		err := json.Unmarshal([]byte(msg.Block), &blk)
// 		if err != nil {
// 			fmt.Println("BFT: unmarshal beacon propose msg fail", err)
// 		}
// 		proposeMsg.Block = blk
// 	} else { //shard
// 		blk := &blockchain.ShardBlock{}
// 		err := json.Unmarshal([]byte(msg.Block), &blk)
// 		if err != nil {
// 			fmt.Println("BFT: unmarshal shard propose msg fail", err)
// 		}
// 		proposeMsg.Block = blk
// 	}
// 	return proposeMsg
// }

// func convertPrepareMsg(msg *wire.MessageBFTPrepareV2) bft.PrepareMsg {
// 	prepareMsg := bft.PrepareMsg{
// 		ChainKey:   msg.ChainKey,
// 		ContentSig: msg.ContentSig,
// 		Pubkey:     msg.Pubkey,
// 		Timestamp:  msg.Timestamp,
// 		RoundKey:   msg.RoundKey,
// 		IsOk:       msg.IsOk,
// 		BlkHash:    msg.BlkHash,
// 	}
// 	return prepareMsg
// }
