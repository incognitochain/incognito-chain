package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/memcache"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"
)

// DecryptOutputCoinByKey process outputcoin to get outputcoin data which relate to keyset
// Param keyset: (private key, payment address, read only key)
// in case private key: return unspent outputcoin tx
// in case read only key: return all outputcoin tx with amount value
// in case payment address: return all outputcoin tx with no amount value
func DecryptOutputCoinByKey(transactionStateDB *statedb.StateDB, outCoin coin.Coin, keySet *incognitokey.KeySet, tokenID *common.Hash, shardID byte) (coin.PlainCoin, error) {
	result, err := outCoin.Decrypt(keySet)
	if err != nil {
		Logger.log.Errorf("Cannot decrypt output coin by key %v", err)
		return nil, err
	}
	keyImage := result.GetKeyImage()
	if keyImage != nil {
		ok, err := statedb.HasSerialNumber(transactionStateDB, *tokenID, keyImage.ToBytesS(), shardID)
		if ok || err != nil {
			Logger.log.Errorf("There is something wrong when check key image %v", err)
			return nil, err
		}
	}
	return result, nil
}

func storePRV(transactionStateRoot *statedb.StateDB) error {
	tokenID := common.PRVCoinID
	name := common.PRVCoinName
	symbol := common.PRVCoinName
	tokenType := 0
	mintable := false
	amount := uint64(0)
	info := []byte{}
	txHash := common.Hash{}
	err := statedb.StorePrivacyToken(transactionStateRoot, tokenID, name, symbol, tokenType, mintable, amount, info, txHash)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) GetTransactionByHash(txHash common.Hash) (byte, common.Hash, int, metadata.Transaction, error) {
	blockHash, index, err := rawdbv2.GetTransactionByHash(blockchain.config.DataBase, txHash)
	if err != nil {
		return byte(255), common.Hash{}, -1, nil, NewBlockChainError(GetTransactionFromDatabaseError, err)
	}
	shardBlock, _, err := blockchain.GetShardBlockByHash(blockHash)
	if err != nil {
		return byte(255), common.Hash{}, -1, nil, NewBlockChainError(GetTransactionFromDatabaseError, err)
	}
	return shardBlock.Header.ShardID, blockHash, index, shardBlock.Body.Transactions[index], nil
}

// GetTransactionHashByReceiver - return list tx id which receiver get from any sender
// this feature only apply on full node, because full node get all data from all shard
func (blockchain *BlockChain) GetTransactionHashByReceiver(keySet *incognitokey.KeySet) (map[byte][]common.Hash, error) {
	result := make(map[byte][]common.Hash)
	var err error
	result, err = rawdbv2.GetTxByPublicKey(blockchain.config.DataBase, keySet.PaymentAddress.Pk)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	return result, nil
}

func (blockchain *BlockChain) ValidateResponseTransactionFromTxsWithMetadata(shardBlock *ShardBlock) error {
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
			Logger.log.Infof("Token ID %+v", requestMeta.TokenID)
			Logger.log.Infof("Coin ID %+v", *coinID)
			Logger.log.Infof("Amount Request %+v", amountRes)
			Logger.log.Infof("Temp Public Key %+v", tempPublicKey)
			amount, err := statedb.GetCommitteeReward(blockchain.BestState.Shard[shardBlock.Header.ShardID].GetCopiedRewardStateDB(), tempPublicKey, requestMeta.TokenID)
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
	transactionStateDB *statedb.StateDB,
	bridgeStateDB *statedb.StateDB,
	meta metadata.Metadata,
	coinID common.Hash,
	shardID byte,
) (metadata.Transaction, error) {
	txType := -1
	if res, err := coinID.Cmp(&common.PRVCoinID); err == nil && res == 0 {
		txType = transaction.NormalCoinType
	}
	if txType == -1 {
		tokenIDs, err := blockchain.ListPrivacyTokenAndBridgeTokenAndPRVByShardID(shardID)
		if err != nil {
			return nil, err
		}
		// coinID must not equal to PRVCoinID
		for _, tokenID := range tokenIDs {
			if res, err := coinID.Cmp(&tokenID); err == nil && res == 0 {
				txType = transaction.CustomTokenPrivacyType
				break
			}
		}
	}
	if txType == -1 {
		return nil, errors.Errorf("Invalid token ID when InitTxSalaryByCoinID. Got %v", coinID)
	}
	buildCoinBaseParams := transaction.NewBuildCoinBaseTxByCoinIDParams(payToAddress,
		amount,
		payByPrivateKey,
		transactionStateDB,
		meta,
		coinID,
		txType,
		coinID.String(),
		shardID,
		bridgeStateDB)
	return transaction.BuildCoinBaseTxByCoinID(buildCoinBaseParams)
}

