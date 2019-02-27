package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"

	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/database/lvdb"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/metadata/frombeaconins"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	cache "github.com/patrickmn/go-cache"
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
	chainLock sync.RWMutex

	//=====cache
	beaconBlock        map[string][]byte // TODO review not use
	highestBeaconBlock string            // TODO review not use

	//channel
	cQuitSync  chan struct{}
	syncStatus struct {
		Beacon bool
		Shards map[byte]struct{}
		sync.Mutex

		CurrentlySyncShardBlkByHash           map[byte]*cache.Cache
		CurrentlySyncShardBlkByHeight         map[byte]*cache.Cache
		CurrentlySyncBeaconBlkByHash          *cache.Cache
		CurrentlySyncBeaconBlkByHeight        *cache.Cache
		CurrentlySyncShardToBeaconBlkByHash   map[byte]*cache.Cache
		CurrentlySyncShardToBeaconBlkByHeight map[byte]*cache.Cache
		CurrentlySyncCrossShardBlkByHash      map[byte]*cache.Cache

		PeersState     map[libp2p.ID]*peerState
		PeersStateLock sync.Mutex
	}
	knownChainState struct {
		Shards map[byte]ChainState
		Beacon ChainState
	}
	PeerStateCh chan *peerState
}
type BestState struct {
	Beacon *BestStateBeacon
	Shard  map[byte]*BestStateShard
}

