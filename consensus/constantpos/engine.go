package constantpos

import (
	"sync"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/connmanager"
	"github.com/ninjadotorg/constant/mempool"
	"github.com/ninjadotorg/constant/wire"
)

type Engine struct {
	sync.Mutex
	started bool

	// channel
	cQuit      chan struct{}
	cBlockSig  chan blockSig
	cQuitSwap  chan struct{}
	cSwapChain chan byte
	cSwapSig   chan swapSig
	cNewBlock  chan blockchain.Block

	config EngineConfig
	Layers struct {
		Beacon *Layerbeacon
		Shard  *Layershard
	}
	CurrentRole string
}

type EngineConfig struct {
	BlockChain  *blockchain.BlockChain
	ConnManager *connmanager.ConnManager
	ChainParams *blockchain.Params
	BlockGen    *blockchain.BlkTmplGenerator
	MemPool     *mempool.TxPool
	UserKeySet  cashec.KeySet
	Server      interface {
		// list functions callback which are assigned from Server struct
		GetPeerIDsFromPublicKey(string) []libp2p.ID
		PushMessageToAll(wire.Message) error
		PushMessageToPeer(wire.Message, libp2p.ID) error
		PushMessageGetChainState() error
	}
}

//Init apply configuration to consensus engine
func (self Engine) Init(cfg *EngineConfig) (*Engine, error) {
	return &Engine{
		config: *cfg,
	}, nil
}

func (self *Engine) Start() error {
	return nil
}

func (self *Engine) Stop() {

}

func (self *Engine) UpdateChain(block *blockchain.Block) {
	err := self.config.BlockChain.ConnectBlock(block)
	if err != nil {
		Logger.log.Error(err)
		return
	}

	// update tx pool
	for _, tx := range block.Transactions {
		self.config.MemPool.RemoveTx(tx)
	}

	// update candidate list
	err = self.config.BlockChain.BestState[block.Header.ChainID].Update(block)
	if err != nil {
		Logger.log.Errorf("Can not update merkle tree for block: %+v", err)
		return
	}
	self.config.BlockChain.StoreBestState(block.Header.ChainID)

	// self.knownChainsHeight.Lock()
	// if self.knownChainsHeight.Heights[block.Header.ChainID] < int(block.Header.Height) {
	// 	self.knownChainsHeight.Heights[block.Header.ChainID] = int(block.Header.Height)
	// 	self.sendBlockMsg(block)
	// }
	// self.knownChainsHeight.Unlock()
	// self.validatedChainsHeight.Lock()
	// self.validatedChainsHeight.Heights[block.Header.ChainID] = int(block.Header.Height)
	// self.validatedChainsHeight.Unlock()

	// self.Committee().UpdateCommitteePoint(block.BlockProducer, block.Header.BlockCommitteeSigs)
}

func (self *Engine) Committee() *CommitteeStruct {
	return &CommitteeStruct{}
}

func (self *Engine) createTmplBlock() *blockchain.BlockV2 {

	return nil
}
