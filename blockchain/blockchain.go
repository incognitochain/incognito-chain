package blockchain

import (
	"errors"
	"time"
	//"fmt"
	//"time"

	"sync"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
)

type BlockChain struct {
	Config    Config
	Blocks    [][]*Block
	Headers   map[common.Hash]int
	BestBlock *Block

	chainLock sync.RWMutex
}

// Config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	// Db defines the database which houses the blocks and will be used to
	// store all metadata created by this package such as the utxo set.
	//
	// This field is required.
	Db database.DB

	// Interrupt specifies a channel the caller can close to signal that
	// long running operations, such as catching up indexes or performing
	// database migrations, should be interrupted.
	//
	// This field can be nil if the caller does not desire the behavior.
	Interrupt <-chan struct{}

	// chainParams identifies which chain parameters the chain is associated
	// with.
	//
	// This field is required.
	ChainParams *Params
}

func (self *BlockChain) Init(config *Config) (error) {

	self.Headers = make(map[common.Hash]int)
	// self.Blocks = make(map[*common.Hash]*Block)

	// Enforce required config fields.
	// TODO
	//if config.Db == nil {
	//	return nil, errors.New("blockchain.New database is nil")
	//}
	if config.ChainParams == nil {
		return errors.New("blockchain.New chain parameters nil")
	}

	self.Config = *config

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := self.InitChainState(); err != nil {
		return err
	}

	return nil
}

// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
func (self *BlockChain) InitChainState() error {
	// TODO
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool

	if !initialized {
		// At this point the database has not already been initialized, so
		// initialize both it and the chain state to the genesis block.
		return self.createChainState()
	}

	// TODO
	// Attempt to load the chain state from the database.
	return nil
}

// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
func (self *BlockChain) createChainState() error {
	// Create a new block from genesis block and set it as best block of chain
	genesisBlock := self.Config.ChainParams.GenesisBlock
	self.Blocks = make([][]*Block, 20)
	self.Blocks[0] = append(self.Blocks[0], genesisBlock)
	self.Headers[*genesisBlock.Hash()] = 0
	self.BestBlock = genesisBlock

	//err := self.Config.Db.StoreBlock(genesisBlock)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (bc *BlockChain) Reset() error {
	//Todo reset genesis bock logic
	return nil
}
