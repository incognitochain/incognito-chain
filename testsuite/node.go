package devframework

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/incognitochain/incognito-chain/consensus_v2"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbftv2"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/testsuite/account"
	"github.com/incognitochain/incognito-chain/testsuite/mock"
	"github.com/incognitochain/incognito-chain/testsuite/rpcclient"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/incognitochain/incognito-chain/pubsub"

	"github.com/incognitochain/incognito-chain/syncker"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/transaction"

	lvdbErrors "github.com/syndtr/goleveldb/leveldb/errors"

	"github.com/pkg/errors"
)

type Config struct {
	ConsensusVersion int
	ChainParam       *ChainParam
}

type NodeEngine struct {
	config      Config
	appNodeMode string
	simName     string
	timer       *TimeEngine

	//for account manager
	accountSeed       string
	accountGenHistory map[int]int
	committeeAccount  map[int][]account.Account
	accounts          []*account.Account

	GenesisAccount account.Account

	//blockchain dependency object
	param       *blockchain.Params
	bc          *blockchain.BlockChain
	ps          *pubsub.PubSubManager
	consensus   mock.ConsensusInterface
	txpool      *mempool.TxPool
	temppool    *mempool.TxPool
	btcrd       *mock.BTCRandom
	syncker     *syncker.SynckerManager
	server      *mock.Server
	cPendingTxs chan metadata.Transaction
	cRemovedTxs chan metadata.Transaction
	rpcServer   *rpcserver.RpcServer
	cQuit       chan struct{}

	RPC               *rpcclient.RPCClient
	listennerRegister map[int][]func(msg interface{})

	userDB        *leveldb.DB
	lightNodeData struct {
		Shards                map[byte]*currentShardState
		ProcessedBeaconHeight uint64
	}
}

type currentShardState struct {
	// BestHeight  uint64
	// BestHash    *common.Hash
	LocalHeight uint64
	LocalHash   *common.Hash
}

func (sim *NodeEngine) NewAccountFromShard(sid int) account.Account {
	lastID := sim.accountGenHistory[sid]
	lastID++
	sim.accountGenHistory[sid] = lastID
	acc, _ := account.GenerateAccountByShard(sid, lastID, sim.accountSeed)
	acc.SetName(fmt.Sprintf("ACC_%v", len(sim.accounts)-len(sim.committeeAccount)+1))
	sim.accounts = append(sim.accounts, &acc)
	return acc
}

func (sim *NodeEngine) NewAccount() account.Account {
	lastID := sim.accountGenHistory[0]
	lastID++
	sim.accountGenHistory[0] = lastID
	acc, _ := account.GenerateAccountByShard(0, lastID, sim.accountSeed)
	return acc
}

func (sim *NodeEngine) EnableDebug() {
	dbLogger.SetLevel(common.LevelTrace)
	blockchainLogger.SetLevel(common.LevelInfo)
	bridgeLogger.SetLevel(common.LevelTrace)
	rpcLogger.SetLevel(common.LevelTrace)
	rpcServiceLogger.SetLevel(common.LevelTrace)
	rpcServiceBridgeLogger.SetLevel(common.LevelTrace)
	transactionLogger.SetLevel(common.LevelTrace)
	privacyLogger.SetLevel(common.LevelTrace)
	mempoolLogger.SetLevel(common.LevelTrace)
}

