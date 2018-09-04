package pos

import (
	"encoding/binary"
	"errors"
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ninjadotorg/cash-prototype/mempool"

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
	MemPool     *mempool.TxPool
	Server      interface {
		// list functions callback which are assigned from Server struct
		PushBlockMessage(*blockchain.Block) error
		PushBlockSignature(*wire.MessageSignedBlock) error
		PushRequestSignBlock(*blockchain.Block) error
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

func (self *Engine) createBlock(chainID byte) (*blockchain.Block, error) {
	newblock, err := self.cfg.BlockGen.NewBlockTemplate(self.CurrentLeader, self.cfg.Chain)
	if err != nil {
		return newblock.Block, err
	}
	return newblock.Block, nil
}

func (self *Engine) signBlock(block *blockchain.Block) (*blockchain.Block, error) {
	return block, nil
}

func (self *Engine) getChainValidators(chainID byte) ([]string, error) {
	var validators []string
	for index := 1; index <= 11; index++ {
		validatorID := math.Mod((index + int(chainID)), 21)
		validators = append(validators, self.CurrentCommittee[validatorID])
	}
	if len(validators) == 11 {
		return validators, nil
	}
	return nil, errors.New("can't get chain's validators")
}

func (self *Engine) getSenderChain(senderAddress string) (byte, error) {
	addrBig := new(big.Int)
	addrBig.SetBytes([]byte(senderAddress))

	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(100))
	modNum := new(big.Int)
	modNum.SetBytes(b)

	modResult := new(big.Int)
	modResult = modResult.Mod(addrBig, modNum)

	for index := uint64(0); index < 5; index++ {
		if (modResult.Uint64()-index)%5 == 0 {
			return 0, (modResult.Uint64() - index) / 5
		}
	}

	return nil, errors.New("can't get sender's chainID")

}

func (self *Engine) OnRequestSign(block *blockchain.Block) {
	return
}

// func (self *Engine) filterTx() []*transaction.Transaction {

// }

func New(cfg *Config) *Engine {
	return &Engine{
		cfg: *cfg,
	}
}
