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
	"github.com/ninjadotorg/cash-prototype/transaction"
)

type BlockChain struct {
	Config    Config
	BestState *BestState

	chainLock sync.RWMutex
}

// Config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	// DataBase defines the database which houses the blocks and will be used to
	// store all metadata created by this package such as the utxo set.
	//
	// This field is required.
	DataBase database.DB

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
	if config.DataBase == nil {
		return errors.New("blockchain.New database is nil")
	}
	if config.ChainParams == nil {
		return errors.New("blockchain.New chain parameters nil")
	}

	self.Config = *config

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := self.initChainState(); err != nil {
		return err
	}

	Logger.log.Infof("BlockChain state (height %d, hash %v, totaltx %d)", self.BestState.Height, self.BestState.BestBlockHash.String(), self.BestState.TotalTxns, )

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
	bestStateBytes, err := self.Config.DataBase.FetchBestBlock()
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

	// store block hash by index and index by block hash
	err = self.StoreBlockIndex(genesisBlock)

	// store best state
	err = self.StoreBestState()
	if err != nil {
		return err
	}

	// Spam random blocks
	for index := 0; index < 0; index++ {
		hashBestBlock := self.BestState.BestBlockHash
		newSpamBlock := Block{
			Header: BlockHeader{
				Version:       1,
				PrevBlockHash: hashBestBlock,
				Timestamp:     time.Now(),
				Difficulty:    0,     //@todo should be create Difficulty logic
				Nonce:         index, //@todo should be create Nonce logic
			},
			Height: int32(index + 1),
		}
		// store block genesis
		err := self.StoreBlock(&newSpamBlock)
		if err != nil {
			return err
		}
		err = self.StoreBlockIndex(genesisBlock)
		if err != nil {
			return err
		}
		self.BestState.Init(&newSpamBlock, 0, 0, numTxns, numTxns, time.Unix(newSpamBlock.Header.Timestamp.Unix(), 0))
		err = self.StoreBestState()
		if err != nil {
			return err
		}
	}
	// Spam random blocks

	return err
}

/**
Get block index(height) of block
 */
func (self *BlockChain) GetBlockHeightByBlockHash(hash *common.Hash) (int32, error) {
	return self.Config.DataBase.GetIndexOfBlock(hash)
}

/**
Get block hash by block index(height)
 */
func (self *BlockChain) GetBlockHashByBlockHeight(height int32) (*common.Hash, error) {
	return self.Config.DataBase.GetBlockByIndex(height)
}

/**
Fetch DB and get block by index(height) of block
 */
func (self *BlockChain) GetBlockByBlockHeight(height int32) (*Block, error) {
	hashBlock, err := self.Config.DataBase.GetBlockByIndex(height)
	if err != nil {
		return nil, err
	}
	blockBytes, err := self.Config.DataBase.FetchBlock(hashBlock)
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
	blockBytes, err := self.Config.DataBase.FetchBlock(hash)
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
	return self.Config.DataBase.StoreBestBlock(self.BestState)
}

/**
Store block into Database
 */
func (self *BlockChain) StoreBlock(block *Block) error {
	return self.Config.DataBase.StoreBlock(block)
}

/**
Save index(height) of block by block hash
and
Save block hash by index(height) of block
 */
func (self *BlockChain) StoreBlockIndex(block *Block) error {
	return self.Config.DataBase.StoreBlockIndex(block.Hash(), block.Height)
}

/**
Get all blocks in chain
Return block array
 */
func (self *BlockChain) GetAllBlocks() ([]*Block, error) {
	result := make([]*Block, 0)
	data, err := self.Config.DataBase.FetchAllBlocks()
	if err != nil {
		return nil, err
	}

	for _, item := range data {
		blockBytes, err := self.Config.DataBase.FetchBlock(item)
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
	data, err := self.Config.DataBase.FetchAllBlocks()
	if err != nil {
		return nil, err
	}
	return data, err
}

// FetchUtxoView loads unspent transaction outputs for the inputs referenced by
// the passed transaction from the point of view of the end of the main chain.
// It also attempts to fetch the utxos for the outputs of the transaction itself
// so the returned view can be examined for duplicate transactions.
//
// This function is safe for concurrent access however the returned view is NOT.
func (b *BlockChain) FetchUtxoView(tx transaction.Tx) (*UtxoViewpoint, error) {
	neededSet := make(map[transaction.OutPoint]struct{})

	// create outpoint map for txout of tx by itself hash
	prevOut := transaction.OutPoint{Hash: *tx.Hash()}
	for txOutIdx, _ := range tx.TxOut {
		prevOut.Vout = uint32(txOutIdx)
		neededSet[prevOut] = struct{}{}
	}

	// create outpoint map for txin of tx
	if !IsCoinBaseTx(tx) {
		for _, txIn := range tx.TxIn {
			neededSet[txIn.PreviousOutPoint] = struct{}{}
		}
	}

	// Request the utxos from the point of view of the end of the main
	// chain.
	view := NewUtxoViewpoint()
	b.chainLock.RLock()
	//@todo will implement late
	err := view.fetchUtxosMain(b.Config.DataBase, neededSet)
	b.chainLock.RUnlock()
	return view, err
}

// CheckTransactionInputs performs a series of checks on the inputs to a
// transaction to ensure they are valid.  An example of some of the checks
// include verifying all inputs exist, ensuring the coinbase seasoning
// requirements are met, detecting double spends, validating all values and fees
// are in the legal range and the total output amount doesn't exceed the input
// amount, and verifying the signatures to prove the spender was the owner of
// the bitcoins and therefore allowed to spend them.  As it checks the inputs,
// it also calculates the total fees for the transaction and returns that value.
//
// NOTE: The transaction MUST have already been sanity checked with the
// CheckTransactionSanity function prior to calling this function.
func (self *BlockChain) CheckTransactionInputs(tx *transaction.Transaction, txHeight int32, utxoView *UtxoViewpoint, chainParams *Params) (int64, error) {
	return 0, nil
}
