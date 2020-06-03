package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/relaying/bnb"
	btcrelaying "github.com/incognitochain/incognito-chain/relaying/btc"
	"strconv"
)

// beacon build new instruction from instruction received from Shard Block
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

func buildRequestPortingInst(
	metaType int,
	shardID byte,
	reqStatus string,
	uniqueRegisterId string,
	incogAddressStr string,
	pTokenId string,
	registerAmount uint64,
	portingFee uint64,
	custodian []*statedb.MatchingPortingCustodianDetail,
	txReqID common.Hash,
) []string {
	portingRequestContent := metadata.PortalPortingRequestContent{
		UniqueRegisterId: uniqueRegisterId,
		IncogAddressStr:  incogAddressStr,
		PTokenId:         pTokenId,
		RegisterAmount:   registerAmount,
		PortingFee:       portingFee,
		Custodian:        custodian,
		TxReqID:          txReqID,
		ShardID:          shardID,
	}

	portingRequestContentBytes, _ := json.Marshal(portingRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		reqStatus,
		string(portingRequestContentBytes),
	}
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildReqPTokensInst(
	uniquePortingID string,
	tokenID string,
	incogAddressStr string,
	portingAmount uint64,
	portingProof string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	reqPTokenContent := metadata.PortalRequestPTokensContent{
		UniquePortingID: uniquePortingID,
		TokenID:         tokenID,
		IncogAddressStr: incogAddressStr,
		PortingAmount:   portingAmount,
		PortingProof:    portingProof,
		TxReqID:         txReqID,
		ShardID:         shardID,
	}
	reqPTokenContentBytes, _ := json.Marshal(reqPTokenContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(reqPTokenContentBytes),
	}
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

// buildInstructionsForCustodianDeposit builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForCustodianDeposit(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
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

func (blockchain *BlockChain) buildInstructionsForPortingRequest(
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while decoding content string of portal porting request action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalUserRegisterAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while unmarshal portal porting request action: %+v", err)
		return [][]string{}, nil
	}

	if currentPortalState == nil {
		Logger.log.Warn("Porting request: Current Portal state is null")
		return [][]string{}, nil
	}

	//check unique id from record from db
	portingRequestKeyExist, err := statedb.IsPortingRequestIdExist(stateDB, []byte(actionData.Meta.UniqueRegisterId))

	if err != nil {
		Logger.log.Errorf("Porting request: Get item portal by prefix error: %+v", err)

		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	if portingRequestKeyExist {
		Logger.log.Errorf("Porting request: Porting request id exist, key %v", actionData.Meta.UniqueRegisterId)
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.Meta.UniqueRegisterId)
	if _, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]; ok {
		Logger.log.Errorf("Porting request: Waiting porting request exist, key %v", waitingPortingRequestKey)
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	//get exchange rates
	exchangeRatesState := currentPortalState.FinalExchangeRatesState
	if exchangeRatesState == nil {
		Logger.log.Errorf("Porting request, exchange rates not found")
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	if len(currentPortalState.CustodianPoolState) <= 0 {
		Logger.log.Errorf("Porting request: Custodian not found")
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	var sortCustodianStateByFreeCollateral []CustodianStateSlice
	sortCustodianByAmountAscent(actionData.Meta, currentPortalState.CustodianPoolState, &sortCustodianStateByFreeCollateral)

	if len(sortCustodianStateByFreeCollateral) <= 0 {
		Logger.log.Errorf("Porting request, custodian not found")

		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	//validation porting fees
	exchangePortingFees, err := CalMinPortingFee(actionData.Meta.RegisterAmount, actionData.Meta.PTokenId, exchangeRatesState, portalParams.MinPercentPortingFee)
	if err != nil {
		Logger.log.Errorf("Calculate Porting fee is error %v", err)
		return [][]string{}, nil
	}

	Logger.log.Infof("Porting request, porting fees need %v", exchangePortingFees)

	if actionData.Meta.PortingFee < exchangePortingFees {
		Logger.log.Errorf("Porting request, Porting fees is wrong")

		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)
		return [][]string{inst}, nil
	}

	pickedCustodians, err := pickUpCustodians(actionData.Meta, exchangeRatesState, sortCustodianStateByFreeCollateral, currentPortalState, portalParams)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while picking up custodians for the porting request: %+v", err)
	}
	if len(pickedCustodians) == 0 || err != nil {
		Logger.log.Errorf("Porting request, custodian not found")
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			pickedCustodians,
			actionData.TxReqID,
		)
		return [][]string{inst}, nil
	}

	//verify total amount
	var totalPToken uint64 = 0
	for _, eachCustodian := range pickedCustodians {
		totalPToken = totalPToken + eachCustodian.Amount
	}

	if totalPToken < actionData.Meta.RegisterAmount {
		Logger.log.Errorf("Porting request, total matching amount of picked custodians is less than porting amount %v != %v", totalPToken, actionData.Meta.RegisterAmount)

		Logger.log.Errorf("Porting request, custodian not found")
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedChainStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	inst := buildRequestPortingInst(
		actionData.Meta.Type,
		shardID,
		common.PortalPortingRequestAcceptedChainStatus,
		actionData.Meta.UniqueRegisterId,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.PTokenId,
		actionData.Meta.RegisterAmount,
		actionData.Meta.PortingFee,
		pickedCustodians,
		actionData.TxReqID,
	)

	newPortingRequestStateWaiting := statedb.NewWaitingPortingRequestWithValue(
		actionData.Meta.UniqueRegisterId,
		actionData.TxReqID,
		actionData.Meta.PTokenId,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.RegisterAmount,
		pickedCustodians,
		actionData.Meta.PortingFee,
		beaconHeight+1,
	)

	keyWaitingPortingRequest := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.Meta.UniqueRegisterId)
	currentPortalState.WaitingPortingRequests[keyWaitingPortingRequest.String()] = newPortingRequestStateWaiting

	return [][]string{inst}, nil
}

// buildInstructionsForCustodianDeposit builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForReqPTokens(
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
) ([][]string, error) {

	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRequestPTokensAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForCustodianDeposit]: Current Portal state is null.")
		inst := buildReqPTokensInst(
			meta.UniquePortingID,
			meta.TokenID,
			meta.IncogAddressStr,
			meta.PortingAmount,
			meta.PortingProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check meta.UniquePortingID is in waiting PortingRequests list in portal state or not
	portingID := meta.UniquePortingID
	keyWaitingPortingRequest := statedb.GeneratePortalWaitingPortingRequestObjectKey(portingID)
	keyWaitingPortingRequestStr := keyWaitingPortingRequest.String()
	waitingPortingRequest := currentPortalState.WaitingPortingRequests[keyWaitingPortingRequestStr]
	if waitingPortingRequest == nil {
		Logger.log.Errorf("PortingID is not existed in waiting porting requests list")
		inst := buildReqPTokensInst(
			meta.UniquePortingID,
			meta.TokenID,
			meta.IncogAddressStr,
			meta.PortingAmount,
			meta.PortingProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	//check unique id from record from db
	portingRequest, err := statedb.GetPortalStateStatusMultiple(stateDB, statedb.PortalPortingRequestStatusPrefix(), []byte(meta.UniquePortingID))

	if err != nil {
		Logger.log.Errorf("Can not get porting req status for portingID %v, %v\n", meta.UniquePortingID, err)
		inst := buildReqPTokensInst(
			meta.UniquePortingID,
			meta.TokenID,
			meta.IncogAddressStr,
			meta.PortingAmount,
			meta.PortingProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	var portingRequestStatus metadata.PortingRequestStatus
	err = json.Unmarshal(portingRequest, &portingRequestStatus)
	if err != nil {
		Logger.log.Errorf("Has an error occurred while unmarshal PortingRequestStatus: %+v", err)
		inst := buildReqPTokensInst(
			meta.UniquePortingID,
			meta.TokenID,
			meta.IncogAddressStr,
			meta.PortingAmount,
			meta.PortingProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	if portingRequestStatus.Status != common.PortalPortingReqWaitingStatus {
		Logger.log.Errorf("PortingID status invalid, expected %v , but got %v\n", common.PortalPortingReqWaitingStatus, portingRequestStatus.Status)
		inst := buildReqPTokensInst(
			meta.UniquePortingID,
			meta.TokenID,
			meta.IncogAddressStr,
			meta.PortingAmount,
			meta.PortingProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check tokenID
	if meta.TokenID != waitingPortingRequest.TokenID() {
		Logger.log.Errorf("TokenID is not correct in portingID req")
		inst := buildReqPTokensInst(
			meta.UniquePortingID,
			meta.TokenID,
			meta.IncogAddressStr,
			meta.PortingAmount,
			meta.PortingProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check porting amount
	if meta.PortingAmount != waitingPortingRequest.Amount() {
		Logger.log.Errorf("PortingAmount is not correct in portingID req")
		inst := buildReqPTokensInst(
			meta.UniquePortingID,
			meta.TokenID,
			meta.IncogAddressStr,
			meta.PortingAmount,
			meta.PortingProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	if meta.TokenID == common.PortalBTCIDStr {
		btcChain := blockchain.config.BTCChain
		if btcChain == nil {
			Logger.log.Error("BTC relaying chain should not be null")
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}
		// parse PortingProof in meta
		btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(meta.PortingProof)
		if err != nil {
			Logger.log.Errorf("PortingProof is invalid %v\n", err)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		isValid, err := btcChain.VerifyTxWithMerkleProofs(btcTxProof)
		if !isValid || err != nil {
			Logger.log.Errorf("Verify btcTxProof failed %v", err)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// extract attached message from txOut's OP_RETURN
		btcAttachedMsg, err := btcrelaying.ExtractAttachedMsgFromTx(btcTxProof.BTCTx)
		if err != nil {
			Logger.log.Errorf("Could not extract attached message from BTC tx proof with err: %v", err)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		encodedMsg := btcrelaying.HashAndEncodeBase58(meta.UniquePortingID)
		if btcAttachedMsg != encodedMsg {
			Logger.log.Errorf("PortingId in the btc attached message is not matched with portingID in metadata")
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// check whether amount transfer in txBNB is equal porting amount or not
		// check receiver and amount in tx
		// get list matching custodians in waitingPortingRequest
		custodians := waitingPortingRequest.Custodians()
		outputs := btcTxProof.BTCTx.TxOut
		for _, cusDetail := range custodians {
			remoteAddressNeedToBeTransfer := cusDetail.RemoteAddress
			amountNeedToBeTransfer := cusDetail.Amount
			amountNeedToBeTransferInBTC := btcrelaying.ConvertIncPBTCAmountToExternalBTCAmount(int64(amountNeedToBeTransfer))

			isChecked := false
			for _, out := range outputs {
				addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
				if err != nil {
					Logger.log.Errorf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
					continue
				}
				if addrStr != remoteAddressNeedToBeTransfer {
					continue
				}
				if out.Value < amountNeedToBeTransferInBTC {
					Logger.log.Errorf("BTC-TxProof is invalid - the transferred amount to %s must be equal to or greater than %d, but got %d", addrStr, amountNeedToBeTransferInBTC, out.Value)
					inst := buildReqPTokensInst(
						meta.UniquePortingID,
						meta.TokenID,
						meta.IncogAddressStr,
						meta.PortingAmount,
						meta.PortingProof,
						meta.Type,
						shardID,
						actionData.TxReqID,
						common.PortalReqPTokensRejectedChainStatus,
					)
					return [][]string{inst}, nil
				} else {
					isChecked = true
					break
				}
			}
			if !isChecked {
				Logger.log.Error("BTC-TxProof is invalid")
				inst := buildReqPTokensInst(
					meta.UniquePortingID,
					meta.TokenID,
					meta.IncogAddressStr,
					meta.PortingAmount,
					meta.PortingProof,
					meta.Type,
					shardID,
					actionData.TxReqID,
					common.PortalReqPTokensRejectedChainStatus,
				)
				return [][]string{inst}, nil
			}
		}
		// update holding public token for custodians
		for _, cusDetail := range custodians {
			custodianKey := statedb.GenerateCustodianStateObjectKey(cusDetail.IncAddress)
			UpdateCustodianStateAfterUserRequestPToken(currentPortalState, custodianKey.String(), waitingPortingRequest.TokenID(), cusDetail.Amount)
		}

		inst := buildReqPTokensInst(
			actionData.Meta.UniquePortingID,
			actionData.Meta.TokenID,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PortingAmount,
			actionData.Meta.PortingProof,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensAcceptedChainStatus,
		)

		// remove waiting porting request from currentPortalState
		deleteWaitingPortingRequest(currentPortalState, keyWaitingPortingRequestStr)
		return [][]string{inst}, nil

	} else if meta.TokenID == common.PortalBNBIDStr {
		// parse PortingProof in meta
		txProofBNB, err := bnb.ParseBNBProofFromB64EncodeStr(meta.PortingProof)
		if err != nil {
			Logger.log.Errorf("PortingProof is invalid %v\n", err)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// check minimum confirmations block of bnb proof
		latestBNBBlockHeight, err2 := blockchain.GetLatestBNBBlkHeight()
		if err2 != nil {
			Logger.log.Errorf("Can not get latest relaying bnb block height %v\n", err)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		if latestBNBBlockHeight < txProofBNB.BlockHeight+bnb.MinConfirmationsBlock {
			Logger.log.Errorf("Not enough min bnb confirmations block %v, latestBNBBlockHeight %v - txProofBNB.BlockHeight %v\n",
				bnb.MinConfirmationsBlock, latestBNBBlockHeight, txProofBNB.BlockHeight)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}
		dataHash, err2 := blockchain.GetBNBDataHash(txProofBNB.BlockHeight)
		if err2 != nil {
			Logger.log.Errorf("Error when get data hash in blockHeight %v - %v\n",
				txProofBNB.BlockHeight, err2)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		isValid, err := txProofBNB.Verify(dataHash)
		if !isValid || err != nil {
			Logger.log.Errorf("Verify txProofBNB failed %v", err)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// parse Tx from Data in txProofBNB
		txBNB, err := bnb.ParseTxFromData(txProofBNB.Proof.Data)
		if err != nil {
			Logger.log.Errorf("Data in PortingProof is invalid %v", err)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// check memo attach portingID req:
		memo := txBNB.Memo
		memoBytes, err2 := base64.StdEncoding.DecodeString(memo)
		if err2 != nil {
			Logger.log.Errorf("Can not decode memo in tx bnb proof", err2)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		var portingMemo PortingMemoBNB
		err2 = json.Unmarshal(memoBytes, &portingMemo)
		if err2 != nil {
			Logger.log.Errorf("Can not unmarshal memo in tx bnb proof", err2)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		if portingMemo.PortingID != meta.UniquePortingID {
			Logger.log.Errorf("PortingId in memoTx is not matched with portingID in metadata", err2)
			inst := buildReqPTokensInst(
				meta.UniquePortingID,
				meta.TokenID,
				meta.IncogAddressStr,
				meta.PortingAmount,
				meta.PortingProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqPTokensRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// check whether amount transfer in txBNB is equal porting amount or not
		// check receiver and amount in tx
		// get list matching custodians in waitingPortingRequest
		custodians := waitingPortingRequest.Custodians()
		outputs := txBNB.Msgs[0].(msg.SendMsg).Outputs
		for _, cusDetail := range custodians {
			remoteAddressNeedToBeTransfer := cusDetail.RemoteAddress
			amountNeedToBeTransfer := cusDetail.Amount
			amountNeedToBeTransferInBNB := convertIncPBNBAmountToExternalBNBAmount(int64(amountNeedToBeTransfer))

			isChecked := false
			for _, out := range outputs {
				addr, _ := bnb.GetAccAddressString(&out.Address, blockchain.config.ChainParams.BNBRelayingHeaderChainID)
				if addr != remoteAddressNeedToBeTransfer {
					Logger.log.Warnf("[portal] remoteAddressNeedToBeTransfer: %v - addr: %v\n", remoteAddressNeedToBeTransfer, addr)
					continue
				}

				// calculate amount that was transferred to custodian's remote address
				amountTransfer := int64(0)
				for _, coin := range out.Coins {
					if coin.Denom == bnb.DenomBNB {
						amountTransfer += coin.Amount
					}
				}
				if amountTransfer < amountNeedToBeTransferInBNB {
					Logger.log.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal to or greater than %d, but got %d",
						addr, amountNeedToBeTransferInBNB, amountTransfer)
					inst := buildReqPTokensInst(
						meta.UniquePortingID,
						meta.TokenID,
						meta.IncogAddressStr,
						meta.PortingAmount,
						meta.PortingProof,
						meta.Type,
						shardID,
						actionData.TxReqID,
						common.PortalReqPTokensRejectedChainStatus,
					)
					return [][]string{inst}, nil
				} else {
					isChecked = true
					break
				}
			}
			if !isChecked {
				Logger.log.Errorf("TxProof-BNB is invalid - Receiver address is invalid, expected %v",
					remoteAddressNeedToBeTransfer)
				inst := buildReqPTokensInst(
					meta.UniquePortingID,
					meta.TokenID,
					meta.IncogAddressStr,
					meta.PortingAmount,
					meta.PortingProof,
					meta.Type,
					shardID,
					actionData.TxReqID,
					common.PortalReqPTokensRejectedChainStatus,
				)
				return [][]string{inst}, nil
			}
		}

		// update holding public token for custodians
		for _, cusDetail := range custodians {
			custodianKey := statedb.GenerateCustodianStateObjectKey(cusDetail.IncAddress)
			UpdateCustodianStateAfterUserRequestPToken(currentPortalState, custodianKey.String(), waitingPortingRequest.TokenID(), cusDetail.Amount)
		}

		inst := buildReqPTokensInst(
			actionData.Meta.UniquePortingID,
			actionData.Meta.TokenID,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PortingAmount,
			actionData.Meta.PortingProof,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensAcceptedChainStatus,
		)

		// remove waiting porting request from currentPortalState
		deleteWaitingPortingRequest(currentPortalState, keyWaitingPortingRequestStr)
		return [][]string{inst}, nil
	} else {
		Logger.log.Errorf("TokenID is not supported currently on Portal")
		inst := buildReqPTokensInst(
			meta.UniquePortingID,
			meta.TokenID,
			meta.IncogAddressStr,
			meta.PortingAmount,
			meta.PortingProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqPTokensRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}
}

func (blockchain *BlockChain) buildInstructionsForExchangeRates(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalExchangeRatesAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal exchange rates action: %+v", err)
		return [][]string{}, nil
	}

	//check key from db
	if currentPortalState.ExchangeRatesRequests != nil {
		_, ok := currentPortalState.ExchangeRatesRequests[actionData.TxReqID.String()]
		if ok {
			Logger.log.Errorf("ERROR: exchange rates key is duplicated")

			portalExchangeRatesContent := metadata.PortalExchangeRatesContent{
				SenderAddress: actionData.Meta.SenderAddress,
				Rates:         actionData.Meta.Rates,
				TxReqID:       actionData.TxReqID,
				LockTime:      actionData.LockTime,
			}

			portalExchangeRatesContentBytes, _ := json.Marshal(portalExchangeRatesContent)

			inst := []string{
				strconv.Itoa(metaType),
				strconv.Itoa(int(shardID)),
				common.PortalExchangeRatesRejectedChainStatus,
				string(portalExchangeRatesContentBytes),
			}

			return [][]string{inst}, nil
		}
	}

	//success
	portalExchangeRatesContent := metadata.PortalExchangeRatesContent{
		SenderAddress: actionData.Meta.SenderAddress,
		Rates:         actionData.Meta.Rates,
		TxReqID:       actionData.TxReqID,
		LockTime:      actionData.LockTime,
	}

	portalExchangeRatesContentBytes, _ := json.Marshal(portalExchangeRatesContent)

	inst := []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PortalExchangeRatesAcceptedChainStatus,
		string(portalExchangeRatesContentBytes),
	}

	//update E-R request
	if currentPortalState.ExchangeRatesRequests != nil {
		currentPortalState.ExchangeRatesRequests[actionData.TxReqID.String()] = metadata.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			actionData.Meta.SenderAddress,
			actionData.Meta.Rates,
		)
	} else {
		//new object
		newExchangeRatesRequest := make(map[string]*metadata.ExchangeRatesRequestStatus)
		newExchangeRatesRequest[actionData.TxReqID.String()] = metadata.NewExchangeRatesRequestStatus(
			common.PortalExchangeRatesAcceptedStatus,
			actionData.Meta.SenderAddress,
			actionData.Meta.Rates,
		)

		currentPortalState.ExchangeRatesRequests = newExchangeRatesRequest
	}

	return [][]string{inst}, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildRedeemRequestInst(
	uniqueRedeemID string,
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	remoteAddress string,
	redeemFee uint64,
	matchingCustodianDetail []*statedb.MatchingRedeemCustodianDetail,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	redeemRequestContent := metadata.PortalRedeemRequestContent{
		UniqueRedeemID:          uniqueRedeemID,
		TokenID:                 tokenID,
		RedeemAmount:            redeemAmount,
		RedeemerIncAddressStr:   incAddressStr,
		RemoteAddress:           remoteAddress,
		MatchingCustodianDetail: matchingCustodianDetail,
		RedeemFee:               redeemFee,
		TxReqID:                 txReqID,
		ShardID:                 shardID,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

// buildInstructionsForRedeemRequest builds instruction for redeem request action
func (blockchain *BlockChain) buildInstructionsForRedeemRequest(
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal redeem request action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRedeemRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal redeem request action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForRedeemRequest]: Current Portal state is null.")
		// need to mint ptoken to user
		inst := buildRedeemRequestInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			nil,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	redeemID := meta.UniqueRedeemID

	// check uniqueRedeemID is existed waitingRedeem list or not
	keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(redeemID)
	keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
	waitingRedeemRequest := currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr]
	if waitingRedeemRequest != nil {
		Logger.log.Errorf("RedeemID is existed in waiting redeem requests list %v\n", redeemID)
		inst := buildRedeemRequestInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			nil,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check uniqueRedeemID is existed in db or not
	redeemRequestBytes, err := statedb.GetPortalRedeemRequestStatus(stateDB, meta.UniqueRedeemID)
	if err != nil {
		Logger.log.Errorf("Can not get redeem req status for redeemID %v, %v\n", meta.UniqueRedeemID, err)
		inst := buildRedeemRequestInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			nil,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedChainStatus,
		)
		return [][]string{inst}, nil
	} else if len(redeemRequestBytes) > 0 {
		Logger.log.Errorf("RedeemID is existed in redeem requests list in db %v\n", redeemID)
		inst := buildRedeemRequestInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			nil,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// get tokenID from redeemTokenID
	tokenID := meta.TokenID

	// check redeem fee
	if currentPortalState.FinalExchangeRatesState == nil {
		Logger.log.Errorf("Can not get exchange rate at beaconHeight %v\n", beaconHeight)
		inst := buildRedeemRequestInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			nil,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}
	minRedeemFee, err := CalMinRedeemFee(meta.RedeemAmount, tokenID, currentPortalState.FinalExchangeRatesState, portalParams.MinPercentRedeemFee)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum redeem fee %v\n", err)
		inst := buildRedeemRequestInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			nil,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	if meta.RedeemFee < minRedeemFee {
		Logger.log.Errorf("Redeem fee is invalid, minRedeemFee %v, but get %v\n", minRedeemFee, meta.RedeemFee)
		inst := buildRedeemRequestInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			nil,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// pick custodian(s) who holding public token to return user
	matchingCustodiansDetail, err := pickupCustodianForRedeem(meta.RedeemAmount, tokenID, currentPortalState)
	if err != nil {
		Logger.log.Errorf("Error when pick up custodian for redeem %v\n", err)
		inst := buildRedeemRequestInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			nil,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// update custodian state (holding public tokens)
	for _, cus := range matchingCustodiansDetail {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(cus.GetIncognitoAddress())
		custodianStateKeyStr := custodianStateKey.String()
		if currentPortalState.CustodianPoolState[custodianStateKeyStr].GetHoldingPublicTokens()[tokenID] < cus.GetAmount() {
			Logger.log.Errorf("Amount holding public tokens is less than matching redeem amount")
			inst := buildRedeemRequestInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.RedeemAmount,
				meta.RedeemerIncAddressStr,
				meta.RemoteAddress,
				meta.RedeemFee,
				nil,
				meta.Type,
				actionData.ShardID,
				actionData.TxReqID,
				common.PortalRedeemRequestRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		holdingPubTokenTmp := currentPortalState.CustodianPoolState[custodianStateKeyStr].GetHoldingPublicTokens()
		holdingPubTokenTmp[tokenID] -= cus.GetAmount()
		currentPortalState.CustodianPoolState[custodianStateKeyStr].SetHoldingPublicTokens(holdingPubTokenTmp)
	}

	// add to waiting Redeem list
	redeemRequest := statedb.NewWaitingRedeemRequestWithValue(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemAmount,
		matchingCustodiansDetail,
		meta.RedeemFee,
		beaconHeight+1,
		actionData.TxReqID,
	)
	currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr] = redeemRequest

	Logger.log.Infof("[Portal] Build accepted instruction for redeem request")
	inst := buildRedeemRequestInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemFee,
		matchingCustodiansDetail,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalRedeemRequestAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

/**
Validation:
	- verify each instruct belong shard
	- check amount < fee collateral
	- build PortalCustodianWithdrawRequestContent to send beacon
*/
func (blockchain *BlockChain) buildInstructionsForCustodianWithdraw(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
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

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null")
		return [][]string{}, nil
	}

	if len(currentPortalState.CustodianPoolState) <= 0 {
		Logger.log.Errorf("Custodian state is empty")

		inst := buildCustodianWithdrawInst(
			actionData.Meta.Type,
			shardID,
			common.PortalCustodianWithdrawRequestRejectedStatus,
			actionData.Meta.PaymentAddress,
			actionData.Meta.Amount,
			0,
			actionData.TxReqID,
		)
		return [][]string{inst}, nil
	}

	custodianKey := statedb.GenerateCustodianStateObjectKey(actionData.Meta.PaymentAddress)
	custodianKeyStr := custodianKey.String()
	custodian, ok := currentPortalState.CustodianPoolState[custodianKeyStr]

	if !ok {
		Logger.log.Errorf("Custodian not found")

		inst := buildCustodianWithdrawInst(
			actionData.Meta.Type,
			shardID,
			common.PortalCustodianWithdrawRequestRejectedStatus,
			actionData.Meta.PaymentAddress,
			actionData.Meta.Amount,
			0,
			actionData.TxReqID,
		)
		return [][]string{inst}, nil
	}

	if actionData.Meta.Amount > custodian.GetFreeCollateral() {
		Logger.log.Errorf("Free Collateral is not enough PRV")

		inst := buildCustodianWithdrawInst(
			actionData.Meta.Type,
			shardID,
			common.PortalCustodianWithdrawRequestRejectedStatus,
			actionData.Meta.PaymentAddress,
			actionData.Meta.Amount,
			0,
			actionData.TxReqID,
		)
		return [][]string{inst}, nil
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

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildReqUnlockCollateralInst(
	uniqueRedeemID string,
	tokenID string,
	custodianAddressStr string,
	redeemAmount uint64,
	unlockAmount uint64,
	redeemProof string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	reqUnlockCollateralContent := metadata.PortalRequestUnlockCollateralContent{
		UniqueRedeemID:      uniqueRedeemID,
		TokenID:             tokenID,
		CustodianAddressStr: custodianAddressStr,
		RedeemAmount:        redeemAmount,
		UnlockAmount:        unlockAmount,
		RedeemProof:         redeemProof,
		TxReqID:             txReqID,
		ShardID:             shardID,
	}
	reqUnlockCollateralContentBytes, _ := json.Marshal(reqUnlockCollateralContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(reqUnlockCollateralContentBytes),
	}
}

// buildInstructionsForReqUnlockCollateral builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForReqUnlockCollateral(
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
) ([][]string, error) {

	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal request unlock collateral action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRequestUnlockCollateralAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal request unlock collateral action: %+v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqUnlockCollateral]: Current Portal state is null.")
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check meta.UniqueRedeemID is in waiting RedeemRequests list in portal state or not
	redeemID := meta.UniqueRedeemID
	keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(redeemID)
	keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
	waitingRedeemRequest := currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr]
	if waitingRedeemRequest == nil {
		Logger.log.Errorf("redeemID is not existed in waiting redeem requests list")
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check status of request unlock collateral by redeemID
	redeemReqStatusBytes, err := statedb.GetPortalRedeemRequestStatus(stateDB, redeemID)
	if err != nil {
		Logger.log.Errorf("Can not get redeem request by redeemID from db %v\n", err)
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}
	var redeemRequest metadata.PortalRedeemRequestStatus
	err = json.Unmarshal(redeemReqStatusBytes, &redeemRequest)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal redeem request %v\n", err)
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	if redeemRequest.Status != common.PortalRedeemReqWaitingStatus {
		Logger.log.Errorf("Redeem request %v has invalid status %v\n", redeemID, redeemRequest.Status)
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check tokenID
	if meta.TokenID != waitingRedeemRequest.GetTokenID() {
		Logger.log.Errorf("TokenID is not correct in redeemID req")
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// check redeem amount of matching custodian
	amountMatchingCustodian := uint64(0)
	isFoundCustodian := false
	for _, cus := range waitingRedeemRequest.GetCustodians() {
		if cus.GetIncognitoAddress() == meta.CustodianAddressStr {
			amountMatchingCustodian = cus.GetAmount()
			isFoundCustodian = true
			break
		}
	}

	if !isFoundCustodian {
		Logger.log.Errorf("Custodian address %v is not in redeemID req %v", meta.CustodianAddressStr, meta.UniqueRedeemID)
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	if meta.RedeemAmount != amountMatchingCustodian {
		Logger.log.Errorf("RedeemAmount is not correct in redeemID req")
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	// validate proof and memo in tx
	if meta.TokenID == common.PortalBTCIDStr {
		btcChain := blockchain.config.BTCChain
		if btcChain == nil {
			Logger.log.Error("BTC relaying chain should not be null")
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}
		// parse PortingProof in meta
		btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(meta.RedeemProof)
		if err != nil {
			Logger.log.Errorf("PortingProof is invalid %v\n", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		isValid, err := btcChain.VerifyTxWithMerkleProofs(btcTxProof)
		if !isValid || err != nil {
			Logger.log.Errorf("Verify btcTxProof failed %v", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// extract attached message from txOut's OP_RETURN
		btcAttachedMsg, err := btcrelaying.ExtractAttachedMsgFromTx(btcTxProof.BTCTx)
		if err != nil {
			Logger.log.Errorf("Could not extract message from btc proof with error: ", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		rawMsg := fmt.Sprintf("%s%s", meta.UniqueRedeemID, meta.CustodianAddressStr)
		encodedMsg := btcrelaying.HashAndEncodeBase58(rawMsg)
		if btcAttachedMsg != encodedMsg {
			Logger.log.Errorf("The hash of combination of UniqueRedeemID(%s) and CustodianAddressStr(%s) is not matched to tx's attached message", meta.UniqueRedeemID, meta.CustodianAddressStr)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// check whether amount transfer in txBNB is equal redeem amount or not
		// check receiver and amount in tx
		// get list matching custodians in waitingRedeemRequest

		outputs := btcTxProof.BTCTx.TxOut
		remoteAddressNeedToBeTransfer := waitingRedeemRequest.GetRedeemerRemoteAddress()
		amountNeedToBeTransfer := meta.RedeemAmount
		amountNeedToBeTransferInBTC := btcrelaying.ConvertIncPBTCAmountToExternalBTCAmount(int64(amountNeedToBeTransfer))

		isChecked := false
		for _, out := range outputs {
			addrStr, err := btcChain.ExtractPaymentAddrStrFromPkScript(out.PkScript)
			if err != nil {
				Logger.log.Warnf("[portal] ExtractPaymentAddrStrFromPkScript: could not extract payment address string from pkscript with err: %v\n", err)
				continue
			}
			if addrStr != remoteAddressNeedToBeTransfer {
				continue
			}
			if out.Value < amountNeedToBeTransferInBTC {
				Logger.log.Errorf("BTC-TxProof is invalid - the transferred amount to %s must be equal to or greater than %d, but got %d", addrStr, amountNeedToBeTransferInBTC, out.Value)
				inst := buildReqUnlockCollateralInst(
					meta.UniqueRedeemID,
					meta.TokenID,
					meta.CustodianAddressStr,
					meta.RedeemAmount,
					0,
					meta.RedeemProof,
					meta.Type,
					shardID,
					actionData.TxReqID,
					common.PortalReqUnlockCollateralRejectedChainStatus,
				)
				return [][]string{inst}, nil
			} else {
				isChecked = true
				break
			}
		}

		if !isChecked {
			Logger.log.Error("BTC-TxProof is invalid")
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// calculate unlock amount
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.CustodianAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		unlockAmount, err := CalUnlockCollateralAmount(currentPortalState, custodianStateKeyStr, meta.RedeemAmount, meta.TokenID)
		if err != nil {
			Logger.log.Errorf("Error calculating unlock amount for custodian %v", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// update custodian state (FreeCollateral, LockedAmountCollateral)
		err = updateCustodianStateAfterReqUnlockCollateral(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			unlockAmount, meta.TokenID)
		if err != nil {
			Logger.log.Errorf("Error when updating custodian state after unlocking collateral %v", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
		updatedCustodians, err := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].GetCustodians(), meta.CustodianAddressStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while removing custodian %v from matching custodians", meta.CustodianAddressStr)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}
		currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].SetCustodians(updatedCustodians)

		// remove redeem request from WaitingRedeemRequest list when all matching custodians return public token to user
		// when list matchingCustodianDetail is empty
		if len(currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].GetCustodians()) == 0 {
			deleteWaitingRedeemRequest(currentPortalState, keyWaitingRedeemRequestStr)
		}

		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			unlockAmount,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralAcceptedChainStatus,
		)

		return [][]string{inst}, nil

	} else if meta.TokenID == common.PortalBNBIDStr {
		// parse PortingProof in meta
		txProofBNB, err := bnb.ParseBNBProofFromB64EncodeStr(meta.RedeemProof)
		if err != nil {
			Logger.log.Errorf("RedeemProof is invalid %v\n", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// check minimum confirmations block of bnb proof
		latestBNBBlockHeight, err2 := blockchain.GetLatestBNBBlkHeight()
		if err2 != nil {
			Logger.log.Errorf("Can not get latest relaying bnb block height %v\n", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		if latestBNBBlockHeight < txProofBNB.BlockHeight+bnb.MinConfirmationsBlock {
			Logger.log.Errorf("Not enough min bnb confirmations block %v, latestBNBBlockHeight %v - txProofBNB.BlockHeight %v\n",
				bnb.MinConfirmationsBlock, latestBNBBlockHeight, txProofBNB.BlockHeight)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		dataHash, err2 := blockchain.GetBNBDataHash(txProofBNB.BlockHeight)
		if err2 != nil {
			Logger.log.Errorf("Error when get data hash in blockHeight %v - %v\n",
				txProofBNB.BlockHeight, err2)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		isValid, err := txProofBNB.Verify(dataHash)
		if !isValid || err != nil {
			Logger.log.Errorf("Verify txProofBNB failed %v", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// parse Tx from Data in txProofBNB
		txBNB, err := bnb.ParseTxFromData(txProofBNB.Proof.Data)
		if err != nil {
			Logger.log.Errorf("Data in RedeemProof is invalid %v", err)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// check memo attach redeemID req (compare hash memo)
		memo := txBNB.Memo
		memoHashBytes, err2 := base64.StdEncoding.DecodeString(memo)
		if err2 != nil {
			Logger.log.Errorf("Can not decode memo in tx bnb proof", err2)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		expectedRedeemMemo := RedeemMemoBNB{
			RedeemID:                  redeemID,
			CustodianIncognitoAddress: meta.CustodianAddressStr}
		expectedRedeemMemoBytes, _ := json.Marshal(expectedRedeemMemo)
		expectedRedeemMemoHashBytes := common.HashB(expectedRedeemMemoBytes)

		if !bytes.Equal(memoHashBytes, expectedRedeemMemoHashBytes) {
			Logger.log.Errorf("Memo redeem is invalid")
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// check whether amount transfer in txBNB is equal redeem amount or not
		// check receiver and amount in tx
		// get list matching custodians in waitingRedeemRequest

		outputs := txBNB.Msgs[0].(msg.SendMsg).Outputs

		remoteAddressNeedToBeTransfer := waitingRedeemRequest.GetRedeemerRemoteAddress()
		amountNeedToBeTransfer := meta.RedeemAmount
		amountNeedToBeTransferInBNB := convertIncPBNBAmountToExternalBNBAmount(int64(amountNeedToBeTransfer))

		isChecked := false
		for _, out := range outputs {
			addr, _ := bnb.GetAccAddressString(&out.Address, blockchain.config.ChainParams.BNBRelayingHeaderChainID)
			if addr != remoteAddressNeedToBeTransfer {
				continue
			}

			// calculate amount that was transferred to custodian's remote address
			amountTransfer := int64(0)
			for _, coin := range out.Coins {
				if coin.Denom == bnb.DenomBNB {
					amountTransfer += coin.Amount
					// note: log error for debug
					Logger.log.Infof("TxProof-BNB coin.Amount %d",
						coin.Amount)
				}
			}
			if amountTransfer < amountNeedToBeTransferInBNB {
				Logger.log.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal to or greater than %d, but got %d",
					addr, amountNeedToBeTransferInBNB, amountTransfer)
				inst := buildReqUnlockCollateralInst(
					meta.UniqueRedeemID,
					meta.TokenID,
					meta.CustodianAddressStr,
					meta.RedeemAmount,
					0,
					meta.RedeemProof,
					meta.Type,
					shardID,
					actionData.TxReqID,
					common.PortalReqUnlockCollateralRejectedChainStatus,
				)
				return [][]string{inst}, nil
			} else {
				isChecked = true
				break
			}
		}

		if !isChecked {
			Logger.log.Errorf("TxProof-BNB is invalid - Receiver address is invalid, expected %v",
				remoteAddressNeedToBeTransfer)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// calculate unlock amount
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.CustodianAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		unlockAmount, err2 := CalUnlockCollateralAmount(currentPortalState, custodianStateKeyStr, meta.RedeemAmount, meta.TokenID)
		if err2 != nil {
			Logger.log.Errorf("Error calculating unlock amount for custodian %v", err2)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// update custodian state (FreeCollateral, LockedAmountCollateral)
		err2 = updateCustodianStateAfterReqUnlockCollateral(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			unlockAmount, meta.TokenID)
		if err2 != nil {
			Logger.log.Errorf("Error when updating custodian state after unlocking collateral %v", err2)
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}

		// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
		updatedCustodians, err2 := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].GetCustodians(), meta.CustodianAddressStr)
		if err2 != nil {
			inst := buildReqUnlockCollateralInst(
				meta.UniqueRedeemID,
				meta.TokenID,
				meta.CustodianAddressStr,
				meta.RedeemAmount,
				0,
				meta.RedeemProof,
				meta.Type,
				shardID,
				actionData.TxReqID,
				common.PortalReqUnlockCollateralRejectedChainStatus,
			)
			return [][]string{inst}, nil
		}
		currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].SetCustodians(updatedCustodians)

		// remove redeem request from WaitingRedeemRequest list when all matching custodians return public token to user
		// when list matchingCustodianDetail is empty
		if len(currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr].GetCustodians()) == 0 {
			deleteWaitingRedeemRequest(currentPortalState, keyWaitingRedeemRequestStr)
		}

		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			unlockAmount,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralAcceptedChainStatus,
		)

		return [][]string{inst}, nil
	} else {
		Logger.log.Errorf("TokenID is not supported currently on Portal")
		inst := buildReqUnlockCollateralInst(
			meta.UniqueRedeemID,
			meta.TokenID,
			meta.CustodianAddressStr,
			meta.RedeemAmount,
			0,
			meta.RedeemProof,
			meta.Type,
			shardID,
			actionData.TxReqID,
			common.PortalReqUnlockCollateralRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}
}
