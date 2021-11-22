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
var proposeHistoryPrefix = []byte("p-h" + string(splitter))
var receiveBlockByHeightPrefix = []byte("rb-height" + string(splitter))
var receiveBlockByHashPrefix = []byte("rb-hash" + string(splitter))
var voteHistoryPrefix = []byte("v-h" + string(splitter))

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

func GetProposeHistoryPrefix(chainID int) []byte {
	temp := make([]byte, len(proposeHistoryPrefix))
	copy(temp, proposeHistoryPrefix)
	key := append(temp, common.IntToBytes(chainID)...)
	key = append(key, splitter...)
	return key
}

func GetProposeHistoryKey(chainID int, timeSlot uint64) []byte {
	temp := GetProposeHistoryPrefix(chainID)
	key := append(temp, common.Uint64ToBytes(timeSlot)...)
	return key
}

func GetReceiveBlockByHeightPrefix(chainID int) []byte {
	temp := make([]byte, len(receiveBlockByHeightPrefix))
	copy(temp, receiveBlockByHeightPrefix)
	key := append(temp, common.IntToBytes(chainID)...)
	key = append(key, splitter...)
	return key
}

func GetReceiveBlockByHeightKey(chainID int, height uint64) []byte {
	temp := GetReceiveBlockByHeightPrefix(chainID)
	key := append(temp, common.Uint64ToBytes(height)...)
	return key
}

func GetReceiveBlockByHashPrefix(chainID int) []byte {
	temp := make([]byte, len(receiveBlockByHashPrefix))
	copy(temp, receiveBlockByHashPrefix)
	key := append(temp, common.IntToBytes(chainID)...)
	key = append(key, splitter...)
	return key
}

func GetReceiveBlockByHashKey(chainID int, blockHash string) []byte {
	temp := GetReceiveBlockByHashPrefix(chainID)
	key := append(temp, []byte(blockHash)...)
	return key
}

func GetVoteHistoryPrefix(chainID int) []byte {
	temp := make([]byte, len(voteHistoryPrefix))
	copy(temp, voteHistoryPrefix)
	key := append(temp, common.IntToBytes(chainID)...)
	key = append(key, splitter...)
	return key
}

func GetVoteHistoryKey(chainID int, height uint64) []byte {
	temp := GetVoteHistoryPrefix(chainID)
	key := append(temp, common.Uint64ToBytes(height)...)
	return key
}
