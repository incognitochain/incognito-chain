package blockchain

import (
	"errors"
	"sort"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

// getCrossShardData get cross data (send to a shard) from list of transaction:
// 1. (Privacy) PRV: Output coin
// 2. Tx Custom Token: Tx Token Data
// 3. Privacy Custom Token: Token Data + Output coin
func getCrossShardData(txList []metadata.Transaction, shardID byte) ([]privacy.OutputCoin, []types.ContentCrossShardTokenPrivacyData) {
	coinList := []privacy.OutputCoin{}
	txTokenPrivacyDataMap := make(map[common.Hash]*types.ContentCrossShardTokenPrivacyData)
	var txTokenPrivacyDataList []types.ContentCrossShardTokenPrivacyData
	for _, tx := range txList {
		if tx.GetProof() != nil {
			for _, outCoin := range tx.GetProof().GetOutputCoins() {
				lastByte := common.GetShardIDFromLastByte(outCoin.CoinDetails.GetPubKeyLastByte())
				if lastByte == shardID {
					coinList = append(coinList, *outCoin)
				}
			}
		}
		if tx.GetType() == common.TxCustomTokenPrivacyType {
			customTokenPrivacyTx := tx.(*transaction.TxCustomTokenPrivacy)
			if customTokenPrivacyTx.TxPrivacyTokenData.TxNormal.GetProof() != nil {
				for _, outCoin := range customTokenPrivacyTx.TxPrivacyTokenData.TxNormal.GetProof().GetOutputCoins() {
					lastByte := common.GetShardIDFromLastByte(outCoin.CoinDetails.GetPubKeyLastByte())
					if lastByte == shardID {
						if _, ok := txTokenPrivacyDataMap[customTokenPrivacyTx.TxPrivacyTokenData.PropertyID]; !ok {
							contentCrossTokenPrivacyData := types.CloneTxTokenPrivacyDataForCrossShard(customTokenPrivacyTx.TxPrivacyTokenData)
							txTokenPrivacyDataMap[customTokenPrivacyTx.TxPrivacyTokenData.PropertyID] = &contentCrossTokenPrivacyData
						}
						txTokenPrivacyDataMap[customTokenPrivacyTx.TxPrivacyTokenData.PropertyID].OutputCoin = append(txTokenPrivacyDataMap[customTokenPrivacyTx.TxPrivacyTokenData.PropertyID].OutputCoin, *outCoin)
					}
				}
			}
		}
	}
	if len(txTokenPrivacyDataMap) != 0 {
		for _, value := range txTokenPrivacyDataMap {
			txTokenPrivacyDataList = append(txTokenPrivacyDataList, *value)
		}
		sort.SliceStable(txTokenPrivacyDataList[:], func(i, j int) bool {
			return txTokenPrivacyDataList[i].PropertyID.String() < txTokenPrivacyDataList[j].PropertyID.String()
		})
	}
	return coinList, txTokenPrivacyDataList
}

// VerifyCrossShardBlock Verify CrossShard Block
//- Agg Signature
//- MerklePath
func VerifyCrossShardBlock(crossShardBlock *types.CrossShardBlock, blockchain *BlockChain, committees []incognitokey.CommitteePublicKey) error {
	committeesForSigning, err := blockchain.getCommitteesForSigning(committees, crossShardBlock)
	if err != nil {
		return err
	}
	if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(crossShardBlock, committeesForSigning); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := types.VerifyCrossShardBlockUTXO(crossShardBlock); !ok {
		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
	}
	return nil
}
