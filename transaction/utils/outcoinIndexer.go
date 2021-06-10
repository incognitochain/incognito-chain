package utils

import (
	"context"
	"encoding/hex"
	"fmt"
	"golang.org/x/sync/semaphore"
	"math"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/privacy"
)

var usedSig map[string]bool

type CoinIndexer struct {
	numWorkers          int
	sem                 *semaphore.Weighted
	mtx                 *sync.RWMutex
	db                  incdb.Database
	accessTokens        map[string]bool
	idxQueue            map[byte][]IndexParam
	queueSize           int
	statusChan          chan JobStatus
	quitChan            chan bool
	isAuthorizedRunning bool
	cachedCoins         map[string]interface{}

	ManagedOTAKeys *sync.Map
	IdxChan        chan IndexParam
}

// NewOutCoinIndexer creates a new full node's caching instance for faster output coin retrieval.
func NewOutCoinIndexer(numWorkers int64, db incdb.Database) (*CoinIndexer, error) {
	// view key :-> indexing status
	// 2 means indexer finished
	// while < 2 : `balance` & `createTx` RPCs are not available
	// viewKey map will be loaded from db
	usedSig = make(map[string]bool)

	var sem *semaphore.Weighted
	accessTokens := make(map[string]bool)
	if numWorkers != 0 {
		sem = semaphore.NewWeighted(numWorkers)
		accessTokens[DefaultAccessToken] = true
	}
	Logger.Log.Infof("NewOutCoinIndexer with %v workers\n", numWorkers)

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
	Logger.Log.Infof("Number of cached OTA keys: %v\n", len(loadedKeysRaw))

	cachedCoins := make(map[string]interface{})
	loadRawCachedCoinHashes, err := rawdbv2.GetCachedCoinHashes(db)
	if err == nil {
		for _, coinHash := range loadRawCachedCoinHashes {
			var temp [32]byte
			copy(temp[:], coinHash[:32])
			cachedCoins[fmt.Sprintf("%x", temp)] = true
		}
	}

	Logger.Log.Infof("Number of cached coins: %v\n", len(cachedCoins))

	ci := &CoinIndexer{
		numWorkers:          int(numWorkers),
		sem:                 sem,
		mtx:                 mtx,
		ManagedOTAKeys:      m,
		db:                  db,
		accessTokens:        accessTokens,
		cachedCoins:         cachedCoins,
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

//// IsValidAdminSignature checks if data is signed by a valid admin privateKey.
//func (ci *CoinIndexer) IsValidAdminSignature(data []byte, sig []byte) (bool, error) {
//	hashedSig := fmt.Sprintf("%x", sha3.Sum256(sig))
//	if usedSig[hashedSig] {
//		return false, fmt.Errorf("cannot relay an already used signature")
//	} else {
//		usedSig[hashedSig] = true
//	}
//
//	if len(sig) != 64 {
//		return false, fmt.Errorf("signature length invalid: %v", len(sig))
//	}
//
//	r := new(big.Int).SetBytes(sig[:32])
//	s := new(big.Int).SetBytes(sig[32:])
//
//	hashedMsg := sha3.Sum256(data)
//
//	isValid := ecdsa.Verify(ci.idxPublicKey, hashedMsg[:], r, s)
//
//	return isValid, nil
//}
//
//// AddAccessToken adds an access token to the list.
//func (ci *CoinIndexer) AddAccessToken(accessToken string, signature []byte) error {
//	isValid, err := ci.IsValidAdminSignature([]byte(accessToken), signature)
//	if err != nil || !isValid {
//		Logger.Log.Errorf("Cannot add access token %v, sig %v: isValid = %v, err = %v\n", accessToken, signature, isValid, err)
//		return fmt.Errorf("cannot add access token")
//	}
//
//	accessTokenBytes, err := hex.DecodeString(accessToken)
//	if err != nil {
//		return err
//	}
//	if len(accessTokenBytes) != 32 {
//		return fmt.Errorf("access token is invalid")
//	}
//
//	ci.accessTokens[accessToken] = true
//
//	return nil
//}

//// DeleteAccessToken deletes an access token from the list.
//func (ci *CoinIndexer) DeleteAccessToken(accessToken string, signature []byte) error {
//	//isValid, err := ci.IsValidAdminSignature([]byte(accessToken), signature)
//	//if err != nil || !isValid {
//	//	Logger.Log.Errorf("Cannot delete access token %v, sig %v: isValid = %v, err = %v\n", accessToken, signature, isValid, err)
//	//	return fmt.Errorf("cannot delete access token")
//	//}
//
//	if ci.accessTokens[accessToken] {
//		delete(ci.accessTokens, accessToken)
//	}
//
//	return nil
//}

// RemoveOTAKey removes an OTAKey from the cached database.
//
// TODO: remove cached output coins, access token.
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

func (ci *CoinIndexer) AddCoinHash(coinHash common.Hash) error {
	err := rawdbv2.StoreCachedCoinHash(ci.db, coinHash[:])
	if err != nil {
		return err
	}
	ci.cachedCoins[coinHash.String()] = true
	Logger.Log.Infof("Add coinHash %v success\n", coinHash, coinHash.String())
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
	Logger.Log.Infof("[CoinIndexer] Re-index output coins for key %x", idxParams.OTAKey)
	keyExists, processing := ci.HasOTAKey(vkb)
	if keyExists {
		if processing == 1 {
			Logger.Log.Errorf("[CoinIndexer] ota key %v is being processed", idxParams.OTAKey)
			ci.statusChan <- status
			return
		}
		// resetting entries for this key is reserved for debugging RPCs
		if processing == 2 && !idxParams.IsReset {
			Logger.Log.Errorf("[CoinIndexer] ota key %v has been processed and isReset = false", idxParams.OTAKey)

			ci.statusChan <- status
			return
		}
	}
	ci.ManagedOTAKeys.Store(vkb, 1)
	defer func() {
		if r := recover(); r != nil {
			Logger.Log.Errorf("[CoinIndexer] Recovered from: %v\n", r)
		}
		if exists, processing := ci.HasOTAKey(vkb); exists && processing == 1 {
			ci.ManagedOTAKeys.Delete(vkb)
		}
	}()
	var allOutputCoins []privacy.Coin

	start := time.Now()
	for height := idxParams.FromHeight; height <= idxParams.ToHeight; {
		tmpStart := time.Now()
		nextHeight := height + MaxOutcoinQueryInterval

		ctx, cancel := context.WithTimeout(context.Background(), OutcoinReindexerTimeout*time.Second)
		defer cancel()

		err := ci.sem.Acquire(ctx, 1)
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] semaphore acquiring error: %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		// query token output coins
		currentOutputCoinsToken, err := QueryDbCoinVer2(idxParams.OTAKey, idxParams.ShardID, &common.ConfidentialAssetID, height, nextHeight-1, idxParams.TxDb, getCoinFilterByOTAKey())
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] Error while querying token coins from db - %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		// query PRV output coins
		currentOutputCoinsPRV, err := QueryDbCoinVer2(idxParams.OTAKey, idxParams.ShardID, &common.PRVCoinID, height, nextHeight-1, idxParams.TxDb, getCoinFilterByOTAKey())
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		ci.sem.Release(1)

		Logger.Log.Infof("[CoinIndexer] Key %x, %d to %d: found %d PRV + %d pToken coins, timeElapsed %v\n", vkb, height, nextHeight-1, len(currentOutputCoinsPRV), len(currentOutputCoinsToken), time.Since(tmpStart).Seconds())

		allOutputCoins = append(allOutputCoins, append(currentOutputCoinsToken, currentOutputCoinsPRV...)...)
		height = nextHeight
	}

	// write
	err := rawdbv2.StoreIndexedOTAKey(ci.db, vkb[:])
	if err == nil {
		err = ci.StoreIndexedOutputCoins(idxParams.OTAKey, allOutputCoins, idxParams.ShardID)
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins error: %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}
	} else {
		Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey error: %v\n", err)

		status.err = err
		ci.statusChan <- status
		return
	}

	ci.ManagedOTAKeys.Store(vkb, 2)
	Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x, timeElapsed: %v\n", vkb, time.Since(start).Seconds())

	status.err = nil
	ci.statusChan <- status
	return
}

