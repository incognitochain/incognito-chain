package portalprocess

import (
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
)

func buildNewPortalV4InstsFromActions(
	p PortalInstructionProcessorV4,
	bc metadata.ChainRetriever,
	stateDB *statedb.StateDB,
	currentPortalState *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams) ([][]string, error) {

	instructions := [][]string{}
	actions := p.GetActions()
	var shardIDKeys []int
	for k := range actions {
		shardIDKeys = append(shardIDKeys, int(k))
	}

	sort.Ints(shardIDKeys)
	for _, value := range shardIDKeys {
		shardID := byte(value)
		actions := actions[shardID]
		for _, action := range actions {
			contentStr := action[1]
			optionalData, err := p.PrepareDataForBlockProducer(stateDB, contentStr)
			if err != nil {
				Logger.log.Errorf("Error when preparing data before processing instruction %+v", err)
				continue
			}
			newInst, err := p.BuildNewInsts(
				bc,
				contentStr,
				shardID,
				currentPortalState,
				beaconHeight,
				shardHeights,
				portalParams,
				optionalData,
			)
			if err != nil {
				Logger.log.Errorf("Error when building new instructions : %v", err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	return instructions, nil
}

// handle portal instructions for block producer
func HandlePortalInstsV4(
	bc metadata.ChainRetriever,
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	currentPortalState *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	ppv4 map[int]PortalInstructionProcessorV4,
) ([][]string, error) {
	instructions := [][]string{}

	// producer portal instructions for actions from shards
	// sort metadata type map to make it consistent for every run
	var metaTypes []int
	for metaType := range ppv4 {
		metaTypes = append(metaTypes, metaType)
	}
	sort.Ints(metaTypes)
	for _, metaType := range metaTypes {
		actions := ppv4[metaType]
		newInst, err := buildNewPortalV4InstsFromActions(
			actions,
			bc,
			stateDB,
			currentPortalState,
			beaconHeight,
			shardHeights,
			portalParams)

		if err != nil {
			Logger.log.Error(err)
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}

	// handle intervally
	// handle batching process unshield requests
	if (beaconHeight+1)%uint64(portalParams.BatchNumBlks) == 0 {
		batchUnshieldInsts, err := handleBatchingUnshieldRequests(bc, stateDB, beaconHeight, shardHeights, currentPortalState, portalParams, ppv4)
		if err != nil {
			Logger.log.Error(err)
		}
		if len(batchUnshieldInsts) > 0 {
			instructions = append(instructions, batchUnshieldInsts...)
		}
	}

	return instructions, nil
}

func hasPortalV4Instruction(instructions [][]string) bool {
	hasPortalV4Instruction := false
	for _, inst := range instructions {
		if len(inst) < 4 {
			continue // Not error, just not Portal v4 instruction
		}
		switch inst[0] {
		case strconv.Itoa(metadata.PortalV4ShieldingRequestMeta):
			hasPortalV4Instruction = true
			break
		case strconv.Itoa(metadata.PortalV4ShieldingResponseMeta):
			hasPortalV4Instruction = true
			break
		case strconv.Itoa(metadata.PortalV4UnshieldingRequestMeta):
			hasPortalV4Instruction = true
			break
		case strconv.Itoa(metadata.PortalV4UnshieldBatchingMeta):
			hasPortalV4Instruction = true
			break
		case strconv.Itoa(metadata.PortalV4FeeReplacementRequestMeta):
			hasPortalV4Instruction = true
			break
		case strconv.Itoa(metadata.PortalV4SubmitConfirmedTxMeta):
			hasPortalV4Instruction = true
			break
		}
	}
	return hasPortalV4Instruction
}

func ProcessPortalInstsV4(
	portalStateDB *statedb.StateDB,
	lastState *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	beaconHeight uint64,
	instructions [][]string,
	ppv4 map[int]PortalInstructionProcessorV4,
	epoch uint64) (*CurrentPortalStateV4, error) {

	if !hasPortalV4Instruction(instructions) {
		return lastState, nil
	}

	currentPortalState, err := InitCurrentPortalStateV4FromDB(portalStateDB, lastState)
	if err != nil {
		Logger.log.Error(err)
		return currentPortalState, nil
	}

	// re-use update info of bridge
	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}

	for _, inst := range instructions {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error
		metaType, _ := strconv.Atoi(inst[0])
		processor := ppv4[metaType]
		if processor != nil {
			err = processor.ProcessInsts(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
			if err != nil {
				Logger.log.Errorf("Process portal instruction err: %v, inst %+v", err, inst)
			}
			continue
		}
	}

	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.CountUpAmt > updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.CountUpAmt - updatingInfo.DeductAmt
			updatingType = "+"
		} else if updatingInfo.CountUpAmt < updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.DeductAmt - updatingInfo.CountUpAmt
			updatingType = "-"
		}
		err := statedb.UpdateBridgeTokenInfo(
			portalStateDB,
			updatingInfo.TokenID,
			updatingInfo.ExternalTokenID,
			updatingInfo.IsCentralized,
			updatingAmt,
			updatingType,
		)
		if err != nil {
			return currentPortalState, err
		}
	}
	return currentPortalState, nil
}

func handleBatchingUnshieldRequests(
	bc metadata.ChainRetriever,
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	currentPortalState *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	ppv4 map[int]PortalInstructionProcessorV4) ([][]string, error) {
	return ppv4[metadata.PortalV4UnshieldBatchingMeta].BuildNewInsts(
		bc, "", 0, currentPortalState, beaconHeight, shardHeights, portalParams, nil)
}
