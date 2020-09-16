package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

/* =======
Portal Custodian Deposit Collateral (PRV) Processor
======= */

type portalCustodianDepositProcessor struct {
	*portalInstProcessor
}

func (p *portalCustodianDepositProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalCustodianDepositProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCustodianDepositProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

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
		ShardID:         shardID,
	}
	custodianDepositContentBytes, _ := json.Marshal(custodianDepositContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(custodianDepositContentBytes),
	}
}

func (p *portalCustodianDepositProcessor) buildNewInsts(
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
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
		Logger.log.Errorf("WARN - [buildInstructionsForCustodianDeposit]: Current Portal state is null.")
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
	meta := actionData.Meta

	keyCustodianState := statedb.GenerateCustodianStateObjectKey(meta.IncogAddressStr)
	keyCustodianStateStr := keyCustodianState.String()

	newCustodian := new(statedb.CustodianState)
	if currentPortalState.CustodianPoolState[keyCustodianStateStr] == nil {
		// new custodian
		newCustodian = statedb.NewCustodianStateWithValue(
			meta.IncogAddressStr, meta.DepositedAmount, meta.DepositedAmount,
			nil, nil,
			meta.RemoteAddresses, nil)
	} else {
		// custodian deposited before
		custodian := currentPortalState.CustodianPoolState[keyCustodianStateStr]
		totalCollateral := custodian.GetTotalCollateral() + meta.DepositedAmount
		freeCollateral := custodian.GetFreeCollateral() + meta.DepositedAmount
		holdingPubTokens := custodian.GetHoldingPublicTokens()
		lockedAmountCollateral := custodian.GetLockedAmountCollateral()
		rewardAmount := custodian.GetRewardAmount()
		remoteAddresses := custodian.GetRemoteAddresses()
		// if total collateral is zero, custodians are able to update remote addresses
		if custodian.GetTotalCollateral() == 0 {
			if len(meta.RemoteAddresses) > 0 {
				remoteAddresses = meta.RemoteAddresses
			}
		} else {
			sortedTokenIDs := make([]string, 0)
			for tokenID := range meta.RemoteAddresses {
				sortedTokenIDs = append(sortedTokenIDs, tokenID)
			}

			for _, tokenID := range sortedTokenIDs {
				if remoteAddresses[tokenID] == "" {
					remoteAddresses[tokenID] = meta.RemoteAddresses[tokenID]
				}
			}
		}
		newCustodian = statedb.NewCustodianStateWithValue(meta.IncogAddressStr, totalCollateral, freeCollateral,
			holdingPubTokens, lockedAmountCollateral, remoteAddresses, rewardAmount)
	}
	// update state of the custodian
	currentPortalState.CustodianPoolState[keyCustodianStateStr] = newCustodian

	inst := buildCustodianDepositInst(
		actionData.Meta.IncogAddressStr,
		actionData.Meta.DepositedAmount,
		newCustodian.GetRemoteAddresses(),
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
		common.PortalCustodianDepositAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

/* =======
Portal Custodian Request Withdraw Free Collaterals Processor
======= */

type portalRequestWithdrawCollateralProcessor struct {
	*portalInstProcessor
}

func (p *portalRequestWithdrawCollateralProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestWithdrawCollateralProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestWithdrawCollateralProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildCustodianWithdrawInst(
	metaType int,
	shardID byte,
	reqStatus string,
	paymentAddress string,
	amount uint64,
	remainFreeCollateral uint64,
	txReqID common.Hash,
) []string {
	content := metadata.PortalCustodianWithdrawRequestContent{
		PaymentAddress:       paymentAddress,
		Amount:               amount,
		RemainFreeCollateral: remainFreeCollateral,
		TxReqID:              txReqID,
		ShardID:              shardID,
	}

	contentBytes, _ := json.Marshal(content)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		reqStatus,
		string(contentBytes),
	}
}

func (p *portalRequestWithdrawCollateralProcessor) buildNewInsts(
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while decoding content string of custodian withdraw request action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalCustodianWithdrawRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Have an error occurred while unmarshal  custodian withdraw request action: %+v", err)
		return [][]string{}, nil
	}

	rejectInst := buildCustodianWithdrawInst(
		actionData.Meta.Type,
		shardID,
		common.PortalCustodianWithdrawRequestRejectedStatus,
		actionData.Meta.PaymentAddress,
		actionData.Meta.Amount,
		0,
		actionData.TxReqID,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null")
		return [][]string{rejectInst}, nil
	}

	if len(currentPortalState.CustodianPoolState) <= 0 {
		Logger.log.Errorf("Custodian state is empty")
		return [][]string{rejectInst}, nil
	}

	custodianKey := statedb.GenerateCustodianStateObjectKey(actionData.Meta.PaymentAddress)
	custodianKeyStr := custodianKey.String()
	custodian, ok := currentPortalState.CustodianPoolState[custodianKeyStr]
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return [][]string{rejectInst}, nil
	}

	if actionData.Meta.Amount > custodian.GetFreeCollateral() {
		Logger.log.Errorf("Free Collateral is not enough PRV")
		return [][]string{rejectInst}, nil
	}
	//withdraw
	remainFreeCollateral := custodian.GetFreeCollateral() - actionData.Meta.Amount
	totalFreeCollateral := custodian.GetTotalCollateral() - actionData.Meta.Amount

	inst := buildCustodianWithdrawInst(
		actionData.Meta.Type,
		shardID,
		common.PortalCustodianWithdrawRequestAcceptedStatus,
		actionData.Meta.PaymentAddress,
		actionData.Meta.Amount,
		remainFreeCollateral,
		actionData.TxReqID,
	)

	//update free collateral custodian
	custodian.SetFreeCollateral(remainFreeCollateral)
	custodian.SetTotalCollateral(totalFreeCollateral)
	currentPortalState.CustodianPoolState[custodianKeyStr] = custodian
	return [][]string{inst}, nil
}

/* =======
Portal Custodian Deposit Collaterals V3 (ETH and ERC20) Processor
======= */
type portalCustodianDepositProcessorV3 struct {
	*portalInstProcessor
}

func (p *portalCustodianDepositProcessorV3) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalCustodianDepositProcessorV3) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCustodianDepositProcessorV3) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// TODO
func (p *portalCustodianDepositProcessorV3) buildNewInsts(
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {

	return [][]string{}, nil
}