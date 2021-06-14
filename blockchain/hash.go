package blockchain

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
)

func generateZeroValueHash() (common.Hash, error) {
	hash := common.Hash{}
	hash.SetBytes(make([]byte, 32))
	return hash, nil
}

func generateHashFromStringArray(strs []string) (common.Hash, error) {
	// if input is empty list
	// return hash value of bytes zero
	if len(strs) == 0 {
		return generateZeroValueHash()
	}
	var (
		hash common.Hash
		buf  bytes.Buffer
	)
	for _, value := range strs {
		buf.WriteString(value)
	}
	temp := common.HashB(buf.Bytes())
	if err := hash.SetBytes(temp[:]); err != nil {
		return common.Hash{}, NewBlockChainError(HashError, err)
	}
	return hash, nil
}

func generateHashFromMapStringString(maps1 map[string]string) (common.Hash, error) {
	var keys []string
	var res []string
	for k := range maps1 {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		res = append(res, key)
		res = append(res, maps1[key])
	}
	return generateHashFromStringArray(res)
}

func generateHashFromShardState(allShardState map[byte][]types.ShardState, version int) (common.Hash, error) {
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
			if version != committeestate.SELF_SWAP_SHARD_VERSION {
				res += shardState.ValidationData
				res += shardState.CommitteeFromBlock.String()
			}
		}
		allShardStateStr = append(allShardStateStr, res)
	}
	return generateHashFromStringArray(allShardStateStr)
}

func verifyHashFromStringArray(strs []string, hash common.Hash) (common.Hash, bool) {
	res, err := generateHashFromStringArray(strs)
	if err != nil {
		return common.Hash{}, false
	}
	return res, bytes.Equal(res.GetBytes(), hash.GetBytes())
}

func verifyHashFromShardState(allShardState map[byte][]types.ShardState, hash common.Hash, version int) bool {
	res, err := generateHashFromShardState(allShardState, version)
	if err != nil {
		return false
	}
	return bytes.Equal(res.GetBytes(), hash.GetBytes())
}
