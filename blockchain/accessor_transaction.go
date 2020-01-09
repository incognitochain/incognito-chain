package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"sort"
	"strconv"
	"strings"
	"time"
)

//GetListOutputCoinsByKeysetV2 - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key.
//With private-key, we can check unspent tx by check serialNumber from database
//- Param #1: keyset - (priv-key, payment-address, readonlykey)
//in case priv-key: return unspent outputcoin tx
//in case readonly-key: return all outputcoin tx with amount value
//in case payment-address: return all outputcoin tx with no amount value
//- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
func (blockchain *BlockChain) GetListOutputCoinsByKeysetV2(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash) ([]*privacy.OutputCoin, error) {
	blockchain.BestState.Shard[shardID].lock.Lock()
	defer blockchain.BestState.Shard[shardID].lock.Unlock()
	var outCointsInBytes [][]byte
	var err error
	transactionStateDB := blockchain.BestState.Shard[shardID].transactionStateDB
	if keyset == nil {
		return nil, NewBlockChainError(GetListOutputCoinsByKeysetError, fmt.Errorf("invalid key set, got keyset %+v", keyset))
	}
	if blockchain.config.MemCache != nil {
		// get from cache
		cachedKey := memcache.GetListOutputcoinCachedKey(keyset.PaymentAddress.Pk[:], tokenID, shardID)
		cachedData, _ := blockchain.config.MemCache.Get(cachedKey)
		if cachedData != nil && len(cachedData) > 0 {
			// try to parsing on outCointsInBytes
			_ = json.Unmarshal(cachedData, &outCointsInBytes)
		}
		if len(outCointsInBytes) == 0 {
			// cached data is nil or fail -> get from database
			outCointsInBytes, err = statedb.GetOutcoinsByPubkey(transactionStateDB, *tokenID, keyset.PaymentAddress.Pk[:], shardID)
			if len(outCointsInBytes) > 0 {
				// cache 1 day for result
				cachedData, err = json.Marshal(outCointsInBytes)
				if err == nil {
					blockchain.config.MemCache.PutExpired(cachedKey, cachedData, 1*24*60*60*time.Millisecond)
				}
			}
		}
	}
	if len(outCointsInBytes) == 0 {
		outCointsInBytes, err = statedb.GetOutcoinsByPubkey(transactionStateDB, *tokenID, keyset.PaymentAddress.Pk[:], shardID)
		if err != nil {
			return nil, err
		}
	}
	// convert from []byte to object
	outCoins := make([]*privacy.OutputCoin, 0)
	for _, item := range outCointsInBytes {
		outcoin := &privacy.OutputCoin{}
		outcoin.Init()
		outcoin.SetBytes(item)
		outCoins = append(outCoins, outcoin)
	}
	// loop on all outputcoin to decrypt data
	results := make([]*privacy.OutputCoin, 0)
	for _, out := range outCoins {
		decryptedOut := DecryptOutputCoinByKeyV2(transactionStateDB, out, keyset, tokenID, shardID)
		if decryptedOut == nil {
			continue
		} else {
			results = append(results, decryptedOut)
		}
	}
	return results, nil
}

// DecryptTxByKey - process outputcoin to get outputcoin data which relate to keyset
func DecryptOutputCoinByKeyV2(transactionStateDB *statedb.StateDB, outCoinTemp *privacy.OutputCoin, keySet *incognitokey.KeySet, tokenID *common.Hash, shardID byte) *privacy.OutputCoin {
	/*
		- Param keyset - (priv-key, payment-address, readonlykey)
		in case priv-key: return unspent outputcoin tx
		in case readonly-key: return all outputcoin tx with amount value
		in case payment-address: return all outputcoin tx with no amount value
	*/
	pubkeyCompress := outCoinTemp.CoinDetails.GetPublicKey().ToBytesS()
	if bytes.Equal(pubkeyCompress, keySet.PaymentAddress.Pk[:]) {
		result := &privacy.OutputCoin{
			CoinDetails:          outCoinTemp.CoinDetails,
			CoinDetailsEncrypted: outCoinTemp.CoinDetailsEncrypted,
		}
		if result.CoinDetailsEncrypted != nil && !result.CoinDetailsEncrypted.IsNil() {
			if len(keySet.ReadonlyKey.Rk) > 0 {
				// try to decrypt to get more data
				err := result.Decrypt(keySet.ReadonlyKey)
				if err != nil {
					return nil
				}
			}
		}
		if len(keySet.PrivateKey) > 0 {
			// check spent with private-key
			result.CoinDetails.SetSerialNumber(
				new(privacy.Point).Derive(
					privacy.PedCom.G[privacy.PedersenPrivateKeyIndex],
					new(privacy.Scalar).FromBytesS(keySet.PrivateKey),
					result.CoinDetails.GetSNDerivator()))
			ok, err := statedb.HasSerialNumber(transactionStateDB, *tokenID, result.CoinDetails.GetSerialNumber().ToBytesS(), shardID)
			if ok || err != nil {
				return nil
			}
		}
		return result
	}
	return nil
}

