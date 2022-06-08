package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/incognitochain/incognito-chain/pruner"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/metrics/monitor"
	"github.com/incognitochain/incognito-chain/portal"
	bnbrelaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/incognitochain/incognito-chain/utils"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/databasemp"
	_ "github.com/incognitochain/incognito-chain/databasemp/lvdb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/limits"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/wallet"
)

//go:generate mockery -dir=incdb/ -name=Database

// winServiceMain is only invoked on Windows.  It detects when incognito network is running
// as a service and reacts accordingly.
var winServiceMain func() (bool, error)

func getBTCRelayingChain(btcRelayingChainID, btcDataFolderName string) (*btcrelaying.BlockChain, error) {
	relayingChainParams := map[string]*chaincfg.Params{
		portal.TestnetBTCChainID:  btcrelaying.GetTestNet3Params(),
		portal.Testnet2BTCChainID: btcrelaying.GetTestNet3ParamsForInc2(),
		portal.MainnetBTCChainID:  btcrelaying.GetMainNetParams(),
	}
	relayingChainGenesisBlkHeight := map[string]int32{
		portal.TestnetBTCChainID:  int32(2063133),
		portal.Testnet2BTCChainID: int32(2064989),
		portal.MainnetBTCChainID:  int32(697298),
	}
	return btcrelaying.GetChainV2(
		filepath.Join(config.Config().DataDir, btcDataFolderName),
		relayingChainParams[btcRelayingChainID],
		relayingChainGenesisBlkHeight[btcRelayingChainID],
	)
}

func getBNBRelayingChainState(bnbRelayingChainID string) (*bnbrelaying.BNBChainState, error) {
	bnbChainState := new(bnbrelaying.BNBChainState)
	err := bnbChainState.LoadBNBChainState(
		filepath.Join(config.Config().DataDir, "bnbrelayingv3"),
		bnbRelayingChainID,
	)
	if err != nil {
		log.Printf("Error getBNBRelayingChainState: %v\n", err)
		return nil, err
	}
	return bnbChainState, nil
}

