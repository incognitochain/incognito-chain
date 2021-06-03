package utils

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/semaphore"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type JobStatus struct {
	otaKey privacy.OTAKey
	err    error
}

type CoinMatcher func(*privacy.CoinV2, map[string]interface{}) bool

type CoinIndexer struct {
	sem            *semaphore.Weighted
	numWorkers     int
	ManagedOTAKeys *sync.Map
	db             incdb.Database
	accessTokens   map[string][]privacy.OTAKey
	otaQueue       []privacy.OTAKey
	OTAChan        chan privacy.OTAKey
	statusChan     chan JobStatus
}

// NewOutCoinIndexer creates a new full node's caching instance for faster output coin retrieval.
func NewOutCoinIndexer(numWorkers int64, db incdb.Database) (*CoinIndexer, error) {
	// view key :-> indexing status
	// 2 means indexer finished
	// while < 2 : `balance` & `createTx` RPCs are not available
	// viewKey map will be loaded from db
	sem := semaphore.NewWeighted(numWorkers)

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
	return &CoinIndexer{sem: sem, ManagedOTAKeys: m, db: db, accessTokens: nil}, nil
}

// IsAuthorizedUser checks if a user is authorized to use the enhanced cache.
func (ci *CoinIndexer) IsAuthorizedUser(accessToken string, otaKey privacy.OTAKey) bool {
	if otaList, ok := ci.accessTokens[accessToken]; ok { // valid access token
		if otaList == nil {
			otaList = make([]privacy.OTAKey, 0)
		}

		otaList = append(otaList, otaKey)

		return true
	}

	return false
}

// RemoveOTAKey removes an OTAKey from the cached database.
//
// TODO: remove cached output coins, access token.
func (ci *CoinIndexer) RemoveOTAKey(accessToken string, otaKey privacy.OTAKey) error {
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

// ReIndexOutCoin re-scans all output coins from fromHeight to toHeight and adds them to the cache if they belong to vk.
func (ci *CoinIndexer) ReIndexOutCoin(fromHeight, toHeight uint64, vk privacy.OTAKey, txDb *statedb.StateDB, shardID byte, isReset bool) error {
	vkb := OTAKeyToRaw(vk)
	Logger.Log.Infof("[CoinIndexer] Re-index output coins for key %x", vkb)
	keyExists, processing := ci.HasOTAKey(vkb)
	if keyExists {
		if processing == 1 {
			return nil
		}
		// resetting entries for this key is reserved for debugging RPCs
		if processing == 2 && !isReset {
			return nil
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

	for height := fromHeight; height <= toHeight; {
		nextHeight := height + MaxOutcoinQueryInterval

		ctx, cancel := context.WithTimeout(context.Background(), OutcoinReindexerTimeout*time.Second)
		defer cancel()

		err := ci.sem.Acquire(ctx, 1)
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] semaphore acquiring error: %v\n", err)
			return err
		}

		// query token output coins
		currentOutputCoinsToken, err := QueryDbCoinVer2(vk, shardID, &common.ConfidentialAssetID, height, nextHeight-1, txDb, getCoinFilterByOTAKey())
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] Error while querying token coins from db - %v\n", err)
			return errors.New(fmt.Sprintf("Error while querying token coins from db - %v", err))
		}

		// query PRV output coins
		currentOutputCoinsPRV, err := QueryDbCoinVer2(vk, shardID, &common.PRVCoinID, height, nextHeight-1, txDb, getCoinFilterByOTAKey())
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v\n", err)
			return errors.New(fmt.Sprintf("Error while querying PRV coins from db - %v", err))
		}

		ci.sem.Release(1)

		Logger.Log.Infof("[CoinIndexer] Key %x - %d to %d : found %d PRV + %d pToken coins\n", vkb, height, nextHeight-1, len(currentOutputCoinsPRV), len(currentOutputCoinsToken))

		allOutputCoins = append(allOutputCoins, append(currentOutputCoinsToken, currentOutputCoinsPRV...)...)
		height = nextHeight
	}

	// write
	err := rawdbv2.StoreIndexedOTAKey(ci.db, vkb[:])
	if err == nil {
		err = ci.StoreIndexedOutputCoins(vk, allOutputCoins, shardID)
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins error: %v\n", err)
			return err
		}
	} else {
		Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey error: %v\n", err)
		return err
	}

	ci.ManagedOTAKeys.Store(vkb, 2)
	Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x\n", vkb)
	return nil
}