// CreateAndSaveTxViewPointFromBlock - fetch data from block, put into txviewpoint variable and save into db
// @note: still storage full data of commitments, serialnumbersm snderivator to check double spend
// @note: this function only work for transaction transfer token/prv within shard
func (blockchain *BlockChain) CreateAndSaveTxViewPointFromBlockV2(block *ShardBlock, transactionStateRoot *statedb.StateDB, beaconFeatureStateRoot *statedb.StateDB) error {
	//startTime := time.Now()
	// Fetch data from block into tx View point
	view := NewTxViewPoint(block.Header.ShardID)
	err := view.fetchTxViewPointFromBlockV2(transactionStateRoot, block)
	if err != nil {
		return err
	}
	// check privacy custom token
	// sort by index
	indices := []int{}
	for index := range view.privacyCustomTokenViewPoint {
		indices = append(indices, int(index))
	}
	sort.Ints(indices)
	for _, indexTx := range indices {
		privacyCustomTokenSubView := view.privacyCustomTokenViewPoint[int32(indexTx)]
		privacyCustomTokenTx := view.privacyCustomTokenTxs[int32(indexTx)]
		switch privacyCustomTokenTx.TxPrivacyTokenData.Type {
		case transaction.CustomTokenInit:
			{
				// check is bridge token
				isBridgeToken := false
				allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(beaconFeatureStateRoot)
				if err != nil {
					return err
				}
				if len(allBridgeTokensBytes) > 0 {
					var allBridgeTokens []*rawdb.BridgeTokenInfo
					err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
					for _, bridgeToken := range allBridgeTokens {
						if bridgeToken.TokenID != nil && bytes.Equal(privacyCustomTokenTx.TxPrivacyTokenData.PropertyID[:], bridgeToken.TokenID[:]) {
							isBridgeToken = true
						}
					}
				}
				// not mintable tx
				if !isBridgeToken && !privacyCustomTokenTx.TxPrivacyTokenData.Mintable {
					tokenID := privacyCustomTokenTx.TxPrivacyTokenData.PropertyID
					name := privacyCustomTokenTx.TxPrivacyTokenData.PropertyName
					symbol := privacyCustomTokenTx.TxPrivacyTokenData.PropertySymbol
					tokenType := privacyCustomTokenTx.TxPrivacyTokenData.Type
					mintable := privacyCustomTokenTx.TxPrivacyTokenData.Mintable
					amount := privacyCustomTokenTx.TxPrivacyTokenData.Amount
					txHash := *privacyCustomTokenTx.Hash()
					Logger.log.Info("Store custom token when it is issued", privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, privacyCustomTokenTx.TxPrivacyTokenData.PropertySymbol, privacyCustomTokenTx.TxPrivacyTokenData.PropertyName)
					err = statedb.StorePrivacyToken(transactionStateRoot, tokenID, name, symbol, tokenType, mintable, amount, txHash)
					if err != nil {
						return err
					}
				}
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Info("Transfer custom token %+v", privacyCustomTokenTx)
			}
		}
		err = statedb.StorePrivacyTokenTx(transactionStateRoot, privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, *privacyCustomTokenTx.Hash())
		if err != nil {
			return err
		}

		err = blockchain.StoreSerialNumbersFromTxViewPointV2(transactionStateRoot, *privacyCustomTokenSubView)
		if err != nil {
			return err
		}

		err = blockchain.StoreCommitmentsFromTxViewPointV2(transactionStateRoot, *privacyCustomTokenSubView, block.Header.ShardID)
		if err != nil {
			return err
		}

		err = blockchain.StoreSNDerivatorsFromTxViewPointV2(transactionStateRoot, *privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}

	// updateShardBestState the list serialNumber and commitment, snd set using the state of the used tx view point. This
	// entails adding the new
	// ones created by the block.
	err = blockchain.StoreSerialNumbersFromTxViewPointV2(transactionStateRoot, *view)
	if err != nil {
		return err
	}

	err = blockchain.StoreCommitmentsFromTxViewPointV2(transactionStateRoot, *view, block.Header.ShardID)
	if err != nil {
		return err
	}

	err = blockchain.StoreSNDerivatorsFromTxViewPointV2(transactionStateRoot, *view)
	if err != nil {
		return err
	}

	err = blockchain.StoreTxByPublicKeyV2(blockchain.GetDatabase(), view)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) StoreSerialNumbersFromTxViewPointV2(stateDB *statedb.StateDB, view TxViewPoint) error {
	if len(view.listSerialNumbers) > 0 {
		err := statedb.StoreSerialNumbers(stateDB, *view.tokenID, view.listSerialNumbers, view.shardID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (blockchain *BlockChain) StoreSNDerivatorsFromTxViewPointV2(stateDB *statedb.StateDB, view TxViewPoint) error {
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		snDsArray := view.mapSnD[k]
		err := statedb.StoreSNDerivators(stateDB, *view.tokenID, snDsArray)
		if err != nil {
			return err
		}
	}
	return nil
}

func (blockchain *BlockChain) StoreTxByPublicKeyV2(db incdb.Database, view *TxViewPoint) error {
	for data := range view.txByPubKey {
		dataArr := strings.Split(data, "_")
		pubKey, _, err := base58.Base58Check{}.Decode(dataArr[0])
		if err != nil {
			return err
		}
		txIDInByte, _, err := base58.Base58Check{}.Decode(dataArr[1])
		if err != nil {
			return err
		}
		txID := common.Hash{}
		err = txID.SetBytes(txIDInByte)
		if err != nil {
			return err
		}
		shardID, _ := strconv.Atoi(dataArr[2])

		err = rawdbv2.StoreTxByPublicKey(db, pubKey, txID, byte(shardID))
		if err != nil {
			return err
		}
	}
	return nil
}

func (blockchain *BlockChain) StoreCommitmentsFromTxViewPointV2(stateDB *statedb.StateDB, view TxViewPoint, shardID byte) error {
	// commitment and output are the same key in map
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		publicKey := k
		publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
		if err != nil {
			return err
		}
		lastByte := publicKeyBytes[len(publicKeyBytes)-1]
		publicKeyShardID := common.GetShardIDFromLastByte(lastByte)
		if publicKeyShardID == shardID {
			// commitment
			commitmentsArray := view.mapCommitments[k]
			err = statedb.StoreCommitments(stateDB, *view.tokenID, publicKeyBytes, commitmentsArray, view.shardID)
			if err != nil {
				return err
			}
			// outputs
			outputCoinArray := view.mapOutputCoins[k]
			outputCoinBytesArray := make([][]byte, 0)
			for _, outputCoin := range outputCoinArray {
				outputCoinBytesArray = append(outputCoinBytesArray, outputCoin.Bytes())
			}
			err = statedb.StoreOutputCoins(stateDB, *view.tokenID, publicKeyBytes, outputCoinBytesArray, publicKeyShardID)
			// clear cached data
			if blockchain.config.MemCache != nil {
				cachedKey := memcache.GetListOutputcoinCachedKey(publicKeyBytes, view.tokenID, publicKeyShardID)
				if ok, e := blockchain.config.MemCache.Has(cachedKey); ok && e == nil {
					er := blockchain.config.MemCache.Delete(cachedKey)
					if er != nil {
						Logger.log.Error("can not delete memcache", "GetListOutputcoinCachedKey", base58.Base58Check{}.Encode(cachedKey, 0x0))
					}
				}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) CreateAndSaveCrossTransactionViewPointFromBlockV2(shardBlock *ShardBlock, transactionStateRoot *statedb.StateDB) error {
	// Fetch data from block into tx View point
	view := NewTxViewPoint(shardBlock.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlockV2(transactionStateRoot, shardBlock)
	if err != nil {
		Logger.log.Error("CreateAndSaveCrossTransactionCoinViewPointFromBlock", err)
		return err
	}

	// sort by index
	indices := []int{}
	for index := range view.privacyCustomTokenViewPoint {
		indices = append(indices, int(index))
	}
	sort.Ints(indices)

	for _, index := range indices {
		privacyCustomTokenSubView := view.privacyCustomTokenViewPoint[int32(index)]
		// 0xsirrush updated: check existed tokenID
		tokenID := *privacyCustomTokenSubView.tokenID
		existed := statedb.PrivacyTokenIDExisted(transactionStateRoot, tokenID)
		if !existed {
			Logger.log.Info("Store custom token when it is issued ", tokenID, privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertyName, privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertySymbol, privacyCustomTokenSubView.privacyCustomTokenMetadata.Amount, privacyCustomTokenSubView.privacyCustomTokenMetadata.Mintable)
			name := privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertyName
			symbol := privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertySymbol
			tokenType := privacyCustomTokenSubView.privacyCustomTokenMetadata.Type
			mintable := privacyCustomTokenSubView.privacyCustomTokenMetadata.Mintable
			amount := privacyCustomTokenSubView.privacyCustomTokenMetadata.Amount
			if err := statedb.StorePrivacyToken(transactionStateRoot, tokenID, name, symbol, tokenType, mintable, amount, common.Hash{}); err != nil {
				return err
			}
		}
		// Store both commitment and outcoin
		err = blockchain.StoreCommitmentsFromTxViewPointV2(transactionStateRoot, *privacyCustomTokenSubView, shardBlock.Header.ShardID)
		if err != nil {
			return err
		}
		// store snd
		err = blockchain.StoreSNDerivatorsFromTxViewPointV2(transactionStateRoot, *privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}
	// store commitment
	err = blockchain.StoreCommitmentsFromTxViewPointV2(transactionStateRoot, *view, shardBlock.Header.ShardID)
	if err != nil {
		return err
	}
	// store snd
	err = blockchain.StoreSNDerivatorsFromTxViewPointV2(transactionStateRoot, *view)
	if err != nil {
		return err
	}

	return nil
}
