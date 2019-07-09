package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/btc"
	"github.com/incognitochain/incognito-chain/pubsub"
	"io"
	"log"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"
	
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

/*
blockChain is a view presents for data in blockchain network
because we use 20 chain data to contain all block in system, so
this struct has a array best state with len = 20,
every beststate present for a best block in every chain
*/
type BlockChain struct {
	BestState *BestState
	config    Config
	chainLock sync.Mutex
	//channel
	cQuitSync        chan struct{}
	Synker           synker
	ConsensusOngoing bool
	IsTest           bool
}
type BestState struct {
	Beacon *BestStateBeacon
	Shard  map[byte]*BestStateShard
}

// config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	DataBase          database.DatabaseInterface
	Interrupt         <-chan struct{}
	ChainParams       *Params
	RelayShards       []byte
	NodeMode          string
	ShardToBeaconPool ShardToBeaconPool
	CrossShardPool    map[byte]CrossShardPool
	BeaconPool        BeaconPool
	ShardPool         map[byte]ShardPool
	TxPool            TxPool
	TempTxPool        TxPool
	CRemovedTxs       chan metadata.Transaction
	FeeEstimator      map[byte]FeeEstimator
	IsBlockGenStarted bool
	PubSubManager     *pubsub.PubSubManager
	RandomClient      btc.RandomClient
	Server            interface {
		BoardcastNodeState() error

		PushMessageGetBlockBeaconByHeight(from uint64, to uint64, peerID libp2p.ID) error
		PushMessageGetBlockBeaconByHash(blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
		PushMessageGetBlockBeaconBySpecificHeight(heights []uint64, getFromPool bool, peerID libp2p.ID) error

		PushMessageGetBlockShardByHeight(shardID byte, from uint64, to uint64, peerID libp2p.ID) error
		PushMessageGetBlockShardByHash(shardID byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
		PushMessageGetBlockShardBySpecificHeight(shardID byte, heights []uint64, getFromPool bool, peerID libp2p.ID) error

		PushMessageGetBlockShardToBeaconByHeight(shardID byte, from uint64, to uint64, peerID libp2p.ID) error
		PushMessageGetBlockShardToBeaconByHash(shardID byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
		PushMessageGetBlockShardToBeaconBySpecificHeight(shardID byte, blksHeight []uint64, getFromPool bool, peerID libp2p.ID) error

		PushMessageGetBlockCrossShardByHash(fromShard byte, toShard byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
		PushMessageGetBlockCrossShardBySpecificHeight(fromShard byte, toShard byte, blksHeight []uint64, getFromPool bool, peerID libp2p.ID) error
		UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string)
	}
	UserKeySet *incognitokey.KeySet
}

func NewBlockChain(config *Config, isTest bool) *BlockChain {
	bc := &BlockChain{}
	bc.config = *config
	bc.config.IsBlockGenStarted = false
	bc.IsTest = isTest
	bc.cQuitSync = make(chan struct{})
	bc.BestState = &BestState{
		Beacon: &BestStateBeacon{},
		Shard:  make(map[byte]*BestStateShard),
	}
	for i := 0; i < 255; i++ {
		shardID := byte(i)
		bc.BestState.Shard[shardID] = &BestStateShard{}
	}
	bc.Synker = synker{
		blockchain: bc,
		cQuit:      bc.cQuitSync,
	}
	return bc
}

/*
Init - init a blockchain view from config
*/
func (blockchain *BlockChain) Init(config *Config) error {
	// Enforce required config fields.
	if config.DataBase == nil {
		return NewBlockChainError(UnExpectedError, errors.New("Database is not config"))
	}
	if config.ChainParams == nil {
		return NewBlockChainError(UnExpectedError, errors.New("Chain parameters is not config"))
	}

	blockchain.config = *config
	blockchain.config.IsBlockGenStarted = false
	blockchain.IsTest = false
	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := blockchain.initChainState(); err != nil {
		return err
	}

	blockchain.cQuitSync = make(chan struct{})
	blockchain.Synker = synker{
		blockchain: blockchain,
		cQuit:      blockchain.cQuitSync,
	}
	return nil
}

func (blockchain *BlockChain) SetIsBlockGenStarted(value bool) {
	blockchain.config.IsBlockGenStarted = value
}
func (blockchain *BlockChain) AddTxPool(txpool TxPool) {
	blockchain.config.TxPool = txpool
}

func (blockchain *BlockChain) AddTempTxPool(temptxpool TxPool) {
	blockchain.config.TempTxPool = temptxpool
}

func (blockchain *BlockChain) SetFeeEstimator(feeEstimator FeeEstimator, shardID byte) {
	if len(blockchain.config.FeeEstimator) == 0 {
		blockchain.config.FeeEstimator = make(map[byte]FeeEstimator)
	}
	blockchain.config.FeeEstimator[shardID] = feeEstimator
}

func (blockchain *BlockChain) InitChannelBlockchain(cRemovedTxs chan metadata.Transaction) {
	blockchain.config.CRemovedTxs = cRemovedTxs
}

// -------------- Blockchain retriever's implementation --------------
// GetCustomTokenTxsHash - return list of tx which relate to custom token
func (blockchain *BlockChain) GetCustomTokenTxs(tokenID *common.Hash) (map[common.Hash]metadata.Transaction, error) {
	txHashesInByte, err := blockchain.config.DataBase.CustomTokenTxs(*tokenID)
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash]metadata.Transaction)
	for _, temp := range txHashesInByte {
		_, _, _, tx, err := blockchain.GetTransactionByHash(temp)
		if err != nil {
			return nil, err
		}
		result[*tx.Hash()] = tx
	}
	return result, nil
}

// -------------- End of Blockchain retriever's implementation --------------

