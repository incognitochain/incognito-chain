package devframework

import (
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"os"
)

const (
	ID_MAINNET = iota
	ID_TESTNET
	ID_TESTNET2
	ID_LOCAL
	ID_CUSTOM
)

func NewStandaloneSimulation(name string, conf Config) *NodeEngine {
	if conf.DataDir == "" {
		conf.DataDir = "./data"
	}
	if conf.ResetDB {
		os.RemoveAll(conf.DataDir)
	}

	switch conf.Network {
	case ID_LOCAL:
		os.Setenv(config.NetworkKey, config.LocalNetwork)
	case ID_CUSTOM:
		os.Setenv(config.NetworkKey, config.TestNetNetwork)
		os.Setenv(config.NetworkVersionKey, config.TestNetVersion1)
	case ID_TESTNET:
		os.Setenv(config.NetworkKey, config.TestNetNetwork)
		os.Setenv(config.NetworkVersionKey, config.TestNetVersion1)
	case ID_TESTNET2:
		os.Setenv(config.NetworkKey, config.TestNetNetwork)
		os.Setenv(config.NetworkVersionKey, config.TestNetVersion2)
	case ID_MAINNET:
		os.Setenv(config.NetworkKey, config.MainnetNetwork)
	}

	config.LoadConfig()

	sim := &NodeEngine{
		config:            conf,
		simName:           name,
		timer:             NewTimeEngine(),
		accountSeed:       "master_account",
		accountGenHistory: make(map[int]int),
		committeeAccount:  make(map[int][]account.Account),
		listennerRegister: make(map[int][]func(msg interface{})),
	}

	config.LoadParam()
	portal.SetupParam()
	return sim
}
