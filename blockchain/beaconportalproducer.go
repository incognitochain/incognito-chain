package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	relaying "github.com/incognitochain/incognito-chain/relaying/bnb"
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
	pTokenAddress string,
	registerAmount uint64,
	portingFee uint64,
	custodian map[string]lvdb.MatchingPortingCustodianDetail,
	txReqID common.Hash,
) []string {
	portingRequestContent := metadata.PortalPortingRequestContent{
		UniqueRegisterId: uniqueRegisterId,
		IncogAddressStr:  incogAddressStr,
		PTokenId:         pTokenId,
		PTokenAddress:    pTokenAddress,
		RegisterAmount:   registerAmount,
		PortingFee:       portingFee,
		Custodian:        custodian,
		TxReqID:          txReqID,
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

// buildInstructionsForCustodianDeposit builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForCustodianDeposit(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
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
	meta := actionData.Meta

	keyCustodianState := lvdb.NewCustodianStateKey(beaconHeight, meta.IncogAddressStr)

	if currentPortalState.CustodianPoolState[keyCustodianState] == nil {
		// new custodian
		newCustodian, _ := NewCustodianState(meta.IncogAddressStr, meta.DepositedAmount, meta.DepositedAmount, nil, nil, meta.RemoteAddresses)
		currentPortalState.CustodianPoolState[keyCustodianState] = newCustodian
	} else {
		// custodian deposited before
		// update state of the custodian
		custodian := currentPortalState.CustodianPoolState[keyCustodianState]
		totalCollateral := custodian.TotalCollateral + meta.DepositedAmount
		freeCollateral := custodian.FreeCollateral + meta.DepositedAmount
		holdingPubTokens := custodian.HoldingPubTokens
		lockedAmountCollateral := custodian.LockedAmountCollateral
		remoteAddresses := custodian.RemoteAddresses
		for tokenSymbol, address := range meta.RemoteAddresses {
			if remoteAddresses[tokenSymbol] == "" {
				remoteAddresses[tokenSymbol] = address
			}
		}

		newCustodian, _ := NewCustodianState(meta.IncogAddressStr, totalCollateral, freeCollateral, holdingPubTokens, lockedAmountCollateral, remoteAddresses)
		currentPortalState.CustodianPoolState[keyCustodianState] = newCustodian
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

func (blockchain *BlockChain) buildInstructionsForPortingRequest(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
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

	db := blockchain.GetDatabase()

	//check unique id from record from db
	keyPortingRequest := lvdb.GetNewPortingRequestKeyValid(actionData.Meta.UniqueRegisterId)
	portingRequestExist, err := db.GetItemPortalByKey([]byte(keyPortingRequest))

	if err != nil {
		Logger.log.Errorf("Porting request: Get item portal by prefix error: %+v", err)

		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.PTokenAddress,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	if portingRequestExist != nil {
		Logger.log.Errorf("Porting request: Porting request exist")
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.PTokenAddress,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	waitingPortingRequestKey := lvdb.NewWaitingPortingReqKey(beaconHeight, actionData.Meta.UniqueRegisterId)
	if _, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey]; ok {
		Logger.log.Errorf("Porting request: Waiting porting request exist, key %v", waitingPortingRequestKey)
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.PTokenAddress,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	//get exchange rates
	exchangeRatesKey := lvdb.NewFinalExchangeRatesKey(beaconHeight)
	exchangeRatesState, ok := currentPortalState.FinalExchangeRates[exchangeRatesKey]
	if !ok {
		Logger.log.Errorf("Porting request, exchange rates not found")
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.PTokenAddress,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	//todo: create error instruction
	if currentPortalState.CustodianPoolState == nil {
		Logger.log.Errorf("Porting request: Custodian not found")
		return [][]string{}, nil
	}

	var sortCustodianStateByFreeCollateral []CustodianStateSlice
	err = sortCustodianByAmountAscent(actionData.Meta, currentPortalState.CustodianPoolState, &sortCustodianStateByFreeCollateral)

	if err != nil {
		return [][]string{}, nil
	}

	if len(sortCustodianStateByFreeCollateral) <= 0 {
		Logger.log.Errorf("Porting request, custodian not found")

		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.PTokenAddress,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			nil,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	//pick one
	pickCustodianResult, _ := pickSingleCustodian(actionData.Meta, exchangeRatesState, sortCustodianStateByFreeCollateral)

	Logger.log.Infof("Porting request, pick single custodian result %v", len(pickCustodianResult))
	//pick multiple
	if len(pickCustodianResult) == 0 {
		pickCustodianResult, _ = pickMultipleCustodian(actionData.Meta, exchangeRatesState, sortCustodianStateByFreeCollateral)
		Logger.log.Infof("Porting request, pick multiple custodian result %v", len(pickCustodianResult))
	}

	//end
	if len(pickCustodianResult) == 0 {
		Logger.log.Errorf("Porting request, custodian not found")
		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.PTokenAddress,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			pickCustodianResult,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	//validation porting fees
	pToken2PRV := exchangeRatesState.ExchangePToken2PRVByTokenId(actionData.Meta.PTokenId, actionData.Meta.RegisterAmount)
	exchangePortingFees := CalculatePortingFees(pToken2PRV)
	Logger.log.Infof("Porting request, porting fees need %v", exchangePortingFees)

	if actionData.Meta.PortingFee < exchangePortingFees {
		Logger.log.Errorf("Porting request, Porting fees is wrong")

		inst := buildRequestPortingInst(
			actionData.Meta.Type,
			shardID,
			common.PortalPortingRequestRejectedStatus,
			actionData.Meta.UniqueRegisterId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.PTokenId,
			actionData.Meta.PTokenAddress,
			actionData.Meta.RegisterAmount,
			actionData.Meta.PortingFee,
			pickCustodianResult,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	inst := buildRequestPortingInst(
		actionData.Meta.Type,
		shardID,
		common.PortalPortingRequestAcceptedStatus,
		actionData.Meta.UniqueRegisterId,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.PTokenId,
		actionData.Meta.PTokenAddress,
		actionData.Meta.RegisterAmount,
		actionData.Meta.PortingFee,
		pickCustodianResult,
		actionData.TxReqID,
	) //return  metadata.PortalPortingRequestContent at instruct[3]

	return [][]string{inst}, nil
}

// buildInstructionsForCustodianDeposit builds instruction for custodian deposit action
func (blockchain *BlockChain) buildInstructionsForReqPTokens(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
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
	keyWaitingPortingRequest := lvdb.NewWaitingPortingReqKey(beaconHeight, portingID)
	waitingPortingRequest := currentPortalState.WaitingPortingRequests[keyWaitingPortingRequest]
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
	db := blockchain.GetDatabase()

	// check porting request status of portingID from db
	portingReqStatus, err := db.GetPortingRequestStatusByPortingID(meta.UniquePortingID)
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
	//if len(portingReqStatus) > 0 {
	//	//todo: need to change to portal request status
	//	reqPTokenStatus := metadata.PortalRequestPTokensStatus{}
	//	err := json.Unmarshal(portingReqStatus, &reqPTokenStatus)
	//	if err != nil {
	//		Logger.log.Errorf("Can not unmarshal req ptoken status %v\n", err)
	//		inst := buildReqPTokensInst(
	//			meta.UniquePortingID,
	//			meta.TokenID,
	//			meta.IncogAddressStr,
	//			meta.PortingAmount,
	//			meta.PortingProof,
	//			meta.Type,
	//			shardID,
	//			actionData.TxReqID,
	//			common.PortalReqPTokensRejectedChainStatus,
	//		)
	//		return [][]string{inst}, nil
	//	}
	if portingReqStatus != common.PortalPortingReqWaitingStatus {
		Logger.log.Errorf("PortingID status invalid")
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
	if meta.TokenID != metadata.PortalSupportedTokenMap[waitingPortingRequest.TokenID] {
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
	if meta.PortingAmount != waitingPortingRequest.Amount {
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

	if meta.TokenID == metadata.PortalSupportedTokenMap[metadata.PortalTokenSymbolBTC] {
		//todo:
	} else if meta.TokenID == metadata.PortalSupportedTokenMap[metadata.PortalTokenSymbolBNB] {
		// parse PortingProof in meta
		txProofBNB, err := relaying.ParseBNBProofFromB64EncodeJsonStr(meta.PortingProof)
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

		isValid, err := txProofBNB.Verify(db)
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
		txBNB, err := relaying.ParseTxFromData(txProofBNB.Proof.Data)
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
		type PortingMemoBNB struct {
			PortingID string `json:"PortingID"`
		}
		memo := txBNB.Memo
		Logger.log.Infof("[buildInstructionsForReqPTokens] memo: %v\n", memo)
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
		Logger.log.Infof("[buildInstructionsForReqPTokens] memoBytes: %v\n", memoBytes)

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
		custodians := waitingPortingRequest.Custodians
		//custodians := make(map[string]lvdb.MatchingPortingCustodianDetail, 1)
		//custodians["12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ"] = lvdb.MatchingPortingCustodianDetail{
		//	RemoteAddress: "tbnb1v63crn5slveu50v8x590uwmqf7kk5xca74scwx",
		//	Amount: 10000000000,   // 10 bnb
		//}

		outputs := txBNB.Msgs[0].(msg.SendMsg).Outputs

		for _, cusDetail := range custodians {
			remoteAddressNeedToBeTransfer := cusDetail.RemoteAddress
			amountNeedToBeTransfer := cusDetail.Amount

			for _, out := range outputs {
				addr := string(out.Address)
				if addr != remoteAddressNeedToBeTransfer {
					continue
				}

				// calculate amount that was transferred to custodian's remote address
				amountTransfer := int64(0)
				for _, coin := range out.Coins {
					if coin.Denom == relaying.DenomBNB {
						amountTransfer += coin.Amount
						// note: log error for debug
						Logger.log.Errorf("TxProof-BNB coin.Amount %d",
							coin.Amount)
					}
				}

				//TODO:
				if amountTransfer*10^9 != int64(amountNeedToBeTransfer) {
					Logger.log.Errorf("TxProof-BNB is invalid - Amount transfer to %s must be equal %d, but got %d",
						addr, amountNeedToBeTransfer, amountTransfer)
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
		removeWaitingPortingReqByKey(keyWaitingPortingRequest, currentPortalState)
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

	return [][]string{}, nil
}

func (blockchain *BlockChain) buildInstructionsForExchangeRates(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
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

	exchangeRatesKey := lvdb.NewExchangeRatesRequestKey(
		beaconHeight+1,
		actionData.TxReqID.String(),
		strconv.FormatInt(actionData.LockTime, 10),
		shardID,
	)

	db := blockchain.GetDatabase()
	//check key from db
	exchangeRatesKeyExist, err := db.GetItemPortalByKey([]byte(exchangeRatesKey))
	if err != nil {
		Logger.log.Errorf("ERROR: Get exchange rates error: %+v", err)

		portalExchangeRatesContent := metadata.PortalExchangeRatesContent{
			SenderAddress:   actionData.Meta.SenderAddress,
			Rates:           actionData.Meta.Rates,
			TxReqID:         actionData.TxReqID,
			LockTime:        actionData.LockTime,
			UniqueRequestId: exchangeRatesKey,
		}

		portalExchangeRatesContentBytes, _ := json.Marshal(portalExchangeRatesContent)

		inst := []string{
			strconv.Itoa(metaType),
			strconv.Itoa(int(shardID)),
			common.PortalExchangeRatesRejectedStatus,
			string(portalExchangeRatesContentBytes),
		}

		return [][]string{inst}, nil
	}

	if exchangeRatesKeyExist != nil {
		Logger.log.Errorf("ERROR: exchange rates key is duplicated")

		portalExchangeRatesContent := metadata.PortalExchangeRatesContent{
			SenderAddress:   actionData.Meta.SenderAddress,
			Rates:           actionData.Meta.Rates,
			TxReqID:         actionData.TxReqID,
			LockTime:        actionData.LockTime,
			UniqueRequestId: exchangeRatesKey,
		}

		portalExchangeRatesContentBytes, _ := json.Marshal(portalExchangeRatesContent)

		inst := []string{
			strconv.Itoa(metaType),
			strconv.Itoa(int(shardID)),
			common.PortalExchangeRatesRejectedStatus,
			string(portalExchangeRatesContentBytes),
		}

		return [][]string{inst}, nil
	}

	//success
	portalExchangeRatesContent := metadata.PortalExchangeRatesContent{
		SenderAddress:   actionData.Meta.SenderAddress,
		Rates:           actionData.Meta.Rates,
		TxReqID:         actionData.TxReqID,
		LockTime:        actionData.LockTime,
		UniqueRequestId: exchangeRatesKey,
	}

	portalExchangeRatesContentBytes, _ := json.Marshal(portalExchangeRatesContent)

	inst := []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PortalExchangeRatesSuccessStatus,
		string(portalExchangeRatesContentBytes),
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
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	redeemRequestContent := metadata.PortalRedeemRequestContent{
		UniqueRedeemID: uniqueRedeemID,
		TokenID:        tokenID,
		RedeemAmount:   redeemAmount,
		IncAddressStr:  incAddressStr,
		RemoteAddress:  remoteAddress,
		RedeemFee:      redeemFee,
		TxReqID:        txReqID,
		ShardID:        shardID,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

//todo
// buildInstructionsForRedeemRequest builds instruction for redeem request action
func (blockchain *BlockChain) buildInstructionsForRedeemRequest(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
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
			meta.IncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemRequestRejectedStatus,
		)
		return [][]string{inst}, nil
	}

	// check uniqueRedeemID is existed in db and waitingRedeem list or not
	// pick custodian(s) who holding public token to return user
	// add to waiting Redeem list
	//




	//keyCustodianState := lvdb.NewCustodianStateKey(beaconHeight, meta.IncogAddressStr)
	//
	//if currentPortalState.CustodianPoolState[keyCustodianState] == nil {
	//	// new custodian
	//	newCustodian, _ := NewCustodianState(meta.IncogAddressStr, meta.DepositedAmount, meta.DepositedAmount, nil, nil, meta.RemoteAddresses)
	//	currentPortalState.CustodianPoolState[keyCustodianState] = newCustodian
	//} else {
	//	// custodian deposited before
	//	// update state of the custodian
	//	custodian := currentPortalState.CustodianPoolState[keyCustodianState]
	//	totalCollateral := custodian.TotalCollateral + meta.DepositedAmount
	//	freeCollateral := custodian.FreeCollateral + meta.DepositedAmount
	//	holdingPubTokens := custodian.HoldingPubTokens
	//	lockedAmountCollateral := custodian.LockedAmountCollateral
	//	remoteAddresses := custodian.RemoteAddresses
	//	for tokenSymbol, address := range meta.RemoteAddresses {
	//		if remoteAddresses[tokenSymbol] == "" {
	//			remoteAddresses[tokenSymbol] = address
	//		}
	//	}
	//
	//	newCustodian, _ := NewCustodianState(meta.IncogAddressStr, totalCollateral, freeCollateral, holdingPubTokens, lockedAmountCollateral, remoteAddresses)
	//	currentPortalState.CustodianPoolState[keyCustodianState] = newCustodian
	//}
	//
	//inst := buildCustodianDepositInst(
	//	actionData.Meta.IncogAddressStr,
	//	actionData.Meta.DepositedAmount,
	//	actionData.Meta.RemoteAddresses,
	//	actionData.Meta.Type,
	//	shardID,
	//	actionData.TxReqID,
	//	common.PortalCustodianDepositAcceptedChainStatus,
	//)
	return [][]string{}, nil
}
