package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

const (
	defaultStrategy   = 1
	defaultTotalTxs   = 1000000
	defaultRPCAddress = "http://127.0.0.1:9334"
)

// See loadConfig for details on the configuration load process.
type config struct {
	Strategy      int      `long:"strategy" short:"s" description:"Strategy Id"`
	TotalTxs      int      `long:"txs" short:"t" description:"Total transactions to test"`
	RPCAddress    []string `long:"rpcaddress" short:"r" description:"RPC address of any node"`
	GenesisPrvKey string   `long:"genesisprvkey" short:"g" description:"Genesis Private PubKey which account hold coins"`
}

func loadConfig() (*config, error) {
	cfg := config{
		Strategy:   defaultStrategy,
		TotalTxs:   defaultTotalTxs,
		RPCAddress: []string{defaultRPCAddress},
	}

	preParser := flags.NewParser(&cfg, flags.Default)
	_, err := preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}

	return &cfg, nil
}
