package portalprocess

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	"sort"
	"strconv"
)

func CollectPortalV4Insts(ppv4 map[int]PortalInstructionProcessorV4, metaType int, action []string, shardID byte) {
	switch metaType {
	// shield
	case metadata.PortalV4ShieldingRequestMeta:
		ppv4[metadata.PortalV4ShieldingRequestMeta].PutAction(action, shardID)

	default:
		return
	}
}

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

	return instructions, nil
}

func ProcessPortalInstsV4(
	portalStateDB *statedb.StateDB,
	portalParams portalv4.PortalParams,
	beaconHeight uint64,
	instructions [][]string,
	ppv4 map[int]PortalInstructionProcessorV4,
	epoch uint64) error {
	currentPortalState, err := InitCurrentPortalStateV4FromDB(portalStateDB)
	if err != nil {
		Logger.log.Error(err)
		return nil
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
	//todo: need to be reviewed
	// update info of bridge portal token
	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.CountUpAmt > updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.CountUpAmt - updatingInfo.DeductAmt
			updatingType = "+"
		}
		if updatingInfo.CountUpAmt < updatingInfo.DeductAmt {
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
			return err
		}
	}

	// store updated currentPortalState to leveldb with new beacon height
	err = StorePortalV4StateToDB(portalStateDB, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}
	return nil
}