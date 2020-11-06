package devframework

import (
	"github.com/incognitochain/incognito-chain/devframework/rpcclient"
	"os"
	"time"
)

func NewRPCClient(endpoint string) *rpcclient.RPCClient {
	return rpcclient.NewRPCClient(&RemoteRPCClient{endpoint: endpoint})
}

func NewStandaloneSimulation(name string, config Config) *SimulationEngine {
	os.RemoveAll(name)
	sim := &SimulationEngine{
		config:            config,
		simName:           name,
		timer:             NewTimeEngine(),
		accountSeed:       "master_account",
		accountGenHistory: make(map[int]int),
		committeeAccount:  make(map[int][]Account),
	}
	sim.init()
	time.Sleep(1 * time.Second)
	return sim
}
