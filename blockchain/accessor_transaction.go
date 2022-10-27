package blockchain

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	coinIndexer "github.com/incognitochain/incognito-chain/transaction/coin_indexer"
	"github.com/incognitochain/incognito-chain/wallet"

	"sort"
	"strconv"
	"strings"

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

var (
	EnableIndexingCoinByOTAKey bool
	outcoinIndexer             *coinIndexer.CoinIndexer
)

func GetCoinIndexer() *coinIndexer.CoinIndexer {
	return outcoinIndexer
}

// DecryptOutputCoinByKey process outputcoin to get outputcoin data which relate to keyset
// Param keyset: (private key, payment address, read only key)
// in case private key: return unspent outputcoin tx
// in case read only key: return all outputcoin tx with amount value
// in case payment address: return all outputcoin tx with no amount value
func DecryptOutputCoinByKey(transactionStateDB *statedb.StateDB, outCoin privacy.Coin, keySet *incognitokey.KeySet, tokenID *common.Hash, shardID byte) (privacy.PlainCoin, error) {
	if tokenID == nil {
		clonedTokenID := common.PRVCoinID
		tokenID = &clonedTokenID
	}
	result, err := outCoin.Decrypt(keySet)
	if err != nil {
		return nil, err
	}
	keyImage := result.GetKeyImage()
	if keyImage != nil {
		ok, err := statedb.HasSerialNumber(transactionStateDB, *tokenID, keyImage.ToBytesS(), shardID)
		if err != nil {
			Logger.log.Errorf("There is something wrong when check key image %v", err)
			return nil, err
		} else if ok {
			// The KeyImage is valid but already spent
			return nil, nil
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

func (blockchain *BlockChain) GetTransactionByHash(txHash common.Hash) (byte, common.Hash, uint64, int, metadata.Transaction, error) {
	for _, i := range blockchain.GetShardIDs() {
		shardID := byte(i)
		blockHash, index, err := blockchain.ShardChain[shardID].BlockStorage.GetTXIndex(txHash)
		if err != nil {
			continue
		}
		// error is nil
		shardBlock, _, err := blockchain.GetShardBlockByHashWithShardID(blockHash, shardID)
		if err != nil {
			continue
		}
		return shardBlock.Header.ShardID, blockHash, shardBlock.GetHeight(), index, shardBlock.Body.Transactions[index], nil
	}
	return byte(255), common.Hash{}, 0, -1, nil, NewBlockChainError(GetTransactionFromDatabaseError, fmt.Errorf("Not found transaction with tx hash %+v", txHash))
}

func (blockchain *BlockChain) GetTransactionByHashWithShardID(txHash common.Hash, shardID byte) (common.Hash, int, metadata.Transaction, error) {
	blockHash, index, err := blockchain.ShardChain[shardID].BlockStorage.GetTXIndex(txHash)
	if err != nil {
		return common.Hash{}, -1, nil, NewBlockChainError(GetTransactionFromDatabaseError, fmt.Errorf("Not found transaction with tx hash %+v", txHash))
	}
	// error is nil
	shardBlock, _, err := blockchain.GetShardBlockByHashWithShardID(blockHash, shardID)
	if err != nil {
		return common.Hash{}, -1, nil, NewBlockChainError(GetTransactionFromDatabaseError, fmt.Errorf("Not found transaction with tx hash %+v", txHash))
	}
	return blockHash, index, shardBlock.Body.Transactions[index], nil
}

// GetTransactionHashByReceiver - return list tx id which receiver get from any sender
// this feature only apply on full node, because full node get all data from all shard
func (blockchain *BlockChain) GetTransactionHashByReceiver(keySet *incognitokey.KeySet) (map[byte][]common.Hash, error) {
	result := make(map[byte][]common.Hash)
	for _, i := range blockchain.GetShardIDs() {
		shardID := byte(i)
		var err error
		resultTemp, err := rawdbv2.GetTxByPublicKey(blockchain.GetShardChainDatabase(shardID), keySet.PaymentAddress.Pk)
		if err == nil {
			if resultTemp == nil || len(resultTemp) == 0 {
				continue
			}
			result[shardID] = resultTemp[shardID]
		}
	}
	return result, nil
}

// GetTransactionHashByReceiverV2 - return list tx id which a receiver receives from any senders in paging fashion
// this feature only apply on full node, because full node get all data from all shard
func (blockchain *BlockChain) GetTransactionHashByReceiverV2(
	keySet *incognitokey.KeySet,
	skip, limit uint,
) (map[byte][]common.Hash, error) {
	result := make(map[byte][]common.Hash)
	for _, i := range blockchain.GetShardIDs() {
		shardID := byte(i)
		var err error
		var resultTemp map[byte][]common.Hash
		resultTemp, skip, limit, err = rawdbv2.GetTxByPublicKeyV2(blockchain.GetShardChainDatabase(shardID), keySet.PaymentAddress.Pk, skip, limit)
		if err == nil {
			if resultTemp == nil || len(resultTemp) == 0 {
				continue
			}
			result[shardID] = resultTemp[shardID]
		}
		if limit == 0 {
			break
		}
	}
	return result, nil
}

func (blockchain *BlockChain) ValidateResponseTransactionFromTxsWithMetadata(shardBlock *types.ShardBlock) error {
	// filter double withdraw request
	withdrawReqTable := make(map[string]*metadata.WithDrawRewardRequest)
	for _, tx := range shardBlock.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.WithDrawRewardRequestMeta:
			if tx.GetMetadata() == nil {
				return fmt.Errorf("metadata is nil for type %v", tx.GetMetadataType())
			}

			md, ok := tx.GetMetadata().(*metadata.WithDrawRewardRequest)
			if !ok {
				return fmt.Errorf("cannot parse withdraw request for tx %v", tx.Hash().String())
			}
			if _, ok = withdrawReqTable[tx.Hash().String()]; !ok {
				withdrawReqTable[tx.Hash().String()] = md
			} else {
				return errors.Errorf("Double withdraw request, tx double %v", tx.Hash().String())
			}
		}
	}
	rewardDB := blockchain.GetBestStateShard(shardBlock.Header.ShardID).GetShardRewardStateDB()
	// check tx withdraw response valid with the corresponding request
	for _, tx := range shardBlock.Body.Transactions {
		if tx.GetMetadataType() == metadata.WithDrawRewardResponseMeta {
			//check valid info with tx request
			if tx.GetMetadata() == nil {
				return fmt.Errorf("metadata is nil for type %v", tx.GetMetadataType())
			}

			metaResponse, ok := tx.GetMetadata().(*metadata.WithDrawRewardResponse)
			if !ok {
				return fmt.Errorf("cannot cast %v to a withdraw reward response", tx.GetMetadata())
			}

			metaRequest, ok := withdrawReqTable[metaResponse.TxRequest.String()]
			if !ok {
				return fmt.Errorf("cannot found tx request for tx withdraw reward response %v", tx.Hash().String())
			} else {
				delete(withdrawReqTable, metaResponse.TxRequest.String())
			}
			rewardPaymentAddress := metaRequest.PaymentAddress

			isMinted, mintCoin, coinID, err := tx.GetTxMintData()
			//check tx mint
			if err != nil || !isMinted {
				return errors.Errorf("[Mint Withdraw Reward] It is not tx mint with error: %v", err)
			}
			//check tokenID
			if metaRequest.TokenID.String() != coinID.String() {
				return fmt.Errorf("token in the request (%v) and the minted token mismatch (%v)", metaRequest.TokenID.String(), coinID.String())
			}

			//check amount & receiver
			receiver := base58.Base58Check{}.Encode(metaRequest.PaymentAddress.Pk, common.Base58Version)
			rewardAmount, err := statedb.GetCommitteeReward(rewardDB, receiver, *coinID)
			if err != nil {
				return errors.Errorf("[Mint Withdraw Reward] Cannot get reward amount for receiver %v, token %v, err %v", receiver, coinID.String(), err)
			}
			if ok := mintCoin.CheckCoinValid(rewardPaymentAddress, metaResponse.SharedRandom, rewardAmount); !ok {
				err = errors.Errorf("[Mint Withdraw Reward] CheckMintCoinValid: %v, %v, %v, %v, %v, %v\n", mintCoin.GetVersion(), rewardAmount, mintCoin.GetValue(), mintCoin.GetPublicKey(), rewardPaymentAddress, rewardPaymentAddress.GetPublicSpend().String())
				Logger.log.Error(err)
				return errors.Errorf("[Mint Withdraw Reward] Mint Coin is invalid for receiver or amount, err %v", err)
			}
		}
	}

	return nil
}

func (blockchain *BlockChain) ValidateResponseTransactionFromBeaconInstructions(
	curView *ShardBestState,
	shardBlock *types.ShardBlock,
	beaconBlocks []*types.BeaconBlock,
	shardID byte,
) error {
	//mainnet have two block return double when height < REPLACE_STAKINGTX
	if len(beaconBlocks) > 0 && beaconBlocks[0].GetHeight() < config.Param().ReplaceStakingTxHeight {
		return nil
	}
	return blockchain.ValidateReturnStakingTxFromBeaconInstructions(
		curView,
		beaconBlocks,
		shardBlock,
		shardID,
	)
}

func (blockchain *BlockChain) InitTxSalaryByCoinID(
	payToAddress *privacy.PaymentAddress,
	amount uint64,
	payByPrivateKey *privacy.PrivateKey,
	transactionStateDB *statedb.StateDB,
	bridgeStateDB *statedb.StateDB,
	meta *metadata.WithDrawRewardResponse,
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
func (blockchain *BlockChain) BuildResponseTransactionFromTxsWithMetadata(view *ShardBestState, transactions []metadata.Transaction, blkProducerPrivateKey *privacy.PrivateKey, shardID byte) ([]metadata.Transaction, error) {
	withdrawReqTable := make(map[string]metadata.Transaction)
	for _, tx := range transactions {
		if tx.GetMetadataType() == metadata.WithDrawRewardRequestMeta {
			metaReq := tx.GetMetadata().(*metadata.WithDrawRewardRequest)
			mapKey := fmt.Sprintf("%s-%s", base58.Base58Check{}.Encode(metaReq.PaymentAddress.Pk, common.Base58Version), metaReq.TokenID.String())
			if _, ok := withdrawReqTable[mapKey]; !ok {
				withdrawReqTable[mapKey] = tx
			}
		}
	}
	txsResponse := []metadata.Transaction{}
	for _, txRequest := range withdrawReqTable {
		txResponse, err := blockchain.buildWithDrawTransactionResponse(view, &txRequest, blkProducerPrivateKey, shardID)
		if err != nil {
			Logger.log.Errorf("[Withdraw Reward] Build transactions response for tx %v return errors %v", txRequest, err)
			continue
		}
		txsResponse = append(txsResponse, txResponse)
		Logger.log.Infof("[Withdraw Reward] - BuildWithDrawTransactionResponse for tx %+v, ok: %+v\n", txRequest, txResponse)
	}
	return append(transactions, txsResponse...), nil
}

//Return all coins belonging to the provided keyset
//
//If there is a ReadonlyKey, return decrypted coins; otherwise, just return raw coins
func (blockchain *BlockChain) getOutputCoins(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash, upToHeight uint64, versionsIncluded map[int]bool) ([]privacy.PlainCoin, []privacy.Coin, uint64, error) {
	var outCoins []privacy.Coin
	var lowestHeightForV2 uint64 = config.Param().CoinVersion2LowestHeight
	var fromHeight uint64
	if keyset == nil {
		return nil, nil, 0, NewBlockChainError(GetListDecryptedOutputCoinsByKeysetError, fmt.Errorf("invalid key set, got keyset %+v", keyset))
	}
	bss := blockchain.GetBestStateShard(shardID)
	transactionStateDB := blockchain.GetBestStateTransactionStateDB(shardID)

	if versionsIncluded[1] {
		results, err := coinIndexer.QueryDbCoinVer1(keyset.PaymentAddress.Pk, tokenID, transactionStateDB)
		if err != nil {
			return nil, nil, 0, err
		}
		outCoins = append(outCoins, results...)
	}
	if versionsIncluded[2] {
		if keyset.OTAKey.GetOTASecretKey() == nil || keyset.OTAKey.GetPublicSpend() == nil {
			return nil, nil, 0, errors.New("OTA secretKey is needed when retrieving coinV2")
		}
		if keyset.PaymentAddress.GetOTAPublicKey() == nil {
			return nil, nil, 0, errors.New("OTA publicKey is needed when retrieving coinV2")
		}
		latest := bss.ShardHeight
		if upToHeight > latest || upToHeight == 0 {
			upToHeight = latest
		}
		fromHeight = coinIndexer.GetNextLowerHeight(upToHeight, lowestHeightForV2)
		filter := coinIndexer.GetCoinFilterByOTAKeyAndToken()
		results, err := coinIndexer.QueryDbCoinVer2(keyset.OTAKey, tokenID, fromHeight, upToHeight, transactionStateDB, filter)
		if err != nil {
			return nil, nil, 0, err
		}
		// if this is a submitted OTA key and indexing is enabled, "cache" the coins
		if cInd := GetCoinIndexer(); cInd != nil && keyset.OTAKey.GetOTASecretKey() != nil {
			var coinsToStore []privacy.Coin
			if hasKey, _ := cInd.HasOTAKey(coinIndexer.OTAKeyToRaw(keyset.OTAKey)); hasKey {
				for _, c := range results {
					//indexer supports v2 only
					coinsToStore = append(coinsToStore, c)
				}
				_ = cInd.StoreIndexedOutputCoins(keyset.OTAKey, coinsToStore, shardID)
			}
		}
		outCoins = append(outCoins, results...)
	}

	//If ReadonlyKey found, return decrypted coins
	if keyset.ReadonlyKey.GetPrivateView() != nil && keyset.ReadonlyKey.GetPublicSpend() != nil {
		resultPlainCoins := make([]privacy.PlainCoin, 0)
		for _, outCoin := range outCoins {
			decryptedOut, _ := DecryptOutputCoinByKey(transactionStateDB, outCoin, keyset, tokenID, shardID)
			if decryptedOut != nil {
				resultPlainCoins = append(resultPlainCoins, decryptedOut)
			}
		}

		return resultPlainCoins, nil, fromHeight, nil
	} else { //Just return the raw coins
		return nil, outCoins, fromHeight, nil
	}

}

// getTokenBalanceV1 returns the balance v1 (balance of all v1 UTXOs) of a tokenId.
func (blockchain *BlockChain) getTokenBalanceV1(keySet *incognitokey.KeySet, shardID byte, tokenId *common.Hash) (uint64, error) {
	decryptedCoins, unDecryptedCoins, err := blockchain.GetListDecryptedOutputCoinsVer1ByKeyset(keySet, shardID, tokenId)
	if err != nil {
		return 0, err
	}

	if unDecryptedCoins != nil || len(unDecryptedCoins) > 0 {
		return 0, fmt.Errorf("cannot decrypt all coins, something's wrong with the key-set")
	}

	balance := uint64(0)
	transactionDB := blockchain.GetBestStateTransactionStateDB(shardID)
	for _, plainCoin := range decryptedCoins {
		keyImage := plainCoin.GetKeyImage()
		if keyImage == nil {
			continue
		}
		exists, err := statedb.HasSerialNumber(transactionDB, *tokenId, keyImage.ToBytesS(), shardID)
		if err != nil || exists {
			continue
		}
		balance += plainCoin.GetValue()
	}

	return balance, nil
}

func (blockchain *BlockChain) GetListDecryptedOutputCoinsVer2ByKeyset(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash, startHeight uint64) ([]privacy.PlainCoin, []privacy.Coin, uint64, error) {
	return blockchain.getOutputCoins(keyset, shardID, tokenID, startHeight, map[int]bool{2: true})
}

func (blockchain *BlockChain) GetListDecryptedOutputCoinsVer1ByKeyset(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash) ([]privacy.PlainCoin, []privacy.Coin, error) {
	resPlainCoins, resCoins, _, err := blockchain.getOutputCoins(keyset, shardID, tokenID, 0, map[int]bool{1: true})
	return resPlainCoins, resCoins, err
}

//GetListDecryptedOutputCoinsByKeyset - Read all blocks to get txs(not action tx) which can be decrypt by readonly secret key.
//With private-key, we can check unspent tx by check serialNumber from database
//- Param #1: keyset - (priv-key, payment-address, readonlykey)
//in case priv-key: return unspent outputcoin tx
//in case readonly-key: return all outputcoin tx with amount value
//in case payment-address: return all outputcoin tx with no amount value
//- Param #2: coinType - which type of joinsplitdesc(COIN or BOND)
func (blockchain *BlockChain) GetListDecryptedOutputCoinsByKeyset(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash, shardHeight uint64) ([]privacy.PlainCoin, []privacy.Coin, uint64, error) {
	if keyset.OTAKey.GetPublicSpend() == nil || keyset.OTAKey.GetOTASecretKey() == nil || keyset.PaymentAddress.GetOTAPublicKey() == nil {
		return blockchain.getOutputCoins(keyset, shardID, tokenID, shardHeight, map[int]bool{1: true})
	}
	return blockchain.getOutputCoins(keyset, shardID, tokenID, shardHeight, map[int]bool{1: true, 2: true})
}

func (blockchain *BlockChain) SubmitOTAKey(otaKey privacy.OTAKey, accessToken string, isReset bool, heightToSyncFrom uint64) error {
	if !EnableIndexingCoinByOTAKey {
		return fmt.Errorf("OTA key submission not supported by this node configuration")
	}

	otaBytes := coinIndexer.OTAKeyToRaw(otaKey)
	keyExists, state := outcoinIndexer.HasOTAKey(otaBytes)
	if keyExists && !isReset && state != coinIndexer.StatusKeySubmittedUsual {
		return fmt.Errorf("OTAKey %x has been submitted and status = %v", otaBytes, state)
	}

	if accessToken != "" && !outcoinIndexer.IsAuthorizedRunning() {
		return fmt.Errorf("enhanced caching not supported by this node configuration")
	}

	otaKeyStr := fmt.Sprintf("%x", otaBytes)
	Logger.log.Infof("[SubmitOTAKey] otaKey %x, keyExist %v, status %v, isReset %v\n", otaKeyStr, keyExists, state, isReset)

	pkb := otaKey.GetPublicSpend().ToBytesS()
	shardID := common.GetShardIDFromLastByte(pkb[len(pkb)-1])

	if accessToken != "" {
		if outcoinIndexer.IsValidAccessToken(accessToken) { //if the token is authorized
			if outcoinIndexer.IsQueueFull(shardID) {
				return fmt.Errorf("the current authorized queue is full, please check back later")
			}

			bss := blockchain.GetBestStateShard(shardID)
			transactionStateDB := blockchain.GetBestStateTransactionStateDB(shardID)

			lowestHeightForV2 := config.Param().CoinVersion2LowestHeight
			if heightToSyncFrom < lowestHeightForV2 {
				heightToSyncFrom = lowestHeightForV2
			}

			if heightToSyncFrom > bss.ShardHeight {
				return fmt.Errorf("fromHeight (%v) is larger than the current shard height (%v)", heightToSyncFrom, bss.ShardHeight)
			}

			idxParams := coinIndexer.IndexParam{
				FromHeight: heightToSyncFrom,
				ToHeight:   bss.ShardHeight + 1,
				OTAKey:     otaKey,
				TxDb:       transactionStateDB,
				ShardID:    shardID,
				IsReset:    isReset,
			}

			outcoinIndexer.IdxChan <- &idxParams

			// Add the OTAKey to the CoinIndexer here to avoid missing coins when new blocks arrive
			err := outcoinIndexer.AddOTAKey(idxParams.OTAKey, coinIndexer.StatusIndexing)
			if err != nil {
				Logger.log.Errorf("Adding OTAKey %v error: %v\n", coinIndexer.OTAKeyToRaw(idxParams.OTAKey), err)
				return err
			}

			Logger.log.Infof("Authorized OTA Key Submission %x", otaKey)
			return nil
		} else {
			return fmt.Errorf("invalid access token")
		}
	} else {
		Logger.log.Infof("OTA Key Submission %x using the regular cache", otaKey)
		return outcoinIndexer.AddOTAKey(otaKey, coinIndexer.StatusKeySubmittedUsual)
	}
}

func (blockchain *BlockChain) GetKeySubmissionInfo(keyStr string) (int, error) {
	if !EnableIndexingCoinByOTAKey {
		return 0, fmt.Errorf("OTA key submission not supported by this node configuration")
	}

	w, err := wallet.Base58CheckDeserialize(keyStr)
	if err != nil {
		return 0, fmt.Errorf("cannot deserialize key")
	}

	otaKey := w.KeySet.OTAKey
	if otaKey.GetPublicSpend() == nil || otaKey.GetOTASecretKey() == nil {
		return 0, fmt.Errorf("otaKey is not valid")
	}

	otaBytes := coinIndexer.OTAKeyToRaw(otaKey)
	keyExists, state := outcoinIndexer.HasOTAKey(otaBytes)

	if !keyExists {
		return 0, fmt.Errorf("ota key hasn't been submitted")
	}

	return state, nil
}

// GetAllOutputCoinsByKeyset retrieves and tries to decrypt all output coins of a key-set.
//
// Any coins that failed to decrypt are returned as privacy.Coin
func (blockchain *BlockChain) GetAllOutputCoinsByKeyset(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash, withVersion1 bool) ([]privacy.PlainCoin, []privacy.Coin, error) {
	transactionStateDB := blockchain.GetBestStateTransactionStateDB(shardID)

	var err error
	var decryptedResults []privacy.PlainCoin
	var otherResults []privacy.Coin

	// get output coins v1 (if required)
	if withVersion1 {
		decryptedResults, otherResults, _, err = blockchain.getOutputCoins(keyset, shardID, tokenID, 0, map[int]bool{1: true})
		if err != nil {
			return nil, nil, err
		}
	}

	if !EnableIndexingCoinByOTAKey {
		if !withVersion1 {
			return nil, nil, errors.New("Getting all coins not supported by this node configuration")
		} else {
			return decryptedResults, otherResults, nil
		}
	}

	// get output coins v2 from the cache
	outCoins, state, err := outcoinIndexer.GetIndexedOutCoin(keyset.OTAKey, tokenID, transactionStateDB, shardID)
	if err != nil {
		return nil, nil, err
	}
	Logger.log.Infof("current cache state: %v\n", state)
	for _, outCoin := range outCoins {
		// try to decrypt each outCoin. If the privateKey is provided, it also checks if the outCoin is spent or not.
		// If outCoin cannot be decrypted by the keySet, just return the raw outCoin.
		decryptedOut, _ := DecryptOutputCoinByKey(transactionStateDB, outCoin, keyset, tokenID, shardID)
		if decryptedOut != nil {
			decryptedResults = append(decryptedResults, decryptedOut)
		} else {
			otherResults = append(otherResults, outCoin)
		}
	}
	Logger.log.Infof("Retrieved output coins ver2 for view key %v", keyset.OTAKey.GetOTASecretKey())
	return decryptedResults, otherResults, nil
}

// GetAllTokenBalancesV1 returns the balance v1 of all tokens.
func (blockchain *BlockChain) GetAllTokenBalancesV1(keySet *incognitokey.KeySet) (map[string]uint64, error) {
	// get shardID from the keySet
	pubKeyBytes := keySet.PaymentAddress.Pk
	if len(pubKeyBytes) == 0 {
		pubKeyBytes = keySet.OTAKey.GetPublicSpend().ToBytesS()
	}
	if len(pubKeyBytes) == 0 || len(keySet.PrivateKey) == 0 {
		return nil, fmt.Errorf("invalid key set")
	}
	shardID := common.GetShardIDFromLastByte(pubKeyBytes[len(pubKeyBytes)-1])

	// get all token
	allTokens, err := blockchain.ListPrivacyTokenAndBridgeTokenAndPRVByShardID(shardID)
	if err != nil {
		return nil, err
	}

	res := make(map[string]uint64)

	// get balanceV1 for each token
	var balance uint64
	for _, tokenID := range allTokens {
		balance, err = blockchain.getTokenBalanceV1(keySet, shardID, &tokenID)
		if err != nil {
			return nil, err
		}
		res[tokenID.String()] = balance
	}

	return res, nil
}

// GetAllTokenBalancesV2 returns the balance v2 of all tokens.
func (blockchain *BlockChain) GetAllTokenBalancesV2(keySet *incognitokey.KeySet) (map[string]uint64, error) {
	if !EnableIndexingCoinByOTAKey {
		return nil, errors.New("getting v2 coins not supported by this node configuration")
	}

	// get shardID from the keySet
	pubKeyBytes := keySet.PaymentAddress.Pk
	if len(pubKeyBytes) == 0 {
		pubKeyBytes = keySet.OTAKey.GetPublicSpend().ToBytesS()
	}
	if len(pubKeyBytes) == 0 || len(keySet.PrivateKey) == 0 {
		return nil, fmt.Errorf("invalid key set")
	}
	shardID := common.GetShardIDFromLastByte(pubKeyBytes[len(pubKeyBytes)-1])

	// get all token
	allTokens, err := blockchain.ListPrivacyTokenAndBridgeTokenAndPRVByShardID(shardID)
	if err != nil {
		return nil, err
	}

	// create a map from Hash(TokenID) => TokenID
	rawAssetTags := make(map[string]*common.Hash)
	for _, tokenID := range allTokens {
		clonedTokenID, err := new(common.Hash).NewHash(tokenID[:])
		if err != nil {
			return nil, err
		}
		rawAssetTags[privacy.HashToPoint(clonedTokenID[:]).String()] = clonedTokenID
	}

	res := make(map[string]uint64)

	// get all token output coins v2 using the generic common.ConfidentialAssetID
	decryptedOutCoins, err := blockchain.TryGetAllOutputCoinsByKeyset(keySet, shardID, &common.ConfidentialAssetID, false)
	if err != nil {
		return nil, err
	}

	// try to get the tokenId of each output coin from its assetTag.
	for _, outCoin := range decryptedOutCoins {
		outCoinV2, ok := outCoin.(*privacy.CoinV2)
		if !ok {
			return nil, fmt.Errorf("cannot parse outCoin as a CoinV2")
		}

		tokenId, err := outCoinV2.GetTokenId(keySet, rawAssetTags)
		if err != nil {
			return nil, err
		}

		res[tokenId.String()] += outCoin.GetValue()
	}

	return res, nil
}

// TryGetAllOutputCoinsByKeyset gets and decrypts all output coins of a tokenID given the keySet.
// Any coins that are failed to decrypt are skipped.
func (blockchain *BlockChain) TryGetAllOutputCoinsByKeyset(keyset *incognitokey.KeySet, shardID byte, tokenID *common.Hash, withVersion1 bool) ([]privacy.PlainCoin, error) {
	res, _, err := blockchain.GetAllOutputCoinsByKeyset(keyset, shardID, tokenID, withVersion1)
	return res, err
}

// CreateAndSaveTxViewPointFromBlock - fetch data from block, put into txviewpoint variable and save into db
// still storage full data of commitments, serial number, snderivator to check double spend
// this function only work for transaction transfer token/prv within shard
func (blockchain *BlockChain) CreateAndSaveTxViewPointFromBlock(shardBlock *types.ShardBlock, transactionStateRoot *statedb.StateDB) error {
	// Fetch data from shardBlock into tx View point
	if shardBlock.Header.Height == 1 {
		err := storePRV(transactionStateRoot)
		if err != nil {
			return err
		}
	}
	var err error
	bridgeStateDB := blockchain.GetBeaconBestState().GetBeaconFeatureStateDB()
	view := NewTxViewPoint(shardBlock.Header.ShardID, shardBlock.Header.BeaconHeight, shardBlock.Header.Height)
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
		tokenData := privacyCustomTokenTx.GetTxTokenData()
		isBridgeToken, err := statedb.IsBridgeToken(bridgeStateDB, tokenData.PropertyID)
		if err != nil {
			return err
		}
		switch tokenData.Type {
		case transaction.CustomTokenInit:
			{
				tokenID := tokenData.PropertyID

				// Add the tokenID to the cache for fast retrieval.
				if EnableIndexingCoinByOTAKey {
					outcoinIndexer.AddTokenID(tokenID)
				}

				existed := statedb.PrivacyTokenIDExisted(transactionStateRoot, tokenID)
				if !existed {
					// check is bridge token
					tokenID := tokenData.PropertyID
					name := tokenData.PropertyName
					symbol := tokenData.PropertySymbol
					mintable := tokenData.Mintable
					amount := tokenData.Amount
					info := privacyCustomTokenTx.GetInfo()
					txHash := *privacyCustomTokenTx.Hash()
					tokenType := statedb.InitToken
					if isBridgeToken {
						tokenType = statedb.BridgeToken
					}
					Logger.log.Info("Store custom token when it is issued", tokenData.PropertyID, tokenData.PropertySymbol, tokenData.PropertyName)
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
		err = statedb.StorePrivacyTokenTx(transactionStateRoot, tokenData.PropertyID, *privacyCustomTokenTx.Hash())
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

		err = blockchain.StoreOnetimeAddressesFromTxViewPoint(transactionStateRoot, *privacyCustomTokenSubView, shardBlock.Header.ShardID)
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

	err = blockchain.StoreOnetimeAddressesFromTxViewPoint(transactionStateRoot, *view, shardBlock.Header.ShardID)
	if err != nil {
		return err
	}

	err = blockchain.StoreSNDerivatorsFromTxViewPoint(transactionStateRoot, *view)
	if err != nil {
		return err
	}

	err = blockchain.StoreTxByPublicKey(blockchain.GetShardChainDatabase(shardBlock.Header.ShardID), view)
	if err != nil {
		return err
	}

	err = blockchain.StoreTxBySerialNumber(shardBlock.Body.Transactions, shardBlock.Header.ShardID)
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

func (blockchain *BlockChain) StoreOnetimeAddressesFromTxViewPoint(stateDB *statedb.StateDB, view TxViewPoint, shardID byte) error {
	// commitment and output are the same key in map
	keys := make([]string, 0, len(view.mapCommitments))
	for k := range view.mapCommitments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, decl := range view.otaDeclarations {
		if err := statedb.StoreOccupiedOneTimeAddress(stateDB, decl.TokenID, decl.PublicKey); err != nil {
			return err
		}
	}

	// Start to store to db
	for _, publicKey := range keys {
		publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
		if err != nil {
			return err
		}
		if (view.beaconHeight >= config.Param().ConsensusParam.NotUseBurnedCoins) && common.IsPublicKeyBurningAddress(publicKeyBytes) {
			continue
		}
		senderShardID, recvShardID, _, _ := privacy.DeriveShardInfoFromCoin(publicKeyBytes)
		if recvShardID == int(shardID) {
			// outputs
			outputCoinArray := view.mapOutputCoins[publicKey]
			otaCoinArray := make([][]byte, 0)
			onetimeAddressArray := make([][]byte, 0)
			for _, outputCoin := range outputCoinArray {
				if outputCoin.GetVersion() != 2 {
					continue
				}
				if EnableIndexingCoinByOTAKey {
					handler := func(k, v interface{}) bool {
						vkArr, ok := k.([64]byte)
						if !ok {
							return false
						}
						processing, ok := v.(int)
						if !ok {
							return false
						}
						if processing == 0 {
							return false
						}
						otaKey := coinIndexer.OTAKeyFromRaw(vkArr)
						ks := &incognitokey.KeySet{}
						ks.OTAKey = otaKey
						belongs, _ := outputCoin.DoesCoinBelongToKeySet(ks)
						if belongs {
							err = outcoinIndexer.StoreIndexedOutputCoins(otaKey, []privacy.Coin{outputCoin}, shardID)
							if err != nil {
								Logger.log.Errorf("StoreIndexedOutputCoins in viewpoint for OTAKey %x error: %v\n", vkArr, err)
							}
						}
						return true
					}
					outcoinIndexer.GetManagedOTAKeys().Range(handler)
				}
				otaCoinArray = append(otaCoinArray, outputCoin.Bytes())
				onetimeAddressArray = append(onetimeAddressArray, outputCoin.GetPublicKey().ToBytesS())
			}
			if err = statedb.StoreOTACoinsAndOnetimeAddresses(stateDB, *view.tokenID, view.height, otaCoinArray, onetimeAddressArray, shardID); err != nil {
				return err
			}
		} else if view.height >= config.Param().BCHeightBreakPointCoinOrigin {
			if senderShardID == int(shardID) {
				var b [32]byte
				copy(b[:], publicKeyBytes)
				err := statedb.StoreOccupiedOneTimeAddress(stateDB, *view.tokenID, b)
				if err != nil {
					return err
				}
			}
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
	for _, publicKey := range keys {
		publicKeyBytes, _, err := base58.Base58Check{}.Decode(publicKey)
		if err != nil {
			return err
		}
		if (view.beaconHeight >= config.Param().ConsensusParam.NotUseBurnedCoins) && common.IsPublicKeyBurningAddress(publicKeyBytes) {
			continue
		}
		publicKeyShardID := common.GetShardIDFromLastByte(publicKeyBytes[len(publicKeyBytes)-1])
		if publicKeyShardID == shardID {
			// outputs
			outputCoinArray := view.mapOutputCoins[publicKey]
			outputCoinBytesArray := make([][]byte, 0)
			for _, outputCoin := range outputCoinArray {
				if outputCoin.GetVersion() == 1 {
					outputCoinBytesArray = append(outputCoinBytesArray, outputCoin.Bytes())
				}
			}
			err = statedb.StoreOutputCoins(stateDB, *view.tokenID, publicKeyBytes, outputCoinBytesArray, publicKeyShardID)
			if err != nil {
				return err
			}

			// commitment
			commitmentsArray := view.mapCommitments[publicKey]
			err = statedb.StoreCommitments(stateDB, *view.tokenID, commitmentsArray, view.shardID)
			if err != nil {
				return err
			}

			//Logger.log.Infof("BUGLOG4 finished storing %v cmts of pk %v\n", len(commitmentsArray), publicKey)

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
		}
	}
	return nil
}

func (blockchain *BlockChain) CreateAndSaveCrossTransactionViewPointFromBlock(shardBlock *types.ShardBlock, transactionStateRoot *statedb.StateDB) error {
	Logger.log.Critical("Fetch Cross transaction", shardBlock.Body.CrossTransactions)
	// Fetch data from block into tx View point
	view := NewTxViewPoint(shardBlock.Header.ShardID, shardBlock.Header.BeaconHeight, shardBlock.Header.Height)
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
			Logger.log.Info("Cross-shard tx: store custom token when it is issued ", tokenID, privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertyName, privacyCustomTokenSubView.privacyCustomTokenMetadata.PropertySymbol, privacyCustomTokenSubView.privacyCustomTokenMetadata.Amount, privacyCustomTokenSubView.privacyCustomTokenMetadata.Mintable)
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

		err = blockchain.StoreOnetimeAddressesFromTxViewPoint(transactionStateRoot, *privacyCustomTokenSubView, shardBlock.Header.ShardID)
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

	// store otas
	err = blockchain.StoreOnetimeAddressesFromTxViewPoint(transactionStateRoot, *view, shardBlock.Header.ShardID)
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

func (blockchain *BlockChain) StoreTxBySerialNumber(txList []metadata.Transaction, shardID byte) error {
	var err error
	db := blockchain.GetShardChainDatabase(shardID)

	for _, tx := range txList {
		//if tx.GetVersion() < 2 {//Only process txver2
		//	continue
		//}
		txHash := *tx.Hash()
		tokenID := *tx.GetTokenID()
		Logger.log.Infof("Process StoreTxBySerialNumber for tx %v, tokenID %v\n", txHash.String(), tokenID.String())

		if tokenID.String() != common.PRVIDStr {
			txToken, ok := tx.(transaction.TransactionToken)
			if !ok {
				return fmt.Errorf("cannot parse tx %v to transactionToken", txHash.String())
			}

			txFee := txToken.GetTxBase()
			txNormal := txToken.GetTxNormal()
			//Process storing serialNumber for PRV
			if txFee.GetProof() != nil {
				for _, inputCoin := range txFee.GetProof().GetInputCoins() {
					serialNumber := inputCoin.GetKeyImage().ToBytesS()
					err = rawdbv2.StoreTxBySerialNumber(db, serialNumber, common.PRVCoinID, shardID, txHash)
					if err != nil {
						Logger.log.Errorf("StoreTxBySerialNumber with serialNumber %v, tokenID %v, shardID %v, txHash %v returns an error: %v\n", serialNumber, common.PRVCoinID.String(), shardID, txHash.String())
						return err
					}
				}
			} else {
				Logger.log.Infof("txFee of %v has no proof\n", txHash.String())
			}

			//Process storing serialNumber for token
			if txNormal.GetProof() != nil {
				for _, inputCoin := range txNormal.GetProof().GetInputCoins() {
					serialNumber := inputCoin.GetKeyImage().ToBytesS()
					err = rawdbv2.StoreTxBySerialNumber(db, serialNumber, tokenID, shardID, txHash)
					if err != nil {
						Logger.log.Errorf("StoreTxBySerialNumber with serialNumber %v, tokenID %v, shardID %v, txHash %v returns an error: %v\n", serialNumber, tokenID.String(), shardID, txHash.String())
						return err
					}
				}
			} else {
				Logger.log.Infof("txToken of %v has no proof\n", txHash.String())
			}
		} else {
			if tx.GetProof() != nil {
				for _, inputCoin := range tx.GetProof().GetInputCoins() {
					serialNumber := inputCoin.GetKeyImage().ToBytesS()
					err = rawdbv2.StoreTxBySerialNumber(db, serialNumber, tokenID, shardID, txHash)
					if err != nil {
						Logger.log.Errorf("StoreTxBySerialNumber with serialNumber %v, tokenID %v, shardID %v, txHash %v returns an error: %v\n", serialNumber, tokenID.String(), shardID, txHash.String())
						return err
					}
				}
			} else {
				Logger.log.Infof("tx %v has no proof\n", txHash.String())
			}
		}
	}

	Logger.log.Infof("Finish StoreTxBySerialNumber, #txs: %v!!!\n", len(txList))

	return nil
}
