package coinIndexer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"math"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/privacy"
)

type CoinIndexer struct {
	numWorkers          int // the maximum number of indexing go-routines for the enhanced cache.
	mtx                 *sync.RWMutex
	db                  incdb.Database
	accessTokens        map[string]bool
	idxQueue            map[byte][]IndexParam
	queueSize           int
	statusChan          chan JobStatus
	quitChan            chan bool
	isAuthorizedRunning bool
	cachedCoinPubKeys   map[string]interface{}

	ManagedOTAKeys *sync.Map
	IdxChan        chan IndexParam
}

//nolint:gocritic
// NewOutCoinIndexer creates a new full node's caching instance for faster output coin retrieval.
func NewOutCoinIndexer(numWorkers int64, db incdb.Database, accessToken string) (*CoinIndexer, error) {
	accessTokens := make(map[string]bool)
	if numWorkers != 0 && len(accessToken) > 0 {
		accessTokenBytes, err := hex.DecodeString(accessToken)
		if err != nil {
			utils.Logger.Log.Errorf("cannot decode the access token %v\n", accessToken)
			return nil, fmt.Errorf("cannot decode the access token %v", accessToken)
		} else if len(accessTokenBytes) != 32 {
			utils.Logger.Log.Errorf("access token is invalid")
			return nil, fmt.Errorf("access token is invalid")
		} else {
			accessTokens[accessToken] = true
		}
	} else {
		numWorkers = 0
	}
	utils.Logger.Log.Infof("NewOutCoinIndexer with %v workers\n", numWorkers)

	mtx := new(sync.RWMutex)
	m := &sync.Map{}

	// load from db once after startup
	loadedKeysRaw, err := rawdbv2.GetIndexedOTAKeys(db)
	if err == nil {
		for _, b := range loadedKeysRaw {
			var temp [64]byte
			copy(temp[:], b[0:64])
			m.Store(temp, 2)
		}
	}
	utils.Logger.Log.Infof("Number of cached OTA keys: %v\n", len(loadedKeysRaw))

	cachedCoins := make(map[string]interface{})
	loadRawCachedCoinHashes, err := rawdbv2.GetCachedCoinHashes(db)
	if err == nil {
		for _, coinHash := range loadRawCachedCoinHashes {
			var temp [32]byte
			copy(temp[:], coinHash[:32])
			cachedCoins[fmt.Sprintf("%x", temp)] = true
		}
	}
	utils.Logger.Log.Infof("Number of cached coins: %v\n", len(cachedCoins))

	ci := &CoinIndexer{
		numWorkers:          int(numWorkers),
		mtx:                 mtx,
		ManagedOTAKeys:      m,
		db:                  db,
		accessTokens:        accessTokens,
		cachedCoinPubKeys:   cachedCoins,
		isAuthorizedRunning: false,
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

// RemoveOTAKey removes an OTAKey from the cached database.
//
//nolint // TODO: remove cached output coins, access token.
func (ci *CoinIndexer) RemoveOTAKey(otaKey privacy.OTAKey) error {
	keyBytes := OTAKeyToRaw(otaKey)
	err := rawdbv2.DeleteIndexedOTAKey(ci.db, keyBytes[:])
	if err != nil {
		return err
	}
	ci.ManagedOTAKeys.Delete(keyBytes)

	return nil
}

// AddOTAKey adds a new OTAKey to the cache list.
func (ci *CoinIndexer) AddOTAKey(otaKey privacy.OTAKey) error {
	keyBytes := OTAKeyToRaw(otaKey)
	err := rawdbv2.StoreIndexedOTAKey(ci.db, keyBytes[:])
	if err != nil {
		return err
	}
	ci.ManagedOTAKeys.Store(keyBytes, 2)
	return nil
}

func (ci *CoinIndexer) HasOTAKey(k [64]byte) (bool, int) {
	var result int
	val, ok := ci.ManagedOTAKeys.Load(k)
	if ok {
		result, ok = val.(int)
	}
	return ok, result
}

func (ci *CoinIndexer) CacheCoinPublicKey(coinPublicKey *privacy.Point) error {
	err := rawdbv2.StoreCachedCoinHash(ci.db, coinPublicKey.ToBytesS())
	if err != nil {
		return err
	}
	ci.cachedCoinPubKeys[coinPublicKey.String()] = true
	utils.Logger.Log.Infof("Add coinPublicKey %v success\n", coinPublicKey.String())
	return nil
}

// IsQueueFull checks if the current indexing queue is full.
//
// The idxQueue size for each shard is as large as the number of workers.
func (ci *CoinIndexer) IsQueueFull(shardID byte) bool {
	return len(ci.idxQueue[shardID]) >= ci.numWorkers
}

// ReIndexOutCoin re-scans all output coins from idxParams.FromHeight to idxParams.ToHeight and adds them to the cache if the belongs to idxParams.OTAKey.
func (ci *CoinIndexer) ReIndexOutCoin(idxParams IndexParam) {
	status := JobStatus{
		otaKey: idxParams.OTAKey,
		err:    nil,
	}

	vkb := OTAKeyToRaw(idxParams.OTAKey)
	utils.Logger.Log.Infof("[CoinIndexer] Re-index output coins for key %x", idxParams.OTAKey)
	keyExists, processing := ci.HasOTAKey(vkb)
	if keyExists {
		if processing == 1 {
			utils.Logger.Log.Errorf("[CoinIndexer] ota key %v is being processed", idxParams.OTAKey)
			ci.statusChan <- status
			return
		}
		// resetting entries for this key is reserved for debugging RPCs
		if processing == 2 && !idxParams.IsReset {
			utils.Logger.Log.Errorf("[CoinIndexer] ota key %v has been processed and isReset = false", idxParams.OTAKey)

			ci.statusChan <- status
			return
		}
	}
	ci.ManagedOTAKeys.Store(vkb, 1)
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Recovered from: %v\n", r)
		}
		if exists, processing := ci.HasOTAKey(vkb); exists && processing == 1 {
			ci.ManagedOTAKeys.Delete(vkb)
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
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying token coins from db - %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		// query PRV output coins
		currentOutputCoinsPRV, err := QueryDbCoinVer2(idxParams.OTAKey, &common.PRVCoinID, height, nextHeight-1, idxParams.TxDb, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		utils.Logger.Log.Infof("[CoinIndexer] Key %x, %d to %d: found %d PRV + %d pToken coins, timeElapsed %v\n", vkb, height, nextHeight-1, len(currentOutputCoinsPRV), len(currentOutputCoinsToken), time.Since(tmpStart).Seconds())

		allOutputCoins = append(allOutputCoins, append(currentOutputCoinsToken, currentOutputCoinsPRV...)...)
		height = nextHeight
	}

	// write
	err := rawdbv2.StoreIndexedOTAKey(ci.db, vkb[:])
	if err == nil {
		err = ci.StoreIndexedOutputCoins(idxParams.OTAKey, allOutputCoins, idxParams.ShardID)
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins error: %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}
	} else {
		utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey error: %v\n", err)

		status.err = err
		ci.statusChan <- status
		return
	}

	ci.ManagedOTAKeys.Store(vkb, 2)
	utils.Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x, timeElapsed: %v\n", vkb, time.Since(start).Seconds())

	status.err = nil
	ci.statusChan <- status
}

// ReIndexOutCoinBatch re-scans all output coins for a list of indexing params of the same shardID.
//
// Callers must manage to make sure all indexing params belong to the same shard.
func (ci *CoinIndexer) ReIndexOutCoinBatch(idxParams []IndexParam, txDb *statedb.StateDB, id string) {
	if len(idxParams) == 0 {
		return
	}

	// create some map instances and necessary params
	mapIdxParams := make(map[string]IndexParam)
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
		keyExists, processing := ci.HasOTAKey(vkb)
		if keyExists {
			if processing == 1 {
				utils.Logger.Log.Errorf("[CoinIndexer] ota key %x is being processed", idxParam.OTAKey)
				ci.statusChan <- mapStatuses[otaStr]
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
			// resetting entries for this key is reserved for debugging RPCs
			if processing == 3 && !idxParam.IsReset {
				utils.Logger.Log.Errorf("[CoinIndexer] ota key %v has been processed with status %v and isReset = false", idxParam.OTAKey, processing)

				ci.statusChan <- mapStatuses[otaStr]
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
		}
		ci.ManagedOTAKeys.Store(vkb, 1)

		defer func() {
			if r := recover(); r != nil {
				utils.Logger.Log.Errorf("[CoinIndexer] Recovered from: %v\n", r)
			}
			if exists, tmpProcessing := ci.HasOTAKey(vkb); exists && tmpProcessing == 1 {
				ci.ManagedOTAKeys.Delete(vkb)
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

	// Clone the current cachedCoinPubKeys to avoid collisions
	cachedCoins := ci.cloneCachedCoins()
	utils.Logger.Log.Infof("len(clonedCachedCoins) = %v\n", len(cachedCoins))

	start := time.Now()
	for height := minHeight; height <= maxHeight; {
		tmpStart := time.Now() // measure time for each round
		nextHeight := height + utils.MaxOutcoinQueryInterval

		// query token output coins
		currentOutputCoinsToken, err := QueryBatchDbCoinVer2(mapIdxParams, shardID, &common.ConfidentialAssetID, height, nextHeight-1, txDb, cachedCoins, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying token coins from db - %v\n", err)

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
		currentOutputCoinsPRV, err := QueryBatchDbCoinVer2(mapIdxParams, shardID, &common.PRVCoinID, height, nextHeight-1, txDb, cachedCoins, getCoinFilterByOTAKey())
		if err != nil {
			utils.Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v\n", err)

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

			utils.Logger.Log.Infof("[CoinIndexer] Key %v, %d to %d: found %d PRV + %d pToken coins, current #coins %v, timeElapsed %v\n", otaStr, height, nextHeight-1, len(currentOutputCoinsPRV[otaStr]), len(currentOutputCoinsToken[otaStr]), len(listOutputCoins), time.Since(tmpStart).Seconds())
			mapOutputCoins[otaStr] = listOutputCoins
		}

		height = nextHeight
	}

	// write
	for otaStr, idxParam := range mapIdxParams {
		vkb := OTAKeyToRaw(idxParam.OTAKey)
		allOutputCoins := mapOutputCoins[otaStr]
		err := rawdbv2.StoreIndexedOTAKey(ci.db, vkb[:])
		if err == nil {
			utils.Logger.Log.Infof("[CoinIndexer] About to store %v output coins for OTAKey %x\n", len(allOutputCoins), vkb)
			err = ci.StoreIndexedOutputCoins(idxParam.OTAKey, allOutputCoins, shardID)
			if err != nil {
				utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins for OTA key %x error: %v\n", vkb, err)

				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
				continue
			}
		} else {
			utils.Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey %x, error: %v\n", vkb, err)

			status := mapStatuses[otaStr]
			status.err = err
			ci.statusChan <- status
			delete(mapIdxParams, otaStr)
			delete(mapStatuses, otaStr)
			delete(mapOutputCoins, otaStr)
			continue
		}

		ci.ManagedOTAKeys.Store(vkb, 3)
		utils.Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x, found %v coins, timeElapsed: %v\n", vkb, len(allOutputCoins), time.Since(start).Seconds())

		ci.statusChan <- mapStatuses[otaStr]
		delete(mapIdxParams, otaStr)
		delete(mapStatuses, otaStr)
		delete(mapOutputCoins, otaStr)
	}
}

func (ci *CoinIndexer) GetIndexedOutCoin(otaKey privacy.OTAKey, tokenID *common.Hash, txDb *statedb.StateDB, shardID byte) ([]privacy.Coin, int, error) {
	vkb := OTAKeyToRaw(otaKey)
	utils.Logger.Log.Infof("Retrieve re-indexed coins for %x from db %v", vkb, ci.db)
	_, processing := ci.HasOTAKey(vkb)
	if processing == 1 {
		return nil, 1, fmt.Errorf("OTA Key %x not ready : Sync still in progress", otaKey)
	}
	if processing == 0 {
		// this is a new view key
		return nil, 0, fmt.Errorf("OTA Key %x not synced", otaKey)
	}
	ocBytes, err := rawdbv2.GetOutCoinsByIndexedOTAKey(ci.db, common.ConfidentialAssetID, shardID, vkb[:])
	if err != nil {
		return nil, 0, err
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
			return nil, 0, fmt.Errorf("coin by OTAKey storage is corrupted")
		}
		if filter(temp, params) {
			// eliminate forked coins
			if dbHasOta, _, err := statedb.HasOnetimeAddress(txDb, *tokenID, temp.GetPublicKey().ToBytesS()); dbHasOta && err == nil {
				result = append(result, temp)
			}
		}
	}
	return result, 2, nil
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

// Start starts the CoinIndexer in case the authorized cache is employed.
// It is a hub to
//	- record key submission from users;
//	- record the indexing status of keys;
//	- collect keys into batches and index them all together in a batching way.
func (ci *CoinIndexer) Start() {
	ci.mtx.Lock()
	if ci.isAuthorizedRunning {
		ci.mtx.Unlock()
		return
	}

	ci.IdxChan = make(chan IndexParam, 2*ci.numWorkers)
	ci.statusChan = make(chan JobStatus, 2*ci.numWorkers)
	ci.quitChan = make(chan bool)
	ci.idxQueue = make(map[byte][]IndexParam)
	for shardID := 0; shardID < common.MaxShardNumber; shardID++ {
		ci.idxQueue[byte(shardID)] = make([]IndexParam, 0)
	}

	// A map to keep track of the number of IdxParam's per go-routine
	tracking := make(map[string]int)
	var id string

	utils.Logger.Log.Infof("Start CoinIndexer....\n")
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
				utils.Logger.Log.Errorf("IndexOutCoin for otaKey %x failed: %v\n", otaKeyBytes, status.err)
				err = ci.RemoveOTAKey(status.otaKey)
				if err != nil {
					utils.Logger.Log.Errorf("Remove OTAKey %v error: %v\n", otaKeyBytes, err)
				}
			} else {
				utils.Logger.Log.Infof("Finished indexing output coins for otaKey: %x\n", otaKeyBytes)
			}

		case idxParams := <-ci.IdxChan:
			otaKeyBytes := OTAKeyToRaw(idxParams.OTAKey)
			utils.Logger.Log.Infof("New authorized OTAKey received: %x\n", otaKeyBytes)

			ci.mtx.Lock()
			ci.idxQueue[idxParams.ShardID] = append(ci.idxQueue[idxParams.ShardID], idxParams)
			ci.queueSize++
			ci.mtx.Unlock()

		case <-ci.quitChan:
			ci.mtx.Lock()
			ci.isAuthorizedRunning = false
			ci.mtx.Unlock()

			utils.Logger.Log.Infof("Stopped coinIndexer!!\n")
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

						utils.Logger.Log.Infof("Re-index for %v new OTA keys, shard %v, id %v\n", numParams, shard, id)

						// increase 1 go-routine

						numWorking += 1
						go ci.ReIndexOutCoinBatch(idxParams, idxParams[0].TxDb, id)

						ci.mtx.Unlock()
					}
				}
				start = time.Now()
			} else {
				utils.Logger.Log.Infof("CoinIndexer is full or no OTA key is found in queue, numWorking %v, queueSize %v\n", numWorking, ci.queueSize)
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func (ci *CoinIndexer) Stop() {
	if ci.isAuthorizedRunning {
		ci.quitChan <- true
	}
}
