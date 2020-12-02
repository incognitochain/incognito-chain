package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal/portalprocess"
)

func (blockchain *BlockChain) processPortalInstructions(portalStateDB *statedb.StateDB, block *BeaconBlock) error {
	// Note: should comment this code if you need to create local chain.
	if blockchain.config.ChainParams.Net == Testnet && block.Header.Height < 1580600 {
		return nil
	}

	beaconHeight := block.Header.Height - 1
	instructions := block.Body.Instructions
	portalParams := blockchain.GetPortalParams(block.GetHeight())
	pm := portalprocess.NewPortalManager()

	return portalprocess.ProcessPortalInstructions(portalStateDB, portalParams, beaconHeight, instructions, pm)
}

func (blockchain *BlockChain) processRelayingInstructions(block *BeaconBlock) error {
	relayingState, err := blockchain.InitRelayingHeaderChainStateFromDB()
	if err != nil {
		Logger.log.Error(err)
		return nil
	}
	// because relaying instructions in received beacon block were sorted already as desired so dont need to do sorting again over here
	return portalprocess.ProcessRelayingInstructions(block.Body.Instructions, relayingState)
}