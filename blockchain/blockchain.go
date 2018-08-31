package blockchain

import (
	"errors"
	//"fmt"
	//"time"

	"sync"

	"github.com/ninjadotorg/cash-prototype/database"
	"time"
	"encoding/json"
	"github.com/ninjadotorg/cash-prototype/common"
)

type BlockChain struct {
	Config Config
	//Blocks []*Block
	//Headers   map[common.Hash]int
	BestState *BestState

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
	// Enforce required config fields.
	// TODO
	if config.Db == nil {
		return errors.New("blockchain.New database is nil")
	}
	if config.ChainParams == nil {
		return errors.New("blockchain.New chain parameters nil")
	}

	//self.Headers = make(map[common.Hash]int)
	// self.Blocks = make(map[*common.Hash]*Block)

	self.Config = *config

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := self.initChainState(); err != nil {
		return err
	}

	Logger.log.Infof("Chain state (height %d, hash %v, totaltx %d)", self.BestState.Height, self.BestState.BestBlockHash.String(), self.BestState.TotalTxns, )

	return nil
}

// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
func (self *BlockChain) initChainState() error {
	// TODO
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool
	bestStateBytes, err := self.Config.Db.FetchBestBlock()
	if err == nil {
		err = json.Unmarshal(bestStateBytes, &self.BestState)
		if err != nil {
			initialized = false
		} else {
			initialized = true
		}
	} else {
		initialized = false
	}

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
	genesisBlock.Height = 0
	//self.Blocks = append(self.Blocks, genesisBlock)

	// Initialize the state related to the best block.  Since it is the
	// genesis block, use its timestamp for the median time.
	numTxns := uint64(len(genesisBlock.Transactions))
	//blockSize := uint64(genesisBlock.SerializeSize())
	//blockWeight := uint64(GetBlockWeight(genesisBlock))
	self.BestState = &BestState{}
	self.BestState.Init(genesisBlock, 0, 0, numTxns, numTxns, time.Unix(genesisBlock.Header.Timestamp.Unix(), 0))

	// store block genesis
	err := self.StoreBlock(genesisBlock)
	if err != nil {
		return err
	}

	// store best state
	err = self.StoreBestState()
	if err != nil {
		return err
	}

	// store block hash by index and index by block hash
	err = self.StoreBlockIndex(genesisBlock)

	return err
}

/**
Get block index(height) of block
 */
func (self *BlockChain) GetBlockHeightByBlockHash(hash *common.Hash) (int32, error) {
	return self.Config.Db.GetIndexOfBlock(hash)
}

/**
Get block hash by block index(height)
 */
func (self *BlockChain) GetBlockHashByBlockHeight(height int32) (*common.Hash, error) {
	return self.Config.Db.GetBlockByIndex(height)
}

/**
Fetch DB and get block by index(height) of block
 */
func (self *BlockChain) GetBlockByBlockHeight(height int32) (*Block, error) {
	hashBlock, err := self.Config.Db.GetBlockByIndex(height)
	if err != nil {
		return nil, err
	}
	blockBytes, err := self.Config.Db.FetchBlock(hashBlock)
	if err != nil {
		return nil, err
	}
	block := Block{}
	err = json.Unmarshal(blockBytes, &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

/**
Fetch DB and get block data by block hash
 */
func (self *BlockChain) GetBlockByBlockHash(hash *common.Hash) (*Block, error) {
	blockBytes, err := self.Config.Db.FetchBlock(hash)
	if err != nil {
		return nil, err
	}
	block := Block{}
	err = json.Unmarshal(blockBytes, &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

/**
Store best state of block(best block, num of tx, ...) into Database
 */
func (self *BlockChain) StoreBestState() (error) {
	return self.Config.Db.StoreBestBlock(self.BestState)
}

/**
Store block into Database
 */
func (self *BlockChain) StoreBlock(block *Block) error {
	return self.Config.Db.StoreBlock(block)
}

/**
Save index(height) of block by block hash
and
Save block hash by index(height) of block
 */
func (self *BlockChain) StoreBlockIndex(block *Block) error {
	return self.Config.Db.StoreBlockIndex(block.Hash(), block.Height)
}

/**
Get all blocks in chain
Return block array
 */
func (self *BlockChain) GetAllBlocks() ([]*Block, error) {
	result := make([]*Block, 0)
	data, err := self.Config.Db.FetchAllBlocks()
	if err != nil {
		return nil, err
	}

	for _, item := range data {
		blockBytes, err := self.Config.Db.FetchBlock(item)
		if err != nil {
			return nil, err
		}
		block := Block{}
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			return nil, err
		}
		result = append(result, &block)
	}
	return result, nil
}

/**
Get all hash of blocks in chain
Return hashes array
 */
func (self *BlockChain) GetAllHashBlocks() ([]*common.Hash, error) {
	data, err := self.Config.Db.FetchAllBlocks()
	if err != nil {
		return nil, err
	}
	return data, err
}
