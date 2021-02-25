package portal

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
)

func CollectPortalInstructions(pm *PortalManager, metaType int, action []string, shardID byte) bool {
	isCollected := true
	if metadata.IsPortalRelayingMetaType(metaType) {
		portalrelaying.CollectPortalRelayingInsts(pm.RelayingChainsProcessors, metaType, action, shardID)
	} else if metadata.IsPortalMetaTypeV3(metaType) {
		portalprocessv3.CollectPortalInstsV3(pm.PortalInstProcessorsV3, metaType, action, shardID)
	} else {
		isCollected = false
	}

	return isCollected
}

func HandlePortalInsts(
	bc metadata.ChainRetriever,
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	shardHeight map[byte]uint64,
	currentPortalState *portalprocessv3.CurrentPortalState,
	relayingState *portalrelaying.RelayingHeaderChainState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParams PortalParams,
	pm *PortalManager,
) ([][]string, error) {
	instructions := [][]string{}

	// handle portal instructions v3
	portalInstsV3, err := portalprocessv3.HandlePortalInstsV3(
		bc, stateDB, beaconHeight, shardHeight, currentPortalState, rewardForCustodianByEpoch,
		portalParams.GetPortalParamsV3(beaconHeight), pm.PortalInstProcessorsV3)
	if err != nil {
		Logger.log.Error(err)
	}
	if len(portalInstsV3) > 0 {
		instructions = append(instructions, portalInstsV3...)
	}

	// Handle relaying instruction
	relayingInsts := portalrelaying.HandleRelayingInsts(bc, relayingState, pm.RelayingChainsProcessors)
	if err != nil {
		Logger.log.Error(err)
	}
	if len(relayingInsts) > 0 {
		instructions = append(instructions, relayingInsts...)
	}

	// Handle next things ...

	return instructions, nil
}

func ProcessPortalInsts(
	portalStateDB *statedb.StateDB,
	relayingState *portalrelaying.RelayingHeaderChainState,
	portalParams PortalParams,
	beaconHeight uint64,
	instructions [][]string,
	pm *PortalManager,
	epoch uint64,
	isSkipPortalV3Ints bool) error {
	// process portal instructions v3
	if !isSkipPortalV3Ints {
		err := portalprocessv3.ProcessPortalInstsV3(
			portalStateDB, portalParams.GetPortalParamsV3(beaconHeight),
			beaconHeight, instructions, pm.PortalInstProcessorsV3, epoch)
		if err != nil {
			Logger.log.Error(err)
			return err
		}
	}

	// process relaying instructions
	err := portalrelaying.ProcessRelayingInstructions(instructions, relayingState)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	// Handle next things ...

	return nil
}