/*
// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
*/
func (blockchain *BlockChain) initChainState() error {
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool

	blockchain.BestState = &BestState{
		Beacon: nil,
		Shard:  make(map[byte]*BestStateShard),
	}

	bestStateBeaconBytes, err := blockchain.config.DataBase.FetchBeaconBestState()
	if err == nil {
		beacon := &BestStateBeacon{}
		err = json.Unmarshal(bestStateBeaconBytes, beacon)
		//update singleton object
		SetBestStateBeacon(beacon)
		//update beacon field in blockchain Beststate
		blockchain.BestState.Beacon = GetBestStateBeacon()

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
		err := blockchain.initBeaconState()
		if err != nil {
			return err
		}

	}

	for shard := 1; shard <= blockchain.BestState.Beacon.ActiveShards; shard++ {
		shardID := byte(shard - 1)
		bestStateBytes, err := blockchain.config.DataBase.FetchShardBestState(shardID)
		if err == nil {
			shardBestState := &BestStateShard{}
			err = json.Unmarshal(bestStateBytes, shardBestState)
			//update singleton object
			SetBestStateShard(shardID, shardBestState)
			//update Shard field in blockchain Beststate
			blockchain.BestState.Shard[shardID] = GetBestStateShard(shardID)
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
			err := blockchain.initShardState(shardID)
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
func (blockchain *BlockChain) initShardState(shardID byte) error {
	blockchain.BestState.Shard[shardID] = InitBestStateShard(shardID, blockchain.config.ChainParams)
	// Create a new block from genesis block and set it as best block of chain
	initBlock := ShardBlock{}
	initBlock = *blockchain.config.ChainParams.GenesisShardBlock
	initBlock.Header.ShardID = shardID

	_, newShardCandidate := GetStakingCandidate(*blockchain.config.ChainParams.GenesisBeaconBlock)

	blockchain.BestState.Shard[shardID].ShardCommittee = append(blockchain.BestState.Shard[shardID].ShardCommittee, newShardCandidate[int(shardID)*blockchain.config.ChainParams.ShardCommitteeSize:(int(shardID)*blockchain.config.ChainParams.ShardCommitteeSize)+blockchain.config.ChainParams.ShardCommitteeSize]...)

	genesisBeaconBlk, err := blockchain.GetBeaconBlockByHeight(1)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	err = blockchain.BestState.Shard[shardID].Update(&initBlock, []*BeaconBlock{genesisBeaconBlk})
	if err != nil {
		return err
	}
	blockchain.ProcessStoreShardBlock(&initBlock)
	return nil
}

func (blockchain *BlockChain) initBeaconState() error {
	blockchain.BestState.Beacon = InitBestStateBeacon(blockchain.config.ChainParams)
	initBlock := blockchain.config.ChainParams.GenesisBeaconBlock
	blockchain.BestState.Beacon.Update(initBlock, blockchain)

	// Insert new block into beacon chain
	if err := blockchain.StoreBeaconBestState(); err != nil {
		Logger.log.Error("Error Store best state for block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}
	if err := blockchain.config.DataBase.StoreBeaconBlock(&blockchain.BestState.Beacon.BestBlock, blockchain.BestState.Beacon.BestBlock.Header.Hash()); err != nil {
		Logger.log.Error("Error store beacon block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return err
	}
	if err := blockchain.config.DataBase.StoreCommitteeByEpoch(initBlock.Header.Epoch, blockchain.BestState.Beacon.GetShardCommittee()); err != nil {
		return err
	}
	blockHash := initBlock.Hash()
	if err := blockchain.config.DataBase.StoreBeaconBlockIndex(*blockHash, initBlock.Header.Height); err != nil {
		return err
	}
	return nil
}

/*
Get block index(height) of block
*/
func (blockchain *BlockChain) GetBlockHeightByBlockHash(hash common.Hash) (uint64, byte, error) {
	return blockchain.config.DataBase.GetIndexOfBlock(hash)
}

/*
Get block hash by block index(height)
*/
func (blockchain *BlockChain) GetBeaconBlockHashByHeight(height uint64) (common.Hash, error) {
	return blockchain.config.DataBase.GetBeaconBlockHashByIndex(height)
}

/*
Fetch DatabaseInterface and get block by index(height) of block
*/
func (blockchain *BlockChain) GetBeaconBlockByHeight(height uint64) (*BeaconBlock, error) {
	if blockchain.IsTest {
		return &BeaconBlock{}, nil
	}
	hashBlock, err := blockchain.config.DataBase.GetBeaconBlockHashByIndex(height)
	if err != nil {
		return nil, err
	}
	block, _, err := blockchain.GetBeaconBlockByHash(hashBlock)
	if err != nil {
		return nil, err
	}
	return block, nil
}

/*
Fetch DatabaseInterface and get block data by block hash
*/
func (blockchain *BlockChain) GetBeaconBlockByHash(hash common.Hash) (*BeaconBlock, uint64, error) {
	if blockchain.IsTest {
		return &BeaconBlock{}, 2, nil
	}
	blockBytes, err := blockchain.config.DataBase.FetchBeaconBlock(hash)
	if err != nil {
		return nil, 0, err
	}
	block := BeaconBlock{}
	err = json.Unmarshal(blockBytes, &block)
	if err != nil {
		return nil, 0, err
	}
	return &block, uint64(len(blockBytes)), nil
}

/*
Get block index(height) of block
*/
func (blockchain *BlockChain) GetShardBlockHeightByHash(hash common.Hash) (uint64, byte, error) {
	return blockchain.config.DataBase.GetIndexOfBlock(hash)
}

/*
Get block hash by block index(height)
*/
func (blockchain *BlockChain) GetShardBlockHashByHeight(height uint64, shardID byte) (common.Hash, error) {
	return blockchain.config.DataBase.GetBlockByIndex(height, shardID)
}

/*
Fetch DatabaseInterface and get block by index(height) of block
*/
func (blockchain *BlockChain) GetShardBlockByHeight(height uint64, shardID byte) (*ShardBlock, error) {
	hashBlock, err := blockchain.config.DataBase.GetBlockByIndex(height, shardID)
	if err != nil {
		return nil, err
	}
	block, _, err := blockchain.GetShardBlockByHash(hashBlock)

	return block, err
}

/*
Fetch DatabaseInterface and get block data by block hash
*/
func (blockchain *BlockChain) GetShardBlockByHash(hash common.Hash) (*ShardBlock, uint64, error) {
	if blockchain.IsTest {
		return &ShardBlock{}, 2, nil
	}
	blockBytes, err := blockchain.config.DataBase.FetchBlock(hash)
	if err != nil {
		return nil, 0, err
	}

	block := ShardBlock{}
	err = json.Unmarshal(blockBytes, &block)
	if err != nil {
		return nil, 0, err
	}
	return &block, uint64(len(blockBytes)), nil
}

/*
Store best state of block(best block, num of tx, ...) into Database
*/
func (blockchain *BlockChain) StoreBeaconBestState() error {
	return blockchain.config.DataBase.StoreBeaconBestState(blockchain.BestState.Beacon)
}

/*
Store best state of block(best block, num of tx, ...) into Database
*/
func (blockchain *BlockChain) StoreShardBestState(shardID byte) error {
	return blockchain.config.DataBase.StoreShardBestState(blockchain.BestState.Shard[shardID], shardID)
}

/*
GetBestState - return a best state from a chain
*/
// #1 - shardID - index of chain
func (blockchain *BlockChain) GetShardBestState(shardID byte) (*BestStateShard, error) {
	bestState := BestStateShard{}
	bestStateBytes, err := blockchain.config.DataBase.FetchShardBestState(shardID)
	if err == nil {
		err = json.Unmarshal(bestStateBytes, &bestState)
	}
	return &bestState, err
}

/*
Store block into Database
*/
func (blockchain *BlockChain) StoreShardBlock(block *ShardBlock) error {
	return blockchain.config.DataBase.StoreShardBlock(block, block.Header.Hash(), block.Header.ShardID)
}

/*
Save index(height) of block by block hash
and
Save block hash by index(height) of block
*/
func (blockchain *BlockChain) StoreShardBlockIndex(block *ShardBlock) error {
	return blockchain.config.DataBase.StoreShardBlockIndex(block.Header.Hash(), block.Header.Height, block.Header.ShardID)
}

func (blockchain *BlockChain) StoreTransactionIndex(txHash *common.Hash, blockHash common.Hash, index int) error {
	return blockchain.config.DataBase.StoreTransactionIndex(*txHash, blockHash, index)
}

/*
Uses an existing database to update the set of used tx by saving list serialNumber of privacy,
this is a list tx-out which are used by a new tx
*/
func (blockchain *BlockChain) StoreSerialNumbersFromTxViewPoint(view TxViewPoint) error {
	//for _, item1 := range view.listSerialNumbers {
	err := blockchain.config.DataBase.StoreSerialNumbers(*view.tokenID, view.listSerialNumbers, view.shardID)
	if err != nil {
		return err
	}
	//}
	return nil
}

/*
Uses an existing database to update the set of used tx by saving list SNDerivator of privacy,
this is a list tx-out which are used by a new tx
*/
func (blockchain *BlockChain) StoreSNDerivatorsFromTxViewPoint(view TxViewPoint, shardID byte) error {
	// commitment
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// Store SND of every transaction in this block
		// UNCOMMENT: TO STORE SND WITH NON-CROSS SHARD TRANSACTION ONLY
		// pubkey := k
		// pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		// if err != nil {
		// 	return err
		// }
		// lastByte := pubkeyBytes[len(pubkeyBytes)-1]
		// pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
		// if pubkeyShardID == shardID {
		snDsArray := view.mapSnD[k]
		//for _, snd := range snDsArray {
		err := blockchain.config.DataBase.StoreSNDerivators(*view.tokenID, snDsArray, view.shardID)
		if err != nil {
			return err
		}
		// }
		//}
	}

	// for pubkey, items := range view.mapSnD {
	// 	pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	lastByte := pubkeyBytes[len(pubkeyBytes)-1]
	// 	pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
	// 	if pubkeyShardID == shardID {
	// 		for _, item1 := range items {
	// 			err := blockchain.config.DataBase.StoreSNDerivators(view.tokenID, item1, view.shardID)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	// 	}
	// }
	return nil
}

/*
Uses an existing database to update the set of not used tx by saving list commitments of privacy,
this is a list tx-in which are used by a new tx
*/
func (blockchain *BlockChain) StoreCommitmentsFromTxViewPoint(view TxViewPoint, shardID byte) error {

	// commitment and output are the same key in map
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		publicKey := k
		publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
		if err != nil {
			return err
		}
		lastByte := publicKeyBytes[len(publicKeyBytes)-1]
		publicKeyShardID := common.GetShardIDFromLastByte(lastByte)
		if publicKeyShardID == shardID {
			// commitment
			commitmentsArray := view.mapCommitments[k]
			err = blockchain.config.DataBase.StoreCommitments(*view.tokenID, publicKeyBytes, commitmentsArray, view.shardID)
			if err != nil {
				return err
			}
			// outputs
			outputCoinArray := view.mapOutputCoins[k]
			outputCoinBytesArray := make([][]byte, 0)
			for _, outputCoin := range outputCoinArray {
				outputCoinBytesArray = append(outputCoinBytesArray, outputCoin.Bytes())
			}
			err = blockchain.config.DataBase.StoreOutputCoins(*view.tokenID, publicKeyBytes, outputCoinBytesArray, publicKeyShardID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateAndSaveTxViewPointFromBlock - fetch data from block, put into txviewpoint variable and save into db
// @note: still storage full data of commitments, serialnumbersm snderivator to check double spend
// @note: this function only work for transaction transfer token/prv within shard
func (blockchain *BlockChain) CreateAndSaveTxViewPointFromBlock(block *ShardBlock) error {
	//startTime := time.Now()
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchTxViewPointFromBlock(blockchain.config.DataBase, block)
	if err != nil {
		return err
	}

	// check normal custom token
	for indexTx, customTokenTx := range view.customTokenTxs {
		switch customTokenTx.TxTokenData.Type {
		case transaction.CustomTokenInit:
			{
				Logger.log.Info("Store custom token when it is issued", customTokenTx.TxTokenData.PropertyID, customTokenTx.TxTokenData.PropertySymbol, customTokenTx.TxTokenData.PropertyName)
				err = blockchain.config.DataBase.StoreCustomToken(customTokenTx.TxTokenData.PropertyID, customTokenTx.Hash()[:])
				if err != nil {
					return err
				}
			}
		case transaction.CustomTokenCrossShard:
			{
				// 0xsirrush updated: check existed token ID
				existedToken := blockchain.CustomTokenIDExisted(&customTokenTx.TxTokenData.PropertyID)
				//If don't exist then create
				if !existedToken {
					Logger.log.Info("Store Cross Shard Custom if It's not existed in DB", customTokenTx.TxTokenData.PropertyID, customTokenTx.TxTokenData.PropertySymbol, customTokenTx.TxTokenData.PropertyName)
					err = blockchain.config.DataBase.StoreCustomToken(customTokenTx.TxTokenData.PropertyID, customTokenTx.Hash()[:])
					if err != nil {
						Logger.log.Error("CreateAndSaveTxViewPointFromBlock", err)
					}
				}
				/*listCustomToken, err := blockchain.ListCustomToken()
				if err != nil {
					panic(err)
				}
				//If don't exist then create
				if _, ok := listCustomToken[customTokenTx.TxTokenData.PropertyID]; !ok {
					Logger.log.Info("Store Cross Shard Custom if It's not existed in DB", customTokenTx.TxTokenData.PropertyID, customTokenTx.TxTokenData.PropertySymbol, customTokenTx.TxTokenData.PropertyName)
					err = blockchain.config.DataBase.StoreCustomToken(&customTokenTx.TxTokenData.PropertyID, customTokenTx.Hash()[:])
				}*/
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Info("Transfer custom token %+v", customTokenTx)
			}
		}
		// save tx which relate to custom token
		// Reject Double spend UTXO before enter this state
		//fmt.Printf("StoreCustomTokenPaymentAddresstHistory/CustomTokenTx: \n VIN %+v VOUT %+v \n", customTokenTx.TxTokenData.Vins, customTokenTx.TxTokenData.Vouts)
		Logger.log.Info("Store Custom Token History")
		err = blockchain.StoreCustomTokenPaymentAddresstHistory(customTokenTx, block.Header.ShardID)
		if err != nil {
			// Skip double spend
			return err
		}
		err = blockchain.config.DataBase.StoreCustomTokenTx(customTokenTx.TxTokenData.PropertyID, block.Header.ShardID, block.Header.Height, indexTx, customTokenTx.Hash()[:])
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}
	}

	// check privacy custom token
	for indexTx, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		privacyCustomTokenTx := view.privacyCustomTokenTxs[indexTx]
		switch privacyCustomTokenTx.TxTokenPrivacyData.Type {
		case transaction.CustomTokenInit:
			{
				Logger.log.Info("Store custom token when it is issued", privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, privacyCustomTokenTx.TxTokenPrivacyData.PropertySymbol, privacyCustomTokenTx.TxTokenPrivacyData.PropertyName)
				err = blockchain.config.DataBase.StorePrivacyCustomToken(privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, privacyCustomTokenTx.Hash()[:])
				if err != nil {
					return err
				}
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Info("Transfer custom token %+v", privacyCustomTokenTx)
			}
		}
		err = blockchain.config.DataBase.StorePrivacyCustomTokenTx(privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, block.Header.ShardID, block.Header.Height, indexTx, privacyCustomTokenTx.Hash()[:])
		if err != nil {
			return err
		}

		err = blockchain.StoreSerialNumbersFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}

		err = blockchain.StoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}

		err = blockchain.StoreSNDerivatorsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}
	}

	// Update the list serialNumber and commitment, snd set using the state of the used tx view point. This
	// entails adding the new
	// ones created by the block.
	err = blockchain.StoreSerialNumbersFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = blockchain.StoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}

	err = blockchain.StoreSNDerivatorsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}
	//endtime := time.Now()
	//runTime := endtime.Sub(startTime)
	//go common.AnalyzeFuncCreateAndSaveTxViewPointFromBlock(runTime.Seconds())
	//Logger.log.Critical("*** CreateAndSaveTxViewPointFromBlock  ***", block.Header.Height, runTime)
	return nil
}

func (blockchain *BlockChain) CreateAndSaveCrossTransactionCoinViewPointFromBlock(block *ShardBlock) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlock(blockchain.config.DataBase, block)
	if err != nil {
		Logger.log.Error("CreateAndSaveCrossTransactionCoinViewPointFromBlock", err)
	}
	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
		// 0xsirrush updated: check existed tokenID
		tokenID := privacyCustomTokenSubView.tokenID
		existed := blockchain.PrivacyCustomTokenIDExisted(tokenID)
		if !existed {
			existedCrossShard := blockchain.PrivacyCustomTokenIDCrossShardExisted(tokenID)
			if !existedCrossShard {
				Logger.log.Info("Store custom token when it is issued ", tokenID, privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertyName, privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertySymbol, privacyCustomTokenSubView.privacyCustomTokenMetadata.Amount, privacyCustomTokenSubView.privacyCustomTokenMetadata.Mintable)
				tokenDataBytes, _ := json.Marshal(privacyCustomTokenSubView.privacyCustomTokenMetadata)

				// crossShardTokenPrivacyMetaData := CrossShardTokenPrivacyMetaData{}
				// json.Unmarshal(tokenDataBytes, &crossShardTokenPrivacyMetaData)
				// fmt.Println("New Token CrossShardTokenPrivacyMetaData", crossShardTokenPrivacyMetaDatla)

				if err := blockchain.config.DataBase.StorePrivacyCustomTokenCrossShard(*tokenID, tokenDataBytes); err != nil {
					return err
				}
			}
		}
		/*listCustomTokens, listCustomTokenCrossShard, err := blockchain.ListPrivacyCustomToken()
		if err != nil {
			return nil
		}
		tokenID := privacyCustomTokenSubView.tokenID
		if _, ok := listCustomTokens[*tokenID]; !ok {
			if _, ok := listCustomTokenCrossShard[*tokenID]; !ok {
				Logger.log.Info("Store custom token when it is issued ", tokenID, privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertyName, privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertySymbol, privacyCustomTokenSubView.privacyCustomTokenMetadata.Amount, privacyCustomTokenSubView.privacyCustomTokenMetadata.Mintable)
				tokenDataBytes, _ := json.Marshal(privacyCustomTokenSubView.privacyCustomTokenMetadata)

				// crossShardTokenPrivacyMetaData := CrossShardTokenPrivacyMetaData{}
				// json.Unmarshal(tokenDataBytes, &crossShardTokenPrivacyMetaData)
				// fmt.Println("New Token CrossShardTokenPrivacyMetaData", crossShardTokenPrivacyMetaData)

				if err := blockchain.config.DataBase.StorePrivacyCustomTokenCrossShard(tokenID, tokenDataBytes); err != nil {
					return err
				}
			}
		}*/
		// Store both commitment and outcoin
		err = blockchain.StoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}
		// store snd
		err = blockchain.StoreSNDerivatorsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}
	}

	// Update the list serialNumber and commitment, snd set using the state of the used tx view point. This
	// entails adding the new
	// ones created by the block.
	err = blockchain.StoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}

	err = blockchain.StoreSNDerivatorsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}

	return nil
}