func (sim *NodeEngine) init() {
	simName := sim.simName
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	InitLogRotator(filepath.Join(path, simName+".log"))

	activeNetParams := sim.config.ChainParam.GetParamData()
	common.MaxShardNumber = activeNetParams.ActiveShards
	common.TIMESLOT = activeNetParams.Timeslot
	sim.GenesisAccount = sim.NewAccount()

	for i := 0; i < activeNetParams.MinBeaconCommitteeSize; i++ {
		acc := sim.NewAccountFromShard(-1)
		sim.committeeAccount[-1] = append(sim.committeeAccount[-1], acc)
		activeNetParams.GenesisParams.PreSelectBeaconNodeSerializedPubkey = append(activeNetParams.GenesisParams.PreSelectBeaconNodeSerializedPubkey, acc.SelfCommitteePubkey)
		activeNetParams.GenesisParams.PreSelectBeaconNodeSerializedPaymentAddress = append(activeNetParams.GenesisParams.PreSelectBeaconNodeSerializedPaymentAddress, acc.PaymentAddress)
	}
	for i := 0; i < activeNetParams.ActiveShards; i++ {
		for a := 0; a < activeNetParams.MinShardCommitteeSize; a++ {
			acc := sim.NewAccountFromShard(i)
			sim.committeeAccount[i] = append(sim.committeeAccount[i], acc)
			activeNetParams.GenesisParams.PreSelectShardNodeSerializedPubkey = append(activeNetParams.GenesisParams.PreSelectShardNodeSerializedPubkey, acc.SelfCommitteePubkey)
			activeNetParams.GenesisParams.PreSelectShardNodeSerializedPaymentAddress = append(activeNetParams.GenesisParams.PreSelectShardNodeSerializedPaymentAddress, acc.PaymentAddress)
		}
	}

	initTxs := createGenesisTx([]account.Account{sim.GenesisAccount})
	activeNetParams.GenesisParams.InitialIncognito = initTxs
	activeNetParams.CreateGenesisBlocks()

	//init blockchain
	bc := blockchain.BlockChain{}

	sim.timer.init(activeNetParams.GenesisBeaconBlock.Header.Timestamp + 10)

	cs := mock.Consensus{}
	txpool := mempool.TxPool{}
	temppool := mempool.TxPool{}
	btcrd := mock.BTCRandom{} // use mock for now
	sync := syncker.NewSynckerManager()
	server := mock.Server{
		BlockChain: &bc,
	}
	ps := pubsub.NewPubSubManager()
	fees := make(map[byte]*mempool.FeeEstimator)
	for i := byte(0); i < byte(activeNetParams.ActiveShards); i++ {
		fees[i] = mempool.NewFeeEstimator(
			mempool.DefaultEstimateFeeMaxRollback,
			mempool.DefaultEstimateFeeMinRegisteredBlocks,
			0)
	}
	cPendingTxs := make(chan metadata.Transaction, 500)
	cRemovedTxs := make(chan metadata.Transaction, 500)
	cQuit := make(chan struct{})
	blockgen, err := blockchain.NewBlockGenerator(&txpool, &bc, sync, cPendingTxs, cRemovedTxs)
	if err != nil {
		panic(err)
	}
	dbpath := filepath.Join("/tmp/database")
	db, err := incdb.OpenMultipleDB("leveldb", dbpath)
	// Create db and use it.
	if err != nil {
		panic(err)
	}

	//listenFunc := net.Listen
	//listener, err := listenFunc("tcp", "0.0.0.0:8000")
	//if err != nil {
	//	panic(err)
	//}

	rpcConfig := rpcserver.RpcServerConfig{
		HttpListenters: []net.Listener{nil},
		RPCMaxClients:  1,
		DisableAuth:    true,
		ChainParams:    activeNetParams,
		BlockChain:     &bc,
		Blockgen:       blockgen,
		TxMemPool:      &txpool,
		Server:         &server,
		Database:       db,
	}
	rpcServer := &rpcserver.RpcServer{}
	rpclocal := &LocalRPCClient{rpcServer}

	btcChain, err := getBTCRelayingChain(activeNetParams.PortalParams.RelayingParam.BTCRelayingHeaderChainID, "btcchain", "/tmp/database")
	if err != nil {
		panic(err)
	}
	bnbChainState, err := getBNBRelayingChainState(activeNetParams.PortalParams.RelayingParam.BNBRelayingHeaderChainID, "/tmp/database")
	if err != nil {
		panic(err)
	}

	txpool.Init(&mempool.Config{
		ConsensusEngine: &cs,
		BlockChain:      &bc,
		DataBase:        db,
		ChainParams:     activeNetParams,
		FeeEstimator:    fees,
		TxLifeTime:      100,
		MaxTx:           1000,
		// DataBaseMempool:   dbmp,
		IsLoadFromMempool: false,
		PersistMempool:    false,
		RelayShards:       nil,
		PubSubManager:     ps,
	})
	// serverObj.blockChain.AddTxPool(serverObj.memPool)
	txpool.InitChannelMempool(cPendingTxs, cRemovedTxs)

	temppool.Init(&mempool.Config{
		BlockChain:    &bc,
		DataBase:      db,
		ChainParams:   activeNetParams,
		FeeEstimator:  fees,
		MaxTx:         1000,
		PubSubManager: ps,
	})
	txpool.IsBlockGenStarted = true
	go temppool.Start(cQuit)
	go txpool.Start(cQuit)

	err = bc.Init(&blockchain.Config{
		BTCChain:        btcChain,
		BNBChainState:   bnbChainState,
		ChainParams:     activeNetParams,
		DataBase:        db,
		MemCache:        memcache.New(),
		BlockGen:        blockgen,
		TxPool:          &txpool,
		TempTxPool:      &temppool,
		Server:          &server,
		Syncker:         sync,
		PubSubManager:   ps,
		FeeEstimator:    make(map[byte]blockchain.FeeEstimator),
		ConsensusEngine: &cs,
		GenesisParams:   blockchain.GenesisParam,
	})
	if err != nil {
		panic(err)
	}
	bc.InitChannelBlockchain(cRemovedTxs)

	sim.param = activeNetParams
	sim.bc = &bc
	sim.consensus = &cs
	sim.txpool = &txpool
	sim.temppool = &temppool
	sim.btcrd = &btcrd
	sim.syncker = sync
	sim.server = &server
	sim.cPendingTxs = cPendingTxs
	sim.cRemovedTxs = cRemovedTxs
	sim.rpcServer = rpcServer
	sim.RPC = rpcclient.NewRPCClient(rpclocal)
	sim.cQuit = cQuit
	sim.listennerRegister = make(map[int][]func(msg interface{}))
	sim.ps = ps
	rpcServer.Init(&rpcConfig)
	go func() {
		for {
			select {
			case <-cQuit:
				return
			case <-cRemovedTxs:
			}
		}
	}()
	go blockgen.Start(cQuit)

	sim.startPubSub()

	//init syncker
	sim.syncker.Init(&syncker.SynckerManagerConfig{Blockchain: sim.bc})

	//init user database
	handles := 256
	cache := 8
	userDBPath := filepath.Join(dbpath, "userdb")
	lvdb, err := leveldb.OpenFile(userDBPath, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if _, corrupted := err.(*lvdbErrors.ErrCorrupted); corrupted {
		lvdb, err = leveldb.RecoverFile(userDBPath, nil)
	}
	sim.userDB = lvdb
	if err != nil {
		panic(errors.Wrapf(err, "levelvdb.OpenFile %s", userDBPath))
	}
}

func (sim *NodeEngine) startPubSub() {
	go sim.ps.Start()
	go func() {
		_, subChan, err := sim.ps.RegisterNewSubscriber(pubsub.BeaconBeststateTopic)
		if err != nil {
			panic("something wrong with subscriber")
		}
		for {
			event := <-subChan
			for _, f := range sim.listennerRegister[BLK_BEACON] {
				f(event.Value)
			}
		}
	}()

	go func() {
		_, subChan, err := sim.ps.RegisterNewSubscriber(pubsub.ShardBeststateTopic)
		if err != nil {
			panic("something wrong with subscriber")
		}
		for {
			event := <-subChan
			for _, f := range sim.listennerRegister[BLK_SHARD] {
				f(event.Value)
			}
		}
	}()
}

func (sim *NodeEngine) StopSync() {
	sim.syncker.Stop()
}

func (sim *NodeEngine) Pause() {
	fmt.Print("Simulation pause! Press Enter to continue ...")
	var input string
	fmt.Scanln(&input)
	fmt.Print("\n")
}

func (sim *NodeEngine) PrintBlockChainInfo() {
	fmt.Println("Beacon Chain:")

	fmt.Println("Shard Chain:")
}

//life cycle of a block generation process:
//PreCreate -> PreValidation -> PreInsert ->
func (sim *NodeEngine) GenerateBlock(args ...interface{}) *NodeEngine {
	time.Sleep(time.Nanosecond)
	var chainArray = []int{-1}
	for i := 0; i < sim.config.ChainParam.ActiveShards; i++ {
		chainArray = append(chainArray, i)
	}
	//beacon
	chain := sim.bc
	var block types.BlockInterface = nil
	var err error

	for _, arg := range args {
		switch arg.(type) {
		case *Execute:
			exec := arg.(*Execute)
			chainArray = exec.appliedChain
		}
	}

	//Create blocks for apply chain
	for _, chainID := range chainArray {
		var proposerPK incognitokey.CommitteePublicKey
		committeeFromBlock := common.Hash{}
		committees := sim.bc.GetChain(chainID).GetBestView().GetCommittee()
		version := 2
		if sim.bc.GetChainParams().StakingFlowV2Height <= sim.bc.GetChain(chainID).GetBestView().GetBeaconHeight() {
			version = 3
		}
		switch version {
		case 2:
			proposerPK, _ = chain.GetChain(chainID).GetBestView().GetProposerByTimeSlot(int64((uint64(sim.timer.Now()) / common.TIMESLOT)), 2)
		case 3:
			proposerPK, _ = chain.GetChain(chainID).GetBestView().GetProposerByTimeSlot(int64((uint64(sim.timer.Now()) / common.TIMESLOT)), 2)
			committeeFromBlock = *chain.BeaconChain.FinalView().GetHash()
			if chainID > -1 {
				committees, _ = sim.bc.GetShardCommitteeFromBeaconHash(committeeFromBlock, byte(chainID))
			}
		}

		proposerPkStr, _ := proposerPK.ToBase58()

		if chainID == -1 {
			block, err = chain.BeaconChain.CreateNewBlock(version, proposerPkStr, 1, sim.timer.Now(), committees, common.Hash{})
			if err != nil {
				Logger.log.Error(err)
				return sim
			}
		} else {
			block, err = chain.ShardChain[byte(chainID)].CreateNewBlock(sim.config.ConsensusVersion, proposerPkStr, 1, sim.timer.Now(), nil, common.Hash{})
			block, err = chain.ShardChain[byte(chainID)].CreateNewBlock(version, proposerPkStr, 1, sim.timer.Now(), committees, committeeFromBlock)
			if err != nil {
				Logger.log.Error(err)
				return sim
			}
		}

		//SignBlock
		proposeAcc := sim.GetAccountByCommitteePubkey(&proposerPK)
		userKey, _ := consensus_v2.GetMiningKeyFromPrivateSeed(proposeAcc.MiningKey)
		sim.SignBlock(userKey, block)

		//Validation
		if chainID == -1 {
			err = chain.VerifyPreSignBeaconBlock(block.(*types.BeaconBlock), true)
			if err != nil {
				Logger.log.Error(err)
				return sim
			}
		} else {
			err = chain.ShardChain[byte(chainID)].ValidatePreSignBlock(block.(*types.ShardBlock), nil)
			if err != nil {
				panic(err)
			}
		}

		//Combine votes
		accs, err := sim.GetListAccountByCommitteePubkey(committees)
		if err != nil {
			panic(err)
		}
		err = sim.SignBlockWithCommittee(block, accs, GenerateCommitteeIndex(len(committees)))
		if err != nil {
			panic(err)
		}

		//Insert
		if chainID == -1 {
			err = chain.InsertBeaconBlock(block.(*types.BeaconBlock), common.FULL_VALIDATION)
			if err != nil {
				panic(err)
			}
			//log.Printf("BEACON | Produced block %v hash %v", block.GetHeight(), block.Hash().String())
		} else {
			err = chain.ShardChain[byte(chainID)].InsertBlock(block.(*types.ShardBlock), common.FULL_VALIDATION)
			if err != nil {
				panic(err)
			} else {
				crossX := blockchain.CreateAllCrossShardBlock(block.(*types.ShardBlock), sim.config.ChainParam.ActiveShards)
				//log.Printf("SHARD %v | Produced block %v hash %v", chainID, block.GetHeight(), block.Hash().String())
				for _, blk := range crossX {
					sim.syncker.InsertCrossShardBlock(blk)
				}
			}
		}
	}

	return sim
}

//number of second we want simulation to forward
//default = round interval
func (sim *NodeEngine) NextRound() {
	sim.timer.Forward(int64(common.TIMESLOT))
}

func (sim *NodeEngine) InjectTx(txBase58 string) error {
	rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58)
	if err != nil {
		return err
	}
	var tx transaction.Tx
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return err
	}
	sim.cPendingTxs <- &tx

	return nil
}

