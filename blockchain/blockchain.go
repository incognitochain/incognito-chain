package blockchain

import (
	"errors"

	//"fmt"
	//"time"

	"sync"

	"encoding/json"

	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/transaction"
)

const (
	ChainCount = 20
)

/*
blockChain is a view presents for data in blockchain network
because we use 20 chain data to contain all block in system, so
this struct has a array best state with len = 20,
every beststate present for a best block in every chain
*/
type BlockChain struct {
	BestState []*BestState //BestState of 20 chain.

	config    Config
	chainLock sync.RWMutex
}

// config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	// dataBase defines the database which houses the blocks and will be used to
	// store all metadata created by this package.
	//
	// This field is required.
	DataBase database.DatabaseInterface

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

/*
Init - init a blockchain view from config
*/
func (self *BlockChain) Init(config *Config) error {
	// Enforce required config fields.
	if config.DataBase == nil {
		return NewBlockChainError(UnExpectedError, errors.New("Database is not config"))
	}
	if config.ChainParams == nil {
		return NewBlockChainError(UnExpectedError, errors.New("Chain parameters is not config"))
	}

	self.config = *config

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := self.initChainState(); err != nil {
		return err
	}

	for chainIndex, bestState := range self.BestState {
		Logger.log.Infof("BlockChain state for chain #%d (Height %d, Best block hash %+v, Total tx %d, Salary fund %d, Gov Param %+v)",
			chainIndex, bestState.Height, bestState.BestBlockHash.String(), bestState.TotalTxns, bestState.BestBlock.Header.SalaryFund, bestState.BestBlock.Header.GovernanceParams)
	}

	return nil
}

