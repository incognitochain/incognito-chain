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

func GetShardFinalityProofPrefix(shardID byte) []byte {
	temp := make([]byte, 0, len(shardFinalityProofPrefix))
	key := append(temp, shardID)
	key = append(key, splitter...)

	return temp
}

func GetShardFinalityProofKey(shardID byte, hash common.Hash) []byte {
	key := GetShardFinalityProofPrefix(shardID)
	key = append(key, hash[:]...)
	return key
}
