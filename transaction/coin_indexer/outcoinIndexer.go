// Package coinIndexer implements a UTXO cache for v2 output coins.
// Privacy V2 boosted up the privacy of a transaction but also resulted in retrieving output coins of users becoming
// more and more costly in terms of time and computing resources. This cache layer makes the retrieval easier.
// However, users have to submit their OTA keys to the cache (for the cache to be able to tell which output coins belong to them).
// This also reduces the anonymity level of a user. The good news is that the cache layer only knows which output coins
// belong to an OTAKey, it does not know their values.
package coinIndexer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/transaction/utils"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/privacy"
)

// CoinIndexer implements a UTXO cache for v2 output coins for faster retrieval.
type CoinIndexer struct {
	numWorkers          int // the maximum number of indexing go-routines for the enhanced cache.
	mtx                 *sync.RWMutex
	db                  incdb.Database
	accessTokens        map[string]bool
	idxQueue            map[byte][]*IndexParam
	queueSize           int
	statusChan          chan JobStatus
	quitChan            chan bool
	isAuthorizedRunning bool
	cachedCoinPubKeys   *sync.Map

	allTokens map[common.Hash]interface{}

	managedOTAKeys *sync.Map
	IdxChan        chan *IndexParam
}

// The following constants indicate the state of an OTAKey.
const (
	// StatusNotSubmitted indicates that an OTAKey has not been submitted
	StatusNotSubmitted = iota

	// StatusIndexing indicates that either the OTAKey is in the queue or it is being indexed.
	// This status usually happens when the OTAKey has been submitted in the enhanced manner.
	StatusIndexing

	// StatusKeySubmittedUsual indicates that the OTAKey has been submitted using the regular method.
	StatusKeySubmittedUsual

	// StatusIndexingFinished indicates the OTAKey has been submitted using the enhanced method and the indexing
	// procedure has been finished.
	StatusIndexingFinished
)

// NewOutCoinIndexer creates CoinIndexer instance.
//
//nolint:gocritic
func NewOutCoinIndexer(numWorkers int64, db incdb.Database, accessToken string, allTokens map[common.Hash]interface{}) (*CoinIndexer, error) {
	accessTokens := make(map[string]bool)
	authorizedCache := true
	if numWorkers != 0 && len(accessToken) > 0 {
		accessTokenBytes, err := hex.DecodeString(accessToken)
		if err != nil {
			utils.Logger.Log.Errorf("cannot decode the access token %v", accessToken)
			return nil, fmt.Errorf("cannot decode the access token %v", accessToken)
		} else if len(accessTokenBytes) != 32 {
			utils.Logger.Log.Errorf("access token is invalid")
			return nil, fmt.Errorf("access token is invalid")
		} else {
			accessTokens[accessToken] = true
		}
	} else {
		authorizedCache = false
		numWorkers = 0
	}
	utils.Logger.Log.Infof("NewOutCoinIndexer with %v workers", numWorkers)

	mtx := new(sync.RWMutex)
	m := &sync.Map{}

	// initialize the queue
	queueSize := 0
	idxQueue := make(map[byte][]*IndexParam)
	for shard := 0; shard < common.MaxShardNumber; shard++ {
		tmpQueue := make([]*IndexParam, 0)
		idxQueue[byte(shard)] = tmpQueue
	}

	// load from db once after startup
	loadedKeysRaw, err := rawdbv2.GetIndexedOTAKeys(db)
	if err == nil {
		for _, b := range loadedKeysRaw {
			var rawOTAKey [64]byte
			status := byte(0)
			copy(rawOTAKey[:], b[0:64])
			if len(b) == 65 {
				status = b[64]
			}

			if status > StatusIndexingFinished {
				utils.Logger.Log.Infof("State of OTAKey %x is invalid: %v", rawOTAKey, status)
				continue
			}

			if status == 0 {
				m.Store(rawOTAKey, StatusKeySubmittedUsual)
			} else {
				m.Store(rawOTAKey, int(status))
			}
			utils.Logger.Log.Infof("Loaded OTAKey %x with status %v", rawOTAKey, status)

			// In case this is an enhanced cache, try to re-add "un-done" params (i.e, params with status = StatusIndexing).
			// For these IdxParam's, as a quick fix, we'll have to re-cache from the beginning.
			//nolint // TODO(thanhn-inc): find a better solution.
			if status == StatusIndexing && authorizedCache {
				otaKey := OTAKeyFromRaw(rawOTAKey)
				if otaKey.GetPublicSpend() == nil || otaKey.GetOTASecretKey() == nil {
					utils.Logger.Log.Infof("invalid otaKey %x, %v", rawOTAKey, otaKey)
				}
				shardID := common.GetShardIDFromLastByte(otaKey.GetPublicSpend().ToBytesS()[len(otaKey.GetPublicSpend().ToBytesS())-1])

				idxParam := &IndexParam{
					FromHeight: config.Param().CoinVersion2LowestHeight,
					ToHeight:   0, // at this stage, we don't have the information about the ToHeight, it will be handled in the `Start` function.
					OTAKey:     otaKey,
					TxDb:       nil, // at this stage, we don't have the information about the TxDb, it will be handled in the `Start` function.
					ShardID:    shardID,
					IsReset:    true,
				}
				utils.Logger.Log.Infof("Added OTAKey %x to queue", rawOTAKey)
				idxQueue[shardID] = append(idxQueue[shardID], idxParam)
				queueSize++
			}
		}
	}
	utils.Logger.Log.Infof("Number of cached OTA keys: %v", len(loadedKeysRaw))

	cachedCoins := &sync.Map{}
	loadRawCachedCoinHashes, err := rawdbv2.GetCachedCoinHashes(db)
	numCached := 0
	if err == nil {
		for _, coinHash := range loadRawCachedCoinHashes {
			var temp [32]byte
			copy(temp[:], coinHash[:32])
			cachedCoins.Store(fmt.Sprintf("%x", temp), true)
			numCached++
		}
	}
	utils.Logger.Log.Infof("Number of cached coins: %v", numCached)
	utils.Logger.Log.Infof("Number of privacy tokens: %v", len(allTokens))

	ci := &CoinIndexer{
		numWorkers:          int(numWorkers),
		mtx:                 mtx,
		managedOTAKeys:      m,
		db:                  db,
		accessTokens:        accessTokens,
		cachedCoinPubKeys:   cachedCoins,
		isAuthorizedRunning: false,
		queueSize:           queueSize,
		idxQueue:            idxQueue,
		allTokens:           allTokens,
	}

	return ci, nil
}

