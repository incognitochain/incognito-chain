package utils

import (
    "errors"
    "fmt"
    "time"
    "context"
    "sync"

    "golang.org/x/sync/semaphore"
    "github.com/incognitochain/incognito-chain/common"
    "github.com/incognitochain/incognito-chain/incdb"
    "github.com/incognitochain/incognito-chain/incognitokey"
    "github.com/incognitochain/incognito-chain/privacy"
    "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
    "github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type CoinMatcher func(*privacy.CoinV2, map[string]interface{}) bool

type CoinReindexer struct{
    sem             *semaphore.Weighted
    ManagedOTAKeys  *sync.Map
    db              incdb.Database
}
func NewOutcoinReindexer(numWorkers int64, db incdb.Database) (*CoinReindexer, error){
    sem := semaphore.NewWeighted(numWorkers)
    // view key :-> indexing status
    // 2 means indexer finished
    // while <2 : `balance` & `createTx` RPCs are not available
    // viewKey map will be loaded from db later

    m  := &sync.Map{}
    // load from db once after startup
    loadedKeysRaw, err := rawdbv2.GetReindexedOTAkeys(db)
    if err==nil{    
        for _, b := range loadedKeysRaw{
            var temp [64]byte
            copy(temp[:], b[0:64])
            m.Store(temp, 2)
        }
    }
    return &CoinReindexer{sem:sem, ManagedOTAKeys: m, db: db}, nil
}

func getCoinFilterByOTAKey() CoinMatcher{
    return func(c *privacy.CoinV2, kvargs map[string]interface{}) bool{
        entry, exists := kvargs["otaKey"]
        if !exists{
            return false
        }
        vk, ok := entry.(privacy.OTAKey)
        if !ok{
            return false
        }
        ks := &incognitokey.KeySet{}
        ks.OTAKey = vk

        pass, _ := c.DoesCoinBelongToKeySet(ks)
        return pass
    }
}

func GetCoinFilterByOTAKeyAndToken() CoinMatcher{
	return func(c *privacy.CoinV2, kvargs map[string]interface{}) bool{
		entry, exists := kvargs["otaKey"]
		if !exists{
			return false
		}
		vk, ok := entry.(privacy.OTAKey)
		if !ok{
			return false
		}
		entry, exists = kvargs["tokenID"]
		if !exists{
			return false
		}
		tokenID, ok := entry.(*common.Hash)
		if !ok{
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

func (ci *CoinReindexer) ReindexOutcoin(toHeight uint64, vk privacy.OTAKey, txdb *statedb.StateDB, shardID byte) error{
    vkb := OTAKeyToRaw(vk)
    Logger.Log.Infof("Re-index output coins for key %x", vkb)
    keyExists, processing := ci.HasOTAKey(vkb)
    if keyExists && (processing==1 || processing==2){
        return nil
    }
    ci.ManagedOTAKeys.Store(vkb, 1)
    defer func(){
        if exists, processing := ci.HasOTAKey(vkb); exists && processing==1{
            ci.ManagedOTAKeys.Delete(vkb)
        }
    }()
    var allOutputCoins []privacy.Coin
    // read
    // if err!=nil{
    //  return err
    // }
    for height:=uint64(0); height <= toHeight;{
        nextHeight := height + MaxOutcoinQueryInterval
        
        ctx, cancel := context.WithTimeout(context.Background(), OutcoinReindexerTimeout * time.Second)
        defer cancel()
        err := ci.sem.Acquire(ctx, 1)
        if err!=nil{
            return err
        }
        currentOutputCoinsToken, err1 := QueryDbCoinVer2(vk, shardID, &common.ConfidentialAssetID, height, nextHeight-1, txdb, getCoinFilterByOTAKey())
        currentOutputCoinsPRV, err2 := QueryDbCoinVer2(vk, shardID, &common.PRVCoinID, height, nextHeight-1 , txdb, getCoinFilterByOTAKey())
        ci.sem.Release(1)
        if err1!=nil || err2!=nil{
            return errors.New(fmt.Sprintf("Error while querying coins from db - %v - %v", err1, err2))
        }
        Logger.Log.Infof("Key %x - %d to %d : found %d PRV + %d pToken coins", vkb, height, nextHeight-1, len(currentOutputCoinsPRV), len(currentOutputCoinsToken))
        
        allOutputCoins = append(allOutputCoins, append(currentOutputCoinsToken, currentOutputCoinsPRV...)...)
        height = nextHeight
    }
    // write
    err := rawdbv2.StoreReindexedOTAkey(ci.db, vkb[:])
    if err==nil{
        err = ci.StoreReindexedOutputCoins(vk, allOutputCoins, shardID)
    }
    ci.ManagedOTAKeys.Store(vkb, 2)
    Logger.Log.Infof("Indexing complete for key %x", vkb)
    return nil
}

func (ci *CoinReindexer) GetReindexedOutcoin(viewKey privacy.OTAKey, tokenID *common.Hash, txdb *statedb.StateDB, shardID byte) ([]privacy.Coin, int, error){
    // keyMap := ci.getOrLoadIndexedOTAKeys(db)
    vkb := OTAKeyToRaw(viewKey)
    Logger.Log.Infof("Retrieve re-indexed coins for %x from db %v", vkb, ci.db)
    _,  processing := ci.HasOTAKey(vkb)
    if processing==1{
        return nil, 1, errors.New(fmt.Sprintf("View Key %x not ready : Sync still in progress", viewKey))
    }
    if processing==0{
        // this is a new view key
        return nil, 0, errors.New(fmt.Sprintf("View Key %x not synced", viewKey))
    }
    ocBytes, err := rawdbv2.GetOutcoinsByReindexedOTAKey(ci.db, common.ConfidentialAssetID, shardID, vkb[:])
    if err!=nil{
        return nil, 0, err
    }
    params := make(map[string]interface{})
    params["otaKey"] = viewKey
    params["tokenID"] = tokenID
    filter := GetCoinFilterByOTAKeyAndToken()
    var result []privacy.Coin
    for _, cb := range ocBytes{
        temp := &privacy.CoinV2{}
        err := temp.SetBytes(cb)
        if err!=nil{
            return nil, 0, errors.New("Coin by View Key storage is corrupted")
        }
        if filter(temp, params){
            // eliminate forked coins
            if dbHasOta, err := statedb.HasOnetimeAddress(txdb, *tokenID, temp.GetPublicKey().ToBytesS()); dbHasOta && err==nil{
                result = append(result, temp)
            }
        }
    }
    return result, 2, nil
}

func (ci *CoinReindexer) StoreReindexedOutputCoins(viewKey privacy.OTAKey, outputCoins []privacy.Coin, shardID byte) error{
    var ocBytes [][]byte
    for _, c := range outputCoins{
        ocBytes = append(ocBytes, c.Bytes())
    }
    vkb := OTAKeyToRaw(viewKey)
    Logger.Log.Infof("Store %d indexed coins to db %v", len(ocBytes), ci.db)
    // all token and PRV coins are grouped together; match them to desired tokenID upon retrieval
    ctx, cancel := context.WithTimeout(context.Background(), OutcoinReindexerTimeout * time.Second)
    defer cancel()
    err := ci.sem.Acquire(ctx, 1)
    if err!=nil{
        return err
    }
    err = rawdbv2.StoreReindexedOutputCoins(ci.db, common.ConfidentialAssetID, vkb[:], ocBytes, shardID)
    ci.sem.Release(1)
    if err!=nil{
        return err
    }
    return err
}

func GetNextLowerHeight(upper uint64) uint64{
    if upper > MaxOutcoinQueryInterval{
        return upper - MaxOutcoinQueryInterval
    }
    return 0
}

func OTAKeyToRaw(vk privacy.OTAKey) [64]byte{
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

func (ci *CoinReindexer) HasOTAKey(k [64]byte) (bool, int){
    var result int
    val, ok := ci.ManagedOTAKeys.Load(k)
    if ok{
        result, ok = val.(int)
    }
    return ok, result
}


func QueryDbCoinVer1(pubkey []byte, shardID byte, tokenID *common.Hash, db *statedb.StateDB) ([]privacy.Coin, error) {
	outCoinsBytes, err := statedb.GetOutcoinsByPubkey(db, *tokenID, pubkey, shardID)
	if err != nil {
		Logger.Log.Error("GetOutcoinsBytesByKeyset Get by PubKey", err)
		return nil, err
	}
	var outCoins []privacy.Coin
	for _, item := range outCoinsBytes{
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
	if shardHeight == 0{
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
			if err!=nil{
				Logger.Log.Error("Get outcoins ver 2 from bytes", err)
				return nil, err
			}
			pass := true
			for _, f := range filters{
				if !f(cv2, params){
					pass = false
				}
			}
			if pass{
				outCoins = append(outCoins, cv2)
			}
		}
	}
	return outCoins, nil
}