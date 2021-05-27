package portalprocess

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
)

// interface for portal instruction processor v3
type PortalInstructionProcessorV4 interface {
	GetActions() map[byte][][]string
	PutAction(action []string, shardID byte)
	// get necessary db from stateDB to verify instructions when producing new block
	PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error)
	// validate and create new instructions in new beacon blocks
	BuildNewInsts(
		bc metadata.ChainRetriever,
		contentStr string,
		shardID byte,
		currentPortalState *CurrentPortalStateV4,
		beaconHeight uint64,
		shardHeights map[byte]uint64,
		portalParams portalv4.PortalParams,
		optionalData map[string]interface{},
	) ([][]string, error)
	// process instructions that confirmed in beacon blocks
	ProcessInsts(
		stateDB *statedb.StateDB,
		beaconHeight uint64,
		instructions []string,
		currentPortalState *CurrentPortalStateV4,
		portalParams portalv4.PortalParams,
		updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
	) error
}

type PortalInstProcessorV4 struct {
	Actions map[byte][][]string
}