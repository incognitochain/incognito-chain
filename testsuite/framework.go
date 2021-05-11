package devframework

import (
	"os"
	"time"

	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func NewStandaloneSimulation(name string, config Config) *NodeEngine {
	if config.DataDir == "" {
		config.DataDir = "/tmp/database"
	}
	if config.ResetDB {
		os.RemoveAll(config.DataDir)
	}

	sim := &NodeEngine{
		config:            config,
		simName:           name,
		timer:             NewTimeEngine(),
		accountSeed:       "master_account",
		accountGenHistory: make(map[int]int),
		committeeAccount:  make(map[int][]account.Account),
		listennerRegister: make(map[int][]func(msg interface{})),
	}

	sim.init()
	time.Sleep(1 * time.Second)
	return sim
}
