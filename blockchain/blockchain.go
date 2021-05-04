package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalv3"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/privacy"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/txpool"
	"github.com/pkg/errors"
)

type BlockChain struct {
	BeaconChain *BeaconChain
	ShardChain  []*ShardChain
	config      Config
	cQuitSync   chan struct{}

	IsTest bool

	beaconViewCache *lru.Cache

	committeeByEpochCache *lru.Cache
}

// Config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	BTCChain      *btcrelaying.BlockChain
	BNBChainState *bnbrelaying.BNBChainState
	DataBase      map[int]incdb.Database
	MemCache      *memcache.MemoryCache
	Interrupt     <-chan struct{}
	ChainParams   *Params
	GenesisParams *GenesisParams
	RelayShards   []byte
	// NodeMode          string
	BlockGen          *BlockGenerator
	TxPool            TxPool
	TempTxPool        TxPool
	CRemovedTxs       chan metadata.Transaction
	FeeEstimator      map[byte]FeeEstimator
	IsBlockGenStarted bool
	PubSubManager     Pubsub
	Syncker           Syncker
	Server            Server
	ConsensusEngine   ConsensusEngine
	Highway           Highway
	PoolManager       *txpool.PoolManager

	relayShardLck sync.Mutex
	usingNewPool  bool
}

func NewBlockChain(config *Config, isTest bool) *BlockChain {
	bc := &BlockChain{}
	bc.config = *config
	bc.config.IsBlockGenStarted = false
	bc.IsTest = isTest
	bc.beaconViewCache, _ = lru.New(100)
	bc.cQuitSync = make(chan struct{})
	bc.GetBeaconBestState().Params = make(map[string]string)
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
	blockchain.beaconViewCache, _ = lru.New(100)
	blockchain.committeeByEpochCache, _ = lru.New(100)
	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any chain state, both it and the chain state
	// will be initialized to contain only the genesis block.
	if err := blockchain.InitChainState(); err != nil {
		return err
	}
	blockchain.cQuitSync = make(chan struct{})
	newTxPool := os.Getenv("TXPOOL_VERSION")
	if newTxPool == "1" {
		blockchain.config.usingNewPool = true
	} else {
		blockchain.config.usingNewPool = false
	}
	return nil
}

// InitChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
func (blockchain *BlockChain) InitChainState() error {
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	blockchain.BeaconChain = NewBeaconChain(multiview.NewMultiView(), blockchain.config.BlockGen, blockchain, common.BeaconChainKey)
	var err error
	blockchain.BeaconChain.hashHistory, err = lru.New(1000)
	blockchain.BeaconChain.committeeCache, err = lru.New(1000)
	if err != nil {
		return err
	}
	if err := blockchain.RestoreBeaconViews(); err != nil {
		Logger.log.Error("debug restore beacon fail, init", err)
		err := blockchain.initBeaconState()
		if err != nil {
			Logger.log.Error("debug beacon state init error")
			return err
		}
	}
	Logger.log.Infof("Init Beacon View height %+v", blockchain.BeaconChain.GetBestView().GetHeight())

	//beaconHash, err := statedb.GetBeaconBlockHashByIndex(blockchain.GetBeaconBestState().GetBeaconConsensusStateDB(), 1)
	//panic(beaconHash.String())
	wl, err := blockchain.GetWhiteList()
	if err != nil {
		Logger.log.Errorf("Can not get whitelist txs, error %v", err)
	}
	blockchain.ShardChain = make([]*ShardChain, blockchain.GetBeaconBestState().ActiveShards)
	for shard := 1; shard <= blockchain.GetBeaconBestState().ActiveShards; shard++ {
		shardID := byte(shard - 1)
		tp, err := blockchain.config.PoolManager.GetShardTxsPool(shardID)
		if err != nil {
			return err
		}
		tv := NewTxsVerifier(
			nil,
			tp,
			wl,
		)
		tp.UpdateTxVerifier(tv)
		blockchain.ShardChain[shardID] = NewShardChain(shard-1, multiview.NewMultiView(), blockchain.config.BlockGen, blockchain, common.GetShardChainKey(shardID), tp, tv)
		blockchain.ShardChain[shardID].hashHistory, err = lru.New(1000)
		if err != nil {
			return err
		}
		if err := blockchain.RestoreShardViews(shardID); err != nil {
			Logger.log.Error("debug restore shard fail, init")
			err := blockchain.InitShardState(shardID)
			if err != nil {
				Logger.log.Error("debug shard state init error")
				return err
			}
		}
		sBestState := blockchain.ShardChain[shardID].GetBestState()
		txDB := sBestState.GetCopiedTransactionStateDB()
		// Logger.log.Infof("[testperformance] SHARD %v | Init txDB from block %v, txdb roothash %v\n", shardID, sBestState.BestBlock.Header.Height, txDB)

		blockchain.ShardChain[shardID].TxsVerifier.UpdateTransactionStateDB(txDB)
		Logger.log.Infof("Init Shard View shardID %+v, height %+v", shardID, blockchain.ShardChain[shardID].GetFinalViewHeight())
	}

	return nil
}

