package coinIndexer

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction/utils"
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

type IndexParam struct {
	FromHeight uint64
	ToHeight   uint64
	OTAKey     privacy.OTAKey
	TxDb       *statedb.StateDB
	ShardID    byte
	IsReset    bool
}

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
			pass, _ = c.ValidateAssetTag(sharedSecret, tokenID)
			return pass
		}
		return false
	}
}

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

func QueryDbCoinVer1(pubKey []byte, shardID byte, tokenID *common.Hash, db *statedb.StateDB) ([]privacy.Coin, error) {
	outCoinsBytes, err := statedb.GetOutcoinsByPubkey(db, *tokenID, pubKey, shardID)
	if err != nil {
		utils.Logger.Log.Error("GetOutcoinsBytesByKeyset Get by PubKey", err)
		return nil, err
	}
	var outCoins []privacy.Coin
	for _, item := range outCoinsBytes {
		outCoin := &privacy.CoinV1{}
		err := outCoin.SetBytes(item)
		if err != nil {
			utils.Logger.Log.Errorf("Cannot create coin from byte %v", err)
			return nil, err
		}
		outCoins = append(outCoins, outCoin)
	}
	return outCoins, nil
}

func QueryDbCoinVer2(otaKey privacy.OTAKey, shardID byte, tokenID *common.Hash, shardHeight, destHeight uint64, db *statedb.StateDB, filters ...CoinMatcher) ([]privacy.Coin, error) {
	var outCoins []privacy.Coin
	// avoid overlap; unless lower height is 0
	start := shardHeight + 1
	if shardHeight == 0 {
		start = 0
	}
	for height := start; height <= destHeight; height++ {
		currentHeightCoins, err := statedb.GetOTACoinsByHeight(db, *tokenID, shardID, height)
		if err != nil {
			utils.Logger.Log.Error("Get outcoins ver 2 bytes by keyset get by height", err)
			return nil, err
		}
		params := make(map[string]interface{})
		params["otaKey"] = otaKey
		params["db"] = db
		params["tokenID"] = tokenID
		for _, coinBytes := range currentHeightCoins {
			cv2 := &privacy.CoinV2{}
			err := cv2.SetBytes(coinBytes)
			if err != nil {
				utils.Logger.Log.Error("Get outcoins ver 2 from bytes", err)
				return nil, err
			}
			pass := true
			for _, f := range filters {
				if !f(cv2, params) {
					pass = false
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
func QueryBatchDbCoinVer2(idxParams map[string]IndexParam, shardID byte, tokenID *common.Hash, shardHeight, destHeight uint64, db *statedb.StateDB, cachedCoins map[string]interface{}, filters ...CoinMatcher) (map[string][]privacy.Coin, error) {
	// avoid overlap; unless lower height is 0
	start := shardHeight + 1
	if shardHeight == 0 {
		start = 0
	}

	res := make(map[string][]privacy.Coin)
	for otaStr := range idxParams {
		res[otaStr] = make([]privacy.Coin, 0)
	}

	countSkipped := 0
	for height := start; height <= destHeight; height++ {
		currentHeightCoins, err := statedb.GetOTACoinsByHeight(db, *tokenID, shardID, height)
		if err != nil {
			utils.Logger.Log.Errorf("Get outCoins ver 2 by height error: %v\n", err)
			return nil, err
		}
		for _, coinBytes := range currentHeightCoins {
			cv2 := &privacy.CoinV2{}
			err = cv2.SetBytes(coinBytes)
			if err != nil {
				utils.Logger.Log.Error("Get outCoins ver 2 from bytes", err)
				return nil, err
			}

			if _, ok := cachedCoins[cv2.GetPublicKey().String()]; ok {
				// coin has already been cached, so we skip.
				countSkipped++
				continue
			}

			for otaStr, idxParam := range idxParams {
				if height < idxParam.FromHeight || height > idxParam.ToHeight {
					// Outside the required range, so we skip.
					continue
				}

				otaKey := idxParam.OTAKey
				params := make(map[string]interface{})
				params["otaKey"] = otaKey
				params["db"] = db
				params["tokenID"] = tokenID

				pass := true
				for _, f := range filters {
					if !f(cv2, params) {
						pass = false
					} else {
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
	utils.Logger.Log.Infof("#skipped for heights %v to %v: %v\n", start, destHeight, countSkipped)
	return res, nil
}

//nolint:gocritic
// getIdxParamsForIndexing chooses the IdxParams for each shard queue based on the current size and the
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

func (ci *CoinIndexer) cloneCachedCoins() map[string]interface{} {
	res := make(map[string]interface{})
	if len(ci.cachedCoinPubKeys) != 0 {
		ci.mtx.Lock()
		for otaStr := range ci.cachedCoinPubKeys {
			res[otaStr] = true
		}
		ci.mtx.Unlock()
	}

	return res
}
