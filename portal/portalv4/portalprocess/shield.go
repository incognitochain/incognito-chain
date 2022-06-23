package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/portal/portalv4"
	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
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

func (p *PortalShieldingRequestProcessor) PrepareDataForBlockProducer(
	stateDB *statedb.StateDB, contentStr string,
	portalParams portalv4.PortalParams,
) (map[string]interface{}, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Shielding request: an error occurred while decoding content string of pToken request action - Error: %v", err)
		return nil, fmt.Errorf("Shielding request: an error occurred while decoding content string of pToken request action - Error: %v", err)
	}

	var actionData metadata.PortalShieldingRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Shielding request: an error occurred while unmarshal shielding request action - Error: %v", err)
		return nil, fmt.Errorf("Shielding request: an error occurred while unmarshal shielding request action - Error: %v", err)
	}

	portalTokenProcessor := portalParams.PortalTokens[actionData.Meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("Shielding Request: TokenID is not supported currently on Portal")
		return nil, fmt.Errorf("Shielding Request: TokenID is not supported currently on Portal")
	}
	shieldTxHash, err := portalTokenProcessor.GetTxHashFromProof(actionData.Meta.ShieldingProof)
	if err != nil {
		Logger.log.Errorf("Shielding Request: Can not get tx hash from shielding proof")
		return nil, fmt.Errorf("Shielding Request: Can not get tx hash from shielding proof")
	}
	proofHash := hashProof(shieldTxHash, actionData.Meta.IncogAddressStr)
	isExistProofTxHash, err := statedb.IsExistsShieldingRequest(stateDB, actionData.Meta.TokenID, proofHash)
	if err != nil {
		Logger.log.Errorf("Shielding request: an error occurred while get pToken request proof from DB - Error: %v", err)
		return nil, fmt.Errorf("Shielding request: an error occurred while get pToken request proof from DB - Error: %v", err)
	}

	optionalData := make(map[string]interface{})
	optionalData["isExistProofTxHash"] = isExistProofTxHash
	optionalData["proofHash"] = proofHash
	return optionalData, nil
}

