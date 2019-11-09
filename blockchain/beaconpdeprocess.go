package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPDEInstructions(block *BeaconBlock, bd *[]database.BatchData) error {
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

func addShareAmountUp(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddrStr string,
	amt uint64,
	currentPDEState *CurrentPDEState,
) {
	pdeShareOnTokenPrefix := string(lvdb.BuildPDESharesKey(beaconHeight, token1IDStr, token2IDStr, contributedTokenIDStr, ""))
	totalSharesOnToken := uint64(0)
	for key, value := range currentPDEState.PDEShares {
		if strings.Contains(key, pdeShareOnTokenPrefix) {
			totalSharesOnToken += value
		}
	}
	pdeShareKey := string(lvdb.BuildPDESharesKey(beaconHeight, token1IDStr, token2IDStr, contributedTokenIDStr, contributorAddrStr))
	if totalSharesOnToken == 0 {
		currentPDEState.PDEShares[pdeShareKey] = amt
		return
	}
	poolPairKey := string(lvdb.BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr))
	poolPair, found := currentPDEState.PDEPoolPairs[poolPairKey]
	if !found || poolPair == nil {
		currentPDEState.PDEShares[pdeShareKey] = amt
		return
	}
	poolValue := poolPair.Token1PoolValue
	if poolPair.Token2IDStr == contributedTokenIDStr {
		poolValue = poolPair.Token2PoolValue
	}
	if poolValue == 0 {
		currentPDEState.PDEShares[pdeShareKey] = amt
	}
	increasingAmt := big.NewInt(0)
	increasingAmt.Mul(big.NewInt(int64(totalSharesOnToken)), big.NewInt(int64(amt)))
	increasingAmt.Div(increasingAmt, big.NewInt(int64(poolValue)))
	currentShare, found := currentPDEState.PDEShares[pdeShareKey]
	addedUpAmt := increasingAmt.Uint64()
	if found {
		addedUpAmt += currentShare
	}
	currentPDEState.PDEShares[pdeShareKey] = addedUpAmt
}

func storePDEPoolForPair(
	pdePoolForPairKey string,
	token1IDStr string,
	token1PoolValue uint64,
	token2IDStr string,
	token2PoolValue uint64,
	currentPDEState *CurrentPDEState,
) {
	pdePoolForPair := &lvdb.PDEPoolForPair{
		Token1IDStr:     token1IDStr,
		Token1PoolValue: token1PoolValue,
		Token2IDStr:     token2IDStr,
		Token2PoolValue: token2PoolValue,
	}
	currentPDEState.PDEPoolPairs[pdePoolForPairKey] = pdePoolForPair
}

func updateWaitingContributionPairToPool(
	beaconHeight uint64,
	waitingContribution1 *lvdb.PDEContribution,
	waitingContribution2 *lvdb.PDEContribution,
	currentPDEState *CurrentPDEState,
) {
	addShareAmountUp(
		beaconHeight,
		waitingContribution1.TokenIDStr,
		waitingContribution2.TokenIDStr,
		waitingContribution1.TokenIDStr,
		waitingContribution1.ContributorAddressStr,
		waitingContribution1.Amount,
		currentPDEState,
	)
	addShareAmountUp(
		beaconHeight,
		waitingContribution1.TokenIDStr,
		waitingContribution2.TokenIDStr,
		waitingContribution2.TokenIDStr,
		waitingContribution2.ContributorAddressStr,
		waitingContribution2.Amount,
		currentPDEState,
	)
	waitingContributions := []*lvdb.PDEContribution{waitingContribution1, waitingContribution2}
	sort.Slice(waitingContributions, func(i, j int) bool {
		return waitingContributions[i].TokenIDStr < waitingContributions[j].TokenIDStr
	})
	pdePoolForPairKey := string(lvdb.BuildPDEPoolForPairKey(beaconHeight, waitingContributions[0].TokenIDStr, waitingContributions[1].TokenIDStr))
	pdePoolForPair, found := currentPDEState.PDEPoolPairs[pdePoolForPairKey]
	if !found || pdePoolForPair == nil {
		storePDEPoolForPair(
			pdePoolForPairKey,
			waitingContributions[0].TokenIDStr,
			waitingContributions[0].Amount,
			waitingContributions[1].TokenIDStr,
			waitingContributions[1].Amount,
			currentPDEState,
		)
		return
	}
	storePDEPoolForPair(
		pdePoolForPairKey,
		waitingContributions[0].TokenIDStr,
		pdePoolForPair.Token1PoolValue+waitingContributions[0].Amount,
		waitingContributions[1].TokenIDStr,
		pdePoolForPair.Token2PoolValue+waitingContributions[1].Amount,
		currentPDEState,
	)
}

