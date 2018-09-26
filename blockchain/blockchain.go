package blockchain

import (
	"errors"
	"fmt"

	//"fmt"
	//"time"

	"sync"

	"encoding/json"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

const (
	CHAIN_COUNT = 20
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
	self.BestState = make([]*BestState, CHAIN_COUNT)
	for chainID := byte(0); chainID < CHAIN_COUNT; chainID++ {
		bestStateBytes, err := self.Config.DataBase.FetchBestState(chainID)
		if err == nil {
			err = json.Unmarshal(bestStateBytes, &self.BestState[chainID])
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

	for _, chain := range self.BestState {
		fmt.Println(chain)
	}

	return nil
}

// UpdateMerkleTreeForBlock adds all transaction's commitments in a block to the newest merkle tree
func UpdateMerkleTreeForBlock(tree *client.IncMerkleTree, block *Block) error {
	for _, blockTx := range block.Transactions {
		if blockTx.GetType() == common.TxNormalType {
			tx, ok := blockTx.(*transaction.Tx)
			if ok == false {
				return fmt.Errorf("Transaction in block not valid")
			}

			for _, desc := range tx.Descs {
				for _, cm := range desc.Commitments {
					tree.AddNewNode(cm[:])
				}
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
	//genesisChainID, _ := common.GetTxSenderChain(self.Config.ChainParams.GenesisBlock.Transactions[0].(*transaction.Tx).AddressLastByte)
	if chainID == 0 {
		initBlock = self.Config.ChainParams.GenesisBlock
	} else {
		initBlock = &Block{}
		initBlock.Header.ChainID = chainID
		initBlock.Header.Timestamp = self.Config.ChainParams.GenesisBlock.Header.Timestamp
		initBlock.Header.NextCommittee = self.Config.ChainParams.GenesisBlock.Header.NextCommittee
	}
	initBlock.Height = 1

	// Initialize the state related to the best block.  Since it is the
	// genesis block, use its timestamp for the median time.
	// numTxns := uint64(len(initBlock.Transactions))
	//blockSize := uint64(genesisBlock.SerializeSize())
	//blockWeight := uint64(GetBlockWeight(genesisBlock))

	tree := new(client.IncMerkleTree) // Build genesis block commitment merkle tree
	if err := UpdateMerkleTreeForBlock(tree, initBlock); err != nil {
		return err
	}

	self.BestState[chainID] = &BestState{}
	self.BestState[chainID].Init(initBlock, tree)

	// save nullifiers and commitments from genesisblock
	view := NewTxViewPoint(chainID)
	err := view.fetchTxViewPoint(self.Config.DataBase, initBlock)
	if err != nil {
		return err
	}
	view.SetBestHash(initBlock.Hash())
	// Update the list nullifiers and commitment set using the state of the used tx view point. This
	// entails adding the new
	// ones created by the block.
	err = self.StoreNullifiersFromTxViewPoint(*view)
	if err != nil {
		return err
	}
	err = self.StoreCommitmentsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	// store block genesis
	err = self.StoreBlock(initBlock)
	if err != nil {
		return err
	}

	// store block hash by index and index by block hash
	err = self.StoreBlockIndex(initBlock)
	if err != nil {
		return err
	}
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
	fmt.Println(hashBlock)
	blockBytes, err := self.Config.DataBase.FetchBlock(hashBlock)
	if err != nil {
		return nil, err
	}

	fmt.Println(blockBytes)
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
/*func (self *BlockChain) StoreUtxoView(view *UtxoViewpoint) error {
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
}*/

/**
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromTxViewPoint(view TxViewPoint) error {
	for typeJoinSplitDesc, item := range view.listNullifiers {
		for _, item1 := range item {
			err := self.Config.DataBase.StoreNullifiers(item1, typeJoinSplitDesc, view.chainId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/**
Uses an existing database to update the set of not used tx by saving list commitments of privacy,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromTxViewPoint(view TxViewPoint) error {
	for typeJoinSplitDesc, item := range view.listCommitments {
		for _, item1 := range item {
			err := self.Config.DataBase.StoreCommitments(item1, typeJoinSplitDesc, view.chainId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/**
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromListNullifier(nullifiers [][]byte, typeJoinSplitDesc string, chainId byte) error {
	for _, nullifier := range nullifiers {
		err := self.Config.DataBase.StoreNullifiers(nullifier, typeJoinSplitDesc, chainId)
		if err != nil {
			return err
		}
	}
	return nil
}

/**
Uses an existing database to update the set of not used tx by saving list commitments of privacy,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromListCommitment(commitments [][]byte, typeJoinSplitDesc string, chainId byte) error {
	for _, item := range commitments {
		err := self.Config.DataBase.StoreCommitments(item, typeJoinSplitDesc, chainId)
		if err != nil {
			return err
		}
	}
	return nil
}

/**
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromTx(tx *transaction.Tx, typeJoinSplitDesc string) error {
	for _, desc := range tx.Descs {
		for _, nullifier := range desc.Nullifiers {
			chainId, err := common.GetTxSenderChain(tx.AddressLastByte)
			if err != nil {
				return err
			}
			err = self.Config.DataBase.StoreNullifiers(nullifier, typeJoinSplitDesc, chainId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/**
Uses an existing database to update the set of not used tx by saving list commitments of privacy,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromTx(tx *transaction.Tx, typeJoinSplitDesc string) error {
	for _, desc := range tx.Descs {
		for _, item := range desc.Commitments {
			chainId, err := common.GetTxSenderChain(tx.AddressLastByte)
			if err != nil {
				return err
			}
			err = self.Config.DataBase.StoreCommitments(item, typeJoinSplitDesc, chainId)
			if err != nil {
				return err
			}
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

func (self *BlockChain) GetChainBlocks(chainID byte) ([]*Block, error) {
	result := make([]*Block, 0)
	data, err := self.Config.DataBase.FetchChainBlocks(chainID)
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
/*func (b *BlockChain) FetchUtxoView(tx transaction.Tx) (*UtxoViewpoint, error) {
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
}*/

/**
FetchTxViewPoint -  return a tx view point, which contain list commitments and nullifiers
Param typeJoinSplitDesc - COIN or BOND
*/
func (self *BlockChain) FetchTxViewPoint(typeJoinSplitDesc string, chainId byte) (*TxViewPoint, error) {
	view := NewTxViewPoint(chainId)
	commitments, err := self.Config.DataBase.FetchCommitments(typeJoinSplitDesc, chainId)
	if err != nil {
		return nil, err
	}
	view.listCommitments[typeJoinSplitDesc] = commitments
	nullifiers, err := self.Config.DataBase.FetchNullifiers(typeJoinSplitDesc, chainId)
	if err != nil {
		return nil, err
	}
	view.listNullifiers[typeJoinSplitDesc] = nullifiers
	// view.SetBestHash(self.BestState.BestBlockHash)
	return view, nil
}

// connectBestChain handles connecting the passed block to the chain while
// respecting proper chain selection according to the chain with the most
// proof of work.  In the typical case, the new block simply extends the main
// chain.  However, it may also be extending (or creating) a side chain (fork)
// which may or may not end up becoming the main chain depending on which fork
// cumulatively has the most proof of work.  It returns whether or not the block
// ended up on the main chain (either due to extending the main chain or causing
// a reorganization to become the main chain).
/*func (b *BlockChain) connectBestChain(block *Block) (bool, error) {
	// We are extending the main (best) chain with a new block.  This is the
	// most common case.
	parentHash := &block.Header.PrevBlockHash
	if parentHash.IsEqual(&b.BestState[block.Header.ChainID].BestBlockHash) {
		view := NewUtxoViewpoint()
		view.SetBestHash(parentHash)

		err := view.fetchInputUtxos(b.Config.DataBase, block)
		stxos := make([]SpentTxOut, 0, countSpentOutputs(block))
		if err != nil {
			return false, err
		}

		//
		err = view.connectTransactions(block, &stxos)
		if err != nil {
			return false, err
		}

		// TODO with stxos
		_ = stxos

		// Update the utxo set using the state of the utxo view.  This
		// entails removing all of the utxos spent and adding the new
		// ones created by the block.
		err = b.StoreUtxoView(view)
		if err != nil {
			return false, err
		}

		return true, nil
	} else {
		// we in sub chain
		return false, nil
	}
}*/
func (b *BlockChain) connectBestChain(block *Block) (bool, error) {
	// We are extending the main (best) chain with a new block.  This is the
	// most common case.
	parentHash := &block.Header.PrevBlockHash
	if parentHash.IsEqual(b.BestState[block.Header.ChainID].BestBlockHash) {
		view := NewTxViewPoint(block.Header.ChainID)

		err := view.fetchTxViewPoint(b.Config.DataBase, block)
		if err != nil {
			return false, err
		}

		view.SetBestHash(block.Hash())
		// Update the list nullifiers and commitment set using the state of the used tx view point. This
		// entails adding the new
		// ones created by the block.
		err = b.StoreNullifiersFromTxViewPoint(*view)
		if err != nil {
			return false, err
		}

		err = b.StoreCommitmentsFromTxViewPoint(*view)
		if err != nil {
			return false, err
		}

		return true, nil
	} else {
		// we in sub chain
		return false, nil
	}
}

// countSpentOutputs returns the number of utxos the passed block spends.
/*func countSpentOutputs(block *Block) int {
	// Exclude the coinbase transaction since it can't spend anything.
	var numSpent int
	for _, tx := range block.Transactions[1:] {
		if tx.GetType() == common.TxNormalType {
			numSpent += len(tx.(*transaction.Tx).TxIn)
		}
	}
	return numSpent
}*/

/**
GetListTxByReadonlyKey - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key
- Param #1: key - key set which contain readonly-key and pub-key
- Param #2: typeJoinSplitDesc - which type of joinsplitdesc(COIN or BOND)
*/
func (self *BlockChain) GetListTxByReadonlyKey(keySet *cashec.KeySet, typeJoinSplitDesc string) (map[byte][]transaction.Tx, error) {
	results := make(map[byte][]transaction.Tx, 0)

	// set default for params
	if typeJoinSplitDesc == "" {
		typeJoinSplitDesc = common.TxOutCoinType
	}

	// lock chain
	self.chainLock.Lock()

	for _, bestState := range self.BestState {
		// get best block
		bestBlock := bestState.BestBlock
		chainId := bestState.BestBlock.Header.ChainID
		blockHeight := bestBlock.Height

		for blockHeight > 0 {
			txsInBlock := bestBlock.Transactions
			txsInBlockAccepted := make([]transaction.Tx, 0)
			for _, txInBlock := range txsInBlock {
				if txInBlock.GetType() == common.TxNormalType {
					tx := txInBlock.(*transaction.Tx)
					copyTx := transaction.Tx{
						Version:  tx.Version,
						JSSig:    tx.JSSig,
						JSPubKey: tx.JSPubKey,
						Fee:      tx.Fee,
						Type:     tx.Type,
						LockTime: tx.LockTime,
						Descs:    make([]*transaction.JoinSplitDesc, 0),
					}
					// try to decrypt each of desc in tx with readonly Key and add to txsInBlockAccepted
					listDesc := make([]*transaction.JoinSplitDesc, 0)
					for _, desc := range tx.Descs {
						copyDesc := &transaction.JoinSplitDesc{
							Anchor:        desc.Anchor,
							Commitments:   make([][]byte, 0),
							EncryptedData: make([][]byte, 0),
						}
						for i, encData := range desc.EncryptedData {
							var epk client.EphemeralPubKey
							copy(epk[:], desc.EphemeralPubKey)
							// var hSig []byte
							// copy(hSig, desc.HSigSeed)
							hSig := client.HSigCRH(desc.HSigSeed, desc.Nullifiers[0], desc.Nullifiers[1], copyTx.JSPubKey)
							note := new(client.Note)
							note, err := client.DecryptNote(encData, keySet.ReadonlyKey.Skenc, keySet.PublicKey.Pkenc, epk, hSig)
							spew.Dump(note)
							if err == nil && note != nil {
								copyDesc.EncryptedData = append(copyDesc.EncryptedData, encData)
								copyDesc.AppendNote(note)
								copyDesc.Commitments = append(copyDesc.Commitments, desc.Commitments[i])
							} else {
								continue
							}
						}
						if len(copyDesc.EncryptedData) > 0 {
							listDesc = append(listDesc, copyDesc)
						}
					}
					if len(listDesc) > 0 {
						copyTx.Descs = listDesc
					}
					txsInBlockAccepted = append(txsInBlockAccepted, copyTx)
				}
			}
			// detected some tx can be accepted
			if len(txsInBlockAccepted) > 0 {
				// add to result
				results[chainId] = append(results[chainId], txsInBlockAccepted...)
			}

			// continue with previous block
			blockHeight--
			if blockHeight > 0 {
				// not is genesis block
				preBlockHash := bestBlock.Header.PrevBlockHash
				bestBlock, err := self.GetBlockByBlockHash(&preBlockHash)
				if blockHeight != bestBlock.Height || err != nil {
					// pre-block is not the same block-height with calculation -> invalid blockchain
					return nil, errors.New("Invalid blockchain")
				}
			}
		}
	}

	// unlock chain
	self.chainLock.Unlock()
	return results, nil
}

/**
GetListTxByPrivateKey - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key.
With private-key, we can check unspent tx by check nullifiers from database
- Param #1: privateKey - byte[] of privatekey
- Param #2: typeJoinSplitDesc - which type of joinsplitdesc(COIN or BOND)
*/
func (self *BlockChain) GetListTxByPrivateKey(privateKey *client.SpendingKey, typeJoinSplitDesc string, sortType int, sortAsc bool) (map[byte][]transaction.Tx, error) {
	results := make(map[byte][]transaction.Tx)

	// Get set of keys from private keybyte
	keys := cashec.KeySet{}
	keys.ImportFromPrivateKey(privateKey)

	// set default for params
	if typeJoinSplitDesc == "" {
		typeJoinSplitDesc = common.TxOutCoinType
	}

	// lock chain
	self.chainLock.Lock()

	for _, bestState := range self.BestState {
		// get best blockFs
		bestBlock := bestState.BestBlock
		chainId := bestBlock.Header.ChainID
		results[chainId] = make([]transaction.Tx, 0)
		blockHeight := bestBlock.Height

		// get list nullifiers from db to check spending
		txViewPoint, err := self.FetchTxViewPoint(typeJoinSplitDesc, chainId)
		if err != nil {
			return nil, err
		}
		nullifiersInDb := txViewPoint.listNullifiers[typeJoinSplitDesc]

		for blockHeight > 0 {
			txsInBlock := bestBlock.Transactions
			txsInBlockAccepted := make([]transaction.Tx, 0)
			for _, txInBlock := range txsInBlock {
				if txInBlock.GetType() == common.TxNormalType {
					tx := txInBlock.(*transaction.Tx)
					copyTx := transaction.Tx{
						Version:         tx.Version,
						JSSig:           tx.JSSig,
						JSPubKey:        tx.JSPubKey,
						Fee:             tx.Fee,
						Type:            tx.Type,
						LockTime:        tx.LockTime,
						Descs:           make([]*transaction.JoinSplitDesc, 0),
						AddressLastByte: tx.AddressLastByte,
					}
					// try to decrypt each of desc in tx with readonly Key and add to txsInBlockAccepted
					listDesc := make([]*transaction.JoinSplitDesc, 0)
					for _, desc := range tx.Descs {
						copyDesc := &transaction.JoinSplitDesc{
							Anchor:        desc.Anchor,
							Reward:        desc.Reward,
							Commitments:   make([][]byte, 0),
							EncryptedData: make([][]byte, 0),
						}
						for i, encData := range desc.EncryptedData {
							var epk client.EphemeralPubKey
							copy(epk[:], desc.EphemeralPubKey)
							hSig := client.HSigCRH(desc.HSigSeed, desc.Nullifiers[0], desc.Nullifiers[1], copyTx.JSPubKey)
							note := new(client.Note)
							note, err := client.DecryptNote(encData, keys.ReadonlyKey.Skenc, keys.PublicKey.Pkenc, epk, hSig)
							if err == nil && note != nil && note.Value > 0 {
								// can decrypt data -> got candidate commitment
								candidateCommitment := desc.Commitments[i]
								if len(nullifiersInDb) > 0 {
									// -> check commitment with db nullifiers
									var rho [32]byte
									copy(rho[:], note.Rho)
									candidateNullifier := client.GetNullifier(keys.PrivateKey, rho)
									if len(candidateNullifier) == 0 {
										continue
									}
									checkCandiateNullifier, err := common.SliceBytesExists(nullifiersInDb, candidateNullifier)
									if err != nil || checkCandiateNullifier == true {
										// candidate nullifier is not existed in db
										continue
									}
								}
								copyDesc.EncryptedData = append(copyDesc.EncryptedData, encData)
								copyDesc.AppendNote(note)
								note.Cm = candidateCommitment
								note.Apk = client.GenPaymentAddress(keys.PrivateKey).Apk
								copyDesc.Commitments = append(copyDesc.Commitments, candidateCommitment)
							} else {
								continue
							}
						}
						if len(copyDesc.EncryptedData) > 0 {
							listDesc = append(listDesc, copyDesc)
						}
					}
					if len(listDesc) > 0 {
						copyTx.Descs = listDesc
					}
					if len(copyTx.Descs) > 0 {
						txsInBlockAccepted = append(txsInBlockAccepted, copyTx)
					}
				}
			}
			// detected some tx can be accepted
			if len(txsInBlockAccepted) > 0 {
				// add to result
				results[chainId] = append(results[chainId], txsInBlockAccepted...)
			}

			// continue with previous block
			blockHeight--
			if blockHeight > 0 {
				// not is genesis block
				preBlockHash := bestBlock.Header.PrevBlockHash
				preBlock, err := self.GetBlockByBlockHash(&preBlockHash)
				if err != nil || blockHeight != preBlock.Height {
					// pre-block is not the same block-height with calculation -> invalid blockchain
					self.chainLock.Unlock()
					return nil, errors.New("Invalid blockchain")
				}
				bestBlock = preBlock
			}
		}
		// sort txs
		transaction.SortArrayTxs(results[chainId], sortType, sortAsc)
	}

	// unlock chain
	self.chainLock.Unlock()

	return results, nil
}

func (self *BlockChain) GetAllUnitCoinSupplier() (map[string]uint64, error) {
	result := make(map[string]uint64)
	result[common.TxOutCoinType] = uint64(0)
	result[common.TxOutBondType] = uint64(0)

	// lock chain
	self.chainLock.Lock()
	for _, bestState := range self.BestState {
		// get best block of each chain
		bestBlock := bestState.BestBlock
		blockHeight := bestBlock.Height

		for blockHeight > 0 {

			txsInBlock := bestBlock.Transactions
			totalFeeInBlock := uint64(0)
			for _, txInBlock := range txsInBlock {
				tx := txInBlock.(*transaction.Tx)
				fee := tx.Fee
				totalFeeInBlock += fee
			}

			coinbaseTx := txsInBlock[0].(*transaction.Tx)
			rewardBond := uint64(0)
			rewardCoin := uint64(0)
			for _, desc := range coinbaseTx.Descs {
				unitType := desc.Type
				switch unitType {
				case common.TxOutCoinType:
					rewardCoin += desc.Reward
				case common.TxOutBondType:
					rewardBond += desc.Reward
				}
			}
			rewardCoin -= totalFeeInBlock
			result[common.TxOutCoinType] += rewardCoin
			result[common.TxOutBondType] += rewardBond

			// continue with previous block
			blockHeight--
			if blockHeight > 0 {
				// not is genesis block
				preBlockHash := bestBlock.Header.PrevBlockHash
				bestBlock, err := self.GetBlockByBlockHash(&preBlockHash)
				if blockHeight != bestBlock.Height || err != nil {
					// pre-block is not the same block-height with calculation -> invalid blockchain
					return nil, errors.New("Invalid blockchain")
				}
			}
		}
	}
	// unlock chain
	self.chainLock.Unlock()
	return result, nil
}
