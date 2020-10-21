package devframework

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/simulation/mock"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	ShardNumber   int
	RoundInterval int
}

type Hook struct {
	Create     func(chainID int, chain interface{}, doCreate func() (blk interface{}, err error))
	Validation func(chainID int, chain interface{}, block interface{}, doValidation func(blk interface{}) error)
	Insert     func(chainID int, chain interface{}, block interface{}, doInsert func(blk interface{}) error)
}
type SimulationEngine struct {
	config      Config
	simName     string
	param       *blockchain.Params
	bc          *blockchain.BlockChain
	cs          *mock.Consensus
	txpool      *mempool.TxPool
	temppool    *mempool.TxPool
	btcrd       *mock.BTCRandom
	sync        *mock.Syncker
	server      *mock.Server
	cPendingTxs chan metadata.Transaction
	cRemovedTxs chan metadata.Transaction
	rpcServer   *rpcserver.RpcServer
	cQuit       chan struct{}
}

func NewStandaloneSimulation(name string, config Config) *SimulationEngine {
	sim := &SimulationEngine{
		config:  config,
		simName: name,
	}
	sim.init()
	time.Sleep(1 * time.Second)
	return sim
}

func (sim *SimulationEngine) init() {
	simName := sim.simName
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	initLogRotator(filepath.Join(path, "sim.log"))
	dbLogger.SetLevel(common.LevelTrace)
	blockchainLogger.SetLevel(common.LevelTrace)
	bridgeLogger.SetLevel(common.LevelTrace)
	rpcLogger.SetLevel(common.LevelTrace)
	rpcServiceLogger.SetLevel(common.LevelTrace)
	rpcServiceBridgeLogger.SetLevel(common.LevelTrace)
	transactionLogger.SetLevel(common.LevelTrace)
	privacyLogger.SetLevel(common.LevelTrace)
	mempoolLogger.SetLevel(common.LevelTrace)
	activeNetParams := &blockchain.ChainTest2Param
	cs := mock.Consensus{}
	txpool := mempool.TxPool{}
	temppool := mempool.TxPool{}
	btcrd := mock.BTCRandom{} // use mock for now
	sync := mock.Syncker{}
	server := mock.Server{}
	ps := mock.Pubsub{}
	fees := make(map[byte]*mempool.FeeEstimator)
	bc := blockchain.BlockChain{}
	for i := byte(0); i < byte(activeNetParams.ActiveShards); i++ {
		fees[i] = mempool.NewFeeEstimator(
			mempool.DefaultEstimateFeeMaxRollback,
			mempool.DefaultEstimateFeeMinRegisteredBlocks,
			0)
	}
	cPendingTxs := make(chan metadata.Transaction, 500)
	cRemovedTxs := make(chan metadata.Transaction, 500)
	cQuit := make(chan struct{})
	blockgen, err := blockchain.NewBlockGenerator(&txpool, &bc, &sync, cPendingTxs, cRemovedTxs)
	if err != nil {
		panic(err)
	}

	listenFunc := net.Listen
	listener, err := listenFunc("tcp", "0.0.0.0:8000")
	if err != nil {
		panic(err)
	}
	rpcConfig := rpcserver.RpcServerConfig{
		HttpListenters: []net.Listener{listener},
		RPCMaxClients:  1,
		DisableAuth:    true,
		ChainParams:    activeNetParams,
		BlockChain:     &bc,
		Blockgen:       blockgen,
		TxMemPool:      &txpool,
	}
	rpcServer := &rpcserver.RpcServer{}

	db, err := incdb.OpenMultipleDB("leveldb", filepath.Join("./"+simName, "database"))
	// Create db and use it.
	if err != nil {
		panic(err)
	}

	btcChain, err := getBTCRelayingChain(activeNetParams.BTCRelayingHeaderChainID, "btcchain", simName)
	if err != nil {
		panic(err)
	}
	bnbChainState, err := getBNBRelayingChainState(activeNetParams.BNBRelayingHeaderChainID, simName)
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
		PubSubManager:     &ps,
	})
	// serverObj.blockChain.AddTxPool(serverObj.memPool)
	txpool.InitChannelMempool(cPendingTxs, cRemovedTxs)

	temppool.Init(&mempool.Config{
		BlockChain:    &bc,
		DataBase:      db,
		ChainParams:   activeNetParams,
		FeeEstimator:  fees,
		MaxTx:         1000,
		PubSubManager: &ps,
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
		Syncker:         &sync,
		PubSubManager:   &ps,
		FeeEstimator:    make(map[byte]blockchain.FeeEstimator),
		RandomClient:    &btcrd,
		ConsensusEngine: &cs,
		GenesisParams:   blockchain.GenesisParam,
	})
	if err != nil {
		panic(err)
	}
	bc.InitChannelBlockchain(cRemovedTxs)

	sim.param = activeNetParams
	sim.bc = &bc
	sim.cs = &cs
	sim.txpool = &txpool
	sim.temppool = &temppool
	sim.btcrd = &btcrd
	sim.sync = &sync
	sim.server = &server
	sim.cPendingTxs = cPendingTxs
	sim.cRemovedTxs = cRemovedTxs
	sim.rpcServer = rpcServer
	sim.cQuit = cQuit

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
	go rpcServer.Start()

}

