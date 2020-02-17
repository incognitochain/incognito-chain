package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildCustodianDepositInst(
	custodianAddressStr string,
	depositedAmount uint64,
	remoteAddresses map[string]string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	custodianDepositContent := metadata.PortalCustodianDepositContent{
		IncogAddressStr: custodianAddressStr,
		RemoteAddresses: remoteAddresses,
		DepositedAmount: depositedAmount,
		TxReqID:         txReqID,
		ShardID: shardID,
	}
	custodianDepositContentBytes, _ := json.Marshal(custodianDepositContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(custodianDepositContentBytes),
	}
}

func buildRequestPortingInst(
	uniqueRegisterId string,
	incogAddressStr string,
	pTokenId string,
	pTokenAddress string,
	registerAmount uint64,
	portingFee uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	portingRequestContent := metadata.PortalPortingRequestContent{
		UniqueRegisterId: 	uniqueRegisterId,
		IncogAddressStr: 	incogAddressStr,
		PTokenId: 			pTokenId,
		PTokenAddress: 		pTokenAddress,
		RegisterAmount: 	registerAmount,
		PortingFee: 		portingFee,
		TxReqID:         	txReqID,
	}

	portingRequestContentBytes, _ := json.Marshal(portingRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PortalPortingRequestSuccessStatus,
		string(portingRequestContentBytes),
	}
}


// buildInstructionsForCustodianDeposit builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForCustodianDeposit(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
) ([][]string, error) {

	// todo: need to validate instruction ? (should update currentPortalState ?)

	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalCustodianDepositAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForCustodianDeposit]: Current Portal state is null.")
		// need to refund collateral to custodian
		inst := buildCustodianDepositInst(
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.RemoteAddresses,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalCustodianDepositRefundChainStatus,
		)
		return [][]string{inst}, nil
	}

	inst := buildCustodianDepositInst(
		actionData.Meta.IncogAddressStr,
		actionData.Meta.DepositedAmount,
		actionData.Meta.RemoteAddresses,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		common.PortalCustodianDepositAcceptedChainStatus,
		)
	return [][]string{inst}, nil
}

//todo:
func (blockchain *BlockChain) buildInstructionsForPortingRequest(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
) ([][]string, error) {
	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForPortingRequest]: Current Portal state is null.")
		inst := []string{
			strconv.Itoa(metaType),
			strconv.Itoa(int(shardID)),
			common.PortalPortingRequestCanceledStatus,
			contentStr,
		}
		return [][]string{inst}, nil
	}

	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal porting request action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalUserRegisterAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal porting request action: %+v", err)
		return [][]string{}, nil
	}

	inst := buildRequestPortingInst(
		actionData.Meta.UniqueRegisterId,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.PTokenId,
		actionData.Meta.PTokenAddress,
		actionData.Meta.RegisterAmount,
		actionData.Meta.PortingFee,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
	)
	return [][]string{inst}, nil
}