// IsValidAccessToken checks if a user is authorized to use the enhanced cache.
//
// An access token is said to be valid if it is a hex-string of length 64.
func (ci *CoinIndexer) IsValidAccessToken(accessToken string) bool {
	atBytes, err := hex.DecodeString(accessToken)
	if err != nil || len(atBytes) != 32 {
		return false
	}
	return ci.accessTokens[accessToken]
}

// IsAuthorizedRunning checks if the current cache supports the enhanced mode.
func (ci *CoinIndexer) IsAuthorizedRunning() bool {
	return ci.isAuthorizedRunning
}

// GetManagedOTAKeys returns the ci.managedOTAKeys.
func (ci *CoinIndexer) GetManagedOTAKeys() *sync.Map {
	return ci.managedOTAKeys
}

// RemoveOTAKey removes an OTAKey from the cached database.
//
// nolint // TODO: remove cached output coins, access token.
func (ci *CoinIndexer) RemoveOTAKey(otaKey privacy.OTAKey) error {
	keyBytes := OTAKeyToRaw(otaKey)
	err := rawdbv2.DeleteIndexedOTAKey(ci.db, keyBytes[:])
	if err != nil {
		return err
	}
	ci.managedOTAKeys.Delete(keyBytes)

	return nil
}

// AddOTAKey adds a new OTAKey to the cache list.
func (ci *CoinIndexer) AddOTAKey(otaKey privacy.OTAKey, state int) error {
	keyBytes := OTAKeyToRaw(otaKey)
	err := rawdbv2.StoreIndexedOTAKey(ci.db, keyBytes[:], state)
	if err != nil {
		return err
	}
	ci.managedOTAKeys.Store(keyBytes, state)
	return nil
}

func (ci *CoinIndexer) GetAllOTAKey() ([][]byte, error) {
	return rawdbv2.GetIndexedOTAKeys(ci.db)
}

// AddOTAKey adds a new OTAKey to the cache list.
func (ci *CoinIndexer) WriteWhitelistOTAKey(whitelist []privacy.OTAKey) error {
	whitelistBytes := [][rawdbv2.OTAKeyBytesLength]byte{}
	for _, k := range whitelist {
		kByte := OTAKeyToRaw(k)
		whitelistBytes = append(whitelistBytes, kByte)
	}
	err := rawdbv2.StoreWhitelistOTAKeys(ci.db, whitelistBytes)
	if err != nil {
		return err
	}
	return ci.resetCachedOTAByWhitelist()
}

