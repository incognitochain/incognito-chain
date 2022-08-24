package rpcserver

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/pkg/errors"
)

// handleGetBurnProof returns a proof of a tx burning pETH
func (httpServer *HttpServer) handleGetBurnProof(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningConfirmMeta
	if onBeacon {
		confirmMeta = metadata.BurningConfirmMetaV2
	}
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetBSCBurnProof returns a proof of a tx burning pBSC
func (httpServer *HttpServer) handleGetBSCBurnProof(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningBSCConfirmMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetPRVERC20BurnProof returns a proof of a tx burning prv erc20
func (httpServer *HttpServer) handleGetPRVERC20BurnProof(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningPRVERC20ConfirmMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetPRVBEP20BurnProof returns a proof of a tx burning prv bep20
func (httpServer *HttpServer) handleGetPRVBEP20BurnProof(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningPRVBEP20ConfirmMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetBurnProofForDepositToSC returns a proof of a tx burning pETH to deposit to SC
func (httpServer *HttpServer) handleGetBurnProofForDepositToSC(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningConfirmForDepositToSCMeta
	if onBeacon {
		confirmMeta = metadata.BurningConfirmForDepositToSCMetaV2
	}
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetBurnPBSCProofForDepositToSC returns a proof of a tx burning pBSC to deposit to SC
func (httpServer *HttpServer) handleGetBurnPBSCProofForDepositToSC(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningConfirmForDepositToSCMeta
	if onBeacon {
		confirmMeta = metadata.BurningPBSCConfirmForDepositToSCMeta
	}
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetPLGBurnProof returns a proof of a tx burning pPLG ( polygon )
func (httpServer *HttpServer) handleGetPLGBurnProof(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningPLGConfirmMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetBurnPLGProofForDepositToSC returns a proof of a tx burning pPLG to deposit to SC
func (httpServer *HttpServer) handleGetBurnPLGProofForDepositToSC(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningPLGConfirmForDepositToSCMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetFTMBurnProof returns a proof of a tx burning pFTM ( Fantom )
func (httpServer *HttpServer) handleGetFTMBurnProof(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningFantomConfirmMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetBurnFTMProofForDepositToSC returns a proof of a tx burning pFTM to deposit to SC
func (httpServer *HttpServer) handleGetBurnFTMProofForDepositToSC(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningFantomConfirmForDepositToSCMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetAURORABurnProof returns a proof of a tx burning pPLG ( polygon )
func (httpServer *HttpServer) handleGetAURORABurnProof(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningAuroraConfirmMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetAVAXBurnProof returns a proof of a tx burning pPLG ( polygon )
func (httpServer *HttpServer) handleGetAVAXBurnProof(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningAvaxConfirmMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetBurnAURORAProofForDepositToSC returns a proof of a tx burning pFTM to deposit to SC
func (httpServer *HttpServer) handleGetBurnAURORAProofForDepositToSC(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningAuroraConfirmForDepositToSCMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

// handleGetBurnAVAXProofForDepositToSC returns a proof of a tx burning pFTM to deposit to SC
func (httpServer *HttpServer) handleGetBurnAVAXProofForDepositToSC(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	onBeacon, height, txID, err := parseGetBurnProofParams(params, httpServer)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	confirmMeta := metadata.BurningAvaxConfirmForDepositToSCMeta
	return retrieveBurnProof(confirmMeta, onBeacon, height, txID, httpServer, true)
}

func parseGetBurnProofParams(params interface{}, httpServer *HttpServer) (bool, uint64, *common.Hash, error) {
	listParams, ok := params.([]interface{})
	if !ok || len(listParams) < 1 {
		return false, 0, nil, errors.New("param must be an array at least 1 element")
	}

	txIDParam, ok := listParams[0].(string)
	if !ok {
		return false, 0, nil, errors.New("Tx id invalid")
	}

	txID, err := common.Hash{}.NewHashFromStr(txIDParam)
	if err != nil {
		return false, 0, nil, err
	}
	// Get block height from txID
	height, onBeacon, err := httpServer.blockService.GetBurningConfirm(*txID)
	if err != nil {
		return false, 0, nil, fmt.Errorf("proof of tx not found")
	}
	return onBeacon, height, txID, nil
}

func retrieveBurnProof(
	confirmMeta int,
	onBeacon bool,
	height uint64,
	txID *common.Hash,
	httpServer *HttpServer,
	shouldValidateBurningMetaType bool,
) (interface{}, *rpcservice.RPCError) {
	if !onBeacon {
		return getBurnProofByHeight(confirmMeta, httpServer, height, txID, shouldValidateBurningMetaType)
	}
	return getBurnProofByHeightV2(confirmMeta, httpServer, height, txID, shouldValidateBurningMetaType)
}

func getBurnProofByHeightV2(
	burningMetaType int,
	httpServer *HttpServer,
	height uint64,
	txID *common.Hash,
	shouldValidateBurningMetaType bool,
) (interface{}, *rpcservice.RPCError) {
	// Get beacon block
	beaconBlock, err := getSingleBeaconBlockByHeight(httpServer.GetBlockchain(), height)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Get proof of instruction on beacon
	inst, instID := findBurnConfirmInst(burningMetaType, beaconBlock.Body.Instructions, txID, shouldValidateBurningMetaType)
	if instID == -1 {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, fmt.Errorf("cannot find tx %s in beacon block %d", txID.String(), height))
	}

	beaconInstProof, err := getBurnProofOnBeacon(inst, []*types.BeaconBlock{beaconBlock}, httpServer.config.ConsensusEngine)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	// Decode instruction to send to Ethereum without having to decode on client
	decodedInst, beaconHeight := splitAndDecodeInstV2(beaconInstProof.inst)
	return buildProofResult(decodedInst, beaconInstProof, nil, beaconHeight, ""), nil
}

func getBurnProofByHeight(
	burningMetaType int,
	httpServer *HttpServer,
	height uint64,
	txID *common.Hash,
	shouldValidateBurningMetaType bool,
) (interface{}, *rpcservice.RPCError) {

	// Get bridge block and corresponding beacon blocks
	bridgeBlock, beaconBlocks, err := getShardAndBeaconBlocks(height, httpServer.GetBlockchain())
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	// Get proof of instruction on bridge
	bridgeInstProof, err := getBurnProofOnBridge(burningMetaType, txID, bridgeBlock, httpServer.config.ConsensusEngine, shouldValidateBurningMetaType)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	fmt.Println("bridgeInstProof", bridgeInstProof)
	// Get proof of instruction on beacon
	beaconInstProof, err := getBurnProofOnBeacon(bridgeInstProof.inst, beaconBlocks, httpServer.config.ConsensusEngine)
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

// getBurnProofOnBridge finds a beacon committee swap instruction in a given bridge block and returns its proof
func getBurnProofOnBridge(
	burningMetaType int,
	txID *common.Hash,
	bridgeBlock *types.ShardBlock,
	ce ConsensusEngine,
	shouldValidateBurningMetaType bool,
) (*swapProof, error) {
	insts := bridgeBlock.Body.Instructions
	_, instID := findBurnConfirmInst(burningMetaType, insts, txID, shouldValidateBurningMetaType)
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

// getBurnProofOnBeacon finds in given beacon blocks a BurningConfirm instruction and returns its proof
func getBurnProofOnBeacon(
	inst []string,
	beaconBlocks []*types.BeaconBlock,
	ce ConsensusEngine,
) (*swapProof, error) {
	// Get beacon block and check if it contains beacon swap instruction
	b, instID := findBeaconBlockWithBurnInst(beaconBlocks, inst)
	if b == nil {
		return nil, fmt.Errorf("cannot find corresponding beacon block that includes burn instruction")
	}

	insts := b.Body.Instructions
	block := &beaconBlock{BeaconBlock: b}
	return buildProofForBlock(block, insts, instID, ce)
}

// findBeaconBlockWithBurnInst finds a beacon block with a specific burning instruction and the instruction's index; nil if not found
func findBeaconBlockWithBurnInst(beaconBlocks []*types.BeaconBlock, inst []string) (*types.BeaconBlock, int) {
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
func findBurnConfirmInst(
	burningMetaType int,
	insts [][]string,
	txID *common.Hash,
	shouldValidateBurningMetaType bool,
) ([]string, int) {
	instType := strconv.Itoa(burningMetaType)
	for i, inst := range insts {
		if shouldValidateBurningMetaType {
			if inst[0] != instType || len(inst) < 5 {
				continue
			}
		} else {
			metaType, err := strconv.Atoi(inst[0])
			if err != nil {
				Logger.log.Warnf("Cannot find burning confirm instruction err %v", err)
				continue
			}
			if !metadataBridge.IsBurningConfirmMetaType(metaType) {
				continue
			}
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

// splitAndDecodeInst splits BurningConfirm insts (on beacon and bridge) into 2 parts: the inst itself and beaconHeight that contains the inst
func splitAndDecodeInstV2(beaconInst []string) (string, string) {
	// Decode instructions
	beaconInstFlat, _ := blockchain.DecodeInstruction(beaconInst)

	// Split of last 32 bytes (block height)
	beaconHeight := hex.EncodeToString(beaconInstFlat[len(beaconInstFlat)-32:])
	decodedInst := hex.EncodeToString(beaconInstFlat[:len(beaconInstFlat)-32])
	return decodedInst, beaconHeight
}

// handleGetBurningAddress returns a burningAddress
func (httpServer *HttpServer) handleGetBurningAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	listParams, ok := params.([]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array"))
	}

	beaconHeightParam := float64(0)
	if len(listParams) >= 1 {
		beaconHeightParam, ok = listParams[0].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("beacon height is invalid"))
		}
	}

	burningAddress := httpServer.blockService.GetBurningAddress(uint64(beaconHeightParam))

	return burningAddress, nil
}

// Notice: this function is used when getting proof
// make sure it will get from final view
func getSingleBeaconBlockByHeight(bc *blockchain.BlockChain, height uint64) (*types.BeaconBlock, error) {
	beaconBlock, err := bc.GetBeaconBlockByView(bc.BeaconChain.GetFinalView(), height)
	if err != nil {
		return nil, fmt.Errorf("cannot find beacon block with height %d %w", height, err)
	}
	return beaconBlock, nil
}