// mainMaster is the real main function for Incognito network.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func mainMaster(serverChan chan<- *Server) error {
	//read basic config from file or flag
	cfg := config.LoadConfig()

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	initLogRotator(cfg.LogFileName)

	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(cfg.LogLevel); err != nil {
		fmt.Fprintln(os.Stderr, err)
		panic(err)
	}
	config.LoadParam()

	portal.SetupParam()
	err := wallet.InitPublicKeyBurningAddressByte()
	if err != nil {
		Logger.log.Error(err)
		panic(err)
	}

	//create genesis block
	blockchain.CreateGenesisBlocks()

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer Logger.log.Warn("Shutdown complete")
	// Show version at startup.
	version := version()
	Logger.log.Infof("Version %s", version)
	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}
	db, err := incdb.OpenMultipleDB("leveldb", filepath.Join(cfg.DataDir, cfg.DatabaseDir))
	// Create db and use it.
	if err != nil {
		Logger.log.Error("could not open connection to leveldb")
		Logger.log.Error(err)
		panic(err)
	}
	//check if prune flag is available
	if config.Config().StatePrune {
		if err := pruner.NewPrunerWithValue(db).Prune(); err != nil {
			panic(err)
		}
	}

	// Create db for mempool and use it
	consensusDB, err := incdb.Open("leveldb", filepath.Join(cfg.DataDir, "consensus"))
	if err != nil {
		Logger.log.Error("could not open connection to leveldb")
		Logger.log.Error(err)
		panic(err)
	}
	rawdb_consensus.SetConsensusDatabase(consensusDB)
	// Create db for mempool and use it
	dbmp, err := databasemp.Open("leveldbmempool", filepath.Join(cfg.DataDir, cfg.MempoolDir))
	if err != nil {
		Logger.log.Error("could not open connection to leveldb")
		Logger.log.Error(err)
		panic(err)
	}
	// Check wallet and start it
	var walletObj *wallet.Wallet
	if cfg.EnableWallet {
		walletObj = &wallet.Wallet{}
		walletConf := wallet.WalletConfig{
			DataDir:        cfg.DataDir,
			DataFile:       cfg.WalletName,
			DataPath:       filepath.Join(cfg.DataDir, cfg.WalletName),
			IncrementalFee: 0, // 0 mili PRV
		}
		if cfg.WalletShardID >= 0 {
			// check shardID of wallet
			temp := byte(cfg.WalletShardID)
			walletConf.ShardID = &temp
		}
		walletObj.SetConfig(&walletConf)
		err = walletObj.LoadWallet(cfg.WalletPassphrase)
		if err != nil {
			if cfg.WalletAutoInit {
				Logger.log.Critical("\n **** Auto init wallet flag is TRUE ****\n")
				walletObj.Init(cfg.WalletPassphrase, 0, cfg.WalletName)
				walletObj.Save(cfg.WalletPassphrase)
			} else {
				// write log and exit when can not load wallet
				Logger.log.Criticalf("Can not load wallet with %s. Please use incognitoctl to create a new wallet", walletObj.GetConfig().DataPath)
				return err
			}
		}
	}
	// Create btcrelaying chain
	btcChain, err := getBTCRelayingChain(
		portal.GetPortalParams().RelayingParam.BTCRelayingHeaderChainID,
		portal.GetPortalParams().RelayingParam.BTCDataFolderName,
	)
	if err != nil {
		Logger.log.Error("could not get or create btc relaying chain")
		Logger.log.Error(err)
		panic(err)
	}
	defer func() {
		Logger.log.Warn("Gracefully shutting down the btc database...")
		db := btcChain.GetDB()
		db.Close()
	}()

	// Create bnbrelaying chain state
	bnbChainState, err := getBNBRelayingChainState(portal.GetPortalParams().RelayingParam.BNBRelayingHeaderChainID)
	if err != nil {
		Logger.log.Error("could not get or create bnb relaying chain state")
		Logger.log.Error(err)
		panic(err)
	}

	useOutcoinDb := len(cfg.UseOutcoinDatabase) >= 1
	var outcoinDb *incdb.Database = nil
	if useOutcoinDb {
		temp, err := incdb.Open("leveldb", filepath.Join(cfg.DataDir, cfg.OutcoinDatabaseDir))
		if err != nil {
			Logger.log.Error("could not open leveldb instance for coin storing")
		}
		outcoinDb = &temp
	}

	// Create server and start it.
	server := Server{}
	server.wallet = walletObj
	err = server.NewServer(cfg.Listener, db, dbmp, outcoinDb, cfg.NumIndexerWorkers, cfg.IndexerAccessTokens, version, btcChain, bnbChainState, interrupt)
	if err != nil {
		Logger.log.Errorf("Unable to start server on %+v", cfg.Listener)
		Logger.log.Error(err)
		return err
	}

	// Init EVM caller cacher
	evmcaller.InitCacher()

	defer func() {
		Logger.log.Warn("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
		Logger.log.Warn("Server shutdown complete")
	}()
	server.Start()
	if serverChan != nil {
		serverChan <- &server
	}

	// Check Metric analyzation system
	env := os.Getenv("GrafanaURL")
	if env != "" {
		Logger.log.Criticalf("Metric Server: %+v", os.Getenv("GrafanaURL"))
	}
	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil

}
func main() {
	limitThreads := os.Getenv("CPU")
	if limitThreads == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
		monitor.SetGlobalParam("CPU", runtime.NumCPU())
	} else {
		numThreads, err := strconv.Atoi(limitThreads)
		if err != nil {
			panic(err)
		}
		runtime.GOMAXPROCS(numThreads)
		monitor.SetGlobalParam("CPU", numThreads)
	}
	fmt.Println("NumCPU", runtime.NumCPU())
	// Block and transaction processing can cause bursty allocations.  This
	// limits the garbage collector from excessively overallocating during
	// bursts.  This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(100)
	if os.Getenv("Profiling") != "" {
		go http.ListenAndServe(":"+os.Getenv("Profiling"), nil)
	}

	// Up some limits.
	if err := limits.SetLimits(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set limits: %+v\n", err)
		os.Exit(utils.ExitByOs)
	}
	// Call serviceMain on Windows to handle running as a service.  When
	// the return isService flag is true, exit now since we ran as a
	// service.  Otherwise, just fall through to normal operation.
	if runtime.GOOS == "windows" {
		isService, err := winServiceMain()
		if err != nil {
			fmt.Println(err)
			os.Exit(utils.ExitByOs)
		}
		if isService {
			os.Exit(utils.ExitCodeUnknow)
		}
	}
	// Work around defer not working after os.Exit()
	if err := mainMaster(nil); err != nil {
		os.Exit(utils.ExitByOs)
	}
}
