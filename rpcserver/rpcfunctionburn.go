package rpcserver

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

// handleGetBurnProof returns a proof of a tx burning pETH
func (rpcServer RpcServer) handleGetBurnProof(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBurnProof params: %+v", params)
	listParams := params.([]interface{})
	txID, err := common.NewHashFromStr(listParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	bc := rpcServer.config.BlockChain
	db := *rpcServer.config.Database

	// Get block height from txID
	height, err := db.GetBurningConfirm(txID[:])
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, fmt.Errorf("proof of tx not found"))
	}

	// Get bridge block and corresponding beacon blocks
	bridgeBlock, beaconBlocks, err := getShardAndBeaconBlocks(height, bc, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Get proof of instruction on bridge
	bridgeInstProof, err := getBurnProofOnBridge(txID, bridgeBlock, bc, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Get proof of instruction on beacon
	beaconInstProof, err := getBurnProofOnBeacon(bridgeInstProof.inst, beaconBlocks, db)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// Decode instruction to send to Ethereum without having to decode on client
	decodedInst := hex.EncodeToString(blockchain.DecodeInstruction(bridgeInstProof.inst))

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

// getBurnProofOnBridge finds a beacon committee swap instruction in a given bridge block and returns its proof
func getBurnProofOnBridge(
	txID *common.Hash,
	bridgeBlock *blockchain.ShardBlock,
	bc *blockchain.BlockChain,
	db database.DatabaseInterface,
) (*swapProof, error) {
	insts := bridgeBlock.Body.Instructions
	_, instID := findBurnConfirmInst(insts, txID)
	if instID < 0 {
		return nil, fmt.Errorf("cannot find burning instruction in bridge block")
	}

	proof, err := buildProofOnBridge(bridgeBlock, insts, instID, db)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// getBurnProofOnBeacon finds in given beacon blocks a BurningConfirm instruction and returns its proof
func getBurnProofOnBeacon(
	inst []string,
	beaconBlocks []*blockchain.BeaconBlock,
	db database.DatabaseInterface,
) (*swapProof, error) {
	// Get beacon block and check if it contains beacon swap instruction
	beaconBlock, instID := findBeaconBlockWithInst(beaconBlocks, inst)
	if beaconBlock == nil {
		return nil, fmt.Errorf("cannot find corresponding beacon block that includes burn instruction")
	}

	insts := beaconBlock.Body.Instructions
	return buildProofOnBeacon(beaconBlock, insts, instID, db)
}

// findBurnConfirmInst finds a BurningConfirm instruction in a list, returns it along with its index
func findBurnConfirmInst(insts [][]string, txID *common.Hash) ([]string, int) {
	instType := strconv.Itoa(metadata.BurningConfirmMeta)
	for i, inst := range insts {
		if inst[0] != instType {
			continue
		}

		h, err := common.NewHashFromStr(inst[len(inst)-1])
		if err != nil {
			continue
		}

		if h.IsEqual(txID) {
			return inst, i
		}
	}
	return nil, -1
}
