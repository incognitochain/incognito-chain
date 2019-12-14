package statedb

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"strconv"
)

var (
	serialNumberPrefix      = []byte("serial-number-")
	allShardCommitteePrefix = []byte("all-shard-committee-")
	committeePrefix         = []byte("shard-com-")
)

func GenerateCommitteeObjectKey(shardID int, committee incognitokey.CommitteePublicKey) common.Hash {
	committeeBytes, err := committee.Bytes()
	if err != nil {
		panic(err)
	}
	prefixHash := GetShardCommitteePrefixByID(shardID)
	valueHash := common.HashH(committeeBytes)
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func GetShardCommitteePrefixByID(shardID int) []byte {
	temp := []byte(string(committeePrefix) + strconv.Itoa(shardID))
	h := common.HashH(temp)
	return h[:][:prefixHashKeyLength]
}
func GetSerialNumberPrefix() []byte {
	h := common.HashH(serialNumberPrefix)
	return h[:][:prefixHashKeyLength]
}

var _ = func() (_ struct{}) {
	m := make(map[string]string)
	prefixs := [][]byte{}
	for i := -1; i < 256; i++ {
		temp := GetShardCommitteePrefixByID(i)
		prefixs = append(prefixs, temp)
		if v, ok := m[string(temp)]; ok {
			panic("shard-com-" + strconv.Itoa(i) + " same prefix " + v)
		}
		m[string(temp)] = "shard-com-" + strconv.Itoa(i)
	}
	temp := GetSerialNumberPrefix()
	prefixs = append(prefixs, temp)
	if v, ok := m[string(temp)]; ok {
		panic("serial-number-" + " same prefix " + v)
	}
	m[string(temp)] = "serial-number-"
	for i, v1 := range prefixs {
		for j, v2 := range prefixs {
			if i == j {
				continue
			}
			if bytes.HasPrefix(v1, v2) || bytes.HasPrefix(v2, v1) {
				panic("(prefix: " + fmt.Sprintf("%+v", v1) + ", value: " + m[string(v1)] + ")" + " is prefix or being prefix of " + " (prefix: " + fmt.Sprintf("%+v", v1) + ", value: " + m[string(v2)] + ")")
			}
		}
	}
	return
}()
