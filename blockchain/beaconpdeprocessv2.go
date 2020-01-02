package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPDEInstructionsV2(stateDB *statedb.StateDB, beaconBlock *BeaconBlock) error {
	beaconHeight := beaconBlock.Header.Height - 1
	currentPDEState, err := InitCurrentPDEStateFromDBV2(stateDB, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}
	for _, inst := range beaconBlock.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not PDE instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.PDEContributionMeta):
			err = blockchain.processPDEContributionV2_V2(stateDB, beaconHeight, inst, currentPDEState)
		case strconv.Itoa(metadata.PDETradeRequestMeta):
			err = blockchain.processPDETrade_V2(stateDB, beaconHeight, inst, currentPDEState)
		case strconv.Itoa(metadata.PDEWithdrawalRequestMeta):
			err = blockchain.processPDEWithdrawal_V2(stateDB, beaconHeight, inst, currentPDEState)
		}
		if err != nil {
			Logger.log.Error(err)
			return nil
		}
	}
	// store updated currentPDEState to leveldb with new beacon height
	err = storePDEStateToDBV2(
		stateDB,
		beaconHeight+1,
		currentPDEState,
	)
	if err != nil {
		Logger.log.Error(err)
	}
	return nil
}

func (blockchain *BlockChain) processPDEContributionV2_V2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instruction []string,
	currentPDEState *CurrentPDEState,
) error {
	if currentPDEState == nil {
		Logger.log.Warn("WARN - [processPDEContribution]: Current PDE state is null.")
		return nil
	}
	if len(instruction) != 4 {
		return nil // skip the instruction
	}
	contributionStatus := instruction[2]
	if contributionStatus == common.PDEContributionWaitingChainStatus {
		var waitingContribution metadata.PDEWaitingContribution
		err := json.Unmarshal([]byte(instruction[3]), &waitingContribution)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde waiting contribution instruction: %+v", err)
			return nil
		}
		waitingContribPairKey := string(rawdb.BuildWaitingPDEContributionKey(beaconHeight, waitingContribution.PDEContributionPairID))
		currentPDEState.WaitingPDEContributions[waitingContribPairKey] = &rawdb.PDEContribution{
			ContributorAddressStr: waitingContribution.ContributorAddressStr,
			TokenIDStr:            waitingContribution.TokenIDStr,
			Amount:                waitingContribution.ContributedAmount,
			TxReqID:               waitingContribution.TxReqID,
		}
		contribStatus := metadata.PDEContributionStatus{
			Status: byte(common.PDEContributionWaitingStatus),
		}
		contribStatusBytes, _ := json.Marshal(contribStatus)
		err = statedb.TrackPDEContributionStatus(
			stateDB,
			rawdb.PDEContributionStatusPrefix,
			[]byte(waitingContribution.PDEContributionPairID),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde waiting contribution status: %+v", err)
			return nil
		}

	} else if contributionStatus == common.PDEContributionRefundChainStatus {
		var refundContribution metadata.PDERefundContribution
		err := json.Unmarshal([]byte(instruction[3]), &refundContribution)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde refund contribution instruction: %+v", err)
			return nil
		}
		waitingContribPairKey := string(rawdb.BuildWaitingPDEContributionKey(beaconHeight, refundContribution.PDEContributionPairID))
		_, found := currentPDEState.WaitingPDEContributions[waitingContribPairKey]
		if found {
			delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
		}
		contribStatus := metadata.PDEContributionStatus{
			Status: byte(common.PDEContributionRefundStatus),
		}
		contribStatusBytes, _ := json.Marshal(contribStatus)
		err = statedb.TrackPDEContributionStatus(
			stateDB,
			rawdb.PDEContributionStatusPrefix,
			[]byte(refundContribution.PDEContributionPairID),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde refund contribution status: %+v", err)
			return nil
		}

	} else if contributionStatus == common.PDEContributionMatchedChainStatus {
		var matchedContribution metadata.PDEMatchedContribution
		err := json.Unmarshal([]byte(instruction[3]), &matchedContribution)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde matched contribution instruction: %+v", err)
			return nil
		}
		waitingContribPairKey := string(rawdb.BuildWaitingPDEContributionKey(beaconHeight, matchedContribution.PDEContributionPairID))
		existingWaitingContribution, found := currentPDEState.WaitingPDEContributions[waitingContribPairKey]
		if !found || existingWaitingContribution == nil {
			Logger.log.Errorf("ERROR: could not find out existing waiting contribution with unique pair id: %s", matchedContribution.PDEContributionPairID)
			return nil
		}
		incomingWaitingContribution := &rawdb.PDEContribution{
			ContributorAddressStr: matchedContribution.ContributorAddressStr,
			TokenIDStr:            matchedContribution.TokenIDStr,
			Amount:                matchedContribution.ContributedAmount,
			TxReqID:               matchedContribution.TxReqID,
		}
		updateWaitingContributionPairToPoolV2(
			beaconHeight,
			existingWaitingContribution,
			incomingWaitingContribution,
			currentPDEState,
		)
		delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
		contribStatus := metadata.PDEContributionStatus{
			Status: byte(common.PDEContributionAcceptedStatus),
		}
		contribStatusBytes, _ := json.Marshal(contribStatus)
		err = statedb.TrackPDEContributionStatus(
			stateDB,
			rawdb.PDEContributionStatusPrefix,
			[]byte(matchedContribution.PDEContributionPairID),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde accepted contribution status: %+v", err)
			return nil
		}

	} else if contributionStatus == common.PDEContributionMatchedNReturnedChainStatus {
		var matchedNReturnedContrib metadata.PDEMatchedNReturnedContribution
		err := json.Unmarshal([]byte(instruction[3]), &matchedNReturnedContrib)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde matched and returned contribution instruction: %+v", err)
			return nil
		}
		waitingContribPairKey := string(rawdb.BuildWaitingPDEContributionKey(beaconHeight, matchedNReturnedContrib.PDEContributionPairID))
		waitingContribution, found := currentPDEState.WaitingPDEContributions[waitingContribPairKey]
		if found && waitingContribution != nil {
			incomingWaitingContribution := &rawdb.PDEContribution{
				ContributorAddressStr: matchedNReturnedContrib.ContributorAddressStr,
				TokenIDStr:            matchedNReturnedContrib.TokenIDStr,
				Amount:                matchedNReturnedContrib.ActualContributedAmount,
				TxReqID:               matchedNReturnedContrib.TxReqID,
			}
			existingWaitingContribution := &rawdb.PDEContribution{
				ContributorAddressStr: waitingContribution.ContributorAddressStr,
				TokenIDStr:            waitingContribution.TokenIDStr,
				Amount:                matchedNReturnedContrib.ActualWaitingContribAmount,
				TxReqID:               waitingContribution.TxReqID,
			}
			updateWaitingContributionPairToPoolV2(
				beaconHeight,
				existingWaitingContribution,
				incomingWaitingContribution,
				currentPDEState,
			)
			delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
		}
		pdeStatusContentBytes, err := statedb.GetPDEContributionStatus(
			stateDB,
			rawdb.PDEContributionStatusPrefix,
			[]byte(matchedNReturnedContrib.PDEContributionPairID),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while getting pde contribution status: %+v", err)
			return nil
		}
		if len(pdeStatusContentBytes) == 0 {
			return nil
		}

		var contribStatus metadata.PDEContributionStatus
		err = json.Unmarshal(pdeStatusContentBytes, &contribStatus)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde contribution status: %+v", err)
			return nil
		}

		if contribStatus.Status != byte(common.PDEContributionMatchedNReturnedStatus) {
			contribStatus := metadata.PDEContributionStatus{
				Status:             byte(common.PDEContributionMatchedNReturnedStatus),
				TokenID1Str:        matchedNReturnedContrib.TokenIDStr,
				Contributed1Amount: matchedNReturnedContrib.ActualContributedAmount,
				Returned1Amount:    matchedNReturnedContrib.ReturnedContributedAmount,
			}
			contribStatusBytes, _ := json.Marshal(contribStatus)
			err := statedb.TrackPDEContributionStatus(
				stateDB,
				rawdb.PDEContributionStatusPrefix,
				[]byte(matchedNReturnedContrib.PDEContributionPairID),
				contribStatusBytes,
			)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while tracking pde contribution status: %+v", err)
				return nil
			}
		} else {
			var contribStatus metadata.PDEContributionStatus
			err := json.Unmarshal(pdeStatusContentBytes, &contribStatus)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while unmarshaling pde contribution status: %+v", err)
				return nil
			}
			contribStatus.TokenID2Str = matchedNReturnedContrib.TokenIDStr
			contribStatus.Contributed2Amount = matchedNReturnedContrib.ActualContributedAmount
			contribStatus.Returned2Amount = matchedNReturnedContrib.ReturnedContributedAmount
			contribStatusBytes, _ := json.Marshal(contribStatus)
			err = statedb.TrackPDEContributionStatus(
				stateDB,
				rawdb.PDEContributionStatusPrefix,
				[]byte(matchedNReturnedContrib.PDEContributionPairID),
				contribStatusBytes,
			)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while tracking pde contribution status: %+v", err)
				return nil
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) processPDETrade_V2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instruction []string,
	currentPDEState *CurrentPDEState,
) error {
	if len(instruction) != 4 {
		return nil // skip the instruction
	}
	if instruction[2] == common.PDETradeRefundChainStatus {
		contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade instruction: %+v", err)
			return nil
		}
		var pdeTradeReqAction metadata.PDETradeRequestAction
		err = json.Unmarshal(contentBytes, &pdeTradeReqAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade instruction: %+v", err)
			return nil
		}
		err = statedb.TrackPDEStatus(
			stateDB,
			rawdb.PDETradeStatusPrefix,
			pdeTradeReqAction.TxReqID[:],
			byte(common.PDETradeRefundStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde refund trade status: %+v", err)
		}
		return nil
	}
	var pdeTradeAcceptedContent metadata.PDETradeAcceptedContent
	err := json.Unmarshal([]byte(instruction[3]), &pdeTradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while unmarshaling PDETradeAcceptedContent: %+v", err)
		return nil
	}

	pdePoolForPairKey := string(rawdb.BuildPDEPoolForPairKey(beaconHeight, pdeTradeAcceptedContent.Token1IDStr, pdeTradeAcceptedContent.Token2IDStr))
	pdePoolForPair, found := currentPDEState.PDEPoolPairs[pdePoolForPairKey]
	if !found || pdePoolForPair == nil {
		Logger.log.Errorf("WARNING: could not find out pdePoolForPair with token ids: %s & %s", pdeTradeAcceptedContent.Token1IDStr, pdeTradeAcceptedContent.Token2IDStr)
		return nil
	}

	if pdeTradeAcceptedContent.Token1PoolValueOperation.Operator == "+" {
		pdePoolForPair.Token1PoolValue += pdeTradeAcceptedContent.Token1PoolValueOperation.Value
		pdePoolForPair.Token2PoolValue -= pdeTradeAcceptedContent.Token2PoolValueOperation.Value
	} else {
		pdePoolForPair.Token1PoolValue -= pdeTradeAcceptedContent.Token1PoolValueOperation.Value
		pdePoolForPair.Token2PoolValue += pdeTradeAcceptedContent.Token2PoolValueOperation.Value
	}
	err = statedb.TrackPDEStatus(
		stateDB,
		rawdb.PDETradeStatusPrefix,
		pdeTradeAcceptedContent.RequestedTxID[:],
		byte(common.PDETradeAcceptedStatus),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted trade status: %+v", err)
	}
	return nil
}

