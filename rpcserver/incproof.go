package rpcserver

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"math/big"
	"strconv"
)

type IncProofInterface interface {
	findConfirmInst(insts [][]string, txID *common.Hash) ([]string, int)
	findBeaconBlockWithConfirmInst(beaconBlocks []*blockchain.BeaconBlock, inst []string) (*blockchain.BeaconBlock, int)
}

type IncProof struct {
	metaType int
}

//TODO: convert other proofs (burn proof and swap beacon proof) to interface
func NewIncProof(metaType int) IncProofInterface {
	switch metaType {
	case metadata.PortalCustodianWithdrawConfirmMetaV3:
		return PortalWithdrawCollateralProof{
			&IncProof{
				metaType: metaType,
			},
		}
	default:
		return nil
	}
}

type PortalWithdrawCollateralProof struct {
	*IncProof
}

// findConfirmInst finds a specific instruction in a list, returns it along with its index
func (withdrawProof PortalWithdrawCollateralProof) findConfirmInst(insts [][]string, txID *common.Hash) ([]string, int) {
	for i, inst := range insts {
		if inst[0] != strconv.Itoa(withdrawProof.metaType) || len(inst) < 7 {
			continue
		}

		h, err := common.Hash{}.NewHashFromStr(inst[6])
		if err != nil {
			continue
		}

		if h.IsEqual(txID) {
			return inst, i
		}
	}
	return nil, -1
}

// findBeaconBlockWithConfirmInst finds a beacon block with a specific burning instruction and the instruction's index; nil if not found
func (withdrawProof PortalWithdrawCollateralProof) findBeaconBlockWithConfirmInst(beaconBlocks []*blockchain.BeaconBlock, inst []string) (*blockchain.BeaconBlock, int) {
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

func retrieveIncProof(
	confirmMetaType int,
	onBeacon bool,
	height uint64,
	txID *common.Hash,
	httpServer *HttpServer,
) (interface{}, *rpcservice.RPCError) {
	incProof := NewIncProof(confirmMetaType)
	if !onBeacon {
		return getIncProofByHeight(incProof, httpServer, height, txID)
	}
	return getIncProofByHeightV2(incProof, httpServer, height, txID)
}

func getIncProofByHeightV2(
	incProof IncProofInterface,
	httpServer *HttpServer,
	height uint64,
	txID *common.Hash,
) (interface{}, *rpcservice.RPCError) {
	// Get beacon block
	beaconBlock, err := getSingleBeaconBlockByHeight(httpServer.GetBlockchain(), height)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Get proof of instruction on beacon
	inst, instID := incProof.findConfirmInst(beaconBlock.Body.Instructions, txID)
	if instID == -1 {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, fmt.Errorf("cannot find inst %s in beacon block %d", txID.String(), height))
	}

	beaconInstProof, err := getIncProofOnBeacon(incProof, inst, []*blockchain.BeaconBlock{beaconBlock}, httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Decode instruction to send to Ethereum without having to decode on client
	decodedInst := splitAndDecodeInstV3(beaconInstProof.inst)
	bcHeightStr := base58.Base58Check{}.Encode(new(big.Int).SetUint64(height).Bytes(), 0x00)
	return buildProofResult(decodedInst, beaconInstProof, nil, bcHeightStr, ""), nil
}

func getIncProofByHeight(
	incProof IncProofInterface,
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
	decodedInst, bridgeHeight, beaconHeight := splitAndDecodeInst(bridgeInstProof.inst, beaconInstProof.inst)
	fmt.Println("decodedInst", decodedInst)
	fmt.Println("bridgeHeight", bridgeHeight)
	fmt.Println("beaconHeight", beaconHeight)
	//decodedInst := hex.EncodeToString(blockchain.DecodeInstruction(bridgeInstProof.inst))

	return buildProofResult(decodedInst, beaconInstProof, bridgeInstProof, beaconHeight, bridgeHeight), nil
}

// getIncProofOnBridge finds a beacon committee swap instruction in a given bridge block and returns its proof
func getIncProofOnBridge(
	incProof IncProofInterface,
	txID *common.Hash,
	bridgeBlock *blockchain.ShardBlock,
	ce ConsensusEngine,
) (*swapProof, error) {
	insts := bridgeBlock.Body.Instructions
	_, instID := incProof.findConfirmInst(insts, txID)
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
	incProof IncProofInterface,
	inst []string,
	beaconBlocks []*blockchain.BeaconBlock,
	ce ConsensusEngine,
) (*swapProof, error) {
	// Get beacon block and check if it contains beacon swap instruction
	b, instID := incProof.findBeaconBlockWithConfirmInst(beaconBlocks, inst)
	if b == nil {
		return nil, fmt.Errorf("cannot find corresponding beacon block that includes burn instruction")
	}

	insts := b.Body.Instructions
	block := &beaconBlock{BeaconBlock: b}
	return buildProofForBlock(block, insts, instID, ce)
}