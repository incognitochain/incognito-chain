package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/btc"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/transaction"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

type BlockChain struct {
	Chains    map[string]ChainInterface
	BestState *BestState
	config    Config
	chainLock sync.Mutex

	cQuitSync        chan struct{}
	Synker           Synker
	ConsensusOngoing bool
	//RPCClient        *rpccaller.RPCClient
	IsTest bool
}

type BestState struct {
	Beacon *BeaconBestState
	Shard  map[byte]*ShardBestState
}

// config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	DataBase          database.DatabaseInterface
	MemCache          *memcache.MemoryCache
	Interrupt         <-chan struct{}
	ChainParams       *Params
	RelayShards       []byte
	NodeMode          string
	ShardToBeaconPool ShardToBeaconPool
	BlockGen          *BlockGenerator
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
		PublishNodeState(userLayer string, shardID int) error

		PushMessageGetBlockBeaconByHeight(from uint64, to uint64) error
		PushMessageGetBlockBeaconByHash(blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
		PushMessageGetBlockBeaconBySpecificHeight(heights []uint64, getFromPool bool) error

		PushMessageGetBlockShardByHeight(shardID byte, from uint64, to uint64) error
		PushMessageGetBlockShardByHash(shardID byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
		PushMessageGetBlockShardBySpecificHeight(shardID byte, heights []uint64, getFromPool bool) error

		PushMessageGetBlockShardToBeaconByHeight(shardID byte, from uint64, to uint64) error
		PushMessageGetBlockShardToBeaconByHash(shardID byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
		PushMessageGetBlockShardToBeaconBySpecificHeight(shardID byte, blksHeight []uint64, getFromPool bool, peerID libp2p.ID) error

		PushMessageGetBlockCrossShardByHash(fromShard byte, toShard byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
		PushMessageGetBlockCrossShardBySpecificHeight(fromShard byte, toShard byte, blksHeight []uint64, getFromPool bool, peerID libp2p.ID) error
		UpdateConsensusState(role string, userPbk string, currentShard *byte, beaconCommittee []string, shardCommittee map[byte][]string)
		PushBlockToAll(block common.BlockInterface, isBeacon bool) error
	}
	// UserKeySet *incognitokey.KeySet

	ConsensusEngine interface {
		ValidateProducerSig(block common.BlockInterface, consensusType string) error
		ValidateBlockCommitteSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey, consensusType string) error
		GetCurrentMiningPublicKey() (string, string)
		GetMiningPublicKeyByConsensus(consensusName string) (string, error)
		GetUserLayer() (string, int)
		GetUserRole() (string, string, int)
		IsOngoing(chainName string) bool
		CommitteeChange(chainName string)
	}

	Highway interface {
		BroadcastCommittee(uint64, []incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey)
	}
}

func NewBlockChain(config *Config, isTest bool) *BlockChain {
	bc := &BlockChain{}
	bc.config = *config
	bc.config.IsBlockGenStarted = false
	bc.IsTest = isTest
	bc.cQuitSync = make(chan struct{})
	bc.BestState = &BestState{
		Beacon: &BeaconBestState{},
		Shard:  make(map[byte]*ShardBestState),
	}
	for i := 0; i < 255; i++ {
		shardID := byte(i)
		bc.BestState.Shard[shardID] = &ShardBestState{}
	}
	bc.BestState.Beacon.Params = make(map[string]string)
	bc.BestState.Beacon.ShardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	bc.BestState.Beacon.ShardPendingValidator = make(map[byte][]incognitokey.CommitteePublicKey)
	bc.Synker = Synker{
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
	blockchain.Synker = newSyncker(blockchain.cQuitSync, blockchain, blockchain.config.PubSubManager)
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
	blockchain.Chains = make(map[string]ChainInterface)
	blockchain.BestState = &BestState{
		Beacon: nil,
		Shard:  make(map[byte]*ShardBestState),
	}

	bestStateBeaconBytes, err := blockchain.config.DataBase.FetchBeaconBestState()
	if err == nil {
		beacon := &BeaconBestState{}
		err = json.Unmarshal(bestStateBeaconBytes, beacon)
		//update singleton object
		SetBeaconBestState(beacon)
		//update beacon field in blockchain Beststate
		blockchain.BestState.Beacon = GetBeaconBestState()

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
	beaconChain := BeaconChain{
		BestState:  GetBeaconBestState(),
		BlockGen:   blockchain.config.BlockGen,
		ChainName:  common.BeaconChainKey,
		Blockchain: blockchain,
	}
	blockchain.Chains[common.BeaconChainKey] = &beaconChain

	for shard := 1; shard <= blockchain.BestState.Beacon.ActiveShards; shard++ {
		shardID := byte(shard - 1)
		bestStateBytes, err := blockchain.config.DataBase.FetchShardBestState(shardID)
		if err == nil {
			shardBestState := &ShardBestState{}
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
		shardChain := ShardChain{
			BestState:  GetBestStateShard(shardID),
			BlockGen:   blockchain.config.BlockGen,
			ChainName:  common.GetShardChainKey(shardID),
			Blockchain: blockchain,
		}
		blockchain.Chains[shardChain.ChainName] = &shardChain
	}

	return nil
}

/*
// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
*/
func (blockchain *BlockChain) initShardState(shardID byte) error {
	blockchain.BestState.Shard[shardID] = NewBestStateShardWithConfig(shardID, blockchain.config.ChainParams)
	// Create a new block from genesis block and set it as best block of chain
	initBlock := ShardBlock{}
	initBlock = *blockchain.config.ChainParams.GenesisShardBlock
	initBlock.Header.ShardID = shardID

	_, newShardCandidate := GetStakingCandidate(*blockchain.config.ChainParams.GenesisBeaconBlock)
	newShardCandidateStructs := []incognitokey.CommitteePublicKey{}
	for _, candidate := range newShardCandidate {
		key := incognitokey.CommitteePublicKey{}
		err := key.FromBase58(candidate)
		if err != nil {
			return err
		}
		newShardCandidateStructs = append(newShardCandidateStructs, key)
	}

	blockchain.BestState.Shard[shardID].ShardCommittee = append(blockchain.BestState.Shard[shardID].ShardCommittee, newShardCandidateStructs[int(shardID)*blockchain.config.ChainParams.MinShardCommitteeSize:(int(shardID)*blockchain.config.ChainParams.MinShardCommitteeSize)+blockchain.config.ChainParams.MinShardCommitteeSize]...)

	genesisBeaconBlock, err := blockchain.GetBeaconBlockByHeight(1)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlockError, err)
	}
	err = blockchain.BestState.Shard[shardID].initShardBestState(blockchain, &initBlock, genesisBeaconBlock)
	if err != nil {
		return err
	}
	err = blockchain.processStoreShardBlockAndUpdateDatabase(&initBlock)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) initBeaconState() error {
	blockchain.BestState.Beacon = NewBeaconBestStateWithConfig(blockchain.config.ChainParams)
	initBlock := blockchain.config.ChainParams.GenesisBeaconBlock
	err := blockchain.BestState.Beacon.initBeaconBestState(initBlock)
	if err != nil {
		return err
	}
	// Insert new block into beacon chain
	if err := blockchain.StoreBeaconBestState(nil); err != nil {
		Logger.log.Error("Error Store best state for block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}
	if err := blockchain.config.DataBase.StoreBeaconBlock(&blockchain.BestState.Beacon.BestBlock, blockchain.BestState.Beacon.BestBlock.Header.Hash(), nil); err != nil {
		Logger.log.Error("Error store beacon block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return err
	}
	if err := blockchain.config.DataBase.StoreShardCommitteeByHeight(initBlock.Header.Height, blockchain.BestState.Beacon.GetShardCommittee()); err != nil {
		return err
	}
	if err := blockchain.config.DataBase.StoreBeaconCommitteeByHeight(initBlock.Header.Height, blockchain.BestState.Beacon.BeaconCommittee); err != nil {
		return err
	}
	blockHash := initBlock.Hash()
	if err := blockchain.config.DataBase.StoreBeaconBlockIndex(*blockHash, initBlock.Header.Height); err != nil {
		return err
	}
	return nil
}

func (bestState BestState) GetClonedBeaconBestState() (*BeaconBestState, error) {
	result := NewBeaconBestState()
	err := result.cloneBeaconBestStateFrom(bestState.Beacon)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetReadOnlyShard - return a copy of Shard of BestState
func (bestState BestState) GetClonedAllShardBestState() map[byte]*ShardBestState {
	result := make(map[byte]*ShardBestState)
	for k, v := range bestState.Shard {
		v.lock.RLock()
		result[k] = &ShardBestState{}
		err := result[k].cloneShardBestStateFrom(v)
		if err != nil {
			Logger.log.Error(err)
		}
		v.lock.RUnlock()
	}
	return result
}

// GetReadOnlyShard - return a copy of Shard of BestState
func (bestState *BestState) GetClonedAShardBestState(shardID byte) (*ShardBestState, error) {
	shardBestState := NewShardBestState()
	if target, ok := bestState.Shard[shardID]; !ok {
		return shardBestState, fmt.Errorf("Failed to get Shard BestState of ShardID %+v", shardID)
	} else {
		target.lock.RLock()
		defer target.lock.RUnlock()
		if err := shardBestState.cloneShardBestStateFrom(target); err != nil {
			return shardBestState, fmt.Errorf("Failed to clone Shard BestState of ShardID %+v", shardID)
		}
	}
	return shardBestState, nil
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
	beaconBlockHash, err := blockchain.config.DataBase.GetBeaconBlockHashByIndex(height)
	if err != nil {
		return nil, err
	}
	beaconBlock, _, err := blockchain.GetBeaconBlockByHash(beaconBlockHash)
	if err != nil {
		return nil, err
	}
	return beaconBlock, nil
}

/*
Fetch DatabaseInterface and get block data by block hash
*/
func (blockchain *BlockChain) GetBeaconBlockByHash(beaconBlockHash common.Hash) (*BeaconBlock, uint64, error) {
	if blockchain.IsTest {
		return &BeaconBlock{}, 2, nil
	}
	beaconBlockBytes, err := blockchain.config.DataBase.FetchBeaconBlock(beaconBlockHash)
	if err != nil {
		return nil, 0, err
	}
	beaconBlock := NewBeaconBlock()
	err = json.Unmarshal(beaconBlockBytes, beaconBlock)
	if err != nil {
		return nil, 0, err
	}
	return beaconBlock, uint64(len(beaconBlockBytes)), nil
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
func (blockchain *BlockChain) StoreBeaconBestState(bd *[]database.BatchData) error {
	return blockchain.config.DataBase.StoreBeaconBestState(blockchain.BestState.Beacon, bd)
}

/*
Store best state of block(best block, num of tx, ...) into Database
*/
func (blockchain *BlockChain) StoreShardBestState(shardID byte, bd *[]database.BatchData) error {
	return blockchain.config.DataBase.StoreShardBestState(blockchain.BestState.Shard[shardID], shardID, bd)
}

/*
GetBestState - return a best state from a chain
*/
// #1 - shardID - index of chain
func (blockchain *BlockChain) GetShardBestState(shardID byte) (*ShardBestState, error) {
	bestState := ShardBestState{}
	bestStateBytes, err := blockchain.config.DataBase.FetchShardBestState(shardID)
	if err == nil {
		err = json.Unmarshal(bestStateBytes, &bestState)
	}
	return &bestState, err
}

/*
Store block into Database
*/
func (blockchain *BlockChain) StoreShardBlock(block *ShardBlock, bd *[]database.BatchData) error {
	return blockchain.config.DataBase.StoreShardBlock(block, block.Header.Hash(), block.Header.ShardID, bd)
}

/*
Save index(height) of block by block hash
and
Save block hash by index(height) of block
*/
func (blockchain *BlockChain) StoreShardBlockIndex(block *ShardBlock, bd *[]database.BatchData) error {
	return blockchain.config.DataBase.StoreShardBlockIndex(block.Header.Hash(), block.Header.Height, block.Header.ShardID, bd)
}

func (blockchain *BlockChain) StoreTransactionIndex(txHash *common.Hash, blockHash common.Hash, index int, bd *[]database.BatchData) error {
	return blockchain.config.DataBase.StoreTransactionIndex(*txHash, blockHash, index, bd)
}

/*
Uses an existing database to update the set of used tx by saving list serialNumber of privacy,
this is a list tx-out which are used by a new tx
*/
func (blockchain *BlockChain) StoreSerialNumbersFromTxViewPoint(view TxViewPoint) error {
	if len(view.listSerialNumbers) > 0 {
		err := blockchain.config.DataBase.StoreSerialNumbers(*view.tokenID, view.listSerialNumbers, view.shardID)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Uses an existing database to update the set of used tx by saving list SNDerivator of privacy,
this is a list tx-out which are used by a new tx
*/
func (blockchain *BlockChain) StoreSNDerivatorsFromTxViewPoint(view TxViewPoint) error {
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
		err := blockchain.config.DataBase.StoreSNDerivators(*view.tokenID, snDsArray)
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

// StoreTxByPublicKey - store txID by public key of receiver,
// use this data to get tx which send to receiver, because we can get this tx from cross shard
// -> only fullnode data can provide this data for all
func (blockchain *BlockChain) StoreTxByPublicKey(view *TxViewPoint) error {
	for data := range view.txByPubKey {
		dataArr := strings.Split(data, "_")
		pubKey, _, err := base58.Base58Check{}.Decode(dataArr[0])
		if err != nil {
			return err
		}
		txIDInByte, _, err := base58.Base58Check{}.Decode(dataArr[1])
		if err != nil {
			return err
		}
		txID := common.Hash{}
		err = txID.SetBytes(txIDInByte)
		if err != nil {
			return err
		}
		shardID, _ := strconv.Atoi(dataArr[2])

		err = blockchain.config.DataBase.StoreTxByPublicKey(pubKey, txID, byte(shardID))
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
			// clear cached data
			if blockchain.config.MemCache != nil {
				cachedKey := memcache.GetListOutputcoinCachedKey(publicKeyBytes, view.tokenID, publicKeyShardID)
				if ok, e := blockchain.config.MemCache.Has(cachedKey); ok && e == nil {
					er := blockchain.config.MemCache.Delete(cachedKey)
					if er != nil {
						Logger.log.Error("can not delete memcache", "GetListOutputcoinCachedKey", base58.Base58Check{}.Encode(cachedKey, 0x0))
					}
				}
			}
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
func (blockchain *BlockChain) CreateAndSaveTxViewPointFromBlock(block *ShardBlock, bd *[]database.BatchData) error {
	//startTime := time.Now()
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchTxViewPointFromBlock(blockchain.config.DataBase, block)
	if err != nil {
		return err
	}

	// check privacy custom token
	// sort by index
	indices := []int{}
	for index := range view.privacyCustomTokenViewPoint {
		indices = append(indices, int(index))
	}
	sort.Ints(indices)
	for _, indexTx := range indices {
		privacyCustomTokenSubView := view.privacyCustomTokenViewPoint[int32(indexTx)]
		privacyCustomTokenTx := view.privacyCustomTokenTxs[int32(indexTx)]
		switch privacyCustomTokenTx.TxPrivacyTokenData.Type {
		case transaction.CustomTokenInit:
			{
				// check is bridge token
				isBridgeToken := false
				allBridgeTokensBytes, err := blockchain.config.DataBase.GetAllBridgeTokens()
				if err != nil {
					return err
				}
				if len(allBridgeTokensBytes) > 0 {
					var allBridgeTokens []*lvdb.BridgeTokenInfo
					err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
					for _, bridgeToken := range allBridgeTokens {
						if bridgeToken.TokenID != nil && bytes.Equal(privacyCustomTokenTx.TxPrivacyTokenData.PropertyID[:], bridgeToken.TokenID[:]) {
							isBridgeToken = true
						}
					}
				}
				// not mintable tx
				if !isBridgeToken && !privacyCustomTokenTx.TxPrivacyTokenData.Mintable {
					Logger.log.Info("Store custom token when it is issued", privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, privacyCustomTokenTx.TxPrivacyTokenData.PropertySymbol, privacyCustomTokenTx.TxPrivacyTokenData.PropertyName)
					err = blockchain.config.DataBase.StorePrivacyToken(privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, privacyCustomTokenTx.Hash()[:])
					if err != nil {
						return err
					}
				}
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Info("Transfer custom token %+v", privacyCustomTokenTx)
			}
		}
		err = blockchain.config.DataBase.StorePrivacyTokenTx(privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, block.Header.ShardID, block.Header.Height, int32(indexTx), privacyCustomTokenTx.Hash()[:])
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

		err = blockchain.StoreSNDerivatorsFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}

	// updateShardBestState the list serialNumber and commitment, snd set using the state of the used tx view point. This
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

	err = blockchain.StoreSNDerivatorsFromTxViewPoint(*view)
	if err != nil {
		return err
	}

	err = blockchain.StoreTxByPublicKey(view)
	if err != nil {
		return err
	}
	//endtime := time.Now()
	//runTime := endtime.Sub(startTime)
	//go common.AnalyzeFuncCreateAndSaveTxViewPointFromBlock(runTime.Seconds())
	//Logger.log.Critical("*** CreateAndSaveTxViewPointFromBlock  ***", block.Header.Height, runTime)

	return nil
}

func (blockchain *BlockChain) CreateAndSaveCrossTransactionCoinViewPointFromBlock(block *ShardBlock, bd *[]database.BatchData) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlock(blockchain.config.DataBase, block)
	if err != nil {
		Logger.log.Error("CreateAndSaveCrossTransactionCoinViewPointFromBlock", err)
		return err
	}

	// sort by index
	indices := []int{}
	for index := range view.privacyCustomTokenViewPoint {
		indices = append(indices, int(index))
	}
	sort.Ints(indices)

	for _, index := range indices {
		privacyCustomTokenSubView := view.privacyCustomTokenViewPoint[int32(index)]
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

				if err := blockchain.config.DataBase.StorePrivacyTokenCrossShard(*tokenID, tokenDataBytes); err != nil {
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
		err = blockchain.StoreSNDerivatorsFromTxViewPoint(*privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}

	// updateShardBestState the list serialNumber and commitment, snd set using the state of the used tx view point. This
	// entails adding the new
	// ones created by the block.
	err = blockchain.StoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
	if err != nil {
		return err
	}

	err = blockchain.StoreSNDerivatorsFromTxViewPoint(*view)
	if err != nil {
		return err
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
	pubkeyCompress := outCoinTemp.CoinDetails.GetPublicKey().ToBytesS()
	if bytes.Equal(pubkeyCompress, keySet.PaymentAddress.Pk[:]) {
		result := &privacy.OutputCoin{
			CoinDetails:          outCoinTemp.CoinDetails,
			CoinDetailsEncrypted: outCoinTemp.CoinDetailsEncrypted,
		}
		if result.CoinDetailsEncrypted != nil && !result.CoinDetailsEncrypted.IsNil() {
			if len(keySet.ReadonlyKey.Rk) > 0 {
				// try to decrypt to get more data
				err := result.Decrypt(keySet.ReadonlyKey)
				if err != nil {
					return nil
				}
			}
		}
		if len(keySet.PrivateKey) > 0 {
			// check spent with private-key
			result.CoinDetails.SetSerialNumber(
				new(privacy.Point).Derive(
					privacy.PedCom.G[privacy.PedersenPrivateKeyIndex],
					new(privacy.Scalar).FromBytesS(keySet.PrivateKey),
					result.CoinDetails.GetSNDerivator()))
			ok, err := blockchain.config.DataBase.HasSerialNumber(*tokenID, result.CoinDetails.GetSerialNumber().ToBytesS(), shardID)
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

	var outCointsInBytes [][]byte
	var err error
	if keyset == nil {
		return nil, NewBlockChainError(UnExpectedError, errors.New("Invalid keyset"))
	}
	if blockchain.config.MemCache != nil {
		// get from cache
		cachedKey := memcache.GetListOutputcoinCachedKey(keyset.PaymentAddress.Pk[:], tokenID, shardID)
		cachedData, _ := blockchain.config.MemCache.Get(cachedKey)
		if cachedData != nil && len(cachedData) > 0 {
			// try to parsing on outCointsInBytes
			_ = json.Unmarshal(cachedData, &outCointsInBytes)
		}
		if len(outCointsInBytes) == 0 {
			// cached data is nil or fail -> get from database
			outCointsInBytes, err = blockchain.config.DataBase.GetOutcoinsByPubkey(*tokenID, keyset.PaymentAddress.Pk[:], shardID)
			if len(outCointsInBytes) > 0 {
				// cache 1 day for result
				cachedData, err = json.Marshal(outCointsInBytes)
				if err == nil {
					blockchain.config.MemCache.PutExpired(cachedKey, cachedData, 1*24*60*60*time.Millisecond)
				}
			}
		}
	}
	if len(outCointsInBytes) == 0 {
		outCointsInBytes, err = blockchain.config.DataBase.GetOutcoinsByPubkey(*tokenID, keyset.PaymentAddress.Pk[:], shardID)
		if err != nil {
			return nil, err
		}
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
		decryptedOut := blockchain.DecryptOutputCoinByKey(out, keyset, shardID, tokenID)
		if decryptedOut == nil {
			continue
		} else {
			results = append(results, decryptedOut)
		}
	}

	return results, nil
}

// GetTransactionByHash - retrieve tx from txId(txHash)
func (blockchain *BlockChain) GetTransactionByHash(txHash common.Hash) (byte, common.Hash, int, metadata.Transaction, error) {
	blockHash, index, err := blockchain.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		return byte(255), common.Hash{}, -1, nil, NewBlockChainError(UnExpectedError, err)
	}
	block, _, err1 := blockchain.GetShardBlockByHash(blockHash)
	if err1 != nil {
		//Logger.log.Errorf("ERROR", err1, "NO Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Body.Transactions[index])
		return byte(255), common.Hash{}, -1, nil, NewBlockChainError(UnExpectedError, err1)
	}
	//Logger.log.Infof("Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Transactions[index])
	return block.Header.ShardID, blockHash, index, block.Body.Transactions[index], nil
}

// GetTransactionHashByReceiver - return list tx id which receiver get from any sender
// this feature only apply on full node, because full node get all data from all shard
func (blockchain *BlockChain) GetTransactionHashByReceiver(keySet *incognitokey.KeySet) (map[byte][]common.Hash, error) {
	result := make(map[byte][]common.Hash)
	var err error
	result, err = blockchain.config.DataBase.GetTxByPublicKey(keySet.PaymentAddress.Pk)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	return result, nil
}

// Check Privacy Custom token ID is existed
func (blockchain *BlockChain) PrivacyCustomTokenIDExisted(tokenID *common.Hash) bool {
	return blockchain.config.DataBase.PrivacyTokenIDExisted(*tokenID)
}

func (blockchain *BlockChain) PrivacyCustomTokenIDCrossShardExisted(tokenID *common.Hash) bool {
	return blockchain.config.DataBase.PrivacyTokenIDCrossShardExisted(*tokenID)
}

// ListCustomToken - return all custom token which existed in network
func (blockchain *BlockChain) ListPrivacyCustomToken() (map[common.Hash]transaction.TxCustomTokenPrivacy, map[common.Hash]CrossShardTokenPrivacyMetaData, error) {
	data, err := blockchain.config.DataBase.ListPrivacyToken()
	if err != nil {
		return nil, nil, err
	}
	crossShardData, err := blockchain.config.DataBase.ListPrivacyTokenCrossShard()
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
		result[txPrivacyCustomToken.TxPrivacyTokenData.PropertyID] = *txPrivacyCustomToken
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

// GetPrivacyCustomTokenTxsHash - return list hash of tx which relate to custom token
func (blockchain *BlockChain) GetPrivacyCustomTokenTxsHash(tokenID *common.Hash) ([]common.Hash, error) {
	txHashesInByte, err := blockchain.config.DataBase.PrivacyTokenTxs(*tokenID)
	if err != nil {
		return nil, err
	}
	result := []common.Hash{}
	for _, temp := range txHashesInByte {
		result = append(result, temp)
	}
	return result, nil
}

func (blockchain *BlockChain) GetCurrentBeaconBlockHeight(shardID byte) uint64 {
	return blockchain.BestState.Beacon.BestBlock.Header.Height
}

func (blockchain BlockChain) RandomCommitmentsProcess(usableInputCoins []*privacy.InputCoin, randNum int, shardID byte, tokenID *common.Hash) (commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) {
	param := transaction.NewRandomCommitmentsProcessParam(usableInputCoins, randNum, blockchain.config.DataBase, shardID, tokenID)
	return transaction.RandomCommitmentsProcess(param)
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

//BuildInstRewardForBeacons create reward instruction for beacons
func (blockchain *BlockChain) BuildInstRewardForBeacons(epoch uint64, totalReward map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	baseRewards := map[common.Hash]uint64{}
	for key, value := range totalReward {
		baseRewards[key] = value / uint64(len(blockchain.BestState.Beacon.BeaconCommittee))
	}
	for _, beaconpublickey := range blockchain.BestState.Beacon.BeaconCommittee {
		// indicate reward pubkey
		singleInst, err := metadata.BuildInstForBeaconReward(baseRewards, beaconpublickey.GetNormalKey())
		if err != nil {
			Logger.log.Errorf("BuildInstForBeaconReward error %+v\n Totalreward: %+v, epoch: %+v, reward: %+v\n", err, totalReward, epoch, baseRewards)
			return nil, err
		}
		resInst = append(resInst, singleInst)
	}
	return resInst, nil
}

func (blockchain *BlockChain) GetAllCoinID() ([]common.Hash, error) {
	mapPrivacyCustomToken, mapCrossShardCustomToken, err := blockchain.ListPrivacyCustomToken()
	if err != nil {
		return nil, err
	}
	db := blockchain.GetDatabase()
	allBridgeTokensBytes, err := db.GetAllBridgeTokens()
	if err != nil {
		return nil, err
	}
	var allBridgeTokens []*lvdb.BridgeTokenInfo
	err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)

	if err != nil {
		return nil, err
	}
	allCoinID := make([]common.Hash, len(mapPrivacyCustomToken)+len(mapCrossShardCustomToken)+len(allBridgeTokens)+1)
	allCoinID[0] = common.PRVCoinID
	index := 1
	for key := range mapPrivacyCustomToken {
		allCoinID[index] = key
		index++
	}
	for key := range mapCrossShardCustomToken {
		allCoinID[index] = key
		index++
	}

	for _, bridgeTokens := range allBridgeTokens {
		allCoinID[index] = *bridgeTokens.TokenID
		index++
	}
	return allCoinID, nil
}

func (blockchain *BlockChain) BuildInstRewardForIncDAO(epoch uint64, totalReward map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	devRewardInst, err := metadata.BuildInstForIncDAOReward(totalReward, blockchain.config.ChainParams.IncognitoDAOAddress)
	if err != nil {
		Logger.log.Errorf("BuildInstRewardForIncDAO error %+v\n Totalreward: %+v, epoch: %+v\n", err, totalReward, epoch)
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

func (blockchain *BlockChain) BuildResponseTransactionFromTxsWithMetadata(
	transactions []metadata.Transaction,
	blkProducerPrivateKey *privacy.PrivateKey,
) (
	[]metadata.Transaction,
	error,
) {
	txRequestTable := reqTableFromReqTxs(transactions)
	txsResponse := []metadata.Transaction{}
	for key, value := range txRequestTable {
		txRes, err := blockchain.buildWithDrawTransactionResponse(&value, blkProducerPrivateKey)
		if err != nil {
			Logger.log.Errorf("Build Withdraw transactions response for tx %v return errors %v", value, err)
			delete(txRequestTable, key)
			continue
		} else {
			Logger.log.Infof("[Reward] - BuildWithDrawTransactionResponse for tx %+v, ok: %+v\n", value, txRes)
		}
		txsResponse = append(txsResponse, txRes)
	}
	txsSpamRemoved := filterReqTxs(transactions, txRequestTable)
	Logger.log.Infof("Number of metadata txs: %v; number of tx request %v; number of tx spam %v; number of tx response %v",
		len(transactions),
		len(txRequestTable),
		len(transactions)-len(txsSpamRemoved),
		len(txsResponse))
	txsSpamRemoved = append(txsSpamRemoved, txsResponse...)
	return txsSpamRemoved, nil
}

func (blockchain *BlockChain) ValidateResponseTransactionFromTxsWithMetadata(blk *ShardBlock) error {
	blkBody := blk.Body
	txRequestTable := reqTableFromReqTxs(blkBody.Transactions)
	// validate more than with spam request tx
	if blk.Header.Timestamp > ValidateTimeForSpamRequestTxs {
		txsSpamRemoved := filterReqTxs(blkBody.Transactions, txRequestTable)
		if len(blkBody.Transactions) != len(txsSpamRemoved) {
			return errors.Errorf("This block contains txs spam request reward. Number of spam: %v", len(blkBody.Transactions)-len(txsSpamRemoved))
		}
	}
	txReturnTable := map[string]bool{}
	db := blockchain.config.DataBase
	for _, tx := range blkBody.Transactions {
		switch tx.GetMetadataType() {
		case metadata.WithDrawRewardResponseMeta:
			_, requesterRes, amountRes, coinID := tx.GetTransferData()
			requester := getRequesterFromPKnCoinID(requesterRes, *coinID)
			txReq, isExist := txRequestTable[requester]
			if !isExist {
				return errors.New("This response dont match with any request")
			}
			requestMeta := txReq.GetMetadata().(*metadata.WithDrawRewardRequest)
			responseMeta := tx.GetMetadata().(*metadata.WithDrawRewardResponse)
			if res, err := coinID.Cmp(&requestMeta.TokenID); err != nil || res != 0 {
				return errors.Errorf("Invalid token ID when check metadata of tx response. Got %v, want %v, error %v", coinID, requestMeta.TokenID, err)
			}
			if cmp, err := responseMeta.TxRequest.Cmp(txReq.Hash()); (cmp != 0) || (err != nil) {
				Logger.log.Errorf("Response does not match with request, response link to txID %v, request txID %v, error %v", responseMeta.TxRequest.String(), txReq.Hash().String(), err)
			}
			amount, err := db.GetCommitteeReward(requesterRes, requestMeta.TokenID)
			if (amount == 0) || (err != nil) {
				return errors.Errorf("Invalid request %v, amount from db %v, error %v", requester, amount, err)
			}
			if amount != amountRes {
				return errors.Errorf("Wrong amount for token %v, get from db %v, response amount %v", requestMeta.TokenID, amount, amountRes)
			}
			delete(txRequestTable, requester)
			continue
		case metadata.ReturnStakingMeta:
			returnMeta := tx.GetMetadata().(*metadata.ReturnStakingMetadata)
			if _, ok := txReturnTable[returnMeta.StakerAddress.String()]; !ok {
				txReturnTable[returnMeta.StakerAddress.String()] = true
			} else {
				return errors.New("Double spent transaction return staking for a candidate.")
			}
		}
	}
	if blk.Header.Timestamp > ValidateTimeForSpamRequestTxs {
		if len(txRequestTable) > 0 {
			return errors.Errorf("Not match request and response, num of unresponse request: %v", len(txRequestTable))
		}
	}
	return nil
}

/*func (blockchain BlockChain) GetRPCClient() *rpccaller.RPCClient {
	return blockchain.RPCClient
}

func (blockchain *BlockChain) SetRPCClientChain(rpcClient *rpccaller.RPCClient) {
	blockchain.RPCClient = rpcClient
}*/

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

		db := blockchain.GetDatabase()
		allBridgeTokensBytes, err := db.GetAllBridgeTokens()
		if err != nil {
			return nil, err
		}
		var allBridgeTokens []*lvdb.BridgeTokenInfo
		err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)

		if err != nil {
			return nil, err
		}

		for _, bridgeTokenIDs := range allBridgeTokens {
			if res, err := coinID.Cmp(bridgeTokenIDs.TokenID); err == nil && res == 0 {
				txType = transaction.CustomTokenPrivacyType
				break
			}
		}
	}
	if txType == -1 {
		mapPrivacyCustomToken, mapCrossShardCustomToken, err := blockchain.ListPrivacyCustomToken()
		if err != nil {
			return nil, err
		}
		if mapPrivacyCustomToken != nil {
			if _, ok := mapPrivacyCustomToken[coinID]; ok {
				txType = transaction.CustomTokenPrivacyType
			}
		}
		if mapCrossShardCustomToken != nil {
			if _, ok := mapCrossShardCustomToken[coinID]; ok {
				txType = transaction.CustomTokenPrivacyType
			}
		}
	}
	if txType == -1 {
		return nil, errors.Errorf("Invalid token ID when InitTxSalaryByCoinID. Got %v", coinID)
	}
	buildCoinBaseParams := transaction.NewBuildCoinBaseTxByCoinIDParams(payToAddress,
		amount,
		payByPrivateKey,
		db,
		meta,
		coinID,
		txType,
		coinID.String(),
		shardID)
	return transaction.BuildCoinBaseTxByCoinID(buildCoinBaseParams)
}

func CalculateNumberOfByteToRead(amountBytes int) []byte {
	var result = make([]byte, 8)
	binary.LittleEndian.PutUint32(result, uint32(amountBytes))
	return result
}
func GetNumberOfByteToRead(value []byte) (int, error) {
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
	shardBestState := &ShardBestState{}
	err = json.Unmarshal(bestStateBytes, shardBestState)
	bestShardHeight := shardBestState.ShardHeight
	var i uint64
	for i = 1; i < bestShardHeight; i++ {
		block, err := blockchain.GetShardBlockByHeight(i, shardID)
		if err != nil {
			return err
		}
		data, err := json.Marshal(block)
		if err != nil {
			return err
		}
		_, err = writer.Write(CalculateNumberOfByteToRead(len(data)))
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		if err != nil {
			return err
		}
		if i%100 == 0 {
			log.Printf("Backup Shard %+v Block %+v", block.Header.ShardID, i)
		}
		if i == bestShardHeight-1 {
			log.Printf("Finish Backup Shard %+v with Block %+v", block.Header.ShardID, i)
		}
	}
	return nil
}
func (blockchain *BlockChain) BackupBeaconChain(writer io.Writer) error {
	bestStateBytes, err := blockchain.config.DataBase.FetchBeaconBestState()
	if err != nil {
		return err
	}
	beaconBestState := &BeaconBestState{}
	err = json.Unmarshal(bestStateBytes, beaconBestState)
	bestBeaconHeight := beaconBestState.BeaconHeight
	var i uint64
	for i = 1; i < bestBeaconHeight; i++ {
		block, err := blockchain.GetBeaconBlockByHeight(i)
		if err != nil {
			return err
		}
		data, err := json.Marshal(block)
		if err != nil {
			return err
		}
		numOfByteToRead := CalculateNumberOfByteToRead(len(data))
		_, err = writer.Write(numOfByteToRead)
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		if err != nil {
			return err
		}
		if i%100 == 0 {
			log.Printf("Backup Beacon Block %+v", i)
		}
		if i == bestBeaconHeight-1 {
			log.Printf("Finish Backup Beacon with Block %+v", i)
		}
	}
	return nil
}

func (blockchain *BlockChain) StoreIncomingCrossShard(block *ShardBlock, bd *[]database.BatchData) error {
	crossShardMap, _ := block.Body.ExtractIncomingCrossShardMap()
	for crossShard, crossBlks := range crossShardMap {
		for _, crossBlk := range crossBlks {
			err := blockchain.config.DataBase.StoreIncomingCrossShard(block.Header.ShardID, crossShard, block.Header.Height, crossBlk, bd)
			if err != nil {
				return NewBlockChainError(StoreIncomingCrossShardError, err)
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) DeleteIncomingCrossShard(block *ShardBlock) error {
	crossShardMap, _ := block.Body.ExtractIncomingCrossShardMap()
	for crossShard, crossBlks := range crossShardMap {
		for _, crossBlk := range crossBlks {
			err := blockchain.config.DataBase.DeleteIncomingCrossShard(block.Header.ShardID, crossShard, crossBlk)
			if err != nil {
				return NewBlockChainError(DeleteIncomingCrossShardError, err)
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) GetActiveShardNumber() int {
	return blockchain.BestState.Beacon.ActiveShards
}

// func (blockchain *BlockChain) BackupCurrentShardState(block *ShardBlock, beaconblks []*BeaconBlock) error {

// 	//Steps:
// 	// 1. Backup beststate
// 	// 2.	Backup data that will be modify by new block data

// 	tempMarshal, err := json.Marshal(blockchain.BestState.Shard[block.Header.ShardID])
// 	if err != nil {
// 		return NewBlockChainError(UnmashallJsonShardBlockError, err)
// 	}

// 	if err := blockchain.config.DataBase.StorePrevBestState(tempMarshal, false, block.Header.ShardID); err != nil {
// 		return NewBlockChainError(UnExpectedError, err)
// 	}

// 	if err := blockchain.createBackupFromTxViewPoint(block); err != nil {
// 		return err
// 	}

// 	if err := blockchain.createBackupFromCrossTxViewPoint(block); err != nil {
// 		return err
// 	}

// 	if err := blockchain.backupDatabaseFromBeaconInstruction(beaconblks, block.Header.ShardID); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (blockchain *BlockChain) backupDatabaseFromBeaconInstruction(beaconBlocks []*BeaconBlock,
// 	shardID byte) error {

// 	shardCommittee := make(map[byte][]string)
// 	isInit := false
// 	epoch := uint64(0)
// 	db := blockchain.config.DataBase
// 	// listShardCommittee := blockchain.config.DataBase.FetchCommitteeByEpoch
// 	for _, beaconBlock := range beaconBlocks {
// 		for _, l := range beaconBlock.Body.Instructions {
// 			if l[0] == StakeAction || l[0] == RandomAction {
// 				continue
// 			}
// 			if len(l) <= 2 {
// 				continue
// 			}
// 			shardToProcess, err := strconv.Atoi(l[1])
// 			if err != nil {
// 				continue
// 			}
// 			if shardToProcess == int(shardID) {
// 				metaType, err := strconv.Atoi(l[0])
// 				if err != nil {
// 					return err
// 				}
// 				switch metaType {
// 				case metadata.BeaconRewardRequestMeta:
// 					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
// 					if err != nil {
// 						return err
// 					}
// 					publicKeyCommittee, _, err := base58.Base58Check{}.Decode(beaconBlkRewardInfo.PayToPublicKey)
// 					if err != nil {
// 						return err
// 					}
// 					for key := range beaconBlkRewardInfo.BeaconReward {
// 						err = db.BackupCommitteeReward(publicKeyCommittee, key)
// 						if err != nil {
// 							return err
// 						}
// 					}
// 					continue

// 				case metadata.DevRewardRequestMeta:
// 					devRewardInfo, err := metadata.NewDevRewardInfoFromStr(l[3])
// 					if err != nil {
// 						return err
// 					}
// 					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(common.IncognitoDAOAddress)
// 					if err != nil {
// 						return err
// 					}
// 					for key := range devRewardInfo.DevReward {
// 						err = db.BackupCommitteeReward(keyWalletDevAccount.KeySet.PaymentAddress.Pk, key)
// 						if err != nil {
// 							return err
// 						}
// 					}
// 					continue

// 				case metadata.ShardBlockRewardRequestMeta:
// 					shardRewardInfo, err := metadata.NewShardBlockRewardInfoFromString(l[3])
// 					if err != nil {
// 						return err
// 					}
// 					if (!isInit) || (epoch != shardRewardInfo.Epoch) {
// 						isInit = true
// 						epoch = shardRewardInfo.Epoch
// 						temp, err := blockchain.config.DataBase.FetchShardCommitteeByHeight(epoch * blockchain.config.ChainParams.Epoch)
// 						if err != nil {
// 							return err
// 						}
// 						json.Unmarshal(temp, &shardCommittee)
// 					}
// 					err = blockchain.backupShareRewardForShardCommittee(shardRewardInfo.Epoch, shardRewardInfo.ShardReward, shardCommittee[shardID])
// 					if err != nil {
// 						return err
// 					}
// 					continue
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// func (blockchain *BlockChain) backupShareRewardForShardCommittee(epoch uint64, totalReward map[common.Hash]uint64, listCommitee []string) error {
// 	// reward := totalReward / uint64(len(listCommitee))
// 	reward := map[common.Hash]uint64{}
// 	for key, value := range totalReward {
// 		reward[key] = value / uint64(len(listCommitee))
// 	}
// 	for key := range totalReward {
// 		for _, committee := range listCommitee {
// 			committeeBytes, _, err := base58.Base58Check{}.Decode(committee)
// 			if err != nil {
// 				return err
// 			}
// 			err = blockchain.config.DataBase.BackupCommitteeReward(committeeBytes, key)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func (blockchain *BlockChain) createBackupFromTxViewPoint(block *ShardBlock) error {
// 	// Fetch data from block into tx View point
// 	view := NewTxViewPoint(block.Header.ShardID)
// 	err := view.fetchTxViewPointFromBlock(blockchain.config.DataBase, block)
// 	if err != nil {
// 		return err
// 	}

// 	// check privacy custom token
// 	backupedView := make(map[string]bool)
// 	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
// 		if ok := backupedView[privacyCustomTokenSubView.tokenID.String()]; !ok {
// 			err = blockchain.backupSerialNumbersFromTxViewPoint(*privacyCustomTokenSubView)
// 			if err != nil {
// 				return err
// 			}

// 			err = blockchain.backupCommitmentsFromTxViewPoint(*privacyCustomTokenSubView)
// 			if err != nil {
// 				return err
// 			}
// 			backupedView[privacyCustomTokenSubView.tokenID.String()] = true
// 		}

// 	}
// 	err = blockchain.backupSerialNumbersFromTxViewPoint(*view)
// 	if err != nil {
// 		return err
// 	}

// 	err = blockchain.backupCommitmentsFromTxViewPoint(*view)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (blockchain *BlockChain) createBackupFromCrossTxViewPoint(block *ShardBlock) error {
// 	view := NewTxViewPoint(block.Header.ShardID)
// 	err := view.fetchCrossTransactionViewPointFromBlock(blockchain.config.DataBase, block)

// 	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
// 		err = blockchain.backupCommitmentsFromTxViewPoint(*privacyCustomTokenSubView)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	err = blockchain.backupCommitmentsFromTxViewPoint(*view)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (blockchain *BlockChain) backupSerialNumbersFromTxViewPoint(view TxViewPoint) error {
// 	err := blockchain.config.DataBase.BackupSerialNumbersLen(*view.tokenID, view.shardID)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (blockchain *BlockChain) backupCommitmentsFromTxViewPoint(view TxViewPoint) error {

// 	// commitment
// 	keys := make([]string, 0, len(view.mapCommitments))
// 	for k := range view.mapCommitments {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)

// 	for _, k := range keys {
// 		pubkey := k
// 		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
// 		if err != nil {
// 			return err
// 		}
// 		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
// 		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
// 		if pubkeyShardID == view.shardID {
// 			err = blockchain.config.DataBase.BackupCommitmentsOfPubkey(*view.tokenID, view.shardID, pubkeyBytes)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	// outputs
// 	keys = make([]string, 0, len(view.mapOutputCoins))
// 	for k := range view.mapOutputCoins {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)

// 	// for _, k := range keys {
// 	// 	pubkey := k

// 	// 	pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	lastByte := pubkeyBytes[len(pubkeyBytes)-1]
// 	// 	pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
// 	// 	if pubkeyShardID == view.shardID {
// 	// 		err = blockchain.config.DataBase.BackupOutputCoin(*view.tokenID, pubkeyBytes, pubkeyShardID)
// 	// 		if err != nil {
// 	// 			return err
// 	// 		}
// 	// 	}
// 	// }
// 	return nil
// }

// func (blockchain *BlockChain) restoreDatabaseFromBeaconInstruction(beaconBlocks []*BeaconBlock,
// 	shardID byte) error {

// 	shardCommittee := make(map[byte][]string)
// 	isInit := false
// 	epoch := uint64(0)
// 	db := blockchain.config.DataBase
// 	// listShardCommittee := blockchain.config.DataBase.FetchCommitteeByEpoch
// 	for _, beaconBlock := range beaconBlocks {
// 		for _, l := range beaconBlock.Body.Instructions {
// 			if l[0] == StakeAction || l[0] == RandomAction {
// 				continue
// 			}
// 			if len(l) <= 2 {
// 				continue
// 			}
// 			shardToProcess, err := strconv.Atoi(l[1])
// 			if err != nil {
// 				continue
// 			}
// 			if shardToProcess == int(shardID) {
// 				metaType, err := strconv.Atoi(l[0])
// 				if err != nil {
// 					return err
// 				}
// 				switch metaType {
// 				case metadata.BeaconRewardRequestMeta:
// 					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
// 					if err != nil {
// 						return err
// 					}
// 					publicKeyCommittee, _, err := base58.Base58Check{}.Decode(beaconBlkRewardInfo.PayToPublicKey)
// 					if err != nil {
// 						return err
// 					}
// 					for key := range beaconBlkRewardInfo.BeaconReward {
// 						err = db.RestoreCommitteeReward(publicKeyCommittee, key)
// 						if err != nil {
// 							return err
// 						}
// 					}
// 					continue

// 				case metadata.DevRewardRequestMeta:
// 					devRewardInfo, err := metadata.NewDevRewardInfoFromStr(l[3])
// 					if err != nil {
// 						return err
// 					}
// 					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(common.IncognitoDAOAddress)
// 					if err != nil {
// 						return err
// 					}
// 					for key := range devRewardInfo.DevReward {
// 						err = db.RestoreCommitteeReward(keyWalletDevAccount.KeySet.PaymentAddress.Pk, key)
// 						if err != nil {
// 							return err
// 						}
// 					}
// 					continue

// 				case metadata.ShardBlockRewardRequestMeta:
// 					shardRewardInfo, err := metadata.NewShardBlockRewardInfoFromString(l[3])
// 					if err != nil {
// 						return err
// 					}
// 					if (!isInit) || (epoch != shardRewardInfo.Epoch) {
// 						isInit = true
// 						epoch = shardRewardInfo.Epoch
// 						temp, err := blockchain.config.DataBase.FetchShardCommitteeByHeight(epoch * blockchain.config.ChainParams.Epoch)
// 						if err != nil {
// 							return err
// 						}
// 						json.Unmarshal(temp, &shardCommittee)
// 					}
// 					err = blockchain.restoreShareRewardForShardCommittee(shardRewardInfo.Epoch, shardRewardInfo.ShardReward, shardCommittee[shardID])
// 					if err != nil {
// 						return err
// 					}
// 					continue
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// func (blockchain *BlockChain) restoreShareRewardForShardCommittee(epoch uint64, totalReward map[common.Hash]uint64, listCommitee []string) error {
// 	// reward := totalReward / uint64(len(listCommitee))
// 	reward := map[common.Hash]uint64{}
// 	for key, value := range totalReward {
// 		reward[key] = value / uint64(len(listCommitee))
// 	}
// 	for key := range totalReward {
// 		for _, committee := range listCommitee {
// 			committeeBytes, _, err := base58.Base58Check{}.Decode(committee)
// 			if err != nil {
// 				return err
// 			}
// 			err = blockchain.config.DataBase.RestoreCommitteeReward(committeeBytes, key)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func (blockchain *BlockChain) restoreFromTxViewPoint(block *ShardBlock) error {
// 	// Fetch data from block into tx View point
// 	view := NewTxViewPoint(block.Header.ShardID)
// 	err := view.fetchTxViewPointFromBlock(blockchain.config.DataBase, block)
// 	if err != nil {
// 		return err
// 	}

// 	// check normal custom token
// 	for indexTx, customTokenTx := range view.customTokenTxs {
// 		switch customTokenTx.TxTokenData.Type {
// 		case transaction.CustomTokenInit:
// 			{
// 				err = blockchain.config.DataBase.DeleteNormalToken(customTokenTx.TxTokenData.PropertyID)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case transaction.CustomTokenCrossShard:
// 			{
// 				err = blockchain.config.DataBase.DeleteNormalToken(customTokenTx.TxTokenData.PropertyID)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 		err = blockchain.config.DataBase.DeleteNormalTokenTx(customTokenTx.TxTokenData.PropertyID, indexTx, block.Header.ShardID, block.Header.Height)
// 		if err != nil {
// 			return err
// 		}

// 	}

// 	// check privacy custom token
// 	for indexTx, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
// 		privacyCustomTokenTx := view.privacyCustomTokenTxs[indexTx]
// 		switch privacyCustomTokenTx.TxPrivacyTokenData.Type {
// 		case transaction.CustomTokenInit:
// 			{
// 				err = blockchain.config.DataBase.DeletePrivacyToken(privacyCustomTokenTx.TxPrivacyTokenData.PropertyID)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 		err = blockchain.config.DataBase.DeletePrivacyTokenTx(privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, indexTx, block.Header.ShardID, block.Header.Height)
// 		if err != nil {
// 			return err
// 		}

// 		err = blockchain.restoreSerialNumbersFromTxViewPoint(*privacyCustomTokenSubView)
// 		if err != nil {
// 			return err
// 		}

// 		err = blockchain.restoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	err = blockchain.restoreSerialNumbersFromTxViewPoint(*view)
// 	if err != nil {
// 		return err
// 	}

// 	err = blockchain.restoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (blockchain *BlockChain) restoreFromCrossTxViewPoint(block *ShardBlock) error {
// 	view := NewTxViewPoint(block.Header.ShardID)
// 	err := view.fetchCrossTransactionViewPointFromBlock(blockchain.config.DataBase, block)

// 	for _, privacyCustomTokenSubView := range view.privacyCustomTokenViewPoint {
// 		tokenID := privacyCustomTokenSubView.tokenID
// 		if err := blockchain.config.DataBase.DeletePrivacyTokenCrossShard(*tokenID); err != nil {
// 			return err
// 		}
// 		err = blockchain.restoreCommitmentsFromTxViewPoint(*privacyCustomTokenSubView, block.Header.ShardID)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	err = blockchain.restoreCommitmentsFromTxViewPoint(*view, block.Header.ShardID)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (blockchain *BlockChain) restoreSerialNumbersFromTxViewPoint(view TxViewPoint) error {
// 	err := blockchain.config.DataBase.RestoreSerialNumber(*view.tokenID, view.shardID, view.listSerialNumbers)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (blockchain *BlockChain) restoreCommitmentsFromTxViewPoint(view TxViewPoint, shardID byte) error {

// 	// commitment
// 	keys := make([]string, 0, len(view.mapCommitments))
// 	for k := range view.mapCommitments {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)

// 	for _, k := range keys {
// 		pubkey := k
// 		item1 := view.mapCommitments[k]
// 		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
// 		if err != nil {
// 			return err
// 		}
// 		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
// 		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
// 		if pubkeyShardID == view.shardID {
// 			err = blockchain.config.DataBase.RestoreCommitmentsOfPubkey(*view.tokenID, view.shardID, pubkeyBytes, item1)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	// outputs
// 	for _, k := range keys {
// 		publicKey := k
// 		publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
// 		if err != nil {
// 			return err
// 		}
// 		lastByte := publicKeyBytes[len(publicKeyBytes)-1]
// 		publicKeyShardID := common.GetShardIDFromLastByte(lastByte)
// 		if publicKeyShardID == shardID {
// 			outputCoinArray := view.mapOutputCoins[k]
// 			outputCoinBytesArray := make([][]byte, 0)
// 			for _, outputCoin := range outputCoinArray {
// 				outputCoinBytesArray = append(outputCoinBytesArray, outputCoin.Bytes())
// 			}
// 			err = blockchain.config.DataBase.DeleteOutputCoin(*view.tokenID, publicKeyBytes, outputCoinBytesArray, publicKeyShardID)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
