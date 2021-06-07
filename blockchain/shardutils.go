package blockchain

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

func FetchBeaconBlockFromHeight(blockchain *BlockChain, from uint64, to uint64) ([]*types.BeaconBlock, error) {
	beaconBlocks := []*types.BeaconBlock{}
	for i := from; i <= to; i++ {
		beaconHash, err := blockchain.GetBeaconBlockHashByHeight(blockchain.BeaconChain.GetFinalView(), blockchain.BeaconChain.GetBestView(), i)
		if err != nil {
			return nil, err
		}
		beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(blockchain.GetBeaconChainDatabase(), *beaconHash)
		if err != nil {
			return beaconBlocks, err
		}
		beaconBlock := types.BeaconBlock{}
		err = json.Unmarshal(beaconBlockBytes, &beaconBlock)
		if err != nil {
			return beaconBlocks, NewBlockChainError(UnmashallJsonShardBlockError, err)
		}
		beaconBlocks = append(beaconBlocks, &beaconBlock)
	}
	return beaconBlocks, nil
}

func CreateCrossShardByteArray(txList []metadata.Transaction, fromShardID byte) []byte {
	crossIDs := []byte{}
	byteMap := make([]byte, common.MaxShardNumber)
	for _, tx := range txList {
		if tx.GetProof() != nil {
			for _, outCoin := range tx.GetProof().GetOutputCoins() {
				lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
				shardID := common.GetShardIDFromLastByte(lastByte)
				byteMap[common.GetShardIDFromLastByte(shardID)] = 1
			}
		}

		switch tx.GetType() {
		case common.TxCustomTokenPrivacyType:
			{
				customTokenTx := tx.(*transaction.TxCustomTokenPrivacy)
				if customTokenTx.TxPrivacyTokenData.TxNormal.GetProof() != nil {
					for _, outCoin := range customTokenTx.TxPrivacyTokenData.TxNormal.GetProof().GetOutputCoins() {
						lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
						shardID := common.GetShardIDFromLastByte(lastByte)
						byteMap[common.GetShardIDFromLastByte(shardID)] = 1
					}
				}
			}
		}
	}

	for k := range byteMap {
		if byteMap[k] == 1 && k != int(fromShardID) {
			crossIDs = append(crossIDs, byte(k))
		}
	}
	return crossIDs
}

func checkReturnStakingTxExistence(txId string, shardBlock *types.ShardBlock) bool {
	for _, tx := range shardBlock.Body.Transactions {
		if tx.GetMetadata() != nil {
			if tx.GetMetadata().GetType() == metadata.ReturnStakingMeta {
				if returnStakingMeta, ok := tx.GetMetadata().(*metadata.ReturnStakingMetadata); ok {
					if returnStakingMeta.TxID == txId {
						return true
					}
				}
			}
		}
	}
	return false
}

func getRequesterFromPKnCoinID(pk privacy.PublicKey, coinID common.Hash) string {
	requester := base58.Base58Check{}.Encode(pk, common.Base58Version)
	return fmt.Sprintf("%s-%s", requester, coinID.String())
}

func reqTableFromReqTxs(
	transactions []metadata.Transaction,
) map[string]metadata.Transaction {
	txRequestTable := map[string]metadata.Transaction{}
	for _, tx := range transactions {
		if tx.GetMetadataType() == metadata.WithDrawRewardRequestMeta {
			requestMeta := tx.GetMetadata().(*metadata.WithDrawRewardRequest)
			key := getRequesterFromPKnCoinID(requestMeta.PaymentAddress.Pk, requestMeta.TokenID)
			txRequestTable[key] = tx
		}
	}
	return txRequestTable
}

func filterReqTxs(
	transactions []metadata.Transaction,
	txRequestTable map[string]metadata.Transaction,
) []metadata.Transaction {
	res := []metadata.Transaction{}
	for _, tx := range transactions {
		if tx.GetMetadataType() == metadata.WithDrawRewardRequestMeta {
			requestMeta := tx.GetMetadata().(*metadata.WithDrawRewardRequest)
			key := getRequesterFromPKnCoinID(requestMeta.PaymentAddress.Pk, requestMeta.TokenID)
			txReq, ok := txRequestTable[key]
			if !ok {
				continue
			}
			cmp, err := txReq.Hash().Cmp(tx.Hash())
			if (err != nil) || (cmp != 0) {
				continue
			}
		}
		res = append(res, tx)
	}
	return res
}

func CreateMerkleCrossTransaction(crossTransactions map[byte][]types.CrossTransaction) (*common.Hash, error) {
	if len(crossTransactions) == 0 {
		res, err := generateZeroValueHash()
		return &res, err
	}
	keys := []int{}
	crossTransactionHashes := []*common.Hash{}
	for k := range crossTransactions {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range crossTransactions[byte(shardID)] {
			hash := value.Hash()
			crossTransactionHashes = append(crossTransactionHashes, &hash)
		}
	}
	merkle := types.Merkle{}
	merkleTree := merkle.BuildMerkleTreeOfHashes(crossTransactionHashes, len(crossTransactionHashes))
	return merkleTree[len(merkleTree)-1], nil
}

func VerifyMerkleCrossTransaction(crossTransactions map[byte][]types.CrossTransaction, rootHash common.Hash) bool {
	res, err := CreateMerkleCrossTransaction(crossTransactions)
	if err != nil {
		return false
	}
	hashByte := rootHash.GetBytes()
	newHash, err := common.Hash{}.NewHash(hashByte)
	if err != nil {
		return false
	}
	return newHash.IsEqual(res)
}

//updateCommitteesWithAddedAndRemovedListValidator :
func updateCommitteesWithAddedAndRemovedListValidator(
	source,
	addedCommittees []incognitokey.CommitteePublicKey) ([]incognitokey.CommitteePublicKey, error) {
	newShardPendingValidator := []incognitokey.CommitteePublicKey{}
	m := make(map[string]bool)
	for _, v := range source {
		str, err := v.ToBase58()
		if err != nil {
			return nil, err
		}
		if m[str] == false {
			newShardPendingValidator = append(newShardPendingValidator, v)
		}
	}
	newShardPendingValidator = append(newShardPendingValidator, addedCommittees...)

	return newShardPendingValidator, nil
}