// ReIndexOutCoinBatch re-scans all output coins for a list of indexing params of the same shardID.
//
// Callers must manage to make sure all indexing params belong to the same shard.
func (ci *CoinIndexer) ReIndexOutCoinBatch(idxParams []IndexParam, txDb *statedb.StateDB) {
	if len(idxParams) == 0 {
		return
	}
	//create some map instances and necessary params
	mapIdxParams := make(map[string]IndexParam)
	mapStatuses := make(map[string]JobStatus)
	mapOutputCoins := make(map[string][]privacy.Coin)
	minHeight := uint64(math.MaxUint64)
	maxHeight := uint64(0)
	shardID := idxParams[0].ShardID
	for _, idxParam := range idxParams {
		otaStr := fmt.Sprintf("%x", OTAKeyToRaw(idxParam.OTAKey))
		mapIdxParams[otaStr] = idxParam
		mapStatuses[otaStr] = JobStatus{otaKey: idxParam.OTAKey, err: nil}
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
		Logger.Log.Infof("[CoinIndexer] Re-index output coins for key %x", idxParam.OTAKey)
		keyExists, processing := ci.HasOTAKey(vkb)
		if keyExists {
			if processing == 1 {
				Logger.Log.Errorf("[CoinIndexer] ota key %x is being processed", idxParam.OTAKey)
				ci.statusChan <- mapStatuses[otaStr]
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
			// resetting entries for this key is reserved for debugging RPCs
			if processing == 2 && !idxParam.IsReset {
				Logger.Log.Errorf("[CoinIndexer] ota key %v has been processed and isReset = false", idxParam.OTAKey)

				ci.statusChan <- mapStatuses[otaStr]
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
		}
		ci.ManagedOTAKeys.Store(vkb, 1)

		func() {
			if r := recover(); r != nil {
				Logger.Log.Errorf("[CoinIndexer] Recovered from: %v\n", r)
			}
			if exists, processing := ci.HasOTAKey(vkb); exists && processing == 1 {
				ci.ManagedOTAKeys.Delete(vkb)
			}
		}()
	}

	// in case minHeight > maxHeight, all indexing params will fail
	Logger.Log.Infof("fromHeight: %v, toHeight %v\n", minHeight, maxHeight)
	if minHeight > maxHeight {
		err := fmt.Errorf("minHeight (%v) > maxHeight (%v) when re-indexing outcoins", minHeight, maxHeight)
		for otaStr, _ := range mapStatuses {
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
		tmpStart := time.Now()
		nextHeight := height + MaxOutcoinQueryInterval

		ctx, cancel := context.WithTimeout(context.Background(), OutcoinReindexerTimeout*time.Second)
		defer cancel()

		err := ci.sem.Acquire(ctx, 1)
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] semaphore acquiring error: %v\n", err)

			for otaStr, _ := range mapStatuses {
				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
			return
		}

		// query token output coins
		currentOutputCoinsToken, err := QueryBatchDbCoinVer2(mapIdxParams, shardID, &common.ConfidentialAssetID, height, nextHeight-1, txDb, ci.cachedCoins, getCoinFilterByOTAKey())
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] Error while querying token coins from db - %v\n", err)

			for otaStr, _ := range mapStatuses {
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
		currentOutputCoinsPRV, err := QueryBatchDbCoinVer2(mapIdxParams, shardID, &common.PRVCoinID, height, nextHeight-1, txDb, ci.cachedCoins, getCoinFilterByOTAKey())
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v\n", err)

			for otaStr, _ := range mapStatuses {
				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
			}
			return
		}

		ci.sem.Release(1)

		//Add output coins to maps
		for otaStr, listOutputCoins := range mapOutputCoins {
			listOutputCoins = append(listOutputCoins, currentOutputCoinsToken[otaStr]...)
			listOutputCoins = append(listOutputCoins, currentOutputCoinsPRV[otaStr]...)

			Logger.Log.Infof("[CoinIndexer] Key %v, %d to %d: found %d PRV + %d pToken coins, current #coins %v, timeElapsed %v\n", otaStr, height, nextHeight-1, len(currentOutputCoinsPRV[otaStr]), len(currentOutputCoinsToken[otaStr]), len(listOutputCoins), time.Since(tmpStart).Seconds())
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
			Logger.Log.Infof("[CoinIndexer] About to store %v output coins for OTAKey %x\n", len(allOutputCoins), vkb)
			err = ci.StoreIndexedOutputCoins(idxParam.OTAKey, allOutputCoins, shardID)
			if err != nil {
				Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins for OTA key %x error: %v\n", vkb, err)

				status := mapStatuses[otaStr]
				status.err = err
				ci.statusChan <- status
				delete(mapIdxParams, otaStr)
				delete(mapStatuses, otaStr)
				delete(mapOutputCoins, otaStr)
				continue
			}
		} else {
			Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey %x, error: %v\n", vkb, err)

			status := mapStatuses[otaStr]
			status.err = err
			ci.statusChan <- status
			delete(mapIdxParams, otaStr)
			delete(mapStatuses, otaStr)
			delete(mapOutputCoins, otaStr)
			continue
		}

		ci.ManagedOTAKeys.Store(vkb, 2)
		Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x, found %v coins, timeElapsed: %v\n", vkb, len(allOutputCoins), time.Since(start).Seconds())

		ci.statusChan <- mapStatuses[otaStr]
		delete(mapIdxParams, otaStr)
		delete(mapStatuses, otaStr)
		delete(mapOutputCoins, otaStr)
	}

	return
}

func (ci *CoinIndexer) GetIndexedOutCoin(otaKey privacy.OTAKey, tokenID *common.Hash, txDb *statedb.StateDB, shardID byte) ([]privacy.Coin, int, error) {
	vkb := OTAKeyToRaw(otaKey)
	Logger.Log.Infof("Retrieve re-indexed coins for %x from db %v", vkb, ci.db)
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
	Logger.Log.Infof("Store %d indexed coins to db %x", len(ocBytes), vkb)
	// all token and PRV coins are grouped together; match them to desired tokenID upon retrieval
	err := rawdbv2.StoreIndexedOutCoins(ci.db, common.ConfidentialAssetID, vkb[:], ocBytes, shardID)
	if err != nil {
		return err
	}

	// cache the coin's hash to reduce the number of checks
	for _, c := range outputCoins {
		err = ci.AddCoinHash(common.HashH(c.Bytes()))
		if err != nil {
			return err
		}
	}

	return nil
}

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

	Logger.Log.Infof("Start CoinIndexer....\n")
	ci.isAuthorizedRunning = true
	ci.mtx.Unlock()

	var err error
	numWorking := 0
	start := time.Now()
	for {
		select {
		case status := <-ci.statusChan:
			numWorking--
			otaKeyBytes := OTAKeyToRaw(status.otaKey)
			if status.err != nil {
				Logger.Log.Errorf("IndexOutCoin for otaKey %x failed: %v\n", otaKeyBytes, status.err)
				err = ci.RemoveOTAKey(status.otaKey)
				if err != nil {
					Logger.Log.Errorf("Remove OTAKey %v error: %v\n", otaKeyBytes, err)
				}
			} else {
				Logger.Log.Infof("Finished indexing output coins for otaKey: %x\n", otaKeyBytes)
			}

		case idxParams := <-ci.IdxChan:
			otaKeyBytes := OTAKeyToRaw(idxParams.OTAKey)
			Logger.Log.Infof("New authorized OTAKey received: %x\n", otaKeyBytes)
			ci.mtx.Lock()
			ci.idxQueue[idxParams.ShardID] = append(ci.idxQueue[idxParams.ShardID], idxParams)
			ci.queueSize++
			ci.mtx.Unlock()

		case <-ci.quitChan:
			ci.mtx.Lock()
			ci.isAuthorizedRunning = false
			ci.mtx.Unlock()
			Logger.Log.Infof("Stopped coinIndexer!!\n")
			return
		default:
			if numWorking < ci.numWorkers && ci.queueSize > 0 {
				if time.Since(start).Seconds() < BatchWaitingTime { //collect indexing params by intervals to (possibly) reduce the number of go routines
					continue
				}
				remainingWorker := ci.numWorkers - numWorking
				workersForEach := ci.splitWorkers(remainingWorker)
				for shard, numWorkers := range workersForEach {
					if numWorkers != 0 {
						numWorking += numWorkers
						ci.queueSize -= numWorkers
						idxParams := ci.idxQueue[shard][:numWorkers]
						ci.idxQueue[shard] = ci.idxQueue[shard][numWorkers:]

						Logger.Log.Infof("Re-index for %v new OTA keys, shard %v\n", numWorkers, shard)
						go ci.ReIndexOutCoinBatch(idxParams, idxParams[0].TxDb)
					}
				}
				start = time.Now()
			} else {
				Logger.Log.Infof("CoinIndexer is full or no OTA key is found in queue, numWorking %v, queueSize %v\n", numWorking, ci.queueSize)
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
