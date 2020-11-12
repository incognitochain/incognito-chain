package blockchain

 import (
 	"errors"
 	"fmt"
 	"os"
 	"time"
 	"context"

 	"golang.org/x/sync/semaphore"
 	"github.com/incognitochain/incognito-chain/common"
 	"github.com/incognitochain/incognito-chain/incognitokey"
 	"github.com/incognitochain/incognito-chain/privacy"
 	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

 )

type coinReindexer struct{
	sem 					*semaphore.Weighted
	ManagedOTAKeys	map[[64]byte]int
}
func NewOutcoinReindexer(numWorkers int64) *coinReindexer{
	sem := semaphore.NewWeighted(numWorkers)
	// view key :-> indexing status
	// 2 means indexer finished
	// while <2 : `balance` & `createTx` RPCs are not available
	// viewKey map will be loaded from db later
	return &coinReindexer{sem:sem, ManagedOTAKeys: nil}
}

func getCoinFilterByOTAKey() coinMatcher{
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

func (ci *coinReindexer) ReindexOutcoin(toHeight uint64, vk privacy.OTAKey, db *statedb.StateDB, shardID byte) error{
	keyMap := ci.getOrLoadIndexedOTAKeys(db)
	vkb := toRawViewKey(vk)
	keyMap[vkb] = 1
	defer func(){
		if keyMap[vkb]==1{
			delete(keyMap, vkb)
		}
	}()
	var allOutputCoins []privacy.Coin
	// read
	for ; toHeight > 0;{
		fromHeight := getLowerHeight(toHeight)
		ctx, cancel := context.WithTimeout(context.Background(), OutcoinReindexerTimeout * time.Second)
		defer cancel()
		err := ci.sem.Acquire(ctx, 1)
		if err!=nil{
			return err
		}
		currentOutputCoinsToken, err1 := queryDbCoinVer2(vk, shardID, &common.ConfidentialAssetID, fromHeight, toHeight, db, getCoinFilterByOTAKey())
		currentOutputCoinsPRV, err2 := queryDbCoinVer2(vk, shardID, &common.PRVCoinID, fromHeight, toHeight, db, getCoinFilterByOTAKey())
		ci.sem.Release(1)
		if err1!=nil || err2!=nil{
			return errors.New(fmt.Sprintf("Error while querying coins from db - %v - %v", err1, err2))
		}
		fmt.Fprintf(os.Stderr, "%d to %d : found %d + %d coins\n", fromHeight, toHeight, len(currentOutputCoinsPRV), len(currentOutputCoinsToken))
		// yield
		time.Sleep(50 * time.Millisecond)
		allOutputCoins = append(allOutputCoins, append(currentOutputCoinsToken, currentOutputCoinsPRV...)...)
		toHeight = fromHeight
	}
	// write
	ctx, cancel := context.WithTimeout(context.Background(), OutcoinReindexerTimeout * time.Second)
	defer cancel()
	err := ci.sem.Acquire(ctx, 1)
	if err!=nil{
		return err
	}
	err = ci.StoreReindexedOutputCoins(db, vk, allOutputCoins, shardID)
	if err==nil{
		err = statedb.StoreReindexedOTAkey(db, vkb[:])
		if err==nil{
			db.Commit(false)
		}
	}
	ci.sem.Release(1)
	if err!=nil{
		return err
	}
	
	keyMap[vkb] = 2
	return nil
}

func (ci *coinReindexer) GetReindexedOutcoin(viewKey privacy.OTAKey, tokenID *common.Hash, db *statedb.StateDB, shardID byte) ([]privacy.Coin, int, error){
	keyMap := ci.getOrLoadIndexedOTAKeys(db)
	vkb := toRawViewKey(viewKey)
	processing := keyMap[vkb]
	if processing==1{
		return nil, 1, errors.New(fmt.Sprintf("View Key %x not ready : Sync still in progress", viewKey))
	}
	if processing==0{
		// this is a new view key
		return nil, 0, nil
	}
	ocBytes, err := statedb.GetOutcoinsByReindexedOTAKey(db, common.ConfidentialAssetID, vkb[:], shardID)
	if err!=nil{
		return nil, 0, err
	}
	params := make(map[string]interface{})
	params["otaKey"] = viewKey
	params["tokenID"] = tokenID
	filter := getCoinFilterByOTAKeyAndToken()
	var result []privacy.Coin
	for _, cb := range ocBytes{
		temp := &privacy.CoinV2{}
		err := temp.SetBytes(cb)
		if err!=nil{
			return nil, 0, errors.New("Coin by View Key storage is corrupted")
		}
		if filter(temp, params){
			result = append(result, temp)
		}
	}
	return result, 2, nil
}

func (ci coinReindexer) StoreReindexedOutputCoins(db *statedb.StateDB, viewKey privacy.OTAKey, outputCoins []privacy.Coin, shardID byte) error{
	// has a default element to check for existence
	var ocBytes [][]byte
	for _, c := range outputCoins{
		ocBytes = append(ocBytes, c.Bytes())
	}
	vkb := toRawViewKey(viewKey)
	fmt.Fprintf(os.Stderr, "Storing %d indexed coins\n", len(ocBytes))
	// all token and PRV coins are grouped together; match them to desired tokenID upon retrieval
	return statedb.StoreReindexedOutputCoins(db, common.ConfidentialAssetID, vkb[:], ocBytes, shardID)
}

func (ci *coinReindexer) getOrLoadIndexedOTAKeys(db *statedb.StateDB) map[[64]byte]int{
	if ci.ManagedOTAKeys==nil{
		// load from db once after startup
		ci.ManagedOTAKeys = make(map[[64]byte]int)
		loadedKeysRaw, err := statedb.GetReindexedOTAKeys(db)
		if err==nil{	
			for _, b := range loadedKeysRaw{
				var temp [64]byte
				copy(temp[:], b[0:64])
				ci.ManagedOTAKeys[temp] = 2
			}
		}
	}
	return ci.ManagedOTAKeys
}

func getLowerHeight(upper uint64) uint64{
	if upper > MaxOutcoinQueryInterval{
		return upper - MaxOutcoinQueryInterval
	}
	return 0
}

// avoid overlap with version 1 coin entries
func toRawViewKey(vk privacy.OTAKey) [64]byte{
	var result [64]byte
	copy(result[0:32], vk.GetOTASecretKey().ToBytesS())
	copy(result[32:64], vk.GetPublicSpend().ToBytesS())
	return result
}