func (ci *CoinIndexer) AddWhitelistOTAKey(whitelist []privacy.OTAKey) error {
	whitelistBytes := [][rawdbv2.OTAKeyBytesLength]byte{}
	for _, k := range whitelist {
		kByte := OTAKeyToRaw(k)
		whitelistBytes = append(whitelistBytes, kByte)
	}
	curKeyMap := map[[rawdbv2.OTAKeyBytesLength]byte]interface{}{}
	curList, err := rawdbv2.GetWhitelistOTAKeys(ci.db)
	if (len(curList) > 0) && (err != nil) {
		for _, k := range curList {
			curKeyMap[k] = nil
		}
	}
	for _, k := range whitelistBytes {
		if _, has := curKeyMap[k]; !has {
			curList = append(curList, k)
		}
	}
	err = rawdbv2.StoreWhitelistOTAKeys(ci.db, curList)
	if err != nil {
		return err
	}
	return ci.resetCachedOTAByWhitelist()
}

// AddOTAKey adds a new OTAKey to the cache list.
func (ci *CoinIndexer) GetMapWhitelistOTAKey() (map[[rawdbv2.OTAKeyBytesLength]byte]interface{}, error) {
	curList, err := rawdbv2.GetWhitelistOTAKeys(ci.db)
	if err != nil {
		return nil, err
	}
	curKeyMap := map[[rawdbv2.OTAKeyBytesLength]byte]interface{}{}
	for _, k := range curList {
		curKeyMap[k] = nil
	}
	return curKeyMap, nil
}

func (ci *CoinIndexer) resetCachedOTAByWhitelist() error {
	curMap, err := ci.GetMapWhitelistOTAKey()
	if err != nil {
		return err
	}
	needToDel := []interface{}{}
	ci.managedOTAKeys.Range(func(key, value interface{}) bool {
		if keyBytes, ok := key.([64]byte); ok {
			if _, has := curMap[keyBytes]; !has {
				needToDel = append(needToDel, key)
			}
		}
		return true
	})
	for _, k := range needToDel {
		ci.managedOTAKeys.Delete(k)
	}
	for _, k := range curMap {
		ci.managedOTAKeys.LoadOrStore(k, StatusIndexing)
	}
	return nil
}

// HasOTAKey checks if an OTAKey has been added to the indexer and returns the state of OTAKey.
// The returned state could be:
//   - StatusNotSubmitted
//   - StatusIndexing
//   - StatusKeySubmittedUsual
//   - StatusIndexingFinished
func (ci *CoinIndexer) HasOTAKey(k [64]byte) (bool, int) {
	var result int
	val, ok := ci.managedOTAKeys.Load(k)
	if ok {
		result, ok = val.(int)
	}
	return ok, result
}

// CacheCoinPublicKey stores the public key of a cached output coin to mark an output coin as cached.
func (ci *CoinIndexer) CacheCoinPublicKey(coinPublicKey *privacy.Point) error {
	err := rawdbv2.StoreCachedCoinHash(ci.db, coinPublicKey.ToBytesS())
	if err != nil {
		return err
	}
	ci.cachedCoinPubKeys.Store(coinPublicKey.String(), true)
	utils.Logger.Log.Infof("Add coinPublicKey %v success", coinPublicKey.String())
	return nil
}

// IsQueueFull checks if the current indexing queue is full.
//
// The idxQueue size for each shard is as large as the number of workers.
func (ci *CoinIndexer) IsQueueFull(shardID byte) bool {
	return len(ci.idxQueue[shardID]) >= ci.numWorkers
}

