package portalprocess

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/portal/portalv4"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
)

/* =======
Portal Shielding Request Processor V4
======= */

type PortalShieldingRequestProcessor struct {
	*PortalInstProcessorV4
}

func (p *PortalShieldingRequestProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalShieldingRequestProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalShieldingRequestProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Shielding request: an error occurred while decoding content string of pToken request action: %+v", err)
		return nil, fmt.Errorf("Shielding request: an error occurred while decoding content string of pToken request action: %+v", err)
	}

	var actionData metadata.PortalShieldingRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Shielding request: an error occurred while unmarshal shielding request action: %+v", err)
		return nil, fmt.Errorf("Shielding request: an error occurred while unmarshal shielding request action: %+v", err)
	}

	proofHash := hashProof(actionData.Meta.ShieldingProof)

	isExistProofTxHash, err := statedb.IsExistsShieldingRequest(stateDB, actionData.Meta.TokenID, proofHash)
	if err != nil {
		Logger.log.Errorf("Shielding request: an error occurred while get pToken request proof from DB: %+v", err)
		return nil, fmt.Errorf("Shielding request: an error occurred while get pToken request proof from DB: %+v", err)
	}

	optionalData := make(map[string]interface{})
	optionalData["isExistProofTxHash"] = isExistProofTxHash
	return optionalData, nil
}

