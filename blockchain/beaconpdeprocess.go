package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/core/rawdb"
	"github.com/incognitochain/incognito-chain/incdb"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPDEInstructions(block *BeaconBlock, bd *[]incdb.BatchData) error {
	beaconHeight := block.Header.Height - 1
	db := blockchain.GetDatabase()
	currentPDEState, err := InitCurrentPDEStateFromDB(db, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil
	}
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not PDE instruction
		}
		var err error
		switch inst[0] {
		case strconv.Itoa(metadata.PDEContributionMeta):
			err = blockchain.processPDEContributionV2(beaconHeight, inst, currentPDEState)
		case strconv.Itoa(metadata.PDETradeRequestMeta):
			err = blockchain.processPDETrade(beaconHeight, inst, currentPDEState)
		case strconv.Itoa(metadata.PDEWithdrawalRequestMeta):
			err = blockchain.processPDEWithdrawal(beaconHeight, inst, currentPDEState)
		}
		if err != nil {
			Logger.log.Error(err)
			return nil
		}
	}
	// store updated currentPDEState to leveldb with new beacon height
	err = storePDEStateToDB(
		db,
		beaconHeight+1,
		currentPDEState,
	)
	if err != nil {
		Logger.log.Error(err)
	}
	return nil
}

func storePDEPoolForPair(
	pdePoolForPairKey string,
	token1IDStr string,
	token1PoolValue uint64,
	token2IDStr string,
	token2PoolValue uint64,
	currentPDEState *CurrentPDEState,
) {
	pdePoolForPair := &rawdb.PDEPoolForPair{
		Token1IDStr:     token1IDStr,
		Token1PoolValue: token1PoolValue,
		Token2IDStr:     token2IDStr,
		Token2PoolValue: token2PoolValue,
	}
	currentPDEState.PDEPoolPairs[pdePoolForPairKey] = pdePoolForPair
}

func (blockchain *BlockChain) processPDEContributionV2(
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
	if contributionStatus == "waiting" {
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
		err = rawdb.TrackPDEStatus(
			blockchain.config.DataBase,
			rawdb.PDEContributionStatusPrefix,
			[]byte(waitingContribution.PDEContributionPairID),
			byte(common.PDEContributionWaitingStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde waiting contribution status: %+v", err)
			return nil
		}

	} else if contributionStatus == "refund" {
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
		err = rawdb.TrackPDEStatus(
			blockchain.config.DataBase,
			rawdb.PDEContributionStatusPrefix,
			[]byte(refundContribution.PDEContributionPairID),
			byte(common.PDEContributionRefundStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde refund contribution status: %+v", err)
			return nil
		}

	} else if contributionStatus == "matched" {
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
		err = rawdb.TrackPDEStatus(
			blockchain.config.DataBase,
			rawdb.PDEContributionStatusPrefix,
			[]byte(matchedContribution.PDEContributionPairID),
			byte(common.PDEContributionAcceptedStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde accepted contribution status: %+v", err)
			return nil
		}
	}
	return nil
}

func (blockchain *BlockChain) processPDETrade(
	beaconHeight uint64,
	instruction []string,
	currentPDEState *CurrentPDEState,
) error {
	if len(instruction) != 4 {
		return nil // skip the instruction
	}
	if instruction[2] == "refund" {
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
		err = rawdb.TrackPDEStatus(
			blockchain.config.DataBase,
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
	err = rawdb.TrackPDEStatus(
		blockchain.config.DataBase,
		rawdb.PDETradeStatusPrefix,
		pdeTradeAcceptedContent.RequestedTxID[:],
		byte(common.PDETradeAcceptedStatus),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted trade status: %+v", err)
	}
	return nil
}

func deductSharesForWithdrawal(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	withdrawerAddressStr string,
	amt uint64,
	currentPDEState *CurrentPDEState,
) {
	pdeShareKey := string(rawdb.BuildPDESharesKeyV2(beaconHeight, token1IDStr, token2IDStr, withdrawerAddressStr))
	adjustingAmt := uint64(0)
	currentAmt, found := currentPDEState.PDEShares[pdeShareKey]
	if found && amt <= currentAmt {
		adjustingAmt = currentAmt - amt
	}
	currentPDEState.PDEShares[pdeShareKey] = adjustingAmt
}

func (blockchain *BlockChain) processPDEWithdrawal(
	beaconHeight uint64,
	instruction []string,
	currentPDEState *CurrentPDEState,
) error {
	if len(instruction) != 4 {
		return nil // skip the instruction
	}
	if instruction[2] == "rejected" {
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
		err = rawdb.TrackPDEStatus(
			blockchain.GetDatabase(),
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

	err = rawdb.TrackPDEStatus(
		blockchain.GetDatabase(),
		rawdb.PDEWithdrawalStatusPrefix,
		wdAcceptedContent.TxReqID[:],
		byte(common.PDEWithdrawalAcceptedStatus),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted withdrawal status: %+v", err)
	}
	return nil
}
