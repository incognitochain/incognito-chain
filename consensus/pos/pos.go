package pos

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/mining"
	"github.com/ninjadotorg/cash-prototype/wire"
)

// PoSEngine only need to start if node runner want to be a validator

type Engine struct {
	started   int32
	shutdown  int32
	waitgroup sync.WaitGroup
	quit      chan struct{}

	cfg                Config
	CurrentCommittee   []string
	NextBlockCandidate []string
	CurrentLeader      string
}

type Config struct {
	Chain       *blockchain.BlockChain
	ChainParams *blockchain.Params
	BlockGen    *mining.BlkTmplGenerator
	Server      interface {
		// list functions callback which are assigned from Server struct
		PushBlockMessage(*blockchain.Block) error
		PushInvalidBlockMessage(*wire.MessageInvalidBlock) error
		UpdateChain(*blockchain.Block)
	}
}

func (self *Engine) Start() {
	// Already started?
	if atomic.AddInt32(&self.started, 1) != 1 {
		return
	}
	self.quit = make(chan struct{})
	Logger.log.Info("Starting Proof of Stake engine")
	self.waitgroup.Add(1)
	time.AfterFunc(2*time.Second, func() {

	})
}

// Stop gracefully shuts down the sync manager by stopping all asynchronous
// handlers and waiting for them to finish.
func (self *Engine) Stop() error {
	if atomic.AddInt32(&self.shutdown, 1) != 1 {
		Logger.log.Info("Sync manager is already in the process of " +
			"shutting down")
		return nil
	}

	Logger.log.Info("Sync manager shutting down")
	close(self.quit)
	self.waitgroup.Wait()
	return nil
}

func (self *Engine) createBlock() (*blockchain.Block, error) {
	newblock, err := self.cfg.BlockGen.NewBlockTemplate(self.CurrentLeader, self.cfg.Chain)
	if err != nil {
		return newblock.Block, err
	}
	return newblock.Block, nil
}

func (self *Engine) signBlock(block *blockchain.Block) (*blockchain.Block, error) {
	return block, nil
}

func New(cfg *Config) *Engine {
	return &Engine{
		cfg: *cfg,
	}
}
