package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
)

// See loadConfig for details on the configuration load process.
type config struct {
	RPCPort int `long:"rpcport" short:"p" description:"Linsten port of RPC server"`
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *config, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	return parser
}

// loadConfig
// - set default config
// - read config from cmd line params
// - return config object
func loadConfig() (*config, error) {
	// create config object from default values
	cfg := config{
		RPCPort: DefaultRPCServerPort,
	}

	//preCfg := cfg
	preParser := newConfigParser(&cfg, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}

	return &cfg, nil
}
