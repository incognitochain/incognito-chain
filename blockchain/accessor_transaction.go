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
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (blockchain *BlockChain) ValidateResponseTransactionFromTxsWithMetadataV2(shardBlock *ShardBlock) error {
	txRequestTable := reqTableFromReqTxs(shardBlock.Body.Transactions)
	if shardBlock.Header.Timestamp > ValidateTimeForSpamRequestTxs {
		txsSpamRemoved := filterReqTxs(shardBlock.Body.Transactions, txRequestTable)
		if len(shardBlock.Body.Transactions) != len(txsSpamRemoved) {
			return errors.Errorf("This block contains txs spam request reward. Number of spam: %v", len(shardBlock.Body.Transactions)-len(txsSpamRemoved))
		}
	}
	txReturnTable := map[string]bool{}
	for _, tx := range shardBlock.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.WithDrawRewardResponseMeta:
			_, requesterRes, amountRes, coinID := tx.GetTransferData()
			requester := getRequesterFromPKnCoinID(requesterRes, *coinID)
			txReq, isExist := txRequestTable[requester]
			if !isExist {
				return errors.New("This response dont match with any request")
			}
			requestMeta := txReq.GetMetadata().(*metadata.WithDrawRewardRequest)
			responseMeta := tx.GetMetadata().(*metadata.WithDrawRewardResponse)
			if res, err := coinID.Cmp(&requestMeta.TokenID); err != nil || res != 0 {
				return errors.Errorf("Invalid token ID when check metadata of tx response. Got %v, want %v, error %v", coinID, requestMeta.TokenID, err)
			}
			if cmp, err := responseMeta.TxRequest.Cmp(txReq.Hash()); (cmp != 0) || (err != nil) {
				Logger.log.Errorf("Response does not match with request, response link to txID %v, request txID %v, error %v", responseMeta.TxRequest.String(), txReq.Hash().String(), err)
			}
			tempPublicKey := base58.Base58Check{}.Encode(requesterRes, common.Base58Version)
			amount, err := statedb.GetCommitteeReward(blockchain.GetShardRewardStateDB(shardBlock.Header.ShardID), tempPublicKey, requestMeta.TokenID)
			if (amount == 0) || (err != nil) {
				return errors.Errorf("Invalid request %v, amount from db %v, error %v", requester, amount, err)
			}
			if amount != amountRes {
				return errors.Errorf("Wrong amount for token %v, get from db %v, response amount %v", requestMeta.TokenID, amount, amountRes)
			}
			delete(txRequestTable, requester)
			continue
		case metadata.ReturnStakingMeta:
			returnMeta := tx.GetMetadata().(*metadata.ReturnStakingMetadata)
			if _, ok := txReturnTable[returnMeta.StakerAddress.String()]; !ok {
				txReturnTable[returnMeta.StakerAddress.String()] = true
			} else {
				return errors.New("Double spent transaction return staking for a candidate.")
			}
		}
	}
	if shardBlock.Header.Timestamp > ValidateTimeForSpamRequestTxs {
		if len(txRequestTable) > 0 {
			return errors.Errorf("Not match request and response, num of unresponse request: %v", len(txRequestTable))
		}
	}
	return nil
}

