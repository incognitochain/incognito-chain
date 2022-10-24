package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
)

var c *config

func Config() *config {
	return c
}

// AbortConfig use for unit test only
// DO NOT use this function for development process
func AbortConfig() {
	c = &config{}
}

type config struct {
	//Basic config
	DataDir     string `mapstructure:"data_dir" short:"d" long:"datadir" description:"Directory to store data"`
	DatabaseDir string `mapstructure:"database_dir" long:"datapre" description:"Database dir"`
	MempoolDir  string `mapstructure:"mempool_dir" short:"m" long:"mempooldir" description:"Mempool Directory"`
	LogDir      string `mapstructure:"log_dir" short:"l" long:"logdir" description:"Directory to log output."`
	LogLevel    string `mapstructure:"log_level" long:"loglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	LogFileName string `mapstructure:"log_file_name" long:"logfilename" description:"log file name"`

	//Peer Config
	AddPeers             []string `mapstructure:"add_peers" short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	ConnectPeers         []string `mapstructure:"connect_peers" short:"c" long:"connect" description:"Connect only to the specified peers at startup"`
	DisableListen        bool     `mapstructure:"disable_listen" long:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen"`
	Listener             string   `mapstructure:"listener" long:"listen" description:"Add an interface/port to listen for connections (default all interfaces port: 9333, testnet: 9444)"`
	MaxPeers             int      `mapstructure:"max_peers" long:"maxpeers" description:"Max number of inbound and outbound peers"`
	MaxOutPeers          int      `mapstructure:"max_out_peers" long:"maxoutpeers" description:"Max number of outbound peers"`
	MaxInPeers           int      `mapstructure:"max_in_peers" long:"maxinpeers" description:"Max number of inbound peers"`
	DiscoverPeers        bool     `mapstructure:"discover_peers" long:"discoverpeers" description:"Enable discover peers"`
	DiscoverPeersAddress string   `mapstructure:"discover_peers_address" long:"discoverpeersaddress" description:"Url to connect discover peers server"`
	MaxPeersSameShard    int      `mapstructure:"max_peers_same_shard" long:"maxpeersameshard" description:"Max peers in same shard for connection"`
	MaxPeersOtherShard   int      `mapstructure:"max_pmax_peers_other_shard" long:"maxpeerothershard" description:"Max peers in other shard for connection"`
	MaxPeersOther        int      `mapstructure:"max_peers_other" long:"maxpeerother" description:"Max peers in other for connection"`
	MaxPeersNoShard      int      `mapstructure:"max_peers_no_shard" long:"maxpeernoshard" description:"Max peers in no shard for connection"`
	MaxPeersBeacon       int      `mapstructure:"max_peers_beacon" long:"maxpeerbeacon" description:"Max peers in beacon for connection"`

	//Rpc Config
	ExternalAddress             string   `mapstructure:"external_address" long:"externaladdress" description:"External address"`
	RPCDisableAuth              bool     `mapstructure:"rpc_disable_auth" long:"norpcauth" description:"Disable RPC authorization by username/password"`
	RPCUser                     string   `mapstructure:"rpc_user" yaml:"rpc_user" short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass                     string   `mapstructure:"rpc_pass" short:"P" long:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCLimitUser                string   `mapstructure:"rpc_limit_user" long:"rpclimituser" description:"Username for limited RPC connections"`
	RPCLimitPass                string   `mapstructure:"rpc_limit_pass" long:"rpclimitpass" default-mask:"-" description:"Password for limited RPC connections"`
	RPCListeners                []string `mapstructure:"rpc_listeners" long:"rpclisten" description:"Add an interface/port to listen for RPC connections (default port: 9334, testnet: 9334)"`
	RPCWSListeners              []string `mapstructure:"rpc_ws_listeners" long:"rpcwslisten" description:"Add an interface/port to listen for RPC Websocket connections (default port: 19334, testnet: 19334)"`
	RPCCert                     string   `mapstructure:"rpc_cert" long:"rpccert" description:"File containing the certificate file"`
	RPCKey                      string   `mapstructure:"rpc_key" long:"rpckey" description:"File containing the certificate key"`
	RPCLimitRequestPerDay       int      `mapstructure:"rpc_limit_request_per_day" long:"rpclimitrequestperday" description:"Max request per day by remote address"`
	RPCLimitRequestErrorPerHour int      `mapstructure:"rpc_limit_request_error_per_hour" long:"rpclimitrequesterrorperhour" description:"Max request error per hour by remote address"`
	RPCMaxClients               int      `mapstructure:"rpc_max_clients" long:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	RPCMaxWSClients             int      `mapstructure:"rpc_max_ws_clients" long:"rpcmaxwsclients" description:"Max number of RPC clients for standard connections"`
	RPCQuirks                   bool     `mapstructure:"rpc_quirks" long:"rpcquirks" description:"Mirror some JSON-RPC quirks of coin Core -- NOTE: Discouraged unless interoperability issues need to be worked around"`
	DisableRPC                  bool     `mapstructure:"disable_rpc" long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass or rpclimituser/rpclimitpass is specified"`
	DisableTLS                  bool     `mapstructure:"disable_tls" long:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	Proxy                       string   `mapstructure:"proxy" long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`

	//Network Config
	IsLocal        bool `description:"Use the local network"`
	IsTestNet      bool `description:"Use the testnet network"`
	IsMainNet      bool `description:"Use the mainnet network"`
	TestNetVersion int  `description:"Use the test network"`

	RelayShards string `mapstructure:"relay_shards" long:"relayshards" description:"set relay shards of this node when in 'relay' mode if noderole is auto then it only sync shard data when user is a shard producer/validator"`
	// For Wallet
	EnableWallet     bool   `mapstructure:"enable_wallet" long:"enablewallet" description:"Enable wallet"`
	WalletName       string `mapstructure:"wallet_name" long:"wallet" description:"Wallet Database Name file, default is 'wallet'"`
	WalletPassphrase string `mapstructure:"wallet_passphrase" long:"walletpassphrase" description:"Wallet passphrase"`
	WalletAutoInit   bool   `mapstructure:"wallet_auto_init" long:"walletautoinit" description:"Init wallet automatically if not exist"`
	WalletShardID    int    `mapstructure:"wallet_shard_id" long:"walletshardid" description:"ShardID which wallet use to create account"`

	//Fast start up config
	FastStartup bool `mapstructure:"fast_start_up" long:"faststartup" description:"Load existed shard/chain dependencies instead of rebuild from block data"`

	//Txpool config
	TxPoolTTL   uint   `mapstructure:"tx_pool_ttl" long:"txpoolttl" description:"Set Time To Live (TTL) Value for transaction that enter pool"`
	TxPoolMaxTx uint64 `mapstructure:"tx_pool_max_tx" long:"txpoolmaxtx" description:"Set Maximum number of transaction in pool"`
	LimitFee    uint64 `mapstructure:"limit_fee" long:"limitfee" description:"Limited fee for tx(per Kb data), default is 0.00 PRV"`

	//Mempool config
	IsLoadFromMempool bool `mapstructure:"is_load_from_mem_pool" long:"loadmempool" description:"Load transactions from Mempool database"`
	IsPersistMempool  bool `mapstructure:"is_persist_mem_pool" long:"persistmempool" description:"Persistence transaction in memepool database"`

	//Mining config
	EnableMining bool   `mapstructure:"enable_mining" long:"mining" description:"enable mining"`
	MiningKeys   string `mapstructure:"mining_keys" long:"miningkeys" description:"keys used for different consensus algorigthm"`
	PrivateKey   string `mapstructure:"private_key" long:"privatekey" description:"your wallet privatekey"`
	Accelerator  bool   `mapstructure:"accelerator" long:"accelerator" description:"Relay Node Configuration For Consensus"`

	// Highway
	Libp2pPrivateKey string `mapstructure:"p2p_private_key" long:"libp2pprivatekey" description:"Private key used to create node's PeerID, empty to generate random key each run"`

	//backup
	BootstrapAddress string `mapstructure:"bootstrap" yaml:"bootstrap" long:"bootstrap" description:"Endpoint of fullnodes to download backup database"`
	Backup           bool   `mapstructure:"backup" long:"backup" description:"backup mode"`
	IsFullValidation bool   `mapstructure:"is_full_validation" long:"is_full_validation" description:"fully validation data"`

	// Optional : db to store coin by OTA key (for v2)
	OutcoinDatabaseDir   string `mapstructure:"coin_data_pre" long:"coindatapre" description:"Output coins by OTA key database dir"`
	NumIndexerWorkers    int64  `mapstructure:"num_indexer_workers" long:"numindexerworkers" description:"Number of workers for caching output coins"`
	IndexerAccessTokens  string `mapstructure:"indexer_access_token" long:"indexeraccesstoken" description:"The access token for caching output coins"`
	UseOutcoinDatabase   []bool `mapstructure:"use_coin_data" long:"usecoindata" description:"Store output coins by known OTA keys"`
	AllowStatePruneByRPC bool   `mapstructure:"allow_state_prune_by_rpc" long:"allowstateprunebyrpc" description:"allow state pruning flag"`
	OfflinePrune         bool   `mapstructure:"offline_prune" long:"offlineprune" description:"offline pruning flag"`
	StateBloomSize       uint64 `mapstructure:"state_bloom_size" long:"statebloomsize" description:"state pruning bloom size"`
	EnableAutoPrune      bool   `mapstructure:"enable_auto_prune" long:"enableautoprune" description:"enable auto prune"`
	NumBlockTriggerPrune uint64 `mapstructure:"num_block_trigger_prune" long:"numblocktriggerprune" description:"number block trigger prune"`
	//backup and bootstrap
	BackupInterval int64 `mapstructure:"backup_interval" long:"backupinterval" description:"Backup Interval"`
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

func (c *config) loadNetwork() {
	switch utils.GetEnv(NetworkKey, MainnetNetwork) {
	case LocalNetwork:
		c.IsLocal = true
	case LocalDCSNetwork:
		c.IsLocal = true
	case TestNetNetwork:
		c.IsTestNet = true
		testnetVersion := utils.GetEnv(NetworkVersionKey, TestNetVersion1)
		version, err := strconv.Atoi(testnetVersion)
		if err != nil {
			panic(err)
		}
		c.TestNetVersion = version
	case MainnetNetwork:
		c.IsMainNet = true
	}
}

func (c *config) Network() string {
	res := utils.GetEnv(NetworkKey, MainnetNetwork)
	if res == TestNetNetwork {
		res = res + "-" + utils.GetEnv(NetworkVersionKey, TestNetVersion1)
	}
	return res
}

func (c *config) verify() {
	network := c.Network()
	// Multiple networks can't be selected simultaneously.
	numNets := 0
	if c.IsLocal {
		numNets++
	}
	if c.IsTestNet {
		numNets++
	}
	if c.IsMainNet {
		numNets++
	}

	if numNets > 1 {
		log.Println("The network can not be used together -- choose one of them")
		os.Exit(utils.ExitCodeUnknow)
	}

	// Append the network type to the data directory so it is "namespaced"
	// per network.  In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	c.DataDir = filepath.Join(c.DataDir, network)

	// Append the network type to the log directory so it is "namespaced"
	// per network in the same fashion as the data directory.
	c.LogDir = filepath.Join(c.DataDir, c.LogDir)
	c.LogFileName = filepath.Join(c.LogDir, c.LogFileName)

	// --addPeer and --connect do not mix.
	if len(c.AddPeers) > 0 && len(c.ConnectPeers) > 0 {
		str := "%s: the --addpeer and --connect options can not be mixed"
		fmt.Fprintln(os.Stderr, errors.New(str))
		panic(str)
	}

	// --proxy or --connect without --listen disables listening.
	if (c.Proxy != utils.EmptyString || len(c.ConnectPeers) > 0) &&
		len(c.Listener) == 0 {
		c.DisableListen = true
	}

	// Add the default listener if none were specified. The default
	// listener is all addresses on the listen port for the network
	// we are to connect to.
	if len(c.Listener) == 0 {
		c.Listener = net.JoinHostPort("", DefaultPort)
	}

	if !c.RPCDisableAuth {
		if c.RPCUser == c.RPCLimitUser && c.RPCUser != "" {
			str := "%s: --rpcuser and --rpclimituser must not specify the same username"
			fmt.Fprintln(os.Stderr, errors.New(str))
			panic(str)
		}

		// Check to make sure limited and admin users don't have the same password
		if c.RPCPass == c.RPCLimitPass && c.RPCPass != "" {
			str := "%s: --rpcpass and --rpclimitpass must not specify the same password"
			fmt.Fprintln(os.Stderr, errors.New(str))
			panic(str)
		}

		// The RPC server is disabled if no username or password is provided.
		if (c.RPCUser == "" || c.RPCPass == "") &&
			(c.RPCLimitUser == "" || c.RPCLimitPass == "") {
			log.Println("The RPC server is disabled if no username or password is provided.")
			c.DisableRPC = true
		}
	}

	if c.DisableRPC {
		log.Println("RPC service is disabled")
	}

	// Default RPC to listen on localhost only.
	if !c.DisableRPC && len(c.RPCListeners) == 0 {
		addrs, err := net.LookupHost("0.0.0.0")
		if err != nil {
			panic(err)
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
		c.RPCListeners = make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, DefaultRPCPort)
			c.RPCListeners = append(c.RPCListeners, addr)
		}
	}

	// Default RPC Ws to listen on localhost only.
	if !c.DisableRPC && len(c.RPCWSListeners) == 0 {
		addrs, err := net.LookupHost("0.0.0.0")
		if err != nil {
			panic(err)
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
		c.RPCWSListeners = make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, DefaultWSPort)
			c.RPCWSListeners = append(c.RPCWSListeners, addr)
		}
	}

	// Add default port to all listener addresses if needed and remove
	// duplicate addresses.
	c.Listener = normalizeAddress(c.Listener, DefaultPort)

	// Add default port to all rpc listener addresses if needed and remove
	// duplicate addresses.
	c.RPCListeners = normalizeAddresses(c.RPCListeners, DefaultRPCPort)
	// Add default port to all rpc listener addresses if needed and remove
	// duplicate addresses.
	c.RPCWSListeners = normalizeAddresses(c.RPCWSListeners, DefaultWSPort)

	// Only allow TLS to be disabled if the RPC is bound to localhost
	// addresses.
	if !c.DisableRPC && c.DisableTLS {
		allowedTLSListeners := map[string]struct{}{
			"localhost": {},
			"127.0.0.1": {},
			"::1":       {},
			"0.0.0.0":   {},
		}

		for _, addr := range c.RPCListeners {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				str := "%s: RPC listen interface '%s' is " +
					"invalid: %v"
				fmt.Fprintln(os.Stderr, errors.New(str))
				panic(str)
			}
			if _, ok := allowedTLSListeners[host]; !ok {
				str := "%s: the --notls option may not be used when binding RPC to non localhost addresses: %s"
				fmt.Fprintln(os.Stderr, errors.New(str))
				panic(str)
			}
		}
		for _, addr := range c.RPCWSListeners {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				str := "%s: WS RPC listen interface '%s' is " +
					"invalid: %v"
				fmt.Fprintln(os.Stderr, errors.New(str))
				panic(str)
			}
			if _, ok := allowedTLSListeners[host]; !ok {
				str := "%s: the --notls option may not be used when binding WS RPC to non localhost addresses: %s"
				fmt.Fprintln(os.Stderr, errors.New(str))
				panic(str)
			}
		}
	}

	if c.DiscoverPeers {
		if c.DiscoverPeersAddress == "" {
			err := errors.New("discover peers server is empty")
			panic(err)
		}
	}

	if c.Backup && c.BootstrapAddress != "" {
		err := errors.New("Backup and Bootstrap cannot be set together!")
		panic(err)
	}
}

func LoadConfig() *config {
	c = &config{}

	//get network
	c.loadNetwork()
	//load config from file
	c.loadConfig()
	//verify config
	c.verify()

	return c
}

func (c *config) loadConfig() {
	network := c.Network()
	//read config from file
	viper.SetConfigName(utils.GetEnv(ConfigFileKey, DefaultConfigFile))         // name of config file (without extension)
	viper.SetConfigType(utils.GetEnv(ConfigFileTypeKey, DefaultConfigFileType)) // REQUIRED if the config file does not have the extension in the name
	path := filepath.Join(utils.GetEnv(ConfigDirKey, DefaultConfigDir), network)
	fmt.Println(path)
	viper.AddConfigPath(path) // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			if err != nil {
				panic(err)
			}
		}
	} else {
		err = viper.Unmarshal(&c)
		if err != nil {
			panic(err)
		}
	}

	parser := flags.NewParser(c, flags.IgnoreUnknown)
	parser.Parse()
}

type gethParam struct {
	Host []string `mapstructure:"host"`
}

func (gethPram *gethParam) GetFromEnv() {
	var host, protocol, port string
	if utils.GetEnv(GethHostKey, utils.EmptyString) != utils.EmptyString {
		host = utils.GetEnv(GethHostKey, utils.EmptyString)
	}
	if utils.GetEnv(GethProtocolKey, utils.EmptyString) != utils.EmptyString {
		protocol = utils.GetEnv(GethProtocolKey, utils.EmptyString)
	}
	if utils.GetEnv(GethPortKey, utils.EmptyString) != utils.EmptyString {
		port = utils.GetEnv(GethPortKey, utils.EmptyString)
	}

	if host != utils.EmptyString || protocol != utils.EmptyString || port != utils.EmptyString {
		gethPram.Host = []string{rpccaller.BuildRPCServerAddress(protocol, host, port)}
	}
}