/*
// 	KeyWallet: token-paymentAddress  -[-]-  {tokenId}  -[-]-  {paymentAddress}  -[-]-  {txHash}  -[-]-  {voutIndex}
//   H: value-spent/unspent
*/
func (blockchain *BlockChain) StoreCustomTokenPaymentAddresstHistory(customTokenTx *transaction.TxCustomToken, shardID byte) error {
	Splitter := lvdb.Splitter
	TokenPaymentAddressPrefix := lvdb.TokenPaymentAddressPrefix
	unspent := lvdb.Unspent
	spent := lvdb.Spent

	tokenKey := TokenPaymentAddressPrefix
	tokenKey = append(tokenKey, Splitter...)
	tokenKey = append(tokenKey, []byte((customTokenTx.TxTokenData.PropertyID).String())...)
	for _, vin := range customTokenTx.TxTokenData.Vins {
		paymentAddressBytes := base58.Base58Check{}.Encode(vin.PaymentAddress.Bytes(), 0x00)
		utxoHash := []byte(vin.TxCustomTokenID.String())
		voutIndex := vin.VoutIndex
		paymentAddressKey := tokenKey
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, paymentAddressBytes...)
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, utxoHash[:]...)
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, common.Int32ToBytes(int32(voutIndex))...)
		_, err := blockchain.config.DataBase.HasValue(paymentAddressKey)
		if err != nil {
			return err
		}
		value, err := blockchain.config.DataBase.Get(paymentAddressKey)
		if err != nil {
			return err
		}
		// old value: {value}-unspent
		values := strings.Split(string(value), string(Splitter))
		if strings.Compare(values[1], string(unspent)) != 0 {
			return errors.New("Double Spend Detected")
		}
		// new value: {value}-spent
		newValues := values[0] + string(Splitter) + string(spent)
		if err := blockchain.config.DataBase.Put(paymentAddressKey, []byte(newValues)); err != nil {
			return err
		}
	}
	for index, vout := range customTokenTx.TxTokenData.Vouts {
		// check vout by type and receiver
		txCustomTokenType := customTokenTx.TxTokenData.Type
		if txCustomTokenType == transaction.CustomTokenInit || txCustomTokenType == transaction.CustomTokenTransfer {
			// check receiver's shard and current shard ID
			shardIDOfReceiver := common.GetShardIDFromLastByte(vout.PaymentAddress.Pk[len(vout.PaymentAddress.Pk)-1])
			if shardIDOfReceiver != shardID {
				continue
			}
		} else if txCustomTokenType == transaction.CustomTokenCrossShard {
			shardIDOfReceiver := common.GetShardIDFromLastByte(vout.PaymentAddress.Pk[len(vout.PaymentAddress.Pk)-1])
			if shardIDOfReceiver != shardID {
				continue
			}
		}
		paymentAddressBytes := base58.Base58Check{}.Encode(vout.PaymentAddress.Bytes(), 0x00)
		utxoHash := []byte(customTokenTx.Hash().String())
		voutIndex := index
		value := vout.Value
		paymentAddressKey := tokenKey
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, paymentAddressBytes...)
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, utxoHash[:]...)
		paymentAddressKey = append(paymentAddressKey, Splitter...)
		paymentAddressKey = append(paymentAddressKey, common.Int32ToBytes(int32(voutIndex))...)
		ok, err := blockchain.config.DataBase.HasValue(paymentAddressKey)
		// Vout already exist
		if ok {
			return errors.New("UTXO already exist")
		}
		if err != nil {
			return err
		}
		// init value: {value}-unspent
		paymentAddressValue := strconv.Itoa(int(value)) + string(Splitter) + string(unspent) + string(Splitter)
		if err := blockchain.config.DataBase.Put(paymentAddressKey, []byte(paymentAddressValue)); err != nil {
			return err
		}
		fmt.Printf("STORE UTXO FOR CUSTOM TOKEN: tokenID %+v \n paymentAddress %+v \n txHash %+v, voutIndex %+v, value %+v \n", (customTokenTx.TxTokenData.PropertyID).String(), vout.PaymentAddress, customTokenTx.Hash(), voutIndex, value)
	}
	return nil
}

