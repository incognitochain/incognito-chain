package lvdb

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type PDEContribution struct {
	ContributorAddressStr string
	TokenIDStr            string
	Amount                uint64
	TxReqID               common.Hash
}

type PDEPoolForPair struct {
	Token1IDStr     string
	Token1PoolValue uint64
	Token2IDStr     string
	Token2PoolValue uint64
}

func BuildPDEStatusKey(
	prefix []byte,
	suffix []byte,
) []byte {
	return append(prefix, suffix...)
}

func BuildPDESharesKey(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddressStr string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	pdeSharesByBCHeightPrefix := append(PDESharePrefix, beaconHeightBytes...)
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	return append(pdeSharesByBCHeightPrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+contributedTokenIDStr+"-"+contributorAddressStr)...)
}

func BuildPDESharesKeyV2(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributorAddressStr string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	pdeSharesByBCHeightPrefix := append(PDESharePrefix, beaconHeightBytes...)
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	return append(pdeSharesByBCHeightPrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+contributorAddressStr)...)
}

func BuildPDEPoolForPairKey(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	pdePoolForPairByBCHeightPrefix := append(PDEPoolPrefix, beaconHeightBytes...)
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	return append(pdePoolForPairByBCHeightPrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1])...)
}

func BuildPDETradeFeesKey(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	tokenForFeeIDStr string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	pdeTradeFeesByBCHeightPrefix := append(PDETradeFeePrefix, beaconHeightBytes...)
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	return append(pdeTradeFeesByBCHeightPrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+tokenForFeeIDStr)...)
}

func BuildWaitingPDEContributionKey(
	beaconHeight uint64,
	pairID string,
) []byte {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	waitingPDEContribByBCHeightPrefix := append(WaitingPDEContributionPrefix, beaconHeightBytes...)
	return append(waitingPDEContribByBCHeightPrefix, []byte(pairID)...)
}

func (db *db) DeleteWaitingPDEContributionByPairID(
	beaconHeight uint64,
	pairID string,
) error {
	key := BuildWaitingPDEContributionKey(beaconHeight, pairID)
	dbErr := db.Delete(key)
	if dbErr != nil {
		return database.NewDatabaseError(database.DeleteWaitingPDEContributionError, errors.Wrap(dbErr, "db.lvdb.del"))
	}
	return nil
}

func (db *db) storeWaitingPDEContribution(
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
		return database.NewDatabaseError(database.StoreWaitingPDEContributionError, errors.Wrap(err, "marshal.to.bytes"))
	}
	key := BuildWaitingPDEContributionKey(beaconHeight, pairID)
	err = db.Put(key, waitingPDEContributionBytes)
	if err != nil {
		return database.NewDatabaseError(database.StoreWaitingPDEContributionError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) ContributeToPDE(
	beaconHeight uint64,
	pairID string,
	contributorAddressStr string,
	tokenIDStr string,
	contributedAmount uint64,
) error {
	waitingContributionPairKey := BuildWaitingPDEContributionKey(beaconHeight, pairID)
	waitingContributionBytes, err := db.lvdb.Get(waitingContributionPairKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.GetWaitingPDEContributionByPairIDError, err)
	}
	if len(waitingContributionBytes) == 0 {
		return db.storeWaitingPDEContribution(beaconHeight+1, pairID, contributorAddressStr, tokenIDStr, contributedAmount)
	}

	// there was a waiting pde contribution with the same pairID
	var waitingPDEContribution PDEContribution
	err = json.Unmarshal(waitingContributionBytes, &waitingPDEContribution)
	if err != nil {
		return err
	}
	if tokenIDStr == waitingPDEContribution.TokenIDStr {
		return db.storeWaitingPDEContribution(beaconHeight, pairID, contributorAddressStr, tokenIDStr, contributedAmount+waitingPDEContribution.Amount)
	}

	// contributing on the remaining token type of existing pair -> move that pair to pde pool for trading
	err = db.DeleteWaitingPDEContributionByPairID(beaconHeight, pairID)
	if err != nil {
		return err
	}
	return db.updateWaitingContributionPairToPool(
		beaconHeight,
		PDEContribution{
			ContributorAddressStr: contributorAddressStr,
			TokenIDStr:            tokenIDStr,
			Amount:                contributedAmount,
		},
		waitingPDEContribution,
	)
}

