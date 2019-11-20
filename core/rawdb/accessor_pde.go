package rawdb

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/incdb"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func DeleteWaitingPDEContributionByPairID(
	db incdb.Database,
	beaconHeight uint64,
	pairID string,
) error {
	key := BuildWaitingPDEContributionKey(beaconHeight, pairID)
	dbErr := db.Delete(key)
	if dbErr != nil {
		return NewRawdbError(DeleteWaitingPDEContributionError, errors.Wrap(dbErr, "db.lvdb.del"))
	}
	return nil
}

func storeWaitingPDEContribution(
	db incdb.Database,
	beaconHeight uint64,
	pairID string,
	contributorAddressStr string,
	tokenIDStr string,
	amount uint64,
) error {
	waitingPDEContribution := PDEContribution{
		ContributorAddressStr: contributorAddressStr,
		TokenIDStr:            tokenIDStr,
		Amount:                amount,
	}
	waitingPDEContributionBytes, err := json.Marshal(waitingPDEContribution)
	if err != nil {
		return NewRawdbError(StoreWaitingPDEContributionError, errors.Wrap(err, "marshal.to.bytes"))
	}
	key := BuildWaitingPDEContributionKey(beaconHeight, pairID)
	err = db.Put(key, waitingPDEContributionBytes)
	if err != nil {
		return NewRawdbError(StoreWaitingPDEContributionError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func ContributeToPDE(
	db incdb.Database,
	beaconHeight uint64,
	pairID string,
	contributorAddressStr string,
	tokenIDStr string,
	contributedAmount uint64,
) error {
	waitingContributionPairKey := BuildWaitingPDEContributionKey(beaconHeight, pairID)
	waitingContributionBytes, err := db.Get(waitingContributionPairKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return NewRawdbError(GetWaitingPDEContributionByPairIDError, err)
	}
	if len(waitingContributionBytes) == 0 {
		return storeWaitingPDEContribution(db, beaconHeight+1, pairID, contributorAddressStr, tokenIDStr, contributedAmount)
	}

	// there was a waiting pde contribution with the same pairID
	var waitingPDEContribution PDEContribution
	err = json.Unmarshal(waitingContributionBytes, &waitingPDEContribution)
	if err != nil {
		return err
	}
	if tokenIDStr == waitingPDEContribution.TokenIDStr {
		return storeWaitingPDEContribution(db, beaconHeight, pairID, contributorAddressStr, tokenIDStr, contributedAmount+waitingPDEContribution.Amount)
	}

	// contributing on the remaining token type of existing pair -> move that pair to pde pool for trading
	err = DeleteWaitingPDEContributionByPairID(db, beaconHeight, pairID)
	if err != nil {
		return err
	}
	return updateWaitingContributionPairToPool(
		db,
		beaconHeight,
		PDEContribution{
			ContributorAddressStr: contributorAddressStr,
			TokenIDStr:            tokenIDStr,
			Amount:                contributedAmount,
		},
		waitingPDEContribution,
	)
}

func storePDEPoolForPair(
	db incdb.Database,
	pdePoolForPairKey []byte,
	token1IDStr string,
	token1PoolValue uint64,
	token2IDStr string,
	token2PoolValue uint64,
) error {
	pdePoolForPair := PDEPoolForPair{
		Token1IDStr:     token1IDStr,
		Token1PoolValue: token1PoolValue,
		Token2IDStr:     token2IDStr,
		Token2PoolValue: token2PoolValue,
	}
	pdePoolForPairBytes, err := json.Marshal(pdePoolForPair)
	if err != nil {
		return NewRawdbError(StorePDEPoolForPairError, errors.Wrap(err, "marshal.to.bytes"))
	}
	err = db.Put(pdePoolForPairKey, pdePoolForPairBytes)
	if err != nil {
		return NewRawdbError(StorePDEPoolForPairError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func addShareAmountUp(
	db incdb.Database,
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddrStr string,
	amt uint64,
) error {
	pdeShareKey := BuildPDESharesKey(beaconHeight, token1IDStr, token2IDStr, contributedTokenIDStr, contributorAddrStr)
	pdeShareBytes, err := db.Get(pdeShareKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return NewRawdbError(GetPDEShareError, err)
	}
	addedUpAmt := amt
	if len(pdeShareBytes) > 0 {
		currentAmt := uint64(binary.LittleEndian.Uint64(pdeShareBytes))
		addedUpAmt += currentAmt
	}
	buf := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(buf, addedUpAmt)
	dbErr := db.Put(pdeShareKey, buf)
	if dbErr != nil {
		return NewRawdbError(AddShareAmountUpError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func GetSharesOfContributorForTokenIDOnAPair(
	db incdb.Database,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddrStr string,
) (uint64, error) {
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	pdeShareKey := append(PDESharePrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+contributedTokenIDStr+"-"+contributorAddrStr)...)
	pdeShareBytes, err := db.Get(pdeShareKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return 0, NewRawdbError(GetPDEShareError, err)
	}
	if len(pdeShareBytes) > 0 {
		return uint64(binary.LittleEndian.Uint64(pdeShareBytes)), nil
	}
	return 0, nil
}

func GetTotalSharesForTokenIDOnAPair(
	db incdb.Database,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
) (uint64, error) {
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	pdeShareForTokenIDPrefix := append(PDESharePrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+contributedTokenIDStr)...)

	totalShares := uint64(0)
	iter := db.NewIteratorWithPrefix(pdeShareForTokenIDPrefix)
	for iter.Next() {
		value := iter.Value()
		itemBytes := make([]byte, len(value))
		copy(itemBytes, value)
		totalShares += uint64(binary.LittleEndian.Uint64(itemBytes))
	}
	iter.Release()
	err := iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return 0, NewRawdbError(GetPDEShareError, err)
	}
	return totalShares, nil
}

func updateWaitingContributionPairToPool(
	db incdb.Database,
	beaconHeight uint64,
	waitingContribution1 PDEContribution,
	waitingContribution2 PDEContribution,
) error {
	err := addShareAmountUp(
		db,
		beaconHeight,
		waitingContribution1.TokenIDStr,
		waitingContribution2.TokenIDStr,
		waitingContribution1.TokenIDStr,
		waitingContribution1.ContributorAddressStr,
		waitingContribution1.Amount,
	)
	if err != nil {
		return err
	}
	err = addShareAmountUp(
		db,
		beaconHeight,
		waitingContribution1.TokenIDStr,
		waitingContribution2.TokenIDStr,
		waitingContribution2.TokenIDStr,
		waitingContribution2.ContributorAddressStr,
		waitingContribution2.Amount,
	)
	if err != nil {
		return err
	}

	waitingContributions := []PDEContribution{waitingContribution1, waitingContribution2}
	sort.Slice(waitingContributions, func(i, j int) bool {
		return waitingContributions[i].TokenIDStr < waitingContributions[j].TokenIDStr
	})
	pdePoolForPairKey := BuildPDEPoolForPairKey(beaconHeight, waitingContributions[0].TokenIDStr, waitingContributions[1].TokenIDStr)
	pdePoolForPairBytes, err := db.Get(pdePoolForPairKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return NewRawdbError(GetPDEPoolForPairKeyError, err)
	}
	if len(pdePoolForPairBytes) == 0 {
		return storePDEPoolForPair(
			db,
			pdePoolForPairKey,
			waitingContributions[0].TokenIDStr,
			waitingContributions[0].Amount,
			waitingContributions[1].TokenIDStr,
			waitingContributions[1].Amount,
		)
	}
	var pdePoolForPair PDEPoolForPair
	err = json.Unmarshal(pdePoolForPairBytes, &pdePoolForPair)
	if err != nil {
		return err
	}
	return storePDEPoolForPair(
		db,
		pdePoolForPairKey,
		waitingContributions[0].TokenIDStr,
		pdePoolForPair.Token1PoolValue+waitingContributions[0].Amount,
		waitingContributions[1].TokenIDStr,
		pdePoolForPair.Token2PoolValue+waitingContributions[1].Amount,
	)
}

func GetPDEPoolForPair(
	db incdb.Database,
	beaconHeight uint64,
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
) ([]byte, error) {
	pdePoolForPairKey := BuildPDEPoolForPairKey(beaconHeight, tokenIDToBuyStr, tokenIDToSellStr)
	pdePoolForPairBytes, err := db.Get(pdePoolForPairKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return []byte{}, NewRawdbError(GetPDEPoolForPairKeyError, err)
	}
	return pdePoolForPairBytes, nil
}

func GetLatestPDEPoolForPair(
	db incdb.Database,
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
) ([]byte, error) {
	iter := db.NewIteratorWithPrefix(PDEPoolPrefix)
	ok := iter.Last()
	if !ok {
		return []byte{}, nil
	}
	key := iter.Key()
	keyBytes := make([]byte, len(key))
	copy(keyBytes, key)
	iter.Release()
	err := iter.Error()
	if err != nil {
		return []byte{}, err
	}

	parts := strings.Split(string(keyBytes), "-")
	if len(parts) <= 1 {
		return []byte{}, nil
	}
	beaconHeight, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return []byte{}, err
	}

	pdePoolForPairKey := BuildPDEPoolForPairKey(beaconHeight, tokenIDToBuyStr, tokenIDToSellStr)
	pdePoolForPairBytes, err := db.Get(pdePoolForPairKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return []byte{}, NewRawdbError(GetPDEPoolForPairKeyError, err)
	}
	return pdePoolForPairBytes, nil
}

func UpdatePDEPoolForPair(
	db incdb.Database,
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	pdePoolForPairBytes []byte,
) error {
	pdePoolForPairKey := BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr)
	dbErr := db.Put(pdePoolForPairKey, pdePoolForPairBytes)
	if dbErr != nil {
		return NewRawdbError(StoreWaitingPDEContributionError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func AddTradeFeeUp(
	db incdb.Database,
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	targetingTokenIDStr string,
	amt uint64,
) error {
	pdeTradeFeeKey := BuildPDETradeFeesKey(beaconHeight, token1IDStr, token2IDStr, targetingTokenIDStr)
	pdeTradeFeeBytes, err := db.Get(pdeTradeFeeKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return NewRawdbError(GetPDETradeFeeError, err)
	}
	addedAmt := amt
	if len(pdeTradeFeeBytes) > 0 {
		currentAmt := uint64(binary.LittleEndian.Uint64(pdeTradeFeeBytes))
		addedAmt = currentAmt + amt
	}

	buf := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(buf, addedAmt)
	dbErr := db.Put(pdeTradeFeeKey, buf)
	if dbErr != nil {
		return NewRawdbError(AddTradeFeeUpError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func DeductTradeFee(
	db incdb.Database,
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	targetingTokenIDStr string,
	amt uint64,
) error {
	pdeTradeFeeKey := BuildPDETradeFeesKey(beaconHeight, token1IDStr, token2IDStr, targetingTokenIDStr)
	pdeTradeFeeBytes, err := db.Get(pdeTradeFeeKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return NewRawdbError(GetPDETradeFeeError, err)
	}
	adjustingAmt := uint64(0)
	if len(pdeTradeFeeBytes) > 0 {
		currentAmt := uint64(binary.LittleEndian.Uint64(pdeTradeFeeBytes))
		if amt <= currentAmt {
			adjustingAmt = currentAmt - amt
		}
	}
	buf := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(buf, adjustingAmt)
	dbErr := db.Put(pdeTradeFeeKey, buf)
	if dbErr != nil {
		return NewRawdbError(DeduceTradeFeeError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func DeductSharesForWithdrawal(
	db incdb.Database,
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	targetingTokenIDStr string,
	withdrawerAddressStr string,
	amt uint64,
) error {
	pdeShareKey := BuildPDESharesKey(beaconHeight, token1IDStr, token2IDStr, targetingTokenIDStr, withdrawerAddressStr)
	pdeShareBytes, err := db.Get(pdeShareKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return NewRawdbError(GetPDEShareError, err)
	}
	adjustingAmt := uint64(0)
	if len(pdeShareBytes) > 0 {
		currentAmt := uint64(binary.LittleEndian.Uint64(pdeShareBytes))
		if amt <= currentAmt {
			adjustingAmt = currentAmt - amt
		}
	}
	buf := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(buf, adjustingAmt)
	dbErr := db.Put(pdeShareKey, buf)
	if dbErr != nil {
		return NewRawdbError(DeduceShareError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func GetAllRecordsByPrefix(db incdb.Database, beaconHeight uint64, prefix []byte) ([][]byte, [][]byte, error) {
	keys := [][]byte{}
	values := [][]byte{}
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	prefixByBeaconHeight := append(prefix, beaconHeightBytes...)
	iter := db.NewIteratorWithPrefix(prefixByBeaconHeight)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		keyBytes := make([]byte, len(key))
		valueBytes := make([]byte, len(value))
		copy(keyBytes, key)
		copy(valueBytes, value)
		keys = append(keys, keyBytes)
		values = append(values, valueBytes)
	}
	iter.Release()
	err := iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return keys, values, NewRawdbError(GetAllRecordsByPrefixError, err)
	}
	return keys, values, nil
}

func TrackPDEStatus(
	db incdb.Database,
	prefix []byte,
	suffix []byte,
	status byte,
) error {
	key := BuildPDEStatusKey(prefix, suffix)
	err := db.Put(key, []byte{status})
	if err != nil {
		return NewRawdbError(TrackPDEStatusError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func GetPDEStatus(
	db incdb.Database,
	prefix []byte,
	suffix []byte,
) (byte, error) {
	key := BuildPDEStatusKey(prefix, suffix)
	pdeStatusBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return common.PDENotFoundStatus, NewRawdbError(GetPDEStatusError, dbErr)
	}
	if len(pdeStatusBytes) == 0 {
		return common.PDENotFoundStatus, nil
	}
	return pdeStatusBytes[0], nil
}
