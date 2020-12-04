package rpcserver

import (
	"encoding/hex"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func retrieveIncProof(
	confirmMetaType int,
	onBeacon bool,
	height uint64,
	txID *common.Hash,
	httpServer *HttpServer,
) (interface{}, *rpcservice.RPCError) {
	incProof := blockchain.NewIncProof(confirmMetaType)
	if !onBeacon {
		return getIncProofByHeight(incProof, httpServer, height, txID)
	}
	return getIncProofByHeightV2(incProof, httpServer, height, txID)
}

func getIncProofByHeightV2(
	incProof blockchain.IncProofInterface,
	httpServer *HttpServer,
	height uint64,
	txID *common.Hash,
) (interface{}, *rpcservice.RPCError) {
	// Get beacon block
	beaconBlock, err := getSingleBeaconBlockByHeight(httpServer.GetBlockchain(), height)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// find index of the confirm instruction in list of instructions
	inst, instID := incProof.FindConfirmInst(beaconBlock.Body.Instructions, txID)
	if instID == -1 {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, fmt.Errorf("cannot find inst %s in beacon block %d", txID.String(), height))
	}

	// Get proof of instruction on beacon
	beaconInstProof, err := getIncProofOnBeacon(incProof, inst, []*blockchain.BeaconBlock{beaconBlock}, httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Hex encode instruction bytes to send to Ethereum
	decodedInst, bcHeightStr := splitAndEncodeIncInstV2(incProof, beaconInstProof.inst)
	return buildProofResult(decodedInst, beaconInstProof, nil, bcHeightStr, ""), nil
}

func getIncProofByHeight(
	incProof blockchain.IncProofInterface,
	httpServer *HttpServer,
	height uint64,
	txID *common.Hash,
) (interface{}, *rpcservice.RPCError) {
	// Get bridge block and corresponding beacon blocks
	bridgeBlock, beaconBlocks, err := getShardAndBeaconBlocks(height, httpServer.GetBlockchain())
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	// Get proof of instruction on bridge
	bridgeInstProof, err := getIncProofOnBridge(incProof, txID, bridgeBlock, httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	fmt.Println("bridgeInstProof", bridgeInstProof)
	// Get proof of instruction on beacon
	beaconInstProof, err := getIncProofOnBeacon(incProof, bridgeInstProof.inst, beaconBlocks, httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	fmt.Println("beaconInstProof", beaconInstProof)
	// Decode instruction to send to Ethereum without having to decode on client
	decodedInst, bridgeHeight, beaconHeight := splitAndEncodeIncInst(incProof, bridgeInstProof.inst, beaconInstProof.inst)
	fmt.Println("decodedInst", decodedInst)
	fmt.Println("bridgeHeight", bridgeHeight)
	fmt.Println("beaconHeight", beaconHeight)
	//decodedInst := hex.EncodeToString(blockchain.DecodeInstruction(bridgeInstProof.inst))

	return buildProofResult(decodedInst, beaconInstProof, bridgeInstProof, beaconHeight, bridgeHeight), nil
}

// getIncProofOnBridge finds a beacon committee swap instruction in a given bridge block and returns its proof
func getIncProofOnBridge(
	incProof blockchain.IncProofInterface,
	txID *common.Hash,
	bridgeBlock *blockchain.ShardBlock,
	ce ConsensusEngine,
) (*swapProof, error) {
	insts := bridgeBlock.Body.Instructions
	_, instID := incProof.FindConfirmInst(insts, txID)
	if instID < 0 {
		return nil, fmt.Errorf("cannot find burning instruction in bridge block")
	}

	block := &shardBlock{ShardBlock: bridgeBlock}
	proof, err := buildProofForBlock(block, insts, instID, ce)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// getIncProofOnBeacon finds in given beacon blocks a specific instruction and returns its proof
func getIncProofOnBeacon(
	incProof blockchain.IncProofInterface,
	inst []string,
	beaconBlocks []*blockchain.BeaconBlock,
	ce ConsensusEngine,
) (*swapProof, error) {
	// Get beacon block and check if it contains beacon swap instruction
	b, instID := incProof.FindBeaconBlockWithConfirmInst(beaconBlocks, inst)
	if b == nil {
		return nil, fmt.Errorf("cannot find corresponding beacon block that includes burn instruction")
	}

	insts := b.Body.Instructions
	block := &beaconBlock{BeaconBlock: b}
	return buildProofForBlock(block, insts, instID, ce)
}

// splitAndDecodeInst splits inc insts into 2 parts: the inst itself and beaconHeight that contains the inst
func splitAndEncodeIncInstV2(incProof blockchain.IncProofInterface, beaconInst []string) (string, string) {
	// convert inst to byte array
	beaconInstBytes, _ := incProof.ConvertInstToBytes(beaconInst)

	// Split of last 32 bytes (block height)
	encodedBeaconHeight := hex.EncodeToString(beaconInstBytes[len(beaconInstBytes)-32:])
	encodedInst := hex.EncodeToString(beaconInstBytes[:len(beaconInstBytes)-32])
	return encodedInst, encodedBeaconHeight
}

// splitAndDecodeInst splits inc insts (on beacon and bridge) into 3 parts: the inst itself, bridgeHeight and beaconHeight that contains the inst
func splitAndEncodeIncInst(incProof blockchain.IncProofInterface, bridgeInst, beaconInst []string) (string, string, string) {
	// convert inst to byte array
	// todo: review
	bridgeInstFlat, _ := incProof.ConvertInstToBytes(bridgeInst)
	beaconInstFlat, _ := incProof.ConvertInstToBytes(beaconInst)

	// Split of last 32 bytes (block height)
	encodedBridgeHeight := hex.EncodeToString(bridgeInstFlat[len(bridgeInstFlat)-32:])
	encodedBeaconHeight := hex.EncodeToString(beaconInstFlat[len(beaconInstFlat)-32:])

	encodedInst := hex.EncodeToString(bridgeInstFlat[:len(bridgeInstFlat)-32])
	return encodedInst, encodedBridgeHeight, encodedBeaconHeight
}