// DecryptTxByKey - process outputcoin to get outputcoin data which relate to keyset
func (blockchain *BlockChain) DecryptOutputCoinByKey(outCoinTemp *privacy.OutputCoin, keySet *incognitokey.KeySet, shardID byte, tokenID *common.Hash) *privacy.OutputCoin {
	/*
		- Param keyset - (priv-key, payment-address, readonlykey)
		in case priv-key: return unspent outputcoin tx
		in case readonly-key: return all outputcoin tx with amount value
		in case payment-address: return all outputcoin tx with no amount value
	*/
	pubkeyCompress := outCoinTemp.CoinDetails.PublicKey.Compress()
	if bytes.Equal(pubkeyCompress, keySet.PaymentAddress.Pk[:]) {
		result := &privacy.OutputCoin{
			CoinDetails:          outCoinTemp.CoinDetails,
			CoinDetailsEncrypted: outCoinTemp.CoinDetailsEncrypted,
		}
		if result.CoinDetailsEncrypted != nil {
			if len(keySet.PrivateKey) > 0 || len(keySet.ReadonlyKey.Rk) > 0 {
				// try to decrypt to get more data
				err := result.Decrypt(keySet.ReadonlyKey)
				if err == nil {
					result.CoinDetails = outCoinTemp.CoinDetails
				}
			}
		}
		if len(keySet.PrivateKey) > 0 {
			// check spent with private-key
			result.CoinDetails.SerialNumber = privacy.PedCom.G[privacy.SK].Derive(new(big.Int).SetBytes(keySet.PrivateKey),
				result.CoinDetails.SNDerivator)
			ok, err := blockchain.config.DataBase.HasSerialNumber(*tokenID, result.CoinDetails.SerialNumber.Compress(), shardID)
			if ok || err != nil {
				return nil
			}
		}
		return result
	}
	return nil
}

