package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
)

/* =======
Portal Unshield Request Processor
======= */
type PortalUnshieldRequestProcessor struct {
	*PortalInstProcessorV4
}

func (p *PortalUnshieldRequestProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalUnshieldRequestProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalUnshieldRequestProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("[UnshieldRequest]: an error occured while decoding content string of portal unshield request action: %+v", err)
		return nil, fmt.Errorf("[UnshieldRequest]: an error occured while decoding content string of portal unshield request action: %+v", err)
	}
	var actionData metadata.PortalUnshieldRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("[UnshieldRequest]: an error occured while unmarshal portal unshield request action: %+v", err)
		return nil, fmt.Errorf("[UnshieldRequest]: an error occured while unmarshal portal unshield request action: %+v", err)
	}

	optionalData := make(map[string]interface{})

	// get unshield request by unshieldID from stateDB
	unshieldRequestStatusBytes, err := statedb.GetPortalUnshieldRequestStatus(stateDB, actionData.TxReqID.String())
	if err != nil {
		Logger.log.Errorf("[UnshieldRequest]: an error occurred while get unshield request by id from DB: %+v", err)
		return nil, fmt.Errorf("[UnshieldRequest]: an error occurred while get unshield request by id from DB: %+v", err)
	}

	optionalData["isExistUnshieldID"] = len(unshieldRequestStatusBytes) > 0
	return optionalData, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildUnshieldRequestInst(
	tokenID string,
	redeemAmount uint64,
	otaPubKeyStr string,
	txRandomStr string,
	remoteAddress string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	unshieldRequestContent := metadata.PortalUnshieldRequestContent{
		TokenID:        tokenID,
		UnshieldAmount: redeemAmount,
		OTAPubKeyStr:   otaPubKeyStr,
		TxRandomStr:    txRandomStr,
		RemoteAddress:  remoteAddress,
		TxReqID:        txReqID,
		ShardID:        shardID,
	}
	unshieldRequestContentBytes, _ := json.Marshal(unshieldRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(unshieldRequestContentBytes),
	}
}

