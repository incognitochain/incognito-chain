package rawdb_consensus

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

var consensusDatabase incdb.Database

func SetConsensusDatabase(db incdb.Database) {
	consensusDatabase = db
}

func GetConsensusDatabase() incdb.Database {
	return consensusDatabase
}

var splitter = []byte("-[-]-")
var shardFinalityProofPrefix = []byte("s-fp" + string(splitter))
var blacklistPrefix = []byte("bd-bl" + string(splitter))

func GetShardFinalityProofPrefix(shardID byte) []byte {
	temp := make([]byte, len(shardFinalityProofPrefix))
	copy(temp, shardFinalityProofPrefix)
	key := append(temp, shardID)
	key = append(key, splitter...)

	return temp
}

func GetShardFinalityProofKey(shardID byte, hash common.Hash) []byte {
	key := GetShardFinalityProofPrefix(shardID)
	key = append(key, hash[:]...)
	return key
}

func GetByzantineBlackListKey(validator string) []byte {
	prefix := GetByzantineBlackListPrefix()
	key := append(prefix, []byte(validator)...)

	return key
}

func GetByzantineBlackListPrefix() []byte {
	temp := make([]byte, len(blacklistPrefix))
	copy(temp, blacklistPrefix)
	temp = append(temp, splitter...)
	return temp
}