// ReIndexOutCoin re-scans all output coins from idxParams.FromHeight to idxParams.ToHeight and adds them to the cache if the belongs to idxParams.OTAKey.
func (ci *CoinIndexer) ReIndexOutCoin(idxParams *IndexParam) {
	status := JobStatus{
		otaKey: idxParams.OTAKey,
		err:    nil,
	}

	vkb := OTAKeyToRaw(idxParams.OTAKey)
	utils.Logger.Log.Infof("[CoinIndexer] Re-index output coins for key %x", idxParams.OTAKey)
	keyExists, state := ci.HasOTAKey(vkb)
	if keyExists {
		if state == StatusIndexing {
			utils.Logger.Log.Errorf("[CoinIndexer] ota key %v is being processed", idxParams.OTAKey)
			ci.statusChan <- status
			return
		}
		// resetting entries for this key is reserved for debugging RPCs
		if state == StatusIndexingFinished && !idxParams.IsReset {
			utils.Logger.Log.Errorf("[CoinIndexer] ota key %v has been processed and isReset = false", idxParams.OTAKey)

			ci.statusChan <- status
			return
		}
	}
	if state != StatusIndexing {
		ci.managedOTAKeys.Store(vkb, StatusIndexing)
	}

	defer func() {
		if r := recover(); r != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Recovered from: %v", r)
		}
		if exists, tmpState := ci.HasOTAKey(vkb); exists && tmpState == StatusIndexing {
			ci.managedOTAKeys.Delete(vkb)
		}
	}()
	var allOutputCoins []privacy.Coin

	start := time.Now()
	for height := idxParams.FromHeight; height <= idxParams.ToHeight; {
		tmpStart := time.Now()
		nextHeight := height + utils.MaxOutcoinQueryInterval

		// query token output coins
		currentOutputCoinsToken, err := QueryDbCoinVer2(idxParams.OTAKey, &common.ConfidentialAssetID, height, nextHeight-1, idxParams.TxDb, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying token coins from db - %v", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		// query PRV output coins
		currentOutputCoinsPRV, err := QueryDbCoinVer2(idxParams.OTAKey, &common.PRVCoinID, height, nextHeight-1, idxParams.TxDb, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		utils.Logger.Log.Infof("[CoinIndexer] Key %x, %d to %d: found %d PRV + %d pToken coins, timeElapsed %v", vkb, height, nextHeight-1, len(currentOutputCoinsPRV), len(currentOutputCoinsToken), time.Since(tmpStart).Seconds())

		allOutputCoins = append(allOutputCoins, append(currentOutputCoinsToken, currentOutputCoinsPRV...)...)
		height = nextHeight
	}

	// write
	err := ci.AddOTAKey(idxParams.OTAKey, StatusIndexingFinished)
	if err == nil {
		err = ci.StoreIndexedOutputCoins(idxParams.OTAKey, allOutputCoins, idxParams.ShardID)
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins error: %v", err)

			status.err = err
			ci.statusChan <- status
			return
		}
	} else {
		utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey error: %v", err)

		status.err = err
		ci.statusChan <- status
		return
	}

	utils.Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x, timeElapsed: %v", vkb, time.Since(start).Seconds())

	status.err = nil
	ci.statusChan <- status
}