func (ci *CoinIndexer) GetIndexedOutCoin(viewKey privacy.OTAKey, tokenID *common.Hash, txDb *statedb.StateDB, shardID byte) ([]privacy.Coin, int, error) {
	vkb := OTAKeyToRaw(viewKey)
	Logger.Log.Infof("Retrieve re-indexed coins for %x from db %v", vkb, ci.db)
	_, processing := ci.HasOTAKey(vkb)
	if processing == 1 {
		return nil, 1, errors.New(fmt.Sprintf("View Key %x not ready : Sync still in progress", viewKey))
	}
	if processing == 0 {
		// this is a new view key
		return nil, 0, errors.New(fmt.Sprintf("View Key %x not synced", viewKey))
	}
	ocBytes, err := rawdbv2.GetOutCoinsByIndexedOTAKey(ci.db, common.ConfidentialAssetID, shardID, vkb[:])
	if err != nil {
		return nil, 0, err
	}
	params := make(map[string]interface{})
	params["otaKey"] = viewKey
	params["tokenID"] = tokenID
	filter := GetCoinFilterByOTAKeyAndToken()
	var result []privacy.Coin
	for _, cb := range ocBytes {
		temp := &privacy.CoinV2{}
		err := temp.SetBytes(cb)
		if err != nil {
			return nil, 0, errors.New("Coin by View Key storage is corrupted")
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

func (ci *CoinIndexer) StoreIndexedOutputCoins(viewKey privacy.OTAKey, outputCoins []privacy.Coin, shardID byte) error {
	var ocBytes [][]byte
	for _, c := range outputCoins {
		ocBytes = append(ocBytes, c.Bytes())
	}
	vkb := OTAKeyToRaw(viewKey)
	Logger.Log.Infof("Store %d indexed coins to db %v", len(ocBytes), ci.db)
	// all token and PRV coins are grouped together; match them to desired tokenID upon retrieval
	err := rawdbv2.StoreIndexedOutCoins(ci.db, common.ConfidentialAssetID, vkb[:], ocBytes, shardID)
	if err != nil {
		return err
	}
	return err
}

func (ci *CoinIndexer) Serve() {

	numWorking := 0
	for {
		select {
		case status := <-ci.statusChan:
			numWorking--
			if status.err != nil {
				ci.RemoveOTAKey("", status.otaKey)
			} else {
				Logger.Log.Infof("Finished indexing output coins for otaKey: %v\n", status.otaKey)
			}
		case otaKey := <-ci.OTAChan:
			Logger.Log.Infof("New authorized OTAKey received: %v\n", otaKey)
			//do something
			ci.otaQueue = append(ci.otaQueue, otaKey)
		default:
			if numWorking < ci.numWorkers {
				numWorking++
				//do something
			}
		}
	}
}

func GetNextLowerHeight(upper, floor uint64) uint64 {
	if upper > MaxOutcoinQueryInterval+floor {
		return upper - MaxOutcoinQueryInterval
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

func (ci *CoinIndexer) HasOTAKey(k [64]byte) (bool, int) {
	var result int
	val, ok := ci.ManagedOTAKeys.Load(k)
	if ok {
		result, ok = val.(int)
	}
	return ok, result
}

func QueryDbCoinVer1(pubKey []byte, shardID byte, tokenID *common.Hash, db *statedb.StateDB) ([]privacy.Coin, error) {
	outCoinsBytes, err := statedb.GetOutcoinsByPubkey(db, *tokenID, pubKey, shardID)
	if err != nil {
		Logger.Log.Error("GetOutcoinsBytesByKeyset Get by PubKey", err)
		return nil, err
	}
	var outCoins []privacy.Coin
	for _, item := range outCoinsBytes {
		outCoin := &privacy.CoinV1{}
		err := outCoin.SetBytes(item)
		if err != nil {
			Logger.Log.Errorf("Cannot create coin from byte %v", err)
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
	for height := start; height <= destHeight; height += 1 {
		currentHeightCoins, err := statedb.GetOTACoinsByHeight(db, *tokenID, shardID, height)
		if err != nil {
			Logger.Log.Error("Get outcoins ver 2 bytes by keyset get by height", err)
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
				Logger.Log.Error("Get outcoins ver 2 from bytes", err)
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