func (blockchain *BlockChain) GetWhiteList() (map[string]interface{}, error) {
	netID := blockchain.config.ChainParams.Name
	res := map[string]interface{}{}
	whitelistData, err := ioutil.ReadFile("whitelist.json")
	if err != nil {
		return nil, err
	}
	type WhiteList struct {
		Data map[string][]string
	}
	whiteList := WhiteList{}
	err = json.Unmarshal(whitelistData, &whiteList)
	if err != nil {
		return nil, err
	}
	if wlByNetID, ok := whiteList.Data[netID]; ok {
		for _, txHash := range wlByNetID {
			res[txHash] = nil
		}
	}
	return res, nil
}

/*
// createChainState initializes both the database and the chain state to the
// genesis block.  This includes creating the necessary buckets and inserting
// the genesis block, so it must only be called on an uninitialized database.
*/
func (blockchain *BlockChain) InitShardState(shardID byte) error {
	// Create a new block from genesis block and set it as best block of chain
	initShardBlock := types.ShardBlock{}
	initShardBlock = *blockchain.config.ChainParams.GenesisShardBlock
	initShardBlock.Header.ShardID = shardID
	initShardBlockHeight := initShardBlock.Header.Height
	var committeeEngine committeestate.ShardCommitteeEngine

	if blockchain.config.ChainParams.StakingFlowV2Height == 1 {
		committeeEngine = committeestate.NewShardCommitteeEngineV2(1, initShardBlock.Header.Hash(), shardID, committeestate.NewShardCommitteeStateV2())
	} else {
		committeeEngine = committeestate.NewShardCommitteeEngineV1(1, initShardBlock.Header.Hash(), shardID, committeestate.NewShardCommitteeStateV1())
	}

	initShardState := NewBestStateShardWithConfig(shardID, blockchain.config.ChainParams, committeeEngine)
	beaconBlocks, err := blockchain.GetBeaconBlockByHeight(initShardBlockHeight)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlockError, err)
	}
	genesisBeaconBlock := beaconBlocks[0]

	err = initShardState.initShardBestState(blockchain, blockchain.GetShardChainDatabase(shardID), &initShardBlock, genesisBeaconBlock)
	if err != nil {
		return err
	}
	committeeChange := committeestate.NewCommitteeChange()
	committeeChange.ShardCommitteeAdded[shardID] = initShardState.GetShardCommittee()

	err = blockchain.processStoreShardBlock(initShardState, &initShardBlock, committeeChange, []*types.BeaconBlock{genesisBeaconBlock})
	if err != nil {
		return err
	}

	return nil
}