func (blockchain *BlockChain) processPDEWithdrawal_V2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instruction []string,
	currentPDEState *CurrentPDEState,
) error {
	if len(instruction) != 4 {
		return nil // skip the instruction
	}
	if instruction[2] == common.PDEWithdrawalRejectedChainStatus {
		contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
			return nil
		}
		var pdeWithdrawalRequestAction metadata.PDEWithdrawalRequestAction
		err = json.Unmarshal(contentBytes, &pdeWithdrawalRequestAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde withdrawal request action: %+v", err)
			return nil
		}
		err = statedb.TrackPDEStatus(
			stateDB,
			rawdb.PDEWithdrawalStatusPrefix,
			pdeWithdrawalRequestAction.TxReqID[:],
			byte(common.PDEWithdrawalRejectedStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde rejected withdrawal status: %+v", err)
		}
		return nil
	}

	var wdAcceptedContent metadata.PDEWithdrawalAcceptedContent
	err := json.Unmarshal([]byte(instruction[3]), &wdAcceptedContent)
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while unmarshaling PDEWithdrawalAcceptedContent: %+v", err)
		return nil
	}

	// update pde pool pair
	pdePoolForPairKey := string(rawdb.BuildPDEPoolForPairKey(
		beaconHeight,
		wdAcceptedContent.PairToken1IDStr,
		wdAcceptedContent.PairToken2IDStr,
	))
	pdePoolForPair, found := currentPDEState.PDEPoolPairs[pdePoolForPairKey]
	if !found || pdePoolForPair == nil {
		Logger.log.Errorf("WARNING: could not find out pdePoolForPair with token ids: %s & %s", wdAcceptedContent.PairToken1IDStr, wdAcceptedContent.PairToken2IDStr)
		return nil
	}
	if pdePoolForPair.Token1IDStr == wdAcceptedContent.WithdrawalTokenIDStr {
		pdePoolForPair.Token1PoolValue -= wdAcceptedContent.DeductingPoolValue
	} else {
		pdePoolForPair.Token2PoolValue -= wdAcceptedContent.DeductingPoolValue
	}

	// update pde shares
	deductSharesForWithdrawal(
		beaconHeight,
		wdAcceptedContent.PairToken1IDStr, wdAcceptedContent.PairToken2IDStr,
		wdAcceptedContent.WithdrawerAddressStr,
		wdAcceptedContent.DeductingShares,
		currentPDEState,
	)

	err = statedb.TrackPDEStatus(
		stateDB,
		rawdb.PDEWithdrawalStatusPrefix,
		wdAcceptedContent.TxReqID[:],
		byte(common.PDEWithdrawalAcceptedStatus),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted withdrawal status: %+v", err)
	}
	return nil
}