// config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	// dataBase defines the database which houses the blocks and will be used to
	// store all metadata created by this package.
	//
	// This field is required.
	DataBase database.DatabaseInterface

	// shardBlock *lru.Cache
	// shardBody  *lru.Cache
	//======
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
	RelayShards []byte
	NodeMode    string
	//Light mode flag
	// Light bool
	//Wallet for light mode
	Wallet *wallet.Wallet

	//snapshot reward
	customTokenRewardSnapshot map[string]uint64

	ShardToBeaconPool ShardToBeaconPool
	CrossShardPool    map[byte]CrossShardPool
	NodeBeaconPool    NodeBeaconPool
	NodeShardPool     NodeShardPool
	TxPool            TxPool

	Server interface {
		BoardcastNodeState() error

		PushMessageGetBlockBeaconByHeight(from uint64, to uint64, peerID libp2p.ID) error
		PushMessageGetBlockBeaconByHash(blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error

		PushMessageGetBlockShardByHeight(shardID byte, from uint64, to uint64, peerID libp2p.ID) error
		PushMessageGetBlockShardByHash(shardID byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error

		PushMessageGetBlockShardToBeaconByHeight(shardID byte, from uint64, to uint64, peerID libp2p.ID) error
		PushMessageGetBlockShardToBeaconByHash(shardID byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error

		PushMessageGetBlockCrossShardByHash(fromShard byte, toShard byte, blksHash []common.Hash, getFromPool bool, peerID libp2p.ID) error
	}
	UserKeySet *cashec.KeySet
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

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := blockchain.initChainState(); err != nil {
		return err
	}

	// for chainIndex, bestState := range blockchain.BestState {
	// 	Logger.log.Infof("BlockChain state for chain #%d (Height %d, Best block hash %+v, Total tx %d, Salary fund %d, Gov Param %+v)",
	// 		chainIndex, bestState.Height, bestState.BestBlockHash.String(), bestState.TotalTxns, bestState.BestBlock.Header.SalaryFund, bestState.BestBlock.Header.GOVConstitution)
	// }
	blockchain.cQuitSync = make(chan struct{})
	blockchain.syncStatus.Shards = make(map[byte]struct{})
	blockchain.syncStatus.PeersState = make(map[libp2p.ID]*peerState)
	blockchain.knownChainState.Shards = make(map[byte]ChainState)
	return nil
}

func (blockchain *BlockChain) AddTxPool(txpool TxPool) {
	blockchain.config.TxPool = txpool
}

func (blockchain *BlockChain) InitShardToBeaconPool(db database.DatabaseInterface) {
	beaconBestState := BestStateBeacon{}
	temp, err := db.FetchBeaconBestState()
	if err != nil {
		panic("Fail to get state from db")
	} else {
		if err := json.Unmarshal(temp, &beaconBestState); err != nil {
			Logger.log.Error(err)
			panic("Can't Unmarshal beacon beststate")
		}
		blockchain.config.ShardToBeaconPool.SetShardState(beaconBestState.BestShardHeight)
	}

}

// -------------- Blockchain retriever's implementation --------------
// GetCustomTokenTxsHash - return list of tx which relate to custom token
func (blockchain *BlockChain) GetCustomTokenTxs(tokenID *common.Hash) (map[common.Hash]metadata.Transaction, error) {
	txHashesInByte, err := blockchain.config.DataBase.CustomTokenTxs(tokenID)
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

// GetOracleParams returns oracle params
func (blockchain *BlockChain) GetOracleParams() *params.Oracle {
	return &params.Oracle{}
	// return blockchain.BestState[0].BestBlock.Header.Oracle
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

	//TODO: 0xBahamoot check back later
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
		bestStateBytes, err := blockchain.config.DataBase.FetchBestState(shardID)
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

	// ---- test RPC api data --- remove after
	initTxs := []string{`{"Version":1,"Type":"s","LockTime":1549889112,"Fee":0,"Info":null,"SigPubKey":"A7GGbCnosNljq25A5o4VIGs7r6WOcs3OrDBJUFd28eEA","Sig":"4gzqBc1TnROMjEdGW1DdIlLRA6pAwbcC3r1macAVy8OaOQaWxcSQXubEgm3oKcJAyE7OnEckV35pwAWD4vr7+A==","Proof":"11111116WGHqpGSLR21nkwRaRVR2vJBD6DR8wKQfB5VCC4TNEXz1XeskmWDehJbmDvr4EeC8x5vGFSrNq4KRs4GoDgn85t7CHJPQWu6s8QWhQVRd621qqT5mBofPcB9WGgQPsD7i4WPxoPKVYhS3jaRXbT2C9S1tHQbW9TytbZKbASDgKygqeijEoWsLW4RXct1oGn2wat2Q1kdPX35AKW1B2R","PubKeyLastByteSender":0,"Metadata":null}`, `{"Version":1,"Type":"s","LockTime":1549889112,"Fee":0,"Info":null,"SigPubKey":"AySFA7ksPnDE7zG+ZKwyk8SaadPLOfJuIn5k4kqUgKcA","Sig":"0jcALduldAkey/6EmKW3EyUQGpJCZ5Vr1lmc7QlzOL3FYEHVwF3kXcDkuPXqqjaH8ueJjDGDqx4N8KpWDfSi7Q==","Proof":"11111116WGHqpGNRGpV3VBz1rndCx6TP4A8eLYeocjg8izynA2YAkx7x38mCir9Nm3oCubXdn25F4sj4jHryBtSbdwJj6o4X43YDftZ9nPsrw4m8DyF6NkxNXbvGj9egkUtypup34hdCXv2L8j5tB9cVUCXVqWeC9axqLLoibXEay4fLrroeRnfNhJ1moNDoQqyRVLrcC7yUjDQz6AUsdd3uFB","PubKeyLastByteSender":0,"Metadata":null}`}

	for _, tx := range initTxs {
		testSalaryTX := transaction.Tx{}
		testSalaryTX.UnmarshalJSON([]byte(tx))
		initBlock.Body.Transactions = append(initBlock.Body.Transactions, &testSalaryTX)
	}
	// var initTxs []string
	// testUserkeyList := []string{
	// 	"112t8rqnMrtPkJ4YWzXfG82pd9vCe2jvWGxqwniPM5y4hnimki6LcVNfXxN911ViJS8arTozjH4rTpfaGo5i1KKcG1ayjiMsa4E3nABGAqQh",
	// 	"112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV",
	// }
	// for _, val := range testUserkeyList {

	// 	testUserKey, _ := wallet.Base58CheckDeserialize(val)
	// 	testUserKey.KeySet.ImportFromPrivateKey(&testUserKey.KeySet.PrivateKey)

	// 	testSalaryTX := transaction.Tx{}
	// 	testSalaryTX.InitTxSalary(1000000, &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey,
	// 		blockchain.config.DataBase,
	// 		nil,
	// 	)
	// 	initTx, _ := json.Marshal(testSalaryTX)
	// 	initTxs = append(initTxs, string(initTx))
	// 	initBlock.Body.Transactions = append(initBlock.Body.Transactions, &testSalaryTX)
	// }
	// fmt.Println(initTxs)
	// os.Exit(1)

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
	blockchain.BestState.Beacon.Update(initBlock)
	// TODO(@0xankylosaurus): initialize oracle data properly
	// All values are in millicent
	blockchain.BestState.Beacon.StabilityInfo.Oracle.DCBToken = 1000000 // $10
	blockchain.BestState.Beacon.StabilityInfo.Oracle.GOVToken = 2000000 // $20
	blockchain.BestState.Beacon.StabilityInfo.Oracle.Constant = 1000    // $1, for Nano
	blockchain.BestState.Beacon.StabilityInfo.Oracle.ETH = 10000        // $100, for milliether
	blockchain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams.RaiseReserveData = map[common.Hash]*params.RaiseReserveData{
		common.ETHAssetID: &params.RaiseReserveData{
			EndBlock: 1000,
			Amount:   1000,
		},
	}
	blockchain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams.SpendReserveData = map[common.Hash]*params.SpendReserveData{
		common.ETHAssetID: &params.SpendReserveData{
			EndBlock:        1000,
			ReserveMinPrice: 1000,
			Amount:          10000000,
		},
	}

	bondID, _ := common.NewHashFromStr("4c420b974449ac188c155a7029706b8419a591ee398977d00000000000000000")
	buyBondSaleID := [32]byte{1}
	sellBondSaleID := [32]byte{2}
	saleData := []params.SaleData{
		params.SaleData{
			SaleID:           buyBondSaleID[:],
			EndBlock:         1000,
			BuyingAsset:      *bondID,
			BuyingAmount:     100,
			DefaultBuyPrice:  100,
			SellingAsset:     common.ConstantID,
			SellingAmount:    150,
			DefaultSellPrice: 100,
		},
		params.SaleData{
			SaleID:           sellBondSaleID[:],
			EndBlock:         2000,
			BuyingAsset:      common.ConstantID,
			BuyingAmount:     250,
			DefaultBuyPrice:  100,
			SellingAsset:     *bondID,
			SellingAmount:    200,
			DefaultSellPrice: 100,
		},
	}
	blockchain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams.ListSaleData = saleData
	// Store temp crowdsale states to avoid submiting DCB proposal
	for _, data := range saleData {
		key := getSaleDataKeyBeacon(data.SaleID)
		if _, ok := blockchain.BestState.Beacon.Params[key]; ok {
			continue
		}
		value := getSaleDataValueBeacon(&data)
		blockchain.BestState.Beacon.Params[key] = value
	}

	loanParams := []params.LoanParams{
		params.LoanParams{
			InterestRate:     100,   // 1%
			Maturity:         1000,  // 1 month in blocks
			LiquidationStart: 15000, // 150%
		},
	}
	blockchain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams.ListLoanParams = loanParams

	blockchain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress = []privacy.PaymentAddress{
		// Payment4: 1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba
		privacy.PaymentAddress{
			Pk: []byte{3, 36, 133, 3, 185, 44, 62, 112, 196, 239, 49, 190, 100, 172, 50, 147, 196, 154, 105, 211, 203, 57, 242, 110, 34, 126, 100, 226, 74, 148, 128, 167, 0},
			Tk: []byte{2, 134, 3, 114, 89, 60, 134, 3, 185, 245, 176, 187, 244, 145, 250, 149, 67, 98, 68, 106, 69, 200, 228, 209, 3, 26, 231, 15, 36, 251, 211, 186, 159},
		},
	}

	// Bond
	blockchain.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams.SellingBonds = &params.SellingBonds{
		BondName:       "Bond 1000 blocks",
		BondSymbol:     "BND1000",
		TotalIssue:     1000,
		BondsToSell:    1000,
		BondPrice:      100, // 1 mili constant
		Maturity:       3,
		BuyBackPrice:   120, // 1.2 mili constant
		StartSellingAt: 0,
		SellingWithin:  100000,
	}

	// Insert new block into beacon chain
	if err := blockchain.StoreBeaconBestState(); err != nil {
		Logger.log.Error("Error Store best state for block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}
	if err := blockchain.config.DataBase.StoreBeaconBlock(blockchain.BestState.Beacon.BestBlock); err != nil {
		Logger.log.Error("Error store beacon block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return err
	}
	blockHash := initBlock.Hash()
	if err := blockchain.config.DataBase.StoreBeaconBlockIndex(blockHash, initBlock.Header.Height); err != nil {
		return err
	}
	return nil
}

/*
Get block index(height) of block
*/
func (blockchain *BlockChain) GetBlockHeightByBlockHash(hash *common.Hash) (uint64, byte, error) {
	return blockchain.config.DataBase.GetIndexOfBlock(hash)
}

/*
Get block hash by block index(height)
*/
func (blockchain *BlockChain) GetBeaconBlockHashByHeight(height uint64) (*common.Hash, error) {
	return blockchain.config.DataBase.GetBeaconBlockHashByIndex(height)
}

/*
Fetch DatabaseInterface and get block by index(height) of block
*/
func (blockchain *BlockChain) GetBeaconBlockByHeight(height uint64) (*BeaconBlock, error) {
	hashBlock, err := blockchain.config.DataBase.GetBeaconBlockHashByIndex(height)
	if err != nil {
		return nil, err
	}
	block, err := blockchain.GetBeaconBlockByHash(hashBlock)
	if err != nil {
		return nil, err
	}
	return block, nil
}

/*
Fetch DatabaseInterface and get block data by block hash
*/
func (blockchain *BlockChain) GetBeaconBlockByHash(hash *common.Hash) (*BeaconBlock, error) {
	blockBytes, err := blockchain.config.DataBase.FetchBeaconBlock(hash)
	if err != nil {
		return nil, err
	}
	block := BeaconBlock{}
	err = json.Unmarshal(blockBytes, &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

/*
Get block index(height) of block
*/
func (blockchain *BlockChain) GetShardBlockHeightByHash(hash *common.Hash) (uint64, byte, error) {
	return blockchain.config.DataBase.GetIndexOfBlock(hash)
}

/*
Get block hash by block index(height)
*/
func (blockchain *BlockChain) GetShardBlockHashByHeight(height uint64, shardID byte) (*common.Hash, error) {
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
	block, err := blockchain.GetShardBlockByHash(hashBlock)

	return block, err
}

/*
Fetch DatabaseInterface and get block data by block hash
*/
func (blockchain *BlockChain) GetShardBlockByHash(hash *common.Hash) (*ShardBlock, error) {
	blockBytes, err := blockchain.config.DataBase.FetchBlock(hash)
	if err != nil {
		return nil, err
	}

	block := ShardBlock{}
	err = json.Unmarshal(blockBytes, &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
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
	return blockchain.config.DataBase.StoreBestState(blockchain.BestState.Shard[shardID], shardID)
}

/*
GetBestState - return a best state from a chain
*/
// #1 - shardID - index of chain
func (blockchain *BlockChain) GetShardBestState(shardID byte) (*BestStateShard, error) {
	bestState := BestStateShard{}
	bestStateBytes, err := blockchain.config.DataBase.FetchBestState(shardID)
	if err == nil {
		err = json.Unmarshal(bestStateBytes, &bestState)
	}
	return &bestState, err
}

/*
Store block into Database
*/
func (blockchain *BlockChain) StoreShardBlock(block *ShardBlock) error {
	return blockchain.config.DataBase.StoreShardBlock(block, block.Header.ShardID)
}

/*
	Store Only Block Header into database
*/
func (blockchain *BlockChain) StoreShardBlockHeader(block *ShardBlock) error {
	//Logger.log.Infof("Store Block Header, block header %+v, block hash %+v, chain id %+v",block.Header, block.blockHash, block.Header.shardID)
	return blockchain.config.DataBase.StoreShardBlockHeader(block.Header, block.Hash(), block.Header.ShardID)
}

/*
	Store Transaction in Light mode
*/
// func (blockchain *BlockChain) StoreUnspentTransactionLightMode(privatKey *privacy.SpendingKey, shardID byte, blockHeight int32, txIndex int, tx *transaction.Tx) error {
// 	txJsonBytes, err := json.Marshal(tx)
// 	if err != nil {
// 		return NewBlockChainError(UnExpectedError, errors.New("json.Marshal"))
// 	}
// 	return blockchain.config.DataBase.StoreTransactionLightMode(privatKey, shardID, blockHeight, txIndex, *(tx.Hash()), txJsonBytes)
// }

/*
Save index(height) of block by block hash
and
Save block hash by index(height) of block
*/
func (blockchain *BlockChain) StoreShardBlockIndex(block *ShardBlock) error {
	return blockchain.config.DataBase.StoreShardBlockIndex(block.Hash(), block.Header.Height, block.Header.ShardID)
}

func (blockchain *BlockChain) StoreTransactionIndex(txHash *common.Hash, blockHash *common.Hash, index int) error {
	return blockchain.config.DataBase.StoreTransactionIndex(txHash, blockHash, index)
}

/*
Uses an existing database to update the set of used tx by saving list nullifier of privacy,
this is a list tx-out which are used by a new tx
*/
func (blockchain *BlockChain) StoreSerialNumbersFromTxViewPoint(view TxViewPoint) error {
	for _, item1 := range view.listSerialNumbers {
		err := blockchain.config.DataBase.StoreSerialNumbers(view.tokenID, item1, view.shardID)
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
func (blockchain *BlockChain) StoreSNDerivatorsFromTxViewPoint(view TxViewPoint, shardID byte) error {
	// commitment
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		pubkey := k
		item1 := view.mapSnD[k]
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
		if pubkeyShardID == shardID {
			for _, snd := range item1 {
				err = blockchain.config.DataBase.StoreSNDerivators(view.tokenID, snd, view.shardID)
				if err != nil {
					return err
				}
			}
		}
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

	// commitment
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		pubkey := k
		item1 := view.mapCommitments[k]
		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
		if pubkeyShardID == shardID {
			for _, com := range item1 {
				err = blockchain.config.DataBase.StoreCommitments(view.tokenID, pubkeyBytes, com, view.shardID)
				if err != nil {
					return err
				}
			}
		}
	}

	// outputs
	keys = make([]string, 0, len(view.mapOutputCoins))
	for k := range view.mapOutputCoins {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		pubkey := k
		item1 := view.mapOutputCoins[k]

		pubkeyBytes, _, err := base58.Base58Check{}.Decode(pubkey)
		if err != nil {
			return err
		}
		lastByte := pubkeyBytes[len(pubkeyBytes)-1]
		pubkeyShardID := common.GetShardIDFromLastByte(lastByte)
		if pubkeyShardID == shardID {
			for _, outcoin := range item1 {
				err = blockchain.config.DataBase.StoreOutputCoins(view.tokenID, pubkeyBytes, outcoin.Bytes(), pubkeyShardID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// func (self *BlockChain) GetChainBlocks(shardID byte) ([]*Block, error) {
// 	result := make([]*Block, 0)
// 	data, err := self.config.DataBase.FetchChainBlocks(shardID)
// 	if err != nil {
// 		return nil, err
// 	}

// CreateAndSaveTxViewPointFromBlock - fetch data from block, put into txviewpoint variable and save into db
// need to check light or not light mode
// with light mode - node only fetch outputcoins of account in local wallet -> smaller data
// with not light mode - node fetch all outputcoins of all accounts in network -> big data
// @note: still storage full data of commitments, serialnumbersm snderivator to check double spend
// @note: this function only work for transaction transfer token/constant within shard

func (blockchain *BlockChain) CreateAndSaveTxViewPointFromBlock(block *ShardBlock) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	// TODO: 0xsirrush check lightmode turn off
	err := view.fetchTxViewPointFromBlock(blockchain.config.DataBase, block, nil)
	if err != nil {
		return err
	}

	// check normal custom token
	for indexTx, customTokenTx := range view.customTokenTxs {
		switch customTokenTx.TxTokenData.Type {
		case transaction.CustomTokenInit:
			{
				Logger.log.Info("Store custom token when it is issued", customTokenTx.TxTokenData.PropertyID, customTokenTx.TxTokenData.PropertySymbol, customTokenTx.TxTokenData.PropertyName)
				err = blockchain.config.DataBase.StoreCustomToken(&customTokenTx.TxTokenData.PropertyID, customTokenTx.Hash()[:])
				if err != nil {
					return err
				}
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Info("Transfer custom token %+v", customTokenTx)
			}
		}
		// save tx which relate to custom token
		// Reject Double spend UTXO before enter this state
		err = blockchain.StoreCustomTokenPaymentAddresstHistory(customTokenTx)
		if err != nil {
			// Skip double spend
			return err
		}
		err = blockchain.config.DataBase.StoreCustomTokenTx(&customTokenTx.TxTokenData.PropertyID, block.Header.ShardID, block.Header.Height, indexTx, customTokenTx.Hash()[:])
		if err != nil {
			return err
		}

		// replace 1000 with proper value for snapshot
		if block.Header.Height%1000 == 0 {
			// list of unreward-utxo
			blockchain.config.customTokenRewardSnapshot, err = blockchain.config.DataBase.GetCustomTokenPaymentAddressesBalance(&customTokenTx.TxTokenData.PropertyID)
			if err != nil {
				return err
			}
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
				err = blockchain.config.DataBase.StorePrivacyCustomToken(&privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, privacyCustomTokenTx.Hash()[:])
				if err != nil {
					return err
				}
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Info("Transfer custom token %+v", privacyCustomTokenTx)
			}
		}
		err = blockchain.config.DataBase.StorePrivacyCustomTokenTx(&privacyCustomTokenTx.TxTokenPrivacyData.PropertyID, block.Header.ShardID, block.Header.Height, indexTx, privacyCustomTokenTx.Hash()[:])
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

	// Update the list nullifiers and commitment, snd set using the state of the used tx view point. This
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

	return nil
}

//TODO: @merman Store txcustom token
func (blockchain *BlockChain) CreateAndSaveCrossOutputCoinViewPointFromBlock(block *ShardBlock) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	// TODO: 0xsirrush check lightmode turn off
	err := view.fetchCrossOutputViewPointFromBlock(blockchain.config.DataBase, block, nil)
	// Update the list nullifiers and commitment, snd set using the state of the used tx view point. This
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
//   H: value-spent/unspent-rewarded/unreward
*/
func (blockchain *BlockChain) StoreCustomTokenPaymentAddresstHistory(customTokenTx *transaction.TxCustomToken) error {
	Splitter := lvdb.Splitter
	TokenPaymentAddressPrefix := lvdb.TokenPaymentAddressPrefix
	unspent := lvdb.Unspent
	spent := lvdb.Spent
	unreward := lvdb.Unreward

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
		paymentAddressKey = append(paymentAddressKey, byte(voutIndex))
		_, err := blockchain.config.DataBase.HasValue(paymentAddressKey)
		if err != nil {
			return err
		}
		value, err := blockchain.config.DataBase.Get(paymentAddressKey)
		if err != nil {
			return err
		}
		// old value: {value}-unspent-unreward/reward
		values := strings.Split(string(value), string(Splitter))
		if strings.Compare(values[1], string(unspent)) != 0 {
			return errors.New("Double Spend Detected")
		}
		// new value: {value}-spent-unreward/reward
		newValues := values[0] + string(Splitter) + string(spent) + string(Splitter) + values[2]
		if err := blockchain.config.DataBase.Put(paymentAddressKey, []byte(newValues)); err != nil {
			return err
		}
	}
	for index, vout := range customTokenTx.TxTokenData.Vouts {
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
		paymentAddressKey = append(paymentAddressKey, byte(voutIndex))
		ok, err := blockchain.config.DataBase.HasValue(paymentAddressKey)
		// Vout already exist
		if ok {
			return errors.New("UTXO already exist")
		}
		if err != nil {
			return err
		}
		// init value: {value}-unspent-unreward
		paymentAddressValue := strconv.Itoa(int(value)) + string(Splitter) + string(unspent) + string(Splitter) + string(unreward)
		if err := blockchain.config.DataBase.Put(paymentAddressKey, []byte(paymentAddressValue)); err != nil {
			return err
		}
	}
	return nil
}

// DecryptTxByKey - process outputcoin to get outputcoin data which relate to keyset
func (blockchain *BlockChain) DecryptOutputCoinByKey(outCoinTemp *privacy.OutputCoin, keySet *cashec.KeySet, shardID byte, tokenID *common.Hash) *privacy.OutputCoin {
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
			ok, err := blockchain.config.DataBase.HasSerialNumber(tokenID, result.CoinDetails.SerialNumber.Compress(), shardID)
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
With private-key, we can check unspent tx by check nullifiers from database
- Param #1: keyset - (priv-key, payment-address, readonlykey)
in case priv-key: return unspent outputcoin tx
in case readonly-key: return all outputcoin tx with amount value
in case payment-address: return all outputcoin tx with no amount value
- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
*/
func (blockchain *BlockChain) GetListOutputCoinsByKeyset(keyset *cashec.KeySet, shardID byte, tokenID *common.Hash) ([]*privacy.OutputCoin, error) {
	// lock chain
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()

	// if blockchain.config.Light {
	// 	// Get unspent tx with light mode
	// 	// TODO
	// }
	// get list outputcoin of pubkey from db

	outCointsInBytes, err := blockchain.config.DataBase.GetOutcoinsByPubkey(tokenID, keyset.PaymentAddress.Pk[:], shardID)
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
		pubkeyCompress := out.CoinDetails.PublicKey.Compress()
		if bytes.Equal(pubkeyCompress, keyset.PaymentAddress.Pk[:]) {
			out = blockchain.DecryptOutputCoinByKey(out, keyset, shardID, tokenID)
			if out == nil {
				continue
			} else {
				results = append(results, out)
			}
		}
	}
	if err != nil {
		return nil, err
	}

	return results, nil
}

// func (blockchain *BlockChain) GetCommitteCandidate(pubkeyParam string) *CommitteeCandidateInfo {
// 	for _, bestState := range blockchain.BestState {
// 		for pubkey, candidateInfo := range bestState.Candidates {
// 			if pubkey == pubkeyParam {
// 				return &candidateInfo
// 			}
// 		}
// 	}
// 	return nil
// }

// /*
// Get Candidate List from all chain and merge all to one - return pubkey of them
// */
// func (blockchain *BlockChain) GetCommitteeCandidateList() []string {
// 	candidatePubkeyList := []string{}
// 	for _, bestState := range blockchain.BestState {
// 		for pubkey, _ := range bestState.Candidates {
// 			if common.IndexOfStr(pubkey, candidatePubkeyList) < 0 {
// 				candidatePubkeyList = append(candidatePubkeyList, pubkey)
// 			}
// 		}
// 	}
// 	sort.Slice(candidatePubkeyList, func(i, j int) bool {
// 		cndInfoi := blockchain.GetCommitteeCandidateInfo(candidatePubkeyList[i])
// 		cndInfoj := blockchain.GetCommitteeCandidateInfo(candidatePubkeyList[j])
// 		if cndInfoi.Value == cndInfoj.Value {
// 			if cndInfoi.Timestamp < cndInfoj.Timestamp {
// 				return true
// 			} else if cndInfoi.Timestamp > cndInfoj.Timestamp {
// 				return false
// 			} else {
// 				if cndInfoi.shardID <= cndInfoj.shardID {
// 					return true
// 				} else if cndInfoi.shardID < cndInfoj.shardID {
// 					return false
// 				}
// 			}
// 		} else if cndInfoi.Value > cndInfoj.Value {
// 			return true
// 		} else {
// 			return false
// 		}
// 		return false
// 	})
// 	return candidatePubkeyList
// }

// func (blockchain *BlockChain) GetCommitteeCandidateInfo(nodeAddr string) CommitteeCandidateInfo {
// 	var cndVal CommitteeCandidateInfo
// 	for _, bestState := range blockchain.BestState {
// 		cndValTmp, ok := bestState.Candidates[nodeAddr]
// 		if ok {
// 			cndVal.Value += cndValTmp.Value
// 			if cndValTmp.Timestamp > cndVal.Timestamp {
// 				cndVal.Timestamp = cndValTmp.Timestamp
// 				cndVal.shardID = cndValTmp.shardID
// 			}
// 		}
// 	}
// 	return cndVal
// }

// GetUnspentTxCustomTokenVout - return all unspent tx custom token out of sender
func (blockchain *BlockChain) GetUnspentTxCustomTokenVout(receiverKeyset cashec.KeySet, tokenID *common.Hash) ([]transaction.TxTokenVout, error) {
	data, err := blockchain.config.DataBase.GetCustomTokenPaymentAddressUTXO(tokenID, receiverKeyset.PaymentAddress.Bytes())
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
		// get unspent and unreward transaction output
		if strings.Compare(values[1], string(unspent)) == 0 {

			vout := transaction.TxTokenVout{}
			vout.PaymentAddress = receiverKeyset.PaymentAddress
			txHash, err := common.Hash{}.NewHashFromStr(string(keys[3]))
			if err != nil {
				return nil, err
			}
			vout.SetTxCustomTokenID(*txHash)
			voutIndexByte := []byte(keys[4])[0]
			voutIndex := int(voutIndexByte)
			vout.SetIndex(voutIndex)
			value, err := strconv.Atoi(values[0])
			if err != nil {
				return nil, err
			}
			vout.Value = uint64(value)
			fmt.Println("GetCustomTokenPaymentAddressUTXO VOUT", vout)
			voutList = append(voutList, vout)
		}
	}
	return voutList, nil
}

// GetTransactionByHash - retrieve tx from txId(txHash)
func (blockchain *BlockChain) GetTransactionByHash(txHash *common.Hash) (byte, *common.Hash, int, metadata.Transaction, error) {
	blockHash, index, err := blockchain.config.DataBase.GetTransactionIndexById(txHash)
	if err != nil {
		abc := NewBlockChainError(UnExpectedError, err)
		Logger.log.Error(abc)
		return byte(255), nil, -1, nil, abc
	}
	block, err1 := blockchain.GetShardBlockByHash(blockHash)
	if err1 != nil {
		Logger.log.Errorf("ERROR", err1, "NO Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Body.Transactions[index])
		return byte(255), nil, -1, nil, NewBlockChainError(UnExpectedError, err1)
	}
	//Logger.log.Infof("Transaction in block with hash &+v", blockHash, "and index", index, "contains", block.Transactions[index])
	return block.Header.ShardID, blockHash, index, block.Body.Transactions[index], nil
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
		_, blockHash, index, tx, err := blockchain.GetTransactionByHash(&hash)
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
func (blockchain *BlockChain) ListPrivacyCustomToken() (map[common.Hash]transaction.TxCustomTokenPrivacy, error) {
	data, err := blockchain.config.DataBase.ListPrivacyCustomToken()
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash]transaction.TxCustomTokenPrivacy)
	for _, txData := range data {
		hash := common.Hash{}
		hash.SetBytes(txData)
		_, blockHash, index, tx, err := blockchain.GetTransactionByHash(&hash)
		_ = blockHash
		_ = index
		if err != nil {
			return nil, err
		}
		txPrivacyCustomToken := tx.(*transaction.TxCustomTokenPrivacy)
		result[txPrivacyCustomToken.TxTokenPrivacyData.PropertyID] = *txPrivacyCustomToken
	}
	return result, nil
}

// GetCustomTokenTxsHash - return list hash of tx which relate to custom token
func (blockchain *BlockChain) GetCustomTokenTxsHash(tokenID *common.Hash) ([]common.Hash, error) {
	txHashesInByte, err := blockchain.config.DataBase.CustomTokenTxs(tokenID)
	if err != nil {
		return nil, err
	}
	result := []common.Hash{}
	for _, temp := range txHashesInByte {
		result = append(result, *temp)
	}
	return result, nil
}

// GetPrivacyCustomTokenTxsHash - return list hash of tx which relate to custom token
func (blockchain *BlockChain) GetPrivacyCustomTokenTxsHash(tokenID *common.Hash) ([]common.Hash, error) {
	txHashesInByte, err := blockchain.config.DataBase.PrivacyCustomTokenTxs(tokenID)
	if err != nil {
		return nil, err
	}
	result := []common.Hash{}
	for _, temp := range txHashesInByte {
		result = append(result, *temp)
	}
	return result, nil
}

// GetListTokenHolders - return list paymentaddress (in hexstring) of someone who hold custom token in network
func (blockchain *BlockChain) GetListTokenHolders(tokenID *common.Hash) (map[string]uint64, error) {
	result, err := blockchain.config.DataBase.GetCustomTokenPaymentAddressesBalance(tokenID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (blockchain *BlockChain) GetCustomTokenRewardSnapshot() map[string]uint64 {
	return blockchain.config.customTokenRewardSnapshot
}

func (blockchain *BlockChain) GetNumberOfDCBGovernors() int {
	return common.NumberOfDCBGovernors
}

func (blockchain *BlockChain) GetNumberOfGOVGovernors() int {
	return common.NumberOfGOVGovernors
}

// func (blockchain *BlockChain) GetBestBlock(shardID byte) *Block {
// 	return blockchain.BestState[shardID].BestBlock
// }

func (blockchain *BlockChain) GetConstitutionStartHeight(boardType metadata.BoardType, shardID byte) uint64 {
	if boardType == metadata.DCBBoard {
		return blockchain.GetDCBConstitutionStartHeight(shardID)
	} else {
		return blockchain.GetGOVConstitutionStartHeight(shardID)
	}
}

func (self *BlockChain) GetDCBConstitutionStartHeight(shardID byte) uint64 {
	return self.BestState.Beacon.StabilityInfo.DCBConstitution.StartedBlockHeight
}
func (self *BlockChain) GetGOVConstitutionStartHeight(shardID byte) uint64 {
	return self.BestState.Beacon.StabilityInfo.GOVConstitution.StartedBlockHeight
}

func (blockchain *BlockChain) GetConstitutionEndHeight(boardType metadata.BoardType, shardID byte) uint64 {
	if boardType == metadata.DCBBoard {
		return blockchain.GetDCBConstitutionEndHeight(shardID)
	} else {
		return blockchain.GetGOVConstitutionEndHeight(shardID)
	}
}

func (blockchain *BlockChain) GetBoardEndHeight(boardType metadata.BoardType, chainID byte) uint64 {
	if boardType == metadata.DCBBoard {
		return blockchain.GetDCBBoardEndHeight(chainID)
	} else {
		return blockchain.GetGOVBoardEndHeight(chainID)
	}
}

func (self *BlockChain) GetDCBConstitutionEndHeight(chainID byte) uint64 {
	return self.BestState.Beacon.StabilityInfo.DCBConstitution.GetEndedBlockHeight()
}

func (self *BlockChain) GetGOVConstitutionEndHeight(shardID byte) uint64 {
	return self.BestState.Beacon.StabilityInfo.GOVConstitution.GetEndedBlockHeight()
}

func (self *BlockChain) GetDCBBoardEndHeight(chainID byte) uint64 {
	return self.BestState.Beacon.StabilityInfo.DCBGovernor.EndBlock
}

func (self *BlockChain) GetGOVBoardEndHeight(chainID byte) uint64 {
	return self.BestState.Beacon.StabilityInfo.GOVGovernor.EndBlock
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

// GetFeePerKbTx - return fee (per kb of tx) from GOV params data
func (blockchain BlockChain) GetFeePerKbTx() uint64 {
	// TODO: stability
	// return blockchain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.FeePerKbTx
	return 0
}

func (blockchain *BlockChain) GetCurrentBoardIndex(helper ConstitutionHelper) uint32 {
	board := helper.GetBoard(blockchain)
	return board.GetBoardIndex()
}

func (blockchain *BlockChain) GetConstitutionIndex(helper ConstitutionHelper) uint32 {
	constitutionInfo := helper.GetConstitutionInfo(blockchain)
	return constitutionInfo.ConstitutionIndex
}

//1. Current National welfare (NW)  < lastNW * 0.9 (Emergency case)
//2. Block height == last constitution start time + last constitution window
//This function is called after successful connect block => block height is block height of best state
func (self *BlockChain) NeedToEnterEncryptionPhrase(helper ConstitutionHelper) bool {
	thisBlockHeight := self.BestState.Beacon.BestBlock.Header.Height
	newNationalWelfare := helper.GetCurrentNationalWelfare(self)
	oldNationalWelfare := helper.GetOldNationalWelfare(self)
	thresholdNationalWelfare := oldNationalWelfare * helper.GetThresholdRatioOfCrisis() / common.BasePercentage
	//
	constitutionInfo := helper.GetConstitutionInfo(self)
	endedOfConstitution := constitutionInfo.StartedBlockHeight + constitutionInfo.ExecuteDuration
	pivotOfStart := endedOfConstitution - 3*uint64(common.EncryptionOnePhraseDuration)
	//
	rightTime := newNationalWelfare < thresholdNationalWelfare || pivotOfStart == uint64(thisBlockHeight)

	encryptFlag, _ := self.config.DataBase.GetEncryptFlag(helper.GetBoardType().BoardTypeDB())
	rightFlag := encryptFlag == common.Lv3EncryptionFlag
	if rightTime && rightFlag {
		return true
	}
	return false
}

//This function is called after successful connect block => block height is block height of best state
func (self *BlockChain) NeedEnterEncryptLv1(helper ConstitutionHelper) bool {
	BestBlock := self.BestState.Beacon.BestBlock
	thisBlockHeight := BestBlock.Header.Height
	lastEncryptBlockHeight, _ := self.config.DataBase.GetEncryptionLastBlockHeight(helper.GetBoardType().BoardTypeDB())
	encryptFlag, _ := self.config.DataBase.GetEncryptFlag(helper.GetBoardType().BoardTypeDB())
	if thisBlockHeight == lastEncryptBlockHeight+common.EncryptionOnePhraseDuration &&
		encryptFlag == common.Lv2EncryptionFlag {
		return true
	}
	return false
}

//This function is called after successful connect block => block height is block height of best state
func (self *BlockChain) NeedEnterEncryptNormal(helper ConstitutionHelper) bool {
	BestBlock := self.BestState.Beacon.BestBlock
	thisBlockHeight := BestBlock.Header.Height
	lastEncryptBlockHeight, _ := self.config.DataBase.GetEncryptionLastBlockHeight(helper.GetBoardType().BoardTypeDB())
	encryptFlag, _ := self.config.DataBase.GetEncryptFlag(helper.GetBoardType().BoardTypeDB())
	if thisBlockHeight == lastEncryptBlockHeight+common.EncryptionOnePhraseDuration &&
		encryptFlag == common.Lv1EncryptionFlag {
		return true
	}
	return false
}

//This function is called after successful connect block => block height is block height of best state
func (self *BlockChain) CreateUpdateEncryptPhraseAndRewardConstitutionIns(helper ConstitutionHelper) ([]frombeaconins.InstructionFromBeacon, error) {
	instructions := make([]frombeaconins.InstructionFromBeacon, 0)
	flag := byte(0)
	boardType := helper.GetBoardType()
	if self.NeedToEnterEncryptionPhrase(helper) {
		flag = common.Lv2EncryptionFlag
	} else if self.NeedEnterEncryptLv1(helper) {
		flag = common.Lv1EncryptionFlag
	} else if self.NeedEnterEncryptNormal(helper) {
		flag = common.NormalEncryptionFlag
	} else if self.readyNewConstitution(helper) {
		flag = common.Lv3EncryptionFlag
		newIns, err := self.createAcceptConstitutionAndPunishTxAndRewardSubmitter(helper)
		instructions = append(instructions, newIns...)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		rewardIns, err := self.createRewardProposalWinnerIns(helper)
		instructions = append(instructions, rewardIns)
	}
	//create instruction to force shard to create transaction. Insert block in shard will affect db
	setEncryptionLastBlockIns := frombeaconins.NewSetEncryptionLastBlockIns(boardType, self.BestState.Beacon.BestBlock.Header.Height)
	instructions = append(instructions, setEncryptionLastBlockIns)
	setEncryptionFlagIns := frombeaconins.NewSetEncryptionFlagIns(boardType, flag)
	instructions = append(instructions, setEncryptionFlagIns)
	return instructions, nil
}

func (blockchain *BlockChain) SetEncryptPhrase(helper ConstitutionHelper) {
	flag := byte(0)
	boardType := helper.GetBoardType()
	height := blockchain.BestState.Beacon.BestBlock.Header.Height
	if blockchain.NeedToEnterEncryptionPhrase(helper) {
		flag = common.Lv2EncryptionFlag
		blockchain.config.DataBase.SetEncryptionLastBlockHeight(boardType.BoardTypeDB(), height)
		blockchain.config.DataBase.SetEncryptFlag(boardType.BoardTypeDB(), flag)
	} else if blockchain.NeedEnterEncryptLv1(helper) {
		flag = common.Lv1EncryptionFlag
		blockchain.config.DataBase.SetEncryptionLastBlockHeight(boardType.BoardTypeDB(), height)
		blockchain.config.DataBase.SetEncryptFlag(boardType.BoardTypeDB(), flag)
	} else if blockchain.NeedEnterEncryptNormal(helper) {
		flag = common.NormalEncryptionFlag
		blockchain.config.DataBase.SetEncryptionLastBlockHeight(boardType.BoardTypeDB(), height)
		blockchain.config.DataBase.SetEncryptFlag(boardType.BoardTypeDB(), flag)
	}
}

// GetRecentTransactions - find all recent history txs which are created by user
// by number of block, maximum is 100 newest blocks
func (blockchain *BlockChain) GetRecentTransactions(numBlock uint64, key *privacy.ViewingKey, shardID byte) (map[string]metadata.Transaction, error) {
	if numBlock > 100 { // maximum is 100
		numBlock = 100
	}
	// var err error
	result := make(map[string]metadata.Transaction)
	// bestBlock := blockchain.BestState[shardID].BestBlock
	// for {
	// 	for _, tx := range bestBlock.Transactions {
	// 		info := tx.GetInfo()
	// 		if info == nil {
	// 			continue
	// 		}
	// 		// info of tx with contain encrypted pubkey of creator in 1st 64bytes
	// 		lenInfo := 66
	// 		if len(info) < lenInfo {
	// 			continue
	// 		}
	// 		// decrypt to get pubkey data from info
	// 		pubkeyData, err1 := privacy.ElGamalDecrypt(key.Rk[:], info[0:lenInfo])
	// 		if err1 != nil {
	// 			continue
	// 		}
	// 		// compare to check pubkey
	// 		if !bytes.Equal(pubkeyData.Compress(), key.Pk[:]) {
	// 			continue
	// 		}
	// 		result[tx.Hash().String()] = tx
	// 	}
	// 	numBlock--
	// 	if numBlock == 0 {
	// 		break
	// 	}
	// 	bestBlock, err = blockchain.GetBlockByBlockHash(&bestBlock.Header.PrevBlockHash)
	// 	if err != nil {
	// 		break
	// 	}
	// }
	return result, nil
}

func (blockchain *BlockChain) IsReady(shard bool, shardID byte) bool {

	if shard {
		//TODO check shardChain ready
	} else {
		//TODO check beaconChain ready
	}

	return true
}

//func (blockchain *BlockChain) UpdateDividendPayout(block *Block) error {
//	for _, tx := range block.Transactions {
//		switch tx.GetMetadataType() {
//		case metadata.DividendMeta:
//			{
//				tx := tx.(*transaction.Tx)
//				meta := tx.Metadata.(*metadata.Dividend)
//				if tx.Proof == nil {
//					return errors.New("Miss output in tx")
//				}
//				for _, _ = range tx.Proof.OutputCoins {
//					keySet := cashec.KeySet{
//						PaymentAddress: meta.PaymentAddress,
//					}
//					vouts, err := blockchain.GetUnspentTxCustomTokenVout(keySet, meta.TokenID)
//					if err != nil {
//						return err
//					}
//					for _, vout := range vouts {
//						txHash := vout.GetTxCustomTokenID()
//						err := blockchain.config.DataBase.UpdateRewardAccountUTXO(meta.TokenID, keySet.PaymentAddress.Pk, &txHash, vout.GetIndex())
//						if err != nil {
//							return err
//						}
//					}
//				}
//			}
//		}
//	}
//	return nil
//}
//
//func (blockchain *BlockChain) UpdateVoteCountBoard(block *Block) error {
//	DCBBoardIndex := uint32(0)
//	GOVBoardIndex := uint32(0)
//	if block.Header.Height != 1 {
//		DCBBoardIndex = block.Header.DCBGovernor.BoardIndex + 1
//		GOVBoardIndex = block.Header.GOVGovernor.BoardIndex + 1
//	}
//	for _, tx := range block.Transactions {
//		switch tx.GetMetadataType() {
//		case metadata.VoteDCBBoardMeta:
//			{
//				txCustomToken := tx.(*transaction.TxCustomToken)
//				voteAmount := txCustomToken.GetAmountOfVote()
//				voteDCBBoardMetadata := txCustomToken.Metadata.(*metadata.VoteDCBBoardMetadata)
//				err := blockchain.config.DataBase.AddVoteBoard(common.DCBBoard, DCBBoardIndex, txCustomToken.TxTokenData.Vins[0].PaymentAddress.Bytes(), txCustomToken.TxTokenData.Vins[0].PaymentAddress, voteDCBBoardMetadata.CandidatePaymentAddress, voteAmount)
//				if err != nil {
//					return err
//				}
//			}
//		case metadata.VoteGOVBoardMeta:
//			{
//				txCustomToken := tx.(*transaction.TxCustomToken)
//				voteAmount := txCustomToken.GetAmountOfVote()
//				voteGOVBoardMetadata := txCustomToken.Metadata.(*metadata.VoteGOVBoardMetadata)
//				err := blockchain.config.DataBase.AddVoteBoard(common.GOVBoard, GOVBoardIndex, txCustomToken.TxTokenData.Vins[0].PaymentAddress.Bytes(), txCustomToken.TxTokenData.Vins[0].PaymentAddress, voteGOVBoardMetadata.CandidatePaymentAddress, voteAmount)
//				if err != nil {
//					return err
//				}
//			}
//		}
//	}
//	return nil
//}
//
//func (blockchain *BlockChain) UpdateVoteTokenHolderDB(block *Block) error {
//	for _, tx := range block.Transactions {
//		switch tx.GetMetadataType() {
//		case metadata.SendInitDCBVoteTokenMeta:
//			{
//				meta := tx.GetMetadata().(*metadata.SendInitDCBVoteTokenMetadata)
//				err := blockchain.config.DataBase.SendInitVoteToken(common.DCBBoard, block.Header.DCBGovernor.BoardIndex, meta.ReceiverPaymentAddress, meta.Amount)
//				if err != nil {
//					return err
//				}
//			}
//		case metadata.SendInitGOVVoteTokenMeta:
//			{
//				meta := tx.GetMetadata().(*metadata.SendInitGOVVoteTokenMetadata)
//				err := blockchain.config.DataBase.SendInitVoteToken(common.GOVBoard, block.Header.GOVGovernor.BoardIndex, meta.ReceiverPaymentAddress, meta.Amount)
//				if err != nil {
//					return err
//				}
//			}
//
//		}
//	}
//	return nil
//}
//
//func (blockchain *BlockChain) ProcessVoteProposal(block *Block) error {
//	nextDCBConstitutionIndex := uint32(block.Header.DCBConstitution.GetConstitutionIndex() + 1)
//	nextGOVConstitutionIndex := uint32(block.Header.GOVConstitution.GetConstitutionIndex() + 1)
//	for _, tx := range block.Transactions {
//		meta := tx.GetMetadata()
//		switch tx.GetMetadataType() {
//		case metadata.SealedLv3DCBVoteProposalMeta:
//			underlieMetadata := meta.(*metadata.SealedLv3DCBVoteProposalMetadata)
//			blockchain.config.DataBase.AddVoteLv3Proposal(common.DCBBoard, nextDCBConstitutionIndex, underlieMetadata.Hash())
//		case metadata.SealedLv2DCBVoteProposalMeta:
//			underlieMetadata := meta.(*metadata.SealedLv2DCBVoteProposalMetadata)
//			blockchain.config.DataBase.AddVoteLv1or2Proposal(common.DCBBoard, nextDCBConstitutionIndex, &underlieMetadata.SealedLv2VoteProposalMetadata.PointerToLv3VoteProposal)
//		case metadata.SealedLv1DCBVoteProposalMeta:
//			underlieMetadata := meta.(*metadata.SealedLv1DCBVoteProposalMetadata)
//			blockchain.config.DataBase.AddVoteLv1or2Proposal(common.DCBBoard, nextDCBConstitutionIndex, &underlieMetadata.SealedLv1VoteProposalMetadata.PointerToLv3VoteProposal)
//		case metadata.NormalDCBVoteProposalFromOwnerMeta:
//			underlieMetadata := meta.(*metadata.NormalDCBVoteProposalFromOwnerMetadata)
//			blockchain.config.DataBase.AddVoteNormalProposalFromOwner(common.DCBBoard, nextDCBConstitutionIndex, &underlieMetadata.NormalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal, underlieMetadata.NormalVoteProposalFromOwnerMetadata.VoteProposal.ToBytes())
//		case metadata.NormalDCBVoteProposalFromSealerMeta:
//			underlieMetadata := meta.(*metadata.NormalDCBVoteProposalFromSealerMetadata)
//			blockchain.config.DataBase.AddVoteNormalProposalFromSealer(common.DCBBoard, nextDCBConstitutionIndex, &underlieMetadata.NormalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal, underlieMetadata.NormalVoteProposalFromSealerMetadata.VoteProposal.ToBytes())
//		case metadata.AcceptDCBProposalMeta:
//			underlieMetadata := meta.(*metadata.AcceptDCBProposalMetadata)
//			blockchain.config.DataBase.TakeVoteTokenFromWinner(common.DCBBoard, nextDCBConstitutionIndex, underlieMetadata.Voter.PaymentAddress, underlieMetadata.Voter.AmountOfVote)
//			blockchain.config.DataBase.SetNewProposalWinningVoter(common.DCBBoard, nextDCBConstitutionIndex, underlieMetadata.Voter.PaymentAddress)
//		case metadata.SealedLv3GOVVoteProposalMeta:
//			underlieMetadata := meta.(*metadata.SealedLv3GOVVoteProposalMetadata)
//			blockchain.config.DataBase.AddVoteLv3Proposal(common.GOVBoard, nextGOVConstitutionIndex, underlieMetadata.Hash())
//		case metadata.SealedLv2GOVVoteProposalMeta:
//			underlieMetadata := meta.(*metadata.SealedLv2GOVVoteProposalMetadata)
//			blockchain.config.DataBase.AddVoteLv1or2Proposal(common.GOVBoard, nextGOVConstitutionIndex, &underlieMetadata.SealedLv2VoteProposalMetadata.PointerToLv3VoteProposal)
//		case metadata.SealedLv1GOVVoteProposalMeta:
//			underlieMetadata := meta.(*metadata.SealedLv1GOVVoteProposalMetadata)
//			blockchain.config.DataBase.AddVoteLv1or2Proposal(common.GOVBoard, nextGOVConstitutionIndex, &underlieMetadata.SealedLv1VoteProposalMetadata.PointerToLv3VoteProposal)
//		case metadata.NormalGOVVoteProposalFromOwnerMeta:
//			underlieMetadata := meta.(*metadata.NormalGOVVoteProposalFromOwnerMetadata)
//			blockchain.config.DataBase.AddVoteNormalProposalFromOwner(common.GOVBoard, nextGOVConstitutionIndex, &underlieMetadata.NormalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal, underlieMetadata.NormalVoteProposalFromOwnerMetadata.VoteProposal.ToBytes())
//		case metadata.NormalGOVVoteProposalFromSealerMeta:
//			underlieMetadata := meta.(*metadata.NormalGOVVoteProposalFromSealerMetadata)
//			blockchain.config.DataBase.AddVoteNormalProposalFromSealer(common.GOVBoard, nextGOVConstitutionIndex, &underlieMetadata.NormalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal, underlieMetadata.NormalVoteProposalFromSealerMetadata.VoteProposal.ToBytes())
//		case metadata.AcceptGOVProposalMeta:
//			underlieMetadata := meta.(*metadata.AcceptGOVProposalMetadata)
//			blockchain.config.DataBase.TakeVoteTokenFromWinner(common.GOVBoard, nextGOVConstitutionIndex, underlieMetadata.Voter.PaymentAddress, underlieMetadata.Voter.AmountOfVote)
//			blockchain.config.DataBase.SetNewProposalWinningVoter(common.GOVBoard, nextGOVConstitutionIndex, underlieMetadata.Voter.PaymentAddress)
//		}
//	}
//	return nil
//}
//
//func (blockchain *BlockChain) ProcessCrowdsaleTxs(block *Block) error {
//	// Temp storage to update crowdsale data
//	saleDataMap := make(map[string]*params.SaleData)
//
//	for _, tx := range block.Transactions {
//		switch tx.GetMetadataType() {
//		case metadata.AcceptDCBProposalMeta:
//			{
//             DONE
//			}
//		case metadata.CrowdsalePaymentMeta:
//			{
//				err := blockchain.updateCrowdsalePaymentData(tx, saleDataMap)
//				if err != nil {
//					return err
//				}
//			}
//			//		case metadata.ReserveResponseMeta:
//			//			{
//			//				// TODO(@0xbunyip): move to another func
//			//				meta := tx.GetMetadata().(*metadata.ReserveResponse)
//			//				_, _, _, txRequest, err := blockchain.GetTransactionByHash(meta.RequestedTxID)
//			//				if err != nil {
//			//					return err
//			//				}
//			//				requestHash := txRequest.Hash()
//			//
//			//				hash := tx.Hash()
//			//				if err := blockchain.config.DataBase.StoreCrowdsaleResponse(requestHash[:], hash[:]); err != nil {
//			//					return err
//			//				}
//			//			}
//		}
//	}
//
//	// Save crowdsale data back into db
//	for _, data := range saleDataMap {
//		if err := blockchain.config.DataBase.StoreCrowdsaleData(
//			data.SaleID,
//			data.GetProposalTxHash(),
//			data.BuyingAmount,
//			data.SellingAmount,
//		); err != nil {
//			return err
//		}
//	}
//	return nil
//}

// func (blockchain *BlockChain) ProcessCMBTxs(block *Block) error {
// 	for _, tx := range block.Transactions {
// 		switch tx.GetMetadataType() {
// 		case metadata.CMBInitRequestMeta:
// 			{
// 				err := blockchain.processCMBInitRequest(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBInitResponseMeta:
// 			{
// 				err := blockchain.processCMBInitResponse(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBInitRefundMeta:
// 			{
// 				err := blockchain.processCMBInitRefund(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBDepositSendMeta:
// 			{
// 				err := blockchain.processCMBDepositSend(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBWithdrawRequestMeta:
// 			{
// 				err := blockchain.processCMBWithdrawRequest(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBWithdrawResponseMeta:
// 			{
// 				err := blockchain.processCMBWithdrawResponse(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}

// 	// Penalize late response for cmb withdraw request
// 	return blockchain.findLateWithdrawResponse(uint64(block.Header.Height))
// }
