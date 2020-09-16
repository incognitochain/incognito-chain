package blockchain

import (
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

/* =======
Portal Porting Request Processor
======= */
type portalPortingRequestProcessor struct {
	*portalInstProcessor
}

func (p *portalPortingRequestProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalPortingRequestProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalPortingRequestProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while decoding content string of portal porting request action: %+v", err)
		return nil, fmt.Errorf("Porting request: an error occurred while decoding content string of portal porting request action: %+v", err)
	}

	var actionData metadata.PortalUserRegisterAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while unmarshal portal porting request action: %+v", err)
		return nil, fmt.Errorf("Porting request: an error occurred while unmarshal portal porting request action: %+v", err)
	}

	optionalData := make(map[string]interface{})

	// Get porting request with uniqueID from stateDB
	isExistPortingID, err := statedb.IsPortingRequestIdExist(stateDB, []byte(actionData.Meta.UniqueRegisterId))
	if err != nil{
		Logger.log.Errorf("Porting request: an error occurred while get porting request Id from DB: %+v", err)
		return nil, fmt.Errorf("Porting request: an error occurred while get porting request Id from DB: %+v", err)
	}

	optionalData["isExistPortingID"] = isExistPortingID
	return optionalData, nil
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

func (p *portalPortingRequestProcessor) buildNewInsts(
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
		Logger.log.Errorf("Porting request: an error occurred while decoding content string of portal porting request action: %+v", err)
		return [][]string{}, nil
	}

	var actionData metadata.PortalUserRegisterAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("Porting request: an error occurred while unmarshal portal porting request action: %+v", err)
		return [][]string{}, nil
	}

	rejectInst := buildRequestPortingInst(
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

	if currentPortalState == nil {
		Logger.log.Errorf("Porting request: Current Portal state is null")
		return [][]string{rejectInst}, nil
	}

	// check unique id from optionalData which get from statedb
	if optionalData == nil {
		Logger.log.Errorf("Porting request: optionalData is null")
		return [][]string{rejectInst}, nil
	}
	isExist, ok := optionalData["isExistPortingID"].(bool)
	if !ok {
		Logger.log.Errorf("Porting request: optionalData isExistPortingID is invalid")
		return [][]string{rejectInst}, nil
	}
	if isExist {
		Logger.log.Errorf("Porting request: Porting request id exist in db %v", actionData.Meta.UniqueRegisterId)
		return [][]string{rejectInst}, nil
	}

	waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.Meta.UniqueRegisterId)
	if _, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]; ok {
		Logger.log.Errorf("Porting request: Waiting porting request exist, key %v", waitingPortingRequestKey)
		return [][]string{rejectInst}, nil
	}

	//get exchange rates
	exchangeRatesState := currentPortalState.FinalExchangeRatesState
	if exchangeRatesState == nil {
		Logger.log.Errorf("Porting request, exchange rates not found")
		return [][]string{rejectInst}, nil
	}

	if len(currentPortalState.CustodianPoolState) <= 0 {
		Logger.log.Errorf("Porting request: Custodian not found")
		return [][]string{rejectInst}, nil
	}

	var sortCustodianStateByFreeCollateral []CustodianStateSlice
	sortCustodianByAmountAscent(actionData.Meta, currentPortalState.CustodianPoolState, &sortCustodianStateByFreeCollateral)

	if len(sortCustodianStateByFreeCollateral) <= 0 {
		Logger.log.Errorf("Porting request, custodian not found")
		return [][]string{rejectInst}, nil
	}

	//validation porting fees
	exchangePortingFees, err := CalMinPortingFee(actionData.Meta.RegisterAmount, actionData.Meta.PTokenId, exchangeRatesState, portalParams.MinPercentPortingFee)
	if err != nil {
		Logger.log.Errorf("Porting request: Calculate Porting fee is error %v", err)
		return [][]string{rejectInst}, nil
	}

	if actionData.Meta.PortingFee < exchangePortingFees {
		Logger.log.Errorf("Porting request: Porting fees is invalid: expected %v - get %v", exchangePortingFees, actionData.Meta.PortingFee)
		return [][]string{rejectInst}, nil
	}

	// pick-up custodians
	pickedCustodians, err := pickUpCustodians(actionData.Meta, exchangeRatesState, sortCustodianStateByFreeCollateral, currentPortalState, portalParams)
	if err != nil || len(pickedCustodians) == 0 {
		Logger.log.Errorf("Porting request: an error occurred while picking up custodians for the porting request: %+v", err)
		return [][]string{rejectInst}, nil
	}

	// Update custodian state after finishing choosing enough custodians for the porting request
	for _, cus := range pickedCustodians {
		cusKey := statedb.GenerateCustodianStateObjectKey(cus.IncAddress).String()
		// update custodian state
		err := UpdateCustodianStateAfterMatchingPortingRequest(currentPortalState, cusKey, actionData.Meta.PTokenId, cus.LockedAmountCollateral)
		if err != nil {
			Logger.log.Errorf("Porting request: an error occurred while updating custodian state: %+v", err)
			return [][]string{rejectInst}, nil
		}
	}

	acceptInst := buildRequestPortingInst(
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

	return [][]string{acceptInst}, nil
}

/* =======
Portal Request Ptoken Processor
======= */

type portalRequestPTokenProcessor struct {
	*portalInstProcessor
}