func (blockchain *BlockChain) initBeaconState() error {
	initBlock := blockchain.config.ChainParams.GenesisBeaconBlock
	var committeeEngine committeestate.BeaconCommitteeEngine

	if blockchain.config.ChainParams.StakingFlowV2Height == 1 {
		committeeEngine = committeestate.
			NewBeaconCommitteeEngineV2(1, initBlock.Header.Hash(),
				committeestate.NewBeaconCommitteeStateV2())
	} else {
		committeeEngine = committeestate.
			NewBeaconCommitteeEngineV1(
				1, initBlock.Header.Hash(),
				committeestate.NewBeaconCommitteeStateV1())
	}
	initBeaconBestState := NewBeaconBestStateWithConfig(blockchain.config.ChainParams, committeeEngine)
	err := initBeaconBestState.initBeaconBestState(initBlock, blockchain, blockchain.GetBeaconChainDatabase())
	if err != nil {
		return err
	}
	initBlockHash := initBeaconBestState.BestBlock.Header.Hash()
	initBlockHeight := initBeaconBestState.BestBlock.Header.Height
	// Insert new block into beacon chain
	if err := statedb.StoreAllShardCommittee(initBeaconBestState.consensusStateDB, initBeaconBestState.GetShardCommittee()); err != nil {
		return err
	}
	if err := statedb.StoreBeaconCommittee(initBeaconBestState.consensusStateDB, initBeaconBestState.GetBeaconCommittee()); err != nil {
		return err
	}

	committees := initBeaconBestState.GetShardCommitteeFlattenList()
	missingSignatureCounter := signaturecounter.NewDefaultSignatureCounter(committees)
	initBeaconBestState.SetMissingSignatureCounter(missingSignatureCounter)

	consensusRootHash, err := initBeaconBestState.consensusStateDB.Commit(true)
	err = initBeaconBestState.consensusStateDB.Database().TrieDB().Commit(consensusRootHash, false)
	if err != nil {
		return err
	}
	initBeaconBestState.consensusStateDB.ClearObjects()
	if err := rawdbv2.StoreBeaconBlockByHash(blockchain.GetBeaconChainDatabase(), initBlockHash, &initBeaconBestState.BestBlock); err != nil {
		Logger.log.Error("Error store beacon block", initBeaconBestState.BestBlockHash, "in beacon chain")
		return err
	}
	rawdbv2.StoreFinalizedBeaconBlockHashByIndex(blockchain.GetBeaconChainDatabase(), initBlockHeight, initBlockHash)

	// State Root Hash
	bRH := BeaconRootHash{
		ConsensusStateDBRootHash: consensusRootHash,
		FeatureStateDBRootHash:   common.EmptyRoot,
		RewardStateDBRootHash:    common.EmptyRoot,
		SlashStateDBRootHash:     common.EmptyRoot,
	}

	initBeaconBestState.ConsensusStateDBRootHash = consensusRootHash
	if err := rawdbv2.StoreBeaconRootsHash(blockchain.GetBeaconChainDatabase(), initBlockHash, bRH); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	// Insert new block into beacon chain
	blockchain.BeaconChain.multiView.AddView(initBeaconBestState)
	if err := blockchain.BackupBeaconViews(blockchain.GetBeaconChainDatabase()); err != nil {
		Logger.log.Error("Error Store best state for block", blockchain.GetBeaconBestState().BestBlockHash, "in beacon chain")
		return NewBlockChainError(UnExpectedError, err)
	}

	return nil
}

func (blockchain *BlockChain) GetClonedBeaconBestState() (*BeaconBestState, error) {
	result := NewBeaconBestState()
	err := result.cloneBeaconBestStateFrom(blockchain.GetBeaconBestState())
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetReadOnlyShard - return a copy of Shard of BestState
func (blockchain *BlockChain) GetClonedAllShardBestState() map[byte]*ShardBestState {
	result := make(map[byte]*ShardBestState)
	for _, v := range blockchain.ShardChain {
		sidState := NewShardBestState()
		err := sidState.cloneShardBestStateFrom(blockchain.ShardChain[v.GetShardID()].GetBestState())
		if err != nil {
			return nil
		}
		result[byte(v.GetShardID())] = sidState
	}
	return result
}

// GetReadOnlyShard - return a copy of Shard of BestState
func (blockchain *BlockChain) GetClonedAShardBestState(shardID byte) (*ShardBestState, error) {
	result := NewShardBestState()
	err := result.cloneShardBestStateFrom(blockchain.ShardChain[int(shardID)].GetBestState())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (blockchain *BlockChain) GetCurrentBeaconBlockHeight(shardID byte) uint64 {
	return blockchain.GetBeaconBestState().BestBlock.Header.Height
}

func (blockchain BlockChain) RandomCommitmentsProcess(usableInputCoins []*privacy.InputCoin, randNum int, shardID byte, tokenID *common.Hash) (commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) {
	param := transaction.NewRandomCommitmentsProcessParam(usableInputCoins, randNum, blockchain.GetBestStateShard(shardID).GetCopiedTransactionStateDB(), shardID, tokenID)
	return transaction.RandomCommitmentsProcess(param)
}

func (blockchain *BlockChain) GetActiveShardNumber() int {
	return blockchain.GetBeaconBestState().ActiveShards
}

func (blockchain *BlockChain) GetShardIDs() []int {
	shardIDs := []int{}
	for i := 0; i < blockchain.GetActiveShardNumber(); i++ {
		shardIDs = append(shardIDs, i)
	}
	return shardIDs
}

// -------------- Start of Blockchain retriever's implementation --------------
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

// -------------- Start of Blockchain BackUp And Restore --------------
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
	bestStateBytes, err := rawdbv2.GetShardBestState(blockchain.GetShardChainDatabase(shardID), shardID)
	if err != nil {
		return err
	}
	shardBestState := &ShardBestState{}
	err = json.Unmarshal(bestStateBytes, shardBestState)
	bestShardHeight := shardBestState.ShardHeight
	var i uint64
	for i = 1; i < bestShardHeight; i++ {
		shardBlocks, err := blockchain.GetShardBlockByHeight(i, shardID)
		if err != nil {
			return err
		}
		var shardBlock *types.ShardBlock
		for _, v := range shardBlocks {
			shardBlock = v
		}
		data, err := json.Marshal(shardBlocks)
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
			Logger.log.Infof("Backup Shard %+v Block %+v", shardBlock.Header.ShardID, i)
		}
		if i == bestShardHeight-1 {
			Logger.log.Infof("Finish Backup Shard %+v with Block %+v", shardBlock.Header.ShardID, i)
		}
	}
	return nil
}

