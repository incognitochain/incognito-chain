package devframework

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"sort"
	"strconv"

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
		return common.Hash{}, err
	}
	return hash, nil
}

func generateHashFromShardState(allShardState map[byte][]types.ShardState) (common.Hash, error) {
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
	return generateHashFromStringArray(allShardStateStr)
}
