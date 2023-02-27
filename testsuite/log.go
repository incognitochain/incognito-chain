package devframework

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/pruner"

	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/incognitochain/incognito-chain/syncker/finishsync"

	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	portalcommonv3 "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	portaltokensv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portaltokens"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	portaltokensv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portaltokens"
	"github.com/incognitochain/incognito-chain/txpool"

	"github.com/incognitochain/incognito-chain/addrmanager"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/connmanager"
	consensus "github.com/incognitochain/incognito-chain/consensus_v2"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/databasemp"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/netsync"
	"github.com/incognitochain/incognito-chain/peer"
	"github.com/incognitochain/incognito-chain/peerv2"
	"github.com/incognitochain/incognito-chain/peerv2/wrapper"

	//privacy "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy"
	relaying "github.com/incognitochain/incognito-chain/relaying/bnb"
	btcRelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/syncker"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/jrick/logrotate/rotator"
)

var (
	// logRotator is one of the logging outputs.  It should be closed on
	// application shutdown.
	logRotator *rotator.Rotator

	backendLog             = common.NewBackend(logWriter{})
	addrManagerLoger       = backendLog.Logger("Address log", true)
	connManagerLogger      = backendLog.Logger("Connection Manager log", true)
	mainLogger             = backendLog.Logger("Server log", false)
	rpcLogger              = backendLog.Logger("RPC log", false)
	rpcServiceLogger       = backendLog.Logger("RPC service log", false)
	rpcServiceBridgeLogger = backendLog.Logger("RPC service DeBridge log", false)
	netsyncLogger          = backendLog.Logger("Netsync log", false)
	peerLogger             = backendLog.Logger("Peer log", true)
	dbLogger               = backendLog.Logger("Database log", false)
	dbmpLogger             = backendLog.Logger("Mempool Persistence DB log", false)
	walletLogger           = backendLog.Logger("Wallet log", false)
	blockchainLogger       = backendLog.Logger("BlockChain log", false)
	consensusLogger        = backendLog.Logger("Consensus log", false)
	mempoolLogger          = backendLog.Logger("Mempool log", false)
	transactionLogger      = backendLog.Logger("Transaction log", false)
	privacyLogger          = backendLog.Logger("Privacy log", false)
	randomLogger           = backendLog.Logger("RandomAPI log", false)
	bridgeLogger           = backendLog.Logger("DeBridge log", false)
	metadataLogger         = backendLog.Logger("Metadata log", false)
	trieLogger             = backendLog.Logger("Trie log", false)
	peerv2Logger           = backendLog.Logger("Peerv2 log", false)
	relayingLogger         = backendLog.Logger("Relaying log", false)
	wrapperLogger          = backendLog.Logger("Wrapper log", false)
	daov2Logger            = backendLog.Logger("DAO log", false)
	btcRelayingLogger      = backendLog.Logger("BTC relaying log", false)
	synckerLogger          = backendLog.Logger("Syncker log ", false)
	privacyV1Logger        = backendLog.Logger("Privacy V1 log ", false)
	privacyV2Logger        = backendLog.Logger("Privacy V2 log ", false)
	instructionLogger      = backendLog.Logger("Instruction log ", false)
	committeeStateLogger   = backendLog.Logger("Committee State log ", false)
	pdexLogger             = backendLog.Logger("Pdex log ", false)
	bridgeAggLogger        = backendLog.Logger("BridgeAgg log ", false)
	finishSyncLogger       = backendLog.Logger("Finish Sync log ", false)
	prunerLogger           = backendLog.Logger("Pruner log ", false)

	portalLogger          = backendLog.Logger("Portal log ", false)
	portalRelayingLogger  = backendLog.Logger("Portal relaying log ", false)
	portalV3CommonLogger  = backendLog.Logger("Portal v3 common log ", false)
	portalV3ProcessLogger = backendLog.Logger("Portal v3 process log ", false)
	portalV3TokenLogger   = backendLog.Logger("Portal v3 token log ", false)

	portalV4ProcessLogger = backendLog.Logger("Portal v4 process log ", false)
	portalV4TokenLogger   = backendLog.Logger("Portal v4 token log ", false)

	txPoolLogger    = backendLog.Logger("Txpool log ", false)
	evmCallerLogger = backendLog.Logger("EVMCaller log ", false)
)

// logWriter implements an io.Writer that outputs to both standard output and
// the write-end pipe of an initialized log rotator.
type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	// os.Stdout.Write(p)
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
	rpcservice.Logger.Init(rpcServiceLogger)
	rpcservice.BLogger.Init(rpcServiceBridgeLogger)
	netsync.Logger.Init(netsyncLogger)
	peer.Logger.Init(peerLogger)
	incdb.Logger.Init(dbLogger)
	wallet.Logger.Init(walletLogger)
	blockchain.Logger.Init(blockchainLogger)
	consensus.Logger.Init(consensusLogger)
	mempool.Logger.Init(mempoolLogger)
	transaction.Logger.Init(transactionLogger)
	//privacy.Logger.Init(privacyLogger)
	databasemp.Logger.Init(dbmpLogger)
	blockchain.BLogger.Init(bridgeLogger)
	rpcserver.BLogger.Init(bridgeLogger)
	metadata.Logger.Init(metadataLogger)
	trie.Logger.Init(trieLogger)
	peerv2.Logger.Init(peerv2Logger)
	relaying.Logger.Init(relayingLogger)
	wrapper.Logger.Init(wrapperLogger)
	dataaccessobject.Logger.Init(daov2Logger)
	btcRelaying.Logger.Init(btcRelayingLogger)
	syncker.Logger.Init(synckerLogger)
	finishsync.Logger.Init(finishSyncLogger)
	privacy.LoggerV1.Init(privacyV1Logger)
	privacy.LoggerV2.Init(privacyV2Logger)
	instruction.Logger.Init(instructionLogger)
	committeestate.Logger.Init(committeeStateLogger)
	pdex.Logger.Init(pdexLogger)
	bridgeagg.Logger.Init(bridgeAggLogger)
	pruner.Logger.Init(prunerLogger)

	portal.Logger.Init(portalLogger)
	portalrelaying.Logger.Init(portalRelayingLogger)
	portalcommonv3.Logger.Init(portalV3CommonLogger)
	portalprocessv3.Logger.Init(portalV3ProcessLogger)
	portaltokensv3.Logger.Init(portalV3TokenLogger)

	portalprocessv4.Logger.Init(portalV4ProcessLogger)
	portaltokensv4.Logger.Init(portalV4TokenLogger)

	txpool.Logger.Init(txPoolLogger)
	evmcaller.Logger.Init(evmCallerLogger)
}

