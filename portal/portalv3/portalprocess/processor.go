package portalprocess

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
)

// interface for portal instruction processor v3
type PortalInstructionProcessorV3 interface {
	GetActions() map[byte][][]string
	PutAction(action []string, shardID byte)
	// get necessary db from stateDB to verify instructions when producing new block
	PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error)
	// validate and create new instructions in new beacon blocks
	BuildNewInsts(
		bc metadata.ChainRetriever,
		contentStr string,
		shardID byte,
		currentPortalState *CurrentPortalState,
		beaconHeight uint64,
		shardHeights map[byte]uint64,
		portalParams portalv3.PortalParams,
		optionalData map[string]interface{},
	) ([][]string, error)
	// process instructions that confirmed in beacon blocks
	ProcessInsts(
		stateDB *statedb.StateDB,
		beaconHeight uint64,
		instructions []string,
		currentPortalState *CurrentPortalState,
		portalParams portalv3.PortalParams,
		updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
	) error
}

type PortalInstProcessorV3 struct {
	Actions map[byte][][]string
}

func GetPortalInstProcessorByMetaType(pv3 map[int]PortalInstructionProcessorV3, metaType int) PortalInstructionProcessorV3 {
	// PortalRequestPortingMeta and PortalRequestPortingMetaV3 use the same processor
	if metaType == metadata.PortalRequestPortingMeta {
		metaType = metadata.PortalRequestPortingMetaV3
	}

	// PortalRedeemRequestMeta and PortalRedeemRequestMetaV3 use the same processor
	if metaType == metadata.PortalRedeemRequestMeta {
		metaType = metadata.PortalRedeemRequestMetaV3
	}

	// PortalRequestUnlockCollateralMeta and PortalRequestUnlockCollateralMetaV3 use the same processor
	if metaType == metadata.PortalRequestUnlockCollateralMeta {
		metaType = metadata.PortalRequestUnlockCollateralMetaV3
	}

	// PortalLiquidateCustodianMeta and PortalLiquidateCustodianMetaV3 use the same processor
	if metaType == metadata.PortalLiquidateCustodianMeta {
		metaType = metadata.PortalLiquidateCustodianMetaV3
	}

	return pv3[metaType]
}