func (blockchain *BlockChain) InitTxSalaryByCoinID(
	payToAddress *privacy.PaymentAddress,
	amount uint64,
	payByPrivateKey *privacy.PrivateKey,
	stateDB *statedb.StateDB,
	meta metadata.Metadata,
	coinID common.Hash,
	shardID byte,
) (metadata.Transaction, error) {
	txType := -1
	if res, err := coinID.Cmp(&common.PRVCoinID); err == nil && res == 0 {
		txType = transaction.NormalCoinType
	}
	if txType == -1 {
		allBridgeTokensBytes, err := rawdb.GetAllBridgeTokens(blockchain.config.DataBase)
		if err != nil {
			return nil, err
		}
		var allBridgeTokens []*rawdb.BridgeTokenInfo
		err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)

		if err != nil {
			return nil, err
		}
		for _, bridgeTokenIDs := range allBridgeTokens {
			if res, err := coinID.Cmp(bridgeTokenIDs.TokenID); err == nil && res == 0 {
				txType = transaction.CustomTokenPrivacyType
				break
			}
		}
	}
	if txType == -1 {
		mapPrivacyCustomToken, mapCrossShardCustomToken, err := blockchain.ListPrivacyCustomToken()
		if err != nil {
			return nil, err
		}
		if mapPrivacyCustomToken != nil {
			if _, ok := mapPrivacyCustomToken[coinID]; ok {
				txType = transaction.CustomTokenPrivacyType
			}
		}
		if mapCrossShardCustomToken != nil {
			if _, ok := mapCrossShardCustomToken[coinID]; ok {
				txType = transaction.CustomTokenPrivacyType
			}
		}
	}
	if txType == -1 {
		return nil, errors.Errorf("Invalid token ID when InitTxSalaryByCoinID. Got %v", coinID)
	}
	buildCoinBaseParams := transaction.NewBuildCoinBaseTxByCoinIDParams(payToAddress,
		amount,
		payByPrivateKey,
		stateDB,
		meta,
		coinID,
		txType,
		coinID.String(),
		shardID)
	return transaction.BuildCoinBaseTxByCoinID(buildCoinBaseParams)
}

// @Notice: change from body.Transaction -> transactions
func (blockchain *BlockChain) BuildResponseTransactionFromTxsWithMetadataV2(transactions []metadata.Transaction, blkProducerPrivateKey *privacy.PrivateKey, shardID byte) ([]metadata.Transaction, error) {
	txRequestTable := reqTableFromReqTxs(transactions)
	txsResponse := []metadata.Transaction{}
	for key, value := range txRequestTable {
		txRes, err := blockchain.buildWithDrawTransactionResponseV2(&value, blkProducerPrivateKey, shardID)
		if err != nil {
			Logger.log.Errorf("Build Withdraw transactions response for tx %v return errors %v", value, err)
			delete(txRequestTable, key)
			continue
		} else {
			Logger.log.Infof("[Reward] - BuildWithDrawTransactionResponse for tx %+v, ok: %+v\n", value, txRes)
		}
		txsResponse = append(txsResponse, txRes)
	}
	txsSpamRemoved := filterReqTxs(transactions, txRequestTable)
	Logger.log.Infof("Number of metadata txs: %v; number of tx request %v; number of tx spam %v; number of tx response %v",
		len(transactions),
		len(txRequestTable),
		len(transactions)-len(txsSpamRemoved),
		len(txsResponse))
	txsSpamRemoved = append(txsSpamRemoved, txsResponse...)
	return txsSpamRemoved, nil
}

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

