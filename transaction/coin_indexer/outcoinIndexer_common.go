package coinIndexer

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"github.com/incognitochain/incognito-chain/wallet"
	"sync"
)

const (
	NumWorkers         = 100
	DefaultAccessToken = "0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11" //nolint:gosec // DEV access token
	BatchWaitingTime   = float64(30)
	IndexingBatchSize  = 10
)

type JobStatus struct {
	id     string // id of the indexing go-routine
	otaKey privacy.OTAKey
	err    error
}

// IndexerInitialConfig is the initial config provided when starting the CoinIndexer.
type IndexerInitialConfig struct {
	TxDbs      []*statedb.StateDB
	BestBlocks []uint64
}

type IndexParam struct {
	FromHeight uint64
	ToHeight   uint64
	OTAKey     privacy.OTAKey
	TxDb       *statedb.StateDB
	ShardID    byte
	IsReset    bool
}

// CoinMatcher is an interface for matching a v2 coin and given parameter(s).
type CoinMatcher func(*privacy.CoinV2, map[string]interface{}) bool

// getCoinFilterByOTAKey returns a functions that filters if an output coin belongs to an OTAKey.
func getCoinFilterByOTAKey() CoinMatcher {
	return func(c *privacy.CoinV2, kvargs map[string]interface{}) bool {
		entry, exists := kvargs["otaKey"]
		if !exists {
			return false
		}
		vk, ok := entry.(privacy.OTAKey)
		if !ok {
			return false
		}
		ks := &incognitokey.KeySet{}
		ks.OTAKey = vk

		pass, _ := c.DoesCoinBelongToKeySet(ks)
		return pass
	}
}

// GetCoinFilterByOTAKeyAndToken returns a functions that filters if an output coin is of a specific token and belongs to an OTAKey.
// If the given token is the generic `common.ConfidentialAssetID` and the retrieved coin has an asset tag field,
// it immediately returns true after the belonging check.
func GetCoinFilterByOTAKeyAndToken() CoinMatcher {
	return func(c *privacy.CoinV2, kvargs map[string]interface{}) bool {
		entry, exists := kvargs["otaKey"]
		if !exists {
			return false
		}
		vk, ok := entry.(privacy.OTAKey)
		if !ok {
			return false
		}
		entry, exists = kvargs["tokenID"]
		if !exists {
			return false
		}
		tokenID, ok := entry.(*common.Hash)
		if !ok {
			return false
		}
		ks := &incognitokey.KeySet{}
		ks.OTAKey = vk

		if pass, sharedSecret := c.DoesCoinBelongToKeySet(ks); pass {
			if tokenID.String() == common.ConfidentialAssetID.String() && c.GetAssetTag() != nil {
				return true
			}
			pass, _ = c.ValidateAssetTag(sharedSecret, tokenID)
			return pass
		}
		return false
	}
}

// GetNextLowerHeight returns the next upper height in querying v2 coins given the current upper height and the base height.
func GetNextLowerHeight(upper, floor uint64) uint64 {
	if upper > utils.MaxOutcoinQueryInterval+floor {
		return upper - utils.MaxOutcoinQueryInterval
	}
	return floor
}

func OTAKeyToRaw(vk privacy.OTAKey) [64]byte {
	var result [64]byte
	copy(result[0:32], vk.GetOTASecretKey().ToBytesS())
	copy(result[32:64], vk.GetPublicSpend().ToBytesS())
	return result
}

func OTAKeyFromRaw(b [64]byte) privacy.OTAKey {
	result := &privacy.OTAKey{}
	result.SetOTASecretKey(b[0:32])
	result.SetPublicSpend(b[32:64])
	return *result
}

// QueryDbCoinVer1 returns all v1 output coins of a public key on the given tokenID.
func QueryDbCoinVer1(pubKey []byte, tokenID *common.Hash, db *statedb.StateDB) ([]privacy.Coin, error) {
	shardID := common.GetShardIDFromLastByte(pubKey[len(pubKey)-1])
	outCoinsBytes, err := statedb.GetOutcoinsByPubkey(db, *tokenID, pubKey, shardID)
	if err != nil {
		utils.Logger.Log.Errorf("get outCoins by pubKey error: %v", err)
		return nil, err
	}
	var outCoins []privacy.Coin
	for _, item := range outCoinsBytes {
		outCoin := &privacy.CoinV1{}
		err = outCoin.SetBytes(item)
		if err != nil {
			utils.Logger.Log.Errorf("cannot create coin from byte %v", err)
			return nil, err
		}
		outCoins = append(outCoins, outCoin)
	}
	return outCoins, nil
}