// hashProof returns the hash of shielding proof (include tx proof and user inc address)
func hashProof(proof string, chainCode string) string {
	type shieldingProof struct {
		Proof      string
		IncAddress string
	}

	shieldProof := shieldingProof{
		Proof:      proof,
		IncAddress: chainCode,
	}
	shieldProofBytes, _ := json.Marshal(shieldProof)
	hash := common.HashB(shieldProofBytes)
	return fmt.Sprintf("%x", hash[:])
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildPortalShieldingRequestInstV4(
	tokenID string,
	incogAddressStr string,
	proofHash string,
	shieldingUTXO []*statedb.UTXO,
	mintingAmt uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	externalTxID string,
	status string,
	errorStr string,
) []string {
	shieldingReqContent := metadata.PortalShieldingRequestContent{
		TokenID:         tokenID,
		IncogAddressStr: incogAddressStr,
		ProofHash:       proofHash,
		ShieldingUTXO:   shieldingUTXO,
		MintingAmount:   mintingAmt,
		TxReqID:         txReqID,
		ExternalTxID:    externalTxID,
		ShardID:         shardID,
	}
	shieldingReqContentBytes, _ := json.Marshal(shieldingReqContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(shieldingReqContentBytes),
		errorStr,
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
		Logger.log.Errorf("Shielding request: an error occurred while decoding content string of portal shielding request action - Error: %v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalShieldingRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Shielding request: an error occurred while unmarshal portal shielding request action - Error: %v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	rejectInst := buildPortalShieldingRequestInstV4(
		meta.TokenID,
		meta.IncogAddressStr,
		"",
		[]*statedb.UTXO{},
		0,
		meta.Type,
		shardID,
		actionData.TxReqID,
		"",
		portalcommonv4.PortalV4RequestRejectedChainStatus,
		"isInvalidProof",
	)

	if currentPortalState == nil {
		Logger.log.Warn("Shielding Request: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	portalTokenProcessor := portalParams.PortalTokens[meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("Shielding Request: TokenID is not supported currently on Portal")
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

	proofHash, ok := optionalData["proofHash"].(string)
	if !ok {
		Logger.log.Errorf("Shielding Request: optionalData proofHash is invalid")
		return [][]string{rejectInst}, nil
	}

	// check unique external proof from portal state
	if currentPortalState.IsExistedShieldingExternalTx(meta.TokenID, proofHash) || isExistInStateDB {
		rejectInst[4] = "IsExistedProof"
		Logger.log.Errorf("Shielding Request: Shielding request proof exist in db %v", meta.ShieldingProof)
		return [][]string{rejectInst}, nil
	}

	// generate expected multisig address from master pubkeys and user payment address
	_, expectedReceivedMultisigAddress, err := portalTokenProcessor.GenerateOTMultisigAddress(
		portalParams.MasterPubKeys[meta.TokenID], int(portalParams.NumRequiredSigs), meta.IncogAddressStr)
	if err != nil {
		Logger.log.Error("Shielding Request: Could not generate multisig address - Error: %v", err)
		return [][]string{rejectInst}, nil
	}

	// verify shielding proof
	isValid, listUTXO, err := portalTokenProcessor.ParseAndVerifyShieldProof(
		meta.ShieldingProof, bc, expectedReceivedMultisigAddress, meta.IncogAddressStr, portalParams.MinShieldAmts[meta.TokenID])
	if !isValid || err != nil {
		Logger.log.Error("Shielding Request: Parse proof and verify shielding proof failed - Error: %v", err)
		return [][]string{rejectInst}, nil
	}

	// calculate shielding amount and minting amount
	shieldingAmount := uint64(0)
	for _, utxo := range listUTXO {
		shieldingAmount += utxo.GetOutputAmount()
	}
	mintingAmount := portalTokenProcessor.ConvertExternalToIncAmount(shieldingAmount)

	// update portal state
	currentPortalState.AddUTXOs(listUTXO, meta.TokenID)
	currentPortalState.AddShieldingExternalTx(
		meta.TokenID, proofHash, listUTXO[0].GetTxHash(),
		meta.IncogAddressStr, shieldingAmount)

	// calculate receiving shardID
	key, _ := wallet.Base58CheckDeserialize(actionData.Meta.IncogAddressStr)
	receivingShardID, _ := metadataBridge.GetShardIDFromPaymentAddress(key.KeySet.PaymentAddress)

	inst := buildPortalShieldingRequestInstV4(
		actionData.Meta.TokenID,
		actionData.Meta.IncogAddressStr,
		proofHash,
		listUTXO,
		mintingAmount,
		actionData.Meta.Type,
		receivingShardID,
		actionData.TxReqID,
		listUTXO[0].GetTxHash(),
		portalcommonv4.PortalV4RequestAcceptedChainStatus,
		"None",
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
		Logger.log.Errorf("Shielding Request: Current Portal state is nil")
		return nil
	}

	if len(instructions) != 5 {
		Logger.log.Errorf("Shielding Request: Instructions are in wrong format")
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalShieldingRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Shielding Request: Could not unmarshal instruction content %v - Error: %v\n", instructions[3], err)
		return nil
	}

	var shieldStatus byte
	reqStatus := instructions[2]
	if reqStatus == portalcommonv4.PortalV4RequestAcceptedChainStatus {
		shieldStatus = portalcommonv4.PortalV4RequestAcceptedStatus

		shieldingExternalTxHash := actionData.ShieldingUTXO[0].GetTxHash()
		shieldingAmount := uint64(0)
		for _, utxo := range actionData.ShieldingUTXO {
			shieldingAmount += utxo.GetOutputAmount()
		}

		currentPortalState.AddUTXOs(actionData.ShieldingUTXO, actionData.TokenID)
		currentPortalState.AddShieldingExternalTx(actionData.TokenID, actionData.ProofHash,
			shieldingExternalTxHash, actionData.IncogAddressStr, shieldingAmount)

		// update bridge token info
		err := metadata.UpdatePortalBridgeTokenInfo(updatingInfoByTokenID, actionData.TokenID, actionData.MintingAmount, false)
		if err != nil {
			Logger.log.Errorf("Shielding Request: Update Portal token info for UnshieldID - Error %v\n", actionData.TxReqID.String(), err)
			return nil
		}
	} else if reqStatus == portalcommonv4.PortalV4RequestRejectedChainStatus {
		shieldStatus = portalcommonv4.PortalV4RequestRejectedStatus
	}

	// track shieldingReq status by txID into DB
	shieldingReqTrackData := metadata.PortalShieldingRequestStatus{
		Status:          shieldStatus,
		Error:           instructions[4],
		TokenID:         actionData.TokenID,
		IncogAddressStr: actionData.IncogAddressStr,
		ProofHash:       actionData.ProofHash,
		ShieldingUTXO:   actionData.ShieldingUTXO,
		MintingAmount:   actionData.MintingAmount,
		TxReqID:         actionData.TxReqID,
		ExternalTxID:    actionData.ExternalTxID,
	}
	shieldingReqTrackDataBytes, _ := json.Marshal(shieldingReqTrackData)
	err = statedb.StoreShieldingRequestStatus(
		stateDB,
		actionData.TxReqID.String(),
		shieldingReqTrackDataBytes,
	)
	if err != nil {
		Logger.log.Errorf("Shielding Request: An error occurred while tracking shielding request tx - Error: %v", err)
	}

	return nil
}
