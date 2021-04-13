package types

import (
	"encoding/binary"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type ContentCrossShardTokenPrivacyData struct {
	OutputCoin     []privacy.OutputCoin
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string
	Type           int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
}

func (contentCrossShardTokenPrivacyData ContentCrossShardTokenPrivacyData) Bytes() []byte {
	res := []byte{}
	for _, item := range contentCrossShardTokenPrivacyData.OutputCoin {
		res = append(res, item.Bytes()...)
	}
	res = append(res, contentCrossShardTokenPrivacyData.PropertyID.GetBytes()...)
	res = append(res, []byte(contentCrossShardTokenPrivacyData.PropertyName)...)
	res = append(res, []byte(contentCrossShardTokenPrivacyData.PropertySymbol)...)
	typeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(typeBytes, uint32(contentCrossShardTokenPrivacyData.Type))
	res = append(res, typeBytes...)
	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(amountBytes, uint32(contentCrossShardTokenPrivacyData.Amount))
	res = append(res, amountBytes...)
	if contentCrossShardTokenPrivacyData.Mintable {
		res = append(res, []byte("true")...)
	} else {
		res = append(res, []byte("false")...)
	}
	return res
}

func (contentCrossShardTokenPrivacyData ContentCrossShardTokenPrivacyData) Hash() common.Hash {
	return common.HashH(contentCrossShardTokenPrivacyData.Bytes())
}

func CalHashTxTokenPrivacyDataHashList(txTokenPrivacyDataList []ContentCrossShardTokenPrivacyData) common.Hash {
	tmpByte := []byte{}
	if len(txTokenPrivacyDataList) != 0 {
		for _, txTokenPrivacyData := range txTokenPrivacyDataList {
			tempHash := txTokenPrivacyData.Hash()
			tmpByte = append(tmpByte, tempHash.GetBytes()...)

		}
	} else {
		return common.HashH([]byte(""))
	}
	return common.HashH(tmpByte)
}
