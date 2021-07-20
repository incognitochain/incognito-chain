package pdex

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type stateProcessorV2 struct {
	stateProcessorBase
}

func (sp *stateProcessorV2) addLiquidity(
	insts [][]string,
) error {
	for _, inst := range insts {
		switch inst[1] {
		case instruction.WaitingStatus:
			waitingAddLiquidityInst := instruction.WaitingAddLiquidity{}
			err := waitingAddLiquidityInst.FromStringArr(inst)
			//TODO: Update state with current instruction
			if err != nil {
				return err
			}
		}
	}
	return nil
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
	var actionData metadataPdexV3.PDexV3ParamsModifyingRequestContent
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

	modifyingReqStatus := metadataPdexV3.PDexV3ParamsModifyingRequestStatus{
		Status:       reqTrackStatus,
		PDexV3Params: metadataPdexV3.PDexV3Params(actionData.Content),
	}
	modifyingReqStatusBytes, _ := json.Marshal(modifyingReqStatus)
	err = statedb.TrackPDexV3Status(
		stateDB,
		statedb.PDexV3ParamsModifyingStatusPrefix(),
		[]byte(actionData.TxReqID.String()),
		modifyingReqStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("PDex Params Modifying: An error occurred while tracking shielding request tx - Error: %v", err)
	}

	return params, nil
}