/*
GetListOutputCoinsByKeyset - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key.
With private-key, we can check unspent tx by check serialNumber from database
- Param #1: keyset - (priv-key, payment-address, readonlykey)
in case priv-key: return unspent outputcoin tx
in case readonly-key: return all outputcoin tx with amount value
in case payment-address: return all outputcoin tx with no amount value
- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
*/
func (blockchain *BlockChain) GetListOutputCoinsByKeyset(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash) ([]*privacy.OutputCoin, error) {
	// lock chain
	blockchain.BestState.Shard[shardID].lock.Lock()
	defer blockchain.BestState.Shard[shardID].lock.Unlock()

	outCointsInBytes, err := blockchain.config.DataBase.GetOutcoinsByPubkey(*tokenID, keyset.PaymentAddress.Pk[:], shardID)
	if err != nil {
		return nil, err
	}
	// convert from []byte to object
	outCoints := make([]*privacy.OutputCoin, 0)
	for _, item := range outCointsInBytes {
		outcoin := &privacy.OutputCoin{}
		outcoin.Init()
		outcoin.SetBytes(item)
		outCoints = append(outCoints, outcoin)
	}

	// loop on all outputcoin to decrypt data
	results := make([]*privacy.OutputCoin, 0)
	for _, out := range outCoints {
		out = blockchain.DecryptOutputCoinByKey(out, keyset, shardID, tokenID)
		if out == nil {
			continue
		} else {
			results = append(results, out)
		}
	}
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetUnspentTxCustomTokenVout - return all unspent tx custom token out of sender
func (blockchain *BlockChain) GetUnspentTxCustomTokenVout(receiverKeyset incognitokey.KeySet, tokenID *common.Hash) ([]transaction.TxTokenVout, error) {
	data, err := blockchain.config.DataBase.GetCustomTokenPaymentAddressUTXO(*tokenID, receiverKeyset.PaymentAddress.Bytes())
	fmt.Println(data)
	if err != nil {
		return nil, err
	}
	splitter := []byte("-[-]-")
	unspent := []byte("unspent")
	voutList := []transaction.TxTokenVout{}
	for key, value := range data {
		keys := strings.Split(key, string(splitter))
		values := strings.Split(value, string(splitter))
		// values: [amount-value, spent/unspent]
		// get unspent transaction output
		if strings.Compare(values[1], string(unspent)) == 0 {
			vout := transaction.TxTokenVout{}
			vout.PaymentAddress = receiverKeyset.PaymentAddress
			txHash, err := common.Hash{}.NewHashFromStr(string(keys[3]))
			if err != nil {
				return nil, err
			}
			vout.SetTxCustomTokenID(*txHash)
			voutIndexByte := []byte(keys[4])
			voutIndex, err := common.BytesToInt32(voutIndexByte)
			if err != nil {
				return nil, err
			}
			vout.SetIndex(int(voutIndex))
			value, err := strconv.Atoi(values[0])
			if err != nil {
				return nil, err
			}
			vout.Value = uint64(value)
			Logger.log.Info("GetCustomTokenPaymentAddressUTXO VOUT", vout)
			voutList = append(voutList, vout)
		}
	}
	return voutList, nil
}

// GetTransactionByHash - retrieve tx from txId(txHash)
func (blockchain *BlockChain) GetTransactionByHash(txHash common.Hash) (byte, common.Hash, int, metadata.Transaction, error) {
	blockHash, index, err := blockchain.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		abc := NewBlockChainError(UnExpectedError, err)
		Logger.log.Error(abc)
		return byte(255), common.Hash{}, -1, nil, abc
	}
	block, _, err1 := blockchain.GetShardBlockByHash(blockHash)
	if err1 != nil {
		Logger.log.Errorf("ERROR", err1, "NO Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Body.Transactions[index])
		return byte(255), common.Hash{}, -1, nil, NewBlockChainError(UnExpectedError, err1)
	}
	//Logger.log.Infof("Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Transactions[index])
	return block.Header.ShardID, blockHash, index, block.Body.Transactions[index], nil
}

// Check Custom token ID is existed
func (blockchain *BlockChain) CustomTokenIDExisted(tokenID *common.Hash) bool {
	return blockchain.config.DataBase.CustomTokenIDExisted(*tokenID)
}

// Check Privacy Custom token ID is existed
func (blockchain *BlockChain) PrivacyCustomTokenIDExisted(tokenID *common.Hash) bool {
	return blockchain.config.DataBase.PrivacyCustomTokenIDExisted(*tokenID)
}

func (blockchain *BlockChain) PrivacyCustomTokenIDCrossShardExisted(tokenID *common.Hash) bool {
	return blockchain.config.DataBase.PrivacyCustomTokenIDCrossShardExisted(*tokenID)
}

// ListCustomToken - return all custom token which existed in network
func (blockchain *BlockChain) ListCustomToken() (map[common.Hash]transaction.TxCustomToken, error) {
	data, err := blockchain.config.DataBase.ListCustomToken()
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash]transaction.TxCustomToken)
	for _, txData := range data {
		hash := common.Hash{}
		hash.SetBytes(txData)
		_, blockHash, index, tx, err := blockchain.GetTransactionByHash(hash)
		_ = blockHash
		_ = index
		if err != nil {
			return nil, NewBlockChainError(UnExpectedError, err)
		}
		txCustomToken := tx.(*transaction.TxCustomToken)
		result[txCustomToken.TxTokenData.PropertyID] = *txCustomToken
	}
	return result, nil
}

