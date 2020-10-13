package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/jrick/logrotate/rotator"
)

var (
	// logRotator is one of the logging outputs.  It should be closed on
	// application shutdown.
	logRotator *rotator.Rotator

	backendLog             = common.NewBackend(logWriter{})
	dbLogger               = backendLog.Logger("Database log", false)
	blockchainLogger       = backendLog.Logger("BlockChain log", false)
	bridgeLogger           = backendLog.Logger("DeBridge log", false)
	rpcLogger              = backendLog.Logger("RPC log", false)
	rpcServiceLogger       = backendLog.Logger("RPC service log", false)
	rpcServiceBridgeLogger = backendLog.Logger("RPC service DeBridge log", false)
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
	incdb.Logger.Init(dbLogger)
	blockchain.Logger.Init(blockchainLogger)
	blockchain.BLogger.Init(bridgeLogger)
	rpcserver.Logger.Init(rpcLogger)
	rpcservice.Logger.Init(rpcServiceLogger)
	rpcservice.BLogger.Init(rpcServiceBridgeLogger)
}

// subsystemLoggers maps each subsystem identifier to its associated logger.
var subsystemLoggers = map[string]common.Logger{
	"DABA":              dbLogger,
	"BLOC":              blockchainLogger,
	"DEBR":              bridgeLogger,
	"RPCS":              rpcLogger,
	"RPCSservice":       rpcServiceLogger,
	"RPCSbridgeservice": rpcServiceBridgeLogger,
}

// initLogRotator initializes the logging rotater to write logs to logFile and
// create roll files in the same directory.  It must be called before the
// package-global log rotater variables are used.
func initLogRotator(logFile string) {
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		os.Exit(common.ExitByLogging)
	}
	r, err := rotator.New(logFile, 10*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		os.Exit(common.ExitByLogging)
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
