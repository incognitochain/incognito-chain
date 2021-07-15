package pdex

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type stateProcessorV2 struct {
	stateProcessorBase
}

func (sp *stateProcessorV2) modifyParams(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	params Params,
) (Params, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return params, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadata.PDexV3ParamsModifyingRequestContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return params, err
	}

	modifyingStatus := inst[2]
	var reqTrackStatus int
	if modifyingStatus == RequestAcceptedChainStatus {
		params = Params(actionData.Content)
		reqTrackStatus = ParamsModifyingSucceedStatus
	} else {
		reqTrackStatus = ParamsModifyingFailedStatus
	}

	modifyingReqStatus := metadata.PDexV3ParamsModifyingRequestStatus{
		Status:       reqTrackStatus,
		PDexV3Params: metadata.PDexV3Params(actionData.Content),
	}
	modifyingReqStatusBytes, _ := json.Marshal(modifyingReqStatus)
	err = statedb.TrackPDexV3Status(
		stateDB,
		statedb.PDexV3ParamsModifyingStatusPrefix(),
		[]byte(actionData.TxReqID.String()),
		modifyingReqStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("PDex Params Modifying: An error occurred while tracking modifying request tx - Error: %v", err)
	}

	return params, nil
}
