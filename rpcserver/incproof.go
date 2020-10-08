package rpcserver

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"math/big"
	"strconv"
)

type IncProofInterface interface {
	findConfirmInst(insts [][]string, txID *common.Hash) ([]string, int)
	findBeaconBlockWithConfirmInst(beaconBlocks []*blockchain.BeaconBlock, inst []string) (*blockchain.BeaconBlock, int)
	convertInstToBytes(inst []string) ([]byte, error)
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
		if inst[0] != strconv.Itoa(withdrawProof.metaType) || len(inst) < 8 {
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

// findBeaconBlockWithConfirmInst finds a beacon block with a specific incognito instruction and the instruction's index; nil if not found
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

func (withdrawProof PortalWithdrawCollateralProof) convertInstToBytes(inst []string) ([]byte, error) {
	if len(inst) < 8 {
		return nil, errors.New("invalid length of WithdrawCollateralConfirm inst")
	}

	m, _ := strconv.Atoi(inst[0])
	metaType := byte(m)
	s, _ := strconv.Atoi(inst[1])
	shardID := byte(s)
	cusPaymentAddress := []byte(inst[2])
	externalAddress, err := common.DecodeETHAddr(inst[3])
	if err != nil {
		Logger.log.Errorf("Decode external address error: ", err)
		return nil, err
	}
	externalTokenID, err := common.DecodeETHAddr(inst[4])
	if err != nil {
		Logger.log.Errorf("Decode externalTokenID error: ", err)
		return nil, err
	}
	amount, _ := new(big.Int).SetString(inst[5], 10)
	amountBytes := common.AddPaddingBigInt(amount, 32)

	txIDStr := inst[6]
	txID, _ := common.Hash{}.NewHashFromStr(txIDStr)

	beaconHeightStr := inst[7]
	bcHeightBN, _ := new(big.Int).SetString(beaconHeightStr, 10)
	bcHeightBytes := common.AddPaddingBigInt(bcHeightBN, 32)

	//Logger.log.Errorf("metaType: %v", metaType)
	//Logger.log.Errorf("shardID: %v", shardID)
	//Logger.log.Errorf("cusPaymentAddress: %v - %v", cusPaymentAddress, len(cusPaymentAddress))
	//Logger.log.Errorf("externalAddress: %v - %v", externalAddress, len(externalAddress))
	//Logger.log.Errorf("externalTokenID: %v - %v", externalTokenID, len(externalTokenID))
	//Logger.log.Errorf("amountBytes: %v - %v", amountBytes, len(amountBytes))
	//Logger.log.Errorf("txID: %v - %v", txID[:])

	//BLogger.log.Infof("Decoded WithdrawCollateralConfirm inst, amount: %d, remoteAddr: %x, externalTokenID: %x", amount, externalAddress, externalTokenID)
	flatten := []byte{}
	flatten = append(flatten, metaType)
	flatten = append(flatten, shardID)
	flatten = append(flatten, cusPaymentAddress...)
	flatten = append(flatten, externalAddress...)
	flatten = append(flatten, externalTokenID...)
	flatten = append(flatten, amountBytes...)
	flatten = append(flatten, txID[:]...)
	flatten = append(flatten, bcHeightBytes...)
	Logger.log.Errorf("flatten: %v - %v", flatten, len(flatten))
	return flatten, nil
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

	// find index of the confirm instruction in list of instructions
	inst, instID := incProof.findConfirmInst(beaconBlock.Body.Instructions, txID)
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
	decodedInst, bridgeHeight, beaconHeight := splitAndEncodeIncInst(incProof, bridgeInstProof.inst, beaconInstProof.inst)
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

// splitAndDecodeInst splits inc insts into 2 parts: the inst itself and beaconHeight that contains the inst
func splitAndEncodeIncInstV2(incProof IncProofInterface, beaconInst []string) (string, string) {
	// convert inst to byte array
	beaconInstBytes, _ := incProof.convertInstToBytes(beaconInst)

	// Split of last 32 bytes (block height)
	encodedBeaconHeight := hex.EncodeToString(beaconInstBytes[len(beaconInstBytes)-32:])
	encodedInst := hex.EncodeToString(beaconInstBytes[:len(beaconInstBytes)-32])
	return encodedInst, encodedBeaconHeight
}

// splitAndDecodeInst splits inc insts (on beacon and bridge) into 3 parts: the inst itself, bridgeHeight and beaconHeight that contains the inst
func splitAndEncodeIncInst(incProof IncProofInterface, bridgeInst, beaconInst []string) (string, string, string) {
	// convert inst to byte array
	// todo: review
	bridgeInstFlat, _ := incProof.convertInstToBytes(bridgeInst)
	beaconInstFlat, _ := incProof.convertInstToBytes(beaconInst)

	// Split of last 32 bytes (block height)
	encodedBridgeHeight := hex.EncodeToString(bridgeInstFlat[len(bridgeInstFlat)-32:])
	encodedBeaconHeight := hex.EncodeToString(beaconInstFlat[len(beaconInstFlat)-32:])

	encodedInst := hex.EncodeToString(bridgeInstFlat[:len(bridgeInstFlat)-32])
	return encodedInst, encodedBridgeHeight, encodedBeaconHeight
}