func contributeToPDE(
	beaconHeight uint64,
	pairID string,
	contributorAddressStr string,
	tokenIDStr string,
	contributedAmount uint64,
	currentPDEState *CurrentPDEState,
) {
	waitingContribPairKey := string(lvdb.BuildWaitingPDEContributionKey(beaconHeight, pairID))
	waitingContribution, found := currentPDEState.WaitingPDEContributions[waitingContribPairKey]
	if !found || waitingContribution == nil {
		currentPDEState.WaitingPDEContributions[waitingContribPairKey] = &lvdb.PDEContribution{
			ContributorAddressStr: contributorAddressStr,
			TokenIDStr:            tokenIDStr,
			Amount:                contributedAmount,
		}
		return
	}
	// there was a waiting pde contribution with the same pairID
	if tokenIDStr == waitingContribution.TokenIDStr {
		currentPDEState.WaitingPDEContributions[waitingContribPairKey] = &lvdb.PDEContribution{
			ContributorAddressStr: contributorAddressStr,
			TokenIDStr:            tokenIDStr,
			Amount:                contributedAmount + waitingContribution.Amount,
		}
		return
	}
	// contributing on the remaining token type of existing pair -> move that pair to pde pool for trading
	delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
	updateWaitingContributionPairToPool(
		beaconHeight,
		&lvdb.PDEContribution{
			ContributorAddressStr: contributorAddressStr,
			TokenIDStr:            tokenIDStr,
			Amount:                contributedAmount,
		},
		waitingContribution,
		currentPDEState,
	)
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
		waitingContribPairKey := string(lvdb.BuildWaitingPDEContributionKey(beaconHeight, waitingContribution.PDEContributionPairID))
		currentPDEState.WaitingPDEContributions[waitingContribPairKey] = &lvdb.PDEContribution{
			ContributorAddressStr: waitingContribution.ContributorAddressStr,
			TokenIDStr:            waitingContribution.TokenIDStr,
			Amount:                waitingContribution.ContributedAmount,
		}

	} else if contributionStatus == "refund" {
		var refundContribution metadata.PDERefundContribution
		err := json.Unmarshal([]byte(instruction[3]), &refundContribution)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde refund contribution instruction: %+v", err)
			return nil
		}
		waitingContribPairKey := string(lvdb.BuildWaitingPDEContributionKey(beaconHeight, refundContribution.PDEContributionPairID))
		_, found := currentPDEState.WaitingPDEContributions[waitingContribPairKey]
		if found {
			delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
		}

	} else if contributionStatus == "matched" {
		var matchedContribution metadata.PDEMatchedContribution
		err := json.Unmarshal([]byte(instruction[3]), &matchedContribution)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde matched contribution instruction: %+v", err)
			return nil
		}
		waitingContribPairKey := string(lvdb.BuildWaitingPDEContributionKey(beaconHeight, matchedContribution.PDEContributionPairID))
		existingWaitingContribution, found := currentPDEState.WaitingPDEContributions[waitingContribPairKey]
		if !found || existingWaitingContribution == nil {
			Logger.log.Errorf("ERROR: could not find out existing waiting contribution with unique pair id: %s", matchedContribution.PDEContributionPairID)
			return nil
		}
		incomingWaitingContribution := &lvdb.PDEContribution{
			ContributorAddressStr: matchedContribution.ContributorAddressStr,
			TokenIDStr:            matchedContribution.TokenIDStr,
			Amount:                matchedContribution.ContributedAmount,
		}
		updateWaitingContributionPairToPoolV2(
			beaconHeight,
			existingWaitingContribution,
			incomingWaitingContribution,
			currentPDEState,
		)
		delete(currentPDEState.WaitingPDEContributions, waitingContribPairKey)
	}
	return nil
}

func (blockchain *BlockChain) processPDEContribution(
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
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while decoding content string of pde contribution instruction: %+v", err)
		return nil
	}
	var pdeContributionInst metadata.PDEContributionAction
	err = json.Unmarshal(contentBytes, &pdeContributionInst)
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while unmarshaling pde contribution instruction: %+v", err)
		return nil
	}
	meta := pdeContributionInst.Meta
	pairID := meta.PDEContributionPairID
	contributeToPDE(
		beaconHeight,
		pairID,
		meta.ContributorAddressStr,
		meta.TokenIDStr,
		meta.ContributedAmount,
		currentPDEState,
	)
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
		return nil
	}
	var pdeTradeAcceptedContent metadata.PDETradeAcceptedContent
	err := json.Unmarshal([]byte(instruction[3]), &pdeTradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while unmarshaling PDETradeAcceptedContent: %+v", err)
		return nil
	}
	pdePoolForPairKey := string(lvdb.BuildPDEPoolForPairKey(beaconHeight, pdeTradeAcceptedContent.Token1IDStr, pdeTradeAcceptedContent.Token2IDStr))
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
	pdeShareKey := string(lvdb.BuildPDESharesKeyV2(beaconHeight, token1IDStr, token2IDStr, withdrawerAddressStr))
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
	var wdAcceptedContent metadata.PDEWithdrawalAcceptedContent
	err := json.Unmarshal([]byte(instruction[3]), &wdAcceptedContent)
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while unmarshaling PDEWithdrawalAcceptedContent: %+v", err)
		return nil
	}

	// update pde pool pair
	pdePoolForPairKey := string(lvdb.BuildPDEPoolForPairKey(
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
	return nil
}
