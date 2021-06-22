package types

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

type CrossTokenPrivacyData struct {
	BlockHeight      uint64
	BlockHash        common.Hash
	TokenPrivacyData []ContentCrossShardTokenPrivacyData
}

func CloneTxTokenPrivacyDataForCrossShard(txTokenPrivacyData transaction.TxTokenData) ContentCrossShardTokenPrivacyData {
	newContentCrossTokenPrivacyData := ContentCrossShardTokenPrivacyData{
		PropertyID:     txTokenPrivacyData.PropertyID,
		PropertyName:   txTokenPrivacyData.PropertyName,
		PropertySymbol: txTokenPrivacyData.PropertySymbol,
		Mintable:       txTokenPrivacyData.Mintable,
		Amount:         txTokenPrivacyData.Amount,
		Type:           transaction.CustomTokenCrossShard,
	}
	newContentCrossTokenPrivacyData.OutputCoin = []privacy.Coin{}
	return newContentCrossTokenPrivacyData
}