// @Notice: change from body.Transaction -> transactions
func (blockchain *BlockChain) BuildResponseTransactionFromTxsWithMetadata(transactions []metadata.Transaction, blkProducerPrivateKey *privacy.PrivateKey, shardID byte) ([]metadata.Transaction, error) {
	txRequestTable := reqTableFromReqTxs(transactions)
	txsResponse := []metadata.Transaction{}
	for key, value := range txRequestTable {
		txRes, err := blockchain.buildWithDrawTransactionResponse(&value, blkProducerPrivateKey, shardID)
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

func (blockchain *BlockChain) QueryDBToGetOutcoinsBytesByKeyset(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash, shardHeight uint64) ([][]byte, error) {
	var err error
	var outCoinsBytes [][]byte
	transactionStateDB := blockchain.BestState.Shard[shardID].transactionStateDB

	// Get outcoins in ver1 by publicSpend key
	outCoinsBytes, err = statedb.GetOutcoinsByPubkey(transactionStateDB, *tokenID, keyset.PaymentAddress.Pk[:], shardID)
	if err != nil {
		Logger.log.Error("GetOutcoinsBytesByKeyset Get by PubKey", err)
		return nil, err
	}

	// Get outcoins in ver2 by privateView key
	currentHeight := blockchain.BestState.Shard[shardID].ShardHeight
	for height := shardHeight; height <= currentHeight; height += 1 {
		currentHeightCoins, err := statedb.GetOnetimeAddressesByHeight(transactionStateDB, *tokenID, shardID, shardHeight)
		if err != nil {
			Logger.log.Error("GetOutcoinsBytesByKeyset Get By Height", err)
			return nil, err
		}
		for _, coinBytes := range currentHeightCoins {
			c, err := coin.CreateCoinFromByte(coinBytes)
			if err != nil {
				Logger.log.Error("GetOutcoinsBytesByKeyset Parse Coin From Bytes", err)
				return nil, err
			}
			if coin.IsCoinBelongToViewKey(c, keyset.ReadonlyKey) {
				outCoinsBytes = append(outCoinsBytes, currentHeightCoins...)
			}
		}
	}
	return outCoinsBytes, nil
}

//GetListDecryptedOutputCoinsByKeyset - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key.
//With private-key, we can check unspent tx by check serialNumber from database
//- Param #1: keyset - (priv-key, payment-address, readonlykey)
//in case priv-key: return unspent outputcoin tx
//in case readonly-key: return all outputcoin tx with amount value
//in case payment-address: return all outputcoin tx with no amount value
//- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
func (blockchain *BlockChain) GetListDecryptedOutputCoinsByKeyset(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash, shardHeight uint64) ([]coin.PlainCoin, error) {
	blockchain.BestState.Shard[shardID].lock.Lock()
	defer blockchain.BestState.Shard[shardID].lock.Unlock()
	var outCoinsInBytes [][]byte
	var err error
	transactionStateDB := blockchain.BestState.Shard[shardID].transactionStateDB
	if keyset == nil {
		return nil, NewBlockChainError(GetListDecryptedOutputCoinsByKeysetError, fmt.Errorf("invalid key set, got keyset %+v", keyset))
	}
	if blockchain.config.MemCache != nil {
		// get from cache
		cachedKey := memcache.GetListOutputcoinCachedKey(keyset.PaymentAddress.Pk[:], tokenID, shardID)
		cachedData, _ := blockchain.config.MemCache.Get(cachedKey)
		if cachedData != nil && len(cachedData) > 0 {
			// try to parsing on outCointsInBytes
			_ = json.Unmarshal(cachedData, &outCoinsInBytes)
		}
		if len(outCoinsInBytes) == 0 {
			// cached data is nil or fail -> get from database
			outCoinsInBytes, err = blockchain.QueryDBToGetOutcoinsBytesByKeyset(keyset, shardID, tokenID, shardHeight)
			if len(outCoinsInBytes) > 0 {
				// cache 1 day for result
				cachedData, err = json.Marshal(outCoinsInBytes)
				if err == nil {
					blockchain.config.MemCache.PutExpired(cachedKey, cachedData, 1*24*60*60*time.Millisecond)
				}
			}
		}
	}
	if len(outCoinsInBytes) == 0 {
		outCoinsInBytes, err = blockchain.QueryDBToGetOutcoinsBytesByKeyset(keyset, shardID, tokenID, shardHeight)
		if err != nil {
			return nil, err
		}
	}
	// loop on all outputcoin to decrypt data
	results := make([]coin.PlainCoin, 0)
	for _, item := range outCoinsInBytes {
		outCoin, err := coin.CreateCoinFromByte(item)
		if err != nil {
			Logger.log.Errorf("Cannot create coin from byte %v", err)
			return nil, err
		}
		decryptedOut, _ := DecryptOutputCoinByKey(transactionStateDB, outCoin, keyset, tokenID, shardID)
		if decryptedOut != nil {
			results = append(results, decryptedOut)
		}
	}
	return results, nil
}

// CreateAndSaveTxViewPointFromBlock - fetch data from block, put into txviewpoint variable and save into db
// still storage full data of commitments, serial number, snderivator to check double spend
// this function only work for transaction transfer token/prv within shard
func (blockchain *BlockChain) CreateAndSaveTxViewPointFromBlock(shardBlock *ShardBlock, transactionStateRoot *statedb.StateDB, beaconFeatureStateRoot *statedb.StateDB) error {
	// Fetch data from shardBlock into tx View point
	if shardBlock.Header.Height == 1 {
		err := storePRV(transactionStateRoot)
		if err != nil {
			return err
		}
	}
	var err error
	_, allBridgeTokens, err := blockchain.GetAllBridgeTokens()
	if err != nil {
		return err
	}
	view := NewTxViewPoint(shardBlock.Header.ShardID)
	err = view.fetchTxViewPointFromBlock(transactionStateRoot, shardBlock)
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
		for _, tempBridgeToken := range allBridgeTokens {
			if tempBridgeToken.TokenID != nil && bytes.Equal(privacyCustomTokenTx.TxPrivacyTokenData.PropertyID[:], tempBridgeToken.TokenID[:]) {
				isBridgeToken = true
			}
		}
		switch privacyCustomTokenTx.TxPrivacyTokenData.Type {
		case transaction.CustomTokenInit:
			{
				tokenID := privacyCustomTokenTx.TxPrivacyTokenData.PropertyID
				existed := statedb.PrivacyTokenIDExisted(transactionStateRoot, tokenID)
				if !existed {
					// check is bridge token
					tokenID := privacyCustomTokenTx.TxPrivacyTokenData.PropertyID
					name := privacyCustomTokenTx.TxPrivacyTokenData.PropertyName
					symbol := privacyCustomTokenTx.TxPrivacyTokenData.PropertySymbol
					mintable := privacyCustomTokenTx.TxPrivacyTokenData.Mintable
					amount := privacyCustomTokenTx.TxPrivacyTokenData.Amount
					info := privacyCustomTokenTx.Tx.Info
					txHash := *privacyCustomTokenTx.Hash()
					tokenType := statedb.InitToken
					if isBridgeToken {
						tokenType = statedb.BridgeToken
					}
					Logger.log.Info("Store custom token when it is issued", privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, privacyCustomTokenTx.TxPrivacyTokenData.PropertySymbol, privacyCustomTokenTx.TxPrivacyTokenData.PropertyName)
					err := statedb.StorePrivacyToken(transactionStateRoot, tokenID, name, symbol, tokenType, mintable, amount, info, txHash)
					if err != nil {
						return err
					}
				}
			}
		case transaction.CustomTokenTransfer:
			{
				Logger.log.Infof("Transfer custom token %+v", privacyCustomTokenTx)
			}
		}
		err = statedb.StorePrivacyTokenTx(transactionStateRoot, privacyCustomTokenTx.TxPrivacyTokenData.PropertyID, *privacyCustomTokenTx.Hash())
		if err != nil {
			return err
		}

		err = blockchain.StoreSerialNumbersFromTxViewPoint(transactionStateRoot, *privacyCustomTokenSubView)
		if err != nil {
			return err
		}

		err = blockchain.StoreCommitmentsFromTxViewPoint(transactionStateRoot, *privacyCustomTokenSubView, shardBlock.Header.ShardID)
		if err != nil {
			return err
		}

		err = blockchain.StoreSNDerivatorsFromTxViewPoint(transactionStateRoot, *privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}

	// updateShardBestState the list serialNumber and commitment, snd set using the state of the used tx view point. This
	// entails adding the new
	// ones created by the shardBlock.
	err = blockchain.StoreSerialNumbersFromTxViewPoint(transactionStateRoot, *view)
	if err != nil {
		return err
	}

	err = blockchain.StoreCommitmentsFromTxViewPoint(transactionStateRoot, *view, shardBlock.Header.ShardID)
	if err != nil {
		return err
	}

	err = blockchain.StoreSNDerivatorsFromTxViewPoint(transactionStateRoot, *view)
	if err != nil {
		return err
	}

	err = blockchain.StoreTxByPublicKey(blockchain.GetDatabase(), view)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) StoreSerialNumbersFromTxViewPoint(stateDB *statedb.StateDB, view TxViewPoint) error {
	if len(view.listSerialNumbers) > 0 {
		err := statedb.StoreSerialNumbers(stateDB, *view.tokenID, view.listSerialNumbers, view.shardID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (blockchain *BlockChain) StoreSNDerivatorsFromTxViewPoint(stateDB *statedb.StateDB, view TxViewPoint) error {
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

func (blockchain *BlockChain) StoreTxByPublicKey(db incdb.Database, view *TxViewPoint) error {
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

func (blockchain *BlockChain) StoreCommitmentsFromTxViewPoint(stateDB *statedb.StateDB, view TxViewPoint, shardID byte) error {
	// commitment and output are the same key in map
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Start to store to db
	for _, k := range keys {
		publicKey := k
		publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
		if err != nil {
			return err
		}
		publicKeyShardID := common.GetShardIDFromLastByte(publicKeyBytes[len(publicKeyBytes)-1])
		if publicKeyShardID == shardID {
			var mapComSnd map[string][]byte = make(map[string][]byte)

			// outputs
			outputCoinArray := view.mapOutputCoins[k]
			outputCoinBytesArray := make([][]byte, 0)
			for _, outputCoin := range outputCoinArray {
				commitment := outputCoin.GetCommitment()
				commitmentBytes := commitment.ToBytesS()
				snd := outputCoin.GetSNDerivator()
				mapComSnd[string(commitmentBytes)] = snd.ToBytesS()

				outputCoinBytesArray = append(outputCoinBytesArray, outputCoin.Bytes())
			}
			err = statedb.StoreOnetimeAddress(stateDB, *view.tokenID, view.height, outputCoinBytesArray, publicKeyShardID)
			if err != nil {
				return err
			}

			err = statedb.StoreOutputCoins(stateDB, *view.tokenID, publicKeyBytes, outputCoinBytesArray, publicKeyShardID)
			if err != nil {
				return err
			}

			// commitment
			commitmentsArray := view.mapCommitments[k]
			err = statedb.StoreCommitments(stateDB, *view.tokenID, publicKeyBytes, mapComSnd, commitmentsArray, view.shardID)
			if err != nil {
				return err
			}

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

func (blockchain *BlockChain) CreateAndSaveCrossTransactionViewPointFromBlock(shardBlock *ShardBlock, transactionStateRoot *statedb.StateDB) error {
	Logger.log.Critical("Fetch Cross transaction", shardBlock.Body.CrossTransactions)
	// Fetch data from block into tx View point
	view := NewTxViewPoint(shardBlock.Header.ShardID)
	err := view.fetchCrossTransactionViewPointFromBlock(transactionStateRoot, shardBlock)
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
			mintable := privacyCustomTokenSubView.privacyCustomTokenMetadata.Mintable
			amount := privacyCustomTokenSubView.privacyCustomTokenMetadata.Amount
			info := []byte{}
			if err := statedb.StorePrivacyToken(transactionStateRoot, tokenID, name, symbol, statedb.CrossShardToken, mintable, amount, info, common.Hash{}); err != nil {
				return err
			}
		}
		// Store both commitment and outcoin
		err = blockchain.StoreCommitmentsFromTxViewPoint(transactionStateRoot, *privacyCustomTokenSubView, shardBlock.Header.ShardID)
		if err != nil {
			return err
		}
		// store snd
		err = blockchain.StoreSNDerivatorsFromTxViewPoint(transactionStateRoot, *privacyCustomTokenSubView)
		if err != nil {
			return err
		}
	}
	// store commitment
	err = blockchain.StoreCommitmentsFromTxViewPoint(transactionStateRoot, *view, shardBlock.Header.ShardID)
	if err != nil {
		return err
	}
	// store snd
	err = blockchain.StoreSNDerivatorsFromTxViewPoint(transactionStateRoot, *view)
	if err != nil {
		return err
	}
	return nil
}