func (sim *NodeEngine) GetBlockchain() *blockchain.BlockChain {
	return sim.bc
}

func (s *NodeEngine) GetUserDatabase() *leveldb.DB {
	return s.userDB
}

func (s *NodeEngine) SignBlockWithCommittee(block types.BlockInterface, committees []account.Account, committeeIndex []int) error {
	committeePubKey := []incognitokey.CommitteePublicKey{}
	miningKeys := []*signatureschemes.MiningKey{}
	if block.GetVersion() >= 2 {
		votes := make(map[string]*blsbftv2.BFTVote)
		for _, committee := range committees {
			miningKey, _ := consensus_v2.GetMiningKeyFromPrivateSeed(committee.MiningKey)
			committeePubKey = append(committeePubKey, *miningKey.GetPublicKey())
			miningKeys = append(miningKeys, miningKey)
		}
		for _, committeeID := range committeeIndex {
			vote, _ := blsbftv2.CreateVote(miningKeys[committeeID], block, committeePubKey)
			vote.IsValid = 1
			votes[vote.Validator] = vote
		}
		committeeBLSString, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(committeePubKey, common.BlsConsensus)
		aggSig, brigSigs, validatorIdx, err := blsbftv2.CombineVotes(votes, committeeBLSString)

		valData, err := blsbftv2.DecodeValidationData(block.GetValidationField())
		if err != nil {
			return errors.New("decode validation data")
		}
		valData.AggSig = aggSig
		valData.BridgeSig = brigSigs
		valData.ValidatiorsIdx = validatorIdx
		validationDataString, _ := blsbftv2.EncodeValidationData(*valData)
		block.AddValidationField(validationDataString)
	}
	return nil
}