func (blockchain *BlockChain) BackupBeaconChain(writer io.Writer) error {
	bestStateBytes, err := rawdbv2.GetBeaconViews(blockchain.GetBeaconChainDatabase())
	if err != nil {
		return err
	}
	beaconBestState := &BeaconBestState{}
	err = json.Unmarshal(bestStateBytes, beaconBestState)
	bestBeaconHeight := beaconBestState.BeaconHeight
	var i uint64
	for i = 1; i < bestBeaconHeight; i++ {
		beaconBlocks, err := blockchain.GetBeaconBlockByHeight(i)
		if err != nil {
			return err
		}
		beaconBlock := beaconBlocks[0]
		data, err := json.Marshal(beaconBlock)
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
			Logger.log.Infof("Backup Beacon Block %+v", i)
		}
		if i == bestBeaconHeight-1 {
			Logger.log.Infof("Finish Backup Beacon with Block %+v", i)
		}
	}
	return nil
}

//TODO:
// current implement: backup all view data
// Optimize: backup view -> backup view hash instead of view
// restore: get view from hash and create new view, then insert into multiview
/*
Backup all BeaconView into Database
*/
func (blockchain *BlockChain) BackupBeaconViews(db incdb.KeyValueWriter) error {
	allViews := []*BeaconBestState{}
	for _, v := range blockchain.BeaconChain.multiView.GetAllViewsWithBFS() {
		allViews = append(allViews, v.(*BeaconBestState))
	}
	b, _ := json.Marshal(allViews)
	return rawdbv2.StoreBeaconViews(db, b)
}

/*
Restart all BeaconView from Database
*/
func (blockchain *BlockChain) RestoreBeaconViews() error {
	allViews := []*BeaconBestState{}
	bcDB := blockchain.GetBeaconChainDatabase()
	b, err := rawdbv2.GetBeaconViews(bcDB)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &allViews)
	if err != nil {
		return err
	}
	sID := []int{}
	for i := 0; i < blockchain.config.ChainParams.ActiveShards; i++ {
		sID = append(sID, i)
	}

	blockchain.BeaconChain.multiView.Reset()
	for _, v := range allViews {
		if err := v.RestoreBeaconViewStateFromHash(blockchain, true); err != nil {
			return NewBlockChainError(BeaconError, err)
		}
		// finish reproduce
		if !blockchain.BeaconChain.multiView.AddView(v) {
			panic("Restart beacon views fail")
		}
	}
	for _, beaconState := range allViews {
		if beaconState.missingSignatureCounter == nil {
			block := beaconState.BestBlock
			err = initMissingSignatureCounter(blockchain, beaconState, &block)
			if err != nil {
				return err
			}
			Logger.log.Infof("Init Missing Signature Counter, %+v, height %+v", beaconState.missingSignatureCounter, beaconState.BeaconHeight)
		}
	}
	return nil
}