func (p *portalRequestPTokenProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRequestPTokenProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRequestPTokenProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
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

func (p *portalRequestPTokenProcessor) buildNewInsts(
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
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal request ptoken action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRequestPTokensAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal request ptoken action: %+v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	rejectInst := buildReqPTokensInst(
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

	if currentPortalState == nil {
		Logger.log.Warn("Request PTokens: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	// check meta.UniquePortingID is in waiting PortingRequests list in portal state or not
	portingID := meta.UniquePortingID
	keyWaitingPortingRequestStr := statedb.GeneratePortalWaitingPortingRequestObjectKey(portingID).String()
	waitingPortingRequest := currentPortalState.WaitingPortingRequests[keyWaitingPortingRequestStr]
	if waitingPortingRequest == nil {
		Logger.log.Errorf("PortingID is not existed in waiting porting requests list")
		return [][]string{rejectInst}, nil
	}

	// check tokenID
	if meta.TokenID != waitingPortingRequest.TokenID() {
		Logger.log.Errorf("TokenID is not correct in portingID req")
		return [][]string{rejectInst}, nil
	}

	// check porting amount
	if meta.PortingAmount != waitingPortingRequest.Amount() {
		Logger.log.Errorf("PortingAmount is not correct in portingID req")
		return [][]string{rejectInst}, nil
	}

	if meta.TokenID == common.PortalBTCIDStr {
		btcChain := bc.config.BTCChain
		if btcChain == nil {
			Logger.log.Error("BTC relaying chain should not be null")
			return [][]string{rejectInst}, nil
		}
		// parse PortingProof in meta
		btcTxProof, err := btcrelaying.ParseBTCProofFromB64EncodeStr(meta.PortingProof)
		if err != nil {
			Logger.log.Errorf("PortingProof is invalid %v\n", err)
			return [][]string{rejectInst}, nil
		}

		isValid, err := btcChain.VerifyTxWithMerkleProofs(btcTxProof)
		if !isValid || err != nil {
			Logger.log.Errorf("Verify btcTxProof failed %v", err)
			return [][]string{rejectInst}, nil
		}

		// extract attached message from txOut's OP_RETURN
		btcAttachedMsg, err := btcrelaying.ExtractAttachedMsgFromTx(btcTxProof.BTCTx)
		if err != nil {
			Logger.log.Errorf("Could not extract attached message from BTC tx proof with err: %v", err)
			return [][]string{rejectInst}, nil
		}

		encodedMsg := btcrelaying.HashAndEncodeBase58(meta.UniquePortingID)
		if btcAttachedMsg != encodedMsg {
			Logger.log.Errorf("PortingId in the btc attached message is not matched with portingID in metadata")
			return [][]string{rejectInst}, nil
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
					return [][]string{rejectInst}, nil
				} else {
					isChecked = true
					break
				}
			}
			if !isChecked {
				Logger.log.Error("BTC-TxProof is invalid")
				return [][]string{rejectInst}, nil
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
			return [][]string{rejectInst}, nil
		}

		// check minimum confirmations block of bnb proof
		latestBNBBlockHeight, err2 := bc.GetLatestBNBBlkHeight()
		if err2 != nil {
			Logger.log.Errorf("Can not get latest relaying bnb block height %v\n", err)
			return [][]string{rejectInst}, nil
		}

		if latestBNBBlockHeight < txProofBNB.BlockHeight+bnb.MinConfirmationsBlock {
			Logger.log.Errorf("Not enough min bnb confirmations block %v, latestBNBBlockHeight %v - txProofBNB.BlockHeight %v\n",
				bnb.MinConfirmationsBlock, latestBNBBlockHeight, txProofBNB.BlockHeight)
			return [][]string{rejectInst}, nil
		}
		dataHash, err2 := bc.GetBNBDataHash(txProofBNB.BlockHeight)
		if err2 != nil {
			Logger.log.Errorf("Error when get data hash in blockHeight %v - %v\n",
				txProofBNB.BlockHeight, err2)
			return [][]string{rejectInst}, nil
		}

		isValid, err := txProofBNB.Verify(dataHash)
		if !isValid || err != nil {
			Logger.log.Errorf("Verify txProofBNB failed %v", err)
			return [][]string{rejectInst}, nil
		}

		// parse Tx from Data in txProofBNB
		txBNB, err := bnb.ParseTxFromData(txProofBNB.Proof.Data)
		if err != nil {
			Logger.log.Errorf("Data in PortingProof is invalid %v", err)
			return [][]string{rejectInst}, nil
		}

		// check memo attach portingID req:
		memo := txBNB.Memo
		memoBytes, err2 := base64.StdEncoding.DecodeString(memo)
		if err2 != nil {
			Logger.log.Errorf("Can not decode memo in tx bnb proof", err2)
			return [][]string{rejectInst}, nil
		}

		var portingMemo PortingMemoBNB
		err2 = json.Unmarshal(memoBytes, &portingMemo)
		if err2 != nil {
			Logger.log.Errorf("Can not unmarshal memo in tx bnb proof", err2)
			return [][]string{rejectInst}, nil
		}

		if portingMemo.PortingID != meta.UniquePortingID {
			Logger.log.Errorf("PortingId in memoTx is not matched with portingID in metadata", err2)
			return [][]string{rejectInst}, nil
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
				addr, _ := bnb.GetAccAddressString(&out.Address, bc.config.ChainParams.BNBRelayingHeaderChainID)
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
					return [][]string{rejectInst}, nil
				} else {
					isChecked = true
					break
				}
			}
			if !isChecked {
				Logger.log.Errorf("TxProof-BNB is invalid - Receiver address is invalid, expected %v",
					remoteAddressNeedToBeTransfer)
				return [][]string{rejectInst}, nil
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
		return [][]string{rejectInst}, nil
	}
}
