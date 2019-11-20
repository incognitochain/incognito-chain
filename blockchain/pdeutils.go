package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/incognitochain/incognito-chain/core/rawdb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/pkg/errors"
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

func storePDEShares(
	db incdb.Database,
	beaconHeight uint64,
	pdeShares map[string]uint64,
) error {
	for shareKey, shareAmt := range pdeShares {
		newKey := replaceNewBCHeightInKeyStr(shareKey, beaconHeight)
		buf := make([]byte, binary.MaxVarintLen64)
		binary.LittleEndian.PutUint64(buf, shareAmt)
		dbErr := db.Put([]byte(newKey), buf)
		if dbErr != nil {
			return incdb.NewDatabaseError(incdb.AddShareAmountUpError, errors.Wrap(dbErr, "db.lvdb.put"))
		}
	}
	return nil
}

func storeWaitingPDEContributions(
	db incdb.Database,
	beaconHeight uint64,
	waitingPDEContributions map[string]*rawdb.PDEContribution,
) error {
	for contribKey, contribution := range waitingPDEContributions {
		newKey := replaceNewBCHeightInKeyStr(contribKey, beaconHeight)
		contributionBytes, err := json.Marshal(contribution)
		if err != nil {
			return err
		}
		err = db.Put([]byte(newKey), contributionBytes)
		if err != nil {
			return incdb.NewDatabaseError(incdb.StoreWaitingPDEContributionError, errors.Wrap(err, "db.lvdb.put"))
		}
	}
	return nil
}

func storePDEPoolPairs(
	db incdb.Database,
	beaconHeight uint64,
	pdePoolPairs map[string]*rawdb.PDEPoolForPair,
) error {
	for poolPairKey, poolPair := range pdePoolPairs {
		newKey := replaceNewBCHeightInKeyStr(poolPairKey, beaconHeight)
		poolPairBytes, err := json.Marshal(poolPair)
		if err != nil {
			return err
		}
		err = db.Put([]byte(newKey), poolPairBytes)
		if err != nil {
			return incdb.NewDatabaseError(incdb.StorePDEPoolForPairError, errors.Wrap(err, "db.lvdb.put"))
		}
	}
	return nil
}

func getWaitingPDEContributions(
	db incdb.Database,
	beaconHeight uint64,
) (map[string]*rawdb.PDEContribution, error) {
	waitingPDEContributions := make(map[string]*rawdb.PDEContribution)
	waitingContribKeysBytes, waitingContribValuesBytes, err := rawdb.GetAllRecordsByPrefix(db, beaconHeight, rawdb.WaitingPDEContributionPrefix)
	if err != nil {
		return nil, err
	}
	for idx, waitingContribKeyBytes := range waitingContribKeysBytes {
		var waitingContrib rawdb.PDEContribution
		err = json.Unmarshal(waitingContribValuesBytes[idx], &waitingContrib)
		if err != nil {
			return nil, err
		}
		waitingPDEContributions[string(waitingContribKeyBytes)] = &waitingContrib
	}
	return waitingPDEContributions, nil
}

func getPDEPoolPair(
	db incdb.Database,
	beaconHeight uint64,
) (map[string]*rawdb.PDEPoolForPair, error) {
	pdePoolPairs := make(map[string]*rawdb.PDEPoolForPair)
	poolPairsKeysBytes, poolPairsValuesBytes, err := rawdb.GetAllRecordsByPrefix(db, beaconHeight, rawdb.PDEPoolPrefix)
	if err != nil {
		return nil, err
	}
	for idx, poolPairsKeyBytes := range poolPairsKeysBytes {
		var padePoolPair rawdb.PDEPoolForPair
		err = json.Unmarshal(poolPairsValuesBytes[idx], &padePoolPair)
		if err != nil {
			return nil, err
		}
		pdePoolPairs[string(poolPairsKeyBytes)] = &padePoolPair
	}
	return pdePoolPairs, nil
}

func getPDEShares(
	db incdb.Database,
	beaconHeight uint64,
) (map[string]uint64, error) {
	pdeShares := make(map[string]uint64)
	sharesKeysBytes, sharesValuesBytes, err := rawdb.GetAllRecordsByPrefix(db, beaconHeight, rawdb.PDESharePrefix)
	if err != nil {
		return nil, err
	}
	for idx, sharesKeyBytes := range sharesKeysBytes {
		shareAmt := uint64(binary.LittleEndian.Uint64(sharesValuesBytes[idx]))
		pdeShares[string(sharesKeyBytes)] = shareAmt
	}
	return pdeShares, nil
}

func InitCurrentPDEStateFromDB(
	db incdb.Database,
	beaconHeight uint64,
) (*CurrentPDEState, error) {
	waitingPDEContributions, err := getWaitingPDEContributions(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdePoolPairs, err := getPDEPoolPair(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	pdeShares, err := getPDEShares(db, beaconHeight)
	if err != nil {
		return nil, err
	}
	return &CurrentPDEState{
		WaitingPDEContributions: waitingPDEContributions,
		PDEPoolPairs:            pdePoolPairs,
		PDEShares:               pdeShares,
	}, nil
}

func storePDEStateToDB(
	db incdb.Database,
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
) error {
	err := storeWaitingPDEContributions(db, beaconHeight, currentPDEState.WaitingPDEContributions)
	if err != nil {
		return err
	}
	err = storePDEPoolPairs(db, beaconHeight, currentPDEState.PDEPoolPairs)
	if err != nil {
		return err
	}
	err = storePDEShares(db, beaconHeight, currentPDEState.PDEShares)
	if err != nil {
		return err
	}
	return nil
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
