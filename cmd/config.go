package main

import (
	"github.com/jessevdk/go-flags"
	"os"
	"fmt"
)

const (
	defaultRPCPort = 9339
)

// See loadConfig for details on the configuration load process.
type config struct {
	Command string `long:"cmd" short:"c" description:"Command name"`

	// For Wallet
	WalletName       string `long:"wallet" description:"Wallet Database Name file, default is 'wallet'"`
	WalletPassphrase string `long:"walletpassphrase" description:"Wallet passphrase"`
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *config, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	return parser
}

func loadConfig() (*config, error) {
	cfg := config{
	}

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
