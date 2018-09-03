package blockchain

import (
	"errors"

	"sync"

	"encoding/json"
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

type BlockChain struct {
	Config    Config
	BestState []*BestState //BestState of 20 chain.

	chainLock sync.RWMutex
}

type blockIdx struct {
	Idx     int
	ChainID byte
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

func (self *BlockChain) Init(config *Config) error {
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

	// Logger.log.Infof("BlockChain state (height %d, hash %v, totaltx %d)", self.BestState.Height, self.BestState.BestBlockHash.String(), self.BestState.TotalTxns)

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
	for chainID := byte(0); chainID <= 19; chainID++ {
		bestStateBytes, err := self.Config.DataBase.FetchBestState(chainID)
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
			err := self.createChainState(chainID)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
func (self *BlockChain) createChainState(chainID byte) error {
	// Create a new block from genesis block and set it as best block of chain
	var initBlock *Block
	if chainID == 0 {
		initBlock = self.Config.ChainParams.GenesisBlock
	} else {
		initBlock = &Block{}
	}
	initBlock.Height = 0

	// Initialize the state related to the best block.  Since it is the
	// genesis block, use its timestamp for the median time.
	numTxns := uint64(len(initBlock.Transactions))
	//blockSize := uint64(initBlock.SerializeSize())
	//blockWeight := uint64(GetBlockWeight(initBlock))
	self.BestState[chainID] = &BestState{}
	self.BestState[chainID].Init(initBlock, 0, 0, numTxns, numTxns, time.Unix(initBlock.Header.Timestamp.Unix(), 0))

	// store block genesis
	err := self.StoreBlock(initBlock)
	if err != nil {
		return err
	}

	// store block hash by index and index by block hash
	err = self.StoreBlockIndex(initBlock)

	// store best state
	err = self.StoreBestState(chainID)
	if err != nil {
		return err
	}

	// Spam random blocks
	// for index := 0; index < 0; index++ {
	// 	hashBestBlock := self.BestState[0].BestBlockHash
	// 	newSpamBlock := Block{
	// 		Header: BlockHeader{
	// 			Version:       1,
	// 			PrevBlockHash: hashBestBlock,
	// 			Timestamp:     time.Now(),
	// 			Difficulty:    0,     //@todo should be create Difficulty logic
	// 			Nonce:         index, //@todo should be create Nonce logic
	// 		},
	// 		Height: int32(index + 1),
	// 	}
	// 	// store block genesis
	// 	err := self.StoreBlock(&newSpamBlock)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = self.StoreBlockIndex(initBlock)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	self.BestState[chainID].Init(&newSpamBlock, 0, 0, numTxns, numTxns, time.Unix(newSpamBlock.Header.Timestamp.Unix(), 0))
	// 	err = self.StoreBestState(chainID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

/**
Get block index(height) of block
*/
func (self *BlockChain) GetBlockHeightByBlockHash(hash *common.Hash) (int32, byte, error) {
	return self.Config.DataBase.GetIndexOfBlock(hash)
}

/**
Get block hash by block index(height)
*/
func (self *BlockChain) GetBlockHashByBlockHeight(height int32, chainID byte) (*common.Hash, error) {
	return self.Config.DataBase.GetBlockByIndex(height, chainID)
}

/**
Fetch DB and get block by index(height) of block
*/
func (self *BlockChain) GetBlockByBlockHeight(height int32, chainID byte) (*Block, error) {
	hashBlock, err := self.Config.DataBase.GetBlockByIndex(height, chainID)
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
func (self *BlockChain) StoreBestState(chainID byte) error {
	return self.Config.DataBase.StoreBestBlock(self.BestState[chainID], chainID)
}

func (self *BlockChain) GetBestState(chainID byte) (*BestState, error) {
	bestState := BestState{}
	bestStateBytes, err := self.Config.DataBase.FetchBestState(chainID)
	if err == nil {
		err = json.Unmarshal(bestStateBytes, &bestState)
	}
	return &bestState, err
}

/**
Store block into Database
*/
func (self *BlockChain) StoreBlock(block *Block) error {
	return self.Config.DataBase.StoreBlock(block, block.Header.ChainID)
}

/**
Save index(height) of block by block hash
and
Save block hash by index(height) of block
*/
func (self *BlockChain) StoreBlockIndex(block *Block) error {
	return self.Config.DataBase.StoreBlockIndex(block.Hash(), block.Height, block.Header.ChainID)
}

// Uses an existing database to update the utxo set
// in the database based on the provided utxo view contents and state.  In
// particular, only the entries that have been marked as modified are written
// to the database.
func (self *BlockChain) StoreUtxoView(view *UtxoViewpoint) error {
	for outpoint, entry := range view.entries {
		// No need to update the database if the entry was not modified.
		if entry == nil || !entry.isModified() {
			continue
		}

		// Remove the utxo entry if it is spent.
		if entry.IsSpent() {
			err := self.Config.DataBase.DeleteUtxoEntry(&outpoint)
			//recycleOutpointKey(key)
			if err != nil {
				return err
			}
			continue
		}

		err := self.Config.DataBase.StoreUtxoEntry(&outpoint, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

/**
Get all blocks in chain
Return block array
*/
func (self *BlockChain) GetAllBlocks() ([][]*Block, error) {
	result := make([][]*Block, 0)
	data, err := self.Config.DataBase.FetchAllBlocks()
	if err != nil {
		return nil, err
	}

	for chainID, chain := range data {
		for _, item := range chain {
			blockBytes, err := self.Config.DataBase.FetchBlock(item)
			if err != nil {
				return nil, err
			}
			block := Block{}
			err = json.Unmarshal(blockBytes, &block)
			if err != nil {
				return nil, err
			}
			result[chainID] = append(result[chainID], &block)
		}
	}

	return result, nil
}

/**
Get all hash of blocks in chain
Return hashes array
*/
func (self *BlockChain) GetAllHashBlocks() ([][]*common.Hash, error) {
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
func (self *BlockChain) CheckTransactionInputs(tx *transaction.Transaction, txHeight int32, utxoView *UtxoViewpoint, chainParams *Params) (float64, error) {
	return 0, nil
}