// ListCustomToken - return all custom token which existed in network
func (blockchain *BlockChain) ListPrivacyCustomToken() (map[common.Hash]transaction.TxCustomTokenPrivacy, map[common.Hash]CrossShardTokenPrivacyMetaData, error) {
	data, err := blockchain.config.DataBase.ListPrivacyCustomToken()
	if err != nil {
		return nil, nil, err
	}
	crossShardData, err := blockchain.config.DataBase.ListPrivacyCustomTokenCrossShard()
	if err != nil {
		return nil, nil, err
	}
	result := make(map[common.Hash]transaction.TxCustomTokenPrivacy)
	for _, txData := range data {
		hash := common.Hash{}
		hash.SetBytes(txData)
		_, blockHash, index, tx, err := blockchain.GetTransactionByHash(hash)
		_ = blockHash
		_ = index
		if err != nil {
			return nil, nil, err
		}
		txPrivacyCustomToken := tx.(*transaction.TxCustomTokenPrivacy)
		result[txPrivacyCustomToken.TxTokenPrivacyData.PropertyID] = *txPrivacyCustomToken
	}
	resultCrossShard := make(map[common.Hash]CrossShardTokenPrivacyMetaData)
	for _, tokenData := range crossShardData {
		crossShardTokenPrivacyMetaData := CrossShardTokenPrivacyMetaData{}
		err = json.Unmarshal(tokenData, &crossShardTokenPrivacyMetaData)
		if err != nil {
			return nil, nil, err
		}
		resultCrossShard[crossShardTokenPrivacyMetaData.TokenID] = crossShardTokenPrivacyMetaData
	}
	return result, resultCrossShard, nil
}

// GetCustomTokenTxsHash - return list hash of tx which relate to custom token
func (blockchain *BlockChain) GetCustomTokenTxsHash(tokenID *common.Hash) ([]common.Hash, error) {
	txHashesInByte, err := blockchain.config.DataBase.CustomTokenTxs(*tokenID)
	if err != nil {
		return nil, err
	}
	result := []common.Hash{}
	for _, temp := range txHashesInByte {
		result = append(result, temp)
	}
	return result, nil
}

// GetPrivacyCustomTokenTxsHash - return list hash of tx which relate to custom token
func (blockchain *BlockChain) GetPrivacyCustomTokenTxsHash(tokenID *common.Hash) ([]common.Hash, error) {
	txHashesInByte, err := blockchain.config.DataBase.PrivacyCustomTokenTxs(*tokenID)
	if err != nil {
		return nil, err
	}
	result := []common.Hash{}
	for _, temp := range txHashesInByte {
		result = append(result, temp)
	}
	return result, nil
}

