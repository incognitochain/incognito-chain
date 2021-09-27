package rawdbv2

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"math/big"
)

var decoyPrefix = []byte("tx-decoy")

func GetStoreTxDecoyPrefix(tokenId common.Hash, otaIdx uint64) []byte {
	temp := make([]byte, 0, len(decoyPrefix))
	temp = append(temp, decoyPrefix...)
	temp = append(temp, tokenId.Bytes()...)

	idxBig := new(big.Int).SetUint64(otaIdx)
	return append(temp, idxBig.Bytes()...)
}

func GetStoreTxDecoyKey(tokenId common.Hash, otaIdx uint64, txHash common.Hash, count uint64) []byte {
	key := GetStoreTxDecoyPrefix(tokenId, otaIdx)

	key = append(key, txHash.GetBytes()...)
	countBig := new(big.Int).SetUint64(count)
	return append(key, countBig.Bytes()...)
}

func StoreTxDecoy(db incdb.Database, tokenId common.Hash, otaIdx uint64, txHash common.Hash, count uint64) error {
	if tokenId.String() != common.PRVIDStr {
		tokenId = common.ConfidentialAssetID
	}
	key := GetStoreTxDecoyKey(tokenId, otaIdx, txHash, count)
	value := []byte{}
	if err := db.Put(key, value); err != nil {
		return NewRawdbError(StoreTxByPublicKeyError, err, tokenId.String(), txHash.String(), otaIdx)
	}
	return nil
}

func GetTxByDecoyIndex(db incdb.Database, tokenId common.Hash, otaIdx uint64) (map[string]uint64, error) {
	if tokenId.String() != common.PRVIDStr {
		tokenId = common.ConfidentialAssetID
	}
	prefix := GetStoreTxDecoyPrefix(tokenId, otaIdx)
	iterator := db.NewIteratorWithPrefix(GetStoreTxDecoyPrefix(tokenId, otaIdx))
	result := make(map[string]uint64)
	for iterator.Next() {
		key := iterator.Key()
		tempKey := make([]byte, len(key))
		copy(tempKey, key)

		countBig := new(big.Int)
		txID := common.Hash{}
		start := len(prefix)
		err := txID.SetBytes(tempKey[start:start+32])
		if err != nil {
			return nil, NewRawdbError(GetTxByPublicKeyError, err, tokenId)
		}
		countBig.SetBytes(tempKey[start+32:])
		result[txID.String()] = countBig.Uint64()
	}
	return result, nil
}