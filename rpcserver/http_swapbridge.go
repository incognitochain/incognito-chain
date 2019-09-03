package rpcserver

import (
	"encoding/hex"
	"fmt"

	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
)

// handleGetBridgeSwapProof returns a proof of a new bridge committee (for a given beacon block height)
func (httpServer *HttpServer) handleGetBridgeSwapProof(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Infof("handleGetBridgeSwapProof params: %+v", params)
	listParams := params.([]interface{})
	height := uint64(listParams[0].(float64))
	bc := httpServer.config.BlockChain
	db := *httpServer.config.Database

	// Get proof of instruction on beacon
	beaconInstProof, beaconBlock, err := getSwapProofOnBeacon(height, db, httpServer.config.ConsensusEngine, metadata.BridgeSwapConfirmMeta)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Get proof of instruction on bridge
	bridgeInstProof, err := getBridgeSwapProofOnBridge(beaconBlock, bc, db, httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Decode instruction to send to Ethereum without having to decode on client
	decodedInst, err := blockchain.DecodeInstruction(beaconInstProof.inst)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	inst := hex.EncodeToString(decodedInst)

	return buildProofResult(inst, beaconInstProof, bridgeInstProof, "", ""), nil
}

// getBridgeSwapProofOnBridge finds a bridge committee swap instruction in a bridge block and returns its proof; the bridge block must be included in a given beaconBlock
func getBridgeSwapProofOnBridge(
	beaconBlock *blockchain.BeaconBlock,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
	ce ConsensusEngine,
) (*swapProof, error) {
	// Get bridge block and check if it contains bridge swap instruction
	b, instID, err := findBridgeBlockWithInst(beaconBlock, bc, db)
	if err != nil {
		return nil, err
	}
	insts := b.Body.Instructions
	block := &shardBlock{ShardBlock: b}
	return buildProofForBlock(block, insts, instID, db, ce)
}

// findBridgeBlockWithInst traverses all shard blocks included in a beacon block and returns the one containing a bridge swap instruction
func findBridgeBlockWithInst(
	beaconBlock *blockchain.BeaconBlock,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) (*blockchain.ShardBlock, int, error) {
	bridgeID := byte(common.BridgeShardID)
	for _, state := range beaconBlock.Body.ShardState[bridgeID] {
		bridgeBlock, _, err := getShardAndBeaconBlocks(state.Height, bc, db)
		if err != nil {
			return nil, 0, err
		}

		_, bridgeInstID := findCommSwapInst(bridgeBlock.Body.Instructions, metadata.BridgeSwapConfirmMeta)
		BLogger.log.Debugf("Finding swap bridge inst in bridge block %d %d", state.Height, bridgeInstID)
		if bridgeInstID >= 0 {
			return bridgeBlock, bridgeInstID, nil
		}
	}

	return nil, 0, fmt.Errorf("cannot find bridge swap instruction in bridge block")
}