/*
Backup shard views
*/
func (blockchain *BlockChain) BackupShardViews(db incdb.KeyValueWriter, shardID byte) error {
	allViews := []*ShardBestState{}
	for _, v := range blockchain.ShardChain[shardID].multiView.GetAllViewsWithBFS() {
		allViews = append(allViews, v.(*ShardBestState))
	}
	// fmt.Println("debug BackupShardViews", len(allViews))
	return rawdbv2.StoreShardBestState(db, shardID, allViews)
}

/*
Restart all BeaconView from Database
*/
func (blockchain *BlockChain) RestoreShardViews(shardID byte) error {
	allViews := []*ShardBestState{}
	b, err := rawdbv2.GetShardBestState(blockchain.GetShardChainDatabase(shardID), shardID)
	if err != nil {
		fmt.Println("debug Cannot see shard best state")
		return err
	}
	err = json.Unmarshal(b, &allViews)
	if err != nil {
		fmt.Println("debug Cannot unmarshall shard best state", string(b))
		return err
	}
	// fmt.Println("debug RestoreShardViews", len(allViews))
	blockchain.ShardChain[shardID].multiView.Reset()

	for _, v := range allViews {
		block, _, err := blockchain.GetShardBlockByHash(v.BestBlockHash)
		if err != nil || block == nil {
			fmt.Println("block ", block)
			panic(err)
		}
		v.BestBlock = block

		err = v.InitStateRootHash(blockchain.GetShardChainDatabase(shardID), blockchain)
		if err != nil {
			panic(err)
		}
		var shardCommitteeEngine committeestate.ShardCommitteeEngine
		if v.BeaconHeight > blockchain.config.ChainParams.StakingFlowV2Height {
			shardCommitteeEngine = InitShardCommitteeEngineV2(
				v.consensusStateDB,
				v.ShardHeight, v.ShardID, v.BestBlockHash,
				block.Header.CommitteeFromBlock, blockchain)
		} else {
			shardCommitteeEngine = InitShardCommitteeEngineV1(
				v.consensusStateDB, v.ShardHeight, v.ShardID, v.BestBlockHash)
		}
		v.shardCommitteeEngine = shardCommitteeEngine
		if v.BeaconHeight == blockchain.config.ChainParams.StakingFlowV2Height {
			err := v.upgradeCommitteeEngineV2(blockchain)
			if err != nil {
				panic(err)
			}
		}
		if !blockchain.ShardChain[shardID].multiView.AddView(v) {
			panic("Restart shard views fail")
		}
	}
	return nil
}

func (blockchain *BlockChain) GetShardStakingTx(shardID byte, beaconHeight uint64) (map[string]string, error) {
	//build staking tx
	beaconConsensusRootHash, err := blockchain.GetBeaconConsensusRootHash(blockchain.GetBeaconBestState(), beaconHeight)
	if err != nil {
		Logger.log.Error("Cannot restore shard, beacon not ready!")
		return nil, err
	}

	beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
	if err != nil {
		Logger.log.Error("Cannot restore shard, beacon not ready!")
		return nil, err
	}
	mapStakingTx, err := beaconConsensusStateDB.GetAllStakingTX(blockchain.GetShardIDs())

	if err != nil {
		fmt.Println(err)
		panic("Something wrong when retrieve mapStakingTx")
	}

	sdb := blockchain.GetShardChainDatabase(byte(shardID))
	shardStakingTx := map[string]string{}
	for _, stakingtx := range mapStakingTx {
		if stakingtx != common.HashH([]byte{0}).String() {
			stakingTxHash, _ := common.Hash{}.NewHashFromStr(stakingtx)
			blockHash, txindex, err := rawdbv2.GetTransactionByHash(sdb, *stakingTxHash)
			if err != nil { //no transaction in this node
				continue
			}
			shardBlockBytes, err := rawdbv2.GetShardBlockByHash(sdb, blockHash)
			if err != nil { //no transaction in this node
				panic("Have transaction but cannot found block")
			}
			shardBlock := types.NewShardBlock()
			err = json.Unmarshal(shardBlockBytes, shardBlock)
			if err != nil {
				panic("Cannot unmarshal shardblock")
			}
			if shardBlock.GetShardID() != int(shardID) {
				continue
			}
			txData := shardBlock.Body.Transactions[txindex]
			committeePk := txData.GetMetadata().(*metadata.StakingMetadata).CommitteePublicKey
			shardStakingTx[committeePk] = stakingtx
		}
	}
	return shardStakingTx, nil
}

// -------------- End of Blockchain BackUp And Restore --------------

