package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/constant-money/constant-chain/addrmanager"
	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/blockchain/btc/btcapi"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/connmanager"
	"github.com/constant-money/constant-chain/consensus/constantbft"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/mempool"
	"github.com/constant-money/constant-chain/netsync"
	"github.com/constant-money/constant-chain/peer"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/rpcserver"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/jrick/logrotate/rotator"
)

var (
	// logRotator is one of the logging outputs.  It should be closed on
	// application shutdown.
	logRotator *rotator.Rotator

	backendLog        = common.NewBackend(logWriter{})
	addrManagerLoger  = backendLog.Logger("Address Log", false)
	connManagerLogger = backendLog.Logger("Connection Manager Log", false)
	mainLogger        = backendLog.Logger("Server Log", false)
	rpcLogger         = backendLog.Logger("RPC Log", false)
	netsyncLogger     = backendLog.Logger("Netsync Log", false)
	peerLogger        = backendLog.Logger("Peer Log", false)
	dbLogger          = backendLog.Logger("Database Log", false)
	walletLogger      = backendLog.Logger("Wallet log", false)
	blockchainLogger  = backendLog.Logger("BlockChain log", false)
	consensusLogger   = backendLog.Logger("Consensus log", false)
	mempoolLogger     = backendLog.Logger("Mempool log", false)
	transactionLogger = backendLog.Logger("Transaction log", false)
	privacyLogger     = backendLog.Logger("Privacy log", false)
	randomLogger      = backendLog.Logger("RandomAPI log", false)
)

// logWriter implements an io.Writer that outputs to both standard output and
// the write-end pipe of an initialized log rotator.
type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	logRotator.Write(p)
	return len(p), nil
}

func init() {
	// for main thread
	Logger.Init(mainLogger)

	// for other components
	connmanager.Logger.Init(connManagerLogger)
	addrmanager.Logger.Init(addrManagerLoger)
	rpcserver.Logger.Init(rpcLogger)
	netsync.Logger.Init(netsyncLogger)
	peer.Logger.Init(peerLogger)
	database.Logger.Init(dbLogger)
	wallet.Logger.Init(walletLogger)
	blockchain.Logger.Init(blockchainLogger)
	constantbft.Logger.Init(consensusLogger)
	mempool.Logger.Init(mempoolLogger)
	btcapi.Logger.Init(randomLogger)
	transaction.Logger.Init(transactionLogger)
	privacy.Logger.Init(privacyLogger)

}

// subsystemLoggers maps each subsystem identifier to its associated logger.
var subsystemLoggers = map[string]common.Logger{
	"MAIN": mainLogger,

	"AMGR": addrManagerLoger,
	"CMGR": connManagerLogger,
	"RPCS": rpcLogger,
	"NSYN": netsyncLogger,
	"PEER": peerLogger,
	"DABA": dbLogger,
	"WALL": walletLogger,
	"BLOC": blockchainLogger,
	"CONS": consensusLogger,
	"MEMP": mempoolLogger,
	"RAND": randomLogger,
	"TRAN": transactionLogger,
	"PRIV": privacyLogger,
}

// initLogRotator initializes the logging rotater to write logs to logFile and
// create roll files in the same directory.  It must be called before the
// package-global log rotater variables are used.
func initLogRotator(logFile string) {
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		os.Exit(1)
	}
	r, err := rotator.New(logFile, 10*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		os.Exit(1)
	}

	logRotator = r
}

// setLogLevel sets the logging level for provided subsystem.  Invalid
// subsystems are ignored.  Uninitialized subsystems are dynamically created as
// needed.
func setLogLevel(subsystemID string, logLevel string) {
	// Ignore invalid subsystems.
	logger, ok := subsystemLoggers[subsystemID]
	if !ok {
		return
	}

	// Defaults to info if the log level is invalid.
	level, _ := common.LevelFromString(logLevel)
	logger.SetLevel(level)
}

// setLogLevels sets the log level for all subsystem loggers to the passed
// level.  It also dynamically creates the subsystem loggers as needed, so it
// can be used to initialize the logging system.
func setLogLevels(logLevel string) {
	// Configure all sub-systems with the new logging level.  Dynamically
	// create loggers as needed.
	for subsystemID := range subsystemLoggers {
		setLogLevel(subsystemID, logLevel)
	}
}

type MainLogger struct {
	log common.Logger
}

func (mainLogger *MainLogger) Init(inst common.Logger) {
	mainLogger.log = inst
}

// Global instant to use
var Logger = MainLogger{}
