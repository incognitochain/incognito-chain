package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/davecgh/go-spew/spew"
	"github.com/jessevdk/go-flags"
)

// default config
const (
	defaultConfigFilename         = "config.conf"
	defaultDataDirname            = "data"
	defaultDatabaseDirname        = "block"
	defaultDatabaseMempoolDirname = "mempool"
	defaultLogLevel               = "info"
	defaultLogDirname             = "logs"
	defaultLogFilename            = "log.log"
	defaultMaxPeers               = 125
	defaultMaxPeersSameShard      = 10
	defaultMaxPeersOtherShard     = 10
	defaultMaxPeersOther          = 125
	defaultMaxPeersNoShard        = 125
	defaultMaxPeersBeacon         = 20
	defaultMaxRPCClients          = 10
	sampleConfigFilename          = "sample-config.conf"
	defaultDisableRpcTLS          = true
	defaultFastStartup            = true
	defaultNodeMode               = common.NODEMODE_RELAY
	defaultTxPoolTTL              = uint(86400) * 10 // in second
	defaultTxPoolMaxTx            = uint64(20000)
	// For wallet
	defaultWalletName     = "wallet"
	defaultPersistMempool = false
)

var (
	defaultHomeDir     = common.AppDataDir("constant", false)
	defaultConfigFile  = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultDataDir     = filepath.Join(defaultHomeDir, defaultDataDirname)
	defaultRPCKeyFile  = filepath.Join(defaultHomeDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(defaultHomeDir, "rpc.cert")
	defaultLogDir      = filepath.Join(defaultHomeDir, defaultLogDirname)
)

// runServiceCommand is only set to a real function on Windows.  It is used
// to parse and execute service commands specified via the -s flag.
var runServiceCommand func(string) error

// See loadConfig for details on the configuration load process.
type config struct {
	Nodename           string `short:"n" long:"name" description:"Node name"`
	ShowVersion        bool   `short:"V" long:"version" description:"Display version information and exit"`
	ConfigFile         string `short:"C" long:"configfile" description:"Path to configuration file"`
	DataDir            string `short:"D" long:"datadir" description:"Directory to store data"`
	DatabaseDir        string `short:"d" long:"datapre" description:"Database dir"`
	DatabaseMempoolDir string `short:"m" long:"datamempool" description:"Mempool Database Dir"`
	LogDir             string `short:"l" long:"logdir" description:"Directory to log output."`
	LogLevel           string `long:"loglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`

	AddPeers             []string `short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	ConnectPeers         []string `short:"c" long:"connect" description:"Connect only to the specified peers at startup"`
	DisableListen        bool     `long:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen"`
	Listener             string   `long:"listen" description:"Add an interface/port to listen for connections (default all interfaces port: 9333, testnet: 9444)"`
	MaxPeers             int      `long:"maxpeers" description:"Max number of inbound and outbound peers"`
	MaxOutPeers          int      `long:"maxoutpeers" description:"Max number of outbound peers"`
	MaxInPeers           int      `long:"maxinpeers" description:"Max number of inbound peers"`
	DiscoverPeers        bool     `long:"discoverpeers" description:"Enable discover peers"`
	DiscoverPeersAddress string   `long:"discoverpeersaddress" description:"Url to connect discover peers server"`
	MaxPeersSameShard    int      `long:"maxpeersameshard" description:"Max peers in same shard for connection"`
	MaxPeersOtherShard   int      `long:"maxpeerothershard" description:"Max peers in other shard for connection"`
	MaxPeersOther        int      `long:"maxpeerother" description:"Max peers in other for connection"`
	MaxPeersNoShard      int      `long:"maxpeernoshard" description:"Max peers in no shard for connection"`
	MaxPeersBeacon       int      `long:"maxpeerbeacon" description:"Max peers in beacon for connection"`

	ExternalAddress string `long:"externaladdress" description:"External address"`

	RPCDisableAuth bool     `long:"norpcauth" description:"Disable RPC authorization by username/password"`
	RPCUser        string   `short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass        string   `short:"P" long:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCLimitUser   string   `long:"rpclimituser" description:"Username for limited RPC connections"`
	RPCLimitPass   string   `long:"rpclimitpass" default-mask:"-" description:"Password for limited RPC connections"`
	RPCListeners   []string `long:"rpclisten" description:"Add an interface/port to listen for RPC connections (default port: 9334, testnet: 9334)"`
	RPCCert        string   `long:"rpccert" description:"File containing the certificate file"`
	RPCKey         string   `long:"rpckey" description:"File containing the certificate key"`
	RPCMaxClients  int      `long:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	RPCQuirks      bool     `long:"rpcquirks" description:"Mirror some JSON-RPC quirks of coin Core -- NOTE: Discouraged unless interoperability issues need to be worked around"`
	DisableRPC     bool     `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass or rpclimituser/rpclimitpass is specified"`
	DisableTLS     bool     `long:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`

	Proxy     string `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser string `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass string `long:"proxypass" default-mask:"-" description:"Password for proxy server"`
	// Generate  bool   `long:"generate" description:"Generate (mine) coins using the CPU"`

	// Net config
	TestNet bool `long:"testnet" description:"Use the test network"`

	PrivateKey  string `long:"privatekey" description:"User spending key used for operation in consensus"`
	NodeMode    string `long:"nodemode" description:"Role of this node (beacon/shard/wallet/relay | default role is 'relay' (relayshards must be set to run), 'auto' mode will switch between 'beacon' and 'shard')"`
	RelayShards string `long:"relayshards" description:"set relay shards of this node when in 'relay' mode if noderole is auto then it only sync shard data when user is a shard producer/validator"`
	// For Wallet
	Wallet           bool   `long:"enablewallet" description:"Enable wallet"`
	WalletName       string `long:"wallet" description:"Wallet Database Name file, default is 'wallet'"`
	WalletPassphrase string `long:"walletpassphrase" description:"Wallet passphrase"`
	WalletAutoInit   bool   `long:"walletautoinit" description:"Init wallet automatically if not exist"`

	FastStartup bool `long:"faststartup" description:"Load existed shard/chain dependencies instead of rebuild from block data"`

	TxPoolTTL   uint   `long:"txpoolttl" description:"Set Time To Live (TTL) Value for transaction that enter pool"`
	TxPoolMaxTx uint64 `long:"txpoolmaxtx" description:"Set Maximum number of transaction in pool"`

	LoadMempool    bool `long:"loadmempool" description:"Load transactions from Mempool database"`
	PersistMempool bool `long:"persistmempool" description:"Persistence transaction in memepool database"`
}

// serviceOptions defines the configuration options for the daemon as a service on
// Windows.
type serviceOptions struct {
	ServiceCommand string `short:"s" long:"service" description:"Service command {install, remove, start, stop}"`
}

// filesExists reports whether the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *config, so *serviceOptions, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	if runtime.GOOS == "windows" {
		parser.AddGroup("Service Options", "Service Options", so)
	}
	return parser
}