// func (blockchain *BlockChain) GetNodeMode() string {
// 	return blockchain.config.NodeMode
// }

func (blockchain *BlockChain) GetWantedShard(isBeaconCommittee bool) map[byte]struct{} {
	res := map[byte]struct{}{}
	if isBeaconCommittee {
		for sID := byte(0); sID < byte(blockchain.config.ChainParams.ActiveShards); sID++ {
			res[sID] = struct{}{}
		}
	} else {
		blockchain.config.relayShardLck.Lock()
		for _, sID := range blockchain.config.RelayShards {
			res[sID] = struct{}{}
		}
		blockchain.config.relayShardLck.Unlock()
	}
	return res
}

// GetConfig returns blockchain's config
func (blockchain *BlockChain) GetConfig() *Config {
	return &blockchain.config
}

// GetPortalParams returns portal params in beaconheight
func (blockchain *BlockChain) GetPortalParams() portal.PortalParams {
	return blockchain.GetConfig().ChainParams.PortalParams
}

func (blockchain *BlockChain) GetPortalParamsV3(beaconHeight uint64) portalv3.PortalParams {
	return blockchain.GetConfig().ChainParams.PortalParams.GetPortalParamsV3(beaconHeight)
}

func (blockchain *BlockChain) GetBeaconChainDatabase() incdb.Database {
	return blockchain.config.DataBase[common.BeaconChainDataBaseID]
}

func (blockchain *BlockChain) GetShardChainDatabase(shardID byte) incdb.Database {
	return blockchain.config.DataBase[int(shardID)]
}

func (blockchain *BlockChain) GetBeaconViewStateDataFromBlockHash(blockHash common.Hash, includeCommittee bool) (*BeaconBestState, error) {
	v, ok := blockchain.beaconViewCache.Get(blockHash)
	if ok {
		return v.(*BeaconBestState), nil
	}
	bcDB := blockchain.GetBeaconChainDatabase()
	rootHash, err := rawdbv2.GetBeaconRootsHash(bcDB, blockHash)
	if err != nil {
		return nil, err
	}
	bRH := &BeaconRootHash{}
	err = json.Unmarshal(rootHash, bRH)
	if err != nil {
		return nil, err
	}

	beaconView := &BeaconBestState{
		BestBlockHash:            blockHash,
		ActiveShards:             blockchain.config.ChainParams.ActiveShards, //we assume active shard not change (if not, we must store active shard in db)
		ConsensusStateDBRootHash: bRH.ConsensusStateDBRootHash,
		FeatureStateDBRootHash:   bRH.FeatureStateDBRootHash,
		RewardStateDBRootHash:    bRH.RewardStateDBRootHash,
		SlashStateDBRootHash:     bRH.SlashStateDBRootHash,
	}

	err = beaconView.RestoreBeaconViewStateFromHash(blockchain, includeCommittee)
	if err != nil {
		Logger.log.Error(err)
	}
	sID := []int{}
	for i := 0; i < blockchain.config.ChainParams.ActiveShards; i++ {
		sID = append(sID, i)
	}
	return beaconView, err
}

// GetFixedRandomForShardIDCommitment returns the fixed randomness for shardID commitments
// if bc height is greater than or equal to BCHeightBreakPointFixRandShardCM
// otherwise, return nil
func (blockchain *BlockChain) GetFixedRandomForShardIDCommitment(beaconHeight uint64) *privacy.Scalar {
	if beaconHeight == 0 {
		beaconHeight = blockchain.GetBeaconBestState().GetHeight()
	}
	if beaconHeight >= blockchain.GetConfig().ChainParams.BCHeightBreakPointNewZKP {
		return privacy.FixedRandomnessShardID
	}

	return nil
}

func (blockchain *BlockChain) IsAfterNewZKPCheckPoint(beaconHeight uint64) bool {
	if beaconHeight == 0 {
		beaconHeight = blockchain.GetBeaconBestState().GetHeight()
	}

	return beaconHeight >= blockchain.GetConfig().ChainParams.BCHeightBreakPointNewZKP
}

func (s *BlockChain) GetChainParams() *Params {
	return s.config.ChainParams
}

func (s *BlockChain) AddRelayShard(sid int) error {
	s.config.relayShardLck.Lock()
	for _, shard := range s.config.RelayShards {
		if shard == byte(sid) {
			s.config.relayShardLck.Unlock()
			return errors.New("already relay this shard" + strconv.Itoa(sid))
		}
	}
	s.config.RelayShards = append(s.config.RelayShards, byte(sid))
	s.config.relayShardLck.Unlock()
	return nil
}