// GetListTokenHolders - return list paymentaddress (in hexstring) of someone who hold custom token in network
func (blockchain *BlockChain) GetListTokenHolders(tokenID *common.Hash) (map[string]uint64, error) {
	result, err := blockchain.config.DataBase.GetCustomTokenPaymentAddressesBalance(*tokenID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *BlockChain) GetCurrentBeaconBlockHeight(shardID byte) uint64 {
	return self.BestState.Beacon.BestBlock.Header.Height
}

func (blockchain BlockChain) RandomCommitmentsProcess(usableInputCoins []*privacy.InputCoin, randNum int, shardID byte, tokenID *common.Hash) (commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) {
	return transaction.RandomCommitmentsProcess(usableInputCoins, randNum, blockchain.config.DataBase, shardID, tokenID)
}

func (blockchain BlockChain) CheckSNDerivatorExistence(tokenID *common.Hash, snd *big.Int, shardID byte) (bool, error) {
	return transaction.CheckSNDerivatorExistence(tokenID, snd, shardID, blockchain.config.DataBase)
}

// func (blockchain *BlockChain) SetReadyState(shard bool, shardID byte, ready bool) {
// 	// fmt.Println("SetReadyState", shard, shardID, ready)
// 	blockchain.syncStatus.IsReady.Lock()
// 	defer blockchain.syncStatus.IsReady.Unlock()
// 	if shard {
// 		blockchain.syncStatus.IsReady.Shards[shardID] = ready
// 	} else {
// 		blockchain.syncStatus.IsReady.Beacon = ready
// 		if ready {
// 			fmt.Println("blockchain is ready")
// 		}
// 	}
// }

// func (blockchain *BlockChain) IsReady(shard bool, shardID byte) bool {
// 	blockchain.syncStatus.IsReady.Lock()
// 	defer blockchain.syncStatus.IsReady.Unlock()
// 	if shard {
// 		if _, ok := blockchain.syncStatus.IsReady.Shards[shardID]; !ok {
// 			return false
// 		}
// 		return blockchain.syncStatus.IsReady.Shards[shardID]
// 	}
// 	return blockchain.syncStatus.IsReady.Beacon
// }

// func (blockchain *BlockChain) BuildAcceptRewardInstructions(rewardInstructionsRequest [][]string, beaconHeight uint64, beaconPaymentAddress *privacy.PaymentAddress) ([][]string, error) {
// 	resIns := [][]string{}
// 	totalBeaconReward := blockchain.getRewardAmount(beaconHeight)
// 	var shardRewards map[string]uint64
// 	var shardPaymentAddress map[string]*privacy.PaymentAddress
// 	for _, rewardRequestIns := range rewardInstructionsRequest {
// 		rewardRequest, err := metadata.NewShardBlockSalaryRequestFromStr(rewardRequestIns[2])
// 		if err != nil {
// 			return [][]string{}, err
// 		}
// 		totalBeaconReward += rewardRequest.TxsFeeForBeacon
// 		shardAddressStr := rewardRequest.PayToAddress.String()
// 		if shardRewards[shardAddressStr] == 0 {
// 			shardPaymentAddress[shardAddressStr] = rewardRequest.PayToAddress
// 		}
// 		shardRewards[shardAddressStr] += rewardRequest.TxsFeeForShard + blockchain.getRewardAmount(rewardRequest.ShardBlockHeight)
// 	}
// 	beaconRewardIns, err := metadata.BuildInstForBeaconSalary(totalBeaconReward, beaconHeight, beaconPaymentAddress)
// 	if err != nil {
// 		return [][]string{}, err
// 	}
// 	resIns = append(resIns, beaconRewardIns)
// 	return resIns, nil
// }

//TODO implement more logic for fair
func (blockchain *BlockChain) BuildInstRewardForBeacons(epoch uint64, totalReward map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	baseRewards := map[common.Hash]uint64{}
	for key, value := range totalReward {
		baseRewards[key] = value / uint64(blockchain.BestState.Beacon.BeaconCommitteeSize)
	}
	for _, publickeyStr := range blockchain.BestState.Beacon.BeaconCommittee {
		singleInst, err := metadata.BuildInstForBeaconReward(baseRewards, publickeyStr)
		if err != nil {
			Logger.log.Errorf("BuildInstForBeaconReward error %+v\n Totalreward: %+v, epoch: %+v, reward: %+v\n", err, totalReward, epoch, baseRewards)
			return nil, err
		}
		resInst = append(resInst, singleInst)
	}
	return resInst, nil
}

func (blockchain *BlockChain) GetAllCoinID() ([]common.Hash, error) {
	mapCustomToken, err := blockchain.ListCustomToken()
	if err != nil {
		return nil, err
	}
	mapPrivacyCustomToken, mapCrossShardCustomToken, err := blockchain.ListPrivacyCustomToken()
	if err != nil {
		return nil, err
	}
	mapBridgeTokenID, err := blockchain.GetDatabase().GetBridgeTokensAmounts()
	if err != nil {
		return nil, err
	}
	allCoinID := make([]common.Hash, len(mapCustomToken)+len(mapPrivacyCustomToken)+len(mapCrossShardCustomToken)+len(mapBridgeTokenID)+1)
	allCoinID[0] = common.PRVCoinID
	index := 1
	for key := range mapCustomToken {
		allCoinID[index] = key
		index++
	}
	for key := range mapPrivacyCustomToken {
		allCoinID[index] = key
		index++
	}
	for key := range mapCrossShardCustomToken {
		allCoinID[index] = key
		index++
	}

	for _, bridgeTokenIDBytes := range mapBridgeTokenID {
		var tokenWithAmount lvdb.TokenWithAmount
		err := json.Unmarshal(bridgeTokenIDBytes, &tokenWithAmount)
		if err != nil {
			return nil, err
		}
		allCoinID[index] = *tokenWithAmount.TokenID
		index++
	}
	return allCoinID, nil
}

func (blockchain *BlockChain) BuildInstRewardForDev(epoch uint64, totalReward map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	devRewardInst, err := metadata.BuildInstForDevReward(totalReward)
	if err != nil {
		Logger.log.Errorf("BuildInstRewardForDev error %+v\n Totalreward: %+v, epoch: %+v\n", err, totalReward, epoch)
		return nil, err
	}
	resInst = append(resInst, devRewardInst)
	return resInst, nil
}

func (blockchain *BlockChain) BuildInstRewardForShards(epoch uint64, totalRewards []map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	for i, reward := range totalRewards {
		if len(reward) > 0 {
			shardRewardInst, err := metadata.BuildInstForShardReward(reward, epoch, byte(i))
			if err != nil {
				Logger.log.Errorf("BuildInstForShardReward error %+v\n Totalreward: %+v, epoch: %+v\n; shard:%+v", err, reward, epoch, byte(i))
				return nil, err
			}
			resInst = append(resInst, shardRewardInst...)
		}
	}
	return resInst, nil
}

func (blockchain *BlockChain) BuildResponseTransactionFromTxsWithMetadata(blkBody *ShardBody, blkProducerPrivateKey *privacy.PrivateKey) error {
	txRequestTable := map[string]metadata.Transaction{}
	txsRes := []metadata.Transaction{}
	for _, tx := range blkBody.Transactions {
		if tx.GetMetadataType() == metadata.WithDrawRewardRequestMeta {
			requestMeta := tx.GetMetadata().(*metadata.WithDrawRewardRequest)
			requester := base58.Base58Check{}.Encode(requestMeta.PaymentAddress.Pk, VERSION)
			txRequestTable[requester] = tx
		}
	}
	for _, value := range txRequestTable {
		txRes, err := blockchain.buildWithDrawTransactionResponse(&value, blkProducerPrivateKey)
		if err != nil {
			fmt.Printf("[ndh] - buildWithDrawTransactionResponse for tx %+v, error: %+v\n", value, err)
			return err
		} else {
			fmt.Printf("[ndh] - buildWithDrawTransactionResponse for tx %+v, ok: %+v\n", value, txRes)
		}
		txsRes = append(txsRes, txRes)
	}
	blkBody.Transactions = append(blkBody.Transactions, txsRes...)
	return nil
}

func (blockchain *BlockChain) ValidateResponseTransactionFromTxsWithMetadata(blkBody *ShardBody) error {
	txRequestTable := map[string]metadata.Transaction{}
	for _, tx := range blkBody.Transactions {
		if tx.GetMetadataType() == metadata.WithDrawRewardRequestMeta {
			requestMeta := tx.GetMetadata().(*metadata.WithDrawRewardRequest)
			requester := base58.Base58Check{}.Encode(requestMeta.PaymentAddress.Pk, VERSION)
			txRequestTable[requester] = tx
		}
	}
	numberOfTxRequest := len(txRequestTable)
	db := blockchain.config.DataBase
	numberOfTxResponse := 0
	for _, tx := range blkBody.Transactions {
		if tx.GetMetadataType() == metadata.WithDrawRewardResponseMeta {
			_, requesterRes, amountRes, coinID := tx.GetTransferData()
			fmt.Printf("[ndh] - response %+v\n", tx)

			requester := base58.Base58Check{}.Encode(requesterRes, VERSION)
			if txRequestTable[requester] == nil {
				fmt.Printf("[ndh] - - [error] This response dont match with any request %+v \n", requester)
				return errors.New("This response dont match with any request")
			}
			requestMeta := txRequestTable[requester].GetMetadata().(*metadata.WithDrawRewardRequest)
			if res, err := coinID.Cmp(&requestMeta.TokenID); err == nil && res != 0 {
				return errors.New("Invalid token ID")
			}
			amount, err := db.GetCommitteeReward(requesterRes, requestMeta.TokenID)
			if (amount == 0) || (err != nil) {
				fmt.Printf("[ndh] - - [error] Not enough reward %+v %+v\n", amount, err)
				return errors.New("Not enough reward")
			}
			if amount != amountRes {
				fmt.Printf("[ndh] - - [error] Wrong amount %+v %+v\n", amount, amountRes)
				return errors.New("Wrong amount")
			}

			if res, err := txRequestTable[requester].Hash().Cmp(tx.GetMetadata().Hash()); err == nil && res != 0 {
				fmt.Printf("[ndh] - - [error] This response dont match with any request %+v %+v\n", amount, amountRes)
				return errors.New("This response dont match with any request")
			}
			txRequestTable[requester] = nil

			numberOfTxResponse++
		}
	}
	if numberOfTxRequest != numberOfTxResponse {
		fmt.Printf("[ndh] - - [error] Not match request and response %+v %+v\n", numberOfTxRequest, numberOfTxResponse)
		return errors.New("Not match request and response")
	}
	return nil
}

func (blockchain *BlockChain) InitTxSalaryByCoinID(
	payToAddress *privacy.PaymentAddress,
	amount uint64,
	payByPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	meta metadata.Metadata,
	coinID common.Hash,
	shardID byte,
) (metadata.Transaction, error) {
	txType := -1
	if res, err := coinID.Cmp(&common.PRVCoinID); err == nil && res == 0 {
		txType = transaction.NormalCoinType
	}
	if txType == -1 {
		mapBridgeTokenID, err := blockchain.GetDatabase().GetBridgeTokensAmounts()
		if err != nil {
			return nil, err
		}
		for _, bridgeTokenIDBytes := range mapBridgeTokenID {
			var tokenWithAmount lvdb.TokenWithAmount
			err := json.Unmarshal(bridgeTokenIDBytes, &tokenWithAmount)
			if err != nil {
				return nil, err
			}

			if res, err := coinID.Cmp(tokenWithAmount.TokenID); err == nil && res == 0 {
				txType = transaction.CustomTokenPrivacyType
				fmt.Printf("[ndh] eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee %+v \n", tokenWithAmount.TokenID)
				break
			}
		}
	}
	if txType == -1 {
		mapCustomToken, err := blockchain.ListCustomToken()
		if err != nil {
			return nil, err
		}
		if mapCustomToken != nil {
			if _, ok := mapCustomToken[coinID]; ok {
				txType = transaction.CustomTokenType
			}
		}
	}
	if txType == -1 {
		mapPrivacyCustomToken, _, err := blockchain.ListPrivacyCustomToken()
		if err != nil {
			return nil, err
		}
		if mapPrivacyCustomToken != nil {
			if _, ok := mapPrivacyCustomToken[coinID]; ok {
				txType = transaction.CustomTokenPrivacyType
			}
		}
	}
	if txType == -1 {
		return nil, errors.New("Invalid token ID")
	}
	return transaction.BuildCoinbaseTxByCoinID(
		payToAddress,
		amount,
		payByPrivateKey,
		db,
		meta,
		coinID,
		txType,
		coinID.String(),
		shardID,
	)
}

func CalculateNumberOfByteToRead(amountBytes int) []byte {
	var result = make([]byte,8)
	binary.LittleEndian.PutUint32(result, uint32(amountBytes))
	return result
}
func GetNumberOfByteToRead(value []byte) (int,error) {
	var result uint32
	err := binary.Read(bytes.NewBuffer(value), binary.LittleEndian, &result)
	if err != nil {
		return -1, err
	}
	return int(result), nil
}
func (blockchain *BlockChain) BackupShardChain(writer io.Writer, shardID byte) error {
	bestStateBytes, err := blockchain.config.DataBase.FetchShardBestState(shardID)
	if err != nil {
		return err
	}
		shardBestState := &BestStateShard{}
		err = json.Unmarshal(bestStateBytes, shardBestState)
		bestShardHeight := shardBestState.ShardHeight
		var i uint64
		for i = 1;i < bestShardHeight; i++{
			block, err := blockchain.GetShardBlockByHeight(i, shardID)
			if err != nil {
				return err
			}
			data, err := json.Marshal(block)
			if err != nil {
				return err
			}
			log.Printf("Byte len block %+v: %+v \n", i, len(data))
			_, err = writer.Write(CalculateNumberOfByteToRead(len(data)))
			if err != nil {
				return err
			}
			_, err = writer.Write(data)
			if err != nil {
				return err
			}
		}
	return nil
}
func (blockchain *BlockChain) RestoreShardChain(reader io.Reader, shardID byte) error {
	
	return nil
}