// subsystemLoggers maps each subsystem identifier to its associated logger.
var subsystemLoggers = map[string]common.Logger{
	"MAIN": mainLogger,

	"AMGR":              addrManagerLoger,
	"CMGR":              connManagerLogger,
	"RPCS":              rpcLogger,
	"RPCSservice":       rpcServiceLogger,
	"RPCSbridgeservice": rpcServiceBridgeLogger,
	"NSYN":              netsyncLogger,
	"PEER":              peerLogger,
	"DABA":              dbLogger,
	"WALL":              walletLogger,
	"BLOC":              blockchainLogger,
	"CONS":              consensusLogger,
	"MEMP":              mempoolLogger,
	"RAND":              randomLogger,
	"TRAN":              transactionLogger,
	"PRIV":              privacyLogger,
	"DBMP":              dbmpLogger,
	"DEBR":              bridgeLogger,
	"META":              metadataLogger,
	"TRIE":              trieLogger,
	"PEERV2":            peerv2Logger,
	"DAO":               daov2Logger,
	"BTCRELAYING":       btcRelayingLogger,
	"SYNCKER":           synckerLogger,
	"INST":              instructionLogger,
	"COMS":              committeeStateLogger,
	"PORTAL":            portalLogger,
	"PORTALRELAYING":    portalRelayingLogger,
	"PORTALV3COMMON":    portalV3CommonLogger,
	"PORTALV3PROCESS":   portalV3ProcessLogger,
	"PORTALV3TOKENS":    portalV3TokenLogger,
	"EVMCALLER":         evmCallerLogger,
}

// initLogRotator initializes the logging rotater to write logs to logFile and
// create roll files in the same directory.  It must be called before the
// package-global log rotater variables are used.
func InitLogRotator(logFile string) {
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		os.Exit(-1)
	}
	r, err := rotator.New(logFile, 100*1024*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		os.Exit(-1)
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

// supportedSubsystems returns a sorted slice of the supported subsystems for
// logging purposes.
func supportedSubsystems() []string {
	// Convert the subsystemLoggers map keys to a slice.
	subsystems := make([]string, 0, len(subsystemLoggers))
	for subsysID := range subsystemLoggers {
		subsystems = append(subsystems, subsysID)
	}

	// Sort the subsystems for stable display.
	sort.Strings(subsystems)
	return subsystems
}

// validLogLevel returns whether or not logLevel is a valid debug log level.
func validLogLevel(logLevel string) bool {
	switch logLevel {
	case "trace":
		fallthrough
	case "debug":
		fallthrough
	case "info":
		fallthrough
	case "warn":
		fallthrough
	case "error":
		fallthrough
	case "critical":
		return true
	}
	return false
}

// parseAndSetDebugLevels attempts to parse the specified debug level and set
// the levels accordingly.  An appropriate error is returned if anything is
// invalid.
func parseAndSetDebugLevels(debugLevel string) error {
	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(debugLevel, ",") && !strings.Contains(debugLevel, "=") {
		// ValidateTransaction debug log level.
		if !validLogLevel(debugLevel) {
			str := "the specified debug level [%v] is invalid"
			return fmt.Errorf(str, debugLevel)
		}

		// Change the logging level for all subsystems.
		setLogLevels(debugLevel)

		return nil
	}

	// Split the specified string into subsystem/level pairs while detecting
	// issues and update the log levels accordingly.
	for _, logLevelPair := range strings.Split(debugLevel, ",") {
		if !strings.Contains(logLevelPair, "=") {
			str := "the specified debug level contains an invalid subsystem/level pair [%v]"
			return fmt.Errorf(str, logLevelPair)
		}

		// Extract the specified subsystem and log level.
		fields := strings.Split(logLevelPair, "=")
		subsysID, logLevel := fields[0], fields[1]

		// ValidateTransaction subsystem.
		if _, exists := subsystemLoggers[subsysID]; !exists {
			str := "the specified subsystem [%v] is invalid -- supported subsytems %v"
			return fmt.Errorf(str, subsysID, supportedSubsystems())
		}

		// ValidateTransaction log level.
		if !validLogLevel(logLevel) {
			str := "the specified debug level [%v] is invalid"
			return fmt.Errorf(str, logLevel)
		}

		setLogLevel(subsysID, logLevel)
	}
	return nil
}