func getBTCRelayingChain(btcRelayingChainID string, btcDataFolderName string, dataFolder string) (*btcrelaying.BlockChain, error) {
	relayingChainParams := map[string]*chaincfg.Params{
		blockchain.TestnetBTCChainID:  btcrelaying.GetTestNet3Params(),
		blockchain.Testnet2BTCChainID: btcrelaying.GetTestNet3ParamsForInc2(),
		blockchain.MainnetBTCChainID:  btcrelaying.GetMainNetParams(),
	}
	relayingChainGenesisBlkHeight := map[string]int32{
		blockchain.TestnetBTCChainID:  int32(1833130),
		blockchain.Testnet2BTCChainID: int32(1833130),
		blockchain.MainnetBTCChainID:  int32(634140),
	}
	return btcrelaying.GetChainV2(
		filepath.Join("./"+dataFolder, btcDataFolderName),
		relayingChainParams[btcRelayingChainID],
		relayingChainGenesisBlkHeight[btcRelayingChainID],
	)
}

func getBNBRelayingChainState(bnbRelayingChainID string, dataFolder string) (*bnbrelaying.BNBChainState, error) {
	bnbChainState := new(bnbrelaying.BNBChainState)
	err := bnbChainState.LoadBNBChainState(
		filepath.Join("./"+dataFolder, "bnbrelayingv3"),
		bnbRelayingChainID,
	)
	if err != nil {
		log.Printf("Error getBNBRelayingChainState: %v\n", err)
		return nil, err
	}
	return bnbChainState, nil
}

func (sim *SimulationEngine) Pause() {
	fmt.Print("Simulation pause! Press Enter to continue ...")
	var input string
	fmt.Scanln(&input)
}

//Auto generate block
func (sim *SimulationEngine) AutoGenerateBlock(numBlk int) {
	for i := 0; i < numBlk; i++ {
		sim.GenerateBlock(nil)
	}
}

//life cycle of a block generation process:
//PreCreate -> PreValidation -> PreInsert ->
func (sim *SimulationEngine) GenerateBlock(h *Hook) {
	//beacon
	chain := sim.bc

	block, err := chain.NewBlockBeacon(chain.GetBeaconBestState(), 2, "x", 1, time.Now().Unix())
	if err != nil {
		fmt.Println("NewBlockError", err)
		goto PASS_BEACON
	}

	err = chain.VerifyPreSignBeaconBlock(block, true)
	if err != nil {
		fmt.Println("VerifyBlockErr", err)
		goto PASS_BEACON
	}

	err = chain.InsertBeaconBlock(block, true)
	if err != nil {
		fmt.Println("InsertBlkErr", err)
		goto PASS_BEACON
	}

PASS_BEACON:

	////shard
	//for i := 0; i < sim.config.ShardNumber; i++ {
	//	sim.bc
	//	block, err := chain.NewBlockShard(chain.GetBestStateShard(byte(i)), 2, "x", 1, time.Now())
	//	if err != nil {
	//		goto PASS_SHARD
	//	}
	//
	//	err = chain.VerifyPreSignShardBlock(block, byte(i))
	//	if err != nil {
	//		goto PASS_SHARD
	//	}
	//
	//	err = chain.InsertShardBlock(block, true)
	//	if err != nil {
	//		goto PASS_SHARD
	//	}
	//PASS_SHARD:
	//}

	sim.ForwardToFuture()
}

//number of second we want simulation to forward
//default = round interval
func (sim *SimulationEngine) ForwardToFuture(args ...interface{}) {

}