func (db *db) storePDEPoolForPair(
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
		return database.NewDatabaseError(database.StorePDEPoolForPairError, errors.Wrap(err, "marshal.to.bytes"))
	}
	err = db.Put(pdePoolForPairKey, pdePoolForPairBytes)
	if err != nil {
		return database.NewDatabaseError(database.StorePDEPoolForPairError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) addShareAmountUp(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddrStr string,
	amt uint64,
) error {
	pdeShareKey := BuildPDESharesKey(beaconHeight, token1IDStr, token2IDStr, contributedTokenIDStr, contributorAddrStr)
	pdeShareBytes, err := db.lvdb.Get(pdeShareKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.GetPDEShareError, err)
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
		return database.NewDatabaseError(database.AddShareAmountUpError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetSharesOfContributorForTokenIDOnAPair(
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
	contributorAddrStr string,
) (uint64, error) {
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	pdeShareKey := append(PDESharePrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+contributedTokenIDStr+"-"+contributorAddrStr)...)
	pdeShareBytes, err := db.lvdb.Get(pdeShareKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return 0, database.NewDatabaseError(database.GetPDEShareError, err)
	}
	if len(pdeShareBytes) > 0 {
		return uint64(binary.LittleEndian.Uint64(pdeShareBytes)), nil
	}
	return 0, nil
}

func (db *db) GetTotalSharesForTokenIDOnAPair(
	token1IDStr string,
	token2IDStr string,
	contributedTokenIDStr string,
) (uint64, error) {
	tokenIDStrs := []string{token1IDStr, token2IDStr}
	sort.Strings(tokenIDStrs)
	pdeShareForTokenIDPrefix := append(PDESharePrefix, []byte(tokenIDStrs[0]+"-"+tokenIDStrs[1]+"-"+contributedTokenIDStr)...)

	totalShares := uint64(0)
	iter := db.lvdb.NewIterator(util.BytesPrefix(pdeShareForTokenIDPrefix), nil)
	for iter.Next() {
		value := iter.Value()
		itemBytes := make([]byte, len(value))
		copy(itemBytes, value)
		totalShares += uint64(binary.LittleEndian.Uint64(itemBytes))
	}
	iter.Release()
	err := iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return 0, database.NewDatabaseError(database.GetPDEShareError, err)
	}
	return totalShares, nil
}

func (db *db) updateWaitingContributionPairToPool(
	beaconHeight uint64,
	waitingContribution1 PDEContribution,
	waitingContribution2 PDEContribution,
) error {
	err := db.addShareAmountUp(
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
	err = db.addShareAmountUp(
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
	pdePoolForPairBytes, err := db.lvdb.Get(pdePoolForPairKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.GetPDEPoolForPairKeyError, err)
	}
	if len(pdePoolForPairBytes) == 0 {
		return db.storePDEPoolForPair(
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
	return db.storePDEPoolForPair(
		pdePoolForPairKey,
		waitingContributions[0].TokenIDStr,
		pdePoolForPair.Token1PoolValue+waitingContributions[0].Amount,
		waitingContributions[1].TokenIDStr,
		pdePoolForPair.Token2PoolValue+waitingContributions[1].Amount,
	)
}

func (db *db) GetPDEPoolForPair(
	beaconHeight uint64,
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
) ([]byte, error) {
	pdePoolForPairKey := BuildPDEPoolForPairKey(beaconHeight, tokenIDToBuyStr, tokenIDToSellStr)
	pdePoolForPairBytes, err := db.lvdb.Get(pdePoolForPairKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return []byte{}, database.NewDatabaseError(database.GetPDEPoolForPairKeyError, err)
	}
	return pdePoolForPairBytes, nil
}

func (db *db) GetLatestPDEPoolForPair(
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
) ([]byte, error) {
	iter := db.lvdb.NewIterator(util.BytesPrefix(PDEPoolPrefix), nil)
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
	pdePoolForPairBytes, err := db.lvdb.Get(pdePoolForPairKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return []byte{}, database.NewDatabaseError(database.GetPDEPoolForPairKeyError, err)
	}
	return pdePoolForPairBytes, nil
}

func (db *db) UpdatePDEPoolForPair(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	pdePoolForPairBytes []byte,
) error {
	pdePoolForPairKey := BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr)
	dbErr := db.Put(pdePoolForPairKey, pdePoolForPairBytes)
	if dbErr != nil {
		return database.NewDatabaseError(database.StoreWaitingPDEContributionError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func (db *db) AddTradeFeeUp(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	targetingTokenIDStr string,
	amt uint64,
) error {
	pdeTradeFeeKey := BuildPDETradeFeesKey(beaconHeight, token1IDStr, token2IDStr, targetingTokenIDStr)
	pdeTradeFeeBytes, err := db.lvdb.Get(pdeTradeFeeKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.GetPDETradeFeeError, err)
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
		return database.NewDatabaseError(database.AddTradeFeeUpError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func (db *db) DeductTradeFee(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	targetingTokenIDStr string,
	amt uint64,
) error {
	pdeTradeFeeKey := BuildPDETradeFeesKey(beaconHeight, token1IDStr, token2IDStr, targetingTokenIDStr)
	pdeTradeFeeBytes, err := db.lvdb.Get(pdeTradeFeeKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.GetPDETradeFeeError, err)
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
		return database.NewDatabaseError(database.DeduceTradeFeeError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func (db *db) DeductSharesForWithdrawal(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	targetingTokenIDStr string,
	withdrawerAddressStr string,
	amt uint64,
) error {
	pdeShareKey := BuildPDESharesKey(beaconHeight, token1IDStr, token2IDStr, targetingTokenIDStr, withdrawerAddressStr)
	pdeShareBytes, err := db.lvdb.Get(pdeShareKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.GetPDEShareError, err)
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
		return database.NewDatabaseError(database.DeduceShareError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetAllRecordsByPrefix(beaconHeight uint64, prefix []byte) ([][]byte, [][]byte, error) {
	keys := [][]byte{}
	values := [][]byte{}
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	prefixByBeaconHeight := append(prefix, beaconHeightBytes...)
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefixByBeaconHeight), nil)
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
		return keys, values, database.NewDatabaseError(database.GetAllRecordsByPrefixError, err)
	}
	return keys, values, nil
}

func (db *db) TrackPDEStatus(
	prefix []byte,
	suffix []byte,
	status byte,
) error {
	key := BuildPDEStatusKey(prefix, suffix)
	err := db.Put(key, []byte{status})
	if err != nil {
		return database.NewDatabaseError(database.TrackPDEStatusError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetPDEStatus(
	prefix []byte,
	suffix []byte,
) (byte, error) {
	key := BuildPDEStatusKey(prefix, suffix)
	pdeStatusBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return common.PDENotFoundStatus, database.NewDatabaseError(database.GetPDEStatusError, dbErr)
	}
	if len(pdeStatusBytes) == 0 {
		return common.PDENotFoundStatus, nil
	}
	return pdeStatusBytes[0], nil
}

func (db *db) TrackPDEContributionStatus(
	prefix []byte,
	suffix []byte,
	statusContent []byte,
) error {
	key := BuildPDEStatusKey(prefix, suffix)
	err := db.Put(key, statusContent)
	if err != nil {
		return database.NewDatabaseError(database.TrackPDEStatusError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetPDEContributionStatus(
	prefix []byte,
	suffix []byte,
) ([]byte, error) {
	key := BuildPDEStatusKey(prefix, suffix)
	pdeStatusContentBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return []byte{}, database.NewDatabaseError(database.GetPDEStatusError, dbErr)
	}
	return pdeStatusContentBytes, nil
}
