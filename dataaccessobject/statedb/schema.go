package statedb

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"strconv"
)

var (
	serialNumberPrefix      = []byte("serial-number-")
	allShardCommitteePrefix = []byte("all-shard-committee-")
	committeePrefix         = []byte("shard-com-")
	rewardReceiverPrefix    = []byte("reward-receiver-")
)

func GetCommitteePrefixByShardID(shardID int) []byte {
	temp := []byte(string(committeePrefix) + strconv.Itoa(shardID))
	h := common.HashH(temp)
	return h[:][:prefixHashKeyLength]
}
func GetRewardReceiverPrefix() []byte {
	temp := []byte(rewardReceiverPrefix)
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
		temp := GetCommitteePrefixByShardID(i)
		prefixs = append(prefixs, temp)
		if v, ok := m[string(temp)]; ok {
			panic("shard-com-" + strconv.Itoa(i) + " same prefix " + v)
		}
		m[string(temp)] = "shard-com-" + strconv.Itoa(i)
	}
	// serial number
	tempSerialNumber := GetSerialNumberPrefix()
	prefixs = append(prefixs, tempSerialNumber)
	if v, ok := m[string(tempSerialNumber)]; ok {
		panic("serial-number-" + " same prefix " + v)
	}
	m[string(tempSerialNumber)] = "serial-number-"
	// reward receiver
	tempRewardReceiver := GetRewardReceiverPrefix()
	prefixs = append(prefixs, tempRewardReceiver)
	if v, ok := m[string(tempRewardReceiver)]; ok {
		panic("reward-receiver-" + " same prefix " + v)
	}
	m[string(tempRewardReceiver)] = "reward-receiver-"
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