func hashProof(proof string) string {
	hash := sha256.Sum256([]byte(proof))
	return fmt.Sprintf("%x", hash[:])
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildReqPTokensInstV4(
	tokenID string,
	incogAddressStr string,
	proofHash string,
	shieldingUTXO []*statedb.UTXO,
	mintingAmt uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	shieldingReqContent := metadata.PortalShieldingRequestContent{
		TokenID:         tokenID,
		IncogAddressStr: incogAddressStr,
		ProofHash:       proofHash,
		ShieldingUTXO:   shieldingUTXO,
		MintingAmount:   mintingAmt,
		TxReqID:         txReqID,
		ShardID:         shardID,
	}
	shieldingReqContentBytes, _ := json.Marshal(shieldingReqContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(shieldingReqContentBytes),
	}
}

func (p *PortalShieldingRequestProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalStateV4,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv4.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal shielding request action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalShieldingRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal shielding request action: %+v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	rejectInst := buildReqPTokensInstV4(
		meta.TokenID,
		meta.IncogAddressStr,
		"",
		[]*statedb.UTXO{},
		0,
		meta.Type,
		shardID,
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Shielding Request: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	portalTokenProcessor := portalParams.PortalTokens[meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("TokenID is not supported currently on Portal")
		return [][]string{rejectInst}, nil
	}

	// check unique external proof from optionalData which get from statedb
	if optionalData == nil {
		Logger.log.Errorf("Shielding Request: optionalData is null")
		return [][]string{rejectInst}, nil
	}
	isExistInStateDB, ok := optionalData["isExistProofTxHash"].(bool)
	if !ok {
		Logger.log.Errorf("Shielding Request: optionalData isExistProofTxHash is invalid")
		return [][]string{rejectInst}, nil
	}

	proofHash := hashProof(meta.ShieldingProof)

	// check unique external proof from portal state
	if IsExistsProofInPortalState(currentPortalState, meta.TokenID, proofHash) || isExistInStateDB {
		Logger.log.Errorf("Shielding Request: Shielding request proof exist in db %v", meta.ShieldingProof)
		return [][]string{rejectInst}, nil
	}

	// generate expected multisig address from master pubkeys and user payment address
	_, expectedReceivedMultisigAddress, err := portalTokenProcessor.GenerateOTMultisigAddress(bc, portalParams.MasterPubKeys[meta.TokenID], int(portalParams.NumRequiredSigs), meta.IncogAddressStr)
	isValid, listUTXO, err := portalTokenProcessor.ParseAndVerifyShieldProof(meta.ShieldingProof, bc, expectedReceivedMultisigAddress, meta.IncogAddressStr)
	if !isValid || err != nil {
		Logger.log.Error("Parse proof and verify shielding proof failed: %v", err)
		return [][]string{rejectInst}, nil
	}

	UpdatePortalStateUTXOs(currentPortalState, meta.TokenID, listUTXO)
	shieldingAmount := uint64(0)
	for _, utxo := range listUTXO {
		shieldingAmount += utxo.GetOutputAmount()
	}
	UpdatePortalStateShieldingExternalTx(currentPortalState, meta.TokenID, proofHash, listUTXO[0].GetTxHash(), meta.IncogAddressStr, shieldingAmount)

	mintingAmount := portalTokenProcessor.ConvertExternalToIncAmount(shieldingAmount)

	inst := buildReqPTokensInstV4(
		actionData.Meta.TokenID,
		actionData.Meta.IncogAddressStr,
		proofHash,
		listUTXO,
		mintingAmount,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		portalcommonv4.PortalV4RequestAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

func (p *PortalShieldingRequestProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalStateV4,
	portalParams portalv4.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalShieldingRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error: %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		UpdatePortalStateUTXOs(currentPortalState, actionData.TokenID, actionData.ShieldingUTXO)
		shieldingExternalTxHash := actionData.ShieldingUTXO[0].GetTxHash()
		shieldingAmount := uint64(0)
		for _, utxo := range actionData.ShieldingUTXO {
			shieldingAmount += utxo.GetOutputAmount()
		}
		UpdatePortalStateShieldingExternalTx(currentPortalState, actionData.TokenID, actionData.ProofHash, shieldingExternalTxHash, actionData.IncogAddressStr, shieldingAmount)

		// track shieldingReq status by txID into DB
		shieldingReqTrackData := metadata.PortalShieldingRequestStatus{
			Status:          portalcommonv4.PortalV4RequestAcceptedStatus,
			TokenID:         actionData.TokenID,
			IncogAddressStr: actionData.IncogAddressStr,
			ProofHash:       actionData.ProofHash,
			ShieldingUTXO:   actionData.ShieldingUTXO,
			MintingAmount:   actionData.MintingAmount,
			TxReqID:         actionData.TxReqID,
		}
		shieldingReqTrackDataBytes, _ := json.Marshal(shieldingReqTrackData)
		err = statedb.StoreShieldingRequestStatus(
			stateDB,
			actionData.TxReqID.String(),
			shieldingReqTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking shielding request by TxReqID: %+v", err)
			return nil
		}

		// update bridge/portal token info
		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: Can not new hash from shielding incTokenID: %+v", err)
			return nil
		}
		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
		if found {
			updatingInfo.CountUpAmt += actionData.MintingAmount
		} else {
			updatingInfo = metadata.UpdatingInfo{
				CountUpAmt:      actionData.MintingAmount,
				DeductAmt:       0,
				TokenID:         *incTokenID,
				ExternalTokenID: nil,
				IsCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo

	} else if reqStatus == portalcommonv4.PortalV4RequestRejectedChainStatus {
		shieldingReqTrackData := metadata.PortalShieldingRequestStatus{
			Status:          portalcommonv4.PortalV4RequestRejectedStatus,
			TokenID:         actionData.TokenID,
			IncogAddressStr: actionData.IncogAddressStr,
			ProofHash:       actionData.ProofHash,
			ShieldingUTXO:   actionData.ShieldingUTXO,
			MintingAmount:   actionData.MintingAmount,
			TxReqID:         actionData.TxReqID,
		}
		shieldingReqTrackDataBytes, _ := json.Marshal(shieldingReqTrackData)
		err = statedb.StoreShieldingRequestStatus(
			stateDB,
			actionData.TxReqID.String(),
			shieldingReqTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking shielding request tx: %+v", err)
			return nil
		}
	}

	return nil
}
