package rpcserver

import (
	"encoding/hex"
	"fmt"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/pkg/errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
)

// handleGetBurnProof returns a proof of a tx burning pETH
func (httpServer *HttpServer) handleGetBurnProof(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Infof("handleGetBurnProof params: %+v", params)
	listParams, ok := params.([]interface{})
	if !ok || len(listParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	txIDParam, ok := listParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx id invalid"))
	}

	txID, err := common.Hash{}.NewHashFromStr(txIDParam)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	bc := httpServer.config.BlockChain
	db := *httpServer.config.Database

	// Get block height from txID
	height, err := httpServer.databaseService.GetBurningConfirm(*txID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, fmt.Errorf("proof of tx not found"))
	}

	// Get bridge block and corresponding beacon blocks
	bridgeBlock, beaconBlocks, err := getShardAndBeaconBlocks(height, bc, db)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Get proof of instruction on bridge
	bridgeInstProof, err := getBurnProofOnBridge(txID, bridgeBlock, db, httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Get proof of instruction on beacon
	beaconInstProof, err := getBurnProofOnBeacon(bridgeInstProof.inst, beaconBlocks, db, httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Decode instruction to send to Ethereum without having to decode on client
	decodedInst, bridgeHeight, beaconHeight := splitAndDecodeInst(bridgeInstProof.inst, beaconInstProof.inst)
	//decodedInst := hex.EncodeToString(blockchain.DecodeInstruction(bridgeInstProof.inst))

	return buildProofResult(decodedInst, beaconInstProof, bridgeInstProof, beaconHeight, bridgeHeight), nil
}

// getBurnProofOnBridge finds a beacon committee swap instruction in a given bridge block and returns its proof
func getBurnProofOnBridge(
	txID *common.Hash,
	bridgeBlock *blockchain.ShardBlock,
	db database.DatabaseInterface,
	ce ConsensusEngine,
) (*swapProof, error) {
	insts := bridgeBlock.Body.Instructions
	_, instID := findBurnConfirmInst(insts, txID)
	if instID < 0 {
		return nil, fmt.Errorf("cannot find burning instruction in bridge block")
	}

	block := &shardBlock{ShardBlock: bridgeBlock}
	proof, err := buildProofForBlock(block, insts, instID, db, ce)
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
	ce ConsensusEngine,
) (*swapProof, error) {
	// Get beacon block and check if it contains beacon swap instruction
	b, instID := findBeaconBlockWithBurnInst(beaconBlocks, inst)
	if b == nil {
		return nil, fmt.Errorf("cannot find corresponding beacon block that includes burn instruction")
	}

	insts := b.Body.Instructions
	block := &beaconBlock{BeaconBlock: b}
	return buildProofForBlock(block, insts, instID, db, ce)
}

// findBeaconBlockWithBurnInst finds a beacon block with a specific burning instruction and the instruction's index; nil if not found
func findBeaconBlockWithBurnInst(beaconBlocks []*blockchain.BeaconBlock, inst []string) (*blockchain.BeaconBlock, int) {
	for _, b := range beaconBlocks {
		for k, blkInst := range b.Body.Instructions {
			diff := false
			// Ignore block height (last element)
			for i, part := range inst[:len(inst)-1] {
				if i >= len(blkInst) || part != blkInst[i] {
					diff = true
					break
				}
			}
			if !diff {
				return b, k
			}
		}
	}
	return nil, -1
}

// findBurnConfirmInst finds a BurningConfirm instruction in a list, returns it along with its index
func findBurnConfirmInst(insts [][]string, txID *common.Hash) ([]string, int) {
	instType := strconv.Itoa(metadata.BurningConfirmMeta)
	for i, inst := range insts {
		if inst[0] != instType || len(inst) < 5 {
			continue
		}

		h, err := common.Hash{}.NewHashFromStr(inst[5])
		if err != nil {
			continue
		}

		if h.IsEqual(txID) {
			return inst, i
		}
	}
	return nil, -1
}

// splitAndDecodeInst splits BurningConfirm insts (on beacon and bridge) into 3 parts: the inst itself, bridgeHeight and beaconHeight that contains the inst
func splitAndDecodeInst(bridgeInst, beaconInst []string) (string, string, string) {
	// Decode instructions
	bridgeInstFlat, _ := blockchain.DecodeInstruction(bridgeInst)
	beaconInstFlat, _ := blockchain.DecodeInstruction(beaconInst)

	// Split of last 32 bytes (block height)
	bridgeHeight := hex.EncodeToString(bridgeInstFlat[len(bridgeInstFlat)-32:])
	beaconHeight := hex.EncodeToString(beaconInstFlat[len(beaconInstFlat)-32:])

	decodedInst := hex.EncodeToString(bridgeInstFlat[:len(bridgeInstFlat)-32])
	return decodedInst, bridgeHeight, beaconHeight
}