func (s *BlockChain) RemoveRelayShard(sid int) {
	s.config.relayShardLck.Lock()
	for idx, shard := range s.config.RelayShards {
		if shard == byte(sid) {
			s.config.RelayShards = append(s.config.RelayShards[:idx], s.config.RelayShards[idx+1:]...)
			break
		}
	}
	s.config.relayShardLck.Unlock()
	return
}

// GetEpochLength return the current length of epoch
// it depends on current final view height
func (bc *BlockChain) GetCurrentEpochLength(beaconHeight uint64) uint64 {
	params := bc.config.ChainParams
	if params.EpochV2BreakPoint == 0 {
		return params.Epoch
	}
	changeEpochBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
	if beaconHeight > changeEpochBreakPoint {
		return params.EpochV2
	} else {
		return params.Epoch
	}
}

func (bc *BlockChain) GetEpochByHeight(beaconHeight uint64) uint64 {
	params := bc.config.ChainParams
	totalBlockBeforeBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
	newEpochBlockHeight := totalBlockBeforeBreakPoint + 1
	if beaconHeight < newEpochBlockHeight {
		if beaconHeight%params.Epoch == 0 {
			return beaconHeight / params.Epoch
		} else {
			return beaconHeight/params.Epoch + 1
		}
	} else {
		newEpochBlocks := beaconHeight - totalBlockBeforeBreakPoint
		numberOfNewEpochs := newEpochBlocks / params.EpochV2
		if newEpochBlocks%params.EpochV2 != 0 {
			numberOfNewEpochs++
		}
		return (totalBlockBeforeBreakPoint / params.Epoch) +
			numberOfNewEpochs
	}
}

func (bc *BlockChain) GetEpochNextHeight(beaconHeight uint64) (uint64, bool) {
	beaconHeight++
	return bc.getEpochAndIsFistHeightInEpoch(beaconHeight)
}

func (bc *BlockChain) getEpochAndIsFistHeightInEpoch(beaconHeight uint64) (uint64, bool) {
	params := bc.config.ChainParams
	totalBlockBeforeBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
	newEpochBlockHeight := totalBlockBeforeBreakPoint + 1
	if beaconHeight < newEpochBlockHeight {
		newEpoch := beaconHeight/params.Epoch + 1
		if beaconHeight%params.Epoch == 1 {
			if beaconHeight == 1 {
				return newEpoch, false
			} else {
				return newEpoch, true
			}
		} else {
			if beaconHeight%params.Epoch == 0 {
				return newEpoch - 1, false
			} else {
				return newEpoch, false
			}
		}
	} else {
		newEpochBlocks := beaconHeight - totalBlockBeforeBreakPoint
		numberOfNewEpochs := newEpochBlocks / params.EpochV2
		numberOfOldEpochs := params.EpochV2BreakPoint - 1
		if newEpochBlocks%params.EpochV2 != 0 {
			numberOfNewEpochs++
		}
		if newEpochBlocks%params.EpochV2 == 1 {
			return numberOfOldEpochs + numberOfNewEpochs, true
		} else {
			return numberOfOldEpochs + numberOfNewEpochs, false
		}
	}
}

func (bc *BlockChain) IsFirstBeaconHeightInEpoch(beaconHeight uint64) bool {
	_, ok := bc.getEpochAndIsFistHeightInEpoch(beaconHeight)
	return ok
}

func (bc *BlockChain) IsLastBeaconHeightInEpoch(beaconHeight uint64) bool {
	_, ok := bc.getEpochAndIsFistHeightInEpoch(beaconHeight + 1)
	return ok
}

func (bc *BlockChain) GetRandomTimeInEpoch(epoch uint64) uint64 {
	params := bc.config.ChainParams
	if epoch < params.EpochV2BreakPoint {
		return (epoch-1)*params.Epoch + params.RandomTime
	} else {
		totalBlockBeforeBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
		numberOfNewEpoch := epoch - params.EpochV2BreakPoint
		beaconHeightRandomTimeAfterBreakPoint := numberOfNewEpoch*params.EpochV2 + params.RandomTimeV2
		res := totalBlockBeforeBreakPoint + beaconHeightRandomTimeAfterBreakPoint
		return res
	}
}

