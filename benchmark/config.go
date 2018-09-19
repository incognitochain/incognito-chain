package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
)

const (
	defaultStrategy = 1
	defaultTotalTxs = 1000000
	defaultRPCAddress = "127.0.0.1:9334"
)

// See loadConfig for details on the configuration load process.
type config struct {
	Strategy	int		`long:"strategy" short:"s" description:"Strategy ID"`
	TotalTxs	int		`long:"txs" short:"t" description:"Total transactions to test"`
	RPCAddress	string	`long:"rpcaddress" short:"ra" description:"RPC address of any node"`
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *config, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	return parser
}

/**
// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
//
// The above results in btcd functioning properly without any config settings
// while still allowing the user to override settings with config files and
// command line options.  Command line options always take precedence.
*/
func loadConfig() (*config, error) {
	cfg := config{
		Strategy: defaultStrategy,
		TotalTxs: defaultTotalTxs,
		RPCAddress: defaultRPCAddress,
	}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.  Any errors aside from the
	// help message error can be ignored here since they will be caught by
	// the final parse below.
	preCfg := cfg
	preParser := newConfigParser(&preCfg, flags.HelpFlag)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}

	return &cfg, nil
}