/*
// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
*/
func (self *BlockChain) initChainState() error {
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool
	self.BestState = make([]*BestState, ChainCount)
	for chainId := byte(0); chainId < ChainCount; chainId++ {
		bestStateBytes, err := self.config.DataBase.FetchBestState(chainId)
		if err == nil {
			err = json.Unmarshal(bestStateBytes, &self.BestState[chainId])
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
			err := self.createChainState(chainId)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

/*
// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
*/
func (self *BlockChain) createChainState(chainId byte) error {
	// Create a new block from genesis block and set it as best block of chain
	var initBlock *Block
	if chainId == 0 {
		initBlock = self.config.ChainParams.GenesisBlock
	} else {
		initBlock = &Block{}
		initBlock.Header = self.config.ChainParams.GenesisBlock.Header
		initBlock.Header.ChainID = chainId
		initBlock.Header.PrevBlockHash = common.Hash{}
	}
	initBlock.Header.Height = 1

	tree := new(client.IncMerkleTree) // Build genesis block commitment merkle tree
	if err := UpdateMerkleTreeForBlock(tree, initBlock); err != nil {
		return NewBlockChainError(UpdateMerkleTreeForBlockError, err)
	}

	self.BestState[chainId] = &BestState{}
	self.BestState[chainId].Init(initBlock, tree)

	// store block genesis
	err := self.StoreBlock(initBlock)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	// store block hash by index and index by block hash
	err = self.StoreBlockIndex(initBlock)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	// store best state
	err = self.StoreBestState(chainId)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	return nil
}

/*
Get block index(height) of block
*/
func (self *BlockChain) GetBlockHeightByBlockHash(hash *common.Hash) (int32, byte, error) {
	return self.config.DataBase.GetIndexOfBlock(hash)
}

/*
Get block hash by block index(height)
*/
func (self *BlockChain) GetBlockHashByBlockHeight(height int32, chainId byte) (*common.Hash, error) {
	return self.config.DataBase.GetBlockByIndex(height, chainId)
}

/*
Fetch DatabaseInterface and get block by index(height) of block
*/
func (self *BlockChain) GetBlockByBlockHeight(height int32, chainId byte) (*Block, error) {
	hashBlock, err := self.config.DataBase.GetBlockByIndex(height, chainId)
	if err != nil {
		return nil, err
	}
	blockBytes, err := self.config.DataBase.FetchBlock(hashBlock)
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

/*
Fetch DatabaseInterface and get block data by block hash
*/
func (self *BlockChain) GetBlockByBlockHash(hash *common.Hash) (*Block, error) {
	blockBytes, err := self.config.DataBase.FetchBlock(hash)
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

/*
Store best state of block(best block, num of tx, ...) into Database
*/
func (self *BlockChain) StoreBestState(chainId byte) error {
	return self.config.DataBase.StoreBestState(self.BestState[chainId], chainId)
}

/*
GetBestState - return a best state from a chain
*/
// #1 - chainID - index of chain
func (self *BlockChain) GetBestState(chainId byte) (*BestState, error) {
	bestState := BestState{}
	bestStateBytes, err := self.config.DataBase.FetchBestState(chainId)
	if err == nil {
		err = json.Unmarshal(bestStateBytes, &bestState)
	}
	return &bestState, err
}

/*
Store block into Database
*/
func (self *BlockChain) StoreBlock(block *Block) error {
	return self.config.DataBase.StoreBlock(block, block.Header.ChainID)
}

/*
Save index(height) of block by block hash
and
Save block hash by index(height) of block
*/
func (self *BlockChain) StoreBlockIndex(block *Block) error {
	return self.config.DataBase.StoreBlockIndex(block.Hash(), block.Header.Height, block.Header.ChainID)
}

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromTxViewPoint(view TxViewPoint) error {
	for coinType, item := range view.listNullifiers {
		for _, item1 := range item {
			err := self.config.DataBase.StoreNullifiers(item1, coinType, view.chainID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Uses an existing database to update the set of not used tx by saving list commitments of privacy,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromTxViewPoint(view TxViewPoint) error {
	for coinType, item := range view.listCommitments {
		for _, item1 := range item {
			err := self.config.DataBase.StoreCommitments(item1, coinType, view.chainID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromListNullifier(nullifiers [][]byte, coinType string, chainId byte) error {
	for _, nullifier := range nullifiers {
		err := self.config.DataBase.StoreNullifiers(nullifier, coinType, chainId)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of not used tx by saving list commitments of privacy,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromListCommitment(commitments [][]byte, coinType string, chainId byte) error {
	for _, item := range commitments {
		err := self.config.DataBase.StoreCommitments(item, coinType, chainId)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
func (self *BlockChain) StoreNullifiersFromTx(tx *transaction.Tx, coinType string) error {
	for _, desc := range tx.Descs {
		for _, nullifier := range desc.Nullifiers {
			chainId, err := common.GetTxSenderChain(tx.AddressLastByte)
			if err != nil {
				return err
			}
			err = self.config.DataBase.StoreNullifiers(nullifier, coinType, chainId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Uses an existing database to update the set of not used tx by saving list commitments of privacy,
this is a list tx-in which are used by a new tx
*/
func (self *BlockChain) StoreCommitmentsFromTx(tx *transaction.Tx, coinType string) error {
	for _, desc := range tx.Descs {
		for _, item := range desc.Commitments {
			chainId, err := common.GetTxSenderChain(tx.AddressLastByte)
			if err != nil {
				return err
			}
			err = self.config.DataBase.StoreCommitments(item, coinType, chainId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
Get all blocks in chain
Return block array
*/
func (self *BlockChain) GetAllBlocks() ([][]*Block, error) {
	result := make([][]*Block, 0)
	data, err := self.config.DataBase.FetchAllBlocks()
	if err != nil {
		return nil, err
	}

	for chainID, chain := range data {
		for _, item := range chain {
			blockBytes, err := self.config.DataBase.FetchBlock(item)
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
	data, err := self.config.DataBase.FetchChainBlocks(chainID)
	if err != nil {
		return nil, err
	}

	for _, item := range data {
		blockBytes, err := self.config.DataBase.FetchBlock(item)
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

/*
Get all hash of blocks in chain
Return hashes array
*/
func (self *BlockChain) GetAllHashBlocks() (map[byte][]*common.Hash, error) {
	data, err := self.config.DataBase.FetchAllBlocks()
	if err != nil {
		return nil, err
	}
	return data, err
}

/*
FetchTxViewPoint -  return a tx view point, which contain list commitments and nullifiers
Param coinType - COIN or BOND
*/
func (self *BlockChain) FetchTxViewPoint(coinType string, chainId byte) (*TxViewPoint, error) {
	view := NewTxViewPoint(chainId)
	commitments, err := self.config.DataBase.FetchCommitments(coinType, chainId)
	if err != nil {
		return nil, err
	}
	view.listCommitments[coinType] = commitments
	nullifiers, err := self.config.DataBase.FetchNullifiers(coinType, chainId)
	if err != nil {
		return nil, err
	}
	view.listNullifiers[coinType] = nullifiers
	// view.SetBestHash(self.BestState.BestBlockHash)
	return view, nil
}

func (self *BlockChain) CreateTxViewPoint(block *Block) error {
	view := NewTxViewPoint(block.Header.ChainID)

	err := view.fetchTxViewPoint(self.config.DataBase, block)
	if err != nil {
		return err
	}

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

	return nil
}

/*
GetListTxByReadonlyKey - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key
- Param #1: key - key set which contain readonly-key and pub-key
- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
*/
func (self *BlockChain) GetListTxByReadonlyKey(keySet *cashec.KeySet, coinType string) (map[byte][]transaction.Tx, error) {
	results := make(map[byte][]transaction.Tx, 0)

	// set default for params
	if coinType == "" {
		coinType = common.AssetTypeCoin
	}

	// lock chain
	self.chainLock.Lock()

	for _, bestState := range self.BestState {
		// get best block
		bestBlock := bestState.BestBlock
		chainId := bestState.BestBlock.Header.ChainID
		blockHeight := bestBlock.Header.Height

		for blockHeight > 0 {
			txsInBlock := bestBlock.Transactions
			txsInBlockAccepted := make([]transaction.Tx, 0)
			for _, txInBlock := range txsInBlock {
				if txInBlock.GetType() == common.TxNormalType || txInBlock.GetType() == common.TxSalaryType {
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
						if desc.Proof != nil && len(desc.EncryptedData) > 0 {
							// privacy
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
						} else {
							// no privacy
							for i, note := range desc.Note {
								copyDesc.AppendNote(note)
								copyDesc.Commitments = append(copyDesc.Commitments, desc.Commitments[i])
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
				// TODO Voting
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
				if blockHeight != bestBlock.Header.Height || err != nil {
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

/*
GetListTxByPrivateKey - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key.
With private-key, we can check unspent tx by check nullifiers from database
- Param #1: privateKey - byte[] of privatekey
- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
*/
func (self *BlockChain) GetListTxByPrivateKey(privateKey *client.SpendingKey, coinType string, sortType int, sortAsc bool) (map[byte][]transaction.Tx, error) {
	results := make(map[byte][]transaction.Tx)

	// Get set of keys from private keybyte
	keys := cashec.KeySet{}
	keys.ImportFromPrivateKey(privateKey)

	// set default for params
	if coinType == "" {
		coinType = common.AssetTypeCoin
	}

	// lock chain
	self.chainLock.Lock()

	// get list nullifiers from db to check spending
	nullifiersInDb := make([][]byte, 0)
	for _, bestState := range self.BestState {
		bestBlock := bestState.BestBlock
		chainId := bestBlock.Header.ChainID
		txViewPoint, err := self.FetchTxViewPoint(coinType, chainId)
		if err != nil {
			return nil, err
		}
		nullifiersInDb = append(nullifiersInDb, txViewPoint.listNullifiers[coinType]...)
	}

	for _, bestState := range self.BestState {
		// get best blockFs
		bestBlock := bestState.BestBlock
		chainId := bestBlock.Header.ChainID
		results[chainId] = make([]transaction.Tx, 0)
		blockHeight := bestBlock.Header.Height

		for blockHeight > 0 {
			txsInBlock := bestBlock.Transactions
			txsInBlockAccepted := make([]transaction.Tx, 0)
			for _, txInBlock := range txsInBlock {
				if txInBlock.GetType() == common.TxNormalType || txInBlock.GetType() == common.TxSalaryType {
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
						if desc.Proof != nil && len(desc.EncryptedData) > 0 {
							// have privacy
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
						} else {
							for i, note := range desc.Note {
								// no privacy
								candidateCommitment := desc.Commitments[i]
								if len(nullifiersInDb) > 0 {
									// -> check commitment with db nullifiers
									var rho [32]byte
									copy(rho[:], note.Rho)
									candidateNullifier := desc.Nullifiers
									if len(candidateNullifier) == 0 {
										continue
									}
									checkCandiateNullifier, err := common.SliceBytesExists(nullifiersInDb, candidateNullifier)
									if err != nil || checkCandiateNullifier == true {
										// candidate nullifier is not existed in db
										continue
									}
								}
								copyDesc.AppendNote(note)
								note.Cm = candidateCommitment
								note.Apk = client.GenPaymentAddress(keys.PrivateKey).Apk
								copyDesc.Commitments = append(copyDesc.Commitments, candidateCommitment)
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
				// TODO Voting
			}
			// detected some tx can be accepted
			if len(txsInBlockAccepted) > 0 {
				// add to result
				results[chainId] = append(results[chainId], txsInBlockAccepted...)
			}

			// continue with previous block
			blockHeight--
			if chainId != 0 && blockHeight == 1 {
				break
			}
			if blockHeight > 0 {
				// not is genesis block
				preBlockHash := bestBlock.Header.PrevBlockHash
				preBlock, err := self.GetBlockByBlockHash(&preBlockHash)
				if err != nil || blockHeight != preBlock.Header.Height {
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

/*
GetAllUnitCoinSupplier - return all list unit currency(bond, coin, ...) with amount of every of them
*/
func (self *BlockChain) GetAllUnitCoinSupplier() (map[string]uint64, error) {
	result := make(map[string]uint64)
	result[common.AssetTypeCoin] = uint64(0)
	result[common.AssetTypeBond] = uint64(0)

	// lock chain
	self.chainLock.Lock()
	for _, bestState := range self.BestState {
		// get best block of each chain
		bestBlock := bestState.BestBlock
		blockHeight := bestBlock.Header.Height

		for blockHeight > 0 {

			txsInBlock := bestBlock.Transactions
			totalFeeInBlock := uint64(0)
			for _, txInBlock := range txsInBlock {
				tx := txInBlock.(*transaction.Tx)
				fee := tx.Fee
				totalFeeInBlock += fee
			}

			salaryTx := txsInBlock[0].(*transaction.Tx)
			rewardBond := uint64(0)
			rewardCoin := uint64(0)
			for _, desc := range salaryTx.Descs {
				unitType := desc.Type
				switch unitType {
				case common.AssetTypeCoin:
					rewardCoin += desc.Reward
				case common.AssetTypeBond:
					rewardBond += desc.Reward
				}
			}
			rewardCoin -= totalFeeInBlock
			result[common.AssetTypeCoin] += rewardCoin
			result[common.AssetTypeBond] += rewardBond

			// continue with previous block
			blockHeight--
			if blockHeight > 0 {
				// not is genesis block
				preBlockHash := bestBlock.Header.PrevBlockHash
				bestBlock, err := self.GetBlockByBlockHash(&preBlockHash)
				if blockHeight != bestBlock.Header.Height || err != nil {
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

func (self *BlockChain) GetCommitteCandidate(pubkeyParam string) *CommitteeCandidateInfo {
	for _, bestState := range self.BestState {
		for pubkey, candidateInfo := range bestState.Candidates {
			if pubkey == pubkeyParam {
				return &candidateInfo
			}
		}
	}
	return nil
}

/*
Get Candidate List from all chain and merge all to one - return pubkey of them
*/
func (self *BlockChain) GetCommitteeCandidateList() []string {
	candidatePubkeyList := []string{}
	for _, bestState := range self.BestState {
		for pubkey, _ := range bestState.Candidates {
			if common.IndexOfStr(pubkey, candidatePubkeyList) < 0 {
				candidatePubkeyList = append(candidatePubkeyList, pubkey)
			}
		}
	}
	sort.Slice(candidatePubkeyList, func(i, j int) bool {
		cndInfoi := self.GetCommitteeCandidateInfo(candidatePubkeyList[i])
		cndInfoj := self.GetCommitteeCandidateInfo(candidatePubkeyList[j])
		if cndInfoi.Value == cndInfoj.Value {
			if cndInfoi.Timestamp < cndInfoj.Timestamp {
				return true
			} else if cndInfoi.Timestamp > cndInfoj.Timestamp {
				return false
			} else {
				if cndInfoi.ChainID <= cndInfoj.ChainID {
					return true
				} else if cndInfoi.ChainID < cndInfoj.ChainID {
					return false
				}
			}
		} else if cndInfoi.Value > cndInfoj.Value {
			return true
		} else {
			return false
		}
		return false
	})
	return candidatePubkeyList
}

func (self *BlockChain) GetCommitteeCandidateInfo(nodeAddr string) CommitteeCandidateInfo {
	var cndVal CommitteeCandidateInfo
	for _, bestState := range self.BestState {
		cndValTmp, ok := bestState.Candidates[nodeAddr]
		if ok {
			cndVal.Value += cndValTmp.Value
			if cndValTmp.Timestamp > cndVal.Timestamp {
				cndVal.Timestamp = cndValTmp.Timestamp
				cndVal.ChainID = cndValTmp.ChainID
			}
		}
	}
	return cndVal
}