// QueryDbCoinVer2 returns all v2 output coins of a public key on the given tokenID for heights from `startHeight` to
// `destHeight` using the given list of filters.
//nolint // TODO: consider using get coin by index to speed up the performance.
func QueryDbCoinVer2(otaKey privacy.OTAKey, tokenID *common.Hash, startHeight, destHeight uint64, db *statedb.StateDB, filters ...CoinMatcher) ([]privacy.Coin, error) {
	pubKeyBytes := otaKey.GetPublicSpend().ToBytesS()
	shardID := common.GetShardIDFromLastByte(pubKeyBytes[len(pubKeyBytes)-1])

	burningPubKey := wallet.GetBurningPublicKey()

	var outCoins []privacy.Coin
	// avoid overlap; unless lower height is 0
	start := startHeight + 1
	if startHeight == 0 {
		start = 0
	}
	for height := start; height <= destHeight; height++ {
		currentHeightCoins, err := statedb.GetOTACoinsByHeight(db, *tokenID, shardID, height)
		if err != nil {
			utils.Logger.Log.Errorf("get outCoins ver 2 bytes by height error: %v", err)
			return nil, err
		}

		// create parameter(s) for CoinMatcher filters.
		params := make(map[string]interface{})
		params["otaKey"] = otaKey
		params["db"] = db
		params["tokenID"] = tokenID
		for _, coinBytes := range currentHeightCoins {
			cv2 := &privacy.CoinV2{}
			err = cv2.SetBytes(coinBytes)
			if err != nil {
				utils.Logger.Log.Error("get outCoins ver 2 from bytes error: %v", err)
				return nil, err
			}

			// check if the output coin was sent to the burning address
			if bytes.Equal(cv2.GetPublicKey().ToBytesS(), burningPubKey) {
				continue
			}

			pass := false
			for _, f := range filters {
				if f(cv2, params) {
					pass = true
					break
				}
			}
			if pass {
				outCoins = append(outCoins, cv2)
			}
		}
	}
	return outCoins, nil
}

// QueryBatchDbCoinVer2 queries the db to get v2 coins for `shardHeight` to `destHeight` and checks if the coins belong
// to any of the given IndexParam's using the given filters.
//
// The wider the range [startHeight: desHeight] is, the longer time this function should take.
//nolint // TODO: consider using get coin by index to speed up the performance.
func QueryBatchDbCoinVer2(idxParams map[string]*IndexParam, shardID byte, tokenID *common.Hash, startHeight, destHeight uint64, db *statedb.StateDB, cachedCoins *sync.Map, filters ...CoinMatcher) (map[string][]privacy.Coin, error) {
	start := startHeight

	res := make(map[string][]privacy.Coin)
	for otaStr := range idxParams {
		res[otaStr] = make([]privacy.Coin, 0)
	}

	burningPubKey := wallet.GetBurningPublicKey()
	countSkipped := 0
	for height := start; height <= destHeight; height++ {
		currentHeightCoins, err := statedb.GetOTACoinsByHeight(db, *tokenID, shardID, height)
		if err != nil {
			utils.Logger.Log.Errorf("Get outCoins ver 2 by height error: %v", err)
			return nil, err
		}
		for _, coinBytes := range currentHeightCoins {
			cv2 := &privacy.CoinV2{}
			err = cv2.SetBytes(coinBytes)
			if err != nil {
				utils.Logger.Log.Error("Get outCoins ver 2 from bytes", err)
				return nil, err
			}

			// check if the output coin was sent to the burning address
			if bytes.Equal(cv2.GetPublicKey().ToBytesS(), burningPubKey) {
				countSkipped++
				continue
			}

			if _, ok := cachedCoins.Load(cv2.GetPublicKey().String()); ok {
				// coin has already been cached, so we skip.
				countSkipped++
				continue
			}

			for otaStr, idxParam := range idxParams {
				if height < idxParam.FromHeight || height > idxParam.ToHeight {
					// Outside the required range, so we skip.
					continue
				}

				// create parameter(s) for CoinMatcher filters.
				otaKey := idxParam.OTAKey
				params := make(map[string]interface{})
				params["otaKey"] = otaKey
				params["db"] = db
				params["tokenID"] = tokenID

				pass := false
				for _, f := range filters {
					if f(cv2, params) {
						pass = true
						break
					}
				}
				if pass {
					res[otaStr] = append(res[otaStr], cv2)
					break
				}
			}
		}
	}
	utils.Logger.Log.Infof("#skipped for heights [%v,%v], tokenID %v: %v", start, destHeight, tokenID.String(), countSkipped)
	return res, nil
}

