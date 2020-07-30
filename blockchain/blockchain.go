package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/btc"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/pubsub"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/transaction"
	libp2p "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	"io"
	"sort"
	"sync"
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

// Config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	BTCChain          *btcrelaying.BlockChain
	BNBChainState     *bnbrelaying.BNBChainState
	DataBase          incdb.Database
	MemCache          *memcache.MemoryCache
	Interrupt         <-chan struct{}
	ChainParams       *Params
	GenesisParams     *GenesisParams
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

// initChainState attempts to load and initialize the chain state from the
// database.  When the db does not yet contain any chain state, both it and the
// chain state are initialized to the genesis block.
func (blockchain *BlockChain) initChainState() error {
	// Determine the state of the chain database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool
	blockchain.Chains = make(map[string]ChainInterface)
	blockchain.BestState = &BestState{
		Beacon: nil,
		Shard:  make(map[byte]*ShardBestState),
	}

	bestStateBeaconBytes, err := rawdbv2.GetBeaconBestState(blockchain.GetDatabase())
	if err == nil {
		beacon := &BeaconBestState{}
		err = json.Unmarshal(bestStateBeaconBytes, beacon)
		//update singleton object
		SetBeaconBestState(beacon)
		//update beacon field in blockchain Beststate
		blockchain.BestState.Beacon = GetBeaconBestState()
		errStateDB := blockchain.BestState.Beacon.InitStateRootHash(blockchain)
		if errStateDB != nil {
			return errStateDB
		}
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
		bestStateBytes, err := rawdbv2.GetShardBestState(blockchain.config.DataBase, shardID)
		if err == nil {
			shardBestState := &ShardBestState{}
			err = json.Unmarshal(bestStateBytes, shardBestState)
			//update singleton object
			SetBestStateShard(shardID, shardBestState)
			//update Shard field in blockchain Beststate
			blockchain.BestState.Shard[shardID] = GetBestStateShard(shardID)
			errStateDB := blockchain.BestState.Shard[shardID].InitStateRootHash(blockchain.GetDatabase(), blockchain)
			if errStateDB != nil {
				return errStateDB
			}
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

	//go blockchain.buildNewStakingTx()
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
	initShardBlock := ShardBlock{}
	initShardBlock = *blockchain.config.ChainParams.GenesisShardBlock
	initShardBlock.Header.ShardID = shardID
	initShardBlockHeight := initShardBlock.Header.Height
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
	tempShardBestState := blockchain.BestState.Shard[shardID]
	beaconBlocks, err := blockchain.GetBeaconBlockByHeight(initShardBlockHeight)
	genesisBeaconBlock := beaconBlocks[0]
	if err != nil {
		return NewBlockChainError(FetchBeaconBlockError, err)
	}
	err = blockchain.BestState.Shard[shardID].initShardBestState(blockchain, blockchain.GetDatabase(), &initShardBlock, genesisBeaconBlock)
	if err != nil {
		return err
	}
	committeeChange := newCommitteeChange()
	committeeChange.shardCommitteeAdded[shardID] = tempShardBestState.GetShardCommittee()
	err = blockchain.processStoreShardBlock(&initShardBlock, committeeChange, genesisBeaconBlock)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) initBeaconState() error {
	blockchain.BestState.Beacon = NewBeaconBestStateWithConfig(blockchain.config.ChainParams)
	initBlock := blockchain.config.ChainParams.GenesisBeaconBlock
	err := blockchain.BestState.Beacon.initBeaconBestState(initBlock, blockchain, blockchain.GetDatabase())
	if err != nil {
		return err
	}
	tempBeaconBestState := blockchain.BestState.Beacon
	beaconBestStateBytes, err := json.Marshal(tempBeaconBestState)
	if err != nil {
		return NewBlockChainError(InitBeaconStateError, err)
	}
	blockchain.BestState.Beacon.lock.Lock()
	defer blockchain.BestState.Beacon.lock.Unlock()
	initBlockHash := tempBeaconBestState.BestBlock.Header.Hash()
	initBlockHeight := tempBeaconBestState.BestBlock.Header.Height
	// Insert new block into beacon chain
	if err := statedb.StoreAllShardCommittee(tempBeaconBestState.consensusStateDB, tempBeaconBestState.ShardCommittee, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking); err != nil {
		return err
	}
	if err := statedb.StoreBeaconCommittee(tempBeaconBestState.consensusStateDB, tempBeaconBestState.BeaconCommittee, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking); err != nil {
		return err
	}
	consensusRootHash, err := beaconBestState.consensusStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = beaconBestState.consensusStateDB.Database().TrieDB().Commit(consensusRootHash, false)
	if err != nil {
		return err
	}
	beaconBestState.consensusStateDB.ClearObjects()
	if err := rawdbv2.StoreBeaconBestState(blockchain.GetDatabase(), beaconBestStateBytes); err != nil {
		Logger.log.Error("Error Store best state for block", blockchain.BestState.Beacon.BestBlockHash, "in beacon chain")
		return NewBlockChainError(InitBeaconStateError, err)
	}
	if err := rawdbv2.StoreBeaconBlock(blockchain.GetDatabase(), initBlockHeight, initBlockHash, &tempBeaconBestState.BestBlock); err != nil {
		Logger.log.Error("Error store beacon block", tempBeaconBestState.BestBlockHash, "in beacon chain")
		return err
	}
	if err := rawdbv2.StoreBeaconBlockIndex(blockchain.GetDatabase(), initBlockHeight, initBlockHash); err != nil {
		return err
	}
	// State Root Hash
	if err := rawdbv2.StoreConsensusStateRootHash(blockchain.GetDatabase(), initBlockHeight, consensusRootHash); err != nil {
		return err
	}
	if err := rawdbv2.StoreRewardStateRootHash(blockchain.GetDatabase(), initBlockHeight, common.EmptyRoot); err != nil {
		return err
	}
	if err := rawdbv2.StoreFeatureStateRootHash(blockchain.GetDatabase(), initBlockHeight, common.EmptyRoot); err != nil {
		return err
	}
	if err := rawdbv2.StoreSlashStateRootHash(blockchain.GetDatabase(), initBlockHeight, common.EmptyRoot); err != nil {
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

func (blockchain *BlockChain) GetCurrentBeaconBlockHeight(shardID byte) uint64 {
	return blockchain.BestState.Beacon.BestBlock.Header.Height
}

func (blockchain BlockChain) RandomCommitmentsProcess(usableInputCoins []*privacy.InputCoin, randNum int, shardID byte, tokenID *common.Hash) (commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) {
	param := transaction.NewRandomCommitmentsProcessParam(usableInputCoins, randNum, blockchain.BestState.Shard[shardID].GetCopiedTransactionStateDB(), shardID, tokenID)
	return transaction.RandomCommitmentsProcess(param)
}

func (blockchain *BlockChain) GetActiveShardNumber() int {
	return blockchain.BestState.Beacon.ActiveShards
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
	bestStateBytes, err := rawdbv2.GetShardBestState(blockchain.config.DataBase, shardID)
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
		var shardBlock *ShardBlock
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
	bestStateBytes, err := rawdbv2.GetBeaconBestState(blockchain.GetDatabase())
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

// -------------- End of Blockchain BackUp And Restore --------------

// GetConfig returns blockchain's config
func (blockchain *BlockChain) GetConfig() *Config {
	return &blockchain.config
}

// GetPortalParams returns portal params in beaconheight
func (blockchain *BlockChain) GetPortalParams(beaconHeight uint64) PortalParams {
	portalParamMap := blockchain.GetConfig().ChainParams.PortalParams
	// only has one value - default value
	if len(portalParamMap) == 1 {
		return portalParamMap[0]
	}

	bchs := []uint64{}
	for bch := range portalParamMap {
		bchs = append(bchs, bch)
	}
	sort.Slice(bchs, func(i, j int) bool {
		return bchs[i] < bchs[j]
	})

	bchKey := bchs[len(bchs)-1]
	for i := len(bchs) - 1; i >= 0; i-- {
		if beaconHeight < bchs[i] {
			continue
		}
		bchKey = bchs[i]
		break
	}

	return portalParamMap[bchKey]
}