func (s *NodeEngine) SignBlock(userMiningKey *signatureschemes.MiningKey, block types.BlockInterface) {
	var validationData blsbftv2.ValidationData
	validationData.ProducerBLSSig, _ = userMiningKey.BriSignData(block.Hash().GetBytes())
	validationDataString, _ := blsbftv2.EncodeValidationData(validationData)
	block.AddValidationField(validationDataString)
}

func (s *NodeEngine) GetAccountByCommitteePubkey(cpk *incognitokey.CommitteePublicKey) *account.Account {
	miningPK := cpk.GetMiningKeyBase58(common.BlsConsensus)
	for _, acc := range s.accounts {
		if acc.MiningPubkey == miningPK {
			return acc
		}
	}
	return nil
}

func (s *NodeEngine) GetListAccountByCommitteePubkey(cpks []incognitokey.CommitteePublicKey) ([]account.Account, error) {
	accounts := []account.Account{}
	for _, cpk := range cpks {
		if acc := s.GetAccountByCommitteePubkey(&cpk); acc != nil {
			accounts = append(accounts, *acc)
		}
	}
	if len(accounts) != len(cpks) {
		return nil, errors.New("Mismatch number of committee pubkey in beststate")
	}
	return accounts, nil
}

func (sim *NodeEngine) GetListAccountsByChainID(chainID int) ([]account.Account, error) {
	committees := sim.bc.GetChain(chainID).GetBestView().GetCommittee()
	return sim.GetListAccountByCommitteePubkey(committees)
}
