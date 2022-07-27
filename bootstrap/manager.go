package bootstrap

import "github.com/incognitochain/incognito-chain/blockchain"

type bootstrapProcess struct {
	checkPointHeight uint64
}

type BootstrapManager struct {
	blockchain       *blockchain.BlockChain
	lastBootStrap    *bootstrapProcess
	runningBootStrap *bootstrapProcess
}

func NewBootStrapManager(bc *blockchain.BlockChain) *BootstrapManager {
	return &BootstrapManager{bc, nil, nil}
}
func (s *BootstrapManager) Start() {
	go func() {
		//backup beacon
		for cid := -1; cid < s.blockchain.GetActiveShardNumber(); cid++ {

		}
	}()
}
