package blockchain

import (
	"github.com/internet-cash/prototype/database"
	"errors"
	"github.com/internet-cash/prototype/common"
)

type BlockChain struct {
	Config    Config
	Blocks    map[*common.Hash]*Block
	Headers   map[*common.Hash]*BlockHeader
	BestBlock *Block
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

	// ChainParams identifies which chain parameters the chain is associated
	// with.
	//
	// This field is required.
	ChainParams *Params
}

func (self BlockChain) New(config *Config) (*BlockChain, error) {
	// Enforce required config fields.
	// TODO
	//if config.Db == nil {
	//	return nil, errors.New("blockchain.New database is nil")
	//}
	if config.ChainParams == nil {
		return nil, errors.New("blockchain.New chain parameters nil")
	}

	self.Config = *config

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := self.InitChainState(); err != nil {
		return nil, err
	}

	return &self, nil
}

// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
func (self *BlockChain) InitChainState() (error) {
	// TODO
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool

	if !initialized {
		// At this point the database has not already been initialized, so
		// initialize both it and the chain state to the genesis block.
		return self.CreateChainState()
	}

	// TODO
	// Attempt to load the chain state from the database.
	return nil
}

// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
func (self *BlockChain) CreateChainState() error {
	// TODO something
	genesisBlock := self.Config.ChainParams.GenesisBlock
	self.Blocks[genesisBlock.Hash()] = genesisBlock
	self.Headers[genesisBlock.Hash()] = &genesisBlock.Header
	self.BestBlock = genesisBlock
	return nil
}