func (p *PortalUnshieldRequestProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalStateV4 *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("[UnshieldRequest]: an error occured while decoding content string of portal unshield request action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalUnshieldRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("[UnshieldRequest]: an error occured while unmarshal portal unshield request action: %+v", err)
		return [][]string{}, nil
	}

	if currentPortalStateV4 == nil {
		Logger.log.Warn("[UnshieldRequest]: Current Portal state V4 is null.")
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildUnshieldRequestInst(
		meta.TokenID,
		meta.UnshieldAmount,
		meta.OTAPubKeyStr,
		meta.TxRandomStr,
		meta.RemoteAddress,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestRejectedChainStatus,
	)
	refundInst := buildUnshieldRequestInst(
		meta.TokenID,
		meta.UnshieldAmount,
		meta.OTAPubKeyStr,
		meta.TxRandomStr,
		meta.RemoteAddress,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestRefundedChainStatus,
	)

	unshieldID := actionData.TxReqID.String()
	tokenID := meta.TokenID

	// check unshieldID is existed in waitingUnshield list or not
	if currentPortalStateV4.WaitingUnshieldRequests != nil {
		wUnshieldReqsByTokenID := currentPortalStateV4.WaitingUnshieldRequests[tokenID]
		if wUnshieldReqsByTokenID != nil {
			keyWaitingUnshieldRequestStr := statedb.GenerateWaitingUnshieldRequestObjectKey(tokenID, unshieldID).String()
			waitingUnshieldRequest := wUnshieldReqsByTokenID[keyWaitingUnshieldRequestStr]
			if waitingUnshieldRequest != nil {
				Logger.log.Errorf("[UnshieldRequest] unshieldID is existed in waiting unshield requests list %v\n", unshieldID)
				return [][]string{rejectInst}, nil
			}
		}
	}

	// check unshieldID is existed in db or not
	if optionalData == nil {
		Logger.log.Errorf("[UnshieldRequest] optionalData is null")
		return [][]string{rejectInst}, nil
	}
	isExist, ok := optionalData["isExistUnshieldID"].(bool)
	if !ok {
		Logger.log.Errorf("[UnshieldRequest] optionalData isExistUnshieldID is invalid")
		return [][]string{rejectInst}, nil
	}
	if isExist {
		Logger.log.Errorf("[UnshieldRequest] UnshieldID exist in db %v", unshieldID)
		return [][]string{rejectInst}, nil
	}

	// validate unshielding amount
	if meta.UnshieldAmount < portalParams.MinUnshieldAmts[meta.TokenID] {
		Logger.log.Errorf("[UnshieldRequest] Unshield amount %v is less than min amount %v", meta.UnshieldAmount, portalParams.MinUnshieldAmts[meta.TokenID])
		return [][]string{refundInst}, nil
	}

	multipleTokenAmt := portalParams.PortalTokens[meta.TokenID].GetMultipleTokenAmount()
	if meta.UnshieldAmount%multipleTokenAmt != 0 {
		Logger.log.Errorf("[UnshieldRequest] Unshield amount %v is not divisible by %v", meta.UnshieldAmount, multipleTokenAmt)
		return [][]string{refundInst}, nil
	}

	// build accept instruction
	newInst := buildUnshieldRequestInst(
		meta.TokenID,
		meta.UnshieldAmount,
		meta.OTAPubKeyStr,
		meta.TxRandomStr,
		meta.RemoteAddress,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestAcceptedChainStatus,
	)

	// add new waiting unshield request to waiting list
	currentPortalStateV4.AddWaitingUnshieldRequest(unshieldID, meta.TokenID, meta.RemoteAddress, meta.UnshieldAmount, beaconHeight+1)

	return [][]string{newInst}, nil
}

func (p *PortalUnshieldRequestProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalStateV4 *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalStateV4 == nil {
		Logger.log.Errorf("[ProcessUnshieldRequest] current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalUnshieldRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("[ProcessUnshieldRequest] Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	var updatedStatus int
	reqStatus := instructions[2]
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		// update status
		updatedStatus = portalcommonv4.PortalUnshieldReqWaitingStatus

		// add new waiting unshield request to waiting list
		currentPortalStateV4.AddWaitingUnshieldRequest(
			actionData.TxReqID.String(), actionData.TokenID, actionData.RemoteAddress, actionData.UnshieldAmount, beaconHeight+1)

		// update bridge token info
		err := metadata.UpdatePortalBridgeTokenInfo(updatingInfoByTokenID, actionData.TokenID, actionData.UnshieldAmount, true)
		if err != nil {
			Logger.log.Errorf("[ProcessUnshieldRequest] Update Portal token info for UnshieldID - Error %v\n", actionData.TxReqID.String(), err)
			return nil
		}
	} else if reqStatus == portalcommonv4.PortalV4RequestRefundedChainStatus {
		updatedStatus = portalcommonv4.PortalUnshieldReqRefundedStatus
	}

	// store status of unshield request by unshieldID (txID)
	unshieldRequestStatus := metadata.PortalUnshieldRequestStatus{
		OTAPubKeyStr:   actionData.OTAPubKeyStr,
		TxRandomStr:    actionData.TxRandomStr,
		RemoteAddress:  actionData.RemoteAddress,
		TokenID:        actionData.TokenID,
		UnshieldAmount: actionData.UnshieldAmount,
		UnshieldID:     actionData.TxReqID.String(),
		ExternalTxID:   "",
		ExternalFee:    0,
		Status:         updatedStatus,
	}
	unshieldRequestStatusBytes, _ := json.Marshal(unshieldRequestStatus)
	err = statedb.StorePortalUnshieldRequestStatus(
		stateDB,
		actionData.TxReqID.String(),
		unshieldRequestStatusBytes)
	if err != nil {
		Logger.log.Errorf("[ProcessUnshieldRequest] Error when storing status of unshield request by unshieldID: %v\n", err)
		return nil
	}

	return nil
}
