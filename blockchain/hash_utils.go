package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/transaction"
)

//=========================HASH util==================================
func GenerateZeroValueHash() (common.Hash, error) {
	hash := common.Hash{}
	hash.SetBytes(make([]byte, 32))
	return hash, nil
}
func GenerateHashFromHashArray(hashes []common.Hash) (common.Hash, error) {
	// if input is empty list
	// return hash value of bytes zero
	if len(hashes) == 0 {
		return GenerateZeroValueHash()
	}
	strs := []string{}
	for _, value := range hashes {
		str := value.String()
		strs = append(strs, str)
	}
	return GenerateHashFromStringArray(strs)
}

func GenerateHashFromStringArray(strs []string) (common.Hash, error) {
	// if input is empty list
	// return hash value of bytes zero
	if len(strs) == 0 {
		return GenerateZeroValueHash()
	}
	var (
		hash common.Hash
		buf  bytes.Buffer
	)
	for _, value := range strs {
		buf.WriteString(value)
	}
	temp := sha256.Sum256(buf.Bytes())
	if err := hash.SetBytes(temp[:]); err != nil {
		return common.Hash{}, NewBlockChainError(HashError, err)
	}
	return hash, nil
}

func GenerateHashFromMapByteString(maps1 map[byte][]string, maps2 map[byte][]string) (common.Hash, error) {
	var keys1 []int
	for k := range maps1 {
		keys1 = append(keys1, int(k))
	}
	sort.Ints(keys1)
	shardPendingValidator := []string{}
	// To perform the opertion you want
	for _, k := range keys1 {
		shardPendingValidator = append(shardPendingValidator, maps1[byte(k)]...)
	}

	var keys2 []int
	for k := range maps2 {
		keys2 = append(keys2, int(k))
	}
	sort.Ints(keys2)
	shardValidator := []string{}
	// To perform the opertion you want
	for _, k := range keys2 {
		shardValidator = append(shardValidator, maps2[byte(k)]...)
	}
	return GenerateHashFromStringArray(append(shardPendingValidator, shardValidator...))
}

func GenerateHashFromShardState(allShardState map[byte][]ShardState) (common.Hash, error) {
	allShardStateStr := []string{}
	var keys []int
	for k := range allShardState {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		res := ""
		for _, shardState := range allShardState[byte(shardID)] {
			res += strconv.Itoa(int(shardState.Height))
			res += shardState.Hash.String()
			crossShard, _ := json.Marshal(shardState.CrossShard)
			res += string(crossShard)
		}
		allShardStateStr = append(allShardStateStr, res)
	}
	return GenerateHashFromStringArray(allShardStateStr)
}
func VerifyHashFromHashArray(hashes []common.Hash, hash common.Hash) bool {
	strs := []string{}
	for _, value := range hashes {
		str := value.String()
		strs = append(strs, str)
	}
	return VerifyHashFromStringArray(strs, hash)
}

func VerifyHashFromStringArray(strs []string, hash common.Hash) bool {
	res, err := GenerateHashFromStringArray(strs)
	if err != nil {
		return false
	}
	return bytes.Equal(res.GetBytes(), hash.GetBytes())
}

func VerifyHashFromMapByteString(maps1 map[byte][]string, maps2 map[byte][]string, hash common.Hash) bool {
	res, err := GenerateHashFromMapByteString(maps1, maps2)
	if err != nil {
		return false
	}
	return bytes.Equal(res.GetBytes(), hash.GetBytes())
}

func VerifyHashFromShardState(allShardState map[byte][]ShardState, hash common.Hash) bool {
	res, err := GenerateHashFromShardState(allShardState)
	if err != nil {
		return false
	}
	return bytes.Equal(res.GetBytes(), hash.GetBytes())
}
func calHashFromTxTokenDataList(txTokenDataList []transaction.TxTokenData) (common.Hash, error) {
	hashes := []common.Hash{}
	sort.SliceStable(txTokenDataList[:], func(i, j int) bool {
		return txTokenDataList[i].PropertyID.String() < txTokenDataList[j].PropertyID.String()
	})
	for _, txTokenData := range txTokenDataList {
		hash, err := txTokenData.Hash()
		if err != nil {
			return common.Hash{}, err
		}
		hashes = append(hashes, *hash)
	}
	hash, err := GenerateHashFromHashArray(hashes)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}
