package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"math/big"
	"sort"
	"strings"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
)

type CurrentPDEState struct {
	WaitingPDEContributions map[string]*rawdb.PDEContribution
	PDEPoolPairs            map[string]*rawdb.PDEPoolForPair
	PDEShares               map[string]uint64
}

type DeductingAmountsByWithdrawal struct {
	Token1IDStr string
	PoolValue1  uint64
	Token2IDStr string
	PoolValue2  uint64
	Shares      uint64
}

func InitCurrentPDEStateFromDB(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
) (*CurrentPDEState, error) {
	waitingPDEContributions, err := statedb.GetWaitingPDEContributions(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdePoolPairs, err := statedb.GetPDEPoolPair(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdeShares, err := statedb.GetPDEShares(stateDB, beaconHeight)
	if err != nil {
		return nil, err
	}
	return &CurrentPDEState{
		WaitingPDEContributions: waitingPDEContributions,
		PDEPoolPairs:            pdePoolPairs,
		PDEShares:               pdeShares,
	}, nil
}

func storePDEStateToDBV2(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
) error {
	err := statedb.StoreWaitingPDEContributions(stateDB, beaconHeight, currentPDEState.WaitingPDEContributions)
	if err != nil {
		return err
	}
	err = statedb.StorePDEPoolPairs(stateDB, beaconHeight, currentPDEState.PDEPoolPairs)
	if err != nil {
		return err
	}
	err = statedb.StorePDEShares(stateDB, beaconHeight, currentPDEState.PDEShares)
	if err != nil {
		return err
	}
	return nil
}

func replaceNewBCHeightInKeyStr(key string, newBeaconHeight uint64) string {
	parts := strings.Split(key, "-")
	if len(parts) <= 1 {
		return key
	}
	parts[1] = fmt.Sprintf("%d", newBeaconHeight)
	newKey := ""
	for idx, part := range parts {
		if idx == len(parts)-1 {
			newKey += part
			continue
		}
		newKey += (part + "-")
	}
	return newKey
}

func addShareAmountUpV2(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddrStr string,
	amt uint64,
	currentPDEState *CurrentPDEState,
) {
	pdeShareOnTokenPrefix := string(rawdb.BuildPDESharesKeyV2(beaconHeight, token1IDStr, token2IDStr, ""))
	totalSharesOnToken := uint64(0)
	for key, value := range currentPDEState.PDEShares {
		if strings.Contains(key, pdeShareOnTokenPrefix) {
			totalSharesOnToken += value
		}
	}
	pdeShareKey := string(rawdb.BuildPDESharesKeyV2(beaconHeight, token1IDStr, token2IDStr, contributorAddrStr))
	if totalSharesOnToken == 0 {
		currentPDEState.PDEShares[pdeShareKey] = amt
		return
	}
	poolPairKey := string(rawdb.BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr))
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

func updateWaitingContributionPairToPoolV2(
	beaconHeight uint64,
	waitingContribution1 *rawdb.PDEContribution,
	waitingContribution2 *rawdb.PDEContribution,
	currentPDEState *CurrentPDEState,
) {
	addShareAmountUpV2(
		beaconHeight,
		waitingContribution1.TokenIDStr,
		waitingContribution2.TokenIDStr,
		waitingContribution1.TokenIDStr,
		waitingContribution1.ContributorAddressStr,
		waitingContribution1.Amount,
		currentPDEState,
	)

	waitingContributions := []*rawdb.PDEContribution{waitingContribution1, waitingContribution2}
	sort.Slice(waitingContributions, func(i, j int) bool {
		return waitingContributions[i].TokenIDStr < waitingContributions[j].TokenIDStr
	})
	pdePoolForPairKey := string(rawdb.BuildPDEPoolForPairKey(beaconHeight, waitingContributions[0].TokenIDStr, waitingContributions[1].TokenIDStr))
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