// ReIndexOutCoinBatch re-scans all output coins for a list of indexing params of the same shardID.
//
// Callers must manage to make sure all indexing params belong to the same shard.
//
//nolint:revive // using defer in loop
func (ci *CoinIndexer) ReIndexOutCoinBatch(idxParams []*IndexParam, txDb *statedb.StateDB, id string) {
	if len(idxParams) == 0 {
		return
	}

	// create some map instances and necessary params
	mapIdxParams := make(map[string]*IndexParam)
	mapStatuses := make(map[string]JobStatus)
	mapOutputCoins := make(map[string][]privacy.Coin)
	minHeight := uint64(math.MaxUint64)
	maxHeight := uint64(0)
	shardID := idxParams[0].ShardID
	for _, idxParam := range idxParams {
		otaStr := fmt.Sprintf("%x", OTAKeyToRaw(idxParam.OTAKey))
		mapIdxParams[otaStr] = idxParam
		mapStatuses[otaStr] = JobStatus{id: id, otaKey: idxParam.OTAKey, err: nil}
		mapOutputCoins[otaStr] = make([]privacy.Coin, 0)

		if idxParam.FromHeight < minHeight {
			minHeight = idxParam.FromHeight
		}
		if idxParam.ToHeight > maxHeight {
			maxHeight = idxParam.ToHeight
		}
	}

	for otaStr, idxParam := range mapIdxParams {
		vkb := OTAKeyToRaw(idxParam.OTAKey)
		utils.Logger.Log.Infof("[CoinIndexer] Re-index output coins for key %x", idxParam.OTAKey)
		keyExists, state := ci.HasOTAKey(vkb)
		if keyExists {
			if state == StatusIndexingFinished && !idxParam.IsReset {
				utils.Logger.Log.Errorf("[CoinIndexer] ota key %v has been processed with status %v and isReset = false", idxParam.OTAKey, state)

				ci.statusChan <- mapStatuses[otaStr]
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
		}
		if state != StatusIndexing {
			ci.managedOTAKeys.Store(vkb, StatusIndexing)
		}

		defer func() {
			if r := recover(); r != nil {
				utils.Logger.Log.Errorf("[CoinIndexer] Recovered from: %v", r)
			}
			if exists, tmpState := ci.HasOTAKey(vkb); exists && tmpState == StatusIndexing {
				ci.managedOTAKeys.Delete(vkb)
			}
		}()
	}

	if len(mapIdxParams) == 0 {
		utils.Logger.Log.Infof("[CoinIndexer] No indexParam to proceed")
		return
	}

	// in case minHeight > maxHeight, all indexing params will fail
	if minHeight == 0 {
		minHeight = 1
	}
	if minHeight > maxHeight {
		err := fmt.Errorf("minHeight (%v) > maxHeight (%v) when re-indexing outcoins", minHeight, maxHeight)
		for otaStr := range mapStatuses {
			status := mapStatuses[otaStr]
			status.err = err
			ci.statusChan <- status
			delete(mapIdxParams, otaStr)
			delete(mapStatuses, otaStr)
			delete(mapOutputCoins, otaStr)
		}
		return
	}

	start := time.Now()
	for height := minHeight; height <= maxHeight; {
		tmpStart := time.Now() // measure time for each round
		nextHeight := height + utils.MaxOutcoinQueryInterval

		// query token output coins
		currentOutputCoinsToken, err := QueryBatchDbCoinVer2(mapIdxParams, shardID, &common.ConfidentialAssetID, height, nextHeight-1, txDb, ci.cachedCoinPubKeys, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying token coins from db - %v", err)

			for otaStr := range mapStatuses {
				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
			return
		}

		// query PRV output coins
		currentOutputCoinsPRV, err := QueryBatchDbCoinVer2(mapIdxParams, shardID, &common.PRVCoinID, height, nextHeight-1, txDb, ci.cachedCoinPubKeys, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v", err)

			for otaStr := range mapStatuses {
				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
			return
		}

		// Add output coins to maps
		for otaStr, listOutputCoins := range mapOutputCoins {
			listOutputCoins = append(listOutputCoins, currentOutputCoinsToken[otaStr]...)
			listOutputCoins = append(listOutputCoins, currentOutputCoinsPRV[otaStr]...)

			utils.Logger.Log.Infof("[CoinIndexer] Key %v, %d to %d: found %d PRV + %d pToken coins, current #coins %v, timeElapsed %v", otaStr, height, nextHeight-1, len(currentOutputCoinsPRV[otaStr]), len(currentOutputCoinsToken[otaStr]), len(listOutputCoins), time.Since(tmpStart).Seconds())
			mapOutputCoins[otaStr] = listOutputCoins
		}

		height = nextHeight
	}

	// write
	for otaStr, idxParam := range mapIdxParams {
		vkb := OTAKeyToRaw(idxParam.OTAKey)
		allOutputCoins := mapOutputCoins[otaStr]

		err := ci.AddOTAKey(idxParam.OTAKey, StatusIndexingFinished)
		if err == nil {
			utils.Logger.Log.Infof("[CoinIndexer] About to store %v output coins for OTAKey %x", len(allOutputCoins), vkb)
			err = ci.StoreIndexedOutputCoins(idxParam.OTAKey, allOutputCoins, shardID)
			if err != nil {
				utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins for OTA key %x error: %v", vkb, err)

				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
				continue
			}
		} else {
			utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey %x, error: %v", vkb, err)

			status := mapStatuses[otaStr]
			status.err = err
			ci.statusChan <- status
			delete(mapIdxParams, otaStr)
			delete(mapStatuses, otaStr)
			delete(mapOutputCoins, otaStr)
			continue
		}

		utils.Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x, found %v coins, timeElapsed: %v", vkb, len(allOutputCoins), time.Since(start).Seconds())

		ci.statusChan <- mapStatuses[otaStr]
		delete(mapIdxParams, otaStr)
		delete(mapStatuses, otaStr)
		delete(mapOutputCoins, otaStr)
	}
}

// ReIndexOutCoinBatchByIndices re-scans all output coins for a list of indexing params of the same shardID.
//
// Callers must manage to make sure all indexing params belong to the same shard.
//
//nolint:revive // using defer in loop
func (ci *CoinIndexer) ReIndexOutCoinBatchByIndices(idxParams []*IndexParam, txDb *statedb.StateDB, id string) {
	if len(idxParams) == 0 {
		return
	}

	// create some map instances and necessary params
	mapIdxParams := make(map[string]*IndexParam)
	mapStatuses := make(map[string]JobStatus)
	mapOutputCoins := make(map[string][]privacy.Coin)
	shardID := idxParams[0].ShardID
	for _, idxParam := range idxParams {
		otaStr := fmt.Sprintf("%x", OTAKeyToRaw(idxParam.OTAKey))
		mapIdxParams[otaStr] = idxParam
		mapStatuses[otaStr] = JobStatus{id: id, otaKey: idxParam.OTAKey, err: nil}
		mapOutputCoins[otaStr] = make([]privacy.Coin, 0)
	}

	for otaStr, idxParam := range mapIdxParams {
		vkb := OTAKeyToRaw(idxParam.OTAKey)
		utils.Logger.Log.Infof("[CoinIndexer] Re-index output coins for key %x", idxParam.OTAKey)
		keyExists, state := ci.HasOTAKey(vkb)
		if keyExists {
			if state == StatusIndexingFinished && !idxParam.IsReset {
				utils.Logger.Log.Errorf("[CoinIndexer] ota key %v has been processed with status %v and isReset = false", idxParam.OTAKey, state)

				ci.statusChan <- mapStatuses[otaStr]
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
		}
		if state != StatusIndexing {
			ci.managedOTAKeys.Store(vkb, StatusIndexing)
		}

		defer func() {
			if r := recover(); r != nil {
				utils.Logger.Log.Errorf("[CoinIndexer] Recovered from: %v", r)
			}
			if exists, tmpState := ci.HasOTAKey(vkb); exists && tmpState == StatusIndexing {
				ci.managedOTAKeys.Delete(vkb)
			}
		}()
	}

	if len(mapIdxParams) == 0 {
		utils.Logger.Log.Infof("[CoinIndexer] No indexParam to proceed")
		return
	}

	prvCoinLength, err := statedb.GetOTACoinLength(txDb, common.PRVCoinID, shardID)
	if err != nil {
		utils.Logger.Log.Errorf("[CoinIndexer] GetOTACoinLength for PRV error: %v", err)

		for otaStr := range mapStatuses {
			status := mapStatuses[otaStr]
			status.err = err
			ci.statusChan <- status
			delete(mapIdxParams, otaStr)
			delete(mapStatuses, otaStr)
			delete(mapOutputCoins, otaStr)
		}
		return
	}
	tokenCoinLength, err := statedb.GetOTACoinLength(txDb, common.ConfidentialAssetID, shardID)
	if err != nil {
		utils.Logger.Log.Errorf("[CoinIndexer] GetOTACoinLength for TOKEN error: %v", err)

		for otaStr := range mapStatuses {
			status := mapStatuses[otaStr]
			status.err = err
			ci.statusChan <- status
			delete(mapIdxParams, otaStr)
			delete(mapStatuses, otaStr)
			delete(mapOutputCoins, otaStr)
		}
		return
	}
	utils.Logger.Log.Infof("[CoinIndexer] Current #Coins for shard %v: PRV %v, TOKEN %v", shardID, prvCoinLength.Uint64(), tokenCoinLength.Uint64())

	start := time.Now()
	for idx := uint64(0); idx < prvCoinLength.Uint64(); {
		nextIdx := idx + utils.MaxOutcoinQueryInterval
		if nextIdx > prvCoinLength.Uint64() {
			nextIdx = prvCoinLength.Uint64()
		}

		// query token output coins
		currentOutputCoinsPRV, err := QueryBatchDbCoinVer2ByIndices(mapIdxParams, shardID, &common.PRVCoinID, idx, nextIdx-1, txDb, ci.cachedCoinPubKeys, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v", err)

			for otaStr := range mapStatuses {
				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
			return
		}

		// Add output coins to maps
		for otaStr, listOutputCoins := range mapOutputCoins {
			listOutputCoins = append(listOutputCoins, currentOutputCoinsPRV[otaStr]...)

			utils.Logger.Log.Infof("[CoinIndexer] Key %v, %d to %d: found %d PRV coins, current #coins %v, timeElapsed %v", otaStr, idx, nextIdx-1, len(currentOutputCoinsPRV[otaStr]), len(listOutputCoins), time.Since(start).Seconds())
			mapOutputCoins[otaStr] = listOutputCoins
		}

		idx = nextIdx
	}

	for idx := uint64(0); idx < tokenCoinLength.Uint64(); {
		nextIdx := idx + utils.MaxOutcoinQueryInterval
		if nextIdx > tokenCoinLength.Uint64() {
			nextIdx = tokenCoinLength.Uint64()
		}

		// query token output coins
		currentOutputCoinsToken, err := QueryBatchDbCoinVer2ByIndices(mapIdxParams, shardID, &common.ConfidentialAssetID, idx, nextIdx-1, txDb, ci.cachedCoinPubKeys, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying Token coins from db - %v", err)

			for otaStr := range mapStatuses {
				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
			return
		}

		// Add output coins to maps
		for otaStr, listOutputCoins := range mapOutputCoins {
			listOutputCoins = append(listOutputCoins, currentOutputCoinsToken[otaStr]...)

			utils.Logger.Log.Infof("[CoinIndexer] Key %v, %d to %d: found %d Token coins, current #coins %v, timeElapsed %v", otaStr, idx, nextIdx-1, len(currentOutputCoinsToken[otaStr]), len(listOutputCoins), time.Since(start).Seconds())
			mapOutputCoins[otaStr] = listOutputCoins
		}

		idx = nextIdx
	}

	// write
	for otaStr, idxParam := range mapIdxParams {
		vkb := OTAKeyToRaw(idxParam.OTAKey)
		allOutputCoins := mapOutputCoins[otaStr]

		err := ci.AddOTAKey(idxParam.OTAKey, StatusIndexingFinished)
		if err == nil {
			utils.Logger.Log.Infof("[CoinIndexer] About to store %v output coins for OTAKey %x", len(allOutputCoins), vkb)
			err = ci.StoreIndexedOutputCoins(idxParam.OTAKey, allOutputCoins, shardID)
			if err != nil {
				utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins for OTA key %x error: %v", vkb, err)

				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
				continue
			}
		} else {
			utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey %x, error: %v", vkb, err)

			status := mapStatuses[otaStr]
			status.err = err
			ci.statusChan <- status
			delete(mapIdxParams, otaStr)
			delete(mapStatuses, otaStr)
			delete(mapOutputCoins, otaStr)
			continue
		}

		utils.Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x, found %v coins, timeElapsed: %v", vkb, len(allOutputCoins), time.Since(start).Seconds())

		ci.statusChan <- mapStatuses[otaStr]
		delete(mapIdxParams, otaStr)
		delete(mapStatuses, otaStr)
		delete(mapOutputCoins, otaStr)
	}
}

// GetIndexedOutCoin returns the indexed (i.e, cached) output coins of an otaKey w.r.t a tokenID.
// It returns an error if the otaKey hasn't been submitted (state = 0) or the indexing is in progress (state = 1).
func (ci *CoinIndexer) GetIndexedOutCoin(otaKey privacy.OTAKey, tokenID *common.Hash, txDb *statedb.StateDB, shardID byte) ([]privacy.Coin, int, error) {
	vkb := OTAKeyToRaw(otaKey)
	utils.Logger.Log.Infof("Retrieve re-indexed coins for %x from db %v", vkb, ci.db)
	_, status := ci.HasOTAKey(vkb)
	if status == StatusIndexing {
		return nil, status, fmt.Errorf("OTA Key %x not ready : Sync still in progress", otaKey)
	}
	if status == StatusNotSubmitted {
		// this is a new view key
		return nil, status, fmt.Errorf("OTA Key %x not synced", otaKey)
	}
	ocBytes, err := rawdbv2.GetOutCoinsByIndexedOTAKey(ci.db, common.ConfidentialAssetID, shardID, vkb[:])
	if err != nil {
		return nil, status, err
	}
	params := make(map[string]interface{})
	params["otaKey"] = otaKey
	params["tokenID"] = tokenID
	filter := GetCoinFilterByOTAKeyAndToken()
	var result []privacy.Coin
	for _, cb := range ocBytes {
		temp := &privacy.CoinV2{}
		err := temp.SetBytes(cb)
		if err != nil {
			return nil, status, fmt.Errorf("coin by OTAKey storage is corrupted")
		}
		if filter(temp, params) {
			// eliminate forked coins
			if dbHasOta, _, err := statedb.HasOnetimeAddress(txDb, *tokenID, temp.GetPublicKey().ToBytesS()); dbHasOta && err == nil {
				result = append(result, temp)
			}
		}
	}
	return result, status, nil
}

// StoreIndexedOutputCoins stores output coins that have been indexed into the cache db. It also keep tracks of each
// output coin hash to boost up the retrieval process.
func (ci *CoinIndexer) StoreIndexedOutputCoins(otaKey privacy.OTAKey, outputCoins []privacy.Coin, shardID byte) error {
	var ocBytes [][]byte
	for _, c := range outputCoins {
		ocBytes = append(ocBytes, c.Bytes())
	}
	vkb := OTAKeyToRaw(otaKey)
	utils.Logger.Log.Infof("Store %d indexed coins to db %x", len(ocBytes), vkb)
	// all token and PRV coins are grouped together; match them to desired tokenID upon retrieval
	err := rawdbv2.StoreIndexedOutCoins(ci.db, common.ConfidentialAssetID, vkb[:], ocBytes, shardID)
	if err != nil {
		return err
	}

	// cache the coin's hash to reduce the number of checks
	for _, c := range outputCoins {
		err = ci.CacheCoinPublicKey(c.GetPublicKey())
		if err != nil {
			return err
		}
	}

	return nil
}

// AddTokenID add a new TokenID to the cache layer.
func (ci *CoinIndexer) AddTokenID(tokenID common.Hash) {
	if !tokenID.IsZeroValue() {
		ci.mtx.Lock()
		ci.allTokens[tokenID] = true
		ci.mtx.Unlock()
		utils.Logger.Log.Infof("Add tokenID %v to cache", tokenID.String())
	}
}

// GetAllTokenIDs returns all tokenIDs from the cache layer.
func (ci *CoinIndexer) GetAllTokenIDs() map[common.Hash]interface{} {
	res := make(map[common.Hash]interface{})
	ci.mtx.Lock()
	for tokenId := range ci.allTokens {
		res[tokenId] = true
	}
	ci.mtx.Unlock()

	return res
}

// Start starts the CoinIndexer in case the authorized cache is employed.
// It is a hub to
//   - record key submission from users;
//   - record the indexing status of keys;
//   - collect keys into batches and index them all together in a batching way.
func (ci *CoinIndexer) Start(cfg *IndexerInitialConfig) {
	ci.mtx.Lock()
	if ci.isAuthorizedRunning {
		ci.mtx.Unlock()
		return
	}

	ci.IdxChan = make(chan *IndexParam, 10*ci.numWorkers)
	ci.statusChan = make(chan JobStatus, 10*ci.numWorkers)
	ci.quitChan = make(chan bool)

	for shardID := 0; shardID < common.MaxShardNumber; shardID++ {
		for _, idxParam := range ci.idxQueue[byte(shardID)] {
			idxParam.ToHeight = cfg.BestBlocks[shardID]
			idxParam.TxDb = cfg.TxDbs[shardID]
		}
	}

	// A map to keep track of the number of IdxParam's per go-routine
	tracking := make(map[string]int)
	var id string

	utils.Logger.Log.Infof("Start CoinIndexer....")
	ci.isAuthorizedRunning = true
	ci.mtx.Unlock()

	var err error
	numWorking := 0
	start := time.Now()
	for {
		select {
		case status := <-ci.statusChan:
			ci.mtx.Lock()
			tracking[status.id] -= 1
			if tracking[status.id] <= 0 {
				numWorking--
			}
			ci.mtx.Unlock()

			otaKeyBytes := OTAKeyToRaw(status.otaKey)
			if status.err != nil {
				utils.Logger.Log.Errorf("IndexOutCoin for otaKey %x failed: %v", otaKeyBytes, status.err)
				err = ci.RemoveOTAKey(status.otaKey)
				if err != nil {
					utils.Logger.Log.Errorf("Remove OTAKey %v error: %v", otaKeyBytes, err)
				}
			} else {
				utils.Logger.Log.Infof("Finished indexing output coins for otaKey: %x", otaKeyBytes)
			}

		case idxParams := <-ci.IdxChan:
			otaKeyBytes := OTAKeyToRaw(idxParams.OTAKey)
			utils.Logger.Log.Infof("New authorized OTAKey received: %x", otaKeyBytes)

			ci.mtx.Lock()
			ci.idxQueue[idxParams.ShardID] = append(ci.idxQueue[idxParams.ShardID], idxParams)
			ci.queueSize++
			ci.mtx.Unlock()

		case <-ci.quitChan:
			ci.mtx.Lock()
			ci.isAuthorizedRunning = false
			ci.mtx.Unlock()

			utils.Logger.Log.Infof("Stopped coinIndexer!!")
			return
		default:
			if numWorking < ci.numWorkers && ci.queueSize > 0 {
				// collect indexing params by intervals to (possibly) reduce the number of go routines
				if time.Since(start).Seconds() < BatchWaitingTime {
					continue
				}
				remainingWorker := ci.numWorkers - numWorking
				// get idxParams for the ci
				workersForEach := ci.getIdxParamsForIndexing(remainingWorker)
				for shard, numParams := range workersForEach {
					if numParams != 0 {
						ci.mtx.Lock()

						// decrease the queue size
						ci.queueSize -= numParams

						idxParams := ci.idxQueue[shard][:numParams]
						ci.idxQueue[shard] = ci.idxQueue[shard][numParams:]

						jsb, _ := json.Marshal(idxParams)
						id = common.HashH(append(jsb, common.RandBytes(32)...)).String()
						tracking[id] = numParams

						utils.Logger.Log.Infof("Re-index for %v new OTA keys, shard %v, id %v", numParams, shard, id)

						ci.mtx.Unlock()

						// increase 1 go-routine
						numWorking += 1
						go ci.ReIndexOutCoinBatchByIndices(idxParams, idxParams[0].TxDb, id)
					}
				}
				start = time.Now()
			} else {
				utils.Logger.Log.Infof("CoinIndexer is full or no OTA key is found in queue, numWorking %v, queueSize %v", numWorking, ci.queueSize)
				time.Sleep(10 * time.Second)
			}
		}
	}
}

// Stop terminates the CoinIndexer.
func (ci *CoinIndexer) Stop() {
	if ci.isAuthorizedRunning {
		ci.quitChan <- true
	}
}
