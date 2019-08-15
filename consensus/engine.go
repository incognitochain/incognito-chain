package consensus

import (
	"errors"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
)

var AvailableConsensus map[string]ConsensusInterface

type Engine struct {
	sync.Mutex
	cQuit              chan struct{}
	started            bool
	Node               nodeInterface
	ChainConsensusList map[string]ConsensusInterface
	Chains             map[string]ChainInterface
	CurrentMiningChain string
	Blockchain         *blockchain.BlockChain
	BlockGen           *blockchain.BlockGenerator
	// userMiningPublicKeys map[string]string
	// MiningKeys         map[string]string

}

func New(node nodeInterface, blockchain *blockchain.BlockChain, blockgen *blockchain.BlockGenerator) *Engine {
	engine := Engine{
		Node:       node,
		Blockchain: blockchain,
		BlockGen:   blockgen,
	}
	err := engine.LoadMiningKeys(node.GetMiningKeys())
	if err != nil {
		panic(err)
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
	Logger.log.Info("starting consensus...")

	engine.cQuit = make(chan struct{})
	go func() {
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				time.Sleep(time.Millisecond * 100)
				// fmt.Println(engine)
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
	if engine.ChainConsensusList[common.BEACON_CHAINKEY].GetConsensusName() != engine.Blockchain.BestState.Beacon.ConsensusAlgorithm {
		consensus, ok := AvailableConsensus[engine.ChainConsensusList[common.BEACON_CHAINKEY].GetConsensusName()]
		if ok {
			engine.ChainConsensusList[common.BEACON_CHAINKEY] = consensus.NewInstance()
		} else {
			panic("Update code please")
		}
	}
	for idx := 0; idx < engine.Blockchain.BestState.Beacon.ActiveShards; idx++ {
		shard, ok := engine.Blockchain.BestState.Shard[byte(idx)]
		if ok {
			chainKey := common.GetShardChainKey(byte(idx))
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

func RegisterConsensus(name string, consensus ConsensusInterface) error {
	if len(AvailableConsensus) == 0 {
		AvailableConsensus = make(map[string]ConsensusInterface)
	}
	AvailableConsensus[name] = consensus
	return nil
}

func (engine *Engine) ValidateBlockWithConsensus(block common.BlockInterface, chainName string, consensusType string) error {
	consensusModule, ok := engine.ChainConsensusList[chainName]
	if ok && !consensusModule.IsOngoing() {
		consensusModule.ValidateBlock(block)
	}
	return nil
}

func (engine *Engine) ValidateBlockCommitteSig(blockHash *common.Hash, committee []string, validationData string, consensusType string) error {
	// return engine.ChainConsensusList[consensusType].ValidateAggregatedSig(blockHash, validationData, committee)
	return nil
}

func (engine *Engine) IsOngoing(chainName string) bool {
	consensusModule, ok := engine.ChainConsensusList[chainName]
	if ok {
		return consensusModule.IsOngoing()
	}
	return false
}

func (engine *Engine) OnBFTMsg(msg *wire.MessageBFT) {
	if engine.CurrentMiningChain == msg.ChainKey {
		engine.ChainConsensusList[msg.ChainKey].ProcessBFTMsg(msg)
	}
}

func (engine *Engine) GetUserRole() (string, int) {
	if engine.CurrentMiningChain != "" {
		publicKey, _ := engine.GetCurrentMiningPublicKey()
		userRole, _ := engine.Chains[engine.CurrentMiningChain].GetPubkeyRole(publicKey, 0)
		if engine.CurrentMiningChain == common.BEACON_CHAINKEY {
			return userRole, -1
		}
		return userRole, int(engine.Chains[engine.CurrentMiningChain].GetShardID())
	}
	return "", 0
}

func (engine *Engine) VerifyData(data []byte, sig string, publicKey string, consensusType string) error {
	return nil
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

func (engine *Engine) ValidateProducerSig(block common.BlockInterface, consensusType string) error {
	return nil
}

// func (engine *Engine) GetUserMiningKey() string {
// 	return ""
// }