// QueryBatchDbCoinVer2ByIndices queries the db to get v2 coins with indices in [fromIndex,toIndex] checks if the coins belong
// to any of the given IndexParam's using the given filters.
//
// The wider the range [fromIndex: toIndex] is, the longer time this function should take.
func QueryBatchDbCoinVer2ByIndices(idxParams map[string]*IndexParam, shardID byte, tokenID *common.Hash, fromIndex, toIndex uint64, db *statedb.StateDB, cachedCoins *sync.Map, filters ...CoinMatcher) (map[string][]privacy.Coin, error) {
	if fromIndex > toIndex {
		return nil, fmt.Errorf("invalid index range [%v,%v]", fromIndex, toIndex)
	}

	res := make(map[string][]privacy.Coin)
	for otaStr := range idxParams {
		res[otaStr] = make([]privacy.Coin, 0)
	}

	burningPubKey := wallet.GetBurningPublicKey()
	countSkipped := 0
	for idx := fromIndex; idx <= toIndex; idx++ {
		coinBytes, err := statedb.GetOTACoinByIndex(db, *tokenID, idx, shardID)
		if err != nil {
			utils.Logger.Log.Errorf("Get outCoins ver 2 by idx error: %v", err)
			return nil, err
		}

		cv2 := &privacy.CoinV2{}
		err = cv2.SetBytes(coinBytes)
		if err != nil {
			utils.Logger.Log.Error("Get outCoins ver 2 from bytes", err)
			return nil, err
		}

		// check if the output coin was sent to the burning address
		if bytes.Equal(cv2.GetPublicKey().ToBytesS(), burningPubKey) {
			countSkipped++
			continue
		}

		if _, ok := cachedCoins.Load(cv2.GetPublicKey().String()); ok {
			// coin has already been cached, so we skip.
			countSkipped++
			continue
		}

		for otaStr, idxParam := range idxParams {
			// create parameter(s) for CoinMatcher filters.
			otaKey := idxParam.OTAKey
			params := make(map[string]interface{})
			params["otaKey"] = otaKey
			params["db"] = db
			params["tokenID"] = tokenID

			pass := false
			for _, f := range filters {
				if f(cv2, params) {
					pass = true
					break
				}
			}
			if pass {
				res[otaStr] = append(res[otaStr], cv2)
				break
			}
		}

	}
	utils.Logger.Log.Infof("#skipped for indices [%v,%v], tokenID %v: %v", fromIndex, toIndex, tokenID.String(), countSkipped)
	return res, nil
}

//nolint:gocritic
// getIdxParamsForIndexing chooses the IndexParam's for each shard queue based on the current size and the
// number of free workers.
func (ci *CoinIndexer) getIdxParamsForIndexing(totalWorker int) map[byte]int {
	res := make(map[byte]int)
	for i := 0; i < common.MaxShardNumber; i++ {
		res[byte(i)] = 0
	}

	if totalWorker <= common.MaxShardNumber {
		remainingWorker := totalWorker
		attempt := 0
		for remainingWorker > 0 {
			if attempt == 10 {
				// already loop with enough attempts but cannot find more, so we break.
				break
			}
			r := common.RandInt() % common.MaxShardNumber
			shardID := byte(r)
			if res[shardID] > 0 || len(ci.idxQueue[shardID]) == 0 {
				attempt += 1
				continue
			} else if len(ci.idxQueue[shardID]) < IndexingBatchSize {
				res[shardID] = len(ci.idxQueue[shardID])
			} else {
				res[shardID] = IndexingBatchSize
			}
			remainingWorker--
			attempt = 0
		}
	} else {
		for shard := 0; shard < common.MaxShardNumber; shard++ {
			shardID := byte(shard)
			if len(ci.idxQueue[shardID]) == 0 {
				continue
			} else if len(ci.idxQueue[shardID]) < IndexingBatchSize {
				res[shardID] = len(ci.idxQueue[shardID])
			} else {
				res[shardID] = IndexingBatchSize
			}
		}
	}

	return res
}
