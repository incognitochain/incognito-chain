package config

type Config struct {
	DataDir                     string   `json:"data_dir" short:"d" long:"datadir" description:"Directory to store data"`
	MempoolDir                  string   `json:"mempool_dir" short:"m" long:"mempooldir" description:"Mempool Directory"`
	LogDir                      string   `json:"log_dir" short:"l" long:"logdir" description:"Directory to log output."`
	LogLevel                    string   `json:"log_level" long:"loglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	AddPeers                    []string `json:"add_peers" short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	ConnectPeers                []string `json:"connect_peers" short:"c" long:"connect" description:"Connect only to the specified peers at startup"`
	DisableListen               bool     `json:"disable_listen" long:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen"`
	Listener                    string   `json:"listener" long:"listen" description:"Add an interface/port to listen for connections (default all interfaces port: 9333, testnet: 9444)"`
	MaxOutPeers                 int      `json:"max_out_peers" long:"maxoutpeers" description:"Max number of outbound peers"`
	MaxInPeers                  int      `json:"max_in_peers" long:"maxinpeers" description:"Max number of inbound peers"`
	ExternalAddress             string   `json:"external_address" long:"externaladdress" description:"External address"`
	RPCDisableAuth              bool     `json:"" long:"norpcauth" description:"Disable RPC authorization by username/password"`
	RPCUser                     string   `short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass                     string   `short:"P" long:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCLimitUser                string   `long:"rpclimituser" description:"Username for limited RPC connections"`
	RPCLimitPass                string   `long:"rpclimitpass" default-mask:"-" description:"Password for limited RPC connections"`
	RPCListeners                []string `long:"rpclisten" description:"Add an interface/port to listen for RPC connections (default port: 9334, testnet: 9334)"`
	RPCWSListeners              []string `long:"rpcwslisten" description:"Add an interface/port to listen for RPC Websocket connections (default port: 19334, testnet: 19334)"`
	RPCCert                     string   `long:"rpccert" description:"File containing the certificate file"`
	RPCKey                      string   `long:"rpckey" description:"File containing the certificate key"`
	RPCLimitRequestPerDay       int      `long:"rpclimitrequestperday" description:"Max request per day by remote address"`
	RPCLimitRequestErrorPerHour int      `long:"rpclimitrequesterrorperhour" description:"Max request error per hour by remote address"`
	RPCMaxClients               int      `long:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	RPCMaxWSClients             int      `long:"rpcmaxwsclients" description:"Max number of RPC clients for standard connections"`
	RPCQuirks                   bool     `long:"rpcquirks" description:"Mirror some JSON-RPC quirks of coin Core -- NOTE: Discouraged unless interoperability issues need to be worked around"`
	DisableRPC                  bool     `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass or rpclimituser/rpclimitpass is specified"`
	DisableTLS                  bool     `long:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	Proxy                       string   `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	IsLocal                     bool     `json:"is_local" long:"local" description:"Use the local network"`
	IsTestNet                   bool     `json:"is_testnet" long:"local" description:"Use the local network"`
	TestNetVersion              int      `json:"testnet_version" long:"testnetversion" description:"Use the test network"`
}
