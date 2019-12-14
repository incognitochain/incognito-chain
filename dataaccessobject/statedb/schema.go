package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"strconv"
)

var (
	serialNumberPrefix      = []byte("serial-number-")
	allShardCommitteePrefix = []byte("all-shard-committee-")
	committeePrefix         = []byte("shard-com-")
)

func GetShardCommitteePrefixByID(shardID int) []byte {
	temp := []byte(string(committeePrefix) + strconv.Itoa(shardID))
	h := common.HashH(temp)
	return h[:][:prefixHashKeyLength]
}
func GetSerialNumberPrefix() []byte {
	h := common.HashH(serialNumberPrefix)
	return h[:][:prefixHashKeyLength]
}
