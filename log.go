package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jrick/logrotate/rotator"
	"github.com/ninjadotorg/constant/addrmanager"
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/connmanager"
	"github.com/ninjadotorg/constant/consensus/ppos"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/mempool"
	"github.com/ninjadotorg/constant/netsync"
	"github.com/ninjadotorg/constant/peer"
	"github.com/ninjadotorg/constant/rpcserver"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/privacy"
)

var (
	// logRotator is one of the logging outputs.  It should be closed on
	// application shutdown.
	logRotator *rotator.Rotator

	backendLog        = common.NewBackend(logWriter{})
	addrManagerLoger  = backendLog.Logger("Address Log")
	connManagerLogger = backendLog.Logger("Connection Manager Log")
	mainLogger        = backendLog.Logger("Server Log")
	rpcLogger         = backendLog.Logger("RPC Log")
	netsyncLogger     = backendLog.Logger("Netsync Log")
	peerLogger        = backendLog.Logger("Peer Log")
	dbLogger          = backendLog.Logger("Database Log")
	walletLogger      = backendLog.Logger("Wallet log")
	blockchainLogger  = backendLog.Logger("blockChain log")
	consensusLogger   = backendLog.Logger("Consensus log")
	mempoolLogger     = backendLog.Logger("Mempool log")
	transactionLogger = backendLog.Logger("Transaction log")
	privacyLogger     = backendLog.Logger("Privacy log")
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
	ppos.Logger.Init(consensusLogger)
	mempool.Logger.Init(mempoolLogger)
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

// directionString is a helper function that returns a string that represents
// the direction of a connection (inbound or outbound).
func directionString(inbound bool) string {
	if inbound {
		return "inbound"
	}
	return "outbound"
}

// pickNoun returns the singular or plural form of a noun depending
// on the count n.
func pickNoun(n uint64, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

type MainLogger struct {
	log common.Logger
}

func (self *MainLogger) Init(inst common.Logger) {
	self.log = inst
}

// Global instant to use
var Logger = MainLogger{}