// createDefaultConfig copies the file sample-sps.conf to the given destination path,
// and populates it with some randomly generated RPC username and password.
func createDefaultConfigFile(destinationPath string) error {
	// Create the destination directory if it does not exists
	err := os.MkdirAll(filepath.Dir(destinationPath), 0700)
	if err != nil {
		return err
	}

	// We assume sample config file path is same as binary
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	sampleConfigPath := filepath.Join(path, sampleConfigFilename)

	// We generate a random user and password
	randomBytes := make([]byte, 20)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return err
	}
	generatedRPCUser := base64.StdEncoding.EncodeToString(randomBytes)

	_, err = rand.Read(randomBytes)
	if err != nil {
		return err
	}
	generatedRPCPass := base64.StdEncoding.EncodeToString(randomBytes)

	src, err := os.Open(sampleConfigPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.OpenFile(destinationPath,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer dest.Close()

	// We copy every line from the sample config file to the destination,
	// only replacing the two lines for rpcuser and rpcpass
	reader := bufio.NewReader(src)
	for err != io.EOF {
		var line string
		line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		if strings.Contains(line, "rpcuser=") {
			line = "rpcuser=" + generatedRPCUser + "\n"
		} else if strings.Contains(line, "rpcpass=") {
			line = "rpcpass=" + generatedRPCPass + "\n"
		}

		if _, err := dest.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

// normalizeAddresses returns a new slice with all the passed peer addresses
// normalized with the given default port, and all duplicates removed.
func normalizeAddresses(addrs []string, defaultPort string) []string {
	for i, addr := range addrs {
		addrs[i] = normalizeAddress(addr, defaultPort)
	}

	return removeDuplicateAddresses(addrs)
}

// normalizeAddress returns addr with the passed default port appended if
// there is not already a port specified.
func normalizeAddress(addr, defaultPort string) string {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.JoinHostPort(addr, defaultPort)
	}
	return addr
}

// removeDuplicateAddresses returns a new slice with all duplicate entries in
// addrs removed.
func removeDuplicateAddresses(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

/*
// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
//
// The above results in Constant functioning properly without any config settings
// while still allowing the user to override settings with config files and
// command line options.  Command line options always take precedence.
*/
func loadConfig() (*config, []string, error) {
	cfg := config{
		ConfigFile:         defaultConfigFile,
		LogLevel:           defaultLogLevel,
		MaxOutPeers:        defaultMaxPeers,
		MaxInPeers:         defaultMaxPeers,
		MaxPeers:           defaultMaxPeers,
		MaxPeersSameShard:  defaultMaxPeersSameShard,
		MaxPeersOtherShard: defaultMaxPeersOtherShard,
		MaxPeersOther:      defaultMaxPeersOther,
		MaxPeersNoShard:    defaultMaxPeersNoShard,
		MaxPeersBeacon:     defaultMaxPeersBeacon,
		RPCMaxClients:      defaultMaxRPCClients,
		DataDir:            defaultDataDir,
		DatabaseDir:        defaultDatabaseDirname,
		DatabaseMempoolDir: defaultDatabaseMempoolDirname,
		LogDir:             defaultLogDir,
		RPCKey:             defaultRPCKeyFile,
		RPCCert:            defaultRPCCertFile,
		// Generate:             defaultGenerate,
		WalletName:           defaultWalletName,
		DisableTLS:           defaultDisableRpcTLS,
		DisableRPC:           false,
		RPCDisableAuth:       false,
		DiscoverPeers:        true,
		TestNet:              true,
		DiscoverPeersAddress: "127.0.0.1:9330", //"35.230.8.182:9339",
		NodeMode:             defaultNodeMode,
		PrivateKey:           common.EmptyString,
		FastStartup:          defaultFastStartup,
		TxPoolTTL:            defaultTxPoolTTL,
		TxPoolMaxTx:          defaultTxPoolMaxTx,
		PersistMempool:       defaultPersistMempool,
	}

	// Service options which are only added on Windows.
	serviceOpts := serviceOptions{}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.  Any errors aside from the
	// help message error can be ignored here since they will be caught by
	// the final parse below.
	preCfg := cfg
	preParser := newConfigParser(&preCfg, &serviceOpts, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return nil, nil, err
		}
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)
	if preCfg.ShowVersion {
		fmt.Println(appName, "version", "0.0")
		os.Exit(common.ExitCodeUnknow)
	}

	// Perform service command and exit if specified.  Invalid service
	// commands show an appropriate error.  Only runs on Windows since
	// the runServiceCommand function will be nil when not on Windows.
	if serviceOpts.ServiceCommand != "" && runServiceCommand != nil {
		err := runServiceCommand(serviceOpts.ServiceCommand)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(common.ExitCodeUnknow)
	}

	// Load additional config from file.
	var configFileError error
	parser := newConfigParser(&cfg, &serviceOpts, flags.Default)
	if _, err := os.Stat(preCfg.ConfigFile); os.IsNotExist(err) {
		err := createDefaultConfigFile(preCfg.ConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating a default config file: %v\n", err)
		}
	}

	errParse := flags.NewIniParser(parser).ParseFile(preCfg.ConfigFile)
	if errParse != nil {
		if _, ok := errParse.(*os.PathError); !ok {
			fmt.Fprintf(os.Stderr, "Error parsing config file: %v\n", errParse)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, errParse
		}
		configFileError = errParse
	}

	// Parse command line options again to ensure they take precedence.
	remainingArgs, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, usageMessage)
		}
		return nil, nil, err
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	err = os.MkdirAll(defaultHomeDir, 0700)
	if err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = fmt.Errorf(str, e.Path, link)
			}
		}

		str := "%s: Failed to create home directory: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	// Multiple networks can't be selected simultaneously.
	numNets := 0
	// Count number of network flags passed; assign active network component
	// while we're at it
	if cfg.TestNet {
		numNets++
		activeNetParams = &testNetParams
	}

	if numNets > 1 {
		Logger.log.Error("The testnet, regtest, segnet, and simnet component can't be used together -- choose one of the four")
		os.Exit(common.ExitCodeUnknow)
	}

	// Append the network type to the data directory so it is "namespaced"
	// per network.  In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	cfg.DataDir = common.CleanAndExpandPath(cfg.DataDir, defaultHomeDir)
	cfg.DataDir = filepath.Join(cfg.DataDir, netName(activeNetParams))

	// Append the network type to the log directory so it is "namespaced"
	// per network in the same fashion as the data directory.
	cfg.LogDir = common.CleanAndExpandPath(cfg.LogDir, defaultHomeDir)
	cfg.LogDir = filepath.Join(cfg.LogDir, netName(activeNetParams))

	// Special show command to list supported subsystems and exit.
	if cfg.LogLevel == "show" {
		fmt.Println("Supported subsystems", supportedSubsystems())
		os.Exit(common.ExitCodeUnknow)
	}

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	initLogRotator(filepath.Join(cfg.LogDir, defaultLogFilename))

	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(cfg.LogLevel); err != nil {
		err := fmt.Errorf("%s: %v", funcName, err.Error())
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --addPeer and --connect do not mix.
	if len(cfg.AddPeers) > 0 && len(cfg.ConnectPeers) > 0 {
		str := "%s: the --addpeer and --connect options can not be mixed"
		err := fmt.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --proxy or --connect without --listen disables listening.
	if (cfg.Proxy != common.EmptyString || len(cfg.ConnectPeers) > 0) &&
		len(cfg.Listener) == 0 {
		cfg.DisableListen = true
	}

	// Add the default listener if none were specified. The default
	// listener is all addresses on the listen port for the network
	// we are to connect to.
	if len(cfg.Listener) == 0 {
		cfg.Listener = net.JoinHostPort("", activeNetParams.DefaultPort)
	}

	if !cfg.RPCDisableAuth {
		if cfg.RPCUser == cfg.RPCLimitUser && cfg.RPCUser != "" {
			str := "%s: --rpcuser and --rpclimituser must not specify the same username"
			err := fmt.Errorf(str, funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}

		// Check to make sure limited and admin users don't have the same password
		if cfg.RPCPass == cfg.RPCLimitPass && cfg.RPCPass != "" {
			str := "%s: --rpcpass and --rpclimitpass must not specify the same password"
			err := fmt.Errorf(str, funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}

		// The RPC server is disabled if no username or password is provided.
		if (cfg.RPCUser == "" || cfg.RPCPass == "") &&
			(cfg.RPCLimitUser == "" || cfg.RPCLimitPass == "") {
			Logger.log.Info("The RPC server is disabled if no username or password is provided.")
			cfg.DisableRPC = true
		}
	}

	if cfg.DisableRPC {
		Logger.log.Info("RPC service is disabled")
	}

	// Default RPC to listen on localhost only.
	if !cfg.DisableRPC && len(cfg.RPCListeners) == 0 {
		addrs, err := net.LookupHost("0.0.0.0")
		if err != nil {
			return nil, nil, err
		}
		// Get address from env
		externalAddress := os.Getenv("EXTERNAL_ADDRESS")
		if externalAddress != "" {
			host, _, err := net.SplitHostPort(externalAddress)
			if err == nil && host != "" {
				addrs = []string{host}
			}
		}

		//Logger.log.Info(externalAddress, addrs)

		cfg.RPCListeners = make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, activeNetParams.rpcPort)
			cfg.RPCListeners = append(cfg.RPCListeners, addr)
		}
	}

	// Add default port to all listener addresses if needed and remove
	// duplicate addresses.
	cfg.Listener = normalizeAddress(cfg.Listener, activeNetParams.DefaultPort)

	// Add default port to all rpc listener addresses if needed and remove
	// duplicate addresses.
	cfg.RPCListeners = normalizeAddresses(cfg.RPCListeners,
		activeNetParams.rpcPort)

	// Only allow TLS to be disabled if the RPC is bound to localhost
	// addresses.
	if !cfg.DisableRPC && cfg.DisableTLS {
		allowedTLSListeners := map[string]struct{}{
			"localhost": {},
			"127.0.0.1": {},
			"::1":       {},
			"0.0.0.0":   {},
		}

		for _, addr := range cfg.RPCListeners {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				str := "%s: RPC listen interface '%s' is " +
					"invalid: %v"
				err := fmt.Errorf(str, funcName, addr, err)
				fmt.Fprintln(os.Stderr, err)
				fmt.Fprintln(os.Stderr, usageMessage)
				return nil, nil, err
			}
			if _, ok := allowedTLSListeners[host]; !ok {
				str := "%s: the --notls option may not be used when binding RPC to non localhost addresses: %s"
				err := fmt.Errorf(str, funcName, addr)
				fmt.Fprintln(os.Stderr, err)
				fmt.Fprintln(os.Stderr, usageMessage)
				return nil, nil, err
			}
		}
	}

	if cfg.DiscoverPeers {
		if cfg.DiscoverPeersAddress == "" {
			err := errors.New("discover peers server is empty")
			return nil, nil, err
		}
	}

	// Warn about missing config file only after all other configuration is
	// done.  This prevents the warning on help messages and invalid
	// options.  Note this should go directly before the return.
	if configFileError != nil {
		spew.Dump(configFileError)
	}
	return &cfg, remainingArgs, nil
}

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

func (conf *config) GetUserKeySet() (*cashec.KeySet, error) {
	if conf.PrivateKey == common.EmptyString {
		return nil, errors.New("user key set cant be empty")
	}
	KeySetUser := &cashec.KeySet{}
	temp, err := wallet.Base58CheckDeserialize(conf.PrivateKey)
	if err != nil {
		return nil, err
	}
	KeySetUser.ImportFromPrivateKey(&temp.KeySet.PrivateKey)
	// lastByte := KeySetUser.PaymentAddress.Pk[len(KeySetUser.PaymentAddress.Pk)-1]
	// shardIDSender, err := common.GetTxSenderChain(lastByte)
	// Logger.log.Info("shardID: ", shardIDSender)
	return KeySetUser, nil
}