// DecryptOutputCoinByKeyV2 process outputcoin to get outputcoin data which relate to keyset
// Param keyset: (private key, payment address, read only key)
// in case private key: return unspent outputcoin tx
// in case read only key: return all outputcoin tx with amount value
// in case payment address: return all outputcoin tx with no amount value
func DecryptOutputCoinByKeyV2(transactionStateDB *statedb.StateDB, outCoinTemp *privacy.OutputCoin, keySet *incognitokey.KeySet, tokenID *common.Hash, shardID byte) *privacy.OutputCoin {
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
			// check spent with private key
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

func storePRV(transactionStateRoot *statedb.StateDB) error {
	tokenID := common.PRVCoinID
	name := common.PRVCoinName
	symbol := common.PRVCoinName
	tokenType := 0
	mintable := false
	amount := uint64(1000000000000000)
	info := []byte{}
	txHash := common.Hash{}
	err := statedb.StorePrivacyToken(transactionStateRoot, tokenID, name, symbol, tokenType, mintable, amount, info, txHash)
	if err != nil {
		return err
	}
	return nil
}

// CreateAndSaveTxViewPointFromBlock - fetch data from block, put into txviewpoint variable and save into db
// still storage full data of commitments, serial number, snderivator to check double spend
// this function only work for transaction transfer token/prv within shard
func (blockchain *BlockChain) CreateAndSaveTxViewPointFromBlockV2(shardBlock *ShardBlock, transactionStateRoot *statedb.StateDB, beaconFeatureStateRoot *statedb.StateDB) error {
	// Fetch data from shardBlock into tx View point
	if shardBlock.Header.Height == 1 {
		err := storePRV(transactionStateRoot)
		if err != nil {
			return err
		}
	}
	var allBridgeTokens []*rawdb.BridgeTokenInfo
	var err error
	allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(beaconFeatureStateRoot)
	if err != nil {
		return err
	}
	if len(allBridgeTokensBytes) > 0 {
		err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
	}
	view := NewTxViewPoint(shardBlock.Header.ShardID)
	err = view.fetchTxViewPointFromBlockV2(transactionStateRoot, shardBlock)
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
		isBridgeToken := false
		for _, bridgeToken := range allBridgeTokens {
			if bridgeToken.TokenID != nil && bytes.Equal(privacyCustomTokenTx.TxPrivacyTokenData.PropertyID[:], bridgeToken.TokenID[:]) {
				isBridgeToken = true
			}
		}
		switch privacyCustomTokenTx.TxPrivacyTokenData.Type {
		case transaction.CustomTokenInit:
			{
				// check is bridge token
				// not mintable tx
				if !isBridgeToken && !privacyCustomTokenTx.TxPrivacyTokenData.Mintable {
					tokenID := privacyCustomTokenTx.TxPrivacyTokenData.PropertyID
					name := privacyCustomTokenTx.TxPrivacyTokenData.PropertyName
					symbol := privacyCustomTokenTx.TxPrivacyTokenData.PropertySymbol
					tokenType := privacyCustomTokenTx.TxPrivacyTokenData.Type
					mintable := privacyCustomTokenTx.TxPrivacyTokenData.Mintable
					amount := privacyCustomTokenTx.TxPrivacyTokenData.Amount
					info := privacyCustomTokenTx.Tx.Info
					txHash := *privacyCustomTokenTx.Hash()
					Logger.log.Info("Store custom token when it is issued", privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, privacyCustomTokenTx.TxPrivacyTokenData.PropertySymbol, privacyCustomTokenTx.TxPrivacyTokenData.PropertyName)
					err = statedb.StorePrivacyToken(transactionStateRoot, tokenID, name, symbol, tokenType, mintable, amount, info, txHash)
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
		err = statedb.StorePrivacyTokenTx(transactionStateRoot, privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, *privacyCustomTokenTx.Hash(), privacyCustomTokenTx.TxPrivacyTokenData.Mintable || isBridgeToken)
		if err != nil {
			return err
		}

		err = blockchain.StoreSerialNumbersFromTxViewPointV2(transactionStateRoot, *privacyCustomTokenSubView)
		if err != nil {
			return err
		}

		err = blockchain.StoreCommitmentsFromTxViewPointV2(transactionStateRoot, *privacyCustomTokenSubView, shardBlock.Header.ShardID)
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
	// ones created by the shardBlock.
	err = blockchain.StoreSerialNumbersFromTxViewPointV2(transactionStateRoot, *view)
	if err != nil {
		return err
	}

	err = blockchain.StoreCommitmentsFromTxViewPointV2(transactionStateRoot, *view, shardBlock.Header.ShardID)
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
	Logger.log.Critical("Fetch Cross transaction", shardBlock.Body.CrossTransactions)
	// Fetch data from block into tx View point
	view := NewTxViewPoint(shardBlock.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlockV2(transactionStateRoot, shardBlock)
	if err != nil {
		Logger.log.Error("CreateAndSaveCrossTransactionCoinViewPointFromBlock ", err)
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
			info := []byte{}
			if err := statedb.StorePrivacyToken(transactionStateRoot, tokenID, name, symbol, tokenType, mintable, amount, info, common.Hash{}); err != nil {
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