func (bc *BlockChain) GetFirstBeaconHeightInEpoch(epoch uint64) uint64 {
	params := bc.config.ChainParams
	if epoch < params.EpochV2BreakPoint {
		return (epoch-1)*params.Epoch + 1
	} else {
		totalBlockBeforeBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
		numberOfNewEpoch := epoch - params.EpochV2BreakPoint
		lastBeaconHeightAfterBreakPoint := numberOfNewEpoch*params.EpochV2 + 1
		return totalBlockBeforeBreakPoint + lastBeaconHeightAfterBreakPoint
	}
}

func (bc *BlockChain) GetLastBeaconHeightInEpoch(epoch uint64) uint64 {
	params := bc.config.ChainParams
	if epoch < params.EpochV2BreakPoint {
		return epoch * params.Epoch
	} else {
		totalBlockBeforeBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
		numberOfNewEpoch := epoch - params.EpochV2BreakPoint + 1
		lastBeaconHeightAfterBreakPoint := numberOfNewEpoch * params.EpochV2
		return totalBlockBeforeBreakPoint + lastBeaconHeightAfterBreakPoint
	}
}

func (bc *BlockChain) GetBeaconBlockOrderInEpoch(beaconHeight uint64) (uint64, uint64) {
	params := bc.config.ChainParams
	totalBlockBeforeBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
	if beaconHeight < totalBlockBeforeBreakPoint {
		return beaconHeight % params.Epoch, params.Epoch - beaconHeight%params.Epoch
	} else {
		newEpochBlocks := beaconHeight - totalBlockBeforeBreakPoint
		return newEpochBlocks % params.EpochV2, params.EpochV2 - newEpochBlocks%params.EpochV2
	}
}

func (bc *BlockChain) IsGreaterThanRandomTime(beaconHeight uint64) bool {
	params := bc.config.ChainParams
	totalBlockBeforeBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
	if beaconHeight < totalBlockBeforeBreakPoint {
		return beaconHeight%params.Epoch > params.RandomTime
	} else {
		newEpochBlocks := beaconHeight - totalBlockBeforeBreakPoint
		return newEpochBlocks%params.EpochV2 > params.RandomTimeV2
	}
}

func (bc *BlockChain) IsEqualToRandomTime(beaconHeight uint64) bool {
	params := bc.config.ChainParams
	totalBlockBeforeBreakPoint := params.Epoch * (params.EpochV2BreakPoint - 1)
	if beaconHeight < totalBlockBeforeBreakPoint {
		return beaconHeight%params.Epoch == params.RandomTime
	} else {
		newEpochBlocks := beaconHeight - totalBlockBeforeBreakPoint
		return newEpochBlocks%params.EpochV2 == params.RandomTimeV2
	}
}

func (bc *BlockChain) GetAllCommitteeStakeInfoByEpoch(epoch uint64) (map[int][]*statedb.StakerInfo, error) {
	height := bc.GetLastBeaconHeightInEpoch(epoch)
	var beaconConsensusRootHash common.Hash
	beaconConsensusRootHash, err := bc.GetBeaconConsensusRootHash(bc.GetBeaconBestState(), height)
	if err != nil {
		return nil, NewBlockChainError(ProcessSalaryInstructionsError, fmt.Errorf("Beacon Consensus Root Hash of Height %+v not found ,error %+v", height, err))
	}
	beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(bc.GetBeaconChainDatabase()))
	if err != nil {
		return nil, NewBlockChainError(ProcessSalaryInstructionsError, err)
	}
	if cState, has := bc.committeeByEpochCache.Peek(epoch); has {
		if result, ok := cState.(map[int][]*statedb.CommitteeState); ok {
			return statedb.GetAllCommitteeStakeInfoV2(beaconConsensusStateDB, result), nil
		}
	}
	allCommitteeState := statedb.GetAllCommitteeState(beaconConsensusStateDB, bc.GetShardIDs())
	bc.committeeByEpochCache.Add(epoch, allCommitteeState)
	return statedb.GetAllCommitteeStakeInfoV2(beaconConsensusStateDB, allCommitteeState), nil
}

func (blockchain *BlockChain) GetPoolManager() *txpool.PoolManager {
	return blockchain.config.PoolManager
}

func (blockchain *BlockChain) UsingNewPool() bool {
	return blockchain.config.usingNewPool
}
