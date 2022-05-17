package blockchain

import (
	"errors"
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

func FetchBeaconBlockFromHeight(blockchain *BlockChain, from uint64, to uint64) ([]*types.BeaconBlock, error) {
	beaconBlocks := []*types.BeaconBlock{}
	if (config.Config().SyncMode == common.STATEDB_LITE_MODE) && (from != 1) && (to != 1) {
		var (
			err    error
			bcBlks []types.BeaconBlock
		)
		for i := 0; i < 5; i++ {
			bcBlks, err = blockchain.GetConfig().Syncker.ReSyncBeaconBlockByHeight(from, to)
			if err != nil {
				Logger.log.Error(err)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			for _, bcBlk := range bcBlks {
				tmp := bcBlk
				beaconBlocks = append(beaconBlocks, &tmp)
			}
			break
		}
		if err != nil {
			return nil, err
		}
	} else {
		for i := from; i <= to; i++ {
			beaconHash, err := blockchain.GetBeaconBlockHashByHeight(blockchain.BeaconChain.GetFinalView(), blockchain.BeaconChain.GetBestView(), i)
			if err != nil {
				return nil, err
			}
			beaconBlock, _, err := blockchain.GetBeaconBlockByHash(*beaconHash)
			if err != nil {
				return beaconBlocks, err
			}
			beaconBlocks = append(beaconBlocks, beaconBlock)
		}
	}
	return beaconBlocks, nil
}

func CreateCrossShardByteArray(txList []metadata.Transaction, fromShardID byte) ([]byte, error) {
	crossIDs := []byte{}
	byteMap := make([]byte, common.MaxShardNumber)
	for _, tx := range txList {
		var prvProof privacy.Proof
		switch tx.GetType() {
		case common.TxCustomTokenPrivacyType, common.TxTokenConversionType:
			{
				customTokenTx, ok := tx.(transaction.TransactionToken)
				if !ok {
					return nil, errors.New("Cannot cast transaction")
				}
				prvProof = customTokenTx.GetTxBase().GetProof()
				proof := customTokenTx.GetTxTokenData().TxNormal.GetProof()
				if proof != nil {
					for _, outCoin := range proof.GetOutputCoins() {
						if shardID, err := outCoin.GetShardID(); err != nil {
							panic("Cannot get shardID")
						} else {
							byteMap[common.GetShardIDFromLastByte(shardID)] = 1
						}
					}
				}
			}
		default:
			prvProof = tx.GetProof()
		}
		if prvProof != nil {
			for _, outCoin := range prvProof.GetOutputCoins() {
				shardID, err := outCoin.GetShardID()
				if err == nil {
					byteMap[common.GetShardIDFromLastByte(shardID)] = 1
				}
			}
		}
	}

	for k := range byteMap {
		if byteMap[k] == 1 && k != int(fromShardID) {
			crossIDs = append(crossIDs, byte(k))
		}
	}
	return crossIDs, nil
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

func UpdateTxEnvWithSView(sView *ShardBestState, tx metadata.Transaction) metadata.ValidationEnviroment {
	valEnv := transaction.WithShardHeight(tx.GetValidationEnv(), sView.GetHeight())
	valEnv = transaction.WithBeaconHeight(valEnv, sView.GetBeaconHeight())
	valEnv = transaction.WithConfirmedTime(valEnv, sView.GetBlockTime())
	return valEnv
}
