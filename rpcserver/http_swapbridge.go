package rpcserver

import (
	"encoding/hex"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

// handleGetBridgeSwapProof returns a proof of a new bridge committee (for a given beacon block height)
func (httpServer *HttpServer) handleGetBridgeSwapProof(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBridgeSwapProof params: %+v", params)
	listParams := params.([]interface{})
	height := uint64(listParams[0].(float64))
	bc := httpServer.config.BlockChain
	db := *httpServer.config.Database

	// Get proof of instruction on beacon
	beaconInstProof, beaconBlock, err := getBridgeSwapProofOnBeacon(height, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Get proof of instruction on bridge
	bridgeInstProof, err := getBridgeSwapProofOnBridge(beaconBlock, bc, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Decode instruction to send to Ethereum without having to decode on client
	decodedInst := hex.EncodeToString(blockchain.DecodeInstruction(beaconInstProof.inst))

	return jsonresult.GetInstructionProof{
		Instruction: decodedInst,

		BeaconInstPath:         beaconInstProof.instPath,
		BeaconInstPathIsLeft:   beaconInstProof.instPathIsLeft,
		BeaconInstRoot:         beaconInstProof.instRoot,
		BeaconBlkData:          beaconInstProof.blkData,
		BeaconBlkHash:          beaconInstProof.blkHash,
		BeaconSignerPubkeys:    beaconInstProof.signerPubkeys,
		BeaconSignerSig:        beaconInstProof.signerSig,
		BeaconSignerPaths:      beaconInstProof.signerPaths,
		BeaconSignerPathIsLeft: beaconInstProof.signerPathIsLeft,

		BridgeInstPath:         bridgeInstProof.instPath,
		BridgeInstPathIsLeft:   bridgeInstProof.instPathIsLeft,
		BridgeInstRoot:         bridgeInstProof.instRoot,
		BridgeBlkData:          bridgeInstProof.blkData,
		BridgeBlkHash:          bridgeInstProof.blkHash,
		BridgeSignerPubkeys:    bridgeInstProof.signerPubkeys,
		BridgeSignerSig:        bridgeInstProof.signerSig,
		BridgeSignerPaths:      bridgeInstProof.signerPaths,
		BridgeSignerPathIsLeft: bridgeInstProof.signerPathIsLeft,
	}, nil
}

// getBridgeSwapProofOnBridge finds a bridge committee swap instruction in a bridge block and returns its proof; the bridge block must be included in a given beaconBlock
func getBridgeSwapProofOnBridge(
	beaconBlock *blockchain.BeaconBlock,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) (*swapProof, error) {
	// Get bridge block and check if it contains bridge swap instruction
	bridgeBlock, instID, err := findBridgeBlockWithInst(beaconBlock, bc, db)
	if err != nil {
		return nil, err
	}
	insts := bridgeBlock.Body.Instructions
	return buildProofOnBridge(bridgeBlock, insts, instID, db)
}

// getBridgeSwapProofOnBeacon finds in a given beacon block a bridge committee swap instruction and returns its proof
func getBridgeSwapProofOnBeacon(
	height uint64,
	db database.DatabaseInterface,
) (*swapProof, *blockchain.BeaconBlock, error) {
	// Get beacon block
	beaconBlocks, err := blockchain.FetchBeaconBlockFromHeight(db, height, height)
	if len(beaconBlocks) == 0 {
		return nil, nil, fmt.Errorf("cannot find beacon block with height %d", height)
	}
	beaconBlock := beaconBlocks[0]

	// Find bridge swap instruction in beacon block
	insts := beaconBlock.Body.Instructions
	_, instID := findCommSwapInst(insts, metadata.BridgeSwapConfirmMeta)
	if instID < 0 {
		return nil, nil, fmt.Errorf("cannot find bridge swap instruction in beacon block")
	}
	proof, err := buildProofOnBeacon(beaconBlock, insts, instID, db)
	if err != nil {
		return nil, nil, err
	}
	return proof, beaconBlock, nil
}

// findBridgeBlockWithInst traverses all shard blocks included in a beacon block and returns the one containing a bridge swap instruction
func findBridgeBlockWithInst(
	beaconBlock *blockchain.BeaconBlock,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) (*blockchain.ShardBlock, int, error) {
	bridgeID := byte(1) // TODO(@0xbunyip); replace with bridge's shardID
	for _, state := range beaconBlock.Body.ShardState[bridgeID] {
		bridgeBlock, _, err := getShardAndBeaconBlocks(state.Height, bc, db)
		if err != nil {
			return nil, 0, err
		}

		_, bridgeInstID := findCommSwapInst(bridgeBlock.Body.Instructions, metadata.BridgeSwapConfirmMeta)
		fmt.Printf("[db] finding swap bridge inst in bridge block %d %d\n", state.Height, bridgeInstID)
		if bridgeInstID >= 0 {
			return bridgeBlock, bridgeInstID, nil
		}
	}

	return nil, 0, fmt.Errorf("cannot find bridge swap instruction in bridge block")
}